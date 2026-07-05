# HTML 基础结构与常用标签

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0 读前导读（0.1～0.8）、手把手步骤表、骨架逐行读、FAQ≥12、闭卷自测、费曼检验；链 Vue 01 与 examples -->

## 0. 读前导读（零基础也能跟上）

> **读者假设**：会新建文件夹、保存文件、用浏览器打开网页。本章是 [00 路线图](./00-学习路线图与说明.md) 之后的第一份「写代码」文档，也是 [Vue 01](../Vue/01-Vue入门与环境搭建.md) 的 HTML 地基。

### 0.1 用一句话弄懂本章

**一句话**：HTML 是网页的**骨架**——用标签标出标题、段落、链接、图片和列表；浏览器按标签把内容「搭」成你看到的页面。

**生活类比——盖房子**：

| 概念 | 房子类比 | 代码里是什么 |
|------|----------|--------------|
| **HTML 文档** | 整栋楼的建筑图纸 | 一个 `.html` 文件 |
| **标签（tag）** | 房间类型：卧室、厨房、门 | `<h1>` `<p>` `<a>` |
| **head** | 水电煤表、房产证信息（住户不直接看） | `<meta>` `<title>` |
| **body** | 客厅、卧室（住户看得见） | 正文、导航、图片 |
| **CSS**（下几章学） | 装修、刷墙、家具 | 本章先不管外观 |
| **JavaScript**（06 章起） | 开关、智能门锁 | 本章先不管交互 |

**为什么重要**：后面 Vue 模板里写的 `<div>`、`<button>`、`<img>` 全是 HTML；[todo.md](../../todo.md) 项目 notehub 的登录表单、文章列表，底层结构都从这里来。

### 0.2 你需要提前知道什么（真不会就先跳到哪一章）

| 你现在的水平 | 建议动作 |
|--------------|----------|
| 不会新建/保存文件 | 先读 [00 章 §4](./00-学习路线图与说明.md) 工具环境 |
| 完全零基础 | **从 §0.7 手把手跟做**，再读 §1～§4 |
| 看过 HTML 但没系统学 | 重点 §0.8 逐行读 + §19 实体 + 章末自测 |
| 目标 Vue 路线 | 本章 + [02 表单语义](./02-HTML表单表格多媒体与语义化.md) 后再开 [Vue 01](../Vue/01-Vue入门与环境搭建.md) |

**最低门槛**：编辑器 UTF-8；Chrome/Edge 能打开本地 `.html`。

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

- [ ] 能默写 HTML5 骨架（DOCTYPE、html、head、body、charset、viewport）
- [ ] 能解释 `head` 与 `body` 各放什么
- [ ] 会使用 h1～h6、p、strong、em、ul/ol/li、a、img
- [ ] 图片必写 `alt`；链接理解 `href` 与 `target`
- [ ] 能区分 id（唯一）与 class（可复用）
- [ ] 知道块级与行内元素的基本嵌套规则
- [ ] 会写常用实体并在浏览器打开 [examples/01-html-entities-demo.html](./examples/01-html-entities-demo.html)
- [ ] 完成 §29 至少 2 个练习页面
- [ ] 闭卷自测 ≥ 8/10

### 0.4 建议学习时长与节奏

| 阶段 | 建议时间 | 做什么 |
|------|----------|--------|
| §0 导读 + 第一个页面 | 1 小时 | §0.7 手把手跟做 |
| §1～§10 标签基础 | 2～3 小时 | 每类标签写一个 mini 示例 |
| §11～§18 链接图片列表 | 2 小时 | 做一个带导航的页面 |
| §19 实体 + 示例文件 | 45 分钟 | 浏览器打开 entities demo |
| §20～§28 规范与语义 | 1.5 小时 | 对照错误清单改自己的页 |
| 练习 + 自测 | 2 小时 | §29 练习 + 闭卷自测 |

**节奏建议**：第一天必须能在浏览器里看到自己的第一个页面；不要连续三天只读不写。

### 0.5 学完本章你能做什么（可验证的具体动作）

1. **新建** `practice/01-html/index.html`，浏览器打开看到标题和段落。
2. **修改** `<title>`，确认标签页标题变化。
3. **添加** 一张本地图片 + `alt` 文字，图片能显示。
4. **写出** 含 3 个链接的 `nav` 列表（先用 `ul>li>a`）。
5. **打开** [examples/01-html-entities-demo.html](./examples/01-html-entities-demo.html)，F12 对照源码与显示。

### 0.6 核心术语三件套

**术语（HTML / HyperText Markup Language）**：超文本标记语言，用标签描述网页结构与内容，不是编程语言。
**生活类比**：建筑图纸上的房间标注——「这里是主卧」「这里是门」，不负责墙纸颜色。
**为什么重要**：所有前端技术的共同起点；SEO 与无障碍都依赖标签语义。
**本章用到的地方**：§2～§4 全文。

**术语（标签 tag / 元素 element）**：尖括号包裹的标记，多数成对出现，如 `<p>...</p>`。
**生活类比**：乐高积木的凸凹接口——有标准形状才能和别的块拼接（嵌套）。
**为什么重要**：写错或未闭合会导致布局乱、JS 选不到元素。
**本章用到的地方**：§6 标签特点；§22 常见错误。

**术语（语义化 semantic）**：用合适的标签表达内容含义，而非只用 `div` 凑样子。
**生活类比**：门牌写「卫生间」而不是「3 号房间」——一眼就懂用途。
**为什么重要**：搜索引擎、屏幕阅读器、团队协作都更轻松；Vue 组件命名也常对应语义区域。
**本章用到的地方**：§27 语义化；与 [02 章](./02-HTML表单表格多媒体与语义化.md) 深化。

### 0.7 手把手：第一个 HTML 页面在浏览器里跑起来

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 在 `practice/01-html/` 新建 `index.html` | 资源管理器里出现空文件 | 检查路径、文件夹是否建好 |
| 2 | 粘贴 §3 骨架代码（或 §0.8 表对照手敲） | 编辑器里有完整 HTML | 对照 §0.8 逐行检查拼写 |
| 3 | **另存为 UTF-8**（Cursor 右下角选 UTF-8） | 编码显示 UTF-8 | 中文变乱码 → 加 charset meta |
| 4 | 双击文件或用 Live Server 打开 | 浏览器显示 `Hello HTML` 和段落 | 白屏 → F12 Console 看报错 |
| 5 | 改 `<h1>` 文字，保存，刷新 | 标题文字立即变化 | 没变化 → 确认保存的是正确文件 |
| 6 | 删一行 `</p>` 故意破坏，保存刷新 | 可能仍显示但 F12 Elements 结构异常 | 补回闭合标签，见 §22 |
| 7 | 打开 [examples/01-html-entities-demo.html](./examples/01-html-entities-demo.html) | 实体符号正确显示 | 路径错 → 从本章 `examples/` 相对打开 |
| 8 | F12 → Elements，点选 `<h1>` | 右侧看到对应 HTML 节点 | 不会用 → [00 章 §4.2](./00-学习路线图与说明.md) |

### 0.8 骨架 HTML 逐行读（主示例）

§3 第一个页面的源码，建议**手敲**一遍。下表是「改错会怎样」速查：

| 行号/片段 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `<!DOCTYPE html>` | 声明 HTML5，触发标准模式 | 省略可能进入怪异模式，布局偶发异常 |
| `<html lang="zh-CN">` | 文档根元素；语言中文 | `lang` 省略影响读屏与 SEO，一般不致命 |
| `<head>` | 元数据容器，不直接展示 | 把 `<h1>` 写进 head → 页面结构错误 |
| `<meta charset="UTF-8" />` | 字符编码 UTF-8 | **删掉 → 中文乱码** |
| `<meta name="viewport" ...>` | 移动端视口适配 | 删掉 → 手机上网页字极小 |
| `<title>...</title>` | 浏览器标签页标题 | 省略 → 标签页显示文件名，不专业 |
| `</head>` | 结束 head | 忘记闭合 → 后面内容可能被当 head 解析 |
| `<body>` | 用户可见内容区 | 正文写 head 里 → 不规范，样式脚本易出问题 |
| `<h1>Hello HTML</h1>` | 一级标题（每页通常一个主标题） | 用 `div` 代替 → 语义与 SEO 变差 |
| `<p>...</p>` | 段落 | 用多个 `<br>` 代替 → 语义错误，间距难控 |
| `</body></html>` | 闭合 body 与文档 | 漏闭合 → 浏览器猜测解析，维护困难 |

**与后续章节衔接**：

- 学完本章 → [02 表单与语义](./02-HTML表单表格多媒体与语义化.md)
- 准备 Vue → [Vue 01 §0.2](../Vue/01-Vue入门与环境搭建.md) 要求本系列 01～05 + JS 06
- 实体演示 → **[examples/01-html-entities-demo.html](./examples/01-html-entities-demo.html)**（务必浏览器打开）

### 0.9 与 [Vue 01](../Vue/01-Vue入门与环境搭建.md) 的对照（先建立期待）

学完本章后，你在 Vue 里会见到「同一套骨架，多了一层自动化」：

| 本章 HTML 写法 | Vue 01 里类似概念 | 区别 |
|----------------|-------------------|------|
| `<h1>{{ 标题文字 }}</h1>` 手写死 | `<h1>{{ shopName }}</h1>` | Vue 里数据变，标题自动变 |
| `<button onclick="...">` 行内事件 | `<button @click="count++">` | Vue 用声明式绑定事件 |
| 多个 `.html` 文件跳转 | 单页 `App.vue` + 路由（后续章） | 结构仍是 HTML 标签 |
| `class="card"` 配合 CSS | `<style scoped>` 里写 `.card` | 标签名不变，样式作用域不同 |

**关键句**：Vue 没有取代 HTML；[Vue 01 §0.2](../Vue/01-Vue入门与环境搭建.md) 明确要求 **01～05 + JS 06** 再入门。本章过关标志：能**不看资料**默写 §3 骨架并在浏览器打开。

### 0.10 本章学习路径一张图

```text
§0 导读 + 第一个页面（浏览器能打开）
  ↓
§1～§10 文本/标题/列表标签
  ↓
§11～§18 链接、图片、路径、id/class
  ↓
§19 实体 + examples/01-html-entities-demo.html
  ↓
§20～§28 规范、语义化、错误清单
  ↓
§29 练习 → §30 学完标准 → §32 闭卷自测 ≥8/10
```

---

## 1. 这一份文档学什么

这一份是前端学习的真正起点。

学完这一份，你应该能做到：

- 看懂一个最基础的网页结构
- 自己写出简单页面
- 理解常见 HTML 标签的作用
- 知道标签不是“随便写”，而是有语义和结构的

## 2. HTML 是什么

HTML 全称是：

- HyperText Markup Language

中文一般叫：

- 超文本标记语言

它不是编程语言，而是标记语言。

你可以先这样理解：

- HTML 负责网页的内容和结构
- CSS 负责网页的样式和布局
- JavaScript 负责网页的交互和逻辑

如果把网页比作房子：

- HTML 是房子的骨架
- CSS 是装修和摆设
- JavaScript 是电器、开关、联动功能

### 为什么初学前端一定先学 HTML

因为你后面无论写：

- 页面
- 表单
- 登录框
- 商品卡片
- 列表
- 导航栏

都必须先有结构。

很多新手一上来就想学“好看的页面”，结果会发现：

- 样式加不上去
- 结构乱
- 不知道元素之间是什么关系

本质上往往是 HTML 不扎实。

### 学 HTML 时最重要的思维

不是背标签名字，而是学会问：

1. 这部分内容是什么
2. 它在页面里扮演什么角色
3. 应该用什么标签表达它

例如：

- 一段正文，用 `p`
- 一个主标题，用 `h1`
- 一组步骤，用 `ol`
- 一个超链接，用 `a`

这就叫“结构和语义意识”。

## 3. 第一个 HTML 页面

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>我的第一个网页</title>
  </head>
  <body>
    <h1>Hello HTML</h1>
    <p>这是我的第一个网页。</p>
  </body>
</html>
```

## 4. 这段代码逐行解释

### `<!DOCTYPE html>`

表示这是一个 HTML5 文档。

作用：

- 告诉浏览器按 HTML5 标准模式解析页面

### `<html>`

表示整个 HTML 文档的根元素。

### `lang="zh-CN"`

表示页面主要语言是中文。

这对：

- 浏览器
- 搜索引擎
- 屏幕阅读器

都有帮助。

### `<head>`

放页面的配置信息，不是页面主体展示内容。

常见内容：

- 编码
- 标题
- 样式文件
- SEO 信息

### `<body>`

放用户真正看到的页面内容。

### 一个页面为什么一定有这些结构

你可以把 HTML 页面分成两部分：

#### 浏览器和搜索引擎更关心的配置部分

也就是：

- `head`

#### 用户真正看到的展示部分

也就是：

- `body`

如果你把本该放在 `head` 里的信息乱丢到 `body`，或者反过来，就会让页面结构变得不规范。

## 5. 常见 head 标签

### 5.1 `meta charset`

```html
<meta charset="UTF-8" />
```

作用：

- 指定字符编码为 UTF-8

为什么重要：

- 防止中文乱码

### 5.2 `meta viewport`

```html
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
```

作用：

- 让页面在移动端正常缩放和适配

### 5.3 `title`

```html
<title>首页</title>
```

作用：

- 设置浏览器标签页标题

### 5.4 以后你还会在 head 里见到什么

虽然现在还不需要全部展开，但你最好先认识这些常见东西：

- 样式文件 `link`
- 脚本文件 `script`
- 页面描述 `meta name="description"`
- 图标 `link rel="icon"`

也就是说，`head` 不是“没用”，而是“用户不直接看到，但页面运行非常依赖”的部分。

## 6. HTML 标签的基本特点

### 6.1 大多数标签成对出现

```html
<p>这是一段话</p>
```

前面是开始标签，后面是结束标签。

### 6.2 有些标签是单标签

例如：

```html
<br />
<img />
<input />
<meta />
```

### 6.3 标签可以嵌套

```html
<div>
  <p>段落</p>
</div>
```

### 6.4 不要乱嵌套

例如：

- 块级元素和行内元素有结构习惯
- 某些标签不能随便互相包裹

### 为什么标签嵌套要小心

因为浏览器虽然“很宽容”，很多写错的 HTML 它也会帮你猜着解析，但这并不代表：

- 结构就是对的
- 样式就一定稳定
- JavaScript 操作时不会出问题

所以从一开始就要尽量写规范。

## 7. 注释

```html
<!-- 这是 HTML 注释 -->
```

作用：

- 给开发者看
- 浏览器不会显示

## 8. 标题标签

HTML 提供了 `h1` 到 `h6` 六级标题。

```html
<h1>一级标题</h1>
<h2>二级标题</h2>
<h3>三级标题</h3>
```

### 使用建议

- 一个页面通常核心主标题用 `h1`
- 不要为了字体大就乱用标题标签

标题标签不仅影响样式，也影响：

- 语义结构
- SEO

### 为什么不能用标题标签纯粹为了变大

因为标题标签本质上不是“字号按钮”，而是“文档结构层级”。

例如一篇文章：

- 文章主标题：`h1`
- 一级章节：`h2`
- 二级章节：`h3`

如果你只是因为想让文字更大就随便塞一个 `h1`，页面结构就会混乱。

你应该逐步形成这种意识：

- 大小是 CSS 的事
- 结构是 HTML 的事

## 9. 段落与文本标签

### 9.1 `p`

表示一段文字。

```html
<p>这是一个段落。</p>
```

### 9.2 `br`

强制换行。

```html
第一行<br />第二行
```

### 9.3 `hr`

水平分隔线。

```html
<hr />
```

### `p` 和 `br` 不要混着乱用

这是初学者很常见的问题。

很多人会写：

```html
第一段<br /><br />
第二段
```

更合理的写法往往是：

```html
<p>第一段</p>
<p>第二段</p>
```

为什么：

- `p` 表示段落
- `br` 表示行内强制换行

它们语义完全不同。

## 10. 文本强调标签

### 10.1 `strong`

表示强调，语义比单纯加粗更强。

```html
<strong>重要内容</strong>
```

### 10.2 `em`

表示强调语气，默认通常表现为斜体。

```html
<em>重点提醒</em>
```

### 10.3 `span`

通用行内容器，本身没有特别强语义。

```html
<span>一小段文本</span>
```

常用于：

- 配合 CSS 给局部文字加样式

### `strong` 和 `b`、`em` 和 `i` 的区别怎么理解

你以后可能会见到：

- `b`
- `i`

你可以先建立一个简单认知：

- `strong`、`em` 更偏语义
- `b`、`i` 更偏视觉表现

现代开发通常更推荐优先考虑语义标签。

## 11. 列表

### 11.1 无序列表 `ul`

```html
<ul>
  <li>苹果</li>
  <li>香蕉</li>
  <li>橙子</li>
</ul>
```

### 11.2 有序列表 `ol`

```html
<ol>
  <li>注册</li>
  <li>登录</li>
  <li>进入首页</li>
</ol>
```

### 11.3 列表使用场景

- 导航项
- 功能清单
- 步骤说明

### 为什么列表在网页中这么常见

因为网页里很多内容本质上就是“一组同类项”：

- 导航菜单
- 商品列表
- 评论列表
- 步骤流程

所以你越早习惯使用列表标签，后面写页面会越顺。

## 12. 链接 `a`

```html
<a href="https://www.example.com">访问网站</a>
```

### 常见属性

#### `href`

指定跳转地址。

#### `target="_blank"`

在新窗口或新标签页打开。

```html
<a href="https://www.example.com" target="_blank">新标签打开</a>
```

### 可以跳到什么地方

- 外部网址
- 本地页面
- 页面内部某个位置

### 链接为什么重要

网页之所以叫“超文本”，一个核心原因就是：

- 页面和页面之间可以跳转

这也是互联网最基础的连接方式之一。

### 一个更完整的链接示例

```html
<a href="https://developer.mozilla.org" target="_blank" title="打开 MDN 文档">
  学习前端文档
</a>
```

这里你可以看到：

- `href`：去哪
- `target`：怎么打开
- `title`：鼠标提示

## 13. 图片 `img`

```html
<img src="cat.jpg" alt="一只猫" />
```

### 常见属性

#### `src`

图片地址。

#### `alt`

替代文本。

为什么重要：

- 图片加载失败时显示
- 对无障碍和 SEO 有帮助

### 为什么图片必须写 alt

很多新手会偷懒不写 `alt`，但这是不好的习惯。

`alt` 的意义包括：

- 图片坏掉时仍能知道这里本来是什么
- 屏幕阅读器可以读出这张图的含义
- 搜索引擎更容易理解页面内容

### 一个更完整的图片示例

```html
<img src="./images/product.jpg" alt="白色机械键盘商品图" title="机械键盘" />
```

## 14. 路径基础认知

### 14.1 相对路径

```html
<img src="./images/a.png" />
```

### 14.2 绝对路径

```html
<img src="https://example.com/a.png" />
```

初学时你要重点搞清楚相对路径。

### 相对路径到底相对谁

这是初学者非常容易迷糊的点。

相对路径通常是：

- 相对于当前 HTML 文件所在位置

例如：

如果你的文件结构是：

```text
project/
  index.html
  images/
    a.png
```

那么在 `index.html` 中写：

```html
<img src="./images/a.png" />
```

就是正确的。

如果你路径总写不对，很多时候不是标签问题，而是目录结构没想清楚。

## 15. 容器标签：`div` 和 `span`

### `div`

常见块级容器。

```html
<div>这是一个区域</div>
```

### `span`

常见行内容器。

```html
<span>这是一小段文字</span>
```

它们的核心作用不是“有默认样式”，而是：

- 帮你组织结构
- 方便加样式

### 为什么前端初学者特别容易滥用 div

因为 `div` 很通用，感觉“哪里都能放”。

但如果整个页面除了 `div` 还是 `div`，就会出现：

- 结构可读性差
- 后面 CSS class 混乱
- 语义不清晰

所以你要逐渐从“只会用 div”走向“知道什么时候该用更合适的标签”。

## 16. 块级元素和行内元素基础认知

你现在先这样理解：

### 块级元素

常见特点：

- 默认独占一行

例如：

- `div`
- `p`
- `h1`
- `ul`

### 行内元素

常见特点：

- 默认一行内排列

例如：

- `span`
- `a`
- `strong`
- `em`

这不是绝对规则大全，但对初学足够重要。

### 一个简单直观的理解方式

块级元素更像：

- 一整块区域

行内元素更像：

- 一小段文字中的局部内容

这会帮助你在写页面时更自然地判断该用哪类标签。

## 17. HTML 属性

标签除了名字，还可以有属性。

```html
<input type="text" placeholder="请输入用户名" />
```

这里：

- `type`
- `placeholder`

都是属性。

属性的作用是：

- 给标签增加附加信息
- 控制行为或样式钩子

### 属性和值的关系

你可以把属性理解成“标签的附加说明”。

比如：

```html
<a href="https://example.com">链接</a>
```

这里：

- `href` 是属性名
- `https://example.com` 是属性值

以后你会发现 HTML 很多能力都靠属性控制。

## 18. 常见全局属性

### `id`

页面中通常应唯一。

```html
<div id="header"></div>
```

### `class`

可以重复，用于分组和样式控制。

```html
<div class="card"></div>
```

### `title`

鼠标悬停提示文字。

```html
<p title="提示信息">文字</p>
```

### `id` 和 `class` 怎么区分

这是必须讲清的基础点。

#### `id`

- 更强调唯一身份
- 一个页面中通常不要重复

#### `class`

- 更强调分类
- 多个元素可以共享同一个 class

你后面写 CSS 和 JavaScript 时会大量依赖 `class`，所以别一上来就什么都写成 `id`。

## 19. HTML 实体字符

有些字符如果直接写在 HTML 里，会被浏览器当成标签或语法的一部分，所以不能直接写，要改成**实体（entity）**。

实体的一般格式：**和号 + 英文名字 + 英文分号**（三部分紧挨着，中间不要空格）。

### 常见实体（用中文记，避免笔记预览乱码）

| 浏览器里看到 | 实体英文名 | 在 HTML 源码里怎么敲（顺序） |
|--------------|------------|------------------------------|
| 小于 | lt | & → l → t → ; |
| 大于 | gt | & → g → t → ; |
| 空格 | nbsp | & → n → b → s → p → ; |
| 和号 | amp | & → a → m → p → ; |

记忆口诀：**先在源码里写实体，浏览器才会显示对应的符号。**

### 动手查看真实效果（推荐）

本仓库准备了演示页，**请用浏览器打开**（不要只在 Markdown 预览里看）：

**文件路径**：`examples/01-html-entities-demo.html`

操作步骤：

1. 在 Cursor 左侧文件树找到 `examples/01-html-entities-demo.html`
2. 右键 → **Reveal in File Explorer**，双击用 Chrome / Edge 打开  
   或安装 Live Server 后选 **Open with Live Server**
3. 页面上看显示效果，按 **F12 → Elements** 查看每段对应的 HTML 源码

### 示例在源码里长什么样

下面用「拆字」描述，请对照演示页里的灰色源码区核对：

- **示例 1**：段落标签内写：数字 5、空格、**lt 实体**、空格、数字 10  
- **示例 2**：段落标签内写：字母 A、空格、**amp 实体**、空格、字母 B  
- **示例 3**：段落标签内写：文字「词与词之间」、**两个 nbsp 实体**、文字「留空」

### 为什么本节不用符号直接写

在 **Cursor 的 Markdown 预览**里，和号、尖括号容易被当成 HTML 解析，导致显示错乱。  
学实体时以 **`examples/01-html-entities-demo.html`** 为准；写自己的 HTML 文件时按上表顺序输入即可。

## 20. 页面内部锚点

```html
<a href="#contact">跳到联系方式</a>

<h2 id="contact">联系方式</h2>
```

适合：

- 长页面目录导航

### 锚点跳转的常见场景

- 文档目录
- 页面顶部“回到某一节”
- 单页说明页面内部导航

## 21. HTML 书写规范建议

建议你从一开始就养成这些习惯：

- 标签层级缩进统一
- 属性尽量规范书写
- 结构清晰
- 不要为了样式乱用标签

### 再补几条很重要的书写习惯

- 类名尽量见名知意
- 结构尽量一层层清楚缩进
- 一个区域最好有清晰包裹关系
- 页面先搭结构，再加样式

## 22. 初学者常见错误

### 22.1 忘记写 `DOCTYPE`

可能导致浏览器解析模式异常。

### 22.2 中文乱码

通常是忘记写：

```html
<meta charset="UTF-8" />
```

### 22.3 标签不闭合

会导致结构混乱。

### 22.4 只会写 `div`

这是非常常见的问题。

你后面要逐步学会用更有语义的标签。

### 22.5 把 HTML 当作“页面最终效果”来写

这也是常见误区。

你要慢慢建立正确分工：

- HTML：表达内容和结构
- CSS：控制样式和布局
- JavaScript：控制交互和逻辑

如果结构和样式混在脑子里，后面会很乱。

## 23. 一个更完整的示例页面

下面给你一个更像真实网页骨架的 HTML 示例：

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>个人介绍页面</title>
  </head>
  <body>
    <h1>前端学习者小明</h1>

    <p>大家好，我正在学习 HTML、CSS 和 JavaScript。</p>

    <h2>我的学习内容</h2>
    <ul>
      <li>HTML 页面结构</li>
      <li>CSS 页面样式</li>
      <li>JavaScript 页面交互</li>
    </ul>

    <h2>推荐网站</h2>
    <p>
      我常去
      <a href="https://developer.mozilla.org" target="_blank">MDN</a>
      学前端知识。
    </p>

    <img src="./images/avatar.png" alt="学习者头像" />
  </body>
</html>
```

### 你应该从这个示例中看懂什么

1. 页面是有层级结构的
2. 标题、段落、列表、链接、图片各有职责
3. 同样是“显示内容”，不同内容应该用不同标签

## 24. HTML 实体完整参考表

除了前面讲过的常用实体，以下为完整速查：

### 24.1 必备实体（前端必须记住）

> 与 [19 节](#19-html-实体字符) 相同：预览里不直接写尖括号、和号等符号。下表用**中文 + 实体英文名**速记；代码块里是完整写法对照。浏览器真实效果见 **`examples/01-html-entities-demo.html`**。

| 浏览器里看到 | 实体英文名 | 源码敲法（顺序） | 用途 |
|--------------|------------|------------------|------|
| 小于号 | lt | & → l → t → ; | 最常见 |
| 大于号 | gt | & → g → t → ; | |
| 和号 | amp | & → a → m → p → ; | URL 参数中很常见 |
| 双引号 | quot | & → q → u → o → t → ; | 属性值冲突时用 |
| 单引号 | apos | & → a → p → o → s → ; | |
| 不换行空格 | nbsp | & → n → b → s → p → ; | 防止单词间换行 |

完整写法对照（代码块内为字面量，可直接抄写）：

```text
显示         实体写法
<            &lt;
>            &gt;
&            &amp;
"            &quot;
'            &apos;
(空格)       &nbsp;
```

### 24.2 常用符号实体

| 符号名 | 实体英文名 | 源码敲法（顺序） |
|--------|------------|------------------|
| 版权符号 | copy | & → c → o → p → y → ; |
| 注册商标 | reg | & → r → e → g → ; |
| 商标 | trade | & → t → r → a → d → e → ; |
| 乘号 | times | & → t → i → m → e → s → ; |
| 除号 | divide | & → d → i → v → i → d → e → ; |
| 正负号 | plusmn | & → p → l → u → s → m → n → ; |
| 左箭头 | larr | & → l → a → r → r → ; |
| 右箭头 | rarr | & → r → a → r → r → ; |
| 省略号 | hellip | & → h → e → l → l → i → p → ; |
| 长破折号 | mdash | & → m → d → a → s → h → ; |
| 心形 | hearts | & → h → e → a → r → t → s → ; |

完整写法对照：

```text
显示    实体写法
©       &copy;
®       &reg;
™       &trade;
×       &times;
÷       &divide;
±       &plusmn;
←       &larr;
→       &rarr;
…       &hellip;
—       &mdash;
♥       &hearts;
```

```html
<!-- ✅ 显示代码中的 HTML 标签 -->
<p>在 HTML 中，段落用 <code>&lt;p&gt;</code> 标签表示。</p>

<!-- ✅ 版权信息 -->
<footer>&copy; 2026 我的网站 | 用 ❤ 构建</footer>
```

---

## 25. HTML 文档结构与嵌套规则

### 25.1 块级 vs 行内标签速查

| 块级（block） | 行内（inline） | 行内块（默认 inline-block） |
|:--|:--|:--|
| `div`, `p`, `h1-h6` | `span`, `a`, `strong`, `em` | `img`, `input`, `button` |
| `ul`, `ol`, `li` | `code`, `label`, `i`, `b` | `select`, `textarea` |
| `section`, `article`, `nav` | `small`, `sub`, `sup` | |
| `header`, `footer`, `main` | `br`, `abbr` | |
| `table`, `form`, `hr` | | |

### 25.2 嵌套规则（重要！经常踩坑）

```html
<!-- ✅ 块级可以包含行内 -->
<div><span>行内文字</span></div>

<!-- ✅ 块级可以包含块级 -->
<section><p>段落</p></section>

<!-- ❌ 行内不能包含块级 -->
<!-- <span><div>错误！</div></span> -->

<!-- ❌ p 里面不能放块级元素（浏览器会自动关闭 p）-->
<!-- <p><div>这样 p 会被截断</div></p> -->

<!-- ⚠️ 特殊规则：a 在 HTML5 中可包裹块级（但别滥用）-->
<a href=”#” style=”display:block;”>
  <h2>整块卡片都可点击</h2>
  <p>卡片描述</p>
</a>
```

### 25.3 哪些标签不能放什么

| 标签 | 不能包含 | 原因 |
|------|---------|------|
| `p` | 块级元素（div/ul/table） | 浏览器会自动关闭 p |
| `a` | 另一个 `a` | 链接不可嵌套 |
| `button` | 另一个 `button` 或 `a` | 按钮不可嵌套 |
| `ul/ol` | 非 `li` 的直接子元素 | 列表项只能是 li |
| `form` | 另一个 `form` | 表单不可嵌套 |

---

## 26. 标签使用场景与反例

| 标签 | ✅ 正确使用 | ❌ 常见误用 |
|------|-----------|-----------|
| `h1-h6` | 文档标题层级 | 为字体大小而用（应该用CSS） |
| `p` | 正文段落 | 用 `<br>` 模拟段落间距 |
| `ul/li` | 导航菜单、功能列表 | 用 div 模拟列表 |
| `ol/li` | 步骤、排名 | 用 ul 代替（丢失顺序语义） |
| `a` | 跳转、锚点、下载 | 用 span + onclick 代替 |
| `img` | 内容图片（产品图、插图） | 不写 alt、用CSS背景代替内容图 |
| `strong` | 重要强调内容 | 仅为加粗外观 |
| `button` | 触发交互 | 用 div/span 模拟按钮 |
| `table` | 规则数据展示 | 整页布局 |
| `div` | 无更合适标签时的容器 | 替代所有语义标签 |

---

## 27. 完整实战：一篇文章页面

```html
<!DOCTYPE html>
<html lang=”zh-CN”>
<head>
  <meta charset=”UTF-8” />
  <meta name=”viewport” content=”width=device-width, initial-scale=1.0” />
  <title>理解 HTML 语义化 — 前端学习笔记</title>
  <meta name=”description” content=”一篇深入浅出的 HTML 语义化文章” />
  <style>
    body {
      font-family: system-ui, -apple-system, sans-serif;
      max-width: 720px; margin: 0 auto; padding: 24px 16px;
      line-height: 1.8; color: #1e293b; background: #fafafa;
    }
    article h1 { font-size: 2rem; border-bottom: 2px solid #2563eb; padding-bottom: 12px; }
    article h2 { font-size: 1.4rem; margin-top: 32px; color: #1e40af; }
    article p { margin: 1em 0; }
    article code {
      background: #e2e8f0; padding: 2px 6px; border-radius: 4px;
      font-family: monospace; font-size: 0.9em;
    }
    pre {
      background: #1e293b; color: #e2e8f0; padding: 16px;
      border-radius: 8px; overflow-x: auto; font-size: 0.9em;
    }
    blockquote {
      border-left: 4px solid #2563eb; padding: 8px 16px; margin: 16px 0;
      background: #eff6ff; color: #1e40af;
    }
    footer {
      margin-top: 48px; padding-top: 16px; border-top: 1px solid #e2e8f0;
      color: #64748b; font-size: 0.9em; text-align: center;
    }
  </style>
</head>
<body>
  <article>
    <header>
      <h1>理解 HTML 语义化：不只是 div 和 span</h1>
      <p>
        <time datetime=”2026-06-18”>2026 年 6 月 18 日</time>
        · <span>作者：前端学习者</span>
        · <span>标签：<strong>HTML</strong> <strong>语义化</strong></span>
      </p>
    </header>

    <h2>1. 什么是语义化</h2>
    <p>
      语义化（Semantic HTML）是指<strong>用合适的 HTML 标签表达内容的含义</strong>，
      而非仅仅关注视觉呈现。浏览器、搜索引擎、屏幕阅读器依赖标签来理解页面。
    </p>
    <blockquote>
      <p>💡 核心理念：结构归 HTML，样式归 CSS，行为归 JavaScript。</p>
    </blockquote>

    <h2>2. 为什么重要</h2>
    <ul>
      <li><strong>SEO</strong>：搜索引擎更懂你的页面在说什么</li>
      <li><strong>无障碍</strong>：屏幕阅读器能正确朗读页面结构</li>
      <li><strong>可维护性</strong>：`&lt;nav&gt;` 比 `&lt;div class=”nav”&gt;` 更直观</li>
    </ul>

    <h2>3. 对比：语义化 vs 非语义化</h2>
    <h3>❌ div 泛滥（难以理解结构）</h3>
    <pre><code>&lt;div class=”header”&gt;
  &lt;div class=”title”&gt;标题&lt;/div&gt;
&lt;/div&gt;
&lt;div class=”nav”&gt;
  &lt;div class=”link”&gt;首页&lt;/div&gt;
&lt;/div&gt;
&lt;div class=”content”&gt;正文&lt;/div&gt;
&lt;div class=”footer”&gt;版权&lt;/div&gt;</code></pre>

    <h3>✅ 语义化（一目了然）</h3>
    <pre><code>&lt;header&gt;&lt;h1&gt;标题&lt;/h1&gt;&lt;/header&gt;
&lt;nav&gt;&lt;a href=”/”&gt;首页&lt;/a&gt;&lt;/nav&gt;
&lt;main&gt;&lt;p&gt;正文&lt;/p&gt;&lt;/main&gt;
&lt;footer&gt;版权&lt;/footer&gt;</code></pre>

    <h2>4. 总结</h2>
    <p>
      语义化不是教条。有时候 <code>&lt;div&gt;</code> 也没问题（确实没有更合适的标签时）。
      但养成<strong>优先考虑语义标签</strong>的习惯，会让你的 HTML 更专业、更健壮。
    </p>
  </article>

  <footer>
    <p>&copy; 2026 前端学习笔记</p>
  </footer>
</body>
</html>
```

---

## 28. 初学者常见错误排查清单

| 症状 | 可能原因 | 先查什么 |
|------|----------|----------|
| 中文乱码 | 编码问题 | HTML 里有没有 `<meta charset=”UTF-8” />`；编辑器右下角是否 UTF-8 |
| 图片不显示 | 路径错误 | 路径相对于当前 HTML 的位置；Network 面板看 404 |
| 链接点不动 | `href` 为空/拼错 | 检查 `href` 属性值 |
| 页面结构错乱 | 标签未闭合或嵌套错误 | 用 W3C Validator 检查 |
| 移动端字小如蚁 | 缺少 viewport meta | 加上 `<meta name=”viewport” content=”width=device-width, initial-scale=1.0” />` |
| 按钮点击不了 | 用 `div` 代替 `button` | 改用 `<button>` 或加 `role=”button”` + `tabindex=”0”` |
| 样式不生效 | class/id 写错/大小写 | Elements 面板确认属性名 |

---

## 29. 练习建议（深度版）

建议你自己动手写：

1. 一个个人介绍页面（含头像图、技能列表、社交链接）
2. 一个文章页面（含标题层级、时间、引用、代码块、图片）
3. 一个商品列表静态页面（3-5 个卡片，含图片、名称、价格、购买链接）
4. 一个带标题、段落、图片、链接、列表的完整页面

### 练习进阶方法

1. 先照着文档完整示例写一遍
2. 再自己删掉重写（不看原文）
3. 再把页面内容换成你自己的（真实的自我介绍）
4. 再尝试加更多结构（导航栏、页脚、侧边栏）
5. 用 W3C Validator（https://validator.w3.org/）检查 HTML 是否有错误

### 练习自检三层

- **能写**：独立写出完整页面骨架
- **能用对**：导航用 `nav>ul>li>a`、图片必写 `alt`、不用 `<br>` 模拟段落
- **能解释**：能说明为什么用 `h1` 而不是 `div style=“font-size:32px”`

---

## 30. 学完标准

如果你能做到这些，这一份就掌握得扎实了：

### 基础
- [ ] 能默写完整 HTML5 页面骨架（含 DOCTYPE、charset、viewport）
- [ ] 知道 `head` 放配置、`body` 放内容
- [ ] 会使用 h1-h6、p、br、hr、strong、em、span
- [ ] 会用 ul/ol/li 做导航和列表
- [ ] 会用 a（含 target、锚点）和 img（含 alt、路径）
- [ ] 能区分 id（唯一）和 class（可复用）
- [ ] 知道块级和行内的基本区别及嵌套规则
- [ ] 能写出常用 HTML 实体（`&lt;` `&gt;` `&amp;` `&nbsp;` `&copy;`）

### 进阶
- [ ] 能独立写一个语义化完整的文章页面
- [ ] 一眼看出 div 泛滥的问题，知道何时用更合适的标签
- [ ] 路径写不错（理解相对路径的参照）
- [ ] 有 SEO 和无障碍意识（alt、语义标签、合理标题层级）

---

## 31. 常见问题 FAQ

### Q1：HTML 是编程语言吗？

不是。它是**标记语言**，描述结构和内容；逻辑在 JavaScript，样式在 CSS。

### Q2：标签名大小写敏感吗？

HTML5 不敏感（`<P>` 也能解析），但**行业惯例全小写**，请统一小写。

### Q3：必须用 VS Code / Cursor 吗？

任意纯文本编辑器均可；推荐 VS Code/Cursor 是为了 UTF-8 提示、 Emmet 和 Live Server。

### Q4：双击打开和 Live Server 有什么区别？

双击用 `file://` 协议，够学本章；Live Server 用本地 HTTP，**保存自动刷新**，学 CSS/JS 时更方便。

### Q5：为什么我的中文变成乱码？

99% 是编码问题：检查 `<meta charset="UTF-8" />`、文件 UTF-8 保存、编辑器右下角编码。

### Q6：`h1` 可以用多个吗？

技术上可以，但**一个页面通常只有一个主 `h1`**（站点 logo 或文章标题）；小节用 `h2`～`h6` 递进。

### Q7：`br` 和 `p` 怎么选？

**分段用 `p`**；`br` 只在诗歌、地址等需要行内换行时用，不要用多个 `br` 撑间距（间距是 CSS 的事）。

### Q8：图片路径 `./images/a.png` 和 `images/a.png` 一样吗？

多数情况下一样（`./` 表示当前目录）；建议统一风格，并确认图片文件真实存在。

### Q9：`id` 和 `class` 初学者怎么选？

**默认用 `class`**；只有页面唯一标识（如锚点 `id="contact"`）时用 `id`。一个页面不要重复 `id`。

### Q10：`a` 标签的 `href="#"` 可以吗？

占位可以，但点击会跳顶部；开发中用 `javascript:void(0)` 或按钮更合适（交互章再讲）。

### Q11：实体字符要在 Markdown 里直接写 `<` 吗？

笔记里容易乱；**在 `.html` 里用实体**，浏览器打开 [examples/01-html-entities-demo.html](./examples/01-html-entities-demo.html) 对照学习。

### Q12：学完本章能直接学 Vue 吗？

建议再完成 **02 章 + CSS 03～05 + JS 06～07**，再开 [Vue 01](../Vue/01-Vue入门与环境搭建.md)。Vue 模板仍是 HTML，但还会用到样式和基础 JS。

### Q13：和 [todo.md](../../todo.md) 暑假项目什么关系？

notehub 的登录页、文章列表、导航栏，底层都是本章标签 + 02 表单；第 1 周可跟 todo 在 `practice/` 建个人介绍页练手。

### Q14：如何检查 HTML 写得是否规范？

用 [W3C Validator](https://validator.w3.org/) 粘贴代码或提交 URL；初学阶段先保证标签闭合、语义合理。

---

## 32. 闭卷自测

合上书，尝试回答（6 概念 + 2 动手 + 2 综合）。

1. **概念** 用「房子」类比说明 HTML、CSS、JS 各做什么？
2. **概念** `head` 和 `body` 分别放什么？各举一个例子。
3. **概念** 块级元素和行内元素在嵌套上有什么习惯？
4. **概念** 为什么 `img` 要写 `alt`？
5. **概念** `id` 与 `class` 的核心区别是什么？
6. **概念** 什么是语义化？举一个 `nav` 优于 `div class="nav"` 的理由。
7. **动手** 默写 HTML5 最小骨架（含 DOCTYPE、charset、viewport、title）。
8. **动手** 写一段 HTML：一个 `h1`、两个 `p`、一个含 3 项的 `ul` 导航（不看上文）。
9. **综合** 说明从保存 `.html` 到浏览器显示标题，经历了哪几个关键标签？
10. **综合** 你要做 notehub 文章列表静态页：会用本章哪些标签？实体 demo 文件路径是什么？

### 自测参考答案

1. HTML 骨架；CSS 装修；JS 电器与开关。
2. head：元数据，如 charset、title；body：可见内容，如 h1、p、img。
3. 块级常独占一行，可包行内；行内不要包块级（如 `p` 里不要再套 `div`）。
4. 图片加载失败时替代文字；屏幕阅读器朗读；SEO。
5. id 页面内应唯一；class 可多个元素共享。
6. 标签表达含义；`nav` 让搜索引擎/读屏/同事一眼知是导航区。
7. （略，对照 §3 或 §0.8 表自检）
8. （略，结构正确即可，导航项为 `li>a` 更佳）
9. DOCTYPE → html → head（charset/title）→ body → h1 等 → 浏览器解析渲染。
10. `header`/`nav`/`main`/`article`/`h1`～`h3`/`p`/`a`/`img` 等；路径 `./examples/01-html-entities-demo.html`。

---

## 33. 费曼检验

请在不看资料的情况下，用 **3 分钟**向没学过编程的朋友解释：「HTML 是什么、一个最简网页由哪些部分组成」。

### 口述提纲（对照用）

1. **HTML 是什么**：像房子的钢筋骨架，标出标题、段落、链接、图片的位置；浏览器按标签把页面搭出来。
2. **最简结构**：开头告诉浏览器这是 HTML5；`head` 放编码和标题（用户不直接看）；`body` 放真正看见的文字和图片。
3. **为什么要认真学**：以后 Vue/React 只是在骨架上加自动化；`alt`、标题层级、语义标签写对了，页面才专业、好维护、对 Google 和读屏友好。

---

*本章已按 EXPANSION-STANDARD 扩充（§0 导读 0.1～0.8、手把手 §0.7、逐行读 §0.8、FAQ §31、闭卷 §32、费曼 §33）。*

**EXPANSION-STANDARD 自检**：☑ §0 读前导读 ☑ 房子类比 ☑ 步骤表 ☑ 逐行读 ☑ FAQ ≥12 ☑ 闭卷 10 题 ☑ 费曼 ☑ Vue 01 / examples 链接
