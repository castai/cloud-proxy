package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/container/apiv1/containerpb"
	"golang.org/x/sync/errgroup"
)

const projectID = "engineering-test-353509"

func TestParallelCalls(ctx context.Context, server Server, rt http.RoundTripper) error {
	if err := server.StartServer(); err != nil {
		return err
	}
	defer server.GracefulStopServer()

	eg := &errgroup.Group{}
	for i := 0; i < 3; i++ {
		eg.Go(func() error {
			contClient, err := GetContainerClient(ctx, rt)
			if err != nil {
				return err
			}

			_, err = contClient.ListClusters(ctx, &containerpb.ListClustersRequest{
				Parent: fmt.Sprintf("projects/%s/locations/-", projectID),
			})
			if err != nil {
				return err
			}

			return nil
		})
	}
	return eg.Wait()
}

func main() {
	ctx := context.Background()

	logger := log.New(os.Stderr, "[CLOUD-PROXY-E2E] ", log.LstdFlags)

	setup := NewTestSetup(logger)

	tt := map[string]func(ctx context.Context, server Server, rt http.RoundTripper) error{
		"basic test": TestParallelCalls,
	}

	for name, testFunc := range tt {
		setup.ExecuteTest(ctx, name, testFunc)
	}

	if !setup.result {
		os.Exit(1)
	}
}
