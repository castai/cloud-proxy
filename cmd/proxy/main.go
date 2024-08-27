package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/googleapis/gax-go/v2/apierror"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/castai/cloud-proxy/internal/castai/dummy"
	"github.com/castai/cloud-proxy/internal/gcpauth"
	"github.com/castai/cloud-proxy/internal/proxy"
)

const (
	projectID   = "engineering-test-353509"
	location    = "europe-north1-a"
	testCluster = "lachezar-2708"
)

func main() {
	//runBasicTests()
	//go runProxyTest()

	go func() {
		log.Println("Starting mock cast instance")
		mockCast := &dummy.MockCast{}
		if err := mockCast.Run(); err != nil {
			log.Panicln("Error running mock Cast:", err)
		}
	}()

	go func() {
		log.Println("Starting proxy client")
		conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect to server: %v", err)
		}
		defer func(conn *grpc.ClientConn) {
			err := conn.Close()
			if err != nil {
				log.Fatalf("Failed to close gRPC connection: %v", err)
			}
		}(conn)

		src := gcpauth.GCPCredentialsSource{}
		executor := proxy.NewExecutor(src, http.DefaultClient)
		client := proxy.NewClient(executor)
		client.Run(conn)
	}()

	time.Sleep(1 * time.Hour)

	// Start local proxy

	// Option 1 - connect to mothership
	// Option 2 - use local server that "simulates" mothership
}

func runBasicTests() {
	src := gcpauth.GCPCredentialsSource{}
	fmt.Println(src)
	defaultCreds, err := src.GetDefaultCredentials()
	if err != nil {
		panic(err)
	}

	if defaultCreds.ProjectID != projectID {
		panic(fmt.Errorf("expected project ID %q got %q", projectID, defaultCreds.ProjectID))
	}

	log.Printf("Found default credentials: [project: %v, typeof(tokensrc): %T, tokensrc: %+v, jsonCreds: %v\n",
		defaultCreds.ProjectID, defaultCreds.TokenSource, defaultCreds.TokenSource, string(defaultCreds.JSON))

	token, err := defaultCreds.TokenSource.Token()
	if err != nil {
		panic(err)
	}
	log.Printf("Managed to successfully get a token: %+v\n", token)
	log.Printf("token expiration: %+v\n", token.Expiry)

	// Test GKE method
	gkeClient, err := container.NewClusterManagerRESTClient(
		context.Background(),
		option.WithCredentials(defaultCreds))
	if err != nil {
		panic(fmt.Errorf("creating gke cluster client: %w", err))
	}

	getClusterResponse, err := gkeClient.GetCluster(context.Background(), &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%v/locations/%v/clusters/%v", projectID, location, testCluster),
	})
	if err != nil {
		panic(fmt.Errorf("getting cluster: %w", err))
	}
	fmt.Printf("Got successful response for cluster: %v: %v\n", getClusterResponse.Name, getClusterResponse.GetStatus())

}

func runProxyTest() {
	// Ensure that we get 401 when we don't use the "proxy"
	gkeClient, err := container.NewClusterManagerRESTClient(
		context.Background(),
		option.WithoutAuthentication())
	if err != nil {
		panic(fmt.Errorf("creating gke cluster client: %w", err))
	}

	_, err = gkeClient.GetCluster(context.Background(), &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%v/locations/%v/clusters/%v", projectID, location, testCluster),
	})
	if err == nil {
		panic(fmt.Errorf("expected unauthorized error when not passing credentials"))
	}

	// TODO: This feels ugly, isn't there a type-safe way?
	out := &apierror.APIError{}
	if ok := errors.As(err, &out); ok && out.Reason() == "CREDENTIALS_MISSING" {
		fmt.Println("Test successful; with no auth we get 401")
	} else {
		panic(fmt.Errorf("expected error when not passing credentials: %w", err))
	}

	fmt.Println("Setting up custom GCP client with our roundtripper")

	src := gcpauth.GCPCredentialsSource{}
	executor := proxy.NewExecutor(src, http.DefaultClient)
	customRoundTripper := proxy.NewProxyRoundTripper(executor)
	customClientForGCP, _, err := htransport.NewClient(context.Background())
	if err != nil {
		log.Panicf("failed creating http client: %v", err)
	}
	customClientForGCP.Transport = customRoundTripper
	gkeProxiedClient, err := container.NewClusterManagerRESTClient(
		context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(customClientForGCP))
	if err != nil {
		log.Panicf("failed creating gke cluster client with proxy: %v", err)
	}
	fmt.Println("Simulating cast requests forever...")
	for {
		getClusterResponse, err := gkeProxiedClient.GetCluster(context.Background(), &containerpb.GetClusterRequest{
			Name: fmt.Sprintf("projects/%v/locations/%v/clusters/%v", projectID, location, testCluster),
		})
		if err != nil {
			panic(fmt.Errorf("getting cluster through proxy: %w", err))
		}
		fmt.Printf("Got successful response for cluster via local proxy: %v: %v\n", getClusterResponse.Name, getClusterResponse.GetStatus())
		time.Sleep(10 * time.Second)
	}
}
