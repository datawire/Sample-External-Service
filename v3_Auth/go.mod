module github.com/AliceProxy/Envoy-Auth-gRPC/tree/main/v3_Auth

go 1.15

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	k8s.io/api v0.0.0 => k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.20.2
	k8s.io/apiserver v0.0.0 => k8s.io/apiserver v0.20.2
	k8s.io/cli-runtime v0.0.0 => k8s.io/cli-runtime v0.20.2
	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.20.2
	k8s.io/cloud-provider v0.0.0 => k8s.io/cloud-provider v0.20.2
	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.20.2
	k8s.io/code-generator v0.0.0 => k8s.io/code-generator v0.20.2
	k8s.io/component-base v0.0.0 => k8s.io/component-base v0.20.2
	k8s.io/component-helpers v0.0.0 => k8s.io/component-helpers v0.20.2
	k8s.io/controller-manager v0.0.0 => k8s.io/controller-manager v0.20.2
	k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.20.2
	k8s.io/csi-translation-lib v0.0.0 => k8s.io/csi-translation-lib v0.20.2
	k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.20.2
	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.20.2
	k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.20.2
	k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.20.2
	k8s.io/kubectl v0.0.0 => k8s.io/kubectl v0.20.2
	k8s.io/kubelet v0.0.0 => k8s.io/kubelet v0.20.2
	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.20.2
	k8s.io/metrics v0.0.0 => k8s.io/metrics v0.20.2
	k8s.io/mount-utils v0.0.0 => k8s.io/mount-utils v0.20.2
	k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.20.2
)

require (
	github.com/datawire/ambassador/v2 v2.1.2
	github.com/datawire/dlib v1.2.5-0.20211116212847-0316f8d7af2b
	github.com/golang/protobuf v1.5.2
	google.golang.org/genproto v0.0.0-20210506142907-4a47615972c2
	google.golang.org/grpc v1.37.0
)
