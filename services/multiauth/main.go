package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/datawire/Sample-External-Service/services/grpc-v3-authz/server"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	EnvLogLevel = "LOG_LEVEL" // Controls logging verbosity, default: debug

	EnvV3Enabled = "GRPC_V3_ENABLED"       // Controls whether to setup the grpc v3 auth service, default: true
	EnvV3TLS     = "GRPC_V3_TLS"           // Whether to use TLS, default: false
	EnvV3TLSFile = "GRPC_V3_TLS_CERT_FILE" // Custom TLS file to use for certs, defaults to using a generated self-signed certificate
	EnvV3Port    = "GRPC_V3_PORT"          // Port to listen on, default: 3000

	EnvV2Enabled = "GRPC_V2_ENABLED"       // Controls whether to setup the grpc v2 auth service, default: false
	EnvV2TLS     = "GRPC_V2_TLS"           // Whether to use TLS, default: false
	EnvV2LSFile  = "GRPC_V2_TLS_CERT_FILE" // Custom TLS file to use for certs, defaults to using a generated self-signed certificate
	EnvV2Port    = "GRPC_V2_PORT"          // Port to listen on, default: 2000

	EnvV2AEnabled = "GRPC_V2ALPHA_ENABLED"       // Controls whether to setup the grpc v2alpha auth service, default: false
	EnvV2ATLS     = "GRPC_V2ALPHA_TLS"           // Whether to use TLS, default: false
	EnvV2ATLSFile = "GRPC_V2ALPHA_TLS_CERT_FILE" // Custom TLS file to use for certs, defaults to using a generated self-signed certificate
	EnvV2APort    = "GRPC_V2ALPHA_PORT"          // Port to listen on, default: 2500

	EnvHTTPEnabled = "HTTP_ENABLED"       // Controls whether to setup the grpc v2alpha auth service, default: false
	EnvHTTPTLS     = "HTTP_TLS"           // Whether to use TLS, default: false
	EnvHTTPTLSFile = "HTTP_TLS_CERT_FILE" // Custom TLS file to use for certs, defaults to using a generated self-signed certificate
	EnvHTTPPort    = "HTTP_PORT"          // Port to listen on, default: 8000
)

func main() {
	ctx := ctrl.SetupSignalHandler()
	g, gCtx := errgroup.WithContext(ctx)

	// Setup logging
	logLevel := getEnv(EnvLogLevel, "debug")
	var zapLevel zap.AtomicLevel
	switch logLevel {
	case "debug":
		zapLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapLevel = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		fmt.Fprintln(os.Stderr, "error: invalid use of LOG_LEVEL. expected info/debug/warn/info")
		zapLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	loggerConfig := zap.Config{
		Level:            zapLevel,
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, _ := loggerConfig.Build()
	defer logger.Sync()

	logger.Info("starting ambassador multiauth ext_authz service!")
	logger.Info("this service is intended to demo auth functionality and is not intended for production use...")
	logger.Info(fmt.Sprintf("log level set to %s", logLevel))

	// Setup grpc v3 auth service if enabled
	v3Enabled, err := strconv.ParseBool(getEnv(EnvV3Enabled, "true"))
	if err != nil {
		logger.Error("error: invalid use of GRPC_V3_ENABLED. expected true/false")
	}
	if v3Enabled {
		tls, err := strconv.ParseBool(getEnv(EnvV3TLS, "false"))
		if err != nil {
			logger.Error("error: invalid use of GRPC_V3_TLS. expected true/false")
		}
		port, err := strconv.Atoi(getEnv(EnvV3Port, "3000"))
		if err != nil {
			logger.Error("error: invalid use of GRPC_V3_PORT. expected integer that is a valid port value")
		}
		tlsFile := getEnv(EnvV3TLSFile, "")
		serviceInfo := fmt.Sprintf("127.0.0.1:%d, tls: %t", port, tls)
		if tlsFile != "" {
			serviceInfo = fmt.Sprintf("%s, custom tls file: %q", serviceInfo, tlsFile)
		}
		logger.Info("server configured to run with", zap.String("gRPC auth v3 protocol on", serviceInfo))

		// Start the gRPC ext_authz v3 server
		g.Go(func() error {
			logger.Info("starting ambassador grpc ext_authz v3 service!")
			err := server.NewGRPCAuthV3Server(logger, port, tls, tlsFile).Start(gCtx)
			if err != nil {
				err = fmt.Errorf("grpc ext_authz v3 server failed to start: %w", err)
			}
			return err
		})
	}

	err = g.Wait()
	if err != nil {
		logger.Error("error occured starting the ambassador multiauth service", zap.Error(err))
	}
	logger.Info("ambassador multiauth service shutting down...")
}

func getEnv(name, fallback string) string {
	res := os.Getenv(name)
	if res == "" {
		res = fallback
	}
	return res
}
