Stellar is a simplified container platform.

# Design Goals

- Simple to deploy and use
- No master or leader elections
- Multihost networking
- Cluster and node based operations
- Extensible service system

Stellar was designed to be a simple container platform for environments that
do not have large orchestration requirements (IoT, edge, CI, render farms, batch computing, etc).

The following are the high level features it provides:

# Simplicity
Stellar is a single Go binary that is easy to deploy and maintain.  It contains everything
that is needed to run the system.

# Leaderless
Stellar was built as a loose distributed system.  Therefore, it does not fit all use
cases.  It is intended to be used where performance is crucial.  Stellar uses drastically
fewer resources than most other popular container platforms.  This fits well with use
cases such as IoT, edge, batch computing, etc where performance is key but high speed, low
latency networks are not always available.  Stellar tolerates peers coming and going by
maintaining minimal dependencies and simple data storage.

# Clustering
Stellar offers cluster capabilities where multiple Stellar nodes join together to form the platform.
This enables high availability among services as well as failover and fault tolerance.  However,
a core design principle is to allow all actions to be directed at a single node if desired.

# Extensible Services
Stellar uses [GRPC](http://grpc.io) for internal services.  This allows any of the core services
(health, node, cluster, networking, datastore) to be substituted with another implementation.  We
are also planning on a plugin type system in the near future where users can bring their own GRPC
services and have them available throughout the cluster.
