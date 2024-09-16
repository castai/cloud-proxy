package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	CastAI      CastAPI     `mapstructure:"cast"`
	ClusterID   string      `mapstructure:"clusterid"`
	PodMetadata PodMetadata `mapstructure:"podmetadata"`

	//MetricsAddress  string               `mapstructure:"metricsaddress"`
	//HealthAddress   string               `mapstructure:"healthaddress"`
	Log Log `mapstructure:"log"`
}

// PodMetadata stores metadata for the pod, mostly used for logging and debugging purposes.
type PodMetadata struct {
	// PodNamespace is the namespace this pod is running in.
	PodNamespace string `mapstructure:"podnamespace"`
	// PodIP is the IP assigned to the pod.
	PodIP string `mapstructure:"podip"`
	// PodName is the name of the pod.
	PodName string `mapstructure:"podname"`
	// NodeName is the name of the node where this pod is running on.
	NodeName string `mapstructure:"nodename"`
}

// CastAPI contains the configuration for the connection to CAST AI API.
type CastAPI struct {
	// ApiKey is the API key used to authenticate to CAST AI API.
	ApiKey string `mapstructure:"apikey"`
	// URL is the URL of CAST AI REST API.
	URL string `mapstructure:"url"`
	// GrpcURL is the URL of CAST AI gRPC API.
	GrpcURL string `mapstructure:"grpcurl"`
	// DisableGRPCTLS disables TLS for gRPC connection. Should only be used for testing.
	DisableGRPCTLS bool `mapstructure:"disablegrpctls"`
}

type GCP struct {
	CredentialsJSON   string
	UseMetadataServer bool
}

type Log struct {
	Level int `mapstructure:"level"`
}

var cfg *Config = nil

func Get() Config {
	if cfg != nil {
		return *cfg
	}

	v := viper.New()

	v.MustBindEnv("cast.apikey", "CAST_API_KEY")
	v.MustBindEnv("cast.url", "CAST_URL")
	v.MustBindEnv("cast.grpcurl", "CAST_GRPC_URL")
	v.MustBindEnv("cast.disablegrpctls", "CAST_DISABLE_GRPC_TLS")

	v.MustBindEnv("clusterid", "CLUSTER_ID")

	v.MustBindEnv("podmetadata.podnamespace", "POD_NAMESPACE")
	v.MustBindEnv("podmetadata.podip", "POD_IP")
	v.MustBindEnv("podmetadata.nodename", "NODE_NAME")
	v.MustBindEnv("podmetadata.podname", "POD_NAME")

	_ = v.BindEnv("log.level", "LOG_LEVEL")

	cfg = &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		panic(fmt.Errorf("while parsing config: %w", err))
	}

	if cfg.CastAI.ApiKey == "" {
		required("CAST_API_KEY")
	}
	if cfg.CastAI.GrpcURL == "" {
		required("CAST_GRPC_URL")
	}
	if cfg.CastAI.URL == "" {
		required("CAST_URL")
	}
	if cfg.ClusterID == "" {
		required("CLUSTER_ID")
	}

	if cfg.Log.Level == 0 {
		cfg.Log.Level = int(logrus.InfoLevel)
	}

	return *cfg
}

func required(variable string) {
	panic(fmt.Errorf("variable %s is required", variable))
}
