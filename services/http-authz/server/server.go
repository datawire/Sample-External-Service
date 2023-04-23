package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	envoyCoreV3 "github.com/emissary-ingress/emissary/v3/pkg/api/envoy/config/core/v3"
	envoyAuthV3 "github.com/emissary-ingress/emissary/v3/pkg/api/envoy/service/auth/v3"
	envoyTypeV3 "github.com/emissary-ingress/emissary/v3/pkg/api/envoy/type/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
)

type GRPCAuthV3Server struct {
	logger     *zap.Logger
	grpcServer *grpc.Server
	port       int
	tls        bool
	tlsFile    string
}

func NewGRPCAuthV3Server(logger *zap.Logger, port int, tls bool, tlsFile string) *GRPCAuthV3Server {
	grpcServer := grpc.NewServer()
	envoyAuthV3.RegisterAuthorizationServer(grpcServer, &authzServer{
		logger: logger,
	})
	return &GRPCAuthV3Server{
		logger:     logger,
		grpcServer: grpcServer,
		port:       port,
		tls:        tls,
		tlsFile:    tlsFile,
	}
}

// Start this will start the server and block until a shutdown signal or an error occurs
func (s *GRPCAuthV3Server) Start(ctx context.Context) error {
	listenOn := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		return fmt.Errorf("GRPCAuthV3Server failed to listen on %s: %w", listenOn, err)
	}

	errChan := make(chan error)
	go func() {
		s.logger.Info("starting ext_authz grpc server using protocol version: v3", zap.String("address", listenOn))
		errChan <- s.grpcServer.Serve(listener)
	}()

	// wait for shut down or grpc accept error to occur
	select {
	case <-ctx.Done():
		s.logger.Info("ext_authz grpc v3 graceful shutdown started")
		s.grpcServer.GracefulStop()
		s.logger.Info("ext_authz grpc v3 server successfully shutdown")
		return nil
	case err := <-errChan:
		return err
	}
}

// authZServer implements Envoy's `AuthorizationServer`interface for ext_authz
type authzServer struct {
	logger *zap.Logger
}

// Ensure that authzServer implements Envoy's `AuthorizationServer`interface for ext_authz
var _ envoyAuthV3.AuthorizationServer = (*authzServer)(nil)

// Check implements the Envoy ext_authz service Check method
func (as *authzServer) Check(ctx context.Context, req *envoyAuthV3.CheckRequest) (*envoyAuthV3.CheckResponse, error) {

	as.logger.Debug("ACCESS",
		zap.String("Method", req.GetAttributes().GetRequest().GetHttp().GetMethod()),
		zap.String("Host", req.GetAttributes().GetRequest().GetHttp().GetHost()),
		zap.String("Body", req.GetAttributes().GetRequest().GetHttp().GetBody()),
		zap.Binary("RawBody", req.GetAttributes().GetRequest().GetHttp().GetRawBody()),
		zap.Any("HTTP", req.GetAttributes().GetRequest().GetHttp()),
		zap.Any("Check Request", req.GetAttributes().GetRequest()),
	)
	requestURI, err := url.ParseRequestURI(req.GetAttributes().GetRequest().GetHttp().GetPath())
	if err != nil {
		as.logger.Error("ERROR:", zap.Error(err))
		return &envoyAuthV3.CheckResponse{
			Status: &status.Status{Code: int32(code.Code_UNKNOWN)},
			HttpResponse: &envoyAuthV3.CheckResponse_DeniedResponse{
				DeniedResponse: &envoyAuthV3.DeniedHttpResponse{
					Status: &envoyTypeV3.HttpStatus{Code: http.StatusInternalServerError},
					Headers: []*envoyCoreV3.HeaderValueOption{
						{Header: &envoyCoreV3.HeaderValue{Key: "Content-Type", Value: "application/json"}},
					},
					Body: `{"msg": "internal server error"}`,
				},
			},
		}, nil
	}
	as.logger.Info("", zap.String("RequestURI", requestURI.String()))

	// Read over and log the headers for the request
	denyHeader := false
	as.logger.Debug("|~~~~~~~~~~~~ BEGIN HEADERS ~~~~~~~~~~~~|")
	for k, v := range req.GetAttributes().GetRequest().GetHttp().GetHeaders() {
		as.logger.Debug("", zap.String("header", k), zap.String("value", v))
		// Sleep for x seconds when this header is present
		if k == "sleepfor" {
			seconds, _ := strconv.Atoi(v)
			as.logger.Info(fmt.Sprintf("%s%d%s", "Sleeping for ", seconds, " seconds..."))
			time.Sleep(time.Duration(seconds) * time.Second)
		} else if k == "deny-me" {
			denyHeader = true
		}
	}
	as.logger.Debug("|~~~~~~~~~~~~ END HEADERS ~~~~~~~~~~~~|")

	if requestURI.Path == "/deny-me/" || denyHeader {
		as.logger.Info("=> DENYING REQUEST")
		return &envoyAuthV3.CheckResponse{
			Status: &status.Status{Code: int32(code.Code_PERMISSION_DENIED)},
			HttpResponse: &envoyAuthV3.CheckResponse_DeniedResponse{
				DeniedResponse: &envoyAuthV3.DeniedHttpResponse{
					Status: &envoyTypeV3.HttpStatus{Code: http.StatusForbidden},
					Headers: []*envoyCoreV3.HeaderValueOption{
						{Header: &envoyCoreV3.HeaderValue{Key: "Content-Type", Value: "application/json"}},
					},
					Body: `{"msg": "Your request was denied, unauthorized path /deny-me/"}`,
				},
			},
		}, nil
	}

	as.logger.Info("=> ALLOWING REQUEST")
	return &envoyAuthV3.CheckResponse{
		Status: &status.Status{Code: int32(code.Code_OK)},
		HttpResponse: &envoyAuthV3.CheckResponse_OkResponse{
			OkResponse: &envoyAuthV3.OkHttpResponse{
				Headers: []*envoyCoreV3.HeaderValueOption{
					{
						Header: &envoyCoreV3.HeaderValue{Key: "V3Overwrite", Value: "HeaderOverwritten"},
						Append: &wrappers.BoolValue{Value: false},
					},
					{
						Header: &envoyCoreV3.HeaderValue{Key: "V3Append", Value: "HeaderAppended"},
						Append: &wrappers.BoolValue{Value: true},
					},
				},
			},
		},
	}, nil
}
