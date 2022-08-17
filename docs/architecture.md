---
sidebar_label: Architecture
title: ''
---

# Architecture

Elemental is a stack, a set of tools, to build an immutable Linux distribution.

Its primary purpose is to run Rancher and its corresponding Kubernetes distributions [RKE2](https://rke2.io) 
and [k3s](https://k3s.io). It can be configured for any other workload, however
the following documentation focuses on a Rancher use-case.

Initial node configurations is done using a
cloud-init style approach and all further maintenance is done using
Kubernetes operators.

## Use Cases

The OS built by Elemental is intended to be run as the operating system beneath a Rancher Multi-Cluster 
Management server or as a node in a Kubernetes cluster managed by Rancher. As such it
also allows you to build stand alone Kubernetes clusters that run an embedded
and smaller version of Rancher to manage the local cluster. A key attribute of Elemental
is that it is managed by Rancher and thus Rancher will exist either locally in the cluster
or centrally with Rancher Multi-Cluster Manager.

## OCI Image based

Elemental is an image based distribution with an A/B style update mechanism. One first runs
on a read-only image A and to do an upgrade pulls a new read only image
B and then reboots the system to run on B. What is unique about
Elemental is that the runtime images come from OCI Images. Not an
OCI Image containing special artifacts, but an actual Docker runnable
image that is built using standard Docker build processes. Elemental is
built using normal `docker build` and if you wish to customize the OS
image all you need to do is create a new `Dockerfile`.

## {{elemental.operator.name}}

Elemental includes no container runtime, Kubernetes distribution,
or Rancher itself. All of these assets are dynamically pulled at runtime. All that
is included in Elemental is [{{elemental.operator.name}}] which
is responsible for managing OS upgrades and managing a secure device inventory to assist
with zero touch provisioning.

{{elemental.operator.name}} includes a Kubernetes operator installed in the management cluster and a client
side installed in nodes, so they can self register into the management cluster. Once a node is
registered the {{elemental.operator.name}} will kick-start the OS installation and schedule the Kubernetes
provisioning using the [{{ranchersystemagent.name}}].
Rancher System Agent is responsible for bootstrapping RKE2/k3s and Rancher from an OCI registry. This means
an update of containerd, k3s, RKE2, or Rancher does not require an OS upgrade
or node reboot.

## Cloud-init

Elemental is initially configured using a simple version of `cloud-init`.
It is not expected that one will need to do a lot of customization to Elemental
as the core OS's sole purpose is to run Rancher and Kubernetes and not serve as
a generic Linux distribution.

## Elemental Teal

Elemental Teal is the OS, based on SUSE Linux Enterprise (SLE) Micro for Rancher,
built using the Elemental stack. The only assumption from the Elemental stack is that
the underlying distribution is based on Systemd. We choose SLE Micro for Rancher for
obvious reasons, but beyond that Elemental provides a stable layer to build upon
that is well tested and has paths to commercial support, if one chooses.
