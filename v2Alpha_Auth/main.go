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

	envoyCoreV2 "github.com/datawire/ambassador/pkg/api/envoy/api/v2/core"
	envoyAuthV2 "github.com/datawire/ambassador/pkg/api/envoy/service/auth/v2"
	envoyAuthV2alpha "github.com/datawire/ambassador/pkg/api/envoy/service/auth/v2alpha"
	envoyType "github.com/datawire/ambassador/pkg/api/envoy/type"

	"github.com/datawire/dlib/dhttp"
)

func main() {
	grpcHandler := grpc.NewServer()
	envoyAuthV2alpha.RegisterAuthorizationServer(grpcHandler, &AuthService{})
	envoyAuthV2.RegisterAuthorizationServer(grpcHandler, &AuthService{})

	sc := &dhttp.ServerConfig{
		Handler: grpcHandler,
	}

	log.Print("Starting Envoy Auth Service v2Alpha Over gRPC...")
	log.Print("Listening on port: 3000")
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
	log.Println("RequestURI: ", requestURI);

	if requestURI.Path == "/deny-me/" {
		log.Println("=> DENIED REQUEST", err)
		return &envoyAuthV2.CheckResponse{
			Status: &status.Status{Code: int32(code.Code_PERMISSION_DENIED)},
			HttpResponse: &envoyAuthV2.CheckResponse_DeniedResponse{
				DeniedResponse: &envoyAuthV2.DeniedHttpResponse{
					Status: &envoyType.HttpStatus{Code: http.StatusOK},
					Headers: []*envoyCoreV2.HeaderValueOption{
						{Header: &envoyCoreV2.HeaderValue{Key: "Content-Type", Value: "application/json"}},
					},
					Body: `{"msg": "Your request was denied, unauthorized path /deny-me/"}`,
				},
			},
		}, nil
	}

	log.Print("=> ALLOW REQUEST")
	return &envoyAuthV2.CheckResponse{
        Status: &status.Status{Code: int32(code.Code_OK)},
                HttpResponse: &envoyAuthV2.CheckResponse_OkResponse{
                    OkResponse: &envoyAuthV2.OkHttpResponse{
                    	Headers: []*envoyCoreV2.HeaderValueOption{
                    	{
                            Header: &envoyCoreV2.HeaderValue{Key: "v2AlphaOverwrite", Value: "HeaderOverwritten"},
                            Append: &wrappers.BoolValue{Value: false},
                    	},
						{
                            Header: &envoyCoreV2.HeaderValue{Key: "v2AlphaAppend", Value: "HeaderAppended"},
                            Append: &wrappers.BoolValue{Value: true},
                    	},
                    },
                },
        },
	}, nil

}
