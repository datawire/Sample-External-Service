// From DataWire Example service as base

package main

// NOTE: VERY WIP, DOES NOT WORK YET

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

	envoyCoreV3 "github.com/datawire/ambassador/v2/pkg/api/envoy/config/core/v3"
	envoyAuthV3 "github.com/datawire/ambassador/v2/pkg/api/envoy/service/auth/v3"
	envoyType "github.com/datawire/ambassador/v2/pkg/api/envoy/type/v3"

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

func (s *AuthService) Check(ctx context.Context, req *envoyAuthV3.CheckRequest) (*envoyAuthV3.CheckResponse, error) {
	log.Println("ACCESS",
		req.GetAttributes().GetRequest().GetHttp().GetMethod(),
		req.GetAttributes().GetRequest().GetHttp().GetHost(),
		req.GetAttributes().GetRequest().GetHttp().GetBody(),
	)
	log.Println("~~~~~~~~> REQUEST BODY ~~~~~~~~>", req.GetAttributes().GetRequest().GetHttp().GetBody())
	log.Println("~~~~~~~~> REQUEST RAW BODY ~~~~~~~~>", req.GetAttributes().GetRequest().GetHttp().GetRawBody())
	log.Println("~~~~~~~~> REQUEST HTTP ~~~~~~~~>", req.GetAttributes().GetRequest().GetHttp())
	log.Println("~~~~~~~~> REQUEST ~~~~~~~~>", req.GetAttributes().GetRequest())
	requestURI, err := url.ParseRequestURI(req.GetAttributes().GetRequest().GetHttp().GetPath())
	if err != nil {
		log.Println("=> ERROR", err)
		return &envoyAuthV3.CheckResponse{
			Status: &status.Status{Code: int32(code.Code_UNKNOWN)},
			HttpResponse: &envoyAuthV3.CheckResponse_DeniedResponse{
				DeniedResponse: &envoyAuthV3.DeniedHttpResponse{
					Status: &envoyType.HttpStatus{Code: http.StatusInternalServerError},
					Headers: []*envoyCoreV3.HeaderValueOption{
						{Header: &envoyCoreV3.HeaderValue{Key: "Content-Type", Value: "application/json"}},
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

	if requestURI.Path == "/deny-me/" || denyHeader {
		log.Println("=> DENIED REQUEST", err)
		return &envoyAuthV3.CheckResponse{
			Status: &status.Status{Code: int32(code.Code_PERMISSION_DENIED)},
			HttpResponse: &envoyAuthV3.CheckResponse_DeniedResponse{
				DeniedResponse: &envoyAuthV3.DeniedHttpResponse{
					Status: &envoyType.HttpStatus{Code: http.StatusForbidden},
					Headers: []*envoyCoreV3.HeaderValueOption{
						{Header: &envoyCoreV3.HeaderValue{Key: "Content-Type", Value: "application/json"}},
					},
					Body: `{"msg": "Your request was denied, unauthorized path /deny-me/"}`,
				},
			},
		}, nil
	}

	log.Print("=> ALLOW REQUEST")
	return &envoyAuthV3.CheckResponse{
		Status: &status.Status{Code: int32(code.Code_OK)},
		HttpResponse: &envoyAuthV3.CheckResponse_OkResponse{
			OkResponse: &envoyAuthV3.OkHttpResponse{
				Headers: []*envoyCoreV3.HeaderValueOption{
					{
						Header: &envoyCoreV3.HeaderValue{Key: "V3AlphaOverwrite", Value: "HeaderOverwritten"},
						Append: &wrappers.BoolValue{Value: false},
					},
					{
						Header: &envoyCoreV3.HeaderValue{Key: "V3AlphaAppend", Value: "HeaderAppended"},
						Append: &wrappers.BoolValue{Value: true},
					},
				},
			},
		},
	}, nil

}
