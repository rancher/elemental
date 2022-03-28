FROM opensuse/leap:15.3 as base
RUN sed -i -s 's/^# rpm.install.excludedocs/rpm.install.excludedocs/' /etc/zypp/zypp.conf
RUN zypper ref

# This target downloads the rancheros operator and makes it available to the framework target
FROM alpine/helm:3.8.1 as helm
ARG CHART_REPO=https://rancher-sandbox.github.io/rancheros-operator
# ">0.0.0-0" means latest including pre-release versions
ARG CHART_VERSION=">0.0.0-0"
RUN helm repo add rancheros $CHART_REPO
RUN mkdir /usr/chart
RUN helm pull rancheros/rancheros-operator -d /usr/chart/ --version $CHART_VERSION
# naming convention is discarded on building the framework image, look into this
RUN mv /usr/chart/rancheros-operator-*.tgz /usr/chart/rancheros-operator-chart.tgz

# this target builds the ros-installer binary.
FROM base AS ros-installer
RUN zypper in -y openssl-devel gcc go1.16
WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY cmd /src/cmd
COPY pkg /src/pkg
RUN go build -o /usr/sbin/ros-installer ./cmd/ros-installer

# This installs the cos packages that we need
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

# This copies from other images the necessary files
FROM scratch AS framework
COPY --from=framework-build /framework/etc /etc
COPY --from=framework-build /framework/lib /lib
COPY --from=framework-build /framework/usr /usr
COPY --from=framework-build /framework/system /system
COPY --from=framework-build /framework/var/lib /var/lib
COPY --from=ros-installer /usr/sbin/ros-installer /usr/sbin/ros-installer
# This is used by framework/files/usr/sbin/ros-operator-install
# name needs to be exactly rancheros-operator-chart.tgz at the path otherwise it wont load it
COPY --from=helm /usr/chart/rancheros-operator-chart.tgz /usr/share/rancher/os2
# This adds our local overrides into the framework image
COPY framework/files/etc/luet/luet.yaml /etc/luet/luet.yaml
COPY framework/files/ /

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
