# Stellar
Simplified Container Runtime Cluster

Stellar is designed to provide simple container runtime clustering.  One
or more nodes are joined together to create a cluster.  The cluster
is eventually consistent making it ideal for transient workloads or edge
computing where nodes are not always guaranteed to have high bandwidth, low
latency connectivity.

# Building
In order to build Stellar you will need the following:

- A [working Go environment](https://golang.org/doc/code.html)
- Protoc 3.x compiler and headers (download at the [Google releases page](https://github.com/google/protobuf/releases))

Once you have the requirements you can build.

If you change / update the protobuf definitions you will need to generate:

`make generate`

To build the binaries (client and server) run:

`make binaries`

## Docker
Alternatively you can use [Docker](https://www.docker.com) to build:

To generate protobuf:

`make docker-generate`

To build binaries:

`make docker-build`
