#!/bin/bash

# Build deb or rpm packages for clicktail.
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

fpm -s dir -n clicktail \
    -m "Support <support@altinity.com>" \
    -p $GOPATH/bin \
    -v $version \
    -t $pkg_type \
    --pre-install=./preinstall \
    $GOPATH/bin/clicktail=/usr/bin/clicktail \
    ./clicktail.upstart=/etc/init/clicktail.conf \
    ./clicktail.service=/lib/systemd/system/clicktail.service \
    ./clicktail.conf=/etc/clicktail/clicktail-example.conf
