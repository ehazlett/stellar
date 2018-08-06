# Stellar Network Service

The network service is a service that provides cross host networking for containers.  Static routing is used for simplicity and easier debugging.  Static networking was chosen for the following reasons:

- Tunnels add overhead and need to be tuned to improve performance
- Static routing offers native networking speed
- Easier debugging
- Access control using existing tools

In order to maintain simplicity and stable management the network will be divided into subnets.  The size of the network will determine the capacity the cluster can scale.  For example, a 192.168.0.0/24 would be able to have a total of 255 container IP addresses.  Since the divided subnets are used to route, each cluster node will have to have a unique subnet.  By default, Stellar uses the 172.16.0.0-172.31.255.255 address space.  This allows for a 1024 node cluster with 1022 routable containers per host (2 IPs per subnet are for gateway and broadcast).  To scale larger you can use the 10.0.0.0/8 subnet.  This can grow much larger (a 4096 node cluster could have 4096 containers per node).

For easier management, upon node join it will be assigned a subnet from the global network.  Any container that needs networking will be placed in this subnet.  This also means easier IPAM as a node will have a dedicated subnet.  For example, given a three node cluster the following subnets would be assigned:

- 172.16.0.0/22 (node0)
- 172.16.4.0/22 (node1)
- 172.16.8.0/22 (node2)

This format would allow the cluster to have 1,046,528 routable containers on the network (1024 nodes * 1022 container IPs).  The Stellar network service also propagates subnet routes throughout the cluster.

## Container Networking
Stellar manages the networking lifecycle of the container.  For containers that want to use Stellar networking, a label is required.  If the container is deployed via the Stellar API the label will be added automatically.  This allows external systems to create containers but still have networking.

Stellar manages the container ethernet device as follows:

- A veth pair is created (host + container)
- Each device in the veth pair is added to the local Stellar bridge
- The container side of the veth pair is assigned to the container network namespace
- An IP is allocated on the local node subnet and assigned to the container veth side
- A default route is added to the container via the local node bridge IP

## IPAM
The network service also manages IPAM.  IPAM data is stored in the Stellar datastore.
