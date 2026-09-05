<div align="center">

# OpenWrt

**Quan-0505 精简定制固件 · kixdns + DaedNext Rust 版 daed + footstrap 中文主题**

[![License](https://img.shields.io/badge/license-GPL--3.0-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-25.12.5-orange.svg)](https://github.com/Quan-0505/OpenWrt/releases/tag/25.12.5)

基于 [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF)（OpenWrt 25.12）+ [rust-daed](https://github.com/Quan-0505/rust-daed) v3.1.1-sticky（DaedNext Rust 原生 dae）/ kixdns，仅保留精简应用集。

</div>

---

## ✨ 特性

- 🦀 **daed（Rust 原生）**：DaedNext 全 Rust eBPF 透明代理（Aya + BoringSSL + sticky-ip），WebUI 内置，独立 `:2023` 管理
- 📌 **kixdns**：Rust 高性能 DNS 分流转发器，LuCI 可视化配置，作者预编译 musl 静态版
- 🎨 **footstrap 主题 + 中文语言包**：官方无中文，本仓库编译了完整 113 词条中文包，默认启用
- 🧠 **内核 BTF**：`KERNEL_DEBUG_INFO_BTF` 开启（Rust daed eBPF 加载必需）
- ⚡ **YAOF 机型级优化**：BBRv3、LRNG、FullCone、Shortcut-FE、mitigations=off、O2 编译、网卡缓冲加大、R2S/R4S 频率与 PWM 风扇等
- 🧹 **极简**：其余上游插件（SSRP/PassWall/OpenClash/Mihomo/MosDNS/DDNS/FRP/Zerotier 等）全部移除
- 🔄 内置升级可用、物理 Reset 可用、默认管理 IP **192.168.2.1**
- 机型：**R2S / R3S / R4S / X86_64**（R2C 已移除）

## 📦 固件（[25.12.5 Release](https://github.com/Quan-0505/OpenWrt/releases/tag/25.12.5)，ext4.zip + sfs.zip 统一发布）

| 设备 | 文件 | 架构 |
|---|---|---|
| OpenWrt X86 软路由 | `X86-OpenWrt-*-25.12.5-{ext4,sfs}.zip` | x86_64 |
| NanoPi R4S | `R4S-OpenWrt-*-25.12.5-{ext4,sfs}.zip` | aarch64_cortex-a72 |
| NanoPi R3S | `R3S-OpenWrt-*-25.12.5-{ext4,sfs}.zip` | aarch64_cortex-a53 |
| NanoPi R2S | `R2S-OpenWrt-*-25.12.5-{ext4,sfs}.zip` | aarch64_cortex-a53 |

> ext4 = 可扩容分区格式（推荐）；sfs = squashfs 只读根文件系统（恢复出厂更干净）。

## 🚀 快速开始

```sh
# 到 Releases 下载对应设备固件，刷入（R2S/R3S/R4S 用 TF 卡/SD 或网线刷写，X86 用写盘工具）
# 默认后台
地址: http://192.168.2.1
LuCI:  http://192.168.2.1
daed 面板: http://192.168.2.1:2023   # 无 LuCI 菜单，独立网页
```

刷入后 LuCI 默认主题即 footstrap（中文）；daed 守护进程默认开机自启。

## 📋 系统要求

- 硬件：NanoPi R2S/R3S/R4S 或 x86_64 软路由
- 内核 ≥ 5.8 且启用 **BTF**（本固件已开启，勿在关闭 BTF 的内核上运行 daed）
- OpenWrt 25.12 / apk 格式

## ⚙️ 本 fork 相对上游的改动

| 改动 | 说明 |
|---|---|
| `SEED/*/config.seed` | 裁剪为「kixdns + daed」；`dae/luci-app-dae` → `daed`（Rust 版）；开启 `KERNEL_DEBUG_INFO_BTF` |
| `SCRIPTS/02_prepare_package.sh` | 挂载 kixdns feed（v1.5.3）；拉取 footstrap 主题并覆盖中文语言包 |
| `.github/workflows/OpenWrt-Matrix.yml` | 移除 R2C；编译时解包 `rust-daed-<设备>.apk` 载荷并注入 WebUI `PATCH/daed-web`；发布前清理本设备旧 release 资产 |
| `PATCH/daed-pkg`, `PATCH/daed-web` | daed 纯预编译包与内置 WebUI（v1.28） |
| `PATCH/theme-footstrap-zh` | footstrap 主题中文语言包（113 词条） |
| `PATCH/files/etc/uci-defaults/` | LAN=192.168.2.1；默认主题 footstrap |

## 🔧 自行编译（GitHub Actions）

1. Actions → `R2S/R3S/R4S/X86-OpenWrt` 任一 **Run workflow**（只编该机型）；或 `OpenWrt-Matrix` 的 `targets` 填 `R2S,R3S,R4S,X86`（留空 = 全机型并行）。
2. 冷编译约 3~5 小时/机型（已缓存下载源 dl）；发布到 `25.12.5` tag 并保留 run artifact 作为安全网。

## 📄 许可

[GNU General Public License v3.0](LICENSE)。上游 [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF) GPL-3.0；[rust-daed](https://github.com/Quan-0505/rust-daed) AGPL-3.0。

---

*daed 独立包（deb/apk）见 [rust-daed](https://github.com/Quan-0505/rust-daed)；Go 版（kdae 引擎）见 [daed-kdae](https://github.com/Quan-0505/daed-kdae)。*
