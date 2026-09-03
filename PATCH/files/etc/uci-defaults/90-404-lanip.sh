#!/bin/sh
# Quan-0505/YAOF fork: default LAN is 192.168.2.1 (avoid clash with upstream router's 192.168.1.1)
uci -q get network.lan >/dev/null || {
	uci set network.lan=interface
	uci set network.lan.proto='static'
	uci set network.lan.device='br-lan'
}
uci set network.lan.ipaddr='192.168.2.1'
uci set network.lan.netmask='255.255.255.0'
uci commit network
exit 0
