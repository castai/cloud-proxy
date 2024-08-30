package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/castai/cloud-proxy/internal/castai/proto"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc"
)

func TestBasic(ctx context.Context, server Server, rt http.RoundTripper) error {
	if err := server.StartServer(); err != nil {
		return err
	}
	defer server.StopServer()

	zonesClient, err := GetZonesClient(ctx, rt)
	if err != nil {
		return err
	}

	it := zonesClient.List(ctx, &computepb.ListZonesRequest{
		Project: "engineering-test-353509",
	})

	zones := make([]*computepb.Zone, 0)
	for {
		zone, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		zones = append(zones, zone)
	}

	if len(zones) < 1 {
		return fmt.Errorf("no zones")
	}

	return nil
}

func main() {
	ctx := context.Background()

	logger := log.New(os.Stderr, "[CLOUD-PROXY-E2E] ", log.LstdFlags)

	grpcServer := grpc.NewServer()
	setup := NewTestSetup(grpcServer, logger)
	proto.RegisterGCPProxyServerServer(grpcServer, setup)

	tt := map[string]func(ctx context.Context, server Server, rt http.RoundTripper) error{
		"basic test": TestBasic,
	}

	for name, testFunc := range tt {
		setup.ExecuteTest(ctx, name, testFunc)
	}

	if !setup.result {
		os.Exit(1)
	}
}
