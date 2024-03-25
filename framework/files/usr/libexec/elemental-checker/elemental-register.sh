#!/bin/bash

run_checks() {
    # Check first if it is installed
    rpm -q --quiet elemental-register
    test $? -ne 0 && return

    # ignore if elemental-register is not enabled
    systemctl is-enabled -q elemental-register.timer
    test $? -ne 0 && return

    systemctl is-active -q elemental-register.service
    test $? -ne 0 && exit 1
}

case "$1" in
    check)
        run_checks
        ;;
    *)
        echo "Usage: $0 {check}"
        exit 1
        ;;
esac

exit 0
