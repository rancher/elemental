## Introduction

Elemental uses the [Rancher System Agent](https://github.com/rancher/system-agent) renamed to Elemental System Agent,
to initially bootstrap the node with a simple plan that sets some labels for the node,
sets the proper hostname according to the `#!yaml MachineInventory`
and installs the default Rancher System Agent from Rancher Server which then will install the proper Kubernetes components.

This small bootstrap service also accepts local plans stored under `/var/lib/elemental/agent/plans` so any plan written
in there will also be applied during the initial node start (After it has been installed).

!!! note
    This local plans run only during the initial Elemental bootstrap **before** Kubernetes is installed on the node


## Types of Plans

 - One time instructions: Only run once
 - Periodic instructions: They run periodically
 - Files: Creates files
 - Probes: http probes


Both one time instructions and periodic instructions can run either a direct command or a docker image.


## Adding local plans on Elemental

You can add those plans to Elemental as part of the `#!yaml MachineRegistration` CRD, in the `cloud-config` section by adding it like so

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

These plans are provided as a quick reference and not guaranteed to work. To learn more about plans please check [Rancher System Agent](https://github.com/rancher/system-agent)

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
