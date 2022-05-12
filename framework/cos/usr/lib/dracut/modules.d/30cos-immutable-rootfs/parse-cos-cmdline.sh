#!/bin/bash
# immutable root is specified with
# rd.cos.mount=LABEL=<vol_label>:<mountpoint>
# rd.cos.mount=UUID=<vol_uuid>:<mountpoint>
# rd.cos.overlay=tmpfs:<size>
# rd.cos.overlay=LABEL=<vol_label>
# rd.cos.overlay=UUID=<vol_uuid>
# rd.cos.oemtimeout=<seconds>
# rd.cos.oemlabel=<vol_label>
# rd.cos.debugrw
# rd.cos.disable
# cos-img/filename=/cOS/active.img

type getarg >/dev/null 2>&1 || . /lib/dracut-lib.sh

if getargbool 0 rd.cos.disable; then
    return 0
fi

cos_img=$(getarg cos-img/filename=)
[ -z "${cos_img}" ] && return 0
[ -z "${root}" ] && root=$(getarg root=)

cos_root_perm="ro"
if getargbool 0 rd.cos.debugrw; then
    cos_root_perm="rw"
fi

case "${root}" in
    LABEL=*) \
        root="${root//\//\\x2f}"
        root="/dev/disk/by-label/${root#LABEL=}"
        rootok=1 ;;
    UUID=*) \
        root="/dev/disk/by-uuid/${root#UUID=}"
        rootok=1 ;;
    /dev/*) \
        root="${root}"
        rootok=1 ;;
esac

[ "${rootok}" != "1" ] && return 0

info "root device set to root=${root}"

wait_for_dev -n "${root}"
/sbin/initqueue --settled --unique /sbin/cos-loop-img "${cos_img}"

return 0
