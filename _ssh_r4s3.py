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
        print('retry', a, str(e)[:120]); time.sleep(5)
if not ok: raise SystemExit('ssh fail')
def run(cmd, timeout=90):
    stdin, stdout, stderr = c.exec_command(cmd, timeout=90)
    out = stdout.read().decode('utf-8','replace')
    err = stderr.read().decode('utf-8','replace')
    return (out + ('\n[err] '+err if err.strip() else '')).strip()
print('--- grep web root ---')
print(run('grep -R "web_root" /etc/daed 2>/dev/null; ls /etc/daed'))
print('--- gunzip all .gz in web (keep .gz, add plain) ---')
print(run('cd /usr/share/daed/web && find . -name "*.gz" -exec sh -c \'gunzip -k -f "$1" 2>/dev/null; chmod 644 "${1%.gz}" 2>/dev/null\' _ {} \\; ; echo done'))
print('--- verify js now ---')
print(run('ls -l /usr/share/daed/web/assets/index-CShzcAqz.js 2>&1'))
print('--- curl js ---')
print(run('curl -sS -m 5 -o /dev/null -w "JS HTTP %{http_code} size=%{size_download} type=%{content_type}\n" http://127.0.0.1:2023/assets/index-CShzcAqz.js 2>&1'))
print('--- curl index ---')
print(run('curl -sS -m 5 -o /dev/null -w "idx HTTP %{http_code} size=%{size_download} type=%{content_type}\n" http://127.0.0.1:2023/ 2>&1'))
c.close()
