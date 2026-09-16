# kixdns-web

An AdGuardHome-style web console for [kixdns](https://github.com/olicesx/kixdns).

kixdns itself ships no management interface and no metrics endpoint — it only
exposes `run` and `convert-geo-ip`. Its **only** source of observability is the
structured event stream emitted by its built-in `TracingObserver` when started
with `--debug`. This console parses that stream and presents it as a dashboard.

## What it does

| Tab | Contents |
|---|---|
| 仪表盘 | query count, cache hit ratio, end-to-end latency, slow (>500 ms) queries, hourly trend, top domains / clients / qtypes / upstreams / rcodes |
| 查询日志 | tail of any log file in the log directory, with 5 s auto-refresh |
| Pipeline 配置 | edit `pipeline.json` in the browser; validated before writing |
| 服务 | start / stop / restart kixdns, toggle `--debug`, reset statistics, log files and disk usage |

The UI is a single self-contained HTML file embedded into the binary — no CDN,
no build step, no external assets.

## How the statistics work

kixdns `--debug` emits one line per pipeline event:

```
DEBUG request started  event="request_started" request_id=1 client=127.0.0.1:51940 qname="www.qq.com" qtype=A qclass=IN
DEBUG cache hit        event="cache_hit" request_id=2 kind=Fresh remaining_ttl_s=27 original_ttl_s=27
DEBUG cache miss       event="cache_miss" request_id=1
DEBUG request finished event="request_finished" request_id=1 status=Completed latency_us=36131
DEBUG upstream result  event="upstream_result" request_id=1 upstream="127.0.0.1:1053" outcome=Success latency_us=35261 rcode=No Error
```

The collector keeps a **byte offset** into the active log and only parses newly
appended data, persisting its aggregated counters as JSON. Aggregation is
therefore O(new lines), not O(file), which matters on a router.

Values can contain spaces (`rcode=No Error`), so lines are tokenised into
`key=value` pairs rather than matched with per-field regexes — a greedy regex
silently swallows the following field otherwise.

### Log volume and retention

`--debug` is verbose: roughly **66 MB/day** at 40 000 queries. The console
rotates the active log when it exceeds `-log-cap-mb` and keeps `-keep-logs`
rotated files. Defaults (`-keep-logs 2`, `-log-cap-mb 24`) retain **three days**
max, which is what a 1.4 GB `/data` partition can afford.

## Build

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -trimpath -ldflags "-s -w" -o kixdns-web ./cmd/kixdns-web
```

The result is a fully static binary (`CGO_ENABLED=0`), suitable for OpenWrt's
musl userland. CI builds arm64 / amd64 / armv7 / mipsle plus freebsd/amd64 and
runs the parser test-suite first — see `.github/workflows/build-kixdns-web.yml`.

## Run

```sh
kixdns-web \
  -listen 0.0.0.0:8080 \
  -bin    /usr/bin/kixdns \
  -config /etc/kixdns/pipeline.json \
  -logdir /var/log/kixdns \
  -state  /var/lib/kixdns/stats.json
```

Verify the resolved paths and configuration without starting the server:

```sh
kixdns-web -check
```

Directory and binary locations are auto-detected, so on a stock OpenWrt install
the path flags can be omitted.

### Flags

| Flag | Default | Meaning |
|---|---|---|
| `-listen` | `0.0.0.0:8080` | HTTP listen address |
| `-bin` | auto | kixdns binary |
| `-config` | auto | `pipeline.json` |
| `-logdir` | auto | directory holding kixdns logs |
| `-state` | `<logdir>/kixdns-web-stats.json` | aggregated counters |
| `-pidfile` | `/var/run/kixdns.pid` | kixdns PID file |
| `-token` | empty | shared secret (`X-Auth-Token` header or `?token=`) |
| `-readonly` | off | disable all mutating endpoints |
| `-interval` | `5s` | statistics refresh interval |
| `-keep-logs` | `2` | rotated files to keep (active + this = days retained) |
| `-log-cap-mb` | `24` | rotate the active log past this size |
| `-check` | | print resolved config and exit |

## Safety

Configuration writes are validated **before** they can replace anything:

1. JSON syntax and the presence of `pipelines` are checked.
2. The candidate is started once with its listeners rebound to loopback
   ephemeral ports. kixdns has no config-test flag, so a trial start is the only
   way to catch semantic errors; a rejected config exits immediately, a good one
   keeps running and is then stopped. Production listeners cannot be disturbed.
3. Only then is the file replaced by an atomic rename, with the previous
   revision kept as `pipeline.json.bak-<timestamp>`.

kixdns watches its config file and reloads pipelines itself, so no restart is
needed after a save.

## Tests

```sh
go test ./internal/stats/ -v
REAL_LOG=/path/to/captured.log go test ./internal/stats/ -v
```

`REAL_LOG` runs the parser against a captured `kixdns --debug` log and asserts
the counters match the file's contents line for line.
