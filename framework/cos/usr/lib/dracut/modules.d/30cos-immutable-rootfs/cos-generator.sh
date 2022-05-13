#!/bin/bash

type getarg >/dev/null 2>&1 || . /lib/dracut-lib.sh

cos_unit="cos-immutable-rootfs.service"
cos_layout="/run/cos/cos-layout.env"

# Omit any immutable roofs module logic if disabled
if getargbool 0 rd.cos.disable; then
    exit 0
fi

[ -z "${root}" ] && root=$(getarg root=)

cos_root_perm="ro"
if getargbool 0 rd.cos.debugrw; then
    cos_root_perm="rw"
fi

oem_timeout=$(getargnum 120 1 1800 rd.cos.oemtimeout=)
oem_label=$(getarg rd.cos.oemlabel=)
cos_overlay=$(getarg rd.cos.overlay=)
[ -z "${cos_overlay}" ] && cos_overlay="tmpfs:20%"

GENERATOR_DIR="$2"
[ -z "$GENERATOR_DIR" ] && exit 1
[ -d "$GENERATOR_DIR" ] || mkdir "$GENERATOR_DIR"

if [ -n "${oem_label}" ]; then
    dev=$(dev_unit_name /dev/disk/by-label/${oem_label})
    {
        echo "[Unit]"
        echo "DefaultDependencies=no"
        echo "Before=cos-setup-rootfs.service"
        echo "Conflicts=initrd-switch-root.target"
        echo "[Mount]"
        echo "Where=/oem"
        echo "What=/dev/disk/by-label/${oem_label}"
        echo "Options=rw,suid,dev,exec,noauto,nouser,async"
    } > "$GENERATOR_DIR"/oem.mount

    if [ ! -e "$GENERATOR_DIR/cos-setup-rootfs.service.wants/oem.mount" ]; then
        mkdir -p "$GENERATOR_DIR"/cos-setup-rootfs.service.wants
        ln -s "$GENERATOR_DIR"/oem.mount \
            "$GENERATOR_DIR"/cos-setup-rootfs.service.wants/oem.mount
    fi

    mkdir -p "$GENERATOR_DIR/$dev.device.d"
    {
        echo "[Unit]"
        echo "Before=initrd-root-fs.target"
        echo "JobRunningTimeoutSec=${oem_timeout}"
    } > "$GENERATOR_DIR/$dev.device.d/timeout.conf"

    if [ ! -e "$GENERATOR_DIR/initrd-root-fs.target.wants/$dev.device" ]; then
        mkdir -p "$GENERATOR_DIR"/initrd-root-fs.target.wants
        ln -s "$GENERATOR_DIR"/"$dev".device \
            "$GENERATOR_DIR"/initrd-root-fs.target.wants/"$dev".device
    fi
fi

case "${cos_overlay}" in
    UUID=*) \
        cos_overlay="block:/dev/disk/by-uuid/${cos_overlay#UUID=}"
    ;;
    LABEL=*) \
        cos_overlay="block:/dev/disk/by-label/${cos_overlay#LABEL=}"
    ;;
esac

cos_mounts=()
for mount in $(getargs rd.cos.mount=); do
    case "${mount}" in
        UUID=*) \
            mount="/dev/disk/by-uuid/${mount#UUID=}"
        ;;
        LABEL=*) \
            mount="/dev/disk/by-label/${mount#LABEL=}"
        ;;
    esac
    cos_mounts+=("${mount}")
done

mkdir -p "/run/systemd/system/${cos_unit}.d"
{
    echo "[Service]"
    echo "Environment=\"cos_mounts=${cos_mounts[@]}\""
    echo "Environment=\"cos_overlay=${cos_overlay}\""
    echo "Environment=\"cos_root_perm=${cos_root_perm}\""
    echo "EnvironmentFile=${cos_layout}"
} > "/run/systemd/system/${cos_unit}.d/override.conf"

case "${root}" in
    LABEL=*) \
        root="${root//\//\\x2f}"
        root="/dev/disk/by-label/${root#LABEL=}"
        rootok=1 ;;
    UUID=*) \
        root="/dev/disk/by-uuid/${root#UUID=}"
        rootok=1 ;;
    /dev/*) \
        rootok=1 ;;
esac

[ "${rootok}" != "1" ] && exit 0

dev=$(dev_unit_name "${root}")
{
    echo "[Unit]"
    echo "Before=initrd-root-fs.target"
    echo "DefaultDependencies=no"
    echo "[Mount]"
    echo "Where=/sysroot"
    echo "What=${root}"
    echo "Options=${cos_root_perm},suid,dev,exec,auto,nouser,async"
} > "$GENERATOR_DIR"/sysroot.mount

if [ ! -e "$GENERATOR_DIR/initrd-root-fs.target.requires/sysroot.mount" ]; then
    mkdir -p "$GENERATOR_DIR"/initrd-root-fs.target.requires
    ln -s "$GENERATOR_DIR"/sysroot.mount \
        "$GENERATOR_DIR"/initrd-root-fs.target.requires/sysroot.mount
fi

mkdir -p "$GENERATOR_DIR/$dev.device.d"
{
    echo "[Unit]"
    echo "JobTimeoutSec=300"
    echo "JobRunningTimeoutSec=300"
} > "$GENERATOR_DIR/$dev.device.d/timeout.conf"
