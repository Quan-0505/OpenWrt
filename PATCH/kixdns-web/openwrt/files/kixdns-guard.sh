#!/bin/sh
# kixdns + kixdns-web watchdog (OpenWrt deployment reference).
#
# Three non-obvious traps encoded here:
#   1. kixdns must run with --debug and must NOT have RUST_LOG set. RUST_LOG is a
#      Rust log filter: RUST_LOG=info suppresses the DEBUG-level observer events,
#      so kixdns looks healthy while every dashboard counter stays 0.
#   2. Do NOT rotate the log when starting kixdns. kixdns only appends and does
#      not reopen its log on a signal, so any rotation starts a fresh file and
#      resets the collector's counters — leaving kixdns's cumulative total
#      permanently smaller than dnsmasq's, which reads as a contradiction.
#   3. Rotate only when the injection file had to be rebuilt, since dnsmasq must
#      be restarted then anyway; restarting it otherwise discards real client
#      counters for no reason.
K=/data/ufi-tools/kixdns
W=/data/ufi-tools/kixdns-web
INJ=/tmp/dnsmasq.d/99-kixdns.conf
LOG=/tmp/kixdns-guard.log
STATE=$W/state
KEEP_LOGS=2

pidof mihomo >/dev/null 2>&1 || { echo "[$(date '+%F %T')] mihomo absent, skip" >> "$LOG"; exit 0; }
mkdir -p "$STATE" "$K/logs"

if ! pgrep -f 'kixdns/kixdns run' >/dev/null 2>&1; then
    echo "[$(date '+%F %T')] start kixdns --debug" >> "$LOG"
    nohup "$K/kixdns" run -c "$K/pipeline.json" --debug >> "$K/logs/kixdns.log" 2>&1 &
    i=0; while [ $i -lt 16 ]; do sleep 0.5; nslookup -type=A www.baidu.com 127.0.0.1:5354 >/dev/null 2>&1 && break; i=$((i+1)); done
fi

INJ_CHANGED=0
if [ ! -f "$INJ" ] || ! grep -q "127.0.0.1#5354" "$INJ" 2>/dev/null; then
    mkdir -p /tmp/dnsmasq.d
    printf 'no-resolv\nserver=127.0.0.1#5354\n' > "$INJ"
    INJ_CHANGED=1
fi

if [ -x "$W/kixdns-web" ] && ! pgrep -f 'kixdns-web -listen' >/dev/null 2>&1; then
    echo "[$(date '+%F %T')] start kixdns-web" >> "$LOG"
    nohup "$W/kixdns-web" -listen 0.0.0.0:8080 -bin "$K/kixdns" -config "$K/pipeline.json" \
        -logdir "$K/logs" -state "$STATE/stats.json" -pidfile /var/run/kixdns.pid \
        -keep-logs "$KEEP_LOGS" -log-cap-mb 0 >> "$W/web.log" 2>&1 &
fi

if [ "$INJ_CHANGED" = "1" ]; then
    /etc/init.d/dnsmasq restart > /dev/null 2>&1
    i=0; while [ $i -lt 16 ]; do sleep 0.5; nslookup -type=A www.baidu.com 192.168.0.1 >/dev/null 2>&1 && break; i=$((i+1)); done
fi
exit 0
