# SPDX-License-Identifier: Apache-2.0
#!BuildTag: rancher/elemental-teal-iso:%VERSION%
#!BuildTag: rancher/elemental-teal-iso:%VERSION%-%RELEASE%
#!BuildConstraint: hardware:disk:size unit=G 10
#!BuildConstraint: hardware:memory:size unit=G 16

ARG SLE_VERSION

FROM rancher/elemental-teal:latest AS os
FROM rancher/elemental-builder-image:latest AS builder

WORKDIR /iso

COPY manifest.yaml manifest.yaml
COPY --from=os / rootfs

# Version value is taken form the elemental repository tags
# Release value of this image and os image are unrelated
RUN elemental --debug --config-dir . build-iso -o /output -n "elemental-teal.$(uname -m)-%VERSION%-Build%RELEASE%" dir:rootfs

# Only keep the ISO as a result
FROM bci/bci-busybox:$SLE_VERSION
COPY --from=builder /output /elemental-iso

ARG BUILD_REPO=registry.suse.com
ARG IMAGE_REPO=$BUILD_REPO/rancher/elemental-teal

# Define labels according to https://en.opensuse.org/Building_derived_containers
# labelprefix=com.rancher.elemental
LABEL org.opencontainers.image.title="Rancher Elemental Teal ISO"
LABEL org.opencontainers.image.description="Includes the Elemental Teal ISO"
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
