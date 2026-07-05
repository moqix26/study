import os

root = r'f:\study'

# 修复残留断链
fixes = [
    # 计网文件中 ../../../后端学习/Java/04 -> ../../后端学习/Java/04 (上一步误伤)
    ("../../../后端学习/Java/04-SpringBoot核心开发.md", "../../后端学习/Java/04-SpringBoot核心开发.md"),
    # C++/11 中 ../../Python/ -> ../Python/ 和 ../../Java/ -> ../Java/
    ("../../Python/09-LinuxDockerNginx部署基础.md", "../Python/09-LinuxDockerNginx部署基础.md"),
    ("../../Java/09-LinuxDockerNginx部署基础.md", "../Java/09-LinuxDockerNginx部署基础.md"),
    # Java/15 中 ../../Linux/00 -> ../Linux/00
    ("../../Linux/00-学习路线图与说明.md", "../Linux/00-学习路线图与说明.md"),
]

changed = 0
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
        orig = txt
        for old, new in fixes:
            if old in txt:
                txt = txt.replace(old, new)
        if txt != orig:
            open(fp, 'w', encoding='utf-8').write(txt)
            changed += 1
            print(f'Fixed: {os.path.relpath(fp, root)}')

print(f'\nFiles changed: {changed}')
