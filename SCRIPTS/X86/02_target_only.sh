#!/bin/bash

sed -i 's/O2/O2 -march=x86-64-v2/g' include/target.mk

# libsodium
sed -i 's,no-mips16 no-lto,no-mips16,g' feeds/packages/libs/libsodium/Makefile

echo '#!/bin/sh
# Put your custom commands here that should be executed once
# the system init finished. By default this file does nothing.

if ! grep "Default string" /tmp/sysinfo/model > /dev/null; then
    echo should be fine
else
    echo "Generic PC" > /tmp/sysinfo/model
fi

status=$(cat /sys/devices/system/cpu/intel_pstate/status)

if [ "$status" = "passive" ]; then
    echo "active" | tee /sys/devices/system/cpu/intel_pstate/status
fi

exit 0
'> ./package/base-files/files/etc/rc.local

# Vermagic
target="x86/64"
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

exit 0
