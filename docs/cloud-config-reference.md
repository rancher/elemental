---
sidebar_label: Cloud-config reference
title: ''
---

# Cloud-config Reference

All custom configuration applied on top of a fresh deployment should come
from the `cloud-config` section in a `MachineRegistration`.

This will get run by [`elemental-cli run-stage`](https://github.com/rancher/elemental-cli/blob/main/docs/elemental_run-stage.md) during the `boot` stage, and
it will be stored in the node under the `/oem` dir.

Elemental uses [yip](https://github.com/mudler/yip) to run these cloud-config files, so we support the [yip subset cloud-config implementation](https://github.com/mudler/yip#compatibility-with-cloud-init-format).

Below is an example of the supported configuration on a `MachineRegistration` resource.

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
      cloud-config:
        users:
          - name: "bar"
          passwd: "foo"
          groups: "users"
          homedir: "/home/foo"
          shell: "/bin/bash"
          ssh_authorized_keys:
            - faaapploo
        # Assigns these keys to the first user in users or root if there
        # is none
        ssh_authorized_keys:
          - asdd
        # Run these commands once the system has fully booted
        runcmd:
          - foo
        # Write arbitrary files
        write_files:
          - encoding: b64
            content: CiMgVGhpcyBmaWxlIGNvbnRyb2xzIHRoZSBzdGF0ZSBvZiBTRUxpbnV4
            path: /foo/bar
            permissions: "0644"
            owner: "bar"
      elemental:
        install:
          reboot: true
          device: /dev/sda
          debug: true
    machineName: my-machine
    machineInventoryLabels:
      location: "europe"
  ```
  
</details>
