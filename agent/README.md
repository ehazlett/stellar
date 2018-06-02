# Element Agent
The Element agent package handles simplified clustering among nodes.  The agent handles
peer management (joins, expiry, etc) as well as exposing GRPC services and publishing
the GRPC endpoint address for each node throughout the group.  This enables services
to be built on top of the agent and allow simple publishing of the GRPC endpoing
to other nodes for accessing the services.
