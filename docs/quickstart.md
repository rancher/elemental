# Quickstart
Follow this guide to have an auto-deployed cluster via rke2/k3s and managed by Rancher 
with the only help of an Elemental Teal iso

## Introduction

---

#### What is Elemental Teal ?

Elemental Teal is the combination of "SLE Micro for Rancher" with the Rancher Elemental stack

SLE Micro for Rancher is a containerized and "stripped to the bones" operating system layer. It only contains grub2, dracut, a kernel, and systemd.

Its sole purpose is to run Kubernetes (k3s or RKE2), with everything controlled through Rancher Manager.

Elemental Teal is built in the [openSUSE Build Service](https://build.opensuse.org/package/show/isv:Rancher:Elemental:Teal52/node-image)
and available through the [openSUSE Registry](registry.opensuse.org/isv/rancher/elemental/teal52/15.3/rancher/elemental-node-image/5.2:latest)

#### What is the Rancher Elemental Stack ?

The Elemental Stack consists of some packages on top of SLE Micro for Rancher

- **elemental-operator** - this connects to Rancher Manager and handles machineRegistration and machineInventory CRDs
- **elemental-register** - this registers machines via machineRegistrations and installs them via elemental-cli
- **elemental-cli** - This installs any elemental-toolkit based derivative. Basically an installer based on our A/B install and upgrade system
- **rancher-system-agent** - runs on the installed system and gets instructions ("Plans") from Rancher Manager what to install and run on the system

## Prerequisites

 - A rancher cluster (2.6.6)
 - A machine (bare metal or virtualized) with TPM 2.0
   - Hint: Libvirt allows setting virtual TPMs for virtual machines
 - Helm Package Manager (https://helm.sh/)
 - Docker (for building the iso)


## Preparing the cluster

elemental-operator is the management endpoint, running the management 
cluster and taking care of creating inventories, registrations for machines and much more.

We will use the Helm package manager to install the elemental-operator chart into our cluster

```shell
$ helm upgrade --create-namespace -n cattle-elemental-system --install elemental-operator oci://registry.opensuse.org/isv/rancher/elemental/charts/elemental/elemental-operator`
```

There is a few options that can be set in the chart install but that is out of scope for this document. You can see all the values on the chart [values.yaml](https://github.com/rancher/elemental-operator/blob/main/chart/values.yaml)

Now after a few seconds you should see the operator pod appear on the `cattle-elemental-system` namespace.


```shell
$ kubectl get pods -n cattle-elemental-system
NAME                                  READY   STATUS    RESTARTS   AGE
elemental-operator-64f88fc695-b8qhn   1/1     Running   0          16s

```


## Prepare you kubernetes resources

Node deployment starts with a machineRegistration, identifying a set of machines sharing the same configuration (disk drives, network, etc.)

Then it continues with having a Cluster resource that uses a MachineInventorySelectorTemplate to know which machines are for that cluster.

This selector is a simple matcher based on labels set in the `MachineInventory`, so if your selector is matching the `cluster-id` key with a value `myId` 
and your `MachineInventory` has that same key with that value, it will match and be bootstrapped as part of the cluster.

You will need to create the following files.

#### selector.yaml
```yaml
apiVersion: elemental.cattle.io/v1beta1
kind: MachineInventorySelectorTemplate
metadata:
  name: my-machine-selector
  namespace: fleet-default
spec:
  template:
    spec:
      selector:
        matchExpressions:
        - key: node-location
          operator: In
          values: [ 'europe' ]
```

As you can see this is a very simple selector that checks the key `node-location` for the value `europe`


#### cluster.yaml

```yaml
kind: Cluster
apiVersion: provisioning.cattle.io/v1
metadata:
  name: my-cluster
  namespace: fleet-default
spec:
  rkeConfig:
    machinePools:
    - controlPlaneRole: true
      etcdRole: true
      machineConfigRef:
        apiVersion: elemental.cattle.io/v1beta1
        kind: MachineInventorySelectorTemplate
        name: my-machine-selector
      name: pool1
      quantity: 1
      unhealthyNodeTimeout: 0s
      workerRole: true
  kubernetesVersion: v1.23.7+k3s1
```

As you can see we are setting that our `machineConfigRef` is of Kind `MachineInventorySelectorTemplate` with the name `my-machine-selector`, which matches the selector we created.


#### registration.yaml

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
    elemental:
      install:
        reboot: true
        device: /dev/sda
        debug: true
  machineName: my-machine
  machineInventoryLabels:
     - node-location: "europe"
```

This creates a `MachineRegistration` which will provide a unique URL which we will use with `elemental-register` to register
the node during installation, so the operator can create a MachineInventory which will be using to bootstrap the node.
See that we set the label that match our selector here already, although it can always be added later to the `MachineInventory`.

`Note: Make sure to modify the registration.yaml above to set the proper install device to point to a valid device based on your node configuration(i.e. /dev/sda, /dev/vda, /dev/nvme0, etc...)`


Now that we have all the configuration to create the proper resources in Kubernetes just apply them

```shell
$ kubectl apply -f selector.yaml 
machineinventoryselectortemplate.elemental.cattle.io/my-machine-selector created
$ kubectl apply -f cluster.yaml 
cluster.provisioning.cattle.io/my-cluster created
$ kubectl apply -f registration.yaml
machineregistration.elemental.cattle.io/my-nodes created
```

## Preparing the iso

Now this is the last step, we need to prepare an Elemental Teal iso that includes the initial registration config, so
it can be auto registered, installed and fully deployed as part of our cluster. The contents of the file are nothing 
more than the registration url that the node needs to register and the proper server certificate, so it can connect securely.
This iso then can be used to provision an infinite number of machines

Now, our `MachineRegistration` provides the needed config in its resource as part of its `Status.RegistrationURL`,
so we can use that url to obtain the proper yaml needed for the iso.

```shell
$ wget --no-check-certificate `kubectl get machineregistration -n fleet-default my-nodes -o jsonpath="{.status.registrationURL}"` -O initial-registration.yaml
```

This will download the proper yaml from the registration URL and store it on the current directory under the `initial-registration.yaml` name

We provide a ISO build script for ease of use that can create the final ISO and inject the `initial-registration.yaml`:

```shell
$ wget -q https://raw.githubusercontent.com/rancher/elemental/master/elemental-iso-build && chmod +x elemental-iso-build
```

Now that we have the script we can proceed to build the ISO with our configuration injected:

```shell
$ ./elemental-iso-build initial-registration.yaml
```

This will generate an ISO on the current directory with the name `elemental-<timestamp>.iso`


You can now boot your nodes with this ISO, and they will:

 - Boot from the ISO
 - Register with the registrationURL given and create a per-machine `MachineInventory`
 - Install Elemental Teal to the given device
 - Restart
 - Auto-deploy the cluster via k3s

After a few minutes your new cluster will be fully provisioned!!


## How can I follow what is going on behind the scenes?

You should be able to follow along what the machine is doing via:

- During ISO boot
   - ssh into the machine (user/pass: root/ros)
      - running `journalctl -f -t elemental` will show you the output of the elemental-register and the elemental install
- Once the system is installed
   - On the Rancher UI -> `Cluster Management` you should see your new cluster and be able to see the `Provisioning Log` in the cluster details
   - ssh into the machine (user/pass: Whatever your configured on the registration.yaml under `Spec.config.cloud-config.users`)
      - running `journalctl -f -u elemental-system-agent` will show the output of the initial elemental config and install of `rancher-system-agent`
      - running `journalctl -f -u rancher-system-agent` will show the output of the boostrap of cluster components like k3s
      - running `journalctl -f -u k3s` will show the logs of the k3s deployment