# DaeNext (Rust-native dae/daed) minimal binary package for this fork.
# The daed daemon binary is a fully static musl build produced by
# .github/workflows/build-daenext.yml (ksong008/DaeNext @ work/boringssl)
# and downloaded into prebuilt/daed by the firmware workflow.

include $(TOPDIR)/rules.mk

PKG_NAME:=daed
PKG_VERSION:=a2c9519
PKG_RELEASE:=1

PKG_LICENSE:=AGPL-3.0-only
PKG_MAINTAINER:=Quan-0505

include $(INCLUDE_DIR)/package.mk

define Package/daed
  SECTION:=net
  CATEGORY:=Network
  TITLE:=DaeNext transparent proxy daemon (Rust dae/daed)
  URL:=https://github.com/ksong008/DaeNext
  DEPENDS:=+vmlinux-btf +kmod-sched-core +kmod-sched-bpf +kmod-veth
endef

define Package/daed/description
  Rust-native dae core workspace product daemon (eBPF transparent proxy,
  daed product API). Prebuilt static musl binary; web UI assets shipped
  under /usr/share/daed/web and served by the daemon itself.
endef

define Package/daed/conffiles
/etc/config/daed
endef

define Package/daed/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(CURDIR)/prebuilt/daed $(1)/usr/bin/daed

	$(INSTALL_DIR) $(1)/usr/share/daed/web
	$(CP) $(CURDIR)/web-assets/. $(1)/usr/share/daed/web/

	$(INSTALL_DIR) $(1)/etc/config $(1)/etc/init.d $(1)/etc/daed
	$(INSTALL_CONF) $(CURDIR)/files/daed.config $(1)/etc/config/daed
	$(INSTALL_BIN) $(CURDIR)/files/daed.init $(1)/etc/init.d/daed
endef

$(eval $(call BuildPackage,daed))
