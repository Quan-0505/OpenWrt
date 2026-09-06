# -*- coding: utf-8 -*-
import os, gzip, shutil
base = r'C:\Users\quan\Desktop\NetProxy-optimized-config-share\404-OpenWrt\yaof-upstream'
web = os.path.join(base, 'PATCH', 'daed-web')
# A) gunzip every .gz in PATCH/daed-web, then remove the .gz
count = 0
for root, dirs, files in os.walk(web):
    for f in files:
        if f.endswith('.gz'):
            src = os.path.join(root, f)
            dst = src[:-3]
            try:
                with gzip.open(src, 'rb') as gi, open(dst, 'wb') as go:
                    shutil.copyfileobj(gi, go)
                os.remove(src)
                count += 1
            except Exception as e:
                print('ERR', src, e)
print('gunzipped', count)
# B) add KERNEL_DEBUG_INFO=y to seeds
import re
for dev in ['R2S', 'R3S', 'R4S', 'X86']:
    p = os.path.join(base, 'SEED', dev, 'config.seed')
    t = open(p, encoding='utf-8').read()
    if 'CONFIG_KERNEL_DEBUG_INFO=y' in t:
        continue
    # place right before BTF line if present else append after KERNEL_DEBUG_INFO_BTF
    if 'CONFIG_KERNEL_DEBUG_INFO_BTF=y' in t:
        t = t.replace('CONFIG_KERNEL_DEBUG_INFO_BTF=y', 'CONFIG_KERNEL_DEBUG_INFO=y\nCONFIG_KERNEL_DEBUG_INFO_BTF=y', 1)
    else:
        t = t + '\nCONFIG_KERNEL_DEBUG_INFO=y\nCONFIG_KERNEL_DEBUG_INFO_BTF=y\n'
    open(p, 'w', encoding='utf-8', newline='\n').write(t)
    print(dev, 'kernel debug info added')
