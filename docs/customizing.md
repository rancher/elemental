---
sidebar_label: Custom Images
title: ''
---

# Custom Images

Elemental image can easily be remastered using a docker build.
For example, to add `cowsay` to Elemental you would use the
following Dockerfile

## Docker image

```docker
# The version of Elemental to modify
FROM registry.opensuse.org/isv/rancher/elemental/teal52/15.3/rancher/elemental-node-image/5.2:VERSION

# Your custom commands
RUN zypper install -y cowsay

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

```bash
docker build --build-arg IMAGE_REPO=myrepo/custom-build \
             --build-arg IMAGE_TAG=v1.1.1 \
             -t myrepo/custom-build:v1.1.1 .
docker push myrepo/custom-build:v1.1.1
```

Your new customized OS is available at in the docker image `myrepo/custom-build:v1.1.1` and you can
check out your new image using docker with

```bash
docker run -it myrepo/custom-build:v1.1.1 bash
```

## Installation ISO

To create an ISO that upon boot will automatically attempt to register run the `elemental-iso-build` script

```bash
bash elemental-iso-build CONFIG_FILE
```

Where CONFIG_FILE is the path to the configuration file including the registration data to register against the
Rancher management cluster.
