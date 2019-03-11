package utils

import (
	"log"
	"os"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"

	"github.com/envoyproxy/go-control-plane/pkg/util"
)

// MessageToStruct wrapper
func MessageToStruct(m *hcm.HttpConnectionManager) *types.Struct {
	pbst, err := util.MessageToStruct(m)
	if err != nil {
		panic(err)
	}
	return pbst
}

// ClustersToAny converts the contens of a resourcer's Values to the
// respective slice of types.Any.
func ClustersToAny(clusters []*v2.Cluster) ([]types.Any, error) {
	typeUrl := "type.googleapis.com/envoy.api.v2.Cluster"
	resources := make([]types.Any, len(clusters))
	for i := range clusters {
		value := MarshalCluster(clusters[i])
		resources[i] = types.Any{TypeUrl: typeUrl, Value: value}
	}
	return resources, nil
}

// ListenersToAny converts the contens of a resourcer's Values to the
// respective slice of types.Any.
func ListenersToAny(listeners []*v2.Listener) ([]types.Any, error) {
	typeUrl := "type.googleapis.com/envoy.api.v2.Listener"
	resources := make([]types.Any, len(listeners))
	for i := range listeners {
		value := MarshalListener(listeners[i])
		resources[i] = types.Any{TypeUrl: typeUrl, Value: value}
	}
	return resources, nil
}

// MarshalCluster wraps default marshalling for Cluster structs
func MarshalCluster(l *v2.Cluster) []byte {
	m, err := proto.Marshal(l)

	if err != nil {
		log.Fatal(err)
	}
	return m
}

// MarshalListener wraps default marshalling for Listener structs
func MarshalListener(l *v2.Listener) []byte {
	m, err := proto.Marshal(l)

	if err != nil {
		log.Fatal(err)
	}
	return m
}

func GetEnvWithDefault(key string, defaultValue string) string {
	if value, found := os.LookupEnv(key); found {
		return value
	}
	return defaultValue
}
