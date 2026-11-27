#!/bin/sh
# Quan-0505/YAOF fork: default LuCI theme = vantage
uci set luci.main.mediaurlbase='/luci-static/vantage'
uci commit luci
exit 0
