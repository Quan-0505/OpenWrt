include $(TOPDIR)/rules.mk

PKG_NAME:=daed
PKG_VERSION:=3.1.0
PKG_RELEASE:=2
PKG_LICENSE:=AGPL-3.0-only MIT
PKG_MAINTAINER:=Quan-0505

include $(INCLUDE_DIR)/package.mk

define Package/daed
  SECTION:=net
  CATEGORY:=Network
  TITLE:=DaeNext Rust daemon with DaedNext 3.1.0 WebUI
  URL:=https://github.com/ksong008/DaeNext
  DEPENDS:=+ca-bundle +kmod-veth +kmod-sched-bpf +kmod-sched-core
endef

define Package/daed/description
  DaeNext built from pinned upstream source for OpenWrt musl, with the
  matching DaedNext 3.1.0 frontend. Native eBPF and BoringSSL enabled.
endef

define Package/daed/conffiles
/etc/config/daed
/etc/daed/
endef

define Build/Prepare
endef
define Build/Configure
endef
define Build/Compile
	test -s $(CURDIR)/prebuilt-data/usr/bin/daed
	test -s $(CURDIR)/prebuilt-data/usr/share/daed/web/index.html
endef

define Package/daed/install
	$(CP) $(CURDIR)/prebuilt-data/. $(1)/
endef

$(eval $(call BuildPackage,daed))
