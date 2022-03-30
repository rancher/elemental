# Custom Images

RancherOS image can easily be remastered using a docker build.
For example, to add `cowsay` to RancherOS you would use the
following Dockerfile

## Docker image

```Dockerfile
# The version of RancherOS to modify
FROM rancher-sandbox/os2:VERSION

# Your custom commands
RUN zypper install -y cowsay

# IMPORTANT: /usr/lib/rancheros-release is used for versioning/upgrade. The
# values here should reflect the tag of the image currently being built
ARG IMAGE_REPO=norepo
ARG IMAGE_TAG=latest
RUN echo "IMAGE_REPO=${IMAGE_REPO}"          > /usr/lib/rancheros-release && \
    echo "IMAGE_TAG=${IMAGE_TAG}"           >> /usr/lib/rancheros-release && \
    echo "IMAGE=${IMAGE_REPO}:${IMAGE_TAG}" >> /usr/lib/rancheros-release
```

Where VERSION is the base version we want to customize. All version numbers available at [quay.io](https://quay.io/repository/costoolkit/os2?tab=tags) or [github](https://github.com/rancher-sandbox/os2/releases)

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

## Bootable images

To create bootable images from the docker image you just created
run the below command

```bash
# Download the ros-image-build script
curl -o ros-image-build https://raw.githubusercontent.com/rancher-sandbox/os2/main/ros-image-build

# Run the script creating a qcow image, an ISO, and an AMI
bash ros-image-build myrepo/custom-build:v1.1.1 qcow,iso,ami
```

The above command will create an ISO, a qcow image, and publish AMIs. You need not create all
three types and can change to comma seperated list to the types you care for.

## Auto-installing ISO

To create an ISO that upon boot will automatically run an installation, as an alternative to iPXE install,
run the following command.

```bash
bash ros-image-build myrepo/custom-build:v1.1.1 iso mycloud-config-file.txt
```

The third parameter is a path to a file that will be used as the cloud config passed to the installation.
Refer to the [installation](./installation.md) and [configuration reference](./configuration.md) for the
contents of the file.