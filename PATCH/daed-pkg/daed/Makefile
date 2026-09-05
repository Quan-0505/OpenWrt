# daed (DaedNext, Rust-native) v3.1.1 sticky - prebuilt package for this fork.
# Payload (usr/, etc/) is unpacked at build time by the firmware workflow from
# the rust-daed release .apk (Quan-0505/rust-daed, tag v3.1.1-sticky) into
# prebuilt-data/. No compilation happens here.

include $(TOPDIR)/rules.mk

PKG_NAME:=daed
PKG_VERSION:=3.1.1
PKG_RELEASE:=1

PKG_LICENSE:=AGPL-3.0-only
PKG_MAINTAINER:=Quan-0505

include $(INCLUDE_DIR)/package.mk

define Package/daed
  SECTION:=net
  CATEGORY:=Network
  TITLE:=daed (DaedNext Rust-native) transparent proxy v3.1.1 sticky
  URL:=https://github.com/Quan-0505/rust-daed
endef

define Package/daed/description
  Rust-native dae/daed product daemon (eBPF transparent proxy, embedded
  WebUI on :2023, sticky-ip build). Static musl binary; WebUI and geo
  data bundled. Requires a BTF-enabled kernel.
endef

define Package/daed/conffiles
/etc/config/daed
endef

define Package/daed/install
	$(INSTALL_DIR) $(1)
	$(CP) $(CURDIR)/prebuilt-data/. $(1)/
endef

$(eval $(call BuildPackage,daed))
