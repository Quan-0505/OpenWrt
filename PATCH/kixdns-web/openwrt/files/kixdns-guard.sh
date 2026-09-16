#!/bin/sh
# kixdns + kixdns-web watchdog (OpenWrt deployment reference).
#
# Every rule here corresponds to a failure that actually happened:
#
#   * kixdns must run with --debug and must NOT have RUST_LOG set. RUST_LOG is a
#     Rust log filter: RUST_LOG=info suppresses the DEBUG-level observer events,
#     so kixdns looks healthy while every dashboard counter stays 0.
#
#   * Never rotate the log. kixdns only appends and does not reopen its log on a
#     signal, so any rotation starts a fresh file and resets the collector's
#     counters, leaving kixdns's cumulative total permanently smaller than
#     dnsmasq's — which reads as a contradiction on the dashboard.
#
#   * Restart dnsmasq only when the upstream injection file had to be rebuilt,
#     because restarting it otherwise discards real client counters.
#
#   * When the binary is upgraded while the old process is still running, that
#     process keeps executing the unlinked inode: alive, listening, but running
#     the previous code, so the console never updates while the watchdog sees
#     "a process is running" and does nothing. Such a process is detected by
#     checking whether the pid holding the listening socket has a deleted
#     executable, and only it is restarted — never the healthy process.
# 设备时区 Asia/Shanghai (UTC+8)。musl 不读 /etc/TZ，子进程必须靠 TZ 环境变量。
TZ=CST-8
export TZ
K=/data/ufi-tools/kixdns
W=/data/ufi-tools/kixdns-web
INJ=/tmp/dnsmasq.d/99-kixdns.conf
LOG=/tmp/kixdns-guard.log
STATE=$W/state
KEEP_LOGS=2

# Reads /proc/net/tcp{,6} and echoes the inode of the socket listening on 8080.
listen_inode() {
    for f in /proc/net/tcp /proc/net/tcp6; do
        [ -r "$f" ] || continue
        awk 'NR > 1 && $4 == "0A" {
                 split($2, a, ":");
                 if (a[2] == "1F90") { print $10 }
             }' "$f"
    done
}

# Pids holding a listening socket on 8080 whose executable has been unlinked.
#
# "Listens on 8080" is the identifying property rather than the process name or
# a cmdline substring: a copy may be named anything, and its arguments need not
# mention kixdns-web (only the -state path does). Whoever owns the port is the
# console.
stale_web_pids() {
    INODES=$(listen_inode)
    [ -n "$INODES" ] || return 0
    for d in /proc/[0-9]*; do
        case "$(readlink $d/exe 2>/dev/null)" in
            *'(deleted)'*) ;;
            *) continue ;;
        esac
        for fd in $d/fd/*; do
            T=$(readlink "$fd" 2>/dev/null) || continue
            case "$T" in
                socket:\[*\])
                    INO=${T#socket:[}; INO=${INO%]}
                    for want in $INODES; do
                        if [ "$INO" = "$want" ]; then basename "$d"; break 3; fi
                    done ;;
            esac
        done
    done
}

pidof mihomo >/dev/null 2>&1 || { echo "[$(date '+%F %T')] mihomo absent, skip" >> "$LOG"; exit 0; }
mkdir -p "$STATE" "$K/logs"

# --- 1) restart a console that is running a replaced binary ---------------
STALE=$(stale_web_pids)
if [ -n "$STALE" ]; then
    echo "[$(date '+%F %T')] kixdns-web holds a replaced binary, restarting: $(echo $STALE | tr '\n' ' ')" >> "$LOG"
    for p in $STALE; do kill $p 2>/dev/null; done
    i=0
    while [ $i -lt 20 ]; do
        sleep 0.5
        LEFT=""
        for p in $STALE; do [ -d /proc/$p ] && LEFT="$LEFT $p"; done
        [ -z "$LEFT" ] && break
        i=$((i+1))
    done
    for p in $STALE; do [ -d /proc/$p ] && kill -9 $p 2>/dev/null; done
    sleep 1
fi

# --- 2) kixdns ------------------------------------------------------------
if ! pgrep -f 'kixdns/kixdns run' >/dev/null 2>&1; then
    echo "[$(date '+%F %T')] start kixdns --debug" >> "$LOG"
    nohup "$K/kixdns" run -c "$K/pipeline.json" --debug >> "$K/logs/kixdns.log" 2>&1 &
    i=0
    while [ $i -lt 16 ]; do
        sleep 0.5
        nslookup -type=A www.baidu.com 127.0.0.1:5354 >/dev/null 2>&1 && break
        i=$((i+1))
    done
fi

# --- 3) upstream injection (tmpfs, cleared on reboot) --------------------
INJ_CHANGED=0
if [ ! -f "$INJ" ] || ! grep -q '127.0.0.1#5354' "$INJ" 2>/dev/null; then
    echo "[$(date '+%F %T')] rebuild DNS upstream injection" >> "$LOG"
    mkdir -p /tmp/dnsmasq.d
    printf 'no-resolv\nserver=127.0.0.1#5354\n' > "$INJ"
    INJ_CHANGED=1
fi

# --- 4) kixdns-web --------------------------------------------------------
if [ -x "$W/kixdns-web" ] && ! pgrep -f 'kixdns-web -listen' >/dev/null 2>&1; then
    echo "[$(date '+%F %T')] start kixdns-web" >> "$LOG"
    nohup "$W/kixdns-web" \
        -listen 0.0.0.0:8080 \
        -bin "$K/kixdns" \
        -config "$K/pipeline.json" \
        -logdir "$K/logs" \
        -state "$STATE/stats.json" \
        -pidfile /var/run/kixdns.pid \
        -keep-logs "$KEEP_LOGS" \
        -log-cap-mb 0 \
        >> "$W/web.log" 2>&1 &
    i=0
    while [ $i -lt 20 ]; do
        sleep 0.5
        curl -s -o /dev/null --max-time 2 http://127.0.0.1:8080/api/status && break
        i=$((i+1))
    done
fi

# --- 5) dnsmasq, only if the injection was rebuilt -----------------------
if [ "$INJ_CHANGED" = "1" ]; then
    /etc/init.d/dnsmasq restart > /dev/null 2>&1
    i=0
    while [ $i -lt 16 ]; do
        sleep 0.5
        nslookup -type=A www.baidu.com 192.168.0.1 >/dev/null 2>&1 && break
        i=$((i+1))
    done
fi

exit 0
