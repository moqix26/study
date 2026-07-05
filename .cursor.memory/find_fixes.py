import os, re, urllib.parse, difflib

root = r'f:\study'

# 收集真实文件 (relative path, basename)
real_files = []  # list of (relpath_posix, basename)
real_set = set()
for dp, dn, fn in os.walk(root):
    if '.cursor.memory' in dp:
        continue
    for f in fn:
        if f.endswith('.md'):
            p = os.path.relpath(os.path.join(dp, f), root).replace('\\', '/')
            real_files.append((p, f))
            real_set.add(p)

def best_match(basename):
    """找真实文件中 basename 最接近的"""
    names = [b for _, b in real_files]
    matches = difflib.get_close_matches(basename, names, n=1, cutoff=0.4)
    if matches:
        for p, b in real_files:
            if b == matches[0]:
                return p
    return None

pat = re.compile(r'\]\(([^)]+\.md)(?:[^)]*)?\)')
fixes = []  # (file, old_link, new_link)
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
            decoded = urllib.parse.unquote(link0)
            src_dir = os.path.dirname(fp)
            tgt = os.path.normpath(os.path.join(src_dir, decoded)).replace('\\', '/')
            rel = os.path.relpath(tgt, root).replace('\\', '/')
            if rel in real_set:
                continue  # not broken
            # 尝试修复：用 basename 模糊匹配
            base = os.path.basename(decoded)
            match = best_match(base)
            if match:
                # 计算从 src_dir 到 match 的相对路径
                src_rel = os.path.relpath(fp, root).replace('\\', '/')
                src_dir_rel = os.path.dirname(src_rel)
                # match 是相对 root 的 posix 路径
                new_rel = os.path.relpath(os.path.join(root, match.replace('/', os.sep)),
                                          os.path.join(root, src_dir_rel.replace('/', os.sep))).replace('\\', '/')
                # 保留原 anchor
                anchor = ''
                if '#' in link:
                    anchor = link[link.index('#'):]
                new_link = new_rel + anchor
                key = (os.path.relpath(fp, root).replace('\\', '/'), link)
                if key not in seen:
                    seen.add(key)
                    fixes.append((os.path.relpath(fp, root).replace('\\', '/'), link, new_link, match))

out = open(os.path.join(root, '.cursor.memory', 'suggested-fixes.txt'), 'w', encoding='utf-8')
for src, old, new, matched in fixes:
    out.write(f'{src}\n  OLD: {old}\n  NEW: {new}\n  -> matches: {matched}\n\n')
out.write(f'\nTotal suggested fixes: {len(fixes)}\n')
out.close()
print(f'Suggested fixes: {len(fixes)}')
