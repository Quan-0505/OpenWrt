#!/bin/sh
# daed(DaeNext/Rust eBPF 透明代理) 需要看到客户端转发流量。
# nftables flowtable(flow offload) 会把已建立连接走快速路径，绕过 tc/eBPF 钩子，
# 导致 daed 无法拦截 -> 表现为"代理看起来没生效/直连被墙"。
# 因此默认关闭软件与硬件 flow offloading。
uci -q get firewall.@defaults[0] >/dev/null 2>&1 || exit 0

uci -q set firewall.@defaults[0].flow_offloading='0'
uci -q set firewall.@defaults[0].flow_offloading_hw='0'
uci -q commit firewall

exit 0
