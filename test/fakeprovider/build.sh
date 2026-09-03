#!/bin/sh
# Builds test/fakeprovider/build/libprobeme_{fake,broken}.so against the
# installed probeme.h. Falls back to ~/.local/lib/pkgconfig when
# PKG_CONFIG_PATH is unset (cmake --install --prefix ~/.local default).
set -eu
cd "$(dirname "$0")"

: "${PKG_CONFIG_PATH:=$HOME/.local/lib/pkgconfig}"
export PKG_CONFIG_PATH

CFLAGS=$(pkg-config --cflags probeme)

mkdir -p build
cc $CFLAGS -O2 -fPIC -shared -o build/libprobeme_fake.so fake.c
cc -O2 -fPIC -shared -o build/libprobeme_broken.so broken.c
echo "fake provider built: $(pwd)/build/libprobeme_fake.so"
