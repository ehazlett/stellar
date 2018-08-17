This is an overview of the Stellar architecture.

# Terminology
The following are definitions that are used:

- Cluster: A group of compute resources running Stellar
- Node: An individual instance of Stellar
- GRPC: A high-performance RPC framework
- Service: A service that provides operations to the cluster

Stellar is built on GRPC and a simple peer-to-peer [cluster service](https://github.com/ehazlett/element).
Each node registers with the other peers and publishes its GRPC services.
These services then provide various platform features.  The following
describes what services are included in the platform.

# Datastore
The datastore service provides the core database functionality for the cluster.
It is essentially a replicated [BoltDB](https://github.com/coreos/bbolt).  The
data is specifically designed to be as tolerant of failures as possible.  For
example, almost all of the data is namespaced by node to enable nodes to come
and go with reduced risk on overwrites.  For this reason, it is not recommended
to use this service as a general datastore.  It is built specifically for Stellar.

# Node
The node service provides container execution management and resource information
on container content (images, etc) along with node health.

# Health
The health service provides general health information on CPU, Memory, OS, uptime, etc
at the node level.

# Version
The version service provides basic information about the application version.

# Network
The network service provides subnet and IP allocation throughout the cluster.  See the
network service [README](https://github.com/ehazlett/stellar/blob/master/services/network/README.md) for a more detailed
explanation of how it is designed.  Simply put, the network service allocates
a dedicated subnet (in a `/22` by default) for each node.  IP allocation is performed
via a CNI IPAM plugin that reserves IPs in the specified subnet for the node it is
operating on.

# Nameserver
The nameserver service is a simple internal domain naming system.  This provides service
discovery for the cluster and integrates with the application service to automatically
provide naming.  The nameserver service also allows for custom records as needed
by the operator.  Since this is a separate service, the nameserver works regardless
of the CNI implementation.

# Cluster
The cluster service aggregates resource information from the nodes via GRPC.  This enables
a unified view in addition to simple application deployment.

# Application
The application service enables operators to deploy multi-container applications to the
cluster in an easy way.  Networking is automatically enabled as well as service
discovery via the nameserver service.
