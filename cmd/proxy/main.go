package main

import (
	"fmt"
	"net/http"
	"path"
	"runtime"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/castai/cloud-proxy/internal/config"
	"github.com/castai/cloud-proxy/internal/gcpauth"
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

	conn, err := grpc.NewClient(cfg.GRPC.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Panicf("Failed to connect to server: %v", err)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logger.Panicf("Failed to close gRPC connection: %v", err)
		}
	}(conn)

	src := gcpauth.GCPCredentialsSource{}

	executor := proxy.NewExecutor(src, http.DefaultClient)
	client := proxy.NewClient(executor, logger)
	client.Run(conn)
}
