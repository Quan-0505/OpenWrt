# -*- coding: utf-8 -*-
import paramiko, time
c = paramiko.SSHClient()
c.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ok = False
for a in range(5):
    try:
        c.connect('192.168.5.44', port=22, username='root', password='liquansen', timeout=60, banner_timeout=60, auth_timeout=60, look_for_keys=False, allow_agent=False)
        ok = True; break
    except Exception as e:
        print('retry', a, str(e)[:90]); time.sleep(8)
if not ok: raise SystemExit('ssh down')
def run(cmd, timeout=300):
    stdin, stdout, stderr = c.exec_command(cmd, timeout=timeout)
    out = stdout.read().decode('utf-8','replace')
    err = stderr.read().decode('utf-8','replace')
    return (out + ('\n[err] '+err if err.strip() else '')).strip()
print('=== df ==='); print(run('df -h / /build /home 2>/dev/null'))
print('=== lsblk ==='); print(run('lsblk -o NAME,SIZE,FSTYPE,MOUNTPOINT'))
print('=== memory ==='); print(run('free -h | head -3'))
print('=== /build 分布 (一级, GB) ==='); print(run('du -xsh /build/* 2>/dev/null | sort -rh | head -16'))
print('=== 系统一级 (GB) ==='); print(run('du -xsh /root /home /var /usr /opt /srv 2>/dev/null | sort -rh'))
print('=== /var 细分 ==='); print(run('du -xsh /var/cache /var/lib/docker /var/log /var/tmp 2>/dev/null | sort -rh'))
c.close()
