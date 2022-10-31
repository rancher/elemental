---
sidebar_label: Inventory Management
title: ''
---

## Inventory Management

The Elemental operator can hold an inventory of machines and
the mapping of the machine to it's configuration and assigned cluster.

### MachineInventory

#### Reference

```yaml
apiVersion: elemental.cattle.io/v1beta1
kind: MachineInventory
metadata:
  name: machine-a
  # The namespace must match the namespace of the cluster
  # assigned to the clusters.provisioning.cattle.io resource
  namespace: fleet-default
spec:
  # The cluster that this machine is assigned to
  clusterName: some-cluster
  # The hash of the TPM EK public key. This is used if you are
  # using TPM2 to identifiy nodes.  You can obtain the TPM by
  # running `rancherd get-tpm-hash` on the node. Or nodes can
  # report their TPM hash by using the MachineRegister
  tpm: d68795c6192af9922692f050b...
  # Generic SMBIOS fields that are typically populated with
  # the MachineRegister approach
  smbios: {}
  # A reference to a secret that contains a shared secret value to
  # identify a node.  The secret must be of type "elemental.cattle.io/token"
  # and have on field "token" which is the value of the shared secret
  machineTokenSecretName: some-secret-name
  # Arbitrary cloud config that will be added to the machines cloud config
  # during the rancherd bootstrap phase.  The one important field that should
  # be set is the role.
  config:
    role: server
```

### MachineRegistration

#### Reference

```yaml
apiVersion: elemental.cattle.io/v1beta1
kind: MachineRegistration
metadata:
  name: machine-registration
  # The namespace must match the namespace of the cluster
  # assigned to the clusters.provisioning.cattle.io resource
  namespace: fleet-default
spec:
  # Labels to be added to the created MachineInventory object
  machineInventoryLabels: {}
  # Annotations to be added to the created MachineInventory object
  machineInventoryAnnotations: {}
  # The cloud config that will be used to provision the node
  cloudConfig: {}
```
