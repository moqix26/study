import os, re, urllib.parse
root = r'f:\study'
real = set()
for dp, dn, fn in os.walk(root):
    if '.cursor.memory' in dp:
        continue
    for f in fn:
        if f.endswith('.md'):
            p = os.path.relpath(os.path.join(dp, f), root).replace('\\', '/')
            real.add(p)

pat = re.compile(r'\]\(([^)]+\.md)(?:[^)]*)?\)')
broken = []
seen = set()
for dp, dn, fn in os.walk(root):
    if '.cursor.memory' in dp:
        continue
    for f in fn:
        if not f.endswith('.md'):
            continue
        fp = os.path.join(dp, f)
        try:
            txt = open(fp, encoding='utf-8').read()
        except Exception:
            continue
        for m in pat.finditer(txt):
            link = m.group(1)
            if link.startswith('http'):
                continue
            link0 = link.split('#')[0]
            link0 = urllib.parse.unquote(link0)  # decode %20 -> space
            src_dir = os.path.dirname(fp)
            tgt = os.path.normpath(os.path.join(src_dir, link0)).replace('\\', '/')
            rel = os.path.relpath(tgt, root).replace('\\', '/')
            if rel not in real:
                key = (os.path.relpath(fp, root).replace('\\', '/'), link)
                if key not in seen:
                    seen.add(key)
                    broken.append((os.path.relpath(fp, root).replace('\\', '/'), link, rel))

out = open(os.path.join(root, '.cursor.memory', 'broken-links.txt'), 'w', encoding='utf-8')
for src, link, rel in broken:
    out.write(f'{src}  ->  {link}  (resolved: {rel})\n')
out.write(f'\nTotal broken: {len(broken)}\n')
out.close()
print(f'Done. Broken links: {len(broken)}')
