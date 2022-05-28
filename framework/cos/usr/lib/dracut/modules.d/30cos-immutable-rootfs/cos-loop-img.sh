#!/bin/bash

function doLoopMount {
    local label
    local dev

    # Iterate over current device labels
    for dev in /dev/disk/by-label/*; do
        label=$(basename "${dev}")
        [ -e "/tmp/cosloop-${label}" ] && continue
        > "/tmp/cosloop-${label}" 

        mount -t auto -o "${cos_root_perm}" "/dev/disk/by-label/${label}" "${cos_state}" || continue
        if [ -f "${cos_state}/${cos_img}" ]; then

            # FSCHECK if cos_root_perm == "ro" on both
            if [ "$cos_root_perm" == "ro" ]; then
               systemd-fsck "/dev/disk/by-label/${label}"
            fi

            dev=$(losetup --show -f "${cos_state}/${cos_img}")

            # FSCHECK if cos_root_perm == "ro"
            if [ "$cos_root_perm" == "ro" ]; then
               systemd-fsck "$dev"
            fi

            exit 0
        else
            umount "${cos_state}"
        fi
    done
}

function dofsCheck {
    # Iterate over current partitions
    # As fs corruption could lead to partitions with no label, we scan here for all partitions found and we run systemd-fsck
    for dev in /dev/disk/by-partuuid/*; do
        partuuid=$(basename "${dev}")
        systemd-fsck "/dev/disk/by-partuuid/${partuuid}"
    done
}

type getarg > /dev/null 2>&1 || . /lib/dracut-lib.sh

PATH=/usr/sbin:/usr/bin:/sbin:/bin

declare cos_img=$1
declare cos_root_perm="ro"
declare cos_state="/run/initramfs/cos-state"

[ -z "${cos_img}" ] && exit 1

if getargbool 0 rd.cos.debugrw; then
    cos_root_perm="rw"
fi

ismounted "${cos_state}" && exit 0

mkdir -p "${cos_state}"

dofsCheck
doLoopMount

rm -r "${cos_state}"
exit 1
