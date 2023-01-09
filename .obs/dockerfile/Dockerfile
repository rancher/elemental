# SPDX-License-Identifier: Apache-2.0
# Define the names/tags of the container
#!BuildTag: rancher/elemental-teal/5.3:latest
#!BuildTag: rancher/elemental-teal/5.3:%VERSION%
#!BuildTag: rancher/elemental-teal/5.3:%VERSION%-%RELEASE%
#

FROM suse/sle-micro-rancher/5.3:latest

# Elemental Teal essentials
# kernel-firmware-amdgpu installed here until 5.4 is released and includes it
RUN zypper in -y -- \
  NetworkManager-wwan \
  kernel-firmware-all \
  -systemd-presets-branding-SLE-Micro-for-Rancher \
  elemental

# Install extra useful utilities
RUN zypper in -y k9s

# IMPORTANT: Setup elemental-release used for versioning/upgrade. The
# values here should reflect the tag of the image being built
# Also used by elemental-populate-labels
ARG IMAGE_TAG=%VERSION%-%RELEASE%
ARG IMAGE_REPO=%%IMG_REPO%%/rancher/elemental-teal/5.3
RUN echo IMAGE_REPO=\"${IMAGE_REPO}\"         >> /etc/os-release && \
    echo IMAGE_TAG=\"${IMAGE_TAG}\"           >> /etc/os-release && \
    echo IMAGE=\"${IMAGE_REPO}:${IMAGE_TAG}\" >> /etc/os-release && \
    echo TIMESTAMP="`date +'%Y%m%d%H%M%S'`" >> /etc/os-release
RUN echo GRUB_ENTRY_NAME=\"Elemental\" >> /etc/os-release

# Define labels according to https://en.opensuse.org/Building_derived_containers
# labelprefix=com.rancher.elemental
LABEL org.opencontainers.image.title="Rancher Elemental Teal OS image"
LABEL org.opencontainers.image.description="Elemental Teal fully containerized OS"
LABEL org.opencontainers.image.version="%VERSION%"
LABEL org.opencontainers.image.url="https://github.com/rancher/elemental"
LABEL org.opencontainers.image.created="%BUILDTIME%"
LABEL org.opencontainers.image.vendor="SUSE LLC"
LABEL org.opensuse.reference="%%IMG_REPO%%/rancher/elemental-teal/5.3"
LABEL org.openbuildservice.disturl="%DISTURL%"
LABEL com.suse.supportlevel="techpreview"
# endlabelprefix

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
RUN mkinitrd && \
    # aarch64 has an uncompressed kernel so we need to link it to vmlinuz
    kernel=$(ls /boot/Image-* | head -n1) && \
    if [ -e "$kernel" ]; then ln -sf "${kernel#/boot/}" /boot/vmlinuz; fi

