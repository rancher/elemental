# SPDX-License-Identifier: Apache-2.0
# Define the names/tags of the container
#!BuildTag: suse/sle-micro/base-%%SLEMICRO_VERSION%%:latest
#!BuildTag: suse/sle-micro/base-%%SLEMICRO_VERSION%%:%VERSION%
#!BuildTag: suse/sle-micro/base-%%SLEMICRO_VERSION%%:%VERSION%-%RELEASE%
#!BuildConstraint: hardware:disk:size unit=G 8
#

ARG SLE_VERSION
FROM suse/sle15:$SLE_VERSION as host

MAINTAINER SUSE LLC (https://www.suse.com/)

RUN mkdir /osimage

RUN rpm --initdb --root /osimage

RUN zypper --installroot /osimage in --no-recommends -y filesystem

# make system bootable
# iputils: https://github.com/rancher/elemental/issues/1138
RUN zypper --installroot /osimage in --no-recommends -y grub2 shim dracut iputils kernel systemd bash

#!ArchExclusiveLine: x86_64
RUN if [ `uname -m` = "x86-64" ]; then zypper --installroot /osimage in --no-recommends -y syslinux; fi

# make dracut happy
RUN zypper --installroot /osimage in --no-recommends -y squashfs NetworkManager-branding-SLE NetworkManager device-mapper iproute2 tar curl ca-certificates ca-certificates-mozilla

# make ARM happy
#!ArchExclusiveLine: aarch64
RUN if [ `uname -m` = "aarch64" ]; then zypper --installroot /osimage in -y grub2-arm64-efi raspberrypi-firmware raspberrypi-firmware-config raspberrypi-firmware-dt u-boot-rpiarm64; fi

# make SUSE happy
RUN zypper --installroot /osimage in --no-recommends -y SLE-Micro-Rancher-release systemd-presets-branding-SLE-Micro-for-Rancher

# make elemental-register happy
RUN zypper --installroot /osimage in --no-recommends -y dmidecode lvm2

# make Rancher (containerd) happy
RUN zypper --installroot /osimage in --no-recommends -y apparmor-parser

# add elemental
RUN zypper --installroot /osimage in --no-recommends -y elemental

# end of mandatory package installs for SLE Micro

# make derived containers possible
RUN zypper --installroot /osimage in --no-recommends -y zypper

FROM scratch as osimage

COPY --from=host /osimage /

ARG SLEMICRO_VERSION
ARG BUILD_REPO=%%IMG_REPO%%
ARG IMAGE_REPO=$BUILD_REPO/suse/sle-micro/base-$SLEMICRO_VERSION
ARG IMAGE_TAG=%VERSION%-%RELEASE%

# IMPORTANT: Setup elemental-release used for versioning/upgrade. The
# values here should reflect the tag of the image being built
# Also used by elemental-populate-labels
RUN echo IMAGE_REPO=\"${IMAGE_REPO}\"              >> /etc/os-release && \
    echo IMAGE_TAG=\"${IMAGE_TAG}\"                >> /etc/os-release && \
    echo IMAGE=\"${IMAGE_REPO}:${IMAGE_TAG}\"      >> /etc/os-release && \
    echo TIMESTAMP="`date +'%Y%m%d%H%M%S'`"        >> /etc/os-release && \
    echo GRUB_ENTRY_NAME=\"SLE Micro\" >> /etc/os-release

# Define labels according to https://en.opensuse.org/Building_derived_containers
# labelprefix=com.suse.sle.micro
LABEL org.opencontainers.image.title="SLE Micro"
LABEL org.opencontainers.image.description="Image containing SLE Micro - a containerized OS layer for Kubernetes."
LABEL org.opencontainers.image.version="${IMAGE_TAG}"
LABEL org.opencontainers.image.url="https://www.suse.com/products/micro/"
LABEL org.opencontainers.image.created="%BUILDTIME%"
LABEL org.opencontainers.image.vendor="SUSE LLC"
LABEL org.opencontainers.image.source="%SOURCEURL%"
LABEL org.opensuse.reference="${IMAGE_REPO}:${IMAGE_TAG}"
LABEL org.openbuildservice.disturl="%DISTURL%"
LABEL com.suse.supportlevel="l3"
LABEL com.suse.eula="sle-eula"
LABEL com.suse.lifecycle-url="https://www.suse.com/lifecycle#suse-linux-enterprise-micro"
LABEL com.suse.image-type="sle-micro"
LABEL com.suse.release-stage="released"
# endlabelprefix

# Make sure trusted certificates are properly generated
RUN /usr/sbin/update-ca-certificates

# Ensure /tmp is mounted as tmpfs by default
RUN if [ -e /usr/share/systemd/tmp.mount ]; then \
      cp /usr/share/systemd/tmp.mount /etc/systemd/system; \
    fi

# Save some space
RUN zypper clean --all && \
    rm -rf /var/log/update* && \
    >/var/log/lastlog && \
    rm -rf /boot/vmlinux*

# Rebuild initrd to setup dracut with the boot configurations
RUN elemental init --force immutable-rootfs,grub-config,dracut-config,cloud-config-essentials,elemental-setup && \
    # aarch64 has an uncompressed kernel so we need to link it to vmlinuz
    kernel=$(ls /boot/Image-* 2>/dev/null | head -n1) && \
    if [ -e "$kernel" ]; then ln -sf "${kernel#/boot/}" /boot/vmlinuz; fi
