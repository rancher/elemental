---
sidebar_label: Operator overview
title: ''
---

# Operator

[elemental-operator](https://github.com/rancher/elemental-operator) The Elemental operator is responsible for managing the Elemental versions and maintaining a machine inventory to assist with edge or bare metal installations.

This chart bootstraps an elemental-operator deployment on the [Rancher Manager v2.6](https://rancher.com/docs/rancher/v2.6/) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Rancher Manager version v2.6
- Helm client version v3.8.0+

## Get Helm chart info

```console
helm pull oci://registry.opensuse.org/isv/rancher/elemental/charts/elemental/elemental-operator
helm show all oci://registry.opensuse.org/isv/rancher/elemental/charts/elemental/elemental-operator
```

## Installation

The Elemental operator can be added to a cluster running Rancher Multi Cluster
Management server with the Helm chart as follows:

```bash
helm -n cattle-elemental-system install --create-namespace elemental-operator https://github.com/rancher/elemental-operator/releases/download/v0.1.0/elemental-operator-0.1.0.tgz
```

The command deploys elemental-operator on the Kubernetes cluster in the default configuration.

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](./values.yaml), or run these configuration commands:

```console
helm show values oci://registry.opensuse.org/isv/rancher/elemental/charts/elemental/elemental-operator
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| image.empty | string | `rancher/pause:3.1` |  |
| image.repository | string | `quay.io/costoolkit/elemental-operator` | Source image for elemental-operator with repository name  |
| image.tag | tag | `""` |  |
| image.imagePullPolicy | string | `IfNotPresent` |  |
| noProxy | string | `127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,.svc,.cluster.local" | Comma separated list of domains or ip addresses that will not use the proxy |
| global.cattle.systemDefaultRegistry | string | `""` | Default container registry name  |
| sync_interval | string | `"60m"` | Default sync interval for upgrade channel |
| sync_namespaces | list | `[]` | Namespace the operator will watch for, leave empty for all |
| debug | bool | `false` | Enable debug output for operator |
| nodeSelector.kubernetes.io/os | string | `linux` |  |
| tolerations | object | `{}` |  |
| tolerations.key | string | `cattle.io/os` |  |
| tolerations.operator | string | `"Equal"` |  |
| tolerations.value | string | `"linux"` |  |
| tolerations.effect | string | `NoSchedule` |  |


## Managing Upgrades

The Elemental operator will manage the upgrade of the local cluster where the operator
is running and also any downstream cluster managed by Rancher Multi-Cluster
Manager.

### ManagedOSImage

The ManagedOSImage kind used to define what version of Elemental should be
running on each node. The simplest example of this type would be to change
the image of the local nodes.

```bash
kubectl edit -n fleet-default default-os-image
```

```yaml
apiVersion: elemental.cattle.io/v1beta1
kind: ManagedOSImage
metadata:
  name: default-os-image
  namespace: fleet-default
spec:
  osImage: rancher/elemental:v0.0.0
```

A `ManagedOSImage` can also use a `ManagedOSVersion` to drive upgrades. 
To use a `ManagedOSVersion` specify a `managedOSVersionName`, as `osImage` takes precedence, mind to set back as empty:

```bash
kubectl edit -n fleet-default default-os-image
```

```yaml
apiVersion: elemental.cattle.io/v1beta1
kind: ManagedOSImage
metadata:
  name: default-os-image
  namespace: fleet-local
spec:
  osImage: ""
  managedOSVersionName: "version-name"
```

### Reference

Below is reference of the full type

```yaml
apiVersion: elemental.cattle.io/v1beta1
kind: ManagedOSImage
metadata:
  name: arbitrary
  
  # There are two special namespaces to consider.  If you wish to manage
  # nodes on the local cluster this namespace must be `fleet-local`. If
  # you wish to manage nodes in Rancher MCM managed clusters then the
  # namespace must match the namespace of the clusters.provisioning.cattle.io resource
  # which is typically fleet-default.
  namespace: fleet-default
spec:
  # The image name to pull for the OS. Overrides managedOSVersionName when specified
  osImage: rancher/os2:v0.0.0

  # The ManagedOSVersion to use for the upgrade
  managedOSVersionName: ""
  
  # The selector for which nodes will be select.  If null then all nodes
  # will be selected
  nodeSelector:
    matchLabels: {}
    
  # How many nodes in parallel to update.  If empty the default is 1 and
  # if set to 0 the rollout will be paused
  concurrency: 2
    
  # Arbitrary action to perform on the node prior to upgrade
  prepare:
    image: ubuntu
    command: ["/bin/sh"]
    args: ["-c", "true"]
    env:
    - name: TEST_ENV
      value: testValue
      
  # Parameters to control the drain behavior.  If null no draining will happen
  # on the node.
  drain:
    # Refer to kubectl drain --help for the definition of these values
    timeout: 5m
    gracePeriod: 5m
    deleteLocalData: false
    ignoreDaemonSets: true
    force: false
    disableEviction: false
    skipWaitForDeleteTimeout: 5
    
  # Which clusters to target
  # This is used if you are running Rancher MCM and managing
  # multiple clusters.  The syntax of this field matches the
  # Fleet targets and is described at https://fleet.rancher.io/gitrepo-targets/
  clusterTargets: []

  # Overrides the default container created for running the upgrade with a custom one
  # This is optional and used only if specific upgrading mechanisms needs to be applied
  # in place of the default behavior.
  # The image used here overrides ones specified in osImage, depending on the upgrade strategy.
  upgradeContainer:
    image: ubuntu
    command: ["/bin/sh"]
    args: ["-c", "true"]
    env:
    - name: TEST_ENV
      value: testValue
```

## Uninstall Chart

```console
helm uninstall -n cattle-elemental-system elemental-operator
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._
