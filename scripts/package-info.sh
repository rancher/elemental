#!/bin/sh
set -e

version() {
     rpm -qv --qf '%{VERSION}' $1
}

echo "elemental-cli:$(version elemental)"
echo "elemental-installer:$(version ros-installer)"
echo "elemental-toolkit:$(version os2-framework)"
