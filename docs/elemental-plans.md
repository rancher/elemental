## Introduction

Elemental uses the [Rancher System Agent](https://github.com/rancher/system-agent), renamed to Elemental System Agent, to initially bootstrap the node with a simple plan.

 The plan will apply the following configurations:
 - Set some labels for the node
 - Set the proper hostname according to the `#!yaml MachineInventory` value
 - Install the default Rancher System Agent from Rancher Server, and install the proper Kubernetes components

The bootstrap service also accepts local plans stored under `/var/lib/elemental/agent/plans`. Any plan written
in there will also be applied during the initial node start after the installation is completed.

!!! note
    The local plans run only during the initial Elemental bootstrap **before** Kubernetes is installed on the node.


## Types of Plans

The type of plans that Elemental can use are:

 - One time instructions: Only run once
 - Periodic instructions: They run periodically
 - Files: Creates files
 - Probes: http probes

!!! note
    Both one time instructions and periodic instructions can run either a direct command or a docker image.

## Adding local plans on Elemental

You can add local plans to Elemental as part of the `#!yaml MachineRegistration` CRD, in the `cloud-config` section as follow:

```yaml
apiVersion: elemental.cattle.io/v1beta1
kind: MachineRegistration
metadata:
  name: my-nodes
  namespace: fleet-default
spec:
  config:
    cloud-config:
      users:
        - name: root
          passwd: root
      write_files:
        - path: /var/lib/elemental/agent/plans/mycustomplan.plan
          permissions: "0600"
          content: |
            {"instructions":
                [
                  {
                    "name":"set hostname",
                    "command":"hostnamectl",
                    "args": ["set-hostname", "myHostname"]
                  },
                  {
                    "name":"stop sshd service",
                    "command":"systemctl",
                    "args": ["stop", "sshd"]
                  }
                ]
            }
    elemental:
      install:
        reboot: true
        device: /dev/sda
        debug: true
  machineName: my-machine
  machineInventoryLabels:
    location: "europe"
```


## Plan examples

The following plans are provided as a quick reference and are not guaranteed to work in your environment. To learn more about plans please check [Rancher System Agent](https://github.com/rancher/system-agent).

```json title="simple command plan"
{"instructions":
    [
        {
            "name":"set hostname",
            "command":"hostnamectl",
            "args": ["set-hostname", "myHostname"]
        },
        {
            "name":"stop sshd service",
            "command":"systemctl",
            "args": ["stop", "sshd"]
        }
    ]
}
```

```json title="periodic docker plan"
{"periodicInstructions":
    [
        {
            "name":"set hostname",
            "image":"ghcr.io/rancher-sandbox/elemental-example-plan:main"
            "command": "run.sh"
        }
    ]
}
```

```json title="file creation plan"
{"files":
    [
        {
            "content":"Welcome to the system",
            "path":"/etc/motd",
            "permissions": "0644"
        }
    ]
}
```

```json title="probe plan"
{"probes":
    "probe1": {
        "name": "Service Up",
        "httpGet": {
            "url": "http://10.0.0.1/healthz",
            "insecure": "false",
            "clientCert": "....",
            "clientKey": "....",
            "caCert": "....."
        }   
    }
}
```
