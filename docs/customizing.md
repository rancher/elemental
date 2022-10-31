---
sidebar_label: Custom Images
title: ''
---

# Custom Images

Elemental Teal images can be customized in different ways.
One option is to provide
additional resources within the installation media so that during installation, or
eventually at boot time, additional binaries such as drivers can be included.

Another option would be to remaster the Elemental Teal by simply using a docker build.
Elemental Teal is a regular container image, so it is absolutely possible to create
a new image using a Dockerfile based on Elemental Teal image.

## Customize installation ISO and installation process

In order to adapt the installation ISO a simple approach is to append extra configuration
files into the ISO root in the same way a registration yaml configuration file
is added.

### Additional configuration files

Elemental Teal installation can be customized in three different, non-exclusive ways. First, including
some custom Elemental client configuration file, second, by including additional cloud-init files to execute at
boot time, and finally, by including installation hooks.

#### Custom Elemental client configuration file

[Elemental client](https://github.com/rancher/elemental-cli) `install`, `upgrade` and `reset` commands can be configured with a
custom [configuration file](https://rancher.github.io/elemental-toolkit/docs/customizing/general_configuration/).

In order to set a custom configuration file in the installation
media the MachineRegistration resource associated with this ISO should also include
the Elemental client configuration directory. For that purpose, the `install` field
supports the `config-dir` field. See [MachineRegistration reference](/machineregistration-reference#configelementalinstall) and the example
below:

```yaml showLineNumbers
apiVersion: elemental.cattle.io/v1beta1
kind: MachineRegistration
metadata:
  name: my-nodes
  namespace: fleet-default
spec:
  ...
  config:
    ...
    elemental:
      ...
      install:
        ...
        config-dir: "/run/initramfs/live/elemental.conf.d"
```

Elemental Teal live ISOs, when booted, have the ISO root mounted at `/run/initramfs/live`.
So in that case, the ISO will contain the custom Elemental client configuration file
as `/elemental.conf.d/config.yaml`.

#### Adding additional cloud-init files at boot

In order to include additional cloud-init files during the installation they need
to be added to the installation data into the MachineRegistration resource. More specific
the `config-urls` field is used for this exact purpose. See [MachineRegistration reference](/machineregistration-reference) page.

`config-urls` is a list of string literals where each item is an http url pointing to a
cloud-init file or a local path of a cloud init file. Note the local path is evaluated at the
time of execution by the installation media, hence the local path must exist within
the installation media, commonly an ISO image.

Since in Elemental Teal live systems the ISO root is mounted at `/run/initramfs/live`,
the local paths for `config-url` in MachineRegistrations are likely to point there.
See the example below:

```yaml showLineNumbers
apiVersion: elemental.cattle.io/v1beta1
kind: MachineRegistration
metadata:
  name: my-nodes
  namespace: fleet-default
spec:
  ...
  config:
    ...
    elemental:
      ...
      install:
        ...
        config-urls:
        - "/run/initramfs/live/oem/10_install_extra_drivers.yaml"
```

In that case the ISO root is expected to include the `/oem/10_install_extra_drivers.yaml` file.

#### Installation hooks

[Elemental client](https://github.com/rancher/elemental-cli) `install`, `upgrade` and `reset` procedures include three different hooks:

* `before-install`: executed after all partition mountpoints are set.
* `after-install-chroot`: executed after deploying the OS image and before unmounting the associated loop filesystem image. Runs chrooted to the OS image.
* `after-install`: executed before unmounting partitions but after all OS images are set and unmounted.

Hooks are provided as cloud-init stages. Equivalent hooks exist for `reset` and `upgrade` procedures.

Hooks are evaluated at `install`,`reset` and `upgrade` processes from `/oem`, `/system/oem` and `/usr/local/cloud-config`, however
additional paths can be provided with the `cloud-init-paths` flag in [Elemental client configuration](https://rancher.github.io/elemental-toolkit/docs/customizing/general_configuration/).

### Adding extra driver binaries into the ISO example

This example is covering the case in which extra driver binaries are included into the ISO
and during the installation they are installed over the OS image.

For that use case the following files are required:

* additional binaries to install (they could be in the form of RPMs)
* additional hooks file to copy binaries into the persistent storage and to install them
* additional Elemental client configuration file to point hooks file location

Lets create an `overlay` directory to include the overlay root-tree that needs to be
applied over the ISO root. In that case the `overlay` directory could contain:

```yaml showLineNumbers
overlay/
  data/
    extra_drivers/
      some_driver.rpm
  hooks/
    install_hooks.yaml
  elemental/
    config.yaml
```

The Elemental client config file in `overlay/elemental` could be as:

```yaml showLineNumbers
cloud-init-paths:
  - "/run/initramfs/live/hooks"
```

This is just to let Elemental client know where to find installation hooks.

Finally, the `overlay/hooks/install_hooks.yaml` could be as:

```yaml showLineNumbers
name: "Install extra drivers"
stages:
  before-install:
    # Preload data to the persistent storage
    # During installation persistent partition is mounted at /run/cos/persistent
    - commands:
        - rsync -a /run/initramfs/live/data/ /run/cos/persistent
  after-install-chroot:
    # extra_drivers folder is at `/usr/local/extra_drivers` from the OS image chroot
    - commands:
      - rpm -iv /usr/local/extra_drivers/some_driver.rpm
```

Note the installation hooks only cover installation procedures, for upgrades equivalent
`before-upgrade` and/or `after-upgrade-chroot` should be defined.

### Repacking the ISO image with extra files

Assuming an `overlay` folder was created in the current directory containing all
additional files to be appended, the following `xorriso` command adds the extra files:

```bash showLineNumbers
xorriso -indev elemental-teal.x86_64.iso -outdev elemental-teal.custom.x86_64.iso -map overlay / -boot_image any replay
```

For that a `xorriso` equal or higher than version 1.5 is required.

## Remastering a custom docker image

Since Elemental Teal image is a Docker image it can also be used as a base image
in a Dockerfile in order to create a new container image.

Imagine some additional package from an extra repository is required, the following example
show case how this could be added:

```docker showLineNumbers
# The version of Elemental to modify
FROM registry.opensuse.org/isv/rancher/elemental/teal52/15.3/rancher/elemental-node-image/5.2:VERSION

# Custom commands
RUN rpm --import <repo-signing-key-url> && \
    zypper addrepo --refresh <repo_url> extra_repo && \
    zypper install -y <extra_package>

# IMPORTANT: /etc/os-release is used for versioning/upgrade. The
# values here should reflect the tag of the image currently being built
ARG IMAGE_REPO=norepo
ARG IMAGE_TAG=latest
RUN echo "IMAGE_REPO=${IMAGE_REPO}"          > /etc/os-release && \
    echo "IMAGE_TAG=${IMAGE_TAG}"           >> /etc/os-release && \
    echo "IMAGE=${IMAGE_REPO}:${IMAGE_TAG}" >> /etc/os-release
```

Where VERSION is the base version we want to customize.

And then the following commands

```bash showLineNumbers
docker build --build-arg IMAGE_REPO=myrepo/custom-build \
             --build-arg IMAGE_TAG=v1.1.1 \
             -t myrepo/custom-build:v1.1.1 .
docker push myrepo/custom-build:v1.1.1
```

The new customized OS is available as the Docker image `myrepo/custom-build:v1.1.1` and it can
be run and verified using docker with

```bash showLineNumbers
docker run -it myrepo/custom-build:v1.1.1 bash
```
