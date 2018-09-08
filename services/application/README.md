# Stellar Application Service
The Stellar application service provides a simple way of launching container workloads.

# Format
Stellar applications are represented in JSON format for portability and interoperatbility.

```
{
    "name": "example-application",
    "labels": [
        "env=prod",
        "region=us-east"
    ],
    "services": [
        {
            "name": "redis",
            "image": "docker.io/library/redis:alpine",
            "process": {
                "uid": 0,
                "gid": 0,
                "args": ["redis-server"]
            },
            "labels": [
                "env=prod"
            ],
            "network": true,
            "mounts": [
                {
                    "type": "bind",
                    "source": "/mnt/data",
                    "destination": "/data",
                    "options": ["rbind"]
                }
            ]
        },
        {
            "name": "processor",
            "image": "docker.io/ehazlett/processor:alpine",
            "process": {
                "args": ["--debug"],
                "env": ["PROCS=4"]
            },
            "resources": {
                "cpu": 0.1,
                "memory": 128
            },
            "network": true
        },
        {
            "name": "ui",
            "image": "docker.io/ehazlett/ui:alpine",
            "network": true,
            "instances": 4,
	    "endpoints": [
                {
                    "service": "ui",
                    "protocol": "http",
                    "host": "example.com",
                    "port": 8080,
                    "tls": false
                }
            ]

        }
    ]
}
```
