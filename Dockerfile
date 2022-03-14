FROM opensuse/leap:15.3 as base
RUN sed -i -s 's/^# rpm.install.excludedocs/rpm.install.excludedocs/' /etc/zypp/zypp.conf
RUN zypper ref

FROM quay.io/luet/base:0.22.7-1 as luet

FROM base AS build
ENV LUET_NOLOCK=true
ENV USER=root
RUN zypper in -y squashfs xorriso go1.16 upx busybox-static curl tar git gzip openssl-devel
COPY framework/files/etc/luet/luet.yaml /etc/luet/luet.yaml
COPY --from=luet /usr/bin/luet /usr/bin/luet
RUN luet install -y utils/helm
COPY go.mod go.sum /usr/src/
COPY cmd /usr/src/cmd
COPY pkg /usr/src/pkg
COPY scripts /usr/src/scripts
COPY chart /usr/src/chart
ARG IMAGE_TAG=latest
ARG IMAGE_REPO=norepo
RUN TAG=${IMAGE_TAG} REPO=${IMAGE_REPO} /usr/src/scripts/package-helm && \
    cp /usr/src/dist/artifacts/rancheros-operator-*.tgz /usr/src/dist/rancheros-operator-chart.tgz
RUN cd /usr/src && \
    CGO_ENABLED=0 go build -ldflags "-extldflags -static -s" -o /usr/sbin/ros-operator ./cmd/ros-operator && \
    upx /usr/sbin/ros-operator
RUN cd /usr/src && \
    go build -o /usr/sbin/ros-installer ./cmd/ros-installer && \
    upx /usr/sbin/ros-installer

FROM quay.io/luet/base:0.22.7-1 AS framework-build
COPY framework/files/etc/luet/luet.yaml /etc/luet/luet.yaml
ARG CACHEBUST
ENV LUET_NOLOCK=true
ENV USER=root

# We set the shell to /usr/bin/luet, as the base image doesn't have busybox, just luet
# and certificates to be able to correctly handle TLS requests.
SHELL ["/usr/bin/luet", "install", "-y", "--system-target", "/framework"]

# Each package we want to install needs a new line here
RUN meta/cos-core
RUN cloud-config/livecd
RUN cloud-config/recovery
RUN cloud-config/network

RUN meta/cos-verify
RUN utils/k9s
RUN utils/nerdctl
RUN selinux/rancher
RUN selinux/k3s
RUN utils/rancherd@0.0.1-alpha13-3
RUN utils/helm

FROM scratch AS framework
COPY --from=framework-build /framework/etc /etc
COPY --from=framework-build /framework/lib /lib
COPY --from=framework-build /framework/usr /usr
COPY --from=framework-build /framework/system /system
COPY --from=framework-build /framework/var/lib /var/lib
COPY --from=build /usr/src/dist/rancheros-operator-chart.tgz /usr/share/rancher/os2/
COPY framework/files/etc/luet/luet.yaml /etc/luet/luet.yaml
COPY --from=build /usr/sbin/ros-installer /usr/sbin/ros-installer
COPY --from=build /usr/sbin/ros-operator /usr/sbin/ros-operator
COPY framework/files/ /


FROM scratch as ros-operator
COPY --from=build /usr/sbin/ros-operator /usr/sbin/ros-operator

# Make OS image
FROM base as os
RUN zypper dup -y
RUN zypper in -y -- \
    apparmor-parser \
    avahi \
    bash-completion \
    conntrack-tools \
    coreutils \
    curl \
    device-mapper \
    dmidecode \
    dosfstools \
    dracut \
    e2fsprogs \
    ethtool \
    findutils \
    gawk \
    gptfdisk \
    glibc-locale-base \
    grub2-i386-pc \
    grub2-x86_64-efi \
    haveged \
    hdparm \
    iproute2 \
    iptables \
    iputils \
    issue-generator \
    jq \
    kernel-default \
    kernel-firmware-bnx2 \
    kernel-firmware-chelsio \
    kernel-firmware-i915 \
    kernel-firmware-intel \
    kernel-firmware-iwlwifi \
    kernel-firmware-liquidio \
    kernel-firmware-marvell \
    kernel-firmware-mediatek \
    kernel-firmware-mellanox \
    kernel-firmware-network \
    kernel-firmware-platform \
    kernel-firmware-qlogic \
    kernel-firmware-realtek \
    kernel-firmware-usb-network \
    -kubic-locale-archive \
    less \
    lshw \
    lsof \
    lsscsi \
    lvm2 \
    mdadm \
    multipath-tools \
    netcat-openbsd \
    nfs-utils \
    open-iscsi \
    open-vm-tools \
    openssh \
    parted \
    -perl \
    pciutils \
    pigz \
    procps \
    psmisc \
    python-azure-agent \
    qemu-guest-agent \
    rsync \
    squashfs \
    strace \
    sysstat \
    systemd \
    systemd-presets-branding-openSUSE \
    -systemd-presets-branding-MicroOS \
    systemd-sysvinit \
    tar \
    timezone \
    vim-small \
    which \
    zstd

# Copy in some local OS customizations
COPY opensuse/files /

ARG IMAGE_TAG=latest
RUN cat /etc/os-release.tmpl | env \
    "VERSION=${IMAGE_TAG}" \
    "VERSION_ID=$(echo ${IMAGE_TAG} | sed s/^v//)" \
    "PRETTY_NAME=RancherOS ${IMAGE_TAG}" \
    envsubst > /etc/os-release && \
    rm /etc/os-release.tmpl

# Starting from here are the lines needed for RancherOS to work

# IMPORTANT: Setup rancheros-release used for versioning/upgrade. The
# values here should reflect the tag of the image being built
ARG IMAGE_REPO=norepo
RUN echo "IMAGE_REPO=${IMAGE_REPO}"          > /usr/lib/rancheros-release && \
    echo "IMAGE_TAG=${IMAGE_TAG}"           >> /usr/lib/rancheros-release && \
    echo "IMAGE=${IMAGE_REPO}:${IMAGE_TAG}" >> /usr/lib/rancheros-release

# Copy in framework runtime
COPY --from=framework / /

# Rebuild initrd to setup dracut with the boot configurations
RUN mkinitrd && \
    # aarch64 has an uncompressed kernel so we need to link it to vmlinuz
    kernel=$(ls /boot/Image-* | head -n1) && \
    if [ -e "$kernel" ]; then ln -sf "${kernel#/boot/}" /boot/vmlinuz; fi

# Save some space
RUN zypper clean --all && \
    rm -rf /var/log/update* && \
    >/var/log/lastlog && \
    rm -rf /boot/vmlinux*

FROM scratch as default
COPY --from=os / /
