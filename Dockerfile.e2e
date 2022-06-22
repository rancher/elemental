ARG BASE_IMAGE=registry.opensuse.org/home/kwk/elemental/images/sle_15_sp3/rancher/rancher-node-image/5.2
FROM $BASE_IMAGE

# Framework files
COPY framework/cos/ /
COPY framework/files/ /

# Copy in some local OS customizations
COPY system/files /

ARG IMAGE_TAG=latest
RUN cat /etc/os-release.tmpl | env \
    "VERSION=${IMAGE_TAG}" \
    "VERSION_ID=$(echo ${IMAGE_TAG} | sed s/^v//)" \
    "PRETTY_NAME=RancherOS ${IMAGE_TAG}" \
    envsubst > /etc/os-release && \
    rm /etc/os-release.tmpl

# IMPORTANT: Setup rancheros-release used for versioning/upgrade. The
# values here should reflect the tag of the image being built
ARG IMAGE_REPO=norepo
RUN echo "IMAGE_REPO=${IMAGE_REPO}"          > /usr/lib/rancheros-release && \
    echo "IMAGE_TAG=${IMAGE_TAG}"           >> /usr/lib/rancheros-release && \
    echo "IMAGE=${IMAGE_REPO}:${IMAGE_TAG}" >> /usr/lib/rancheros-release

# Rebuild initrd to setup dracut with the boot configurations
RUN mkinitrd && \
    # aarch64 has an uncompressed kernel so we need to link it to vmlinuz
    kernel=$(ls /boot/Image-* | head -n1) && \
    if [ -e "$kernel" ]; then ln -sf "${kernel#/boot/}" /boot/vmlinuz; fi

# Save some space
RUN rm -rf /var/log/update* && \
    >/var/log/lastlog && \
    rm -rf /boot/vmlinux*