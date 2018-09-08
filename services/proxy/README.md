# Stellar Proxy Service
The Stellar Proxy Service provides layer 7 load balancing to applications in the cluster.
It uses the [Caddy](http://caddyserver.com) Server as the foundation and automatically
configures the proxy based upon endpoints specified in the application definition.  For example,
given the following configuration, the application will be available at the specified domain:

```
"endpoints": [
    {
        "service": "web",
        "protocol": "http",
        "host": "example.com",
        "port": 8080,
        "tls": false
    }
]

```

This should be specified in the application config.  Please note that the domain must be routable
and if you enable TLS it must be resolvable by DNS.
