#!/bin/sh
# kixdns + kixdns-web watchdog (deployment reference for OpenWrt).
#
# Two non-obvious traps this script encodes:
#   1. kixdns must run with --debug, and must NOT have RUST_LOG set. RUST_LOG is
#      a Rust log filter: RUST_LOG=info suppresses the DEBUG-level observer
#      events, so kixdns appears to run fine while every counter stays 0.
#   2. Rotate the log before starting kixdns. kixdns appends to the same file
#      across restarts, so a collector holding a byte offset would resume in the
#      middle of the new session and its counters would appear frozen. Giving
#      each session its own file makes the offset unambiguous.
K=/data/ufi-tools/kixdns
W=/data/ufi-tools/kixdns-web
INJ=/tmp/dnsmasq.d/99-kixdns.conf
LOG=/tmp/kixdns-guard.log
STATE=$W/state
KEEP_LOGS=2   # rotated files kept; active + KEEP_LOGS = days retained

pidof mihomo >/dev/null 2>&1 || { echo "[$(date '+%F %T')] mihomo absent, skip" >> "$LOG"; exit 0; }
mkdir -p "$STATE" "$K/logs"

if ! pgrep -f 'kixdns/kixdns run' >/dev/null 2>&1; then
    if [ -s "$K/logs/kixdns.log" ]; then
        mv "$K/logs/kixdns.log" "$K/logs/kixdns-$(date +%Y%m%d-%H%M%S).log"
    fi
    ls -1t "$K"/logs/kixdns-*.log 2>/dev/null | tail -n +$((KEEP_LOGS + 1)) | while read f; do rm -f "$f"; done
    nohup "$K/kixdns" run -c "$K/pipeline.json" --debug >> "$K/logs/kixdns.log" 2>&1 &
    i=0; while [ $i -lt 16 ]; do sleep 0.5; nslookup -type=A www.baidu.com 127.0.0.1:5354 >/dev/null 2>&1 && break; i=$((i+1)); done
fi

if [ ! -f "$INJ" ] || ! grep -q "127.0.0.1#5354" "$INJ" 2>/dev/null; then
    mkdir -p /tmp/dnsmasq.d
    printf 'no-resolv\nserver=127.0.0.1#5354\n' > "$INJ"
    /etc/init.d/dnsmasq restart > /dev/null 2>&1
fi

if [ -x "$W/kixdns-web" ] && ! pgrep -f 'kixdns-web -listen' >/dev/null 2>&1; then
    nohup "$W/kixdns-web" -listen 0.0.0.0:8080 -bin "$K/kixdns" -config "$K/pipeline.json" \
        -logdir "$K/logs" -state "$STATE/stats.json" -pidfile /var/run/kixdns.pid \
        -keep-logs "$KEEP_LOGS" -log-cap-mb 24 >> "$W/web.log" 2>&1 &
fi
exit 0
