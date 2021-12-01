# Understanding Clusters

RancherOS bootstraps a node with Kubernetes (k3s/rke2) and Rancher such
that all future management of Kubernetes and Rancher can be done from
Kubernetes. This is done by running Rancherd once per node on boot. Once the system has
been fully bootstrapped it will not run again. Rancherd is ran from cloud-init
and it's configuration is embedded in the cloud-config file.

## Cluster Initialization

Creating a cluster always starts with one node initializing the cluster, and
all other nodes joining the cluster by pointing to a `server` node. The node
that will initialize a new cluster is the one with `role: server` and
`server: ""` (empty). The new cluster will have a token generated or you can
manually assign a unique string. The token for an existing cluster can be determined
by running `rancherd get-token` on a server node.

## Joining Nodes

Nodes can be joined to the cluster as the role `server` to add more control
plane nodes or as the role `agent` to add more worker nodes. To join a node
you must have the Rancher server URL (which is by default running on port
`8443`) and the token.  The server and token are assigned to the `server` and
`token` fields respectively.

## Node Roles

Rancherd will bootstrap a node with one of the following roles

2. __server__: Joins the cluster as a new control-plane,etcd,worker node
3. __agent__: Joins the cluster as a worker only node.

## Server discovery

It can be quite cumbersome to automate bringing up a clustered system
that requires one bootstrap node.  Also there are more considerations
around load balancing and replacing nodes in a proper production setup.
Rancherd support server discovery based on [go-discover](https://github.com/hashicorp/go-discover).

To use server discovery you must set the `role`, `discovery` and `token` fields.
The `discovery` configuration will be used to dynamically determine what
is the server URL and if the current node should act as the node to initialize the cluster.

Example
```yaml
role: server
discovery:
  params:
    # Corresponds to go-discover provider name
    provider: "mdns"
    # All other key/values are parameters corresponding to what 
    # the go-discover provider is expecting
    service: "rancher-server"
  # If this is a new cluster it will wait until 3 server are 
  # available and they all agree on the same cluster-init node
  expectedServers: 3
  # How long servers are remembered for. It is useful for providers
  # that are not consistent in their responses, like mdns.
  serverCacheDuration: 1m
```