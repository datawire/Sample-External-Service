module github.com/AliceProxy/Envoy-Auth-gRPC/tree/main/v2Alpha_Auth

go 1.15

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)

require (
	github.com/datawire/ambassador v1.13.3
	github.com/datawire/dlib v1.2.1
	github.com/golang/protobuf v1.5.2
	google.golang.org/genproto v0.0.0-20210506142907-4a47615972c2
	google.golang.org/grpc v1.37.0
)
