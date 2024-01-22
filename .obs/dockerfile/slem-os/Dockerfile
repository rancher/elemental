# SPDX-License-Identifier: Apache-2.0
# Define the names/tags of the container
#!BuildTag: suse/sle-micro/%%SLEMICRO_VERSION%%:latest
#!BuildTag: suse/sle-micro/%%SLEMICRO_VERSION%%:%VERSION%
#!BuildTag: suse/sle-micro/%%SLEMICRO_VERSION%%:%VERSION%-%RELEASE%
#!BuildConstraint: hardware:disk:size unit=G 8
#

ARG SLEMICRO_VERSION

FROM suse/sle-micro/base-${SLEMICRO_VERSION}:latest

# dummy zypper call to get elemental into the build context and extract %VERSION% from it via _service
RUN zypper in --no-recommends -y systemd-presets-branding-SLE-Micro-for-Rancher elemental

# extend -base with baremetal and usability packages
RUN zypper in --no-recommends -y procps openssl openssh openssh-server \
    vim-small less kernel-firmware-all NetworkManager-wwan cryptsetup \
    \
    avahi bash-completion catatonit cni cni-plugins conmon conntrack-tools \
    container-selinux ethtool fuse-overlayfs haveged hdparm hostname iptables \
    issue-generator jq k9s keyutils lshw lsof lsscsi mdadm multipath-tools \
    netcat-openbsd net-tools nfs-client nfsidmap \
    nss-mdns open-iscsi open-vm-tools pciutils pciutils-ids podman policycoreutils \
    procmail psmisc rpcbind runc selinux-policy selinux-policy-targeted \
    selinux-tools slirp4netns strace sudo sysstat system-user-nobody \
    timezone traceroute xtables-plugins

ARG SLEMICRO_VERSION
ARG BUILD_REPO=%%IMG_REPO%%
ARG IMAGE_REPO=$BUILD_REPO/suse/sle-micro/$SLEMICRO_VERSION
ARG IMAGE_TAG=%VERSION%-%RELEASE%

# IMPORTANT: Setup elemental-release used for versioning/upgrade. The
# values here should reflect the tag of the image being built
# Also used by elemental-populate-labels
RUN grep -v "IMAGE_REPO\|IMAGE_TAG\|IMAGE=\|TIMESTAMP\|GRUB_ENTRY_NAME" /etc/os-release > /tmp/os-release
RUN mv /tmp/os-release /etc/os-release
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
LABEL org.opencontainers.image.created="%BUILDTIME%"
LABEL org.opencontainers.image.source="%SOURCEURL%"
LABEL org.opensuse.reference="${IMAGE_REPO}:${IMAGE_TAG}"
LABEL org.openbuildservice.disturl="%DISTURL%"
# endlabelprefix

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
