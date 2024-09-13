# Elemental

[![Lint](https://github.com/rancher/elemental/actions/workflows/lint.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/lint.yaml)

## Daily CI

[![CLI-K3s](https://github.com/rancher/elemental/actions/workflows/cli-k3s-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/cli-k3s-matrix.yaml)
[![CLI-K3s-Upgrade](https://github.com/rancher/elemental/actions/workflows/cli-k3s-upgrade-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/cli-k3s-upgrade-matrix.yaml)
[![CLI-RKE2](https://github.com/rancher/elemental/actions/workflows/cli-rke2-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/cli-rke2-matrix.yaml)
[![CLI-RKE2-Upgrade](https://github.com/rancher/elemental/actions/workflows/cli-rke2-upgrade-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/cli-rke2-upgrade-matrix.yaml)

[![UI-K3s](https://github.com/rancher/elemental/actions/workflows/ui-k3s-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/ui-k3s-matrix.yaml)
[![UI-K3s-Upgrade](https://github.com/rancher/elemental/actions/workflows/ui-k3s-upgrade-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/ui-k3s-upgrade-matrix.yaml)
[![UI-RKE2](https://github.com/rancher/elemental/actions/workflows/ui-rke2-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/ui-rke2-matrix.yaml)
[![UI-RKE2-Upgrade](https://github.com/rancher/elemental/actions/workflows/ui-rke2-upgrade-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/ui-rke2-upgrade-matrix.yaml)

## Weekly CI

[![CLI-K3s-Airgap](https://github.com/rancher/elemental/actions/workflows/cli-k3s-airgap-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/cli-k3s-airgap-matrix.yaml)
[![CLI-K3s-Downgrade](https://github.com/rancher/elemental/actions/workflows/cli-k3s-downgrade-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/cli-k3s-downgrade-matrix.yaml)
[![CLI-K3s-Scalability](https://github.com/rancher/elemental/actions/workflows/cli-k3s-scalability-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/cli-k3s-scalability-matrix.yaml)
[![CLI-K3s-SELinux](https://github.com/rancher/elemental/actions/workflows/cli-k3s-selinux-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/cli-k3s-selinux-matrix.yaml)
[![CLI-Full-Backup-Restore](https://github.com/rancher/elemental/actions/workflows/cli-full-backup-restore-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/cli-full-backup-restore-matrix.yaml)
[![CLI-Multicluster](https://github.com/rancher/elemental/actions/workflows/cli-multicluster-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/cli-multicluster-matrix.yaml)
[![CLI-Rancher-Manager-Head-2.7](https://github.com/rancher/elemental/actions/workflows/cli-rm-head-2.7-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/cli-rm-head-2.7-matrix.yaml)
[![CLI-Regression](https://github.com/rancher/elemental/actions/workflows/cli-regression-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/cli-regression-matrix.yaml)
[![UI-Rancher-Manager-Head-2.7](https://github.com/rancher/elemental/actions/workflows/ui-rm-head-2.7-matrix.yaml/badge.svg)](https://github.com/rancher/elemental/actions/workflows/ui-rm-head-2.7-matrix.yaml)

## Goal

Elemental is a software stack enabling a centralized, full cloud-native OS management solution with Kubernetes.

Cluster Node OSes are built and maintained via container images through the [Elemental Toolkit](https://rancher.github.io/elemental-toolkit/) and installed on new hosts using the [Elemental CLI](https://github.com/rancher/elemental-cli).

The [Elemental Operator](https://github.com/rancher/elemental-operator) and the [Rancher System Agent](https://github.com/rancher/system-agent) enable Rancher Manager to fully control Elemental clusters, from the installation and management of the OS on the Nodes to the provisioning of new K3s or RKE2 clusters in a centralized way.

Follow our [Quickstart](https://rancher.github.io/elemental/quickstart/) or see the [full docs](https://rancher.github.io/elemental/) for more info.

## License

Copyright (c) 2020-2024 [SUSE, LLC](http://suse.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
