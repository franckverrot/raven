package xds

import (
	"log"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/franckverrot/raven/utils"
)

func CreateClusterDiscoveryResponse(clusters []*v2.Cluster) *v2.DiscoveryResponse {
	anys, err := utils.ClustersToAny(clusters)
	if err != nil {
		log.Fatal(err)
	}

	return &v2.DiscoveryResponse{
		VersionInfo: "number-2",
		Resources:   anys,
		TypeUrl:     "type.googleapis.com/envoy.api.v2.Cluster",
	}
}
