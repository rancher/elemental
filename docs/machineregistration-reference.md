---
sidebar_label: Machineregistration reference
title: ''
---

<head>
  <meta charset="utf-8" />
  <title>Redirecting to https://elemental.docs.rancher.com</title>
  <meta http-equiv="refresh" content="0; URL=https://elemental.docs.rancher.com/" />
  <link rel="canonical" href="https://elemental.docs.rancher.com/" />
</head>

# MachineRegistration reference

The MachineRegistration resource is the responsible of defining a machine registration end point. Once created it generates a registration URL used by nodes to register so they are inventoried.

There are several keys that can be configured under a `MachineRegistration` resource spec.

There are several keys that can be configured under a `MachineRegistration` resource spec.

```yaml title="MachineRegistration" showLineNumbers
apiVersion: elemental.cattle.io/v1beta1
kind: MachineRegistration
metadata:
  name: my-nodes
  namespace: fleet-default
spec:
  machineName: name
  machineInventoryLabels:
    label: value
  machineInventoryAnnotations:
    annotation: value
  config:
    cloud-config:
        ...
    elemental:
        registration:
            ...
        install:
            ... 
```

#### config.cloud-config

Contains the cloud-configuration to be injected in the node. See the [Cloud Config Reference](cloud-config-reference.md) for full information.

#### config.elemental.registration
Contains the configuration used for the connection and the initial registration to the <Vars name="elemental_operator_name" />.

Supports the following values:

| Key               | Type   | Default value | Description                                                                                                                        |
|-------------------|--------|---------------|------------------------------------------------------------------------------------------------------------------------------------|
| url               | string | empty         | URL to connect to the <Vars name="elemental_operator_name" />                                                                                  |
| ca-cert           | string | empty         | CA to validate the certificate provided by the server at 'url' (required if the certificate is not signed by a public CA)          |
| emulate-tpm       | bool   | false         | this will use software emulation of the TPM (required for hosts without TPM hardware)                                              |
| emulated-tpm-seed | int64  | 1             | fixed seed to use with 'emulate-tpm': use for debug purposes only                                                                  |
| no-smbios         | bool   | false         | whether SMBIOS data should be sent to the <Vars name="elemental_operator_name" /> (see the [SMBIOS reference](smbios.md) for more information) |

#### config.elemental.install

Contains the installation configuration that would be applied via `operator-register` when booted from an ISO and passed to [`elemental-cli install`](https://github.com/rancher/elemental-cli/blob/main/docs/elemental_install.md)

Supports the following values:

| Key         | Type   | Default value | Description                                                                                                                                |
|-------------|--------|---------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| firmware    | string | efi           | Firmware to install ('efi' or 'bios')                                                                                                      |
| device      | string | empty         | Device to install the system to                                                                                                            |
| no-format   | bool   | false         | Don’t format disks. It is implied that COS_STATE, COS_RECOVERY, COS_PERSISTENT, COS_OEM partitions are already existing on the target disk |
| config-urls | list   | empty         | Cloud-init config files locations                                                                                                          |
| iso         | string | empty         | Performs an installation from the ISO url instead of the running ISO                                                                       |
| system-uri  | string | empty         | Sets the system image source and its type (e.g. 'docker:registry.org/image:tag') instead of using the running ISO                          |
| debug       | bool   | false         | Enable debug output                                                                                                                        |
| tty         | string | empty         | Add named tty to grub                                                                                                                      |
| poweroff    | bool   | false         | Shutdown the system after install                                                                                                          |
| reboot      | bool   | false         | Reboot the system after install                                                                                                            |
| eject-cd    | bool   | false         | Try to eject the cd on reboot                                                                                                              |

:::warning warning
In case of using both `iso` and `system-uri` the `iso` value takes precedence
:::

The only required value for a successful installation is the `device` key as we need a target disk to install to. The rest of the parameters are all optional.

<details>
<summary>Example</summary>

  ```yaml showLineNumbers
  apiVersion: elemental.cattle.io/v1beta1
  kind: MachineRegistration
  metadata:
    name: my-nodes
    namespace: fleet-default
  spec:
    config:
      elemental:
        install:
          device: /dev/sda
          debug: true
          reboot: true
          eject-cd: true
          system-uri: registry.opensuse.org/isv/rancher/elemental/teal52/15.3/rancher/elemental-node-image/5.2:latest
  ```
</details>

#### machineName

This refers to the name that will be set to the node and the kubernetes resources that require a hostname (rke2 deployed pods for example, they use the node hostname as part of the pod names)
`String` type.

:::info
When `elemental:registration:no-smbios` is set to `false` (default), machineName is interpolated with [SMBIOS](https://www.dmtf.org/standards/smbios) data which allows you to store hardware information.
See our [SMBIOS docs](smbios.md) for more information.
If no `machineName` is specified, a default one in the form `m-$UUID` will be set.
The UUID will be retrieved from the SMBIOS data if available, otherwise a random UUID will be generated.
:::

<details>
<summary>Example</summary>

  ```yaml showLineNumbers
  apiVersion: elemental.cattle.io/v1beta1
  kind: MachineRegistration
  metadata:
    name: my-nodes
    namespace: fleet-default
  spec:
    machineName: hostname-test-4
  ```

</details>

#### machineInventoryLabels

Labels that will be set to the `MachineInventory` that is created from this `MachineRegistration`
`Key: value` type. These labels will be used to establish a selection criteria in [MachineInventorySelectorTemplate](machineinventoryselectortemplate-reference.md).

:::info
When `elemental:registration:no-smbios` is set to `false` (default), Labels are interpolated with [SMBIOS](https://www.dmtf.org/standards/smbios) data. This allows to store hardware information in custom labels.
See our [SMBIOS docs](smbios.md) for more information.
:::

<details>
<summary>Example</summary>

  ```yaml showLineNumbers
  apiVersion: elemental.cattle.io/v1beta1
  kind: MachineRegistration
  metadata:
    name: my-nodes
    namespace: fleet-default
  spec:
    machineInventoryLabels:
      my.prefix.io/location: europe
      my.prefix.io/cpus: 32
      my.prefix.io/manufacturer: "${System Information/Manufacturer}"
      my.prefix.io/productName: "${System Information/Product Name}"
      my.prefix.io/serialNumber: "${System Information/Serial Number}"
      my.prefix.io/machineUUID: "${System Information/UUID}"
  ```

</details>

#### machineInventoryAnnotations

Annotations that will be set to the `MachineInventory` that is created from this `MachineRegistration`
`Key: value` type

<details>
<summary>Example</summary>

  ```yaml showLineNumbers
  apiVersion: elemental.cattle.io/v1beta1
  kind: MachineRegistration
  metadata:
    name: my-nodes
    namespace: fleet-default
  spec:
    machineInventoryAnnotations:
      owner: bob
      version: 1.0.0
  ```

</details>
