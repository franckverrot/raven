# Raven

Raven is a control plane for Envoy Proxy that is basing its service discovery
on [Consul][consul].


Envoy Proxy supports a lot of different types of service discoveries:

  * `LDS`: Listener Discovery Service: what to listen to;
  * `CDS`: Cluster Discovery Service: what apps are available;
  * And a lot more, soon to be implemented.


## Usage


    HOST_ADDRESS=host.docker.internal make


This should spin up Raven.  Now you can start Envoy

    make envoy


It is strongly recommended to start Envoy interactively but it works either way.


### Environment variables


	* `GRPC_BINDING` : where raven will expose its gRPC interface (default: `127.0.0.1:1984`)
	* `CONSUL_URL` : where Consul is located (default: `localhost:8500`)
	* `HOST_ADDRESS` : address to use if the service address is `localhost` (useful with Docker for Mac)


## TODO

  * [ ] Implement remaining Service Discoveries
  * [ ] Implement `AggregatedServiceDiscovery`
  * [ ] Use flags along with environment variables



## License

Please see [LICENSE] for licensing details.



[consul]: https://www.consul.io