// From DataWire Example service as base

package main

// NOTE: VERY WIP, DOES NOT WORK YET

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"

	envoyCoreV3 "github.com/datawire/ambassador/pkg/api/envoy/config/core/v3" // As of 05/06/2021 there is no /api/v3
	envoyAuthV3 "github.com/datawire/ambassador/pkg/api/envoy/service/auth/v3"
	envoyType "github.com/datawire/ambassador/pkg/api/envoy/type/v3"

	"github.com/datawire/dlib/dhttp"
)

func main() {
	grpcHandler := grpc.NewServer()
	envoyAuthV3.RegisterAuthorizationServer(grpcHandler, &AuthService{})

	sc := &dhttp.ServerConfig{
		Handler: grpcHandler,
	}

	log.Print("starting...")
	log.Fatal(sc.ListenAndServe(context.Background(), ":3000"))
}

type AuthService struct{}

func (s *AuthService) Check(ctx context.Context, req *envoyAuthV2.CheckRequest) (*envoyAuthV2.CheckResponse, error) {
	log.Println("ACCESS",
		req.GetAttributes().GetRequest().GetHttp().GetMethod(),
		req.GetAttributes().GetRequest().GetHttp().GetHost(),
		req.GetAttributes().GetRequest().GetHttp().GetPath(),
	)
	requestURI, err := url.ParseRequestURI(req.GetAttributes().GetRequest().GetHttp().GetPath())
	if err != nil {
		log.Println("=> ERROR", err)
		return &envoyAuthV2.CheckResponse{
			Status: &status.Status{Code: int32(code.Code_UNKNOWN)},
			HttpResponse: &envoyAuthV2.CheckResponse_DeniedResponse{
				DeniedResponse: &envoyAuthV2.DeniedHttpResponse{
					Status: &envoyType.HttpStatus{Code: http.StatusInternalServerError},
					Headers: []*envoyCoreV2.HeaderValueOption{
						{Header: &envoyCoreV2.HeaderValue{Key: "Content-Type", Value: "application/json"}},
					},
					Body: `{"msg": "internal server error"}`,
				},
			},
		}, nil
	}

	log.Print("=> Allow")
	return &envoyAuthV2.CheckResponse{
		Status: &status.Status{Code: int32(code.Code_UNKNOWN)},
		HttpResponse: &envoyAuthV2.CheckResponse_DeniedResponse{
			DeniedResponse: &envoyAuthV2.DeniedHttpResponse{
				Status: &envoyType.HttpStatus{Code: http.StatusOK},
				Headers: []*envoyCoreV2.HeaderValueOption{
					{Header: &envoyCoreV2.HeaderValue{Key: "X-Allowed-Output-Header", Value: "baz"}},
					{Header: &envoyCoreV2.HeaderValue{Key: "X-Disallowed-Output-Header", Value: "qux"}},
					{Header: &envoyCoreV2.HeaderValue{Key: "Content-Type", Value: "application/json"}},
				},
				Body: `{"msg": "intercepted"}`,
			},
		},
	}, nil
}