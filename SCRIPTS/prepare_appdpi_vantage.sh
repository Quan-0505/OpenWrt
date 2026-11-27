#!/usr/bin/env bash
set -euo pipefail
mode=${1:-prepare}
if [[ "$mode" == config ]]; then
    sed -i '/^CONFIG_PACKAGE_luci-.*footstrap.*=/d' .config
    for package in appdpi appdpi-rulepack appdpi-cdn appdpi-geoip-ip2region-v4 appdpi-geoip-ip2region-v6 appdpi-geoip-qqwry appdpi-geoip-zxipv6wry luci-app-appdpi luci-i18n-appdpi-zh-cn luci-theme-vantage luci-i18n-vantage-theme-zh-cn; do
        sed -i "/^CONFIG_PACKAGE_${package}=/d; /^# CONFIG_PACKAGE_${package} is not set$/d" .config
        printf 'CONFIG_PACKAGE_%s=y\n' "$package" >> .config
    done
    exit 0
fi
source_dir="${APPDPI_SOURCE_DIR:-$(cd .. && pwd)/panabit-appdpi-source}"
theme_dir="${VANTAGE_SOURCE_DIR:-$(cd .. && pwd)/vantage-source}"
fetch_source() {
    local directory=$1 url=$2 revision=$3
    if [[ ! -d "$directory/.git" ]]; then git init -q "$directory"; git -C "$directory" remote add origin "$url"; fi
    git -C "$directory" fetch -q --depth 1 origin "$revision"
    git -C "$directory" checkout -q --detach FETCH_HEAD
    test "$(git -C "$directory" rev-parse HEAD)" = "$revision"
}
if [[ -d "$source_dir/.git" ]] && [[ "$(git -C "$source_dir" rev-parse HEAD)" = 27a22fc0237eccbc83febb4077a0b56e1da7526f ]]; then
    :
else
    fetch_source "$source_dir" https://github.com/SuLingGG/panabit-appdpi.git 27a22fc0237eccbc83febb4077a0b56e1da7526f
fi
if [[ ! -d "$theme_dir/.git" ]] || [[ "$(git -C "$theme_dir" rev-parse HEAD)" != 17d46e2dc6d6de2e16b08777acc15bb315f4bb34 ]]; then
    fetch_source "$theme_dir" https://github.com/anothaDev/luci-theme-vantage.git 17d46e2dc6d6de2e16b08777acc15bb315f4bb34
fi
# The engine download must match the package definitions and bundled rulepack.
sed -i 's/^PKG_SOURCE_VERSION:=HEAD$/PKG_SOURCE_VERSION:=27a22fc0237eccbc83febb4077a0b56e1da7526f/' "$source_dir/openwrt/appdpi/Makefile"
test -s "$source_dir/rules/panabit.rulepack"
mkdir -p package
ln -sfn "$source_dir/openwrt" package/appdpi
ln -sfn "$theme_dir/luci-theme-vantage" package/luci-theme-vantage

sh "$source_dir/scripts/gen-cdn-data.sh" "$source_dir/openwrt/appdpi-cdn/files"
