ARG ELEMENTAL_TOOLKIT
ARG ELEMENTAL_REGISTER

FROM ${ELEMENTAL_REGISTER} as register 
FROM ${ELEMENTAL_TOOLKIT} as toolkit

# OS base image of our choice
FROM registry.opensuse.org/opensuse/leap:15.5 as OS

ARG RANCHER_SYSTEM_AGENT_VERSION

# install kernel, systemd, dracut, grub2 and other required tools
RUN ARCH=$(uname -m); \
    [[ "${ARCH}" == "aarch64" ]] && ARCH="arm64"; \
    zypper --non-interactive install --no-recommends -- \
      kernel-default \
      device-mapper \
      dracut \
      grub2 \
      grub2-${ARCH}-efi \
      shim \
      haveged \
      systemd \
      NetworkManager \
      openssh-server \
      openssh-clients \
      timezone \
      parted \
      e2fsprogs \
      dosfstools \
      mtools \
      xorriso \
      findutils \
      gptfdisk \
      rsync \
      squashfs \
      lvm2 \
      tar \
      gzip \
      vim \
      which \
      less \
      sudo \
      curl \
      iproute2 \
      podman \
      sed \
      btrfsprogs \
      snapper

# elemental-register dependencies
RUN ARCH=$(uname -m); \
    [[ "${ARCH}" == "aarch64" ]] && ARCH="arm64"; \
    zypper --non-interactive install --no-recommends -- \
      dmidecode

# Add system files
COPY framework/files/ /
# Add elemental-register
COPY --from=register /usr/sbin/elemental-register /usr/sbin/elemental-register
COPY --from=register /usr/sbin/elemental-support /usr/sbin/elemental-support
# Add the elemental cli
COPY --from=toolkit /usr/bin/elemental /usr/bin/elemental

# Add the elemental-system-agent
ADD --chmod=0755 https://github.com/rancher/system-agent/releases/download/${RANCHER_SYSTEM_AGENT_VERSION}/rancher-system-agent-amd64 /usr/sbin/elemental-system-agent

# Enable essential services
RUN systemctl enable NetworkManager.service

# Enable /tmp to be on tmpfs
RUN cp /usr/share/systemd/tmp.mount /etc/systemd/system

# Generate initrd with required elemental services
RUN elemental init --debug --force

# Update os-release file with some metadata
RUN echo TIMESTAMP="`date +'%Y%m%d%H%M%S'`" >> /etc/os-release && \
    echo GRUB_ENTRY_NAME=\"Elemental Dev\" >> /etc/os-release

# Good for validation after the build
CMD /bin/bash
