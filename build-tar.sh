#!/bin/bash

machine=`uname -m | tr A-Z a-z`
os=`uname -s | tr A-Z a-z`
stamp=`date '+%Y%m%d_%H%M'`

tar czf honeytail.${os}_${machine}.${stamp}.tar.gz honeytail
