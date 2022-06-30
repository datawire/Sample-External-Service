# Envoy-Auth-gRPC-v3

## V3 of the auth service API used for Ambassador Edge Stack has been released as part of Emissary-ingress/Edge-Stack version `2.3.0`. Version `2.3.0` supports both the `v2` and the `v3` API, whereas as of Emissary-ingress/Edge-Stack version `3.0.0` only the `v3` transport protocol is supported.

## Please see the v2 service if you are looking for an example you can use for Emissary-ingress/Edge-Stack versions prior to `2.3.0`.


# v3_Auth
Example service for authorization in Envoy over gRPC using transport protocol `v3`

This is an example service for performing operations on requests. It is meant to be used with
an Ambassador Edge Stack `External Filter`.

When requests pass through this filter it will:
- Read and log all of the headers.
- If the `sleepfor` header is present, it will parse it's value and sleep for that ammount of time to simulate a timeout or delayed response form the service.
- If the `deny-me` header is present, it will deny the request.
- If the `url.path` of the request is `/deny-me/` it will deny the request.
- Rewrite the value of the `v3alphaOverwrite` header.
- Append a new value to the `v3alphaAppend` header.

> You can find a list of status codes to be used with this external filter at the following two links:
> - https://pkg.go.dev/google.golang.org/genproto/googleapis/rpc/codego
> - https://pkg.go.dev/net/http#pkg-variables

## Instructions for building:

`git clone` this repository. 

run `go mod tidy`

run `docker build -t {DOCKER_HUB_USERNAME}/{NAME_FOR_IMAGE}:{VERSION} .`
- I've already built and deployed an image you can use instead: 
- `alicewasko/grpcauth-v3:0.1`


## Instructions for Deploying:

Install [Ambassador Edge Stack](https://www.getambassador.io/docs/edge-stack/latest/tutorials/getting-started/) to your Kubernetes cluster.



Create a new file `Authv3.yaml` and add the following information to it:
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
        image:  alicewasko/grpcauth-v3:0.1
        imagePullPolicy: Always
        ports:
        - name: http-api
          containerPort: 3000
        resources:
          limits:
            cpu: "0.1"
            memory: 100Mi
```



Create a new file `AuthFilterv3.yaml` and add the following to it:
This will create a filter for us, configure that filter to run on all requests, and configure that filter to send requests to this service.

```
---
apiVersion: getambassador.io/v3alpha1
kind: Filter
metadata:
  name: ext-filter-v3
spec:
  External:
    proto: "grpc"
    protocol_version: v3
    auth_service: "grpc-auth.default:3000"
    allowed_request_headers:
    - "v3alphaOverwrite"
    - "v3alphaAppend"
---
apiVersion: getambassador.io/v3alpha1
kind: FilterPolicy
metadata:
  name: ext-filter-v3
spec:
  rules:
  - host: "*"
    path: /**
    filters:
    - name: ext-filter-v3
```

Apply the Service, Deployment, and Filter configuration:

`kubectl apply -f Authv3.yaml`

`kubectl apply -f AuthFilterv3.yaml`


## Need a service to test this out with?

Deploy the [Quote of the Moment service](github.com/datawire/quote/) that is part of the `Ambassador Edge Stack` [quick start guide](https://www.getambassador.io/docs/edge-stack/latest/tutorials/getting-started/).
This will deploy an additional service that we can send requests to in order to test our filter service. This will deploy the quote service and make it reachable on the `/quote/` path.

Create a `QuoteDeployment.yaml` file:
```
---
apiVersion: getambassador.io/v3alpha1
kind: Mapping
metadata:
  name: quote-mapping
spec:
  hostname: "*"
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

## Test it out:

1. Grab the IP of your Ambassador service:
   - `kubectl get svc -A` 
   - Look for the `ambassador` service that has `Type: LoadBalancer` and use its `External IP`:

2. Make some requests to the quote service. Our external filter will run this service on the requests before they get sent off to the quote service.
   The quote service has a `/debug/` endpoint which will print out all the headers for incomming requests so we can validate whether or not our service is working.

  - `curl -kv https://{YOUR_AMBASSADOR_IP}/quote/debug/ -H "v3alphaOverwrite: test" -H "v3alphaAppend: test"`
    - You Should see that the value of the `v3alphaOverwrite` header that we passed was changed
    - The value of the `v3alphaAppend` header should also have a new value along with the value you passed

  - `curl -kv https://{YOUR_AMBASSADOR_IP}/quote/debug/ -H "deny-me: true"`
    - You should see an error is returned since the request was denied by our service

  - `curl -kv https://{YOUR_AMBASSADOR_IP}/quote/debug/ -H "sleepfor: 2"`
    - You should see a 2 second delay before getting the reply from the Quote service since we told the filter service to sleep for 2 seconds.


## Check out the logs:

This service will log information about the requests it processes. You can grab the logs using `kubectl` to view each request that came in:
-  The path of the incomming requests
-  All of the request's headers are logged
-  What action was taken for the request (Allow/Deny/Error)

