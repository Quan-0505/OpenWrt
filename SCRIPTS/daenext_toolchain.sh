# Sourced by build_daenext_r4s.sh; ROOT is the firmware checkout.
# Shell globs also traverse toolchain directories that are symbolic links.
GCC=
for candidate in "$ROOT"/openwrt/staging_dir/toolchain-*/bin/aarch64-openwrt-linux-musl-gcc; do
  if [ -x "$candidate" ]; then
    GCC="$candidate"
    break
  fi
done
if [ -z "$GCC" ]; then
  echo 'DaeNext: no executable aarch64 musl GCC in staging_dir/toolchain-*/bin' >&2
  exit 1
fi
PREFIX=${GCC%gcc}
TOOLCHAIN_DIR=$(cd "$(dirname "$GCC")/.." && pwd -P)
SYSROOT=$("$GCC" -print-sysroot)
# OpenWrt GCC can use configured --with-headers/specs without advertising a
# sysroot through -print-sysroot. Its musl headers then live at toolchain/include.
if [ -z "$SYSROOT" ] || [ ! -d "$SYSROOT" ]; then
  SYSROOT="$TOOLCHAIN_DIR"
fi
if [ -f "$SYSROOT/include/stdio.h" ]; then
  TARGET_INCLUDE="$SYSROOT/include"
elif [ -f "$SYSROOT/usr/include/stdio.h" ]; then
  TARGET_INCLUDE="$SYSROOT/usr/include"
elif [ -f "$TOOLCHAIN_DIR/include/stdio.h" ]; then
  SYSROOT="$TOOLCHAIN_DIR"
  TARGET_INCLUDE="$TOOLCHAIN_DIR/include"
else
  echo "DaeNext: musl headers missing; GCC=$GCC sysroot=$SYSROOT toolchain=$TOOLCHAIN_DIR" >&2
  exit 1
fi
for tool in g++ ar readelf strip; do
  if [ ! -x "${PREFIX}${tool}" ]; then
    echo "DaeNext: required tool missing: ${PREFIX}${tool}" >&2
    exit 1
  fi
done
printf 'DaeNext compiler: %s\nDaeNext sysroot: %s\nDaeNext target headers: %s\n' "$GCC" "$SYSROOT" "$TARGET_INCLUDE"
