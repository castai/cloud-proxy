package main

import (
	"context"
	"fmt"
	"github.com/castai/cloud-proxy/internal/cloud/gcp"
	"github.com/castai/cloud-proxy/internal/gcpauth"
	"net/http"
	"path"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/castai/cloud-proxy/internal/config"
	"github.com/castai/cloud-proxy/internal/proxy"
	"github.com/sirupsen/logrus"
)

var (
	GitCommit = "undefined"
	GitRef    = "no-ref"
	Version   = "local"
)

func main() {
	cfg := config.Get()

	logger := logrus.New()
	logger.SetLevel(logrus.Level(cfg.Log.Level))
	logger.SetReportCaller(true)
	logger.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	}

	logger.WithFields(logrus.Fields{
		"GitCommit": GitCommit,
		"GitRef":    GitRef,
		"Version":   Version,
	}).Println("Starting cloud-proxy")

	dialOpts := make([]grpc.DialOption, 0)
	dialOpts = append(dialOpts, grpc.WithConnectParams(grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  2 * time.Second,
			Jitter:     0.1,
			MaxDelay:   5 * time.Second,
			Multiplier: 1.2,
		},
	}))
	if cfg.GRPC.TLS.Enabled {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.GRPC.Endpoint, dialOpts...)
	if err != nil {
		logger.Panicf("Failed to connect to server: %v", err)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logger.Panicf("Failed to close gRPC connection: %v", err)
		}
	}(conn)

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"authorization", fmt.Sprintf("Token %s", cfg.GRPC.Key),
	))

	src := gcpauth.GCPCredentialsSource{}

	executor := gcp.NewExecutor(src, http.DefaultClient)
	client := proxy.NewClient(executor, logger)
	client.Run(ctx, conn)
}
