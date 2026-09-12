<div align="center">

# OpenWrt for NanoPi & x86_64

**OpenWrt 25.12 定制固件 · 多机型支持 · 网络应用集成 · GitHub Actions 构建**

[![OpenWrt](https://img.shields.io/badge/OpenWrt-25.12-orange?style=flat-square)](https://github.com/Quan-0505/OpenWrt/releases)
[![Branch](https://img.shields.io/badge/branch-25.12-0b5?style=flat-square)](https://github.com/Quan-0505/OpenWrt/tree/25.12)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue?style=flat-square)](LICENSE)
[![Downloads](https://img.shields.io/github/downloads/Quan-0505/OpenWrt/total?style=flat-square)](https://github.com/Quan-0505/OpenWrt/releases)

[![R2C](https://github.com/Quan-0505/OpenWrt/actions/workflows/R2C-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/R2C-OpenWrt.yml)
[![R2S](https://github.com/Quan-0505/OpenWrt/actions/workflows/R2S-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/R2S-OpenWrt.yml)
[![R3S](https://github.com/Quan-0505/OpenWrt/actions/workflows/R3S-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/R3S-OpenWrt.yml)
[![R4S](https://github.com/Quan-0505/OpenWrt/actions/workflows/R4S-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/R4S-OpenWrt.yml)
[![X86](https://github.com/Quan-0505/OpenWrt/actions/workflows/X86-OpenWrt.yml/badge.svg?branch=25.12)](https://github.com/Quan-0505/OpenWrt/actions/workflows/X86-OpenWrt.yml)

基于 [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF)，面向 **NanoPi R2C / R2S / R3S / R4S 与 x86_64** 的个人定制固件。

[固件下载](#downloads) · [分支说明](#branches) · [快速开始](#quick-start) · [自行编译](#build) · [相关项目](#related)

</div>

---

## ✨ 特性

- 📦 **多机型构建**：提供独立机型流程与矩阵构建，支持 ext4、SquashFS 两种镜像。
- 🌐 **网络应用集成**：配置包含 dae、PassWall、OpenClash、SSRP、HomeProxy、Nikki 等，具体以机型配置及构建结果为准。
- 📌 **DNS 与连接管理**：集成 MosDNS、DDNS、FRP、ZeroTier、WireGuard 等应用。
- ⚡ **YAOF 优化基础**：保留 BBRv3、LRNG、FullCone 及机型相关的 CPU、网卡调优。
- 🖥️ **LuCI 管理**：包含 Aurora、Bootstrap 主题，以及 CPU 调频、分区扩容、流量监控等工具。
- 🔧 **配置可追溯**：各机型选包位于 `SEED/`，构建脚本与内核补丁保存在仓库中。

<a id="downloads"></a>

## 📦 固件下载

前往 **[本仓库 Releases](https://github.com/Quan-0505/OpenWrt/releases)**，选择设备对应的附件。文件名包含机型、构建日期和 OpenWrt 版本。

| 设备 | 文件名前缀 | 平台 |
|---|---|---|
| NanoPi R2C | `R2C-OpenWrt-*` | RK3328 / ARM64 |
| NanoPi R2S | `R2S-OpenWrt-*` | RK3328 / ARM64 |
| NanoPi R3S | `R3S-OpenWrt-*` | RK3566 / ARM64 |
| NanoPi R4S | `R4S-OpenWrt-*` | RK3399 / ARM64 |
| x86_64 软路由 | `X86-OpenWrt-*` | x86_64 |

| 格式 | 说明 |
|---|---|
| `*-ext4.zip` | ext4 可写根文件系统 |
| `*-sfs.zip` | SquashFS 只读基础系统，配置写入可写层 |

同一仓库可能保留不同配置时期的固件，请结合 Release 说明和构建来源选择附件。当前 README 描述的是 **`25.12` 分支源码配置**。

<a id="branches"></a>

## 🌿 分支说明

| 分支 | 用途 | 主要配置 |
|---|---|---|
| [`25.12`](https://github.com/Quan-0505/OpenWrt/tree/25.12) | 默认分支，多插件固件 | dae、MosDNS、PassWall 等；Aurora / Bootstrap 主题 |
| [`codex/r4s-daenext-source`](https://github.com/Quan-0505/OpenWrt/tree/codex/r4s-daenext-source) | R4S DaeNext 源码构建分支 | DaeNext Rust 核心 + DaedNext 3.1.0、KixDNS、Footstrap 中文主题 |

**需要新版 R4S DaeNext 固件，请使用 [DaeNext 分支](https://github.com/Quan-0505/OpenWrt/tree/codex/r4s-daenext-source#readme)。** 该分支有独立的打包脚本、BTF 检查和发布标签，尚未合并到默认分支。

<a id="quick-start"></a>

## 🚀 快速开始

1. 下载对应设备的固件 ZIP，解压得到镜像；如仍为 `.img.gz`，按写盘工具要求继续解压。
2. NanoPi 将镜像写入 TF 卡，x86_64 将镜像写入目标启动盘。
3. 启动设备，电脑连接 LAN 口，通过浏览器进入 LuCI，完成网络和账户配置。

| 配置来源 | 默认管理地址 |
|---|---|
| 本分支 `25.12` | [http://192.168.1.1](http://192.168.1.1) |
| DaeNext 定制分支 | [http://192.168.2.1](http://192.168.2.1) |

保留配置升级时，管理地址以原有配置为准。DaeNext 分支的独立 daed 面板和初始化说明见该分支 README。

<a id="build"></a>

## 🔧 自行编译

### 构建单个机型

1. 打开 **Actions**，选择对应的 `R2C-OpenWrt`、`R2S-OpenWrt`、`R3S-OpenWrt`、`R4S-OpenWrt` 或 `X86-OpenWrt`。
2. 点击 **Run workflow**，分支选择 **`25.12`**。
3. 开启 `build_firmware`，等待编译和发布完成。
4. 前往 **Releases** 下载对应设备的固件 ZIP。

### 同时构建多个机型

运行 **OpenWrt-Matrix**，按需填写以下参数：

| 参数 | 用途 |
|---|---|
| `targets` | 逗号分隔的机型，例如 `R2S,R4S,X86`；留空构建全部机型 |
| `build_firmware` | 编译并发布固件 |
| `upload_source` | 上传编译前的源码归档，便于检查构建配置 |

默认分支按 OpenWrt 版本标签发布固件。DaeNext 分支的 R4S 构建使用 `<OpenWrt版本>-daenext-<运行编号>` 独立标签，选择流程时请同时确认分支。

## 📂 仓库结构

| 路径 | 说明 |
|---|---|
| [`SEED/`](SEED/) | R2C / R2S / R3S / R4S / X86 的机型配置 |
| [`SCRIPTS/`](SCRIPTS/) | 源码准备、软件包处理、机型适配和构建辅助脚本 |
| [`PATCH/kernel/`](PATCH/kernel/) | 内核补丁 |
| [`PATCH/pkgs/`](PATCH/pkgs/) | 软件包补丁 |
| [`PATCH/files/`](PATCH/files/) | 写入固件的预设文件 |
| [`.github/workflows/`](.github/workflows/) | 各机型、矩阵构建及清理流程 |

<a id="related"></a>

## 🤝 鸣谢与相关项目

感谢 [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF) 提供构建基础，以及 [OpenWrt](https://github.com/openwrt/openwrt)、[ImmortalWrt](https://github.com/immortalwrt)、[coolsnowwolf](https://github.com/coolsnowwolf)、[Lienol](https://github.com/Lienol) 和各软件包作者的贡献。完整贡献记录可在上游项目及本仓库提交历史中查看。

- [ksong008/DaeNext](https://github.com/ksong008/DaeNext) / [ksong008/DaedNext](https://github.com/ksong008/DaedNext) — Rust 代理核心与网页控制面板。
- [Quan-0505/rust-daed](https://github.com/Quan-0505/rust-daed) — Rust 原生 daed 独立安装包。
- [Quan-0505/daed-kdae](https://github.com/Quan-0505/daed-kdae) — Go / kdae 版 daed 独立安装包。
- [Quan-0505/OPNsense-For-R4S](https://github.com/Quan-0505/OPNsense-For-R4S) — R4S 的 OPNsense 固件与调优记录。
- [Quan-0505/OPNsense-kixdns-web](https://github.com/Quan-0505/OPNsense-kixdns-web) — KixDNS 的 OPNsense 集成。

## 📄 许可

本仓库采用 [GNU General Public License v3.0](LICENSE)。固件内各组件遵循其上游许可证，具体以对应项目的许可文件为准。
