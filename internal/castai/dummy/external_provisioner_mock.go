package dummy

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/castai/cloud-proxy/internal/e2etest"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
)

const (
	projectID   = "engineering-test-353509"
	location    = "europe-central2"
	testCluster = "damianc"
)

type mockEP struct {
	gkeClient *container.ClusterManagerClient
	logger    *log.Logger
}

func newMockEP(dispatcher *e2etest.Dispatcher, logger *log.Logger) (*mockEP, error) {
	httpClient, _, err := htransport.NewClient(context.Background(), option.WithoutAuthentication())
	if err != nil {
		return nil, err
	}
	httpClient.Transport = e2etest.NewHttpOverGrpcRoundTripper(dispatcher, logger)
	gkeProxiedClient, err := container.NewClusterManagerRESTClient(
		context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	return &mockEP{gkeClient: gkeProxiedClient, logger: logger}, nil
}

func (m *mockEP) simulateActivity() {
	m.logger.Println("Simulating External provisioner call forever...")
	for {
		getClusterResponse, err := m.gkeClient.GetCluster(context.Background(), &containerpb.GetClusterRequest{
			Name: fmt.Sprintf("projects/%v/locations/%v/clusters/%v", projectID, location, testCluster),
		})
		if err != nil {
			m.logger.Panicf("%v\n", fmt.Errorf("getting cluster through proxy: %w", err))
		}
		m.logger.Printf("Got successful response for cluster via proxy: %v: %v\n", getClusterResponse.Name, getClusterResponse.GetStatus())
		time.Sleep(time.Duration(rand.IntN(5)) * time.Second)
	}
}
