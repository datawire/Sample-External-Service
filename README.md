# Envoy-Auth-gRPC
Example service for authorization in Envoy over gRPC

This repository holds examples of external services for use with Ambassador Edge Stack as an example
for how you can setup external services to perform processing on requests. 

These services do not perform actual authentication, but demonstrate some of the functionality
available such as:
- Reject and allow requests
- Append and rewrite headers
- Read headers and perform actions based on values
- Perform actions based on the url path of the reuqest

The service may be deployed as an Edge-Stack `External Filter` and/or an Emissary-ingress `AuthService`

There are three versions for historical and future examples.
- v2Alpha # Deprecated, preserved for historical reference
- v2      # This is supported in Emissary-ingress / Edge-Stack `1.x` - `2.2.z`
- v3      # This version is supported as of `2.3.0` and as of `3.0.0`+ is the only supported version.

Instructions for building and deploying them can be found in each directory.
