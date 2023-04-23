// From DataWire Example service as base

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
	EnvLogLevel = "LOG_LEVEL"             // Controls logging verbosity, default: debug
	EnvTLS      = "GRPC_V3_TLS"           // Whether to use TLS, default: false
	EnvTLSFile  = "GRPC_V3_TLS_CERT_FILE" // Custom TLS file to use for certs, defaults to using a generated self-signed certificate
	EnvPort     = "GRPC_V3_PORT"          // Port to listen on, default: 3000
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

	// Check environment variables for custom settings
	tls, err := strconv.ParseBool(getEnv(EnvTLS, "false"))
	if err != nil {
		logger.Error("error: invalid use of GRPC_V3_TLS. expected true/false")
	}
	port, err := strconv.Atoi(getEnv(EnvPort, "3000"))
	if err != nil {
		logger.Error("error: invalid use of GRPC_V3_PORT. expected integer that is a valid port value")
	}
	tlsFile := getEnv(EnvTLSFile, "")

	logger.Info("starting ambassador ext_authz service!")
	logger.Info("this service is intended to demo auth functionality and is not intended for production use...")
	logger.Info(fmt.Sprintf("log level set to %s", logLevel))

	serviceInfo := fmt.Sprintf("127.0.0.1:%d, tls: %t", port, tls)
	if tlsFile != "" {
		serviceInfo = fmt.Sprintf("%s, custom tls file: %q", serviceInfo, tlsFile)
	}
	logger.Info("server configured to run with",
		zap.String("gRPC auth v3 protocol on", serviceInfo),
	)

	// Start the gRPC ext_authz v3 server on port 3000
	g.Go(func() error {
		err := server.NewGRPCAuthV3Server(logger, port, tls, tlsFile).Start(gCtx)
		if err != nil {
			err = fmt.Errorf("grpc ext_authz v3 server failed to start: %w", err)
		}
		return err
	})

	err = g.Wait()
	if err != nil {
		logger.Error("error occured starting ambassador grpc ext_authz v3 service", zap.Error(err))
	}

	logger.Info("ambassador grpc ext_authz v3 service has finished shutting down")
}

func getEnv(name, fallback string) string {
	res := os.Getenv(name)
	if res == "" {
		res = fallback
	}
	return res
}
