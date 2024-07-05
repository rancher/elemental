# 2024-07-08 - Elemental Network configuration demo

## Why

### In this demo

1. Assign static IP addresses to K8s nodes.  
1. Mitigate the risk of DHCP as a single point of failure.
1. Remove the need of "infinite leases" when using DHCP.  
1. Provide a way to automatically configure the machine network, bonding, VLANs, etc.

### Not in this demo

1. Allow a complete DHCP-less setup
1. Allow network configuration per cluster (rather than per inventory pool/registration)

## Context

### This demo

![demo](./images/network-demo.png)  

### Next

![demo-next](./images/network-demo-next.png)

### Ideally

![demo-ideal](./images/network-demo-ideal.png)
