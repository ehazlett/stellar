![Radiant](radiant.jpg)

[Photo by Jakub Novacek on Pexels](https://www.pexels.com/photo/time-lapse-photo-of-stars-on-night-924824/)

# Radiant
Radiant is a GRPC proxy service using [Caddy](https://caddyserver.com).

Radiant uses a datastore to store server info.  By default, there is a simple in-memory datastore.  You can implement whatever you want to integrate with external systems.

# Building
Uses [dep](https://github.com/golang/dep) for dependencies.

- `make`

To just build binaries:

- `make binaries`

# Usage
To start the server, run:

```
$> radiant
```

Or via code:

```go
// create config
cfg := &radiant.Config{
	GRPCAddr:  "unix:///run/radiant.sock",
	HTTPPort:  80,
	HTTPSPort: 443,
	Debug:     true,
}
// instantiate a datastore
memDs := memory.NewMemory()

// create the server
srv, _ := server.NewServer(cfg, memDs)

// run the server
_ = srv.Run()
```

This will start both the proxy and GRPC servers.

There is a Go client available to assist in usage:

```go
client, _ := radiant.NewClient("unix:///run/radiant.sock")

timeout := time.Second * 30
upstreams := []string{
    "http://1.2.3.4",
    "http://5.6.7.8",
}
opts := []radiant.AddOpts{
    radiant.WithUpstreams(upstreams...),
    radiant.WithTimeouts(timeout),
    radiant.WithTLS,
}

// add the server
_ = client.AddServer(host, opts...)

// reload the proxy to take effect
_ = client.Reload()

// remove the server
_ = client.RemoveServer(host)

// reload to take effect
_ = client.Reload()
```
It is safe to reload as often as you wish.  There is zero downtime for the reload operation.

There is also a CLI that can be used directly or as reference.
