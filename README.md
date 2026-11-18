<div align="center">

# OpenWrt for NanoPi R2S / R3S / R4S & x86_64

**OpenWrt 25.12 精简定制固件 · KixDNS + DaedNext · GitHub Actions 构建**

[![OpenWrt](https://img.shields.io/badge/OpenWrt-25.12-orange?style=flat-square)](https://github.com/Quan-0505/OpenWrt/releases)
[![Branch](https://img.shields.io/badge/branch-25.12-0b5?style=flat-square)](https://github.com/Quan-0505/OpenWrt/tree/25.12)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue?style=flat-square)](LICENSE)
[![Downloads](https://img.shields.io/github/downloads/Quan-0505/OpenWrt/total?style=flat-square)](https://github.com/Quan-0505/OpenWrt/releases)

[![R2S](https://github.com/Quan-0505/OpenWrt/actions/workflows/R2S-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/R2S-OpenWrt.yml)
[![R3S](https://github.com/Quan-0505/OpenWrt/actions/workflows/R3S-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/R3S-OpenWrt.yml)
[![R4S](https://github.com/Quan-0505/OpenWrt/actions/workflows/R4S-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/R4S-OpenWrt.yml)
[![X86](https://github.com/Quan-0505/OpenWrt/actions/workflows/X86-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/X86-OpenWrt.yml)

**[简体中文](./README.md)** &nbsp;|&nbsp; **[English](./README.en.md)**

基于 [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF) 的精简固件，面向 **NanoPi R2S / R3S / R4S 与 x86_64**，只保留 **KixDNS + DaedNext** 两条主线。

[固件下载](#downloads) · [快速开始](#quick-start) · [使用说明](#usage) · [自行编译](#build) · [相关项目](#related)

</div>

---

## ✨ 特性

- 🧩 **只做精简**：应用层只集成 **KixDNS**（Rust DNS 分流/转发 + LuCI）与 **DaedNext**（daed Rust 内核 + 3.1.0 Web 面板），不打包 PassWall / OpenClash / MosDNS / SSRP 等整包套件。
- ⚙️ **内核为 dae 对齐**：内核带 eBPF 所需特性——`CONFIG_VETH`、`NET_SCH_INGRESS`（clsact）、`NET_CLS_ACT`、`NET_CLS_BPF` 与 **BTF（`/sys/kernel/btf/vmlinux`）**，R4S 构建时还会在 CI 内校验 BTF 与 web 资产。
- 🎨 **中文主题**：Footstrap 主题 + 自译 `zh_Hans` 语言包（LuCI 界面为中文）。
- ⚡ **YAOF 优化基础**：保留 BBRv3、LRNG、FullCone 及机型相关的 CPU/NIC 调优；Broadcom FullCone 改用 YAOF 当前的 conntrack-helper 补丁组，避免旧 nft-fullcone 与新版 firewall4 补丁冲突。
- 🧰 **补丁跟随上游**：miniupnpd 的端口转发兼容选项按当前 packages feed 配置更新，避免对上游已包含的初始化脚本重复打补丁。
- 🔧 **配置可追溯**：机型选包在 `SEED/`，脚本与补丁在 `SCRIPTS/`、`PATCH/`。

<a id="downloads"></a>

## 📦 固件下载

**最新固件** · [OpenWrt 25.12.6-20261118](https://github.com/Quan-0505/OpenWrt/releases/tag/25.12.6-20261118)

**旧版下载** · [OpenWrt 25.12.5](https://github.com/Quan-0505/OpenWrt/releases/tag/25.12.5)

2026-11-18 构建自 [e0a1823](https://github.com/Quan-0505/OpenWrt/commit/e0a18239b97ef4cf9c93824a1164ab7121912fc0)。本次同步 [YAOF 25.12 至 fb2fc54](https://github.com/QiuSimons/YAOF/commit/fb2fc542411461c4856d2137ae19a263a8ed070e)，更新 firewall4 兼容性，并核验 R4S 的 daed 载荷和内核 BTF。8 个 ZIP 均通过镜像检查和 SHA-256 校验，实机启动待验证。

| 设备 | 平台 | ext4 | SquashFS |
|---|---|---|---|
| NanoPi R2S | RK3328 / ARM64 | [ext4 ZIP](https://github.com/Quan-0505/OpenWrt/releases/download/25.12.6-20261118/R2S-OpenWrt-2026-11-18-25.12.6-ext4.zip) | [SquashFS ZIP](https://github.com/Quan-0505/OpenWrt/releases/download/25.12.6-20261118/R2S-OpenWrt-2026-11-18-25.12.6-sfs.zip) |
| NanoPi R3S | RK3566 / ARM64 | [ext4 ZIP](https://github.com/Quan-0505/OpenWrt/releases/download/25.12.6-20261118/R3S-OpenWrt-2026-11-18-25.12.6-ext4.zip) | [SquashFS ZIP](https://github.com/Quan-0505/OpenWrt/releases/download/25.12.6-20261118/R3S-OpenWrt-2026-11-18-25.12.6-sfs.zip) |
| NanoPi R4S | RK3399 / ARM64 | [ext4 ZIP](https://github.com/Quan-0505/OpenWrt/releases/download/25.12.6-20261118/R4S-OpenWrt-2026-11-18-25.12.6-ext4.zip) | [SquashFS ZIP](https://github.com/Quan-0505/OpenWrt/releases/download/25.12.6-20261118/R4S-OpenWrt-2026-11-18-25.12.6-sfs.zip) |
| x86_64 | x86_64 / EFI | [ext4 ZIP](https://github.com/Quan-0505/OpenWrt/releases/download/25.12.6-20261118/X86-OpenWrt-2026-11-18-25.12.6-ext4.zip) | [SquashFS ZIP](https://github.com/Quan-0505/OpenWrt/releases/download/25.12.6-20261118/X86-OpenWrt-2026-11-18-25.12.6-sfs.zip) |

ext4 使用可写根文件系统。SquashFS 使用只读基础系统和可写覆盖层。每个 ZIP 内含一份压缩磁盘镜像，NanoPi 文件以 `sysupgrade.img.gz` 结尾，x86_64 EFI 文件以 `combined-efi.img.gz` 结尾。

<a id="quick-start"></a>

## 🚀 快速开始

1. 下载对应设备的固件 ZIP，解出其中的 `*.img.gz`（**不要解压**）。
2. NanoPi 在线升级时直接使用 `.img.gz`，以保留升级元数据。把文件传到设备，校验通过后再刷写。
   ```sh
   scp openwrt-*-sysupgrade.img.gz root@192.168.2.1:/tmp/
   ssh root@192.168.2.1 'sysupgrade -T /tmp/openwrt-*-sysupgrade.img.gz && sysupgrade -v -k /tmp/openwrt-*-sysupgrade.img.gz'
   ```
   默认保留 `/etc` 内白名单文件（把要保留的路径写进 `/etc/sysupgrade.conf`）。`-k` 会记录已装包列表。
   离线安装时解压为 `.img`，写入 NanoPi 的 TF 卡或 eMMC，或写入 x86_64 的 EFI 启动盘。
3. 启动后电脑接 LAN 口，浏览器打开 **[http://192.168.2.1](http://192.168.2.1)** 进入 LuCI 完成网络与账户配置。

| 项目 | 默认值 |
|---|---|
| LuCI | http://192.168.2.1 |
| SSH | `ssh root@192.168.2.1` |
| daed 面板 | http://192.168.2.1:2023 |

<a id="usage"></a>

## 🧭 使用说明

**KixDNS**：LuCI 菜单 → 服务 → KixDNS，用于 DNS 分流与转发，按上游文档配置规则即可。

**DaedNext**：浏览器打开 `http://192.168.2.1:2023`，首次进入设置管理密码，然后在面板里添加订阅/节点并开启开关。
- 面板由 daemon 自带（web 根目录 `/usr/share/daed/web`），**不使用 luci-app-daed**，因此 LuCI "服务" 菜单里没有 daed 入口属正常现象。
- 宿主工具（`tc-full` / `bpftool-minimal` / `ip-full` / `ipset`）已随固件内置；若自行裁剪后提示 `required host tool is missing`，手动补：`apk add tc-full bpftool-minimal ip-full ipset`。
- 内核已内置 dae 所需特性（VETH / clsact / BPF sched / BTF），无需额外内核模块。
- **flow offloading 已默认关闭**：nft flowtable 会绕过 tc/eBPF 钩子，导致 daed 抓不到流量（表现为"开了代理但出口 IP 没变"）。
- 已内置 **fw4 规则**（`/usr/share/nftables.d/chain-post/{forward,srcnat}/30-daed-netkit.nft`）：放行 daed 数据面 netkit 设备对（`dae0` ↔ `dae0peer`，`daens` netns）的转发，并为源地址 `169.254.0.11` 做 SNAT。缺这两条会让节点全部拨号失败（日志 `no alive dialer`、面板节点全红），且 `fw4 reload` 后依然生效。
- 日志见 `/etc/daed/logs/current.jsonl`（实际指向 `/tmp/log/daed`，重启清空）；状态库 `/etc/daed/daed.db`（SQLite）。
- 两个已知现象：① 面板的节点延迟/存活有时不刷新（显示灰色但实际转发正常，**以能否打开墙外站点为准**）；② daed 属 fail-closed——策略组无可用节点时被代理流量会被拒绝，国内直连规则不受影响。
- 排查"通了没有"，务必确认**路由器本身到节点 IP 的 TCP** 可达（ICMP 通不代表 TCP 通；上游策略路由/环路会只打死 TCP）：`curl -v --max-time 8 https://<节点IP>:<端口>`。

<a id="packages"></a>

## 📦 安装第三方包（apk v3）

本固件基于 **OpenWrt 25.12**，包管理为 **apk-tools 3.x**，包格式是 **apk v3**（ADB 容器；既不是 tar 也不是 gzip，用 `tar`/apk2 工具打不开是正常的）。

```sh
apk add --allow-untrusted ./some-package.apk     # 未签名的本地包
apk add some-package                             # 来自配置好的源
```

- 安装 **apk v2** 格式的包会直接报 `ERROR: ...: v2 package format error`——需要提供 v3 包。
- [rust-daed](https://github.com/Quan-0505/rust-daed)、[daed-kdae](https://github.com/Quan-0505/daed-kdae) 的 release 里，
  规范资产 `*-<device>.apk` **已重打包为 apk v3**，可在 25.12 上直接 `apk add`；旧格式保留为 `*-<device>-v2.apk`。
  本仓库固件本身已内置 daed，正常无需另装（两者共用 `/usr/bin/daed`，只能装一个）。
- 独立 Footstrap 主题包见 [25.12.5 发布页](https://github.com/Quan-0505/OpenWrt/releases/tag/25.12.5)的附件，包括`luci-theme-footstrap-<ver>.apk` + `luci-i18n-footstrap-zh-cn-<ver>.apk`（25.12 / apk v3）、
  `luci-theme-footstrap_<ver>_all.ipk` + `luci-i18n-footstrap-zh-cn_<ver>_all.ipk`（24.10 / opkg）。


<a id="build"></a>

## 🔧 自行编译

### 构建单个机型

1. 打开 **Actions**，选择 `R2S-OpenWrt`、`R3S-OpenWrt`、`R4S-OpenWrt` 或 `X86-OpenWrt`。
2. **Run workflow**，分支选择 `25.12`。
3. 开启 `build_firmware`，等待编译与发布完成。
4. 在 **Releases** 下载对应 ZIP。

### 同时构建多个机型

运行 **OpenWrt-Matrix**，按需填写：

| 参数 | 用途 |
|---|---|
| `targets` | 逗号分隔机型，如 `R2S,R4S,X86`；留空构建全部 |
| `build_firmware` | 编译并发布固件 |
| `upload_source` | 上传编译前源码归档，便于核查配置 |

主题包（theme / zh_Hans）由独立的 `build-footstrap-ipk` / `build-footstrap-apk` 流程构建发布，与固件构建解耦。

## 📂 仓库结构

| 路径 | 说明 |
|---|---|
| [`SEED/`](SEED/) | R2S / R3S / R4S / X86 机型配置 |
| [`SCRIPTS/`](SCRIPTS/) | 源码准备、软件包处理、机型适配与构建辅助脚本 |
| [`PATCH/kernel/`](PATCH/kernel/) | 内核补丁 |
| [`PATCH/pkgs/`](PATCH/pkgs/) | 软件包补丁 |
| [`PATCH/files/`](PATCH/files/) | 写入固件的预设文件 |
| [`PATCH/daed-pkg/`](PATCH/daed-pkg/) | daed 预编译包定义（非 R4S 机型）|
| [`PATCH/daenext-r4s/`](PATCH/daenext-r4s/) | R4S 的 DaeNext 源码编译包定义 |
| [`.github/workflows/`](.github/workflows/) | 机型、矩阵、主题与清理流程 |

<a id="related"></a>

## 🤝 鸣谢与相关项目

感谢 [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF) 提供构建基础，以及 [OpenWrt](https://github.com/openwrt/openwrt)、[ImmortalWrt](https://github.com/immortalwrt)、[coolsnowwolf](https://github.com/coolsnowwolf)、[Lienol](https://github.com/Lienol) 与各软件包作者的贡献。

- [ksong008/DaeNext](https://github.com/ksong008/DaeNext) / [ksong008/DaedNext](https://github.com/ksong008/DaedNext) — Rust 代理核心与网页控制面板。
- [Quan-0505/rust-daed](https://github.com/Quan-0505/rust-daed) — Rust 原生 daed 独立安装包。
- [Quan-0505/daed-kdae](https://github.com/Quan-0505/daed-kdae) — Go / kdae 版 daed 独立安装包。
- [VizzleTF/luci-theme-footstrap](https://github.com/VizzleTF/luci-theme-footstrap) — Footstrap 主题（含本仓库自译中文包）。
- [JohnsonRan/luci-app-kixdns](https://github.com/JohnsonRan/luci-app-kixdns) — KixDNS 的 LuCI 应用。
- [Quan-0505/OPNsense-For-R4S](https://github.com/Quan-0505/OPNsense-For-R4S) — R4S 的 OPNsense 固件与调优记录。
- [Quan-0505/OPNsense-kixdns-web](https://github.com/Quan-0505/OPNsense-kixdns-web) — KixDNS 的 OPNsense 集成。

## 📄 许可

本仓库采用 [GNU General Public License v3.0](LICENSE)。固件内各组件遵循其上游许可证，具体以对应项目的许可文件为准。

## 2026-11-02 上游同步

已合并 YAOF `2469d54` 的最新源码。上游新增 ImageBuilder 的 profiles/scripts 保留在仓库中；本定制分支继续使用既有源码编译流程，以保留 KixDNS、DAED 和 R2S/R3S/R4S/X86 四机型的内核及插件配置。新的构建结果以 Actions 的实际校验为准。
