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

	"github.com/castai/cloud-proxy/internal/gcpauth"
	"github.com/castai/cloud-proxy/internal/proxy"
)

const (
	projectID   = "engineering-test-353509"
	location    = "europe-north1-a"
	testCluster = "lachezar-2708"
)

func main() {
	fmt.Println("Hi")

	runBasicTests()
	runProxyTest()
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
		fmt.Printf("Got successful response for cluster via proxy: %v: %v\n", getClusterResponse.Name, getClusterResponse.GetStatus())
		time.Sleep(10 * time.Second)
	}
}
