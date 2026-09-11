"""Fail the firmware build if entrypoint JS/CSS are absent or compressed-only."""
import sys
from html.parser import HTMLParser
from pathlib import Path


class Entry(HTMLParser):
    def __init__(self):
        super().__init__()
        self.assets = []

    def handle_starttag(self, tag, attrs):
        attrs = dict(attrs)
        if tag == 'script' and attrs.get('src'):
            self.assets.append(attrs['src'])
        if tag == 'link' and attrs.get('rel') in ('stylesheet', 'modulepreload'):
            self.assets.append(attrs['href'])


root = Path(sys.argv[1]).resolve()
entry = Entry()
entry.feed((root / 'index.html').read_text(encoding='utf-8'))
assert entry.assets, 'No frontend entrypoints'
for ref in entry.assets:
    asset = (root / ref.removeprefix('./').lstrip('/')).resolve()
    assert asset.is_relative_to(root), ref
    assert asset.is_file() and asset.stat().st_size, f'Missing asset: {ref}'
    assert not asset.read_bytes().lstrip().lower().startswith(b'<!doctype'), ref
print(f'Validated {len(entry.assets)} uncompressed frontend entrypoints in {root}')
