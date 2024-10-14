# SPDX-License-Identifier: Apache-2.0
# Define the names/tags of the container
#!BuildName: SL-Micro-baremetal-container
#!BuildTag: suse/sl-micro/%%SLMICRO_VERSION%%/baremetal-os-container:latest
#!BuildTag: suse/sl-micro/%%SLMICRO_VERSION%%/baremetal-os-container:%VERSION%
#!BuildTag: suse/sl-micro/%%SLMICRO_VERSION%%/baremetal-os-container:%VERSION%-%RELEASE%
#!BuildConstraint: hardware:disk:size unit=G 8
#!UseOBSRepositories

ARG SLMICRO_VERSION

FROM suse/sl-micro/${SLMICRO_VERSION}/base-os-container:latest

# dummy zypper call to get elemental into the build context and extract %VERSION% from it via _service
RUN zypper in --no-recommends -y systemd-presets-branding-Elemental elemental

# extend -base with baremetal and usability packages
RUN zypper in --no-recommends -y procps openssh openssh-server \
    vim-small less kernel-firmware-all NetworkManager-wwan cryptsetup \
    \
    avahi bash-completion catatonit cni cni-plugins conmon conntrack-tools \
    container-selinux ethtool fuse-overlayfs hdparm hostname iptables \
    issue-generator jq keyutils lshw lsof lsscsi mdadm multipath-tools \
    netcat-openbsd net-tools nfs-client \
    nss-mdns open-iscsi open-vm-tools pciutils pciutils-ids podman policycoreutils \
    procmail psmisc rpcbind runc selinux-policy selinux-policy-targeted \
    selinux-tools slirp4netns strace sudo sysstat system-user-nobody \
    timezone traceroute xtables-plugins

ARG SLMICRO_VERSION
ARG BUILD_REPO=%%IMG_REPO%%
ARG IMAGE_REPO=$BUILD_REPO/suse/sl-micro/%%SLMICRO_VERSION%%/baremetal-os-container
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
    echo GRUB_ENTRY_NAME=\"SUSE Linux Micro\" >> /etc/os-release

# Define labels according to https://en.opensuse.org/Building_derived_containers
# labelprefix=com.suse.sl.micro
LABEL org.opencontainers.image.title="SUSE Linux Micro Baremetal OS Container"
LABEL org.opencontainers.image.description="SUSE Linux Micro Baremetal OS - a containerized OS layer for Kubernetes."
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
LABEL com.suse.image-type="sl-micro"
LABEL com.suse.release-stage="beta"
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

# Add the arm-firwmware feature
#!ArchExclusiveLine: aarch64
RUN if [[ "$(uname -m)"  == "aarch64" ]]; then elemental init --mkinitrd=false arm-firmware; fi

# Rebuild initrd to setup dracut with the boot configurations
RUN elemental init --force elemental-rootfs,elemental-sysroot,grub-config,dracut-config,cloud-config-essentials,elemental-setup,boot-assessment
