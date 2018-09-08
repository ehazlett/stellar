# Element
Element handles simplified clustering among nodes.  It handles
peer management (joins, expiry, etc) as well as exposing GRPC services and publishing
the GRPC endpoint address for each node throughout the group.  This enables services
to be built using element and allow simple publishing of the GRPC endpoints
to other nodes for accessing the services.
