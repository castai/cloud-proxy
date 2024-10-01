package config

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TODO: test defaults.

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

	t.Setenv("LOG_LEVEL", "3")

	expected := Config{
		CastAI: CastAPI{
			APIKey:         "API_KEY",
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
		Log: Log{
			Level: 3,
		},
		UseCompression:   false,
		KeepAlive:        KeepAliveDefault,
		KeepAliveTimeout: KeepAliveTimeoutDefault,
		HealthAddress:    HealthAddressDefault,
	}

	got := Get()

	r.Equal(expected, got)
}
