package localtest

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

// RunBasicTests verifies that:
//
//   - We can get a local auth token
//   - We can make a call to GKE GetCluster API successfully with that token
//
// Sample usage: verify that the workload's identity has access to GKE without any modifications; proxies; etc
func RunBasicTests(projectID, location, testCluster string) {
	src := gcpauth.GCPCredentialsSource{}
	log.Println(src)
	defaultCreds, err := src.GetDefaultCredentials()
	if err != nil {
		panic(err)
	}

	if defaultCreds.ProjectID != projectID {
		log.Panicf("expected project ID %q got %q", projectID, defaultCreds.ProjectID)
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
		log.Panicf("creating gke cluster client: %v", err)
	}

	getClusterResponse, err := gkeClient.GetCluster(context.Background(), &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%v/locations/%v/clusters/%v", projectID, location, testCluster),
	})
	if err != nil {
		log.Panicf("getting cluster: %v", err)
	}
	log.Printf("Got successful response for cluster: %v: %v\n", getClusterResponse.Name, getClusterResponse.GetStatus())
}

// RunProxyTest verifies that:
//
//   - Making a call to GKE without credentials fails with 401 (to ensure we don't auto-detect credentials somehow).
//   - Making the call via proxy.Executor works as expected (i.e. the modifications proxy.Executor does make the call succeed with proper auth)
//
// This func doesn't return; it loops forever
func RunProxyTest(projectID, location, testCluster string) {
	// Ensure that we get 401 when we don't use the "proxy"
	gkeClient, err := container.NewClusterManagerRESTClient(
		context.Background(),
		option.WithoutAuthentication())
	if err != nil {
		log.Panicf("creating gke cluster client: %v", err)
	}

	_, err = gkeClient.GetCluster(context.Background(), &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%v/locations/%v/clusters/%v", projectID, location, testCluster),
	})
	if err == nil {
		log.Panic("expected unauthorized error when not passing credentials")
	}

	// TODO: This feels ugly, isn't there a type-safe way?
	out := &apierror.APIError{}
	if ok := errors.As(err, &out); ok && out.Reason() == "CREDENTIALS_MISSING" {
		log.Println("Test successful; with no auth we get 401")
	} else {
		log.Panicf("expected error when not passing credentials: %v", err)
	}

	log.Println("Setting up custom GCP client with our roundtripper")

	src := gcpauth.GCPCredentialsSource{}
	executor := proxy.NewExecutor(src, http.DefaultClient)
	customRoundTripper := NewProxyRoundTripper(executor)
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
	log.Println("Simulating cast requests forever...")
	for {
		getClusterResponse, err := gkeProxiedClient.GetCluster(context.Background(), &containerpb.GetClusterRequest{
			Name: fmt.Sprintf("projects/%v/locations/%v/clusters/%v", projectID, location, testCluster),
		})
		if err != nil {
			log.Panicf("getting cluster through proxy: %v", err)
		}
		log.Printf("Got successful response for cluster via local proxy: %v: %v\n", getClusterResponse.Name, getClusterResponse.GetStatus())
		time.Sleep(10 * time.Second)
	}
}
