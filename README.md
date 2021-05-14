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

There are three versions for historical and future examples.
- v2Alpha # Deprecated, preserved for historical reference
- v2      # This is the currently supported version in Ambassador
- v3      # This version is not supported yet, but will be when v2 hits EOL

Instructions for building and deploying them can be found in each directory.