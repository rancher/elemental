# Elemental
CI with latest stable Rancher Manager:

[![K3s-E2E-Stable_RM](https://github.com/rancher/elemental/actions/workflows/e2e-k3s-stable.yaml/badge.svg?branch=main)](https://github.com/rancher/elemental/actions/workflows/e2e-k3s-stable.yaml)
[![RKE2-E2E-Stable_RM](https://github.com/rancher/elemental/actions/workflows/e2e-rke2-stable.yaml/badge.svg?branch=main)](https://github.com/rancher/elemental/actions/workflows/e2e-rke2-stable.yaml)

[![K3s-UI_E2E-Stable_RM](https://github.com/rancher/elemental/actions/workflows/ui-e2e-k3s-stable.yaml/badge.svg?branch=main)](https://github.com/rancher/elemental/actions/workflows/ui-e2e-k3s-stable.yaml)
[![RKE2-UI_E2E-Stable_RM](https://github.com/rancher/elemental/actions/workflows/ui-e2e-rke2-stable.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/ui-e2e-rke2-stable.yaml)

CI with latest devel Rancher Manager:

[![K3s-E2E-Latest_RM](https://github.com/rancher/elemental/actions/workflows/e2e-k3s-latest.yaml/badge.svg?branch=main)](https://github.com/rancher/elemental/actions/workflows/e2e-k3s-latest.yaml)
[![RKE2-E2E-Latest_RM](https://github.com/rancher/elemental/actions/workflows/e2e-rke2-latest.yaml/badge.svg?branch=main)](https://github.com/rancher/elemental/actions/workflows/e2e-rke2-latest.yaml)

[![K3s-UI_E2E-Latest_RM](https://github.com/rancher/elemental/actions/workflows/ui-e2e-k3s-latest.yaml/badge.svg?branch=main)](https://github.com/rancher/elemental/actions/workflows/ui-e2e-k3s-latest.yaml)
[![RKE2-UI_E2E-Latest_RM](https://github.com/rancher/elemental/actions/workflows/ui-e2e-rke2-latest.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/ui-e2e-rke2-latest.yaml)

Elemental is a software stack enabling a centralized, full cloud-native OS management solution with Kubernetes.

Cluster Node OSes are built and maintained via container images through the [Elemental Toolkit](https://rancher.github.io/elemental-toolkit/) and installed on new hosts using the [Elemental CLI](https://github.com/rancher/elemental-cli).

The [Elemental Operator](https://github.com/rancher/elemental-operator) and the [Rancher System Agent](https://github.com/rancher/system-agent) enable Rancher Manager to fully control Elemental clusters, from the installation and management of the OS on the Nodes to the provisioning of new K3s or RKE2 clusters in a centralized way.

Follow our [Quickstart](https://rancher.github.io/elemental/quickstart/) or see the [full docs](https://rancher.github.io/elemental/) for more info.
