<p align="center">
<h1 align="center">OpenWrt（Quan-0505 精简定制版）</h1>
<p align="center">
<img src="https://github.com/Quan-0505/OpenWrt/actions/workflows/R2S-OpenWrt.yml/badge.svg">
<img src="https://github.com/Quan-0505/OpenWrt/actions/workflows/R3S-OpenWrt.yml/badge.svg">
<img src="https://github.com/Quan-0505/OpenWrt/actions/workflows/R4S-OpenWrt.yml/badge.svg">
<img src="https://github.com/Quan-0505/OpenWrt/actions/workflows/X86-OpenWrt.yml/badge.svg">
</p>
<p align="center">
<img alt="GitHub All Releases" src="https://img.shields.io/github/downloads/Quan-0505/OpenWrt/total?style=for-the-badge">
<img alt="GitHub" src="https://img.shields.io/github/license/Quan-0505/OpenWrt?style=for-the-badge">
</p>

<h3 align="center">基于 QiuSimons/YAOF 的 OpenWrt 25.12 精简固件（kixdns + DaeNext Rust 版 daed）</h3>

> ⚠️ 本仓库是 [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF) 的公开 fork（GPL-3.0）。只做了**应用裁剪 + 默认值调整**，保留上游机型级优化与构建链。请勿用于商业用途。

### 特性

- 基于原生 OpenWrt 25.12 编译，**默认管理地址 192.168.2.1**（避免与上级路由 192.168.1.1 冲突）
- 第三方应用只保留两个：
  - **kixdns**（Rust 高性能 DNS 分流转发器，LuCI 可视化配置，作者预编译 musl 静态版）
  - **daed**（[Quan-0505/rust-daed](https://github.com/Quan-0505/rust-daed) v3.1.1-sticky，**DaedNext Rust 原生** eBPF 透明代理，Aya + BoringSSL + sticky-ip 增强）
- **daed 无 LuCI 菜单，独立网页管理**：浏览器访问 `http://<路由器IP>:2023`（WebUI 已内置，随固件打包）
- **LuCI 主题 footstrap**（[VizzleTF/luci-theme-footstrap](https://github.com/VizzleTF/luci-theme-footstrap)），官方无中文，**本仓库额外编译了完整中文语言包**，默认即启用
- 内核开启 **BTF**（`KERNEL_DEBUG_INFO_BTF`）——Rust daed 的 eBPF 加载必需
- 其余上游插件（SSRP/PassWall/OpenClash/Mihomo/HomeProxy/MosDNS/DDNS/FRP/Zerotier/SQM/京东/网易云等）**全部移除**
- 保留 YAOF 机型级优化：BBRv3、LRNG、FullCone、Shortcut-FE 流量分载、`mitigations=off`、O2 编译、网卡 rx/tx 缓冲加大、R2S/R4S 频率与 PWM 风扇支持等
- 内置升级功能可用，物理 Reset 按键可用
- 支持机型：**R2S、R3S、R4S、X86_64**（R2C 已移除）
- 每机型产物：`<机型>-OpenWrt-<日期>-<OpenWrt版本>-ext4.zip` 与 `...-sfs.zip`

### 下载

- 到 [Releases](https://github.com/Quan-0505/OpenWrt/releases) 选择设备对应固件下载（ext4 / sfs 按需选择）

### 自行编译（GitHub Actions）

1. Actions → 选择 `R2S-OpenWrt` / `R3S-OpenWrt` / `R4S-OpenWrt` / `X86-OpenWrt` 其中一个 **Run workflow**（只编译该机型）；
   或选择 `OpenWrt-Matrix`，在 `targets` 填 `R2S,R3S,R4S,X86`（留空 = 全机型并行）。
2. 等待编译完成（冷编译约 3~5 小时/机型；已缓存下载源 dl，可省源码下载时间）。
3. 结果自动发布到 [Releases](https://github.com/Quan-0505/OpenWrt/releases)，`25.12.5` tag 覆盖更新；每次编译产物同时保留为 run artifact 作为安全网。

### 本 fork 相对上游的改动

| 改动 | 说明 |
|------|------|
| `SEED/*/config.seed` | 裁剪为「kixdns + daed」极简应用集；`dae/luci-app-dae` → `daed`（Rust 版）；开启 `KERNEL_DEBUG_INFO_BTF` |
| `SCRIPTS/02_prepare_package.sh` | 挂载 kixdns feed（v1.5.3）；拉取并整包替换为 `rust-daed` apk 载荷；拉取 footstrap 主题 + 覆盖中文语言包 |
| `.github/workflows/OpenWrt-Matrix.yml` | 移除 R2C；下载 kixdns 预编译二进制；编译时解包 `rust-daed-<设备>.apk` 并注入 WebUI 资产；发布前清理本设备旧 release 资产 |
| `.github/workflows/build-daenext.yml` | 可选：musl 静态交叉编译 DaeNext 核心（默认走 rust-daed 官方 apk） |
| `PATCH/daed-pkg/daed` / `PATCH/daed-web` | daed 纯预编译包（无源码编译）与内置 WebUI（v1.28） |
| `PATCH/theme-footstrap-zh` | footstrap 主题的完整中文语言包（113 词条） |
| `PATCH/files/etc/uci-defaults/` | 首次启动设置 LAN 192.168.2.1；默认 LuCI 主题切到 footstrap |

### 鸣谢

[QiuSimons/YAOF](https://github.com/QiuSimons/YAOF) · [OpenWrt](https://github.com/openwrt) · [ImmortalWrt](https://github.com/immortalwrt) · [coolsnowwolf/LEDE](https://github.com/coolsnowwolf/lede) · [daeuniverse](https://github.com/daeuniverse) · [Quan-0505/rust-daed](https://github.com/Quan-0505/rust-daed) · [olicesx/kixdns](https://github.com/olicesx/kixdns) · [JohnsonRan/luci-app-kixdns](https://github.com/JohnsonRan/luci-app-kixdns) · [VizzleTF/luci-theme-footstrap](https://github.com/VizzleTF/luci-theme-footstrap)
