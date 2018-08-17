# Stellar
Simplified Container Platform

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

## About
- [Introduction](intro/about/)
  - [What is Stellar](intro/about.md)
  - [Architecture](intro/architecture.md)
- [Deployment](install/)
  - [Requirements](install/index.md#requirements)
  - [Installation](install/index.md#deployment)
