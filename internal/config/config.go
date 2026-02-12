package config

import (
	"errors"
	"os"
	"strconv"
)

type EnvConfig struct {
	TokenKey string
	GRPCPort int
	HTTPPort int
}

func LoadConfiguration() (*EnvConfig, error) {
	grpcPort, err1 := strconv.Atoi(os.Getenv("GRPC_PORT"))
	httpPort, err2 := strconv.Atoi(os.Getenv("HTTP_PORT"))
	if errs := errors.Join(err1, err2); errs != nil {
		return nil, errs
	}

	return &EnvConfig{
		TokenKey: os.Getenv("TOKEN_KEY"),
		GRPCPort: grpcPort,
		HTTPPort: httpPort,
	}, nil
}
