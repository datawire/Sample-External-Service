# gRPC V2 Authz Service

This is an example service that implements Envoy's ext_authz gRPC v2 protocol.
It is meant to serve as a tool for debugging and testing and as an example starter reference for anyone building their own Envoy authorization service. You can build off this example to implement your own authorization and authenticatoin logic to suit your needs.

The hosted dockerhub images provided in the examples below have multiarch support for
`linux_amd64` and `linux_arm64`. The `make` targets for the binaries have multiarch support for `darwin_amd64`, `darwin_arm64`, `linux_amd64`, and `linux_arm64`.

By default this service listens for gRPC requests on port `2000`.
By default this service will not use TLS, but it can be configured to do so with the
environment variables doccumented below.

When Envoy makes a request to this service for authorizing a request, the service will:

- Log a bunch of information about the request.
- If the `sleepfor` header is present, it will parse it's value and sleep for that number of seconds. This can be used to simulate a slow auth service and to test authentication timeouts.
- If the `deny-me` header is present, it will deny the request.
- If the `url.path` of the request is `/deny-me/` it will deny the request.
- It will set the value of the `x-v2-overwrite` request header to: `overwritten`.
- It will append a new value to the `x-v2-append` header, creating the header if it does not already exist.

## Environment Variable Config

## Emissary-ingress Usage

## Ambassador Edge-Stack Usage

## Envoy Config Example

## Building

## Instructions for building

`git clone` this repository.

run `go mod tidy`

run `docker build -t {DOCKER_HUB_USERNAME}/{NAME_FOR_IMAGE}:{VERSION} .`

- I've already built and deployed an image you can use instead:
- `alicewasko/envoy-grpc-auth-v2:1.6`

## Instructions for Deploying

Install [Ambassador Edge Stack](https://www.getambassador.io/docs/edge-stack/latest/tutorials/getting-started/) to your Kubernetes cluster.

Create a new file `Authv2.yaml` and add the following information to it:
**Make sure to replace `{YOUR_IMAGE_HERE}` with the image you built, or the prebuilt image I provided above**

```
---
apiVersion: v1
kind: Service
metadata:
  name: grpc-auth
spec:
  type: ClusterIP
  selector:
    app: grpc-auth
  ports:
  - port: 3000
    name: http-grpc-auth
    targetPort: http-api
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grpc-auth
spec:
  replicas: 1
  strategy:
    type: RollingUpdate
  selector:
    matchLabels:
      app: grpc-auth
  template:
    metadata:
      labels:
        app: grpc-auth
    spec:
      containers:
      - name: grpc-auth
        image:  {YOUR IMAGE HERE}
        imagePullPolicy: Always
        ports:
        - name: http-api
          containerPort: 3000
        resources:
          limits:
            cpu: "0.1"
            memory: 100Mi
```

Create a new file `AuthFilterv2.yaml` and add the following to it:
This will create a filter for us, configure that filter to run on all requests, and configure that filter to send requests to this service.

```
---
apiVersion: getambassador.io/v2
kind: Filter
metadata:
  name: ext-filter
spec:
  External:
    proto: "grpc"
    auth_service: "grpc-auth:3000"
    allowed_request_headers:
    - "v2Overwrite"
    - "v2Append"
---
apiVersion: getambassador.io/v2
kind: FilterPolicy
metadata:
  name: authentication
spec:
  rules:
  - host: "*"
    path: /*
    filters:
    - name: ext-filter
```

Apply the Service, Deployment, and Filter configuration:

`kubectl apply -f Authv2.yaml`

`kubectl apply -f AuthFilterv2.yaml`

## Need a service to test this out with?

Deploy the [Quote of the Moment service](github.com/datawire/quote/) that is part of the `Ambassador Edge Stack` [quick start guide](https://www.getambassador.io/docs/edge-stack/latest/tutorials/getting-started/).
This will deploy an additional service that we can send requests to in order to test our filter service. This will deploy the quote service and make it reachable on the `/quote/` path.

Create a `QuoteDeployment.yaml` file:

```
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: quote-mapping
spec:
  prefix: /quote/
  service: quote
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: quote
  strategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: quote
    spec:
      containers:
      - name: backend
        image: docker.io/datawire/quote:0.5.0
        ports:
        - name: http
          containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: quote
  annotations:
        a8r.io/description: "Quote of the moment service"
        a8r.io/owner: "No owner"
        a8r.io/chat: "#ambassador"
        a8r.io/bugs: "https://github.com/datawire/quote/issues"
        a8r.io/documentation: "https://github.com/datawire/quote/blob/master/README.md"
        a8r.io/repository: "https://github.com/datawire/quote"
        a8r.io/support: "http://d6e.co/slack"
        a8r.io/runbook: "https://github.com/datawire/quote/blob/master/README.md"
        a8r.io/incidents: "https://github.com/datawire/quote/issues"
        a8r.io/dependencies: "None"    
spec:
  ports:
  - name: http
    port: 80
    targetPort: 8080
  selector:
    app: quote

```

Apply the quote service and it's mapping.

`kubectl apply -f QuoteDeployment.yaml`

## Test it out

1. Grab the IP of your Ambassador service:
   - `kubectl get svc -A`
   - Look for the `ambassador` service that has `Type: LoadBalancer` and use its `External IP`:

2. Make some requests to the quote service. Our external filter will run this service on the requests before they get sent off to the quote service.
   The quote service has a `/debug/` endpoint which will print out all the headers for incomming requests so we can validate whether or not our service is working.

- `curl -kv https://{YOUR_AMBASSADOR_IP}/quote/debug/ -H "v2Overwrite: test" -H "v2Append: test"`
  - You Should see that the value of the `v2Overwrite` header that we passed was changed
  - The value of the `v2Append` header should also have a new value along with the value you passed

- `curl -kv https://{YOUR_AMBASSADOR_IP}/quote/debug/ -H "deny-me: true"`
  - You should see an error is returned since the request was denied by our service

- `curl -kv https://{YOUR_AMBASSADOR_IP}/quote/debug/ -H "sleepfor: 2"`
  - You should see a 2 second delay before getting the reply from the Quote service since we told the filter service to sleep for 2 seconds.

## Check out the logs

This service will log information about the requests it processes. You can grab the logs using `kubectl` to view each request that came in:

- The path of the incomming requests
- All of the request's headers are logged
- What action was taken for the request (Allow/Deny/Error)
