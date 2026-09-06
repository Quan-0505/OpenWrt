# -*- coding: utf-8 -*-
import paramiko, time
c = paramiko.SSHClient()
c.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ok = False
for a in range(4):
    try:
        c.connect('192.168.2.1', port=22, username='root', password='liquansen', timeout=20, banner_timeout=30, auth_timeout=30, look_for_keys=False, allow_agent=False)
        ok = True; break
    except Exception as e:
        print('conn retry', a, str(e)[:120]); time.sleep(5)
if not ok: raise SystemExit('cannot ssh 192.168.2.1')
def run(cmd, timeout=60):
    stdin, stdout, stderr = c.exec_command(cmd, timeout=timeout)
    out = stdout.read().decode('utf-8','replace')
    err = stderr.read().decode('utf-8','replace')
    return (out + ('\n[err] '+err if err.strip() else '')).strip()
print('--- uname ---'); print(run('uname -a'))
print('--- daed binary ---'); print(run('ls -l /usr/bin/daed /etc/init.d/daed; apk list 2>/dev/null | grep -i daed'))
print('--- daemon proc ---'); print(run('ps w | grep -i daed | grep -v grep'))
print('--- init start ---'); print(run('/etc/init.d/daed status 2>&1 || true; /etc/init.d/daed enabled && echo ENABLED || echo not-enabled'))
print('--- port 2023 ---'); print(run('netstat -tlnp 2>/dev/null | grep 2023 || ss -tlnp 2>/dev/null | grep 2023 || echo no-2023'))
print('--- web assets ---'); print(run('ls -l /usr/share/daed/web 2>&1 | head -8; echo ---; ls /usr/share/daed/web 2>/dev/null | wc -l'))
print('--- geo ---'); print(run('ls -l /usr/share/daed/*.dat 2>&1'))
print('--- BTF ---'); print(run('ls -l /sys/kernel/btf/vmlinux 2>&1'))
print('--- daed log ---'); print(run('logread 2>/dev/null | grep -i daed | tail -25'))
print('--- curl 2023 ---'); print(run('curl -sS -m 5 -o /dev/null -w "HTTP %{http_code} size=%{size_download}\n" http://127.0.0.1:2023/ 2>&1; curl -sS -m 5 http://127.0.0.1:2023/ 2>&1 | head -c 400'))
c.close()
