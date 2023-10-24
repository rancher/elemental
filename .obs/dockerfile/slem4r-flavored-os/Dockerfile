# SPDX-License-Identifier: Apache-2.0
# Define the names/tags of the container
#!BuildTag: suse/sle-micro-rancher/%%BUILD_FLAVOR%%%%SLEMICRO_VERSION%%:latest
#!BuildTag: suse/sle-micro-rancher/%%BUILD_FLAVOR%%%%SLEMICRO_VERSION%%:%VERSION%-%RELEASE%
#!BuildTag: suse/sle-micro-rancher/%%BUILD_FLAVOR%%%%SLEMICRO_VERSION%%:%RELEASE%
#!BuildConstraint: hardware:disk:size unit=G 8
#

ARG SLEMICRO_VERSION

FROM suse/sle-micro-rancher/base-${SLEMICRO_VERSION}:latest

# dummy zypper call to get elemental into the build context and extract %VERSION% from it via _service
RUN zypper in --no-recommends -y systemd-presets-branding-SLE-Micro-for-Rancher elemental

# Define labels according to https://en.opensuse.org/Building_derived_containers
# labelprefix=com.suse.sle.micro.rancher
LABEL org.opencontainers.image.version="%%BUILD_FLAVOR%%%%SLEMICRO_VERSION%%.%VERSION%.%RELEASE%"
LABEL org.opencontainers.image.created="%BUILDTIME%"
LABEL org.opencontainers.image.source="%SOURCEURL%"
# endlabelprefix


RUN if [[ "%%BUILD_FLAVOR%%" == "kvm-" ]]; then \
  # replace default kernel with smaller one, sufficient to run kvm guests \
  zypper in --no-recommends -y kernel-default-base -kernel-default; \
else \
  # extend -base with baremetal and usability packages \
  zypper in --no-recommends -y procps openssl openssh vim-small less iputils kernel-firmware-all NetworkManager-wwan cryptsetup; \
fi

ARG BUILD_REPO=%%IMG_REPO%%
ARG IMAGE_REPO=$BUILD_REPO/rancher/sle-micro-for-rancher/$SLEMICRO_VERSION
ARG IMAGE_TAG=%VERSION%-%RELEASE%

# IMPORTANT: Setup elemental-release used for versioning/upgrade. The
# values here should reflect the tag of the image being built
# Also used by elemental-populate-labels
RUN grep -v IMAGE_REPO /etc/os-release | grep -v IMAGE_TAG | grep -v IMAGE= | grep -v TIMESTAMP | grep -v GRUB_ENTRY_NAME > /tmp/os-release
RUN mv /tmp/os-release /etc/os-release
RUN echo IMAGE_REPO=\"${IMAGE_REPO}\"               >> /etc/os-release && \
    echo IMAGE_TAG=\"${IMAGE_TAG}\"                 >> /etc/os-release && \
    echo IMAGE=\"${IMAGE_REPO}:${IMAGE_TAG}\"       >> /etc/os-release && \
    echo TIMESTAMP="`date +'%Y%m%d%H%M%S'`"         >> /etc/os-release && \
    echo GRUB_ENTRY_NAME=\"SLE Micro for Rancher\"  >> /etc/os-release

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
RUN elemental init --force immutable-rootfs,grub-config,dracut-config,cloud-config-essentials,elemental-setup && \
    # aarch64 has an uncompressed kernel so we need to link it to vmlinuz
    kernel=$(ls /boot/Image-* 2>/dev/null | head -n1) && \
    if [ -e "$kernel" ]; then ln -sf "${kernel#/boot/}" /boot/vmlinuz; fi
