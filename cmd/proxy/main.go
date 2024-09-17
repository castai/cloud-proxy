package main

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"runtime"
	"time"

	"cloud-proxy/internal/cloud/gcp"
	"cloud-proxy/internal/cloud/gcp/gcpauth"
	"cloud-proxy/internal/config"
	"cloud-proxy/internal/healthz"
	"cloud-proxy/internal/proxy"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

var (
	GitCommit = "undefined"
	GitRef    = "no-ref"
	Version   = "local"
)

func main() {
	logrus.Info("Starting proxy")
	cfg := config.Get()
	logger := setupLogger(cfg)

	dialOpts := make([]grpc.DialOption, 0)
	if cfg.CastAI.DisableGRPCTLS {
		// ONLY For testing purposes
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	}

	connectParams := grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  2 * time.Second,
			Jitter:     0.1,
			MaxDelay:   5 * time.Second,
			Multiplier: 1.2,
		},
	}
	dialOpts = append(dialOpts, grpc.WithConnectParams(connectParams))
	dialOpts = append(dialOpts, grpc.WithKeepaliveParams(keepalive.ClientParameters{}))

	logger.Infof(
		"Creating grpc channel against (%s) with connection config (%+v) and TLS enabled=%v",
		cfg.CastAI.GrpcURL,
		connectParams,
		!cfg.CastAI.DisableGRPCTLS,
	)
	conn, err := grpc.NewClient(cfg.CastAI.GrpcURL, dialOpts...)
	if err != nil {
		logger.Panicf("Failed to connect to server: %v", err)
		panic(err)
	}

	defer func(conn *grpc.ClientConn) {
		logger.Info("Closing grpc connection")
		err := conn.Close()
		if err != nil {
			logger.Panicf("Failed to close gRPC connection: %v", err)
			panic(err)
		}
	}(conn)

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"authorization", fmt.Sprintf("Token %s", cfg.CastAI.ApiKey),
	))

	client := proxy.New(conn, gcp.New(gcpauth.NewCredentialsSource(), http.DefaultClient), logger,
		cfg.ClusterID, GetVersion(), cfg.KeepAlive, cfg.KeepAliveTimeout)

	go func() {
		healthchecks := healthz.NewServer(logger, client)

		logger.Infof("Starting healthcheck server on address %v", cfg.HealthAddress)

		if err := healthchecks.Run(cfg.HealthAddress); err != nil {
			logger.WithError(err).Errorf("Failed to run healthcheck server")
		}
	}()

	err = client.Run(ctx)
	if err != nil {
		logger.Panicf("Failed to run client: %v", err)
		panic(err)
	}
}

func GetVersion() string {
	return fmt.Sprintf("GitCommit=%q GitRef=%q Version=%q", GitCommit, GitRef, Version)
}

func setupLogger(cfg config.Config) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.Level(cfg.Log.Level))
	logger.SetReportCaller(true)
	logger.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	}

	logger.WithFields(logrus.Fields{
		"GitCommit": GitCommit,
		"GitRef":    GitRef,
		"Version":   Version,
	}).Infof("Starting cloud-proxy: %+v", cfg)

	return logger
}
