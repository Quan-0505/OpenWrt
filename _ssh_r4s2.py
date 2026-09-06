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
def run(cmd, timeout=60):
    stdin, stdout, stderr = c.exec_command(cmd, timeout=60)
    out = stdout.read().decode('utf-8','replace')
    err = stderr.read().decode('utf-8','replace')
    return (out + ('\n[err] '+err if err.strip() else '')).strip()
print('--- assets 内容 ---'); print(run('ls -l /usr/share/daed/web/assets/ | head -8; echo "count=$(ls /usr/share/daed/web/assets | wc -l)"'))
print('--- index引用的js存在? ---'); print(run('ls -l /usr/share/daed/web/assets/ 2>/dev/null | grep -E "index-CShzcAqz" || echo "index-CShzcAqz.js NOT found"'))
print('--- 其 .gz 形式? ---'); print(run('ls -l /usr/share/daed/web/assets/ | grep -E "index.*gz" || echo none'))
print('--- curl js ---'); print(run('curl -sS -m 5 -o /dev/null -w "JS HTTP %{http_code} size=%{size_download}\n" http://127.0.0.1:2023/assets/index-CShzcAqz.js 2>&1'))
print('--- logo webp ---'); print(run('curl -sS -m 5 -o /dev/null -w "logo HTTP %{http_code} size=%{size_download}\n" http://127.0.0.1:2023/logo.webp 2>&1'))
c.close()
