# 前端工程化、调试、Git 与包管理基础

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0 导读、FAQ≥10、闭卷自测、费曼检验 -->

## 0. 读前导读（零基础也能跟上）

> **读者假设**：已完成 HTML/CSS/JS 01～10；本章把「写页面」升级到「像工程师一样组织、调试、管理代码」。

### 0.1 用一句话弄懂本章

**一句话**：前端不只是三件套——还要会 **DevTools 排错、Git 记历史、npm 管依赖**，为学 Vue/React 打工程化底子。

**生活类比**：

| 概念 | 类比 |
|------|------|
| **DevTools** | 网页的 X 光机：看结构、样式、网络、JS |
| **Git** | 游戏的存档点：改坏了能回到上一版 |
| **npm** | 应用商店：下载别人写好的工具库 |
| **package.json** | 项目说明书：依赖清单和快捷命令 |

---

### 0.2 你需要提前知道什么

| 水平 | 建议 |
|------|------|
| 只会写单页 HTML | 先完成 [08 DOM](../HTML%20CSS%20JS/08-JavaScript-DOM-BOM与事件机制.md) |
| 目标 Vue 路线 | 本章 + [12 实战](../HTML%20CSS%20JS/12-前端页面实战组件思维与常见模块.md) 后进 [Vue 01](../Vue/01-Vue入门与环境搭建.md) |
| Git 要系统学 | 继续 [Git 00～05](../Git/00-学习路线图与说明.md) |

---

### 0.3 本章知识地图（☐→☑）

- [ ] 熟练 DevTools 四大面板
- [ ] 会用 git status/add/commit/log
- [ ] 知道 git restore / reset / revert 区别
- [ ] 理解 package.json 与 npm scripts
- [ ] 能按报错速查表定位问题
- [ ] 闭卷自测 ≥ 8/10

---

### 0.4 建议学习时长

| 阶段 | 时间 |
|------|------|
| DevTools §3～§9、§22 | 3 小时 |
| Git §10～§12、§20、§29～§30 | 3 小时 |
| npm §14～§17、§21、§31 | 2 小时 |
| 练习 + 自测 | 2 小时 |

---

### 0.5 可验证成果

1. 用 F12 完成一次：改样式、看 Network、设断点。
2. 练习项目至少 3 次有意义 commit。
3. `npm init -y` 并读懂 package.json 各字段。

---

## 1. 为什么前端也离不开工程化

很多初学者在静态页面阶段，觉得前端就是：

- 写 HTML
- 写 CSS
- 写 JS

但一旦项目变大，你就会发现还需要：

- 文件组织
- 代码管理
- 依赖安装
- 调试排错

这些都属于工程化基础。

## 2. 文件组织基础建议

前端项目不要所有文件乱堆在一起。

至少可以有这些目录意识：

- `html`
- `css`
- `js`
- `images`

后面再复杂一点可能还会有：

- `components`
- `utils`
- `assets`

## 3. 浏览器开发者工具

你必须尽早熟悉它。

常见面板：

- Elements
- Styles
- Console
- Network
- Sources

## 4. Elements 面板

作用：

- 查看页面结构
- 实时改 HTML
- 查元素嵌套关系

## 5. Styles 面板

作用：

- 查看元素最终样式
- 看哪些 CSS 生效
- 看哪些样式被覆盖

这是排 CSS 问题的核心工具之一。

## 6. Console 面板

作用：

- 输出日志
- 看报错
- 临时运行 JavaScript

你要逐渐养成习惯：

- 有问题先看控制台

## 7. Network 面板

作用：

- 看接口请求
- 看资源加载
- 看状态码
- 看请求耗时

## 8. Sources 面板

作用：

- 打断点
- 单步调试 JavaScript
- 观察变量值

## 9. 调试的基础思路

当页面出问题时，不要只靠猜。

建议按这个顺序想：

1. HTML 结构对不对
2. CSS 有没有生效
3. JavaScript 有没有报错
4. DOM 是否取到了
5. 接口有没有请求成功

## 10. Git 是什么

> **系统学习 Git**（分支、远程、PR、团队协作）请学 [Git 学习系列](../Git/00-学习路线图与说明.md)（00～05）。  
> 本节只做**初识**：让你知道 Git 存在、能完成第一次 `commit`；详细内容在 Git 01～05。

Git 是版本控制工具。

它能帮你：

- 记录代码历史
- 回溯修改
- 协作开发

## 11. Git 常见操作

### 查看状态

```bash
git status
```

### 添加文件

```bash
git add .
```

### 提交

```bash
git commit -m "feat: add login page"
```

### 查看日志

```bash
git log
```

## 12. 分支基础认知

分支可以理解为：

- 一条独立开发线

常见操作：

- 创建分支
- 切换分支
- 合并分支

## 13. 为什么前端要会 Git

因为你以后做项目时几乎一定会：

- 改代码
- 提交代码
- 拉代码
- 合并代码

## 14. 包管理基础认知

现代前端项目经常会接触：

- npm
- pnpm
- yarn

它们的作用大致是：

- 安装依赖
- 管理依赖
- 运行脚本

## 15. `package.json` 基础认知

这是很多前端项目里的核心文件之一。

它通常记录：

- 项目名称
- 依赖
- 脚本命令

你现在不一定马上深入，但最好先认识它。

## 16. 依赖是什么

依赖就是项目运行需要的外部库。

比如：

- 工具库
- 构建工具
- 框架

## 17. 前端工程化会逐步接触什么

你以后会看到很多词：

- 打包
- 构建
- 模块化
- 热更新
- 压缩
- 代码规范

现在不需要全懂，但要知道这些是现代前端项目的一部分。

## 18. 代码规范基础认知

建议你从一开始就注意：

- 命名清晰
- 文件结构清晰
- 不要把所有逻辑写一个文件
- 注释写在关键处

## 19. 推荐项目目录结构

```text
my-frontend-practice/
├── index.html
├── pages/
│   ├── about.html
│   └── login.html
├── css/
│   ├── base.css      # 重置、变量
│   ├── layout.css
│   └── components.css
├── js/
│   ├── main.js
│   └── utils.js
├── images/
├── practice/         # 每章练习分子文件夹
│   ├── 03-css/
│   └── 08-dom/
├── .gitignore        # 忽略 node_modules 等
└── README.md         # 项目说明
```

原则：**HTML / CSS / JS / 图片分开**，练习按章节分子目录。

---

## 20. Git 完整工作流（初学者版）

```bash
# 1. 初始化（项目根目录）
git init

# 2. 配置用户信息（每台电脑做一次，用你的名字和邮箱）
git config user.name "你的名字"
git config user.email "your@email.com"

# 3. 日常循环
git status              # 看改了什么
git add .               # 暂存所有改动
git commit -m "feat: 完成待办列表 DOM 练习"

# 4. 查看历史
git log --oneline

# 5. 分支（可选）
git branch feature/todo
git checkout feature/todo   # 或 git switch feature/todo
# ... 开发 ...
git checkout main
git merge feature/todo
```

### 提交信息怎么写（简单规范）

```text
feat: 新功能
fix: 修 bug
docs: 文档
style: 格式（不影响逻辑）
refactor: 重构
```

例：`feat: 添加登录表单校验`、`fix: 修复导航栏移动端错位`

### `.gitignore` 示例

```gitignore
node_modules/
.DS_Store
*.log
.env
```

---

## 21. npm 快速上手

```bash
# 检查是否安装 Node（自带 npm）
node -v
npm -v

# 在项目目录初始化
npm init -y          # 生成 package.json

# 安装依赖（举例）
npm install lodash   # 生产依赖
npm install -D vite  # 开发依赖（-D）

# 运行 package.json 里 scripts
npm run dev
```

### `package.json` 片段解读

```json
{
  "name": "my-practice",
  "version": "1.0.0",
  "scripts": {
    "dev": "vite",
    "build": "vite build"
  },
  "dependencies": {
    "lodash": "^4.17.21"
  },
  "devDependencies": {
    "vite": "^5.0.0"
  }
}
```

- `dependencies`：项目运行需要
- `devDependencies`：仅开发时需要
- `scripts`：用 `npm run <名字>` 执行的命令

初学阶段**不必急着装构建工具**，先把原生 HTML/CSS/JS 练熟；学到框架前再接触 Vite 等。

---

## 22. DevTools 调试 JS 步骤

1. **Console**：看红色报错，点链接跳到出错行
2. **Sources** → 打开 `main.js` → 行号左侧点击设**断点**
3. 触发操作（如点按钮），代码在断点处暂停
4. 右侧看 **Scope** 里变量值，**Step over** 单步执行
5. 用 `console.log` / `console.table` 打印中间结果

### 常用 Console 命令

```js
console.log("普通");
console.warn("警告");
console.error("错误");
console.table([{a:1},{a:2}]);  // 表格形式
```

---

## 23. 按症状排查清单

| 症状 | 先查 |
|------|------|
| 页面空白 | Console 报错；HTML 路径 |
| 样式不对 | Elements → Styles；CSS 文件 404 |
| 点击无反应 | Console；是否 DOM 未就绪；选择器 |
| 接口失败 | Network 状态码；CORS；URL |
| 图片不显示 | Network 是否 404；src 路径 |
| 改了代码没变化 | 强刷 Ctrl+F5；缓存；是否改错文件 |

---

## 24. VS Code / Cursor 推荐扩展

| 扩展 | 作用 |
|------|------|
| Live Server | 保存自动刷新预览 |
| Prettier | 代码格式化 |
| ESLint | JS 语法与规范检查（以后用） |
| Chinese Language Pack | 中文界面（可选） |

**编码设置**：右下角确认 **UTF-8**，与 HTML `meta charset` 一致。

---

## 25. 分级练习

**基础**：用 F12 故意改 Elements 里一段文字和颜色  
**进阶**：练习项目 `git init` + 至少 3 次有意义 commit  
**挑战**：`npm init` 后安装一个包并在 Console 里 `import` 试用（需 module 环境）

---

## 27. 练习建议

1. 把现有练习按推荐目录整理一遍
2. 用 Git 管理，每天至少 commit 一次
3. 用 Sources 断点调试一个按钮点击函数
4. 用 Network + Console 配合排查一个故意写错的接口

---

## 29. Git 撤销操作

Git 撤销有三种常用的，分别应对不同场景：

### 29.1 `git restore` — 撤销工作区修改（未 add）

```bash
# 丢弃某个文件的所有改动（回到上次 commit 的状态）
git restore 文件名

# 丢弃所有文件的改动
git restore .
```

### 29.2 `git restore --staged` — 撤销暂存（已 add 但未 commit）

```bash
# 把文件从暂存区撤回工作区
git restore --staged 文件名
# 或使用旧命令（同样效果）
git reset HEAD 文件名
```

### 29.3 `git reset` — 回退 commit

```bash
# 回退到上一个 commit，保留代码修改在工作区
git reset --soft HEAD~1

# 回退到上一个 commit，保留代码修改但取消暂存（默认）
git reset HEAD~1

# 回退到上一个 commit，完全丢弃修改（危险！）
git reset --hard HEAD~1
```

### 29.4 `git revert` — 安全撤销（推荐）

```bash
# 创建一个新 commit 来撤销之前某次 commit
# 不改变历史，适合已推送的代码
git revert HEAD
```

### 撤销场景速查

| 场景 | 命令 |
|------|------|
| 改乱了还没 add | `git restore 文件名` |
| 已经 add 了想撤销 | `git restore --staged 文件名` |
| 想撤销最近一次 commit（保留代码） | `git reset --soft HEAD~1` |
| 想撤销已推送的 commit | `git revert HEAD` |
| 想回到某个历史版本看看 | `git checkout <commit-hash>` |

---

## 30. Git 远程仓库

```bash
# 关联远程仓库
git remote add origin https://github.com/用户名/仓库名.git

# 查看远程仓库
git remote -v

# 推送代码
git push -u origin main       # 首次推送，-u 设定上游分支
git push                      # 之后直接 push

# 拉取代码
git pull                      # = git fetch + git merge
git fetch                     # 只下载远程更新，不合并

# 克隆仓库
git clone https://github.com/用户名/仓库名.git
```

### 冲突解决示例

当 `git pull` 或 `git merge` 出现冲突时：

```bash
# 1. 打开冲突文件，找到冲突标记：
# <<<<<<< HEAD
# 你的代码
# =======
# 远程的代码
# >>>>>>> branch-name

# 2. 手动编辑，保留想要的代码，删除 <<<<<< ======= >>>>>>> 标记

# 3. 标记为已解决
git add 文件名
git commit -m "merge: 解决冲突"

# 4. 如果解决一半不想继续
git merge --abort   # 回到合并前的状态
```

---

## 31. npm scripts 与 npx

### npm scripts 常用配置

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "lint": "eslint src --fix",
    "format": "prettier --write src",
    "test": "vitest"
  }
}
```

运行：`npm run dev`（`start` 和 `test` 可省略 `run`）

### npx — 直接运行包，不安装

```bash
# 不需要全局安装就能运行
npx create-react-app my-app
npx eslint --init

# 原理：临时下载 → 执行 → 清理
```

---

## 32. ESLint + Prettier 快速配置

```bash
# 1. 安装
npm install -D eslint prettier

# 2. 初始化 ESLint 配置
npx eslint --init   # 交互式回答问题

# 3. .prettierrc 配置
{
  "semi": true,
  "singleQuote": false,
  "tabWidth": 2,
  "trailingComma": "es5",
  "printWidth": 100
}
```

初学阶段不一定要装，但要知道它们是干嘛的：
- **ESLint**：检查代码错误和风格（如未使用的变量）
- **Prettier**：自动格式化代码（如统一缩进、引号）

---

## 32.1 Chrome DevTools 四面板手把手

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | F12 → Elements | 左侧 DOM 树 | 检查是否在 iframe 内 |
| 2 | 选中 `<h1>` → Styles 改 color | 标题变色 | 是否被更高优先级覆盖 |
| 3 | Console 输入 `1+1` | 输出 2 | 是否在正确 frame |
| 4 | Network 刷新页 | 请求列表 | 是否 Preserve log 被关 |
| 5 | Sources 给 main.js 设断点 | 点击后暂停 | script 是否 module 延迟加载 |

---

## 32.2 git 第一次提交手把手

| 步骤 | 命令 | 预期 |
|------|------|------|
| 1 | `git init` | 出现 .git 文件夹 |
| 2 | 写 `.gitignore` 含 node_modules | 避免误提交 |
| 3 | `git add .` | status 显示 staged |
| 4 | `git commit -m "feat: init project"` | log 有一条记录 |
| 5 | `git log --oneline` | 看到短 hash |

---

## 33. 调试进阶（Sources 面板）

### 断点类型
- **普通断点**：点行号左侧，代码执行到这暂停
- **条件断点**：右键行号 → "Add conditional breakpoint" → 输入条件（如 `i === 5`）
- **DOM 断点**：Elements 面板右键元素 → "Break on" → subtree modifications

### 调试操作
| 按钮 | 作用 |
|------|------|
| ▶ Resume | 继续执行，直到下一个断点 |
| ⤵ Step over | 执行当前行，不进入函数内部 |
| ↓ Step into | 进入函数内部 |
| ↑ Step out | 执行完当前函数并跳出 |
| 👁 Watch | 添加监视表达式，实时观察值变化 |

### Call Stack（调用栈）
显示当前函数的调用路径：从入口到此处的链。点击可跳到对应函数。

---

## 34. 性能调试：Performance 面板与 Lighthouse

### Performance 面板
1. F12 → Performance
2. 点击录制按钮 → 操作页面 → 停止
3. 分析：
   - **Frames**：帧率（绿条 = 60fps，红色 = 卡顿）
   - **Main**：主线程 JS 执行时间
   - **Summary**：Scripting / Rendering / Painting 耗时分布

### Lighthouse
F12 → Lighthouse → Generate report。关注：
- Performance（性能得分）
- Accessibility（无障碍得分）
- Best Practices（最佳实践）

---

## 35. 常见报错信息与解决速查表

| 报错信息（关键词） | 可能原因 | 解决方向 |
|---------------------|----------|----------|
| `Cannot read properties of null` | DOM 未就绪或选择器错误 | 检查 `<script>` 位置、选择器拼写 |
| `xxx is not defined` | 变量/函数未声明或拼写错误 | 检查拼写、作用域、引入 |
| `xxx is not a function` | 变量不是函数（可能是 undefined） | 检查赋值、导入 |
| `Cannot read properties of undefined` | 访问了 undefined 的属性 | 可选链 `?.`、检查数据是否存在 |
| `Failed to fetch` | 网络不通、URL 错误、CORS | 检查 URL、Network 面板 |
| `Unexpected token` | 语法错误（少括号、逗号等） | 检查报错位置附近语法 |
| `Maximum call stack size exceeded` | 无限递归 | 检查递归终止条件 |
| `Cannot set properties of null` | 给 null 对象设属性 | 检查 DOM 选择是否成功 |
| `Assignment to constant variable` | 对 `const` 重新赋值了 | 改用 `let` 或不要重新赋值 |
| `Blocked by CORS policy` | 跨域请求被拦截 | 后端配 CORS，或开发代理 |

---

## 26. FAQ

**Q1：Git 和 GitHub 一样吗？**  
Git 是本地版本控制；GitHub 是托管代码的网站，用 `git push` 上传。

**Q2：npm 慢怎么办？**  
可配置国内镜像：`npm config set registry https://registry.npmmirror.com`

**Q3：要不要学 Webpack？**  
初学不必；先会用 Vite 或框架自带脚手架即可。

**Q4：DevTools 改了样式为什么刷新就没了？**  
Elements 里是临时预览；要永久生效需改源 CSS 文件。

**Q5：git add . 会提交 node_modules 吗？**  
若 `.gitignore` 写了 `node_modules/` 就不会；务必配置 gitignore。

**Q6：commit 写错了怎么办？**  
未 push：`git reset --soft HEAD~1`；已 push：用 `git revert`。

**Q7：dependencies 和 devDependencies 区别？**  
前者生产运行需要（如 axios）；后者仅开发需要（如 vite、eslint）。

**Q8：Console 红色报错从哪读起？**  
看最后一行报错信息 + 右侧文件名行号，点链接跳 Sources。

**Q9：Network 里 304 是什么？**  
缓存命中，资源未变；200 才是完整下载。

**Q10：要不要学 Docker？**  
Vue 10 章会涉及；本章知道「容器像打包好的运行环境」即可。

**Q11：ESLint 和 Prettier 冲突怎么办？**  
装 `eslint-config-prettier` 关 ESLint 里和格式有关的规则。

**Q12：Live Server 和 Vite dev 区别？**  
Live Server 只刷新静态页；Vite 还做模块编译、HMR、proxy。

---

## 37. 闭卷自测

1. DevTools 四大面板各干什么？
2. 调试 JS 的基本顺序（§9）是哪五步？
3. `git status / add / commit` 各做什么？
4. git restore 和 git reset 适用场景？
5. package.json 里 scripts、dependencies 是什么？
6. npm 和 npx 区别？
7. **动手**：给按钮点击函数设断点，单步执行一次。
8. **动手**：git init 并完成 3 次 commit。
9. **综合**：页面空白时按 §23 症状表排查。
10. **综合**：解释为什么前端项目需要 .gitignore。

### 37.1 自测参考答案

1. Elements 结构样式；Console 日志；Network 请求；Sources 断点。
2. HTML→CSS→JS 报错→DOM 是否取到→接口是否成功。
3. status 看改动；add 暂存；commit 提交快照。
4. restore 丢工作区/暂存改动；reset 回退 commit。
5. scripts 是 npm run 命令；dependencies 是运行依赖包。
6. npm 安装到 node_modules；npx 临时下载执行不全局装。
7. Sources 点行号 → 触发事件 → Step over。
8. 三次不同 feat/fix 信息即可。
9. 先 Console 报错，再 Network 404，再选择器。
10. 忽略 node_modules、.env 等巨大或敏感文件。

---

## 38. 费曼检验

3 分钟向朋友解释「Git 是干什么的」：

1. **问题**：改代码怕改坏、想回到昨天版本。
2. **Git**：每次 commit 像存档；可以对比、回退、分支并行开发。
3. **和 GitHub**：Git 本地存档；GitHub 是云备份和协作平台。

---

## 36. 学完标准（扩充）

- [ ] 能组织清晰的前端练习目录
- [ ] 熟练使用 DevTools 的 Elements、Console、Network、Sources 四大面板
- [ ] 会用 `git status/add/commit/log` 管理代码历史
- [ ] 知道 `git restore`（撤销修改）、`git reset`（回退 commit）、`git revert`（安全撤销）
- [ ] 了解远程仓库操作（push/pull/clone）
- [ ] 理解 `package.json`、dependencies/devDependencies/scripts 的作用
- [ ] 知道 npx、ESLint、Prettier 是干什么的
- [ ] 会设置断点调试、看调用栈和变量值
- [ ] 能按报错信息对照速查表定位问题
