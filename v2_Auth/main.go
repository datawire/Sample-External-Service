// From DataWire Example service as base

package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"regexp"
	"os"

	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"

	envoyCoreV2 "github.com/datawire/ambassador/pkg/api/envoy/api/v2/core"
	envoyAuthV2 "github.com/datawire/ambassador/pkg/api/envoy/service/auth/v2"
	envoyType "github.com/datawire/ambassador/pkg/api/envoy/type"

	"github.com/datawire/dlib/dhttp"
)

const (
	EnvAction        	= "ACTION"      // Can be "allow, 403, 404, 503. Default: 404
	EnvQuery        	= "QUERY"       // Value of the query to take action on
	EnvQueryRegex       = "REGEX"       // Can be "True/False. Sets whether the above query should be interpreted as a string or a regex. Default: false"
)

var action = "404"
var query = `\?`
var queryRegex = false

func main() {
	grpcHandler := grpc.NewServer()
	envoyAuthV2.RegisterAuthorizationServer(grpcHandler, &AuthService{})

	sc := &dhttp.ServerConfig{
		Handler: grpcHandler,
	}

	
	envQueryRegex, err := strconv.ParseBool(getEnv(EnvQueryRegex, "false"))
	if err != nil {
		log.Println("ERROR: REGEX environment variable must be either 'true' or 'false'.")
	} else {
		queryRegex = envQueryRegex
	}

	envAction := getEnv(EnvAction, "404")
	if envAction == "" {
		log.Println("ACTION envionment variable not set, defaulting to returning 404s on matched queries")
	}
	action = envAction

	envQuery := getEnv(EnvQuery, "")
	if envQuery == "" {
		log.Println("ERROR: query envionment variable cannot be empty.")
	}
	query += envQuery


	log.Print("Starting Envoy Auth Service v2 Over gRPC...")
	log.Print("Rejecting all requests with query parameters...")
	log.Print("Listening on port: 3000")
	log.Fatal(sc.ListenAndServe(context.Background(), ":3000"))
}

type AuthService struct{}

func getEnv(name, fallback string) string {
	res := os.Getenv(name)
	if res == "" {
		res = fallback
	}

	return res
}
func (s *AuthService) Check(ctx context.Context, req *envoyAuthV2.CheckRequest) (*envoyAuthV2.CheckResponse, error) {
	// Log some info about the request
	log.Println("~~~~~~~~> INCOMMING REQUEST ~~~~~~~~>")
	log.Println("Method: ", req.GetAttributes().GetRequest().GetHttp().GetMethod())
	log.Println("Host: ", req.GetAttributes().GetRequest().GetHttp().GetHost())
	log.Println("Path: ", req.GetAttributes().GetRequest().GetHttp().GetPath())
	
	requestURI, err := url.ParseRequestURI(req.GetAttributes().GetRequest().GetHttp().GetPath())
	// Unable to read the request URI
	if err != nil {
		log.Println("<~~~~~~~~ ERROR <~~~~~~~~", err)
		return &envoyAuthV2.CheckResponse{
			Status: &status.Status{Code: int32(code.Code_UNKNOWN)},
			HttpResponse: &envoyAuthV2.CheckResponse_DeniedResponse{
				DeniedResponse: &envoyAuthV2.DeniedHttpResponse{
					Status: &envoyType.HttpStatus{Code: http.StatusInternalServerError},
					Headers: []*envoyCoreV2.HeaderValueOption{
						{Header: &envoyCoreV2.HeaderValue{Key: "Content-Type", Value: "text/html"}},
					},
					Body: `Internal Server Error`,
				},
			},
		}, nil
	}
	log.Println("URI: ", requestURI)
	
	queryRegx := regexp.MustCompile(query)
	containsQuery := queryRegx.FindStringIndex(requestURI.String()) != nil
	
	if !containsQuery {
		log.Println("<~~~~~~~~ ALLOW REQUEST <~~~~~~~~ ")
		log.Println("Request does not contain query, allowing")
		header := make([]*envoyCoreV2.HeaderValueOption, 0, 4+len(req.GetAttributes().GetRequest().GetHttp().GetHeaders()))
		return &envoyAuthV2.CheckResponse{
			Status: &status.Status{Code: int32(code.Code_OK)},
			HttpResponse: &envoyAuthV2.CheckResponse_OkResponse{
				OkResponse: &envoyAuthV2.OkHttpResponse{
					Headers: header,
				},
			},
		}, nil
	}

	log.Println("Request contains matched query")

	switch action {
        case "allow":
			log.Println("<~~~~~~~~ ALLOW REQUEST <~~~~~~~~ ")
			log.Println("Allowing request with query anyways")
			header := make([]*envoyCoreV2.HeaderValueOption, 0, 4+len(req.GetAttributes().GetRequest().GetHttp().GetHeaders()))
			return &envoyAuthV2.CheckResponse{
				Status: &status.Status{Code: int32(code.Code_OK)},
				HttpResponse: &envoyAuthV2.CheckResponse_OkResponse{
					OkResponse: &envoyAuthV2.OkHttpResponse{
						Headers: header,
					},
				},
			}, nil
        case "400":
			log.Println("<~~~~~~~~ DENIED REQUEST <~~~~~~~~")
			log.Println("returning a 400")
			return &envoyAuthV2.CheckResponse{
				Status: &status.Status{Code: int32(code.Code_INVALID_ARGUMENT)},
				HttpResponse: &envoyAuthV2.CheckResponse_DeniedResponse{
					DeniedResponse: &envoyAuthV2.DeniedHttpResponse{
						Status: &envoyType.HttpStatus{Code: http.StatusBadRequest},
						Headers: []*envoyCoreV2.HeaderValueOption{
							{Header: &envoyCoreV2.HeaderValue{Key: "Content-Type", Value: "text/html"}},
						},
						Body: `Bad Request`,
					},
				},
			}, nil
        case "403":
			log.Println("<~~~~~~~~ DENIED REQUEST <~~~~~~~~")
			log.Println("returning a 403")
			return &envoyAuthV2.CheckResponse{
				Status: &status.Status{Code: int32(code.Code_PERMISSION_DENIED)},
				HttpResponse: &envoyAuthV2.CheckResponse_DeniedResponse{
					DeniedResponse: &envoyAuthV2.DeniedHttpResponse{
						Status: &envoyType.HttpStatus{Code: http.StatusForbidden},
						Headers: []*envoyCoreV2.HeaderValueOption{
							{Header: &envoyCoreV2.HeaderValue{Key: "Content-Type", Value: "text/html"}},
						},
						Body: `Unauthorized`,
					},
				},
			}, nil
        case "404":
			log.Println("<~~~~~~~~ DENIED REQUEST <~~~~~~~~")
			log.Println("returning a 404")
			return &envoyAuthV2.CheckResponse{
				Status: &status.Status{Code: int32(code.Code_NOT_FOUND)},
				HttpResponse: &envoyAuthV2.CheckResponse_DeniedResponse{
					DeniedResponse: &envoyAuthV2.DeniedHttpResponse{
						Status: &envoyType.HttpStatus{Code: http.StatusNotFound},
						Headers: []*envoyCoreV2.HeaderValueOption{
							{Header: &envoyCoreV2.HeaderValue{Key: "Content-Type", Value: "text/html"}},
						},
						Body: `Not Found`,
					},
				},
			}, nil
        default:
			log.Println("<~~~~~~~~ DENIED REQUEST <~~~~~~~~")
			log.Println("Defaulting to denying request with a 404")
			return &envoyAuthV2.CheckResponse{
				Status: &status.Status{Code: int32(code.Code_NOT_FOUND)},
				HttpResponse: &envoyAuthV2.CheckResponse_DeniedResponse{
					DeniedResponse: &envoyAuthV2.DeniedHttpResponse{
						Status: &envoyType.HttpStatus{Code: http.StatusNotFound},
						Headers: []*envoyCoreV2.HeaderValueOption{
							{Header: &envoyCoreV2.HeaderValue{Key: "Content-Type", Value: "text/html"}},
						},
						Body: `Not Found`,
					},
				},
			}, nil
	}
}
