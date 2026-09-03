<p align="center">
<h1 align="center">YAOF（Quan-0505 精简定制版）</h1>
<p align="center">
<img src="https://github.com/Quan-0505/YAOF/workflows/R2S-OpenWrt/badge.svg">
<img src="https://github.com/Quan-0505/YAOF/workflows/R3S-OpenWrt/badge.svg">
<img src="https://github.com/Quan-0505/YAOF/workflows/R4S-OpenWrt/badge.svg">
<img src="https://github.com/Quan-0505/YAOF/workflows/X86-OpenWrt/badge.svg">
</p>
<p align="center">
<img alt="GitHub All Releases" src="https://img.shields.io/github/downloads/Quan-0505/YAOF/total?style=for-the-badge">
<img alt="GitHub" src="https://img.shields.io/github/license/Quan-0505/YAOF?style=for-the-badge">
</p>

<h3 align="center">基于 QiuSimons/YAOF 的 OpenWrt 25.12 精简固件（只保留 kixdns + daed）</h3>

> ⚠️ 本仓库是 [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF) 的公开 fork（GPL-3.0）。仅做了**应用裁剪 + 默认值调整**，未改动上游的内核级优化与构建链。请勿用于商业用途。

### 特性

- 基于原生 OpenWrt 25.12 编译，**默认管理地址 192.168.2.1**（避免与上级路由 192.168.1.1 冲突）
- 第三方应用只保留两个：
  - **kixdns**（Rust 高性能 DNS 分流转发器，LuCI 可视化配置，预编译 musl 静态版）
  - **daed**（dae + wing + 新版 Web UI 的一体化 eBPF 透明代理）
- 其余上游插件（SSRP/PassWall/OpenClash/Mihomo/HomeProxy/MosDNS/DDNS/FRP/Zerotier/SQM/京东/网易云等）**全部移除**
- 保留 YAOF 机型级优化：BBRv3、LRNG、FullCone、Shortcut-FE 流量分载、`mitigations=off`、O2 编译、网卡 rx/tx 缓冲加大、R2S/R4S 频率与 PWM 风扇支持等
- 内置升级功能可用，物理 Reset 按键可用
- 支持机型：**R2S、R3S、R4S、X86_64**（R2C 已移除）
- 每个机型产物：`<机型>-OpenWrt-<日期>-<OpenWrt版本>-ext4.zip` 与 `...-sfs.zip`

### 下载

- 到 [Releases](https://github.com/Quan-0505/YAOF/releases) 选择设备对应固件下载（ext4 / sfs 按需选择）

### 自行编译（GitHub Actions）

1. Actions → 选择 `R2S-OpenWrt` / `R3S-OpenWrt` / `R4S-OpenWrt` / `X86-OpenWrt` 其中一个 **Run workflow**（只编译该机型）；
   或选择 `OpenWrt-Matrix`，在 `targets` 输入框填 `R2S,R3S,R4S,X86`（留空 = 全机型并行编译）。
2. 等待编译完成（受 GitHub Actions 队列与缓存影响，一般 1~3 小时/机型，第二次起走缓存会快很多）。
3. 编译结果自动发布到本仓库 [Releases](https://github.com/Quan-0505/YAOF/releases)，同名 tag 自动覆盖更新。

### 本 fork 相对上游的改动

| 改动 | 说明 |
|------|------|
| `SEED/*/config.seed` | 裁剪为「仅 kixdns + daed」极简应用集；`dae/luci-app-dae` 换成 `daed/luci-app-daed` |
| `SCRIPTS/02_prepare_package.sh` | 挂载 `JohnsonRan/luci-app-kixdns` feed（v1.5.3） |
| `.github/workflows/OpenWrt-Matrix.yml` | 移除 R2C；下载 kixdns 预编译静态二进制；产物前缀改为 `<机型>-OpenWrt` |
| `PATCH/files/etc/uci-defaults/90-404-lanip.sh` | 首次启动把 LAN 默认 IP 设为 192.168.2.1 |
| 其余 | 删除 R2C workflow/seed、无用 ssrplus 列表文件 |

### 鸣谢

[QiuSimons/YAOF](https://github.com/QiuSimons/YAOF) · [OpenWrt](https://github.com/openwrt) · [ImmortalWrt](https://github.com/immortalwrt) · [coolsnowwolf/LEDE](https://github.com/coolsnowwolf/lede) · [daeuniverse](https://github.com/daeuniverse) · [olicesx/kixdns](https://github.com/olicesx/kixdns) · [JohnsonRan/luci-app-kixdns](https://github.com/JohnsonRan/luci-app-kixdns)
