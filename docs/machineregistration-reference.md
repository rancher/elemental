
There is several keys that can be configured under a `#!yaml MachineRegistration` resource spec.


```yaml title="MachineRegistration"
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
        install:
            ... 
```


#### config.cloud-config

Contains the cloud-configuration to be injected in the node. See the [Cloud Config Reference](cloud-config-reference.md) for full information.

#### config.elemental.install

Contains the installation configuration that would be applied via `operator-register` when booted from an ISO and passed to [`elemental-cli install`](https://github.com/rancher/elemental-cli/blob/main/docs/elemental_install.md)

Supports the following values:

| Key         | Type   | Description                                                                                                                                |
|-------------|--------|--------------------------------------------------------------------------------------------------------------------------------------------|
| firmware    | string | Firmware to install ('efi' or 'bios') (default "efi")                                                                                      |
| device      | string | Device to install the system to                                                                                                            |
| no-format   | bool   | Don’t format disks. It is implied that COS_STATE, COS_RECOVERY, COS_PERSISTENT, COS_OEM partitions are already existing on the target disk |
| config-urls | list   | Cloud-init config files locations                                                                                                          |
| iso         | string | Performs an installation from the ISO url instead of the running ISO                                                                       |
| system-uri  | string | Sets the system image source and its type (e.g. 'docker:registry.org/image:tag') instead of using the running ISO                          |
| debug       | bool   | Enable debug output                                                                                                                        |
| tty         | string | Add named tty to grub                                                                                                                      |
| poweroff    | bool   | Shutdown the system after install                                                                                                          |
| reboot      | bool   | Reboot the system after install                                                                                                            |
| eject-cd    | bool   | Try to eject the cd on reboot                                                                                                              |

!!! warning
    In case of using both `iso` and `system-uri` the `iso` value takes precedence

The only required value for a successful installation is the `device` key as we need a target disk to install to. The rest of the parameters are all optional.

??? example
    ```yaml
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


#### machineName

This refers to the name that will be set to the node and the kubernetes resources that require a hostname (rk2 pods for example)
`String` type.

??? example
    ```yaml
    apiVersion: elemental.cattle.io/v1beta1
    kind: MachineRegistration
    metadata:
      name: my-nodes
      namespace: fleet-default
    spec:
      machineName: hostname-test-4
    ```

#### machineInventoryLabels

Labels that will be set to the `#!yaml MachineInventory` that is created from this `#!yaml MachineRegistration`
`Key: value` type

??? example
    ```yaml
    apiVersion: elemental.cattle.io/v1beta1
    kind: MachineRegistration
    metadata:
      name: my-nodes
      namespace: fleet-default
    spec:
      machineInventoryLabels:
        my.prefix.io/location: europe
        my.prefix.io/cpus: 32
    ```


#### machineInventoryAnnotations

Annotations that will be set to the `#!yaml MachineInventory` that is created from this `#!yaml MachineRegistration`
`Key: value` type

??? example
    ```yaml
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

