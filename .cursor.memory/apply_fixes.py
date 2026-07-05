import os

root = r'f:\study'

# ===== 全局文件名替换（安全：旧文件名在任何地方都不存在）=====
global_renames = [
    # HTML CSS JS 重命名
    ("04-CSS布局Flex与Grid.md", "05-CSS布局FlexGrid响应式与动画.md"),
    ("07-JavaScript流程控制函数与ES6基础.md", "07-JavaScript流程控制函数对象数组与ES6基础.md"),
    ("08-JavaScriptDOM操作与事件处理.md", "08-JavaScript-DOM-BOM与事件机制.md"),
    ("09-JavaScript异步编程与Fetch.md", "09-JavaScript异步编程网络请求与本地存储.md"),
    ("08-JavaScript异步编程与Promise.md", "09-JavaScript异步编程网络请求与本地存储.md"),
    ("10-Web网络入门与浏览器调试.md", "10-浏览器HTTP网络与Web基础.md"),
    ("09-JavaScript模块化与工程入门.md", "11-前端工程化调试Git与包管理基础.md"),
    ("09-JavaScript模块化.md", "11-前端工程化调试Git与包管理基础.md"),
    # Vue 重命名
    ("06-生命周期与内置组件.md", "05-组合式API与script-setup.md"),
    ("08-前后端联调与Axios封装.md", "08-Axios网络请求与前后端联调.md"),
    # 计算机网络 重命名
    ("04-HTTP协议与请求响应.md", "04-HTTP协议深入.md"),
    ("05-DNS与域名解析.md", "03-IP地址与DNS解析.md"),
    ("04-HTTPS与TLS安全.md", "05-HTTPS与TLS加密.md"),
    # 计网 中不存在的 CORS/安全章 -> Web安全系列
    ("05-跨域CORS与开发代理.md", "../Web安全/05-CORS与同源策略安全.md"),
    ("06-Web安全与HTTPS实战.md", "../Web安全/04-HTTPS与传输安全实战.md"),
    # Java 重命名
    ("05-MyBatis与数据库操作.md", "05-MyBatis事务与接口工程化.md"),
    ("06-Redis缓存与分布式锁.md", "07-Redis核心原理与缓存实战.md"),
    ("07-Redis缓存与分布式锁.md", "07-Redis核心原理与缓存实战.md"),
    ("12-微服务与分布式架构入门.md", "12-高并发与分布式系统基础.md"),
    ("08-RabbitMQ消息队列与Spring整合.md", "08-RabbitMQ与消息队列实战.md"),
    ("04-SpringBoot入门与RESTful接口.md", "04-SpringBoot核心开发.md"),
    ("11-微服务架构与SpringCloud入门.md", "11-微服务与SpringCloud基础.md"),
    ("14-高频场景题与面试专题.md", "14-高频场景设计与面试专题.md"),
    ("06-MySQL与数据库基础.md", "06-MySQL基础索引与事务.md"),
    # AIAgent 连字符
    ("02-Spring-AI核心开发.md", "02-SpringAI核心开发.md"),
    ("04-Function-Calling与Tool设计.md", "04-FunctionCalling与Tool设计.md"),
    # Linux 重命名
    ("07-网络基础与SSH.md", "07-网络命令与防火墙基础.md"),
    ("05-用户权限与sudo.md", "05-用户组与文件权限.md"),
    # Python 重命名
    ("09-Linux-Docker-Nginx部署基础.md", "09-LinuxDockerNginx部署基础.md"),
    ("07-Redis缓存与消息队列.md", "07-Redis核心原理与缓存实战.md"),
    # TypeScript 顿号
    ("05-类、枚举与类型收窄.md", "05-类枚举与类型收窄.md"),
]

# ===== 路径深度修复（仅在特定源目录生效）=====
# (源目录前缀, 旧路径, 新路径)
scoped_fixes = [
    # 后端学习/X/ 里的 ../../数据结构/ -> ../数据结构/
    ("后端学习/", "../../数据结构/", "../数据结构/"),
    # 后端学习/Linux/ 里的 ../../Java/ -> ../Java/
    ("后端学习/Linux/", "../../Java/", "../Java/"),
    # 后端学习/Linux/ 里的 ../../Python/ -> ../Python/
    ("后端学习/Linux/", "../../Python/", "../Python/"),
    # 前端学习/React/ 里的 ../../计算机网络/ -> ../计算机网络/
    ("前端学习/React/", "../../计算机网络/", "../计算机网络/"),
    # 前端学习/Vue/ 里的 ../../计算机网络/ -> ../计算机网络/
    ("前端学习/Vue/", "../../计算机网络/", "../计算机网络/"),
    # 前端学习/Vue/ 里的 ../../后端学习/Java/ -> ../../后端学习/Java/ (already correct depth? check)
    # 计网 06 里 ../后端学习/Java/ -> ../../后端学习/Java/
    ("前端学习/计算机网络/", "../后端学习/Java/04-SpringBoot核心开发.md", "../../后端学习/Java/04-SpringBoot核心开发.md"),
]

# ===== 特殊单点修复 =====
special_fixes = [
    # 修改规范.md (root) 路径
    ("修改规范.md", "../系统设计/00-学习路线图与说明.md", "后端学习/系统设计/00-学习路线图与说明.md"),
    ("修改规范.md", "../../前端学习/Web安全/07-LLM应用安全与Prompt注入防护.md", "前端学习/Web安全/07-LLM应用安全与Prompt注入防护.md"),
    ("修改规范.md", "../../前端学习/计算机网络/00-学习路线图与说明.md", "前端学习/计算机网络/00-学习路线图与说明.md"),
    ("修改规范.md", "../../todo.md", "todo.md"),
    ("修改规范.md", "../Java/16-SSE与WebSocket实时通信.md", "后端学习/Java/16-SSE与WebSocket实时通信.md"),
    # 系统设计 -> Vue (wrong: ../Vue/ -> correct: ../../前端学习/Vue/)
    ("系统设计/00-学习路线图与说明.md", "../Vue/00-学习路线图与说明.md", "../../前端学习/Vue/00-学习路线图与说明.md"),
    # Web安全 -> AIAgent (wrong: ../AIAgent/ -> correct: ../../后端学习/AIAgent/)
    ("Web安全/00-学习路线图与说明.md", "../AIAgent/00-学习路线图与说明.md", "../../后端学习/AIAgent/00-学习路线图与说明.md"),
    # 计网 01 的"（待写）"标记（内容已移至 Web安全）
    ("计算机网络/01-网络分层与通信基础.md", "（待写）", ""),
]

changed_files = {}
total_changes = 0

for dp, dn, fn in os.walk(root):
    if '.cursor.memory' in dp:
        continue
    for f in fn:
        if not f.endswith('.md'):
            continue
        fp = os.path.join(dp, f)
        rel = os.path.relpath(fp, root).replace('\\', '/')
        try:
            txt = open(fp, encoding='utf-8').read()
        except Exception:
            continue
        orig = txt
        changes = []

        # 1. 全局文件名替换
        for old, new in global_renames:
            if old in txt:
                cnt = txt.count(old)
                txt = txt.replace(old, new)
                changes.append(f"global: {old} -> {new} ({cnt}x)")

        # 2. 路径深度修复
        for prefix, old, new in scoped_fixes:
            if rel.startswith(prefix) and old in txt:
                cnt = txt.count(old)
                txt = txt.replace(old, new)
                changes.append(f"scoped[{prefix}]: {old} -> {new} ({cnt}x)")

        # 3. 特殊单点修复
        for target, old, new in special_fixes:
            if rel.endswith(target) and old in txt:
                cnt = txt.count(old)
                txt = txt.replace(old, new)
                changes.append(f"special[{target}]: {old!r} -> {new!r} ({cnt}x)")

        if txt != orig:
            open(fp, 'w', encoding='utf-8').write(txt)
            changed_files[rel] = changes
            total_changes += sum(int(c.split('(')[-1].rstrip('x)')) for c in changes if 'x)' in c)

# 报告
out = open(os.path.join(root, '.cursor.memory', 'fix-report.txt'), 'w', encoding='utf-8')
for f, changes in sorted(changed_files.items()):
    out.write(f'### {f}\n')
    for c in changes:
        out.write(f'  {c}\n')
    out.write('\n')
out.write(f'Total files changed: {len(changed_files)}\n')
out.write(f'Total replacements: {total_changes}\n')
out.close()
print(f'Files changed: {len(changed_files)}, total replacements: {total_changes}')
