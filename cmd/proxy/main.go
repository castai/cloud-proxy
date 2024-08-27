package main

import (
	"context"
	"fmt"
	"log"

	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"google.golang.org/api/option"

	"github.com/castai/cloud-proxy/internal/gcpauth"
)

func main() {
	fmt.Println("Hi")

	src := gcpauth.GCPCredentialsSource{}
	fmt.Println(src)

	log.Println("new iteration")
	log.Println("TESTING REBUILD 2")

	defaultCreds, err := src.GetDefaultCredentials()
	if err != nil {
		panic(err)
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
	gkeClient, err := container.NewClusterManagerClient(
		context.Background(),
		option.WithCredentials(defaultCreds))
	if err != nil {
		panic(fmt.Errorf("creating gke cluster client: %w", err))
	}

	getClusterResponse, err := gkeClient.GetCluster(context.Background(), &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%v/locations/%v/clusters/%v", defaultCreds.ProjectID, "europe-north1-a", "lachezar-2708"),
	})
	if err != nil {
		panic(fmt.Errorf("getting cluster: %w", err))
	}
	fmt.Printf("Got response for cluster: %+v\n", getClusterResponse)

	fmt.Println("Sleeping forever...")
	for {
	}
}
