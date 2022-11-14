---
sidebar_label: Installation
title: ''
---

# Installation

## Overview

Elemental stack provides OS management using OCI containers and Kubernetes. The Elemental
stack installation encompasses the installation of the <Vars name="elemental_operator_name" /> into the
management cluster and the creation and use of Elemental Teal installation media to
provide the OS into the Cluster Nodes. See [Architecture](architecture.md) section to read about the
interaction of the components.

The installation configuration is mostly applied and set as part of the registration process.
The registration process is done by the `elemental-register` (the <Vars name="elemental_operator_name" /> client part)
who is the responsible to register nodes in a Rancher management cluster and fetch the installation configuration.

Please refer to the [Quick Start](quickstart.md) guide for simple step by step deployment instructions.

## Elemental Operator Installation

The <Vars name="elemental_operator_name" /> is responsible for managing the Elemental versions and
maintaining a machine inventory to assist with edge or bare metal installations. <Vars name="elemental_operator_name" />
requires a cluster including the Rancher Manager and it can be installed with a helm chart.

See <Vars name="elemental_operator_name" /> [helm chart reference](elementaloperatorchart-reference.md) for install,
uninstall, upgrade and configuration details.

## Prepare Kubernetes Resources

Once the <Vars name="elemental_operator_name" /> is up and running within the management cluster a couple of kubernetes
resources are required in order to prepare an Elemental based cluster deployment.

* [MachineInventorySelectorTemplate](machineinventoryselectortemplate-reference.md):
  This resource identifies the criteria to match registered boxes (listed as part of the MachineInventory)
  against available Rancher 2.6 Clusters. As soon as there is a match the selected kubernetes cluster takes
  ownership of the registered box.
  
* [MachineRegistration](machineregistration-reference.md):
  This resource defines OS deployment details for any machine attempting to register. The machine
  registration is the entrance for Elemental nodes as it handles the authentication (based on TPM),
  the Elemental Teal deployment and the node inclusion into to the MachineInventory so it can be added
  to a cluster when there is a match based on a MachineInventorySelectorTemplate. The MachineRegistration
  object includes the machine registration URL that nodes use to register against it.

A Rancher Cluster resource is also required to deploy Elemental, it can be manually created as exemplified in
the [Quick Start](quickstart.md) guide or created from the Rancher 2.6 UI.

## Prepare Installation Media

The installation media is the media that will be used to kick start an Elemental Teal deployment. Currently
the supported media is a live ISO. The live ISO must include the registration configuration yaml hence it must
crafted once the MachineRegistration is created. The installation media is created by using the `elemental-iso-add-registration`
helper script (see [quick start](quickstart.md#preparing-the-iso) guide)
or by using the `elemental build-iso` command line utility included as part of the <Vars name="elemental_toolkit_name" link="elemental_toolkit_url/docs/creating-derivatives/build_iso" />.

Within MachineRegistration only a subset of OS installation parameters can be configured, all available parameters are listed
at [MachineRegistration](machineregistration-reference.md) reference page.

In order to configure the installation beyond the common options provided within the
[`elemental.install`](machineregistration-reference.md#configelementalinstall) section a `config.yaml`
configuration file can be included into the ISO (see [Custom Images](customizing.md#custom-elemental-client-configuration-file)).
Note any configuration applied as part of `elemental.install` section of the MachineRegistration will be
applied on top of the settings included in any custom `config.yaml` file.

Most likely the cloud-init configuration is enough to configure and set the deployed node at boot, however
if for some reason firstboot actions or scripts are required it is possible to also include
Rancher System Agent plans into the installation media. Refer to the [Elemental Plans](elemental-plans.md) section for details and
some example plans. The plans could be included into the squashed rootfs at `/var/lib/elemental/agent/plans`
folder and they would be seen by the system agent at firstboot.

## Start Installation Process

The installation starts by booting the installation media on a node. Once the installation media has booted it will
attempt to contact the management cluster and register to it by calling `elemental-register` command.
As the registration yaml configuration is already included into the ISO `elemental-register` knows the registration URL and
any other required data for the registration.

On a succeeded registration the installation media will start the Elemental Teal installation into the host based
on the configuration already included in the media and the MachineRegistration parameters. As soon as the installation
is done the node is ready to reboot. The deployed Elemental Teal includes a system agent plan to
kick start a regular rancher provisioning process to install the selected kubernetes version, once booted, after
some minutes the node installation is finalized and the node is included into the cluster and visible through
the Rancher UI.

## Deployed Elemental Teal Partition Table

Once Elemental Teal is installed the OS partition table, according to default values, will look like

| Label          | Default Size    | Contains                                                    |
|----------------|-----------------|-------------------------------------------------------------|
| COS_GRUB       | 64 MiB          | UEFI Boot partition                                         |
| COS_STATE      | 15 GiB          | A/B bootable file system images constructed from OCI images |
| COS_OEM        | 64 MiB          | OEM cloud-config files and other data                       |
| COS_RECOVERY   | 8 GiB           | Recovery file system image if COS_STATE is destroyed        |
| COS_PERSISTENT | Remaining space | All contents of the persistent folders                      |

Note this is the basic structure of any OS built by the <Vars name="elemental_toolkit_name" link="elemental_toolkit_url" />

## Elemental Teal Immutable Root

One of the characteristics of Elemental OSes is the setup of an immutable root filesystem where some ephemeral or
persistent locations are applied on top of it. Elemental Teal default folders structure is listed in the
matrix below.

| Path                    | Read-Only | Ephemeral | Persistent |
|-------------------------|:---------:|:---------:|:----------:|
| /                       |     x     |           |            |
| /etc                    |           |     x     |            |
| /etc/cni                |           |           |     x      |
| /etc/iscsi              |           |           |     x      |
| /etc/rancher            |           |           |     x      |
| /etc/ssh                |           |           |     x      |
| /etc/systemd            |           |           |     x      |
| /srv                    |           |     x     |            |
| /home                   |           |           |     x      |
| /opt                    |           |           |     x      |
| /root                   |           |           |     x      |
| /var                    |           |     x     |            |
| /usr/libexec            |           |           |     x      |
| /var/lib/cni            |           |           |     x      |
| /var/lib/kubelet        |           |           |     x      |
| /var/lib/longhorn       |           |           |     x      |
| /var/lib/rancher        |           |           |     x      |
| /var/lib/elemetal       |           |           |     x      |
| /var/lib/NetworkManager |           |           |     x      |
| /var/lib/calico         |           |           |     x      |
| /var/log                |           |           |     x      |
