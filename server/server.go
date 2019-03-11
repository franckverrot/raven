package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hashicorp/consul/api"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"google.golang.org/grpc"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/envoyproxy/go-control-plane/pkg/util"
	"github.com/franckverrot/raven/utils"
	"github.com/franckverrot/raven/xds"
)

// Server is the gRPC server implementation
type Server struct {
	GRPCAddress  string
	ConsulURL    string
	Routes       map[string]*RavenRoute
	ListenerChan chan int
	ClusterChan  chan int
	HostAddress  string
}

type RavenRoute struct {
	Domains map[string]bool // Used like a set
	Hosts   map[string]RavenHost
}

type RavenHost struct {
	Address string
	Port    int
}

func NewServer(gRPCAddress string, consulURL string, hostAddress string) *Server {
	newServer := new(Server)
	newServer.GRPCAddress = gRPCAddress
	newServer.ConsulURL = consulURL
	newServer.Routes = map[string]*RavenRoute{}
	newServer.ListenerChan = make(chan int, 1)
	newServer.ClusterChan = make(chan int, 1)
	newServer.HostAddress = hostAddress
	return newServer
}

// Start will setup the gRPC server
func (s *Server) Start() {
	s.startConsulConsumer(s.ConsulURL)
	s.setupGRPC(s.GRPCAddress)
}

func (s *Server) startConsulConsumer(consulURL string) {
	config := api.DefaultConfig()
	client, _ := api.NewClient(config)
	agent := client.Agent()

	ticker := time.NewTicker(time.Millisecond * 1000)
	go func() {
		for _ = range ticker.C {
			log.Printf("Checking consul\n")
			services, _ := agent.Services()

			for _, service := range services {
				serviceName := service.Service
				serviceAddress := service.Address
				servicePort := service.Port

				if _, ok := s.Routes[serviceName]; !ok {
					route := new(RavenRoute)
					route.Hosts = map[string]RavenHost{}
					route.Domains = map[string]bool{}
					s.Routes[serviceName] = route
				}

				route := *s.Routes[serviceName]

				route.Domains[serviceName] = true
				hostName := fmt.Sprintf("%s-%d", serviceAddress, servicePort)
				route.Hosts[hostName] = RavenHost{
					Address: s.determineServiceAddress(serviceAddress),
					Port:    servicePort,
				}
			}

			// Don't notify if stream hasn't started
			if len(s.ListenerChan) < 1 {
				s.ListenerChan <- 1
			}
			if len(s.ClusterChan) < 1 {
				s.ClusterChan <- 1
			}
		}
	}()
}

func (s *Server) setupGRPC(address string) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error getting listener: %+v", err)
	}

	grpc := grpc.NewServer()

	v2.RegisterClusterDiscoveryServiceServer(grpc, s)
	v2.RegisterListenerDiscoveryServiceServer(grpc, s)

	log.Printf("gRPC Listening on: %s", address)
	grpc.Serve(lis)
}

// FetchListeners isn't implemented
func (s *Server) FetchListeners(ctx context.Context, req *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	log.Println("FetchListeners not implemented")
	return nil, nil
}

func (s *Server) newHost(name string, domains []string, clusterName string) route.VirtualHost {
	return route.VirtualHost{
		Name:    name,
		Domains: domains,
		Routes: []route.Route{
			route.Route{
				Match: route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: clusterName,
						},
					},
				},
			},
		},
	}
}

func getDomains(domainsMap map[string]bool) []string {
	domains := make([]string, 0)
	for domain := range domainsMap {
		domains = append(domains, domain)
	}
	return domains
}

func (s *Server) getVirtualHosts() []route.VirtualHost {
	hosts := make([]route.VirtualHost, 0)

	for hostName, routeDef := range s.Routes {
		domains := getDomains(routeDef.Domains)
		hosts = append(hosts, s.newHost(hostName, domains, hostName))
	}

	return hosts
}

// StreamListeners allows streaming the configuration
func (s *Server) StreamListeners(stream v2.ListenerDiscoveryService_StreamListenersServer) error {
	log.Println("StreamListeners")

	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.AUTO,
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &v2.RouteConfiguration{
				Name:         "default-route",
				VirtualHosts: s.getVirtualHosts(),
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: util.Router,
		}},
	}
	listenerStruct := &v2.Listener{
		Name: "main-listener",
		Address: core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Address:  "0.0.0.0",
					Protocol: core.TCP,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: 10000,
					},
				},
			},
		},
		FilterChains: []listener.FilterChain{
			listener.FilterChain{
				Filters: []listener.Filter{
					listener.Filter{
						Name:   "envoy.http_connection_manager",
						Config: utils.MessageToStruct(manager),
					},
				},
			},
		},
	}

	for {
		select {
		case _ = <-s.ListenerChan:
			_, err := stream.Recv()

			res := xds.CreateListenerDiscoveryResponse(
				[]*v2.Listener{
					listenerStruct,
				},
			)
			stream.Send(res)

			log.Printf("Sync'ed listener %s\n", listenerStruct.Name)

			if err != nil {
				return err
			}
		}
	}
}

func (s *Server) FetchClusters(context.Context, *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	return nil, nil
}

func (s *Server) StreamClusters(stream v2.ClusterDiscoveryService_StreamClustersServer) error {
	for {
		select {
		case _ = <-s.ClusterChan:
			_, err := stream.Recv()

			clusters := make([]*v2.Cluster, 0)

			for clusterName, routeDef := range s.Routes {
				hosts := make([]RavenHost, 0)
				for _, host := range routeDef.Hosts {
					hosts = append(hosts, host)
				}

				newCluster := s.newCluster(clusterName, hosts)
				clusters = append(clusters, newCluster)
			}

			res := xds.CreateClusterDiscoveryResponse(clusters)
			stream.Send(res)

			for clusterName := range s.Routes {
				log.Printf("Sync'ed cluster %s\n", clusterName)
			}

			if err != nil {
				return err
			}
		}
	}
}

func (s *Server) newCluster(name string, hosts []RavenHost) *v2.Cluster {
	addresses := make([]*core.Address, 0)

	for _, host := range hosts {
		addresses = append(addresses,
			&core.Address{
				Address: &core.Address_SocketAddress{
					SocketAddress: &core.SocketAddress{
						Address:  host.Address,
						Protocol: core.TCP,
						PortSpecifier: &core.SocketAddress_PortValue{
							PortValue: uint32(host.Port),
						},
					},
				},
			},
		)
	}

	return &v2.Cluster{
		Name:            name,
		ConnectTimeout:  2 * time.Second,
		Type:            v2.Cluster_STRICT_DNS,
		DnsLookupFamily: v2.Cluster_V4_ONLY,
		LbPolicy:        v2.Cluster_ROUND_ROBIN,
		Hosts:           addresses,
	}
}

// determineServiceAddress is here mostly for development purpose as Docker For
// Mac doesn't support the `host` network mode and `127.0.0.1` would represent
// the container's loopback address
func (s *Server) determineServiceAddress(address string) string {
	realAddress := address

	if s.HostAddress != "" {
		realAddress = s.HostAddress
	}

	return realAddress
}
