#!/bin/bash
clear

# 使用专属优化
sed -i 's,-mcpu=generic,-march=armv8-a,g' include/target.mk

# 交换 LAN/WAN 口
sed -i 's,"eth1" "eth0","eth0" "eth1",g' target/linux/rockchip/armv8/base-files/etc/board.d/02_network
sed -i "s,'eth1' 'eth0','eth0' 'eth1',g" target/linux/rockchip/armv8/base-files/etc/board.d/02_network

# remove LRNG for 3328
rm -f target/linux/generic/hack-6.6/696*

#Vermagic
latest_version="$(curl -s https://github.com/openwrt/openwrt/tags | grep -Eo "v[0-9\.]+\-*r*c*[0-9]*.tar.gz" | sed -n '/[2-9][4-9]/p' | sed -n 1p | sed 's/v//g' | sed 's/.tar.gz//g')"
wget https://downloads.openwrt.org/releases/${latest_version}/targets/rockchip/armv8/profiles.json
jq -r '.linux_kernel.vermagic' profiles.json >.vermagic
sed -i -e 's/^\(.\).*vermagic$/\1cp $(TOPDIR)\/.vermagic $(LINUX_DIR)\/.vermagic/' include/kernel-defaults.mk

# 预配置一些插件
cp -rf ../PATCH/files ./files

find ./ -name *.orig | xargs rm -f
find ./ -name *.rej | xargs rm -f

# 使用 LZ4HC 压缩算法，优化设备性能
# 也支持CONFIG_SQUASHFS_ZSTD=y
if patch -p1 < ../PATCH/squashfs4_add_zstd_lz4_support.patch; then
  rm -rf ./tools/squashfs4/patches/
  echo '
CONFIG_SQUASHFS_XZ=n
CONFIG_SQUASHFS_LZ4=y
CONFIG_LZ4_DECOMPRESS=y
' | tee -a ./target/linux/generic/config-6.6 ./target/linux/rockchip/armv8/config-6.6 > /dev/null
else
  echo "squashfs 补丁应用失败，跳过"
fi
#exit 0
