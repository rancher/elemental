ARG ELEMENTAL_OS_IMAGE

FROM ${ELEMENTAL_OS_IMAGE} as os

COPY manifest.yaml manifest.yaml
RUN elemental --debug --config-dir . build-iso -o /output -n "elemental-dev" dir:/

FROM busybox:stable

COPY --from=os /output /elemental-iso
