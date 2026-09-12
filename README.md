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

基于 [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF) 的精简固件，面向 **NanoPi R2S / R3S / R4S 与 x86_64**，只保留 **KixDNS + DaedNext** 两条主线。

[固件下载](#downloads) · [快速开始](#quick-start) · [使用说明](#usage) · [自行编译](#build) · [相关项目](#related)

</div>

---

## ✨ 特性

- 🧩 **只做精简**：应用层只集成 **KixDNS**（Rust DNS 分流/转发 + LuCI）与 **DaedNext**（daed Rust 内核 + 3.1.0 Web 面板），不打包 PassWall / OpenClash / MosDNS / SSRP 等整包套件。
- ⚙️ **内核为 dae 对齐**：内核带 eBPF 所需特性——`CONFIG_VETH`、`NET_SCH_INGRESS`（clsact）、`NET_CLS_ACT`、`NET_CLS_BPF` 与 **BTF（`/sys/kernel/btf/vmlinux`）**，R4S 构建时还会在 CI 内校验 BTF 与 web 资产。
- 🎨 **中文主题**：Footstrap 主题 + 自译 `zh_Hans` 语言包（LuCI 界面为中文）。
- ⚡ **YAOF 优化基础**：保留 BBRv3、LRNG、FullCone 及机型相关的 CPU/NIC 调优。
- 🔧 **配置可追溯**：机型选包在 `SEED/`，脚本与补丁在 `SCRIPTS/`、`PATCH/`。

<a id="downloads"></a>

## 📦 固件下载

前往 **[本仓库 Releases](https://github.com/Quan-0505/OpenWrt/releases)** 选择对应设备附件。文件名含机型、构建日期与 OpenWrt 版本。

| 设备 | 文件名前缀 | 平台 |
|---|---|---|
| NanoPi R2S | `R2S-OpenWrt-*` | RK3328 / ARM64 |
| NanoPi R3S | `R3S-OpenWrt-*` | RK3566 / ARM64 |
| NanoPi R4S | `R4S-OpenWrt-*` | RK3399 / ARM64 |
| x86_64 软路由 | `X86-OpenWrt-*` | x86_64 |

| 格式 | 说明 |
|---|---|
| `*-ext4.zip` | ext4 可写根文件系统 |
| `*-sfs.zip` | SquashFS 只读基础系统，配置写入可写层 |

<a id="quick-start"></a>

## 🚀 快速开始

1. 下载对应设备的固件 ZIP 并解压得到镜像；若仍是 `.img.gz`，按写盘工具要求继续解压。
2. NanoPi 写入 TF 卡（或 eMMC），x86_64 写入目标启动盘。
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
- 开启开关时若提示 `resident candidate preflight failed … required host tool is missing`，安装宿主工具：`apk add tc-full bpftool-minimal ip-full ipset`。
- 若提示 `clsact qdisc add failed` 或 `/sys/kernel/btf/vmlinux` 不存在，说明内核缺少 eBPF 特性，请使用本仓库最新固件（内核已补齐 VETH / clsact / BTF）。
- 内核需要 veth 与 clsact（tc eBPF），本仓库固件已内置；日志见 `/etc/daed/logs/current.jsonl`。

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
