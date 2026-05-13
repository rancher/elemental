ARG ELEMENTAL_TOOLKIT
ARG ELEMENTAL_REGISTER

FROM ${ELEMENTAL_REGISTER} AS register
FROM ${ELEMENTAL_TOOLKIT} AS toolkit

# OS base image of our choice
FROM registry.opensuse.org/opensuse/leap:16.0 AS os

ARG RANCHER_SYSTEM_AGENT_VERSION
ARG RANCHER_SYSTEM_AGENT_CHECKSUM
ARG NMC_x86_CHECKSUM=a0bb9439d14db5071b63c5f8d1407cd243b1093794d207771e69cf506c6ea8c9
ARG NMC_arm_CHECKSUM=eff11304ffb45782cf17c7ec95e52c0b9b9257e8ac14ff5a290645db6efb1740
ARG NMSTATECTL_CHECKSUM=a0aaa72e9fd1c9df35a875339268cbe8acafdf269cfbca2b7a4897ba4033aa69

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
      btrfsmaintenance \
      snapper \
      glibc-gconv-modules-extra \
      wget \
      unzip

# elemental-register dependencies
RUN ARCH=$(uname -m); \
    [[ "${ARCH}" == "aarch64" ]] && ARCH="arm64"; \
    zypper --non-interactive install --no-recommends -- \
      dmidecode

# Install nmstatectl
RUN wget https://github.com/nmstate/nmstate/releases/download/v2.2.55/nmstatectl-linux-x64.zip && \
    echo "${NMSTATECTL_CHECKSUM} nmstatectl-linux-x64.zip" | sha256sum -c - && \
    unzip nmstatectl-linux-x64.zip && \
    chmod +x nmstatectl && \
    mv ./nmstatectl /usr/sbin/nmstatectl && \
    rm nmstatectl-linux-x64.zip

# Install nm-configurator
RUN curl -o /usr/sbin/nmc -L https://github.com/suse-edge/nm-configurator/releases/download/v0.3.5/nmc-linux-$(uname -m)
RUN ARCH=$(uname -m); CHECKSUM=${NMC_x86_CHECKSUM}; \
    if [[ "${ARCH}" == "aarch64" ]]; then \
      CHECKSUM=${NMC_arm_CHECKSUM}; \
    fi; \
    echo "${CHECKSUM} /usr/sbin/nmc" | sha256sum -c -
RUN chmod +x /usr/sbin/nmc

# SELinux policy and tools
RUN ARCH=$(uname -m); \
    [[ "${ARCH}" == "aarch64" ]] && ARCH="arm64"; \
    zypper --non-interactive install --no-recommends -- \
      patterns-base-selinux \
      container-selinux \
      k3s-selinux \
      audit

# Add system files
COPY framework/files/ /

# Enable SELinux (The security=selinux arg is default on Micro, not on Tumbleweed)
RUN sed -i "s/selinux=1/security=selinux selinux=1/g" /etc/elemental/bootargs.cfg
# Enforce SELinux
# RUN sed -i "s/enforcing=0/enforcing=1/g" /etc/elemental/bootargs.cfg

# Add elemental-register
COPY --from=register /usr/sbin/elemental-register /usr/sbin/elemental-register
COPY --from=register /usr/sbin/elemental-support /usr/sbin/elemental-support
# Add the elemental cli
COPY --from=toolkit /usr/bin/elemental /usr/bin/elemental

# Add the elemental-system-agent
ADD --chmod=0755 https://github.com/rancher/system-agent/releases/download/${RANCHER_SYSTEM_AGENT_VERSION}/rancher-system-agent-amd64 /usr/sbin/elemental-system-agent
RUN echo "${RANCHER_SYSTEM_AGENT_CHECKSUM} /usr/sbin/elemental-system-agent" | sha256sum -c -

# Enable essential services
RUN systemctl enable NetworkManager.service sshd elemental-register.timer

# This is for testing purposes, do not do this in production.
RUN echo "PermitRootLogin yes" > /etc/ssh/sshd_config.d/rootlogin.conf

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

# Update os-release file with some metadata
RUN echo TIMESTAMP="`date +'%Y%m%d%H%M%S'`" >> /etc/os-release && \
    echo GRUB_ENTRY_NAME=\"Elemental Dev\" >> /etc/os-release

# Rebuild initrd to setup dracut with the boot configurations
RUN elemental init --force elemental-rootfs,elemental-sysroot,grub-config,dracut-config,cloud-config-essentials,elemental-setup,boot-assessment

# Good for validation after the build
CMD ["/bin/bash"]
