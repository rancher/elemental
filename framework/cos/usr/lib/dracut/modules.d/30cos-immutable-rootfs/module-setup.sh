#!/bin/bash

# called by dracut
check() {
    require_binaries "$systemdutildir"/systemd || return 1
    return 255
}

# called by dracut 
depends() {
    echo systemd rootfs-block dm fs-lib
    return 0
}

# called by dracut
installkernel() {
    instmods overlay
}

# called by dracut
install() {
    declare moddir=${moddir}
    declare systemdutildir=${systemdutildir}
    declare systemdsystemunitdir=${systemdsystemunitdir}
    declare initdir="${initdir}"

    inst_multiple \
        mount mountpoint elemental sort findmnt rmdir findmnt rsync cut

    # Include utilities required for cos-setup services,
    # probably a devoted cos-setup dracut module makes sense
    inst_multiple -o \
        "$systemdutildir"/systemd-fsck partprobe sync udevadm lsblk sgdisk parted mkfs.ext2 mkfs.ext3 mkfs.ext4 mkfs.vfat mkfs.fat mkfs.xfs blkid e2fsck resize2fs mount xfs_growfs umount
    inst_hook cmdline 30 "${moddir}/parse-cos-cmdline.sh"
    inst_script "${moddir}/cos-generator.sh" \
        "${systemdutildir}/system-generators/dracut-cos-generator"
    inst_script "${moddir}/cos-mount-layout.sh" "/sbin/cos-mount-layout"
    inst_script "${moddir}/cos-loop-img.sh" "/sbin/cos-loop-img"
    inst_simple "${moddir}/cos-immutable-rootfs.service" \
        "${systemdsystemunitdir}/cos-immutable-rootfs.service"
    mkdir -p "${initdir}/${systemdsystemunitdir}/initrd-fs.target.requires"
    ln_r "../cos-immutable-rootfs.service" \
        "${systemdsystemunitdir}/initrd-fs.target.requires/cos-immutable-rootfs.service"
    ln_r "$systemdutildir"/systemd-fsck \
        "/sbin/systemd-fsck"
    dracut_need_initqueue
}
