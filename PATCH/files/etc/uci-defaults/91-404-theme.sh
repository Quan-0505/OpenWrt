#!/bin/sh
# Quan-0505/YAOF fork: default LuCI theme = footstrap
uci set luci.main.mediaurlbase='/luci-static/footstrap'
uci commit luci
exit 0
