# Stellar Scheduler Service
The Stellar scheduler service performs simple scheduling.  It handles node placement for Stellar services.
It does not create or monitor services for failures it simply schedules.  Given a service,
the scheduler will check for a `PlacementPreference`.  If there is no preference, then the first
available node is returned.  If there is a placement preference the schedule will handle accordingly.
The schedule will filter available nodes by node ID and node labels.  The preference is additive so it will
return nodes that match either by ID or by a matching labels.  The labels are "AND" so if multiple labels are
specified when creating the service, only nodes that match all labels will be returned.  The number of nodes
returned are determined by the `Replicas` config option.  Note: if `0` is set for replicas, a warning will
be issued in the logs and the replica count will be adjusted to `1`.  If you do not want the service to have
any replicas, remove it from the config.

# Examples
Here are some examples of using the scheduler:

## No Placement Preference

```
{
    "name": "demo",
    "services": [
        {
            "name": "app",
            "image": "docker.io/ehazlett/docker-demo:latest"
        }
    ]
}
```
This will be scheduled to the first available node with a single replica.

## Placement with Node ID

```
{
    "name": "demo",
    "services": [
        {
            "name": "app",
            "image": "docker.io/ehazlett/docker-demo:latest",
	    "placement_preference": {
	        "node_ids": ["node-00"]
	    },
	    "replicas": 3
        }
    ]
}
```
This will have 3 replicas deployed to `node-00`.

## Placement with Node Labels

```
{
    "name": "demo",
    "services": [
        {
            "name": "app",
            "image": "docker.io/ehazlett/docker-demo:latest",
	    "placement_preference": {
                "labels": {
                    "env": "staging"
		}
	    },
	    "replicas": 3
        }
    ]
}
```
This will have 3 replicas deployed to any node that has the label "env=staging".

Note: node labels can be configured in the Stellar config file.


For more examples of scheduling you can look at the tests in the scheduler service.
