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
		zap.String("Path", req.GetAttributes().GetRequest().GetHttp().GetPath()),
		zap.String("Protocol", req.GetAttributes().GetRequest().GetHttp().GetProtocol()),
		zap.String("Query", req.GetAttributes().GetRequest().GetHttp().GetQuery()),
		zap.String("Scheme", req.GetAttributes().GetRequest().GetHttp().GetScheme()),
		zap.Int64("Size", req.GetAttributes().GetRequest().GetHttp().GetSize()),
		zap.Binary("RawBody", req.GetAttributes().GetRequest().GetHttp().GetRawBody()),
		zap.Any("Check Request", req.GetAttributes().GetRequest()),
	)
	requestURI, err := url.ParseRequestURI(req.GetAttributes().GetRequest().GetHttp().GetPath())
	if err != nil {
		as.logger.Error("ERROR:", zap.Error(err))
		return &envoyAuthV3.CheckResponse{
			// Status is the response from this service to envoy
			// An HTTP 200 OK response will tell envoy to allow the request through, any other
			// response status will tell Envoy to deny the request.
			// Setting the status here is still valuable so that you can have visibility in the envoy logs as to the
			// status that the service returned (to envoy) when denying/approving a request.
			// Envoy debug level logs may be necessary to see this information.
			Status: &status.Status{Code: int32(code.Code_UNKNOWN)},
			HttpResponse: &envoyAuthV3.CheckResponse_DeniedResponse{
				// DeniedResponse is what gets sent to the downstream client by envoy when it
				// denies the request.
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
		// When the `sleepfor` header is sent, this service will sleep for x number of seconds
		// before continuing where x is the value of the sleepfor header. Must be an integer.
		// Ocasionally useful for debugging slow auth service behaviour and timeouts
		if k == "sleepfor" {
			seconds, _ := strconv.Atoi(v)
			as.logger.Info(fmt.Sprintf("%s%d%s", "Sleeping for ", seconds, " seconds..."))
			time.Sleep(time.Duration(seconds) * time.Second)
		} else if k == "deny-me" {
			denyHeader = true
		}
	}
	as.logger.Debug("|~~~~~~~~~~~~ END HEADERS ~~~~~~~~~~~~|")

	// This service will deny requests when they match the `/deny-me/` path or have a `deny-me` http header
	// Ocasionally useful for debugging and testing
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
		// By returning an HTTP 200 OK here, we are telling Envoy to allow he request to
		// its destination. HttpResponse can be used to tell Envoy to set/remove/append headers and
		// query parameters to the original request before proxying it to the upstream service.
		// Modifying the Path is not supported, by the time this authorization filter has been called by Envoy,
		// it has already matched a routing rule, so the decision on where to route this request cannot be changed
		// by any of the headers or query parameters you change below.
		HttpResponse: &envoyAuthV3.CheckResponse_OkResponse{
			OkResponse: &envoyAuthV3.OkHttpResponse{
				// You do not need to copy over all of the original headers on the incoming request.
				// Any headers on the request that are not mentioned below will not be modified.
				Headers: []*envoyCoreV3.HeaderValueOption{
					{
						Header: &envoyCoreV3.HeaderValue{Key: "x-v3-overwrite", Value: "overwritten"},
						Append: &wrappers.BoolValue{Value: false},
					},
					{
						Header: &envoyCoreV3.HeaderValue{Key: "x-v3-append", Value: "appended"},
						Append: &wrappers.BoolValue{Value: true},
					},
				},
			},
		},
	}, nil
}
