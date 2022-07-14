#!/bin/bash

function doLoopMount {
    local partdev
    local partname
    local dev

    # Iterate over current device labels
    for partdev in $(lsblk -ln -o path,type | grep part | cut -d" " -f1); do
        partname=$(basename "${partdev}")
        [ -e "/tmp/cosloop-${partname}" ] && continue
        > "/tmp/cosloop-${partname}" 

        # Ensure run system-fsck, at least, for the root partition
        systemd-fsck "${partdev}"

        # Only run systemd-fsck if root is already found
        [ "${found}" == "ok" ] && continue

        mount -t auto -o "${cos_root_perm}" "${partdev}" "${cos_state}" || continue
        if [ -f "${cos_state}/${cos_img}" ]; then

            dev=$(losetup --show -f "${cos_state}/${cos_img}")

            # attempt to run systemd-fsck on the loop device
            systemd-fsck "${dev}"

            found="ok"
        else
            umount "${cos_state}"
        fi
    done
}

type getarg > /dev/null 2>&1 || . /lib/dracut-lib.sh

PATH=/usr/sbin:/usr/bin:/sbin:/bin

declare cos_img=$1
declare cos_root_perm="ro"
declare cos_state="/run/initramfs/cos-state"
declare found=""

[ -z "${cos_img}" ] && exit 1

if getargbool 0 rd.cos.debugrw; then
    cos_root_perm="rw"
fi

ismounted "${cos_state}" && exit 0

mkdir -p "${cos_state}"

doLoopMount
if [ "${found}" == "ok" ]; then
    exit 0
fi

rm -r "${cos_state}"
exit 1
