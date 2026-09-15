<div align="center">

# OpenWrt for NanoPi R2S / R3S / R4S & x86_64

**OpenWrt 25.12 streamlined custom firmware · KixDNS + DaedNext · GitHub Actions builds**

[![OpenWrt](https://img.shields.io/badge/OpenWrt-25.12-orange?style=flat-square)](https://github.com/Quan-0505/OpenWrt/releases)
[![Branch](https://img.shields.io/badge/branch-25.12-0b5?style=flat-square)](https://github.com/Quan-0505/OpenWrt/tree/25.12)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue?style=flat-square)](LICENSE)
[![Downloads](https://img.shields.io/github/downloads/Quan-0505/OpenWrt/total?style=flat-square)](https://github.com/Quan-0505/OpenWrt/releases)

[![R2S](https://github.com/Quan-0505/OpenWrt/actions/workflows/R2S-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/R2S-OpenWrt.yml)
[![R3S](https://github.com/Quan-0505/OpenWrt/actions/workflows/R3S-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/R3S-OpenWrt.yml)
[![R4S](https://github.com/Quan-0505/OpenWrt/actions/workflows/R4S-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/R4S-OpenWrt.yml)
[![X86](https://github.com/Quan-0505/OpenWrt/actions/workflows/X86-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/X86-OpenWrt.yml)

A streamlined firmware based on [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF), targeting **NanoPi R2S / R3S / R4S and x86_64**, keeping only the two main lines: **KixDNS + DaedNext**.

[Firmware download](#downloads) · [Quick start](#quick-start) · [Usage](#usage) · [Build it yourself](#build) · [Related projects](#related)

</div>

**English** &nbsp;|&nbsp; **[简体中文](./README.md)**

---

## ✨ Features

- 🧩 **Streamlined only**: the application layer integrates just **KixDNS** (Rust DNS routing/forwarding + LuCI) and **DaedNext** (daed Rust core + 3.1.0 web panel); full suites such as PassWall / OpenClash / MosDNS / SSRP are not packaged.
- ⚙️ **Kernel aligned with dae**: the kernel carries the features eBPF requires — `CONFIG_VETH`, `NET_SCH_INGRESS` (clsact), `NET_CLS_ACT`, `NET_CLS_BPF` and **BTF (`/sys/kernel/btf/vmlinux`)**; R4S builds additionally verify BTF and the web assets inside CI.
- 🎨 **Chinese theme**: Footstrap theme + self-translated `zh_Hans` language pack (the LuCI interface is in Chinese).
- ⚡ **YAOF-optimized base**: keeps BBRv3, LRNG, FullCone and device-specific CPU/NIC tuning.
- 🔧 **Traceable configuration**: per-device package selection lives in `SEED/`, scripts and patches in `SCRIPTS/` and `PATCH/`.

<a id="downloads"></a>

## 📦 Firmware download

All artifacts are gathered in **[a single Release](https://github.com/Quan-0505/OpenWrt/releases)** (tagged with the current OpenWrt version, such as `25.12.5`):
firmware for all four devices plus the footstrap theme packages are all in there; every CI build publishes the newest artifacts there and automatically cleans up older files for the same device.
Filenames include the device, the build date and the OpenWrt version.

| Device | Filename prefix | Platform |
|---|---|---|
| NanoPi R2S | `R2S-OpenWrt-*` | RK3328 / ARM64 |
| NanoPi R3S | `R3S-OpenWrt-*` | RK3566 / ARM64 |
| NanoPi R4S | `R4S-OpenWrt-*` | RK3399 / ARM64 |
| x86_64 soft router | `X86-OpenWrt-*` | x86_64 |

| Format | Description |
|---|---|
| `*-ext4.zip` | ext4 writable root filesystem |
| `*-sfs.zip` | SquashFS read-only base system, with configuration written to a writable layer |

Inside the zip is `openwrt-<target>-<device>-……-sysupgrade.img.gz` (**a complete disk image**).

- **The image already carries JSON metadata** (appended to the end of the `.img.gz` file), so you can upgrade online directly with `sysupgrade`.
- ⚠️ **Flash the `.gz` directly, do not decompress it first**: the metadata sits after the gzip stream and `gunzip` throws it away,
  after which fstools rejects the image (`Firmware image couldn't be validated: no JSON input`, and `-F` does not help either).
- Offline card writing: decompress and write with `dd` / balenaEtcher. For NanoPi write to a TF card or eMMC; for x86_64 write to the target boot disk.

<a id="quick-start"></a>

## 🚀 Quick start

1. Download the firmware ZIP for your device and extract the `*.img.gz` inside it (**do not decompress it**).
2. Online upgrade: transfer the `.gz` to the device and flash it directly——
   ```sh
   scp openwrt-*-sysupgrade.img.gz root@192.168.2.1:/tmp/
   ssh root@192.168.2.1 'sysupgrade -T /tmp/openwrt-*-sysupgrade.img.gz && sysupgrade -v -k /tmp/openwrt-*-sysupgrade.img.gz'
   ```
   Whitelisted files under `/etc` are kept by default (write the paths you want to keep into `/etc/sysupgrade.conf`). `-k` records the list of installed packages.
   Offline card writing: decompress to `.img`, then write to a TF card / eMMC / boot disk with `dd` or balenaEtcher.
3. After boot, connect your computer to a LAN port and open **[http://192.168.2.1](http://192.168.2.1)** in a browser to enter LuCI and finish the network and account configuration.

| Item | Default value |
|---|---|
| LuCI | http://192.168.2.1 |
| SSH | `ssh root@192.168.2.1` |
| daed panel | http://192.168.2.1:2023 |

<a id="usage"></a>

## 🧭 Usage

**KixDNS**: LuCI menu → Services → KixDNS, used for DNS routing and forwarding; just configure the rules following the upstream documentation.

**DaedNext**: open `http://192.168.2.1:2023` in a browser, set the admin password on first entry, then add subscriptions/nodes in the panel and turn on the switch.
- The panel ships with the daemon itself (web root `/usr/share/daed/web`), and **luci-app-daed is not used**, so it is normal that there is no daed entry in the LuCI "Services" menu.
- The host tools (`tc-full` / `bpftool-minimal` / `ip-full` / `ipset`) are built into the firmware; if you trim them yourself and see `required host tool is missing`, add them back manually: `apk add tc-full bpftool-minimal ip-full ipset`.
- The kernel already includes the features dae needs (VETH / clsact / BPF sched / BTF), so no extra kernel modules are required.
- **flow offloading is disabled by default**: the nft flowtable bypasses the tc/eBPF hooks, so daed cannot capture the traffic (the symptom is "the proxy is on but the egress IP has not changed").
- **fw4 rules** are built in (`/usr/share/nftables.d/chain-post/{forward,srcnat}/30-daed-netkit.nft`): they allow forwarding for the daed data-plane netkit device pair (`dae0` ↔ `dae0peer`, `daens` netns), and do SNAT for the source address `169.254.0.11`. Missing these two rules makes every node fail to dial (log `no alive dialer`, all nodes red in the panel), and they still take effect after `fw4 reload`.
- Logs are in `/etc/daed/logs/current.jsonl` (actually pointing to `/tmp/log/daed`, cleared on reboot); the state database is `/etc/daed/daed.db` (SQLite).
- Two known behaviours: ① node latency/alive status in the panel sometimes does not refresh (shown grey while forwarding actually works — **judge by whether blocked sites open**); ② daed is fail-closed — when a policy group has no usable node, proxied traffic is rejected, while domestic direct rules are unaffected.
- When troubleshooting "is it working or not", make sure to confirm that **TCP from the router itself to the node IP** is reachable (ICMP working does not mean TCP works; upstream policy routing/loops may kill only TCP): `curl -v --max-time 8 https://<节点IP>:<端口>`.

<a id="packages"></a>

## 📦 Installing third-party packages (apk v3)

This firmware is based on **OpenWrt 25.12**, its package manager is **apk-tools 3.x** and the package format is **apk v3** (an ADB container; it is neither tar nor gzip, so it is normal that `tar`/apk2 tools cannot open it).

```sh
apk add --allow-untrusted ./some-package.apk     # 未签名的本地包
apk add some-package                             # 来自配置好的源
```

- Installing a package in **apk v2** format fails outright with `ERROR: ...: v2 package format error` — you need to provide a v3 package.
- In the releases of [rust-daed](https://github.com/Quan-0505/rust-daed) and [daed-kdae](https://github.com/Quan-0505/daed-kdae),
  the canonical asset `*-<device>.apk` **has been repackaged as apk v3** and can be installed directly with `apk add` on 25.12; the old format is kept as `*-<device>-v2.apk`.
  This repository's firmware already includes daed, so there is normally no need to install it separately (the two share `/usr/bin/daed`, so only one can be installed).
- The theme packages (same baseline as this firmware) are in the **same Release**: `luci-theme-footstrap-<ver>.apk` + `luci-i18n-footstrap-zh-cn-<ver>.apk` (25.12 / apk v3),
  `luci-theme-footstrap_<ver>_all.ipk` + `luci-i18n-footstrap-zh-cn_<ver>_all.ipk` (24.10 / opkg).


<a id="build"></a>

## 🔧 Build it yourself

### Building a single device

1. Open **Actions** and choose `R2S-OpenWrt`, `R3S-OpenWrt`, `R4S-OpenWrt` or `X86-OpenWrt`.
2. **Run workflow**, and select the `25.12` branch.
3. Enable `build_firmware` and wait for the build and release to finish.
4. Download the corresponding ZIP from **Releases**.

### Building several devices at once

Run **OpenWrt-Matrix** and fill in the fields as needed:

| Parameter | Purpose |
|---|---|
| `targets` | Comma-separated devices, such as `R2S,R4S,X86`; leave empty to build all |
| `build_firmware` | Build and release the firmware |
| `upload_source` | Upload the pre-build source archive to make configuration review easier |

The theme packages (theme / zh_Hans) are built and released by the separate `build-footstrap-ipk` / `build-footstrap-apk` pipelines, decoupled from the firmware build.

## 📂 Repository structure

| Path | Description |
|---|---|
| [`SEED/`](SEED/) | Device configurations for R2S / R3S / R4S / X86 |
| [`SCRIPTS/`](SCRIPTS/) | Source preparation, package handling, device adaptation and build helper scripts |
| [`PATCH/kernel/`](PATCH/kernel/) | Kernel patches |
| [`PATCH/pkgs/`](PATCH/pkgs/) | Package patches |
| [`PATCH/files/`](PATCH/files/) | Preset files written into the firmware |
| [`PATCH/daed-pkg/`](PATCH/daed-pkg/) | Prebuilt daed package definitions (non-R4S devices)|
| [`PATCH/daenext-r4s/`](PATCH/daenext-r4s/) | DaeNext source-built package definitions for R4S |
| [`.github/workflows/`](.github/workflows/) | Per-device, matrix, theme and cleanup workflows |

<a id="related"></a>

## 🤝 Credits and related projects

Thanks to [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF) for providing the build base, as well as to [OpenWrt](https://github.com/openwrt/openwrt), [ImmortalWrt](https://github.com/immortalwrt), [coolsnowwolf](https://github.com/coolsnowwolf), [Lienol](https://github.com/Lienol) and the authors of the various packages for their contributions.

- [ksong008/DaeNext](https://github.com/ksong008/DaeNext) / [ksong008/DaedNext](https://github.com/ksong008/DaedNext) — Rust proxy core and web control panel.
- [Quan-0505/rust-daed](https://github.com/Quan-0505/rust-daed) — standalone Rust-native daed install package.
- [Quan-0505/daed-kdae](https://github.com/Quan-0505/daed-kdae) — standalone Go / kdae daed install package.
- [VizzleTF/luci-theme-footstrap](https://github.com/VizzleTF/luci-theme-footstrap) — the Footstrap theme (including this repository's self-translated Chinese pack).
- [JohnsonRan/luci-app-kixdns](https://github.com/JohnsonRan/luci-app-kixdns) — the LuCI application for KixDNS.
- [Quan-0505/OPNsense-For-R4S](https://github.com/Quan-0505/OPNsense-For-R4S) — OPNsense firmware and tuning notes for R4S.
- [Quan-0505/OPNsense-kixdns-web](https://github.com/Quan-0505/OPNsense-kixdns-web) — OPNsense integration for KixDNS.

## 📄 License

This repository uses the [GNU General Public License v3.0](LICENSE). Each component inside the firmware follows its upstream license; refer to the license file of the corresponding project for details.
