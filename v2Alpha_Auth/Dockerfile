FROM golang:1.15 as foundation

WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM foundation as builder

COPY . .
RUN make

FROM gcr.io/distroless/base as runtime

COPY --from=builder /build/bin/authv2-linux-amd64 /bin/authv2

ENTRYPOINT ["/bin/authv2"]
