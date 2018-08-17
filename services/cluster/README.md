# Stellar Cluster Service
The cluster service provides a unified view of the cluster.  It aggregates info and
resources across all nodes by querying each node via GRPC.

# Nodes
The cluster service will report information on all nodes in the cluster:

```
NAME                ADDR                OS                          UPTIME              CPUS                MEMORY (USED)
ctr-00              10.0.1.33:9000      Linux (4.13.0-46-generic)   30 minutes          4                   955 MB / 1.0 GB
ctr-01              10.0.1.34:9000      Linux (4.13.0-46-generic)   30 minutes          4                   939 MB / 1.0 GB
ctr-02              10.0.1.35:9000      Linux (4.13.0-46-generic)   30 minutes          4                   942 MB / 1.0 GB

```

# Containers
The cluster service will report all containers throughout the cluster:

```
ID                  IMAGE                                RUNTIME                          NODE
ci.gocd             docker.io/gocd/gocd-server:v18.7.0   io.containerd.runtime.v1.linux   ctr-00
test00.test         docker.io/ehazlett/redis:alpine      io.containerd.runtime.v1.linux   ctr-00
test01.test         docker.io/ehazlett/redis:alpine      io.containerd.runtime.v1.linux   ctr-01

```
