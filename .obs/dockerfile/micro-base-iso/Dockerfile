# SPDX-License-Identifier: Apache-2.0
#!BuildName: build-base-iso-image
#!BuildTag: suse/sl-micro/%%SLMICRO_VERSION%%/base-iso-image:latest
#!BuildTag: suse/sl-micro/%%SLMICRO_VERSION%%/base-iso-image:%VERSION%
#!BuildTag: suse/sl-micro/%%SLMICRO_VERSION%%/base-iso-image:%VERSION%-%RELEASE%
#!BuildConstraint: hardware:disk:size unit=G 10
#!BuildConstraint: hardware:memory:size unit=G 16
#!UseOBSRepositories

ARG SLMICRO_VERSION

FROM suse/sl-micro/${SLMICRO_VERSION}/base-os-container:latest AS os
FROM suse/sl-micro/${SLMICRO_VERSION}/baremetal-os-container:latest AS builder

WORKDIR /iso

COPY manifest.yaml manifest.yaml
COPY --from=os / rootfs

# Prepare a busybox base
RUN mkdir -p /busybox && rpm --initdb --root /busybox
RUN zypper --installroot /busybox in --no-recommends -y busybox

# Version value is taken form the elemental repository tags
# Release value of this image and os image are unrelated
RUN elemental --debug --config-dir . build-iso -o /output -n "sl-micro-%%SLMICRO_VERSION%%-base-$(uname -m)-%VERSION%-Build%RELEASE%" dir:rootfs

# Only keep the ISO as a result and busybox
FROM scratch
COPY --from=builder /busybox /
COPY --from=builder /output /elemental-iso

ARG SLMICRO_VERSION
ARG BUILD_REPO=%%IMG_REPO%%
ARG IMAGE_REPO=$BUILD_REPO/suse/sl-micro/%%SLMICRO_VERSION%%/base-iso-image
ARG IMAGE_TAG=%VERSION%-%RELEASE%

# Define labels according to https://en.opensuse.org/Building_derived_containers
# labelprefix=com.suse.sl.micro
LABEL org.opencontainers.image.title="SUSE Linux Micro base-ISO"
LABEL org.opencontainers.image.description="Includes the SUSE Linux Micro base-ISO"
LABEL org.opencontainers.image.version="${IMAGE_TAG}"
LABEL org.opencontainers.image.url="https://github.com/rancher/elemental"
LABEL org.opencontainers.image.created="%BUILDTIME%"
LABEL org.opencontainers.image.vendor="SUSE LLC"
LABEL org.opensuse.reference="${IMAGE_REPO}:${IMAGE_TAG}"
LABEL org.openbuildservice.disturl="%DISTURL%"
LABEL com.suse.supportlevel="l3"
# endlabelprefix

# By default run a shell
ENTRYPOINT ["/usr/bin/busybox"]
CMD ["sh"]
