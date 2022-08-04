# Installation

## Overview

The design of Elemental is that you boot from an installation image and through cloud-init and Kubernetes mechanisms
the node will be configured. An installation image is essentially a regular Elemental image baked in some installation media,
most likely a bootable ISO or a iPXE setup for network boots, including all the registration metadata to
comunicate with the Rancher management cluster.

## Installation Configuration

The installation configuration is mostly applied and set as part of the registration process.
The registration process is done by the Elemental-operator client who is the responsible to register
the node in a Rancher management cluster and fetch the installation configuration.

## Elemental Partition Table

Elemental requires the following partitions.  These partitions are required by [Elemental-toolkit](https://rancher.github.io/elemental-toolkit/docs)

| Label          | Default Size    | Contains                                                    |
| ---------------|-----------------|------------------------------------------------------------ |
| COS_BOOT       |          50 MiB | UEFI Boot partition                                         |
| COS_STATE      |          15 GiB | A/B bootable file system images constructed from OCI images |
| COS_OEM        |          50 MiB | OEM cloud-config files and other data                       |
| COS_RECOVERY   |           8 GiB | Recovery file system image if COS_STATE is destroyed        |
| COS_PERSISTENT | Remaining space | All contents of the persistent folders                      |

## Folders

| Path              | Read-Only | Ephemeral | Persistent |
| ------------------|:---------:|:---------:|:----------:|
| /                 | x         |           |            |
| /etc              |           | x         |            |
| /etc/cni          |           |           | x          |
| /etc/iscsi        |           |           | x          |
| /etc/rancher      |           |           | x          |
| /etc/ssh          |           |           | x          |
| /etc/systemd      |           |           | x          |
| /srv              |           | x         |            |
| /home             |           |           | x          |
| /opt              |           |           | x          |
| /root             |           |           | x          |
| /var              |           | x         |            |
| /usr/libexec      |           |           | x          |
| /var/lib/cni      |           |           | x          |
| /var/lib/kubelet  |           |           | x          |
| /var/lib/longhorn |           |           | x          |
| /var/lib/rancher  |           |           | x          |
| /var/lib/wicked   |           |           | x          |
| /var/log          |           |           | x          |
