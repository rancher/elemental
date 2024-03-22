#!/bin/bash

run_checks() {
    systemctl is-active -q network-online.target
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
