# Overview

Elemental is a software stack enabling a centralized, full cloud-native OS management with Kubernetes.

Cluster Node OSes are built and mainteined via container images through the [Elemental toolkit](https://rancher.github.io/elemental-toolkit/) and installed on new hosts using the [Elemental cli](https://github.com/rancher/elemental-cli).

The [Elemental Operator](https://github.com/rancher/elemental-operator) and the [Rancher System Agent](https://github.com/rancher/system-agent) enable Rancher Manager to fully control Elemental clusters, from the installation and management of the OS on the Nodes to the provisioning of new K3s or RKE2 clusters in a centralized way.

Ready to give it a try? Get an Elemental Cluster up and running following the [Quickstart](quickstart.md) section.

Want more details? Take a look at the [Architecture](architecture.md) section.