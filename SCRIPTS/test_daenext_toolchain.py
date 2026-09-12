"""Exercise OpenWrt GCC layouts without rebuilding the toolchain."""
import os
from pathlib import Path
import subprocess
import sys
import tempfile

bash = sys.argv[1] if len(sys.argv) > 1 else 'bash'
helper = Path(__file__).with_name('daenext_toolchain.sh').resolve()
with tempfile.TemporaryDirectory(prefix='daenext-toolchain-') as temp:
    root = Path(temp)
    toolchain = root / 'openwrt/staging_dir/toolchain-test'
    (toolchain / 'bin').mkdir(parents=True)
    (toolchain / 'include').mkdir()
    (toolchain / 'include/stdio.h').write_text('/* musl test header */')
    prefix = toolchain / 'bin/aarch64-openwrt-linux-musl-'
    for tool in ('gcc', 'g++', 'ar', 'readelf', 'strip'):
        path = Path(str(prefix) + tool)
        path.write_text('#!/bin/sh\nprintf "%s\\n" "${DAENEXT_TEST_SYSROOT:-}"\n', encoding='utf-8')
        path.chmod(0o755)
    command = [bash, '-c', 'set -euo pipefail; ROOT="$1"; source "$2"',
               'test', root.as_posix(), helper.as_posix()]
    for label, sysroot in [('empty sysroot', ''), ('stale sysroot', '/nonexistent/daenext-sysroot'),
                           ('explicit sysroot', toolchain.as_posix())]:
        env = dict(os.environ, DAENEXT_TEST_SYSROOT=sysroot)
        result = subprocess.run(command, env=env, text=True, capture_output=True)
        assert result.returncode == 0, (label, result.stderr)
        assert 'DaeNext target headers:' in result.stdout
        print('PASS:', label)
    (toolchain / 'include/stdio.h').unlink()
    result = subprocess.run(command, text=True, capture_output=True)
    assert result.returncode != 0 and 'musl headers missing' in result.stderr
    print('PASS: missing headers fail with a diagnostic')
