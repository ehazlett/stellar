```
   _____ __       ____
  / ___// /____  / / /___ ______
  \__ \/ __/ _ \/ / / __ `/ ___/
 ___/ / /_/  __/ / / /_/ / /
/____/\__/\___/_/_/\__,_/_/

```

Simplified Container Runtime Cluster

Stellar is designed to provide simple container runtime clustering.  One
or more nodes are joined together to create a cluster.  The cluster
is eventually consistent making it ideal for transient workloads or edge
computing where nodes are not always guaranteed to have high bandwidth, low
latency connectivity.

# Why
There are several container platforms and container orchestrators out there.
However, they are too complex for my use.  I like simple infrastructure that
is easy to deploy and manage, tolerates failure cases and is easy to debug
when needed.  With the increased tolerance in failure modes, this comes at
a consistency cost.  It may not be for you.  Use the best tool for your use
case.  Enjoy :)

# Features

- Container execution via [containerd](https://github.com/containerd/containerd)
- Multihost Networking via [CNI](https://github.com/containernetworking/cni)
- Single binary for simple daemon deployment
- Masterless design
- Builtin Proxy using [Caddy](https://caddyserver.com) (zero downtime reloads, health checks, automatic HTTPS)
- Efficient use of system resources

# Building
In order to build Stellar you will need the following:

- A [working Go environment](https://golang.org/doc/code.html)
- Protoc 3.x compiler and headers (download at the [Google releases page](https://github.com/google/protobuf/releases))

Once you have the requirements you can build.

If you change / update the protobuf definitions you will need to generate:

`make generate`

To build the binaries (client and server) run:

`make`

## Docker
Alternatively you can use [Docker](https://www.docker.com) to build:

To generate protobuf:

`make docker-generate`

To build binaries:

`make docker-build`

# Running
To run Stellar, once you have a working containerd installation follow these steps:

- Build Stellar (all binaries will be in `/bin`)
- Copy `/bin/stellar` to `/usr/local/bin/`
- Copy `/bin/stellar-cni-ipam` to `/opt/containerd/bin/` or `/opt/cni/bin`

To start the initial node run `stellar -D server --bind-addr <node-ip> --advertise-addr <node-ip>`.
For example:

```
$> stellar -D server --bind-addr 10.0.1.33 --advertise-addr 10.0.1.33
```

To join additional nodes simply add the `--peer` flag.  For example:

```
$> stellar -D server --bind-addr 10.0.1.34 --advertise-addr 10.0.1.34 --peer 10.0.1.33:7946
```

You will now have a two node cluster.  To see node information, use `sctl`.

```
$> sctl --addr 10.0.1.33:9000 cluster nodes
NAME                ADDR                OS                          UPTIME              CPUS                MEMORY (USED)
ctr-00              10.0.1.33:9000      Linux (4.13.0-46-generic)   2 minutes           4                   952 MB / 1.0 GB
ctr-01              10.0.1.34:9000      Linux (4.13.0-46-generic)   13 seconds          4                   943 MB / 1.0 GB
```

# Deploying an Application
To deploy an application, create an application config.  For example, create the following as `example.conf`:

```json
{
    "name": "example",
    "labels": [
        "env=prod",
        "region=us-east"
    ],
    "services": [
        {
            "name": "redis",
            "image": "docker.io/library/redis:alpine",
            "runtime": "io.containerd.runtime.v1.linux",
            "process": {
                "uid": 0,
                "gid": 0,
                "args": ["redis-server"]
            },
            "labels": [
                "env=prod"
            ],
            "network": true
        }
    ]
}

```

Then run the following to deploy:

```
$> sctl --addr 10.0.1.33:9000 apps create -f ./example.conf
```

You should now see the application deployed:

```
$> sctl --addr 10.0.1.33:9000 apps list
NAME                SERVICES
example             1

$> sctl --addr 10.0.1.33:9000 apps inspect example
Name: example

Services:
  Name: example.redis
  Image: docker.io/library/redis:alpine
  Runtime: io.containerd.runtime.v1.linux
  Snapshotter: overlayfs
  Labels:
    containerd.io/restart.status=running
    stellar.io/application=example
    stellar.io/network=true

```

By default all applications that have networking enabled will have a corresponding nameserver record
created.  To view the records use the following:

```
$> sctl --addr 10.0.1.33:9000 nameserver list
NAME                    TYPE                VALUE               OPTIONS
example.redis.stellar   A                   172.16.0.4          nil
```
