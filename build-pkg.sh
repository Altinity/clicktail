#!/bin/bash

# Build deb or rpm packages for honeytail.
set -e

function usage() {
    echo "Usage: build-pkg.sh -v <version> -t <package_type>"
    exit 2
}

while getopts "v:t:" opt; do
    case "$opt" in
    v)
        version=$OPTARG
        ;;
    t)
        pkg_type=$OPTARG
        ;;
    esac
done

if [ -z "$version" ] || [ -z "$pkg_type" ]; then
    usage
fi

fpm -s dir -n honeytail \
    -m "Honeycomb <team@honeycomb.io>" \
    -p $GOPATH/bin \
    -v $version \
    -t $pkg_type \
    --pre-install=./preinstall \
    $GOPATH/bin/honeytail=/usr/bin/honeytail \
    ./honeytail.upstart=/etc/init/honeytail.conf \
    ./honeytail.service=/lib/systemd/system/honeytail.service \
    ./honeytail.conf=/etc/honeytail/honeytail.conf
