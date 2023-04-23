// From DataWire Example service as base

package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"

	envoyCoreV2 "github.com/emissary-ingress/emissary/v3/pkg/api/envoy/api/v2/core"
	envoyAuthV2 "github.com/emissary-ingress/emissary/v3/pkg/api/envoy/service/auth/v2"
	envoyType "github.com/emissary-ingress/emissary/v3/pkg/api/envoy/type"

	"github.com/datawire/dlib/dhttp"
)

func main() {
	grpcHandler := grpc.NewServer()
	envoyAuthV2.RegisterAuthorizationServer(grpcHandler, &AuthService{})

	sc := &dhttp.ServerConfig{
		Handler: grpcHandler,
	}

	log.Print("Starting Envoy Auth Service v2 Over gRPC...")
	log.Print("Listening on port: 3000")
	log.Fatal(sc.ListenAndServe(context.Background(), ":3000"))
}

type AuthService struct{}

func (s *AuthService) Check(ctx context.Context, req *envoyAuthV2.CheckRequest) (*envoyAuthV2.CheckResponse, error) {
	// Log some info about the request
	log.Println("~~~~~~~~> INCOMMING REQUEST ~~~~~~~~>",
		req.GetAttributes().GetRequest().GetHttp().GetMethod(),
		req.GetAttributes().GetRequest().GetHttp().GetHost(),
		req.GetAttributes().GetRequest().GetHttp().GetPath(),
		req.GetAttributes().GetRequest().GetHttp().GetBody(),
	)
	log.Println("~~~~~~~~> REQUEST BODY ~~~~~~~~>", req.GetAttributes().GetRequest().GetHttp().GetBody())
	log.Println("~~~~~~~~> REQUEST HTTP ~~~~~~~~>", req.GetAttributes().GetRequest().GetHttp())
	log.Println("~~~~~~~~> REQUEST ~~~~~~~~>", req.GetAttributes().GetRequest())
	requestURI, err := url.ParseRequestURI(req.GetAttributes().GetRequest().GetHttp().GetPath())
	if err != nil {
		log.Println("<~~~~~~~~ ERROR <~~~~~~~~", err)
		return &envoyAuthV2.CheckResponse{
			// https://pkg.go.dev/google.golang.org/genproto/googleapis/rpc/code
			Status: &status.Status{Code: int32(code.Code_UNKNOWN)},
			HttpResponse: &envoyAuthV2.CheckResponse_DeniedResponse{
				DeniedResponse: &envoyAuthV2.DeniedHttpResponse{
					// https://pkg.go.dev/net/http#pkg-variables
					Status: &envoyType.HttpStatus{Code: http.StatusInternalServerError},
					Headers: []*envoyCoreV2.HeaderValueOption{
						{Header: &envoyCoreV2.HeaderValue{Key: "Content-Type", Value: "application/json"}},
					},
					Body: `{"msg": "internal server error"}`,
				},
			},
		}, nil
	}
	log.Println("RequestURI: ", requestURI)

	// Read over and log the headers for the request
	denyHeader := false
	log.Println("|~~~~~~~~~~~~ BEGIN HEADERS ~~~~~~~~~~~~|")
	for k, v := range req.GetAttributes().GetRequest().GetHttp().GetHeaders() {
		log.Printf("%s: %s", k, v)
		// Sleep for x seconds when this header is present
		if k == "sleepfor" {
			seconds, _ := strconv.Atoi(v)
			log.Printf("%s%d%s", "Sleeping for ", seconds, " seconds...")
			time.Sleep(time.Duration(seconds) * time.Second)
		} else if k == "deny-me" {
			denyHeader = true
		}
	}
	log.Println("|~~~~~~~~~~~~ END HEADERS ~~~~~~~~~~~~|")

	// You can perform operations based on the path, or you can perform them based on the  headers we read in earlier
	if requestURI.Path == "/deny-me/" || denyHeader {
		log.Print("<~~~~~~~~ DENIED REQUEST <~~~~~~~~")
		// Define your headers in the return statement
		return &envoyAuthV2.CheckResponse{
			Status: &status.Status{Code: int32(code.Code_PERMISSION_DENIED)},
			HttpResponse: &envoyAuthV2.CheckResponse_DeniedResponse{
				DeniedResponse: &envoyAuthV2.DeniedHttpResponse{
					Status: &envoyType.HttpStatus{Code: http.StatusForbidden},
					Headers: []*envoyCoreV2.HeaderValueOption{
						{Header: &envoyCoreV2.HeaderValue{Key: "Content-Type", Value: "application/json"}},
					},
					Body: `{"msg": "Your request was denied, unauthorized path /deny-me/"}`,
				},
			},
		}, nil
	}

	log.Print("<~~~~~~~~ ALLOW REQUEST <~~~~~~~~ ")
	// Or work on the headers before building the return statement
	header := make([]*envoyCoreV2.HeaderValueOption, 0, 4+len(req.GetAttributes().GetRequest().GetHttp().GetHeaders()))
	header = append(header, &envoyCoreV2.HeaderValueOption{
		Header: &envoyCoreV2.HeaderValue{Key: "v2Overwrite", Value: "HeaderOverwritten"},
		// This will overwrite the value of this header if it exists
		Append: &wrappers.BoolValue{Value: false},
	})
	header = append(header, &envoyCoreV2.HeaderValueOption{
		Header: &envoyCoreV2.HeaderValue{Key: "v2Append", Value: "HeaderAppended"},
		// This will append this value to the header if it exists
		Append: &wrappers.BoolValue{Value: true},
	})

	return &envoyAuthV2.CheckResponse{
		Status: &status.Status{Code: int32(code.Code_OK)},
		HttpResponse: &envoyAuthV2.CheckResponse_OkResponse{
			OkResponse: &envoyAuthV2.OkHttpResponse{
				Headers: header,
			},
		},
	}, nil
}
