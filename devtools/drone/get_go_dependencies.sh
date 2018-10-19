#!/bin/bash
set -e

CURRENT_DIR=$(pwd)
BASE_PACKAGE=github.com/CaliOpen/Caliopen
PACKAGE=$BASE_PACKAGE/src/backend/$PROG
export GOPATH=$GOPATH:/srv

echo "Getting dependencies for "$PROG

# Respect GOPATH conventions and package imports
mkdir -p /srv/src/$BASE_PACKAGE 
cp -a /srv/caliopen/. /srv/src/$BASE_PACKAGE && cd /srv/src

# Install depth for go dependencies
wget -q https://github.com/KyleBanks/depth/releases/download/v1.2.1/depth_1.2.1_linux_amd64 -O /usr/bin/depth
chmod +x /usr/bin/depth

# It first filters govendor packages, then non-Caliopen packages, seds out special chars,
# cuts path to get path from src directory and makes a list of the dependencies
export DEPS="$(depth $PACKAGE | \
grep -v CaliOpen/Caliopen/src/backend/vendor | \
grep CaliOpen | \
sort | \
sed -e 's/└//' -e 's/├//' -e 's/ *//' | \
uniq | \
cut -d'/' -f 6- | \
tr '\n' ' ')"

cd $CURRENT_DIR
