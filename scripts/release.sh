#!/bin/sh
# Builds release tarballs: probeme-mothership-<ver>-linux-<arch>.tar.gz
# (binary + METRICS.md + systemd unit). Native arch always builds; the
# other arch builds when a matching cross toolchain is installed.
set -eu

cd "$(dirname "$0")/.."

VERSION=${1:?usage: release.sh <version>}
REPO=github.com/xXLegionBinFrogXx/probeme-mothership

native() {
	case "$(uname -m)" in
	x86_64) echo amd64 ;;
	aarch64 | arm64) echo arm64 ;;
	*) echo "" ;;
	esac
}

NATIVE=$(native)
rm -rf release
mkdir -p release

for arch in amd64 arm64; do
	CC=""
	if [ "$arch" != "$NATIVE" ]; then
		case $arch in
		amd64) CC=x86_64-linux-gnu-gcc ;;
		arm64) CC=aarch64-linux-gnu-gcc ;;
		esac
		if ! command -v "$CC" >/dev/null 2>&1; then
			echo "skip linux/$arch: $CC not installed"
			continue
		fi
	fi

	dir="release/probeme-mothership-$VERSION-linux-$arch"
	mkdir -p "$dir"

	echo "building linux/$arch (CC=${CC:-default cc})..."
	CC="$CC" GOOS=linux GOARCH="$arch" CGO_ENABLED=1 \
		go build -trimpath \
		-ldflags "-s -w -X $REPO/internal/buildinfo.Version=$VERSION" \
		-o "$dir/probeme-mothership" ./cmd/probeme-mothership

	cp README.md packaging/probeme-mothership.service "$dir/"
	tar -czf "release/probeme-mothership-$VERSION-linux-$arch.tar.gz" \
		-C release "probeme-mothership-$VERSION-linux-$arch"
	echo "  -> release/probeme-mothership-$VERSION-linux-$arch.tar.gz"
done
