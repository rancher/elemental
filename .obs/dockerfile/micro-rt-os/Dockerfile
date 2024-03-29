# SPDX-License-Identifier: Apache-2.0
# Define the names/tags of the container
#!ExclusiveArch: x86_64
#!BuildName: SL-Micro-rt-container
#!BuildTag: suse/sl-micro/%%SLMICRO_VERSION%%/rt-os-container:latest
#!BuildTag: suse/sl-micro/%%SLMICRO_VERSION%%/rt-os-container:%VERSION%
#!BuildTag: suse/sl-micro/%%SLMICRO_VERSION%%/rt-os-container:%VERSION%-%RELEASE%
#!BuildConstraint: hardware:disk:size unit=G 8
#!UseOBSRepositories

ARG SLMICRO_VERSION

FROM suse/sl-micro/${SLMICRO_VERSION}/baremetal-os-container:latest AS os

# dummy zypper call to get elemental into the build context and extract %VERSION% from it via _service
RUN zypper in --no-recommends -y systemd-presets-branding-Elemental elemental

# turn base image into a rt image
RUN zypper in --no-recommends -y -- -kernel-default kernel-rt

MAINTAINER SUSE LLC (https://www.suse.com/)

ARG SLMICRO_VERSION
ARG BUILD_REPO=%%IMG_REPO%%
ARG IMAGE_REPO=$BUILD_REPO/suse/sl-micro/%%SLMICRO_VERSION%%/rt-os-container
ARG IMAGE_TAG=%VERSION%-%RELEASE%

# Define labels according to https://en.opensuse.org/Building_derived_containers
# labelprefix=com.suse.sl.micro
LABEL org.opencontainers.image.title="SUSE Linux Micro Realtime OS Container"
LABEL org.opencontainers.image.description="Image containing SUSE Linux Micro Realtime - a containerized OS layer for Kubernetes."
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
RUN echo IMAGE_REPO=\"${IMAGE_REPO}\"         >> /etc/os-release && \
    echo IMAGE_TAG=\"${IMAGE_TAG}\"           >> /etc/os-release && \
    echo IMAGE=\"${IMAGE_REPO}:${IMAGE_TAG}\" >> /etc/os-release && \
    echo TIMESTAMP="`date +'%Y%m%d%H%M%S'`" >> /etc/os-release
RUN echo GRUB_ENTRY_NAME=\"SUSE Linux Micro\" >> /etc/os-release

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
RUN elemental init --force elemental-rootfs,elemental-sysroot,grub-config,dracut-config,cloud-config-essentials,elemental-setup,boot-assessment
