package config

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TODO: test defaults

func TestConfig(t *testing.T) {
	r := require.New(t)

	clusterID := uuid.New()

	t.Setenv("CAST_API_KEY", "API_KEY")
	t.Setenv("CAST_URL", "cast-url")
	t.Setenv("CAST_GRPC_URL", "cast-grpc-url")
	t.Setenv("CAST_DISABLE_GRPC_TLS", "true")

	t.Setenv("CLUSTER_ID", clusterID.String())

	t.Setenv("POD_NAMESPACE", "castai-namespace")
	t.Setenv("POD_IP", "192.168.0.1")
	t.Setenv("POD_NAME", "heavy-worker")
	t.Setenv("NODE_NAME", "awesome-node")

	expected := Config{
		CastAI: CastAPI{
			ApiKey:         "API_KEY",
			URL:            "cast-url",
			GrpcURL:        "cast-grpc-url",
			DisableGRPCTLS: true,
		},
		ClusterID: clusterID.String(),
		PodMetadata: PodMetadata{
			PodNamespace: "castai-namespace",
			PodIP:        "192.168.0.1",
			NodeName:     "awesome-node",
			PodName:      "heavy-worker",
		},
	}

	got := Get()

	r.Equal(expected, got)
}
