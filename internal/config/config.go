package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	GRPC GRPC
	Log  Log
}

type GRPC struct {
	Endpoint string
	Key      string
}

type Log struct {
	Level int
}

var cfg *Config = nil

func Get() Config {
	if cfg != nil {
		return *cfg
	}

	_ = viper.BindEnv("grpc.endpoint", "GRPC_ENDPOINT")
	_ = viper.BindEnv("grpc.key", "GRPC_KEY")
	_ = viper.BindEnv("log.level", "LOG_LEVEL")

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		panic(fmt.Errorf("while parsing config: %w", err))
	}

	if cfg.GRPC.Endpoint == "" {
		required("GRPC_ENDPOINT")
	}

	if cfg.GRPC.Key == "" {
		required("GRPC_KEY")
	}

	if cfg.Log.Level == 0 {
		cfg.Log.Level = int(logrus.InfoLevel)
	}

	return *cfg
}

func required(variable string) {
	panic(fmt.Errorf("variable %s is required", variable))
}
