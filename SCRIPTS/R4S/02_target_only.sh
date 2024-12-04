#!/bin/bash
clear

# 使用特定的优化
sed -i 's,-mcpu=generic,-march=armv8-a+crc+crypto,g' include/target.mk
#sed -i 's,kmod-r8169,kmod-r8168,g' target/linux/rockchip/image/armv8.mk

# Vermagic
target="rockchip/armv8"
latest_version="$(curl -s https://github.com/openwrt/openwrt/tags | grep -Eo "v[0-9\.]+\-*r*c*[0-9]*.tar.gz" | sed -n '/[2-9][4-9]/p' | sed -n 1p | sed 's/v//g' | sed 's/.tar.gz//g')"
wget -O profiles.json https://downloads.openwrt.org/releases/${latest_version}/targets/${target}/profiles.json

jq -r '.linux_kernel.vermagic' profiles.json | tr -d '\n' >.vermagic
sed -i -e 's/^\(.\).*vermagic$/\1cp $(TOPDIR)\/.vermagic $(LINUX_DIR)\/.vermagic/' include/kernel-defaults.mk

# kmod feed hack
wget -O sha256sums https://downloads.openwrt.org/releases/${latest_version}/targets/${target}/sha256sums
kmods_path=$(awk '$2 ~ /^\*kmods/ { match($2, /kmods[^ ]+/, arr); split(arr[0], path_parts, "/"); print path_parts[1] "/" path_parts[2]; exit }' "./sha256sums")
mkdir -p ./files/etc/opkg
echo "src/gz openwrt_kmods https://mirrors.pku.edu.cn/openwrt/releases/${latest_version}/targets/${target}/$kmods_path" > ./files/etc/opkg/customfeeds.conf

# 预配置一些插件
cp -rf ../PATCH/files ./files

find ./ -name *.orig | xargs rm -f
find ./ -name *.rej | xargs rm -f

#exit 0
