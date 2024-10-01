package main

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"

	"cloud-proxy/internal/cloud/gcp"
	"cloud-proxy/internal/cloud/gcp/gcpauth"
	"cloud-proxy/internal/config"
	"cloud-proxy/internal/healthz"
	"cloud-proxy/internal/proxy"
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

	ctx := context.Background()

	tokenSource, err := gcpauth.NewTokenSource(ctx)
	if err != nil {
		logger.WithError(err).Panicf("Failed to create GCP credentials source")
	}

	client := proxy.New(gcp.New(tokenSource), logger,
		GetVersion(), &cfg)

	go startHealthServer(logger, cfg.HealthAddress)

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
		CallerPrettyfier: func(f *runtime.Frame) (function, file string) {
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

func startHealthServer(logger *logrus.Logger, addr string) {
	healthchecks := healthz.NewServer(logger)

	logger.Infof("Starting healthcheck server on address %v", addr)

	if err := healthchecks.Run(addr); err != nil {
		logger.WithError(err).Errorf("Failed to run healthcheck server")
	}
}
