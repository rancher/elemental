# SPDX-License-Identifier: Apache-2.0
#!BuildTag: suse/sle-micro-rancher-iso/%%SLEMICRO_VERSION%%:latest
#!BuildTag: suse/sle-micro-rancher-iso/%%SLEMICRO_VERSION%%:%VERSION%
#!BuildTag: suse/sle-micro-rancher-iso/%%SLEMICRO_VERSION%%:%VERSION%-%RELEASE%
#!BuildConstraint: hardware:disk:size unit=G 10
#!BuildConstraint: hardware:memory:size unit=G 16

ARG SLE_VERSION
ARG SLEMICRO_VERSION

FROM suse/sle-micro-rancher/$SLEMICRO_VERSION:latest AS os
FROM suse/sle-micro-rancher/$SLEMICRO_VERSION:latest AS builder

WORKDIR /iso

COPY manifest.yaml manifest.yaml
COPY --from=os / rootfs

# Version value is taken form the elemental repository tags
# Release value of this image and os image are unrelated
RUN elemental --debug --config-dir . build-iso -o /output -n "sle-micro-rancher.$(uname -m)-%VERSION%-Build%RELEASE%" dir:rootfs

# Only keep the ISO as a result
FROM bci/bci-busybox:$SLE_VERSION
COPY --from=builder /output /iso

ARG SLEMICRO_VERSION
ARG BUILD_REPO=%%IMG_REPO%%
ARG IMAGE_REPO=$BUILD_REPO/suse/sle-micro-rancher-iso/$SLEMICRO_VERSION

# Define labels according to https://en.opensuse.org/Building_derived_containers
# labelprefix=com.rancher.slem4r
LABEL org.opencontainers.image.title="SLE Micro for Rancher ISO"
LABEL org.opencontainers.image.description="Includes the SLE Micro for Rancher ISO"
LABEL org.opencontainers.image.version="%VERSION%"
LABEL org.opencontainers.image.url="https://github.com/rancher/elemental"
LABEL org.opencontainers.image.created="%BUILDTIME%"
LABEL org.opencontainers.image.vendor="SUSE LLC"
LABEL org.opensuse.reference=$IMAGE_REPO
LABEL org.openbuildservice.disturl="%DISTURL%"
LABEL com.suse.supportlevel="l3"
# endlabelprefix

# By default run a shell
ENTRYPOINT ["busybox"]
CMD ["sh"]
