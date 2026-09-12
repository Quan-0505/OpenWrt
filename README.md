<div align="center">

# OpenWrt for NanoPi & x86_64

**精简软路由固件 · DaeNext 透明代理 · KixDNS 分流 · Footstrap 中文主题**

[![OpenWrt](https://img.shields.io/badge/OpenWrt-25.12-orange?style=flat-square)](https://github.com/Quan-0505/OpenWrt/releases)
[![DaeNext](https://img.shields.io/badge/R4S-DaeNext%20%2B%20DaedNext%203.1.0-e05d44?style=flat-square)](DAENEXT-R4S.md)
[![R4S Build](https://github.com/Quan-0505/OpenWrt/actions/workflows/R4S-OpenWrt.yml/badge.svg?branch=codex%2Fr4s-daenext-source)](https://github.com/Quan-0505/OpenWrt/actions/workflows/R4S-OpenWrt.yml?query=branch%3Acodex%2Fr4s-daenext-source)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue?style=flat-square)](LICENSE)

基于 [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF)，面向 **NanoPi R2S / R3S / R4S 与 x86_64**，保留透明代理、DNS 分流和中文管理界面。

[固件下载](#downloads) · [快速开始](#quick-start) · [R4S 构建说明](#r4s) · [自行编译](#build) · [相关项目](#related)

</div>

---

## ✨ 特性

- 🦀 **R4S 源码构建 DaeNext**：Rust 原生透明代理核心，启用 Aya/eBPF、BoringSSL 和 jemalloc，搭配 **DaedNext 3.1.0** 网页控制面板。
- 📌 **KixDNS 分流**：集成 Rust DNS 转发器与 LuCI 配置页面。
- 🎨 **Footstrap 中文主题**：内置主题和简体中文语言包，默认启用。
- 🧠 **R4S BTF 检查**：关闭与 BTF 冲突的精简调试信息选项，发布前检查内核实际包含 `.BTF`。
- 📦 **出包校验**：R4S 构建检查二进制架构、静态链接、版本标识和网页入口资源。
- ⚡ **YAOF 优化基础**：保留上游的 BBRv3、FullCone 及机型相关调优，应用集聚焦 daed 与 KixDNS。

<a id="downloads"></a>

## 📦 固件下载

前往 **[GitHub Releases](https://github.com/Quan-0505/OpenWrt/releases)**，按设备和文件系统选择固件。每个 ZIP 文件名包含构建日期和 OpenWrt 版本。

| 设备 | 文件名前缀 | 平台 |
|---|---|---|
| NanoPi R4S | `R4S-OpenWrt-*` | RK3399 / ARM64 |
| NanoPi R3S | `R3S-OpenWrt-*` | RK3566 / ARM64 |
| NanoPi R2S | `R2S-OpenWrt-*` | RK3328 / ARM64 |
| x86_64 软路由 | `X86-OpenWrt-*` | x86_64 |

| 格式 | 说明 |
|---|---|
| `*-ext4.zip` | ext4 可写根文件系统 |
| `*-sfs.zip` | SquashFS 只读基础系统，配置写入可写层 |

**本分支的 R4S DaeNext 固件使用独立发布标签**：`<OpenWrt版本>-daenext-<运行编号>`。请在对应 Release 中下载 R4S 附件；普通版本标签中的历史固件不代表已包含本分支改动。只有通过构建和出包检查的运行才会发布固件。

<a id="quick-start"></a>

## 🚀 快速开始

1. 下载对应设备的 ZIP，解压得到镜像；如镜像仍为 `.img.gz`，按写盘工具要求继续解压。
2. R2S / R3S / R4S 将镜像写入 TF 卡；x86_64 将镜像写入目标启动盘。
3. 启动设备，电脑连接 LAN 口，进入管理页面完成初始化。

| 管理入口 | 默认地址 |
|---|---|
| OpenWrt / LuCI | [http://192.168.2.1](http://192.168.2.1) |
| daed 独立面板 | [http://192.168.2.1:2023](http://192.168.2.1:2023) |

Footstrap 为默认主题。daed 服务配置为开机启动，首次使用仍需在面板完成账户初始化、添加节点或订阅并配置路由。

<a id="r4s"></a>

## 🦀 R4S：DaeNext + DaedNext 3.1.0

**DaeNext** 提供 Rust 代理核心与 `daed` 守护进程，**DaedNext** 提供配套网页界面。本分支对 R4S 同时从固定源码提交构建这两个组件。

| 组件 | 来源 | 固定版本 / 提交 |
|---|---|---|
| Rust 核心与守护进程 | [ksong008/DaeNext](https://github.com/ksong008/DaeNext) | [`218bdf72b5be`](https://github.com/ksong008/DaeNext/commit/218bdf72b5be3e59494513c9805fe701dc0fed04) |
| 网页控制面板 | [ksong008/DaedNext](https://github.com/ksong008/DaedNext) | **3.1.0** · [`b084f2f6f006`](https://github.com/ksong008/DaedNext/commit/b084f2f6f0062a12d6087ad5243bb4a436d37f45) |
| 编译目标 | OpenWrt musl 工具链 | `aarch64-unknown-linux-musl` |

- 使用通用 ARMv8 指令集，兼容 RK3399 的两组 CPU 核心。
- 配置和数据库保存在 `/etc/daed/`；新安装的日志写入内存目录 `/tmp/log/daed/`。
- 网页 JS/CSS 入口在打包和固件生成后均检查，避免压缩文件替代原始资源导致白屏。
- 通过 QEMU 检查程序版本，通过 ELF 检查架构与静态链接；实际代理转发需在 R4S 刷入并配置后验证。

R2S、R3S 和 x86_64 当前仍沿用 [rust-daed v3.1.1-sticky](https://github.com/Quan-0505/rust-daed/releases/tag/v3.1.1-sticky) 预编译载荷与仓库内网页资源。该包的 sticky-ip 增强不作为本分支 R4S 源码构建的功能承诺。

构建细节见 **[DAENEXT-R4S.md](DAENEXT-R4S.md)**。

<a id="build"></a>

## 🔧 自行编译

### R4S DaeNext 固件

1. 打开 **Actions → R4S-OpenWrt → Run workflow**。
2. 分支选择 **`codex/r4s-daenext-source`**，开启 `build_firmware`。
3. 构建依次准备 OpenWrt 工具链、编译 DaeNext 与网页界面、生成完整固件并执行出包检查。
4. 成功后从对应 **Release** 或运行页面的 **Artifacts** 下载固件 ZIP。

### 其他机型

选择对应的 `R2S-OpenWrt`、`R3S-OpenWrt` 或 `X86-OpenWrt` 流程。也可运行 `OpenWrt-Matrix`，在 `targets` 中填入 `R2S,R3S,R4S,X86` 的所需组合；留空会构建全部机型。

### Footstrap 独立安装包

主题 APK 由 **Build footstrap APK for OpenWrt 25.12** 单独构建，发布到 **[footstrap-zh](https://github.com/Quan-0505/OpenWrt/releases/tag/footstrap-zh)**，与固件发布流程分开。

下载主题及中文语言包并上传至路由器 `/tmp/` 后执行：

```sh
apk add --allow-untrusted /tmp/luci-theme-footstrap_*.apk /tmp/luci-i18n-footstrap-zh-cn-*.apk
```

`--allow-untrusted` 用于安装尚未加入路由器信任库的 SDK 本地签名包。安装后在 LuCI 的语言和界面设置中选择 Footstrap 和简体中文；24.10 的 IPK 请使用 Release 中对应格式的附件。

## 📂 仓库结构

| 路径 | 说明 |
|---|---|
| [`SEED/`](SEED/) | 各机型固件配置 |
| [`.github/workflows/`](.github/workflows/) | 固件与主题包构建流程 |
| [`SCRIPTS/build_daenext_r4s.sh`](SCRIPTS/build_daenext_r4s.sh) | R4S 核心与网页源码构建 |
| [`SCRIPTS/daenext_toolchain.sh`](SCRIPTS/daenext_toolchain.sh) | OpenWrt 工具链与 musl 头文件定位 |
| [`PATCH/daenext-r4s/`](PATCH/daenext-r4s/) | R4S 的 daed 包定义及启动服务 |
| [`PATCH/daed-pkg/`](PATCH/daed-pkg/) / [`PATCH/daed-web/`](PATCH/daed-web/) | 其他机型使用的预编译包定义与网页资源 |
| [`PATCH/theme-footstrap-zh/`](PATCH/theme-footstrap-zh/) | Footstrap 简体中文翻译 |
| [`PATCH/files/`](PATCH/files/) | 默认 LAN 地址、主题等固件配置 |

<a id="related"></a>

## 🤝 相关项目

- [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF) — 固件构建基础。
- [ksong008/DaeNext](https://github.com/ksong008/DaeNext) / [ksong008/DaedNext](https://github.com/ksong008/DaedNext) — Rust 代理核心与网页控制面板。
- [JohnsonRan/luci-app-kixdns](https://github.com/JohnsonRan/luci-app-kixdns) — KixDNS 的 OpenWrt 集成。
- [VizzleTF/luci-theme-footstrap](https://github.com/VizzleTF/luci-theme-footstrap) — Footstrap 主题。
- [Quan-0505/rust-daed](https://github.com/Quan-0505/rust-daed) / [Quan-0505/daed-kdae](https://github.com/Quan-0505/daed-kdae) — Rust 与 Go 版 daed 独立安装包。
- [Quan-0505/OPNsense-For-R4S](https://github.com/Quan-0505/OPNsense-For-R4S) — NanoPi R4S 的 OPNsense 固件与调优记录。

## 📄 许可

本仓库采用 [GNU General Public License v3.0](LICENSE)。DaeNext、DaedNext、KixDNS、Footstrap 等组件遵循各自上游许可证，具体以对应项目的许可文件为准。
