package dummy

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
)

const (
	projectID   = "engineering-test-353509"
	location    = "europe-north1-a"
	testCluster = "lachezar-2708"
)

type mockEP struct {
	gkeClient *container.ClusterManagerClient
}

func newMockEP(dispatcher *Dispatcher) (*mockEP, error) {
	httpClient, _, err := htransport.NewClient(context.Background(), option.WithoutAuthentication())
	if err != nil {
		return nil, err
	}
	httpClient.Transport = NewHttpOverGrpcRoundTripper(dispatcher)
	gkeProxiedClient, err := container.NewClusterManagerRESTClient(
		context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	return &mockEP{gkeClient: gkeProxiedClient}, nil
}

func (m *mockEP) simulateActivity() {
	log.Println("Simulating External provisioner call forever...")
	for {
		getClusterResponse, err := m.gkeClient.GetCluster(context.Background(), &containerpb.GetClusterRequest{
			Name: fmt.Sprintf("projects/%v/locations/%v/clusters/%v", projectID, location, testCluster),
		})
		if err != nil {
			log.Panicf("%v\n", fmt.Errorf("getting cluster through proxy: %w", err))
		}
		fmt.Printf("Got successful response for cluster via proxy: %v: %v\n", getClusterResponse.Name, getClusterResponse.GetStatus())
		time.Sleep(time.Duration(rand.IntN(2)) * time.Second)
	}
}
