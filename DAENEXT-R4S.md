# R4S DaeNext firmware

R4S builds compile the Rust daemon directly from
`ksong008/DaeNext@218bdf72b5be3e59494513c9805fe701dc0fed04` and the
DaedNext **3.1.0** frontend from
`ksong008/DaedNext@b084f2f6f0062a12d6087ad5243bb4a436d37f45`.
The previous `rust-daed v3.1.1-sticky` payload and checked-in WebUI are not
used for R4S. Other targets keep their existing package selection.

The daemon is compiled with OpenWrt's aarch64 musl toolchain, upstream default
production features (including native eBPF and BoringSSL), and a generic ARMv8
CPU baseline compatible with both CPU clusters on the RK3399. The build checks
the ELF architecture, absence of a dynamic interpreter, and version output via
QEMU. The firmware build must pass the package and kernel checks before release.

`CONFIG_KERNEL_DEBUG_INFO_REDUCED` is disabled because it conflicts with BTF.
The final kernel must actually contain `.BTF`; setting a seed option alone does
not count as verification. Frontend entrypoint JS/CSS must exist uncompressed
both in the package payload and the final firmware root.

Run `R4S-OpenWrt` on this branch to build. Each R4S run publishes to a separate
`<OpenWrt-version>-daenext-<run-id>` release and preserves previous firmware
release assets. Configuration is at `/etc/daed`, logs for new installations
are in RAM at `/tmp/log/daed`, and the WebUI is on port 2023.

This build does not flash a router automatically. Runtime forwarding still
needs to be checked on R4S after flashing and configuring nodes/routing.
