#!/usr/bin/env bash
set -Eeuo pipefail
trap 'printf "DaeNext build failed at line %s: %s\n" "$LINENO" "$BASH_COMMAND" >&2' ERR

# Run from the firmware repository root after OpenWrt toolchain/install.
ROOT=$(pwd)
CORE_REF=218bdf72b5be3e59494513c9805fe701dc0fed04
WEB_REF=b084f2f6f0062a12d6087ad5243bb4a436d37f45
TARGET=aarch64-unknown-linux-musl
DEST="$ROOT/openwrt/package/new/daed/prebuilt-data"

checkout_pinned() {
  local url=$1 ref=$2 dir=$3
  git init "$dir"
  git -C "$dir" remote add origin "$url"
  git -C "$dir" fetch --depth 1 origin "$ref"
  git -C "$dir" checkout --detach FETCH_HEAD
  test "$(git -C "$dir" rev-parse HEAD)" = "$ref"
}
checkout_pinned https://github.com/ksong008/DaeNext.git "$CORE_REF" "$ROOT/daenext-core"
checkout_pinned https://github.com/ksong008/DaedNext.git "$WEB_REF" "$ROOT/daenext-web"

source "$ROOT/SCRIPTS/daenext_toolchain.sh"
export STAGING_DIR="$ROOT/openwrt/staging_dir"
export PATH="$(dirname "$GCC"):$PATH"
export CARGO_TARGET_AARCH64_UNKNOWN_LINUX_MUSL_LINKER="$GCC"
export CC_aarch64_unknown_linux_musl="$GCC"
export CXX_aarch64_unknown_linux_musl="${PREFIX}g++"
export AR_aarch64_unknown_linux_musl="${PREFIX}ar"
# Bindgen runs on the x86 build host; explicitly give it musl target headers.
export BORING_BSSL_SYSROOT="$SYSROOT"
export BINDGEN_EXTRA_CLANG_ARGS_aarch64_unknown_linux_musl="--sysroot=$SYSROOT -isystem $TARGET_INCLUDE"
export RUSTFLAGS='-C target-cpu=generic -C target-feature=+crt-static'
export CARGO_BUILD_JOBS=2
export CARGO_PROFILE_RELEASE_LTO=thin
export CARGO_PROFILE_RELEASE_CODEGEN_UNITS=1
export DAE_DAEMON_VERSION="DaeNext 3.1.0 core=$CORE_REF target=$TARGET native-ebpf boringssl"
(
  cd "$ROOT/daenext-core"
  cargo build --locked --release --target "$TARGET" -p dae-daemon --bin daed
  git diff --exit-code -- Cargo.lock
)
BIN="$ROOT/daenext-core/target/$TARGET/release/daed"
file "$BIN"
"${PREFIX}readelf" -h "$BIN" | grep -q AArch64
if "${PREFIX}readelf" -l "$BIN" | grep -q INTERP; then
  echo 'Refusing a dynamically linked daemon in the OpenWrt image' >&2
  exit 1
fi
qemu-aarch64 "$BIN" --version | tee "$ROOT/daenext-version.txt"
grep -q "$CORE_REF" "$ROOT/daenext-version.txt"

(
  cd "$ROOT/daenext-web"
  test "$(node -p "require('./apps/web/package.json').version")" = 3.1.0
  pnpm install --frozen-lockfile
  pnpm --filter daed build
)
install -Dm755 "$BIN" "$DEST/usr/bin/daed"
"${PREFIX}strip" "$DEST/usr/bin/daed"
install -Dm755 "$ROOT/PATCH/daenext-r4s/daed.init" "$DEST/etc/init.d/daed"
mkdir -p "$DEST/etc/config" "$DEST/etc/daed" "$DEST/usr/share/daed/web"
printf "config daed 'main'\n\toption enabled '1'\n" > "$DEST/etc/config/daed"
cp -a "$ROOT/daenext-web/apps/web/dist/." "$DEST/usr/share/daed/web/"
for name in geoip geosite; do
  curl --fail --location --retry 3 "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/$name.dat" -o "$DEST/usr/share/daed/$name.dat"
done
python3 "$ROOT/SCRIPTS/verify_daenext_assets.py" "$DEST/usr/share/daed/web"
printf 'core=%s\nweb=%s\n' "$CORE_REF" "$WEB_REF" > "$DEST/usr/share/daed/source-revisions"
sha256sum "$DEST/usr/bin/daed" "$DEST/usr/share/daed/"*.dat >> "$DEST/usr/share/daed/source-revisions"
