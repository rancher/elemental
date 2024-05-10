# SPDX-License-Identifier: Apache-2.0
# Define the names/tags of the container
#!BuildName: SL-Micro-k3s-selinux-container
#!BuildTag: suse/sl-micro/%%SLMICRO_VERSION%%/k3s-selinux-os-container:latest
#!BuildTag: suse/sl-micro/%%SLMICRO_VERSION%%/k3s-selinux-os-container:%VERSION%
#!BuildTag: suse/sl-micro/%%SLMICRO_VERSION%%/k3s-selinux-os-container:%VERSION%-%RELEASE%
#!BuildConstraint: hardware:disk:size unit=G 8
#!UseOBSRepositories

ARG SLMICRO_VERSION

FROM suse/sl-micro/${SLMICRO_VERSION}/base-os-container:latest

# dummy zypper call to get elemental into the build context and extract %VERSION% from it via _service
RUN zypper in --no-recommends -y systemd-presets-branding-Elemental elemental

# replace default kernel with smaller one, sufficient to run kvm guests
RUN zypper in --no-recommends -y kernel-default-base -kernel-default qemu-guest-agent

RUN zypper in --no-recommends -y patterns-microos-selinux k3s-selinux

MAINTAINER SUSE LLC (https://www.suse.com/)

ARG SLMICRO_VERSION
ARG BUILD_REPO=%%IMG_REPO%%
ARG IMAGE_REPO=$BUILD_REPO/suse/sl-micro/%%SLMICRO_VERSION%%/k3s-selinux-os-container
ARG IMAGE_TAG=%VERSION%-%RELEASE%

# Define labels according to https://en.opensuse.org/Building_derived_containers
# labelprefix=com.suse.sl.micro
LABEL org.opencontainers.image.title="SUSE Linux Micro K3S SELinux"
LABEL org.opencontainers.image.description="Image containing SUSE Linux Micro K3s SELinux - a containerized OS layer for Kubernetes."
LABEL org.opencontainers.image.version="${IMAGE_TAG}"
LABEL org.opencontainers.image.created="%BUILDTIME%"
LABEL org.opencontainers.image.source="%SOURCEURL%"
LABEL org.opensuse.reference="${IMAGE_REPO}:${IMAGE_TAG}"
LABEL org.openbuildservice.disturl="%DISTURL%"
# endlabelprefix

# IMPORTANT: Setup elemental-release used for versioning/upgrade. The
# values here should reflect the tag of the image being built
# Also used by elemental-populate-labels
RUN grep -v "IMAGE_REPO\|IMAGE_TAG\|IMAGE=\|TIMESTAMP\|GRUB_ENTRY_NAME" /etc/os-release > /tmp/os-release
RUN mv /tmp/os-release /etc/os-release
RUN echo IMAGE_REPO=\"${IMAGE_REPO}\"              >> /etc/os-release && \
    echo IMAGE_TAG=\"${IMAGE_TAG}\"                >> /etc/os-release && \
    echo IMAGE=\"${IMAGE_REPO}:${IMAGE_TAG}\"      >> /etc/os-release && \
    echo TIMESTAMP="`date +'%Y%m%d%H%M%S'`"        >> /etc/os-release && \
    echo GRUB_ENTRY_NAME=\"SUSE Linux Micro\" >> /etc/os-release

# Ensure /tmp is mounted as tmpfs by default
RUN if [ -e /usr/share/systemd/tmp.mount ]; then \
      cp /usr/share/systemd/tmp.mount /etc/systemd/system; \
    fi

# Save some space
RUN if [ -f /usr/bin/zypper ]; then zypper clean --all; fi
# Save more space
RUN rm -rf /var/log/update* && \
    >/var/log/lastlog && \
    rm -rf /boot/vmlinux*

# Rebuild initrd to setup dracut with the boot configurations
RUN elemental init --force elemental-rootfs,elemental-sysroot,grub-config,dracut-config,cloud-config-essentials,elemental-setup,boot-assessment

# Set SELinux, this should be the default and not needed anymore as soon as SELinux
# is fully supported. Manually set here in a hacky way for testing purposes only
RUN sed -i "s|SELINUX=.*|SELINUX=enforcing|g" /etc/selinux/config && \
    sed -i "s| enforcing=0||g" /etc/elemental/bootargs.cfg && \
    sed -i "s|selinux=0|enforcing=0|g" /etc/elemental/bootargs.cfg && \
    sed -i "s| selinux=1||g" /etc/elemental/bootargs.cfg
