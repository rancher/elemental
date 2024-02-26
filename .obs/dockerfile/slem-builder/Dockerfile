# SPDX-License-Identifier: Apache-2.0
# Define the names/tags of the container
#!BuildName: SLE-Micro-builder-container
#!BuildTag: suse/sle-micro/sle-micro-builder-container-%%SLEMICRO_VERSION%%:latest
#!BuildTag: suse/sle-micro/sle-micro-builder-container-%%SLEMICRO_VERSION%%:1.0
#!BuildTag: suse/sle-micro/sle-micro-builder-container-%%SLEMICRO_VERSION%%:1.0-%RELEASE%
#!BuildConstraint: hardware:disk:size unit=G 8
#

ARG SLE_VERSION
FROM suse/sle15:$SLE_VERSION as host

MAINTAINER SUSE LLC (https://www.suse.com/)

RUN mkdir /builderimage

RUN rpm --initdb --root /builderimage

RUN zypper --installroot /builderimage in --no-recommends -y filesystem

# make derived containers possible
RUN zypper --installroot /builderimage in --no-recommends -y zypper

FROM scratch as builderimage

COPY --from=host /builderimage /

ARG SLEMICRO_VERSION
ARG BUILD_REPO=%%IMG_REPO%%
ARG IMAGE_REPO=$BUILD_REPO/suse/sle-micro/sle-micro-builder-container-$SLEMICRO_VERSION
ARG IMAGE_TAG=1.0-%RELEASE%

# Define labels according to https://en.opensuse.org/Building_derived_containers
# labelprefix=com.suse.sle.micro
LABEL org.opencontainers.image.title="SLE Micro builder Container"
LABEL org.opencontainers.image.description="Build dependency only - do not ship"
LABEL org.opencontainers.image.version="${IMAGE_TAG}"
LABEL org.opencontainers.image.url="https://www.suse.com/products/micro/"
LABEL org.opencontainers.image.created="%BUILDTIME%"
LABEL org.opencontainers.image.vendor="SUSE LLC"
LABEL org.opencontainers.image.source="%SOURCEURL%"
LABEL org.opensuse.reference="${IMAGE_REPO}:${IMAGE_TAG}"
LABEL org.openbuildservice.disturl="%DISTURL%"
LABEL com.suse.supportlevel="none"
LABEL com.suse.eula="sle-eula"
LABEL com.suse.release-stage="beta"
# endlabelprefix
