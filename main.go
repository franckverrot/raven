package main

import (
	server "github.com/franckverrot/raven/server"
	"github.com/franckverrot/raven/utils"
)

const (
	defaultGRPCAddress = "127.0.0.1:1984"
	defaultConsulURL   = "localhost:8500"
	defaultHostAddress = ""
)

func main() {
	gRPCAddress := utils.GetEnvWithDefault("GRPC_BINDING", defaultGRPCAddress)
	consulURL := utils.GetEnvWithDefault("CONSUL_URL", defaultConsulURL)
	hostAddress := utils.GetEnvWithDefault("HOST_ADDRESS", defaultHostAddress)

	s := server.NewServer(gRPCAddress, consulURL, hostAddress)

	s.Start()
}
