# Use distroless as minimal base image to package the eg-plus binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
ARG TARGETPLATFORM
ARG BUILD_DIR=.

WORKDIR /
COPY $BUILD_DIR/$TARGETPLATFORM/grpc-v2-authz /usr/local/bin/

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/grpc-v2-authz"]
