# CSS 基础语法、选择器与文本样式

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0 导读（CSS=装修）、Elements DevTools、主示例逐行读、闭卷自测、费曼检验 -->

## 0. 读前导读（零基础也能跟上）

> **读者假设**：你已能写 HTML 页面骨架（[01](./01-HTML基础结构与常用标签.md)、[02 表单表格](./02-HTML表单表格多媒体与语义化.md)）。本章教「怎么让页面好看、好读、好维护」——**结构是 HTML，样式是 CSS**。

### 0.1 用一句话弄懂本章

**一句话**：CSS 通过「选择器选中元素 + 属性改外观」控制颜色、字体、间距和边框；学会选择器和优先级，后面 Flex/Grid 才不会写一堆不生效的样式。

**生活类比——CSS = 装修**：

| CSS 概念 | 装修类比 | 本章位置 |
|----------|----------|----------|
| **HTML 结构** | 毛坯房：墙、门、窗的位置 | 上一章 |
| **选择器** | 指定「刷哪面墙」而不是全屋 | §5～§16 |
| **class `.card`** | 给同类家具贴标签，批量换风格 | §7 |
| **颜色/字体** | 墙漆、字号、灯具色温 | §17～§21 |
| **背景/边框/圆角** | 壁纸、踢脚线、圆角柜 | §22～§24 |
| **优先级** | 设计师改稿 vs 业主临时要求谁说了算 | §28 |
| **外部 CSS 文件** | 装修图纸单独装订，多套房复用 | §3 |
| **CSS 变量** | 全屋主色卡，改一处全房跟着变 | §30 |

**为什么重要**：Vue/React 组件的 `<style>` 本质仍是 CSS；[02 章表单](./02-HTML表单表格多媒体与语义化.md) 写好后，本章才能做出像样的登录页和导航栏。

---

### 0.2 你需要提前知道什么

| 水平 | 建议 |
|------|------|
| 只会 HTML | 从 §2～§4 三种引入方式和语法开始，每个 demo 改颜色观察效果 |
| 样式总不生效 | 重点 §28 优先级 + §33.5 排查清单 + 本章 Elements 八步 |
| 已会一点 CSS | 补 §14～§16 伪类/伪元素/属性选择器，§18 单位 |
| 目标 Vue | 本章 + [04 Flex/Grid](./05-CSS布局FlexGrid响应式与动画.md) 后再开 Vue 组件样式 |

---

### 0.3 本章知识地图（☐→☑）

- [ ] 会用外部 CSS（推荐）、内部、行内三种方式，知道何时用哪种
- [ ] 掌握元素/class/id/后代/子代/并集/交集选择器
- [ ] 会用 `:hover`、`:focus`、`::before`、`::after`、属性选择器
- [ ] 理解 px/rem/em/%/vw/vh 并会选用
- [ ] 会设颜色、字体、行高、背景、边框、圆角、阴影
- [ ] 知道优先级与继承，**不轻易**用 `!important`
- [ ] 会用 `:root` CSS 变量
- [ ] 能独立完成个人卡片 + 导航栏样式
- [ ] 会用 Elements 面板排查「样式不生效」
- [ ] 闭卷自测 ≥ 8/10

---

### 0.4 建议学习时长

| 阶段 | 时间 |
|------|------|
| CSS 是什么 + 引入 + 语法 §2～§4 | 1.5 小时 |
| 选择器全家桶 §5～§16 | 3 小时 |
| 颜色/单位/文本 §17～§21 | 2 小时 |
| 背景/边框/阴影 §22～§26 | 1.5 小时 |
| 优先级/继承/变量 §28～§30 | 1.5 小时 |
| 实战卡片+导航 + DevTools + 自测 | 2 小时 |

---

### 0.5 可验证成果

1. 给 [02 章登录表单](./02-HTML表单表格多媒体与语义化.md) 写外部 CSS：输入框聚焦蓝色边框、按钮圆角。
2. Elements 面板选中 `.profile-card`，改 `--primary` 变量，卡片按钮色跟着变。
3. 故意写两条冲突规则，用更具体的选择器修复（不用 `!important`）。
4. Tab 键走过导航栏，`:focus-visible` 轮廓清晰可见。

---

### 0.6 核心术语三件套

**术语（选择器 Selector）**：CSS 里用来「选中 HTML 元素」的模式，决定样式作用在谁身上。
**生活类比**：装修时的「仅客厅墙面」——不是全屋刷漆。
**为什么重要**：选错选择器会导致「改 A 却动了 B」或样式完全不生效。
**本章用到的地方**：§5～§16。

**术语（层叠 Cascading）**：多条规则冲突时，浏览器按优先级和顺序决定谁生效。
**生活类比**：公司规定 vs 部门规定 vs 临时便签——不是「后写的永远赢」。
**为什么重要**：排查样式问题必看 Styles 面板里被划掉的规则。
**本章用到的地方**：§28。

**术语（CSS 变量 Custom Properties）**：以 `--name` 定义、用 `var(--name)` 引用，可运行时切换主题。
**生活类比**：全屋「主色卡」——换一张色卡，按钮、链接、强调线一起变。
**为什么重要**：暗色模式、设计系统都靠变量；Vue 里也可在 `:root` 或组件根上定义。
**本章用到的地方**：§30；§31～§32 实战。

---

### 0.7 手把手：给 02 章登录表单写第一版 CSS

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 复制 [02 章 §31](./02-HTML表单表格多媒体与语义化.md) 登录 HTML 为 `login.html` | 无样式但结构完整 | 先结构后样式 |
| 2 | 新建 `login.css`，HTML `<head>` 加 `<link rel="stylesheet" href="login.css" />` | Network 里 css 200 | 路径错则 404 |
| 3 | 写 `body { font-family: system-ui, sans-serif; max-width: 400px; margin: 2rem auto; }` | 表单居中、字体变 UI 风 | margin auto 需块级宽度 |
| 4 | 写 `label { display: block; margin-top: 1rem; font-weight: 600; }` | 标签独占一行 | inline label 与 input 可能挤 |
| 5 | 写 `input { width: 100%; padding: 8px; border: 1px solid #cbd5e1; border-radius: 6px; }` | 输入框通栏圆角 | 缺 box-sizing 时可能溢出 |
| 6 | 写 `input:focus { outline: 2px solid #2563eb; border-color: #2563eb; }` | Tab/点击时蓝框 | 仅 :hover 对键盘用户不够 |
| 7 | 写 `button[type="submit"] { margin-top: 1rem; padding: 10px 20px; background: #2563eb; color: white; border: none; border-radius: 6px; cursor: pointer; }` | 蓝色提交按钮 | button 默认 type 是 submit |
| 8 | F12 Elements 改 `--primary` 或 background 实时看效果 | 学会在 Styles 里调试 | 见 §44 八步实操 |

---

## 1. 这一份文档学什么

这一份是 CSS 的起点。

学完后你应该能做到：

- 给 HTML 页面加基础样式
- 理解 CSS 选择器（包括属性选择器、伪类、伪元素）
- 控制文字、颜色、背景、边框等常见外观
- 理解 CSS 单位、优先级、继承与 CSS 变量
- 独立完成个人卡片和导航栏样式
- 初步建立「结构是 HTML，样式是 CSS」的思维

### 为什么这一份很重要

很多初学者 HTML 写得还行，但页面「丑」或「乱」，问题往往出在 CSS 基础不牢：

- 不知道选中了哪些元素（选择器）
- 不知道样式为什么没生效（优先级）
- 只会用 `px`，响应式一塌糊涂（单位）
- 把所有样式写在行内或 `<style>` 里，改一处要翻半天（引入方式）

把这一份吃透，后面学 Flex、Grid、动画都会顺很多。

### 学习建议

1. **边读边写**：每个示例都复制到本地，改几个值看效果
2. **用开发者工具**：右键 → 检查，看样式是否生效、被谁覆盖
3. **不要背属性表**：先掌握常用 20 个属性，其余用到再查
4. **每天练 30 分钟**：比周末突击 5 小时更有效

### 本章思维导图式小结

```
CSS 入门
├── 是什么：层叠样式表，管外观
├── 怎么用：外部 CSS（推荐）> 内部 > 行内
├── 语法：选择器 { 属性: 值; }
├── 选择器：元素 / class / id / 组合 / 伪类 / 伪元素 / 属性
├── 单位：px / rem / em / % / vw/vh
├── 文本与视觉：颜色、字体、背景、边框、阴影
├── 规则：优先级、继承、!important、CSS 变量
└── 实战：个人卡片 + 导航栏
```

---

## 2. CSS 是什么

CSS 全称是：

- Cascading Style Sheets

中文一般叫：

- 层叠样式表

它的主要作用是：

- 设置页面外观
- 调整排版
- 控制布局
- 添加动效

### 和 HTML、JavaScript 的关系

| 技术 | 职责 | 类比 |
|------|------|------|
| HTML | 内容与结构 | 房子的钢筋和墙体 |
| CSS | 外观与排版 | 油漆、壁纸、家具摆放 |
| JavaScript | 交互与逻辑 | 开关、门锁、智能家居 |

### 为什么叫「层叠」（Cascading）

同一条 CSS 规则可能同时作用在一个元素上，浏览器要决定「谁说了算」。这个过程叫层叠，取决于：

1. 来源（浏览器默认、用户、作者）
2. 优先级（选择器权重、`!important`）
3. 顺序（后写的同权重规则覆盖先写的）

你现在不必背完整算法，但要建立印象：**CSS 不是「后写就一定赢」**。

### 完整可运行示例：三者分工

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <title>HTML/CSS/JS 分工</title>
  <style>
  /* CSS：管好看 */
  .btn {
    padding: 8px 16px;
    background: #2563eb;
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
  }
  .btn:hover {
    background: #1d4ed8;
  }
  </style>
</head>
<body>
  <!-- HTML：管结构 -->
  <button class="btn" id="counter">点击次数：0</button>

  <script>
  // JavaScript：管交互
  let count = 0;
  document.getElementById('counter').addEventListener('click', () => {
    count++;
    document.getElementById('counter').textContent = '点击次数：' + count;
  });
  </script>
</body>
</html>
```

### 常见错误

- 以为 CSS 能改 HTML 结构（不能，只能改外观；改结构要靠 HTML 或 JS）
- 在还没写 HTML 结构时就纠结「用什么颜色」（先结构，后样式）

### 调试技巧

打开开发者工具（F12）→ Elements 面板，选中元素，右侧 Styles 可实时改 CSS 看效果。

### 思维导图式小结

```
CSS 是什么
├── 管外观，不管结构和逻辑
├── 层叠：多条规则冲突时的裁决机制
└── 与 HTML、JS 分工明确，不要混用职责
```

---

## 3. CSS 的三种使用方式

### 3.1 行内样式

```html
<p style="color: red;">红色文字</p>
```

特点：

- 写得快，改起来痛苦
- 无法复用（每个元素都要写一遍）
- 优先级很高，容易覆盖外部样式，造成维护噩梦
- 不利于缓存（HTML 体积变大）

一般不推荐大规模使用。

**什么时候可以用**：临时调试、邮件 HTML（部分场景）、第三方组件强制内联时。

### 3.2 内部样式

```html
<head>
  <style>
    p {
      color: blue;
    }
  </style>
</head>
```

通常写在 `<head>` 中。

适合：单页 demo、学习示例、极小的静态页。

### 3.3 外部样式表

```html
<link rel="stylesheet" href="./style.css" />
```

这是最推荐的方式。

优点：

- 结构和样式分离
- 更利于维护
- 可复用（多个页面共用一个 `style.css`）
- 浏览器可缓存 CSS 文件，加快二次访问

### 完整可运行示例：三种方式对比

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <title>CSS 引入方式</title>
  <!-- 方式3：外部（推荐） -->
  <link rel="stylesheet" href="demo.css" />
  <!-- 方式2：内部 -->
  <style>
    .internal { color: green; }
  </style>
</head>
<body>
  <p class="internal">内部样式：绿色</p>
  <p class="external">外部样式：需在 demo.css 里写 .external { color: purple; }</p>
  <!-- 方式1：行内 -->
  <p style="color: orange;">行内样式：橙色</p>
</body>
</html>
```

`demo.css` 内容：

```css
.external {
  color: purple;
}
```

### 为什么外部 CSS 最重要

实际项目常有几十上百个页面、上万行样式。全写在 HTML 里，任何人都不敢改。外部文件可以：

- 按模块拆分：`base.css`、`layout.css`、`components.css`
- 用构建工具压缩合并
- 团队协作时减少冲突

### 常见错误

| 错误 | 后果 |
|------|------|
| `href` 路径写错 | 样式完全不加载 |
| 忘记 `<link rel="stylesheet">` 只写 `<link href="...">` | 浏览器不认 |
| CSS 写在 `<body>` 底部也能生效，但会阻塞渲染体验 | 应放 `<head>` |
| 行内 + 外部混用同一属性 | 行内往往赢，难以预测 |

### 调试技巧

Network 面板刷新页面，看 `style.css` 是否 200。若是 404，就是路径问题。

### 思维导图式小结

```
CSS 引入
├── 行内 style=""     → 仅调试/特例
├── 内部 <style>      → 小 demo
└── 外部 link.css     → 项目标准 ✅
```

---

## 4. CSS 的基本语法

```css
p {
  color: red;
  font-size: 16px;
}
```

结构可以拆成：

- **选择器**：`p`（选中谁）
- **声明块**：`{ ... }`
- **属性**：`color`（改什么）
- **值**：`red`（改成什么样）
- **声明**：`color: red;`（注意分号）

### 语法细节（初学者常忽略）

```css
/* 注释用这种方式，单行或多行均可 */

.selector {
  property: value;   /* 每条声明必须以分号结尾（最后一行也建议写） */
  another: 10px 20px;  /* 多个值用空格分隔 */
}
```

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  h1 {
    color: #1e293b;
    font-size: 28px;
    margin-bottom: 12px;
  }
  p {
    color: #475569;
    line-height: 1.6;
  }
  </style>
</head>
<body>
  <h1>标题</h1>
  <p>正文段落。</p>
</body>
</html>
```

### 为什么重要

语法错误会导致**整段或整个文件**的后续规则失效。例如少了一个 `}`，后面所有样式都可能废掉。

### 常见错误

- 写成 `color = red`（用了等号，应是冒号）
- 忘记分号（有时下一行会「背锅」）
- 中文冒号 `：` 代替英文 `:` 
- 属性名拼错：`font-size` 写成 `fontsize`

### 调试技巧

开发者工具 Styles 里若整片规则变灰并提示 invalid property，检查拼写和符号。

### 思维导图式小结

```
CSS 语法
├── 选择器 { 属性: 值; }
├── 注释 /* */
└── 符号必须是英文 : 和 ;
```

---

## 5. 选择器是什么

选择器就是：

- 选中页面中的某些元素，然后给它们加样式

如果你不会选择器，就等于不会真正写 CSS。

### 为什么重要

HTML 里可能有几百个 `<p>`，你只想改文章区的段落颜色——靠选择器精确命中，而不是给每个标签加行内样式。

### 选择器大家族预览

| 类型 | 示例 | 作用 |
|------|------|------|
| 元素 | `p` | 所有 p |
| class | `.card` | 带 class="card" |
| id | `#header` | id="header" |
| 属性 | `[type="email"]` | 带特定属性 |
| 伪类 | `a:hover` | 特定状态 |
| 伪元素 | `p::before` | 元素前后虚拟节点 |

后面章节会逐一展开。

### 思维导图式小结

```
选择器 = 定位器
├── 选中目标元素
├── 决定样式作用范围
└── 影响优先级权重
```

---

## 6. 元素选择器

```css
p {
  color: red;
}
```

作用：

- 选中所有 `p` 标签

### 为什么重要

适合设置「全局默认」：所有段落字号、所有链接去掉下划线（再配合更具体的选择器覆盖）。

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  p { color: #334155; margin: 8px 0; }
  h2 { color: #0f172a; border-bottom: 2px solid #e2e8f0; }
  </style>
</head>
<body>
  <h2>小节标题</h2>
  <p>第一段。</p>
  <p>第二段。</p>
</body>
</html>
```

### 常见错误

- 过度使用元素选择器，导致「一改全改」，后面又要用更高优先级覆盖
- 用 `div` 选择器给所有 div 加样式（项目里 div 太多，太宽泛）

### 调试技巧

Elements 面板选中元素，看 Matched CSS Rules 里哪条 `p { }` 生效。

### 思维导图式小结

```
元素选择器 tag
├── 权重低
├── 适合全局默认
└── 慎用过于宽泛的 div、span
```

---

## 7. class 选择器

```html
<div class="card">卡片</div>
```

```css
.card {
  border: 1px solid #ccc;
}
```

这是实际开发里最常用的选择器之一。

### 为什么 class 是主力

- 可重复使用（多个元素同一 class）
- 可组合（`class="card featured"`）
- 语义灵活，不绑定唯一 id
- 权重适中，便于覆盖

### 多个 class

```html
<button class="btn btn-primary">提交</button>
```

```css
.btn { padding: 8px 16px; border-radius: 4px; }
.btn-primary { background: blue; color: white; }
```

### 命名建议（初学阶段）

- 用英文小写 + 连字符：`user-card`、`nav-link`
- 见名知意，避免 `a1`、`box2`
- BEM 等方法后面再学，先保证可读

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  .card {
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 16px;
    max-width: 320px;
  }
  .card-title {
    font-size: 18px;
    font-weight: bold;
    margin-bottom: 8px;
  }
  </style>
</head>
<body>
  <div class="card">
    <div class="card-title">产品名称</div>
    <p>产品简介文字。</p>
  </div>
</body>
</html>
```

### 常见错误

- HTML 写 `class="card"`，CSS 写 `.Card`（class 区分大小写）
- 忘记点号：写成 `card { }` 会被当成未知元素标签

### 思维导图式小结

```
class 选择器 .name
├── 可复用 ✅
├── 开发最常用 ✅
└── 记得加点：.card 不是 card
```

---

## 8. id 选择器

```html
<div id="header">头部</div>
```

```css
#header {
  background: black;
}
```

注意：

- 同一页面中，`id` 理论上应唯一（一个 id 只对应一个元素）
- 权重高于 class

### 为什么现在较少用 id 写样式

- 唯一性导致无法复用
- 权重太高，后期很难覆盖
- JS 用 `getElementById` 时 id 仍然有用，但样式优先 class

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  #page-header {
    background: #0f172a;
    color: white;
    padding: 16px;
  }
  </style>
</head>
<body>
  <header id="page-header">网站头部（仅此一处）</header>
</body>
</html>
```

### 常见错误

- 多个元素用同一个 id（违规且 JS 行为不确定）
- 为了「提高优先级」滥用 id（应该用更合理的选择器结构）

### 思维导图式小结

```
id 选择器 #name
├── 页面内应唯一
├── 权重高，慎用做样式
└── JS 锚点常用，样式优先 class
```

---

## 9. 通配符选择器

```css
* {
  margin: 0;
  padding: 0;
}
```

作用：

- 选中所有元素

常见用途：

- 做基础样式重置（配合更具体的规则）

### 为什么有人要 margin/padding 归零

浏览器对 `h1`、`p`、`ul` 等有**默认样式**（user agent stylesheet）。不同浏览器默认值略有差异，归零后自己统一设。

现代项目也常用 `normalize.css` 或 `reset.css`，不必死记 `* { margin:0 }` 是否最佳，但要理解动机。

### 性能提示

`*` 会匹配每一个节点，极端情况下略影响性能。一般小项目无感；大项目用更精确 reset。

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  * { box-sizing: border-box; }
  body { margin: 0; font-family: sans-serif; }
  </style>
</head>
<body>
  <h1>标题默认有 margin，body margin:0 只去 body 外边距</h1>
</body>
</html>
```

### 思维导图式小结

```
* 通配符
├── 全选，权重最低
├── 常用于 reset、box-sizing
└── 不要滥用做具体样式
```

---

## 10. 后代选择器

```css
.article p {
  color: gray;
}
```

表示：

- 选中 `.article` **内部任意层级**的所有 `p`（儿子、孙子、曾孙……）

### 为什么重要

页面结构是嵌套的。后代选择器让你只影响「某一区块里的段落」，而不影响页脚、侧栏的段落。

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  .article p { color: #475569; line-height: 1.7; }
  .sidebar p { color: #94a3b8; font-size: 14px; }
  </style>
</head>
<body>
  <article class="article">
    <p>文章段落（灰色）。</p>
    <div><p>文章内嵌套的段落也是灰色。</p></div>
  </article>
  <aside class="sidebar">
    <p>侧栏段落（浅灰小字）。</p>
  </aside>
</body>
</html>
```

### 常见错误

- 以为 `.article p` 只选直接子级（那是子代选择器 `>`）
- 选择器写太长：`.page .content .article .body p`（难维护，应简化 HTML 或 class）

### 思维导图式小结

```
后代 A B（空格）
├── 任意深度的 B
└── 范围比子代更宽
```

---

## 11. 子代选择器

```css
.menu > li {
  color: blue;
}
```

表示：

- 只选中 `.menu` 的**直接子元素** `li`

### 与后代选择器对比

```html
<ul class="menu">
  <li>一级 <ul><li>二级</li></ul></li>
</ul>
```

- `.menu li`：一级和二级 `li` 都选中
- `.menu > li`：只选中一级 `li`

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  .menu > li {
    border-bottom: 1px solid #e2e8f0;
    padding: 8px;
  }
  .menu li li {
    border-bottom: none;
    font-size: 14px;
  }
  </style>
</head>
<body>
  <ul class="menu">
    <li>首页</li>
    <li>产品
      <ul>
        <li>子产品 A</li>
      </ul>
    </li>
  </ul>
</body>
</html>
```

### 思维导图式小结

```
子代 A > B
├── 仅直接子元素
└── 多级菜单样式时常用
```

---

## 12. 并集选择器

```css
h1,
h2,
h3 {
  color: navy;
}
```

作用：

- 多个选择器共用同一套样式（逗号分隔）

### 完整可运行示例

```css
h1, h2, h3 {
  color: #1e293b;
  line-height: 1.3;
}
```

### 常见错误

- 逗号写错：`.a .b`（后代）和 `.a, .b`（并集）完全不同
- 最后一项多了逗号：某些旧浏览器可能出问题

### 思维导图式小结

```
并集 A, B, C
├── 共用样式，减少重复
└── 逗号 = 或的关系（满足任一即可）
```

---

## 13. 交集选择器

```css
button.primary {
  background: blue;
}
```

表示：

- 既是 `button` 元素，又带有 `class="primary"`（或 class 列表中包含 primary）

```html
<button class="primary">是</button>
<div class="primary">不是 button，不匹配</div>
```

### 与多个 class 的区别

```css
button.primary { }  /* button 且 .primary */
.btn.primary { }     /* 同时有 btn 和 primary 两个 class */
```

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  button { padding: 8px 16px; }
  button.primary { background: #2563eb; color: white; border: none; }
  button.ghost { background: transparent; border: 1px solid #ccc; }
  </style>
</head>
<body>
  <button class="primary">主要按钮</button>
  <button class="ghost">次要按钮</button>
</body>
</html>
```

### 思维导图式小结

```
交集 A.B（无空格连着写）
├── 同时满足
└── 比单独 .B 更精确
```

---

## 14. 伪类选择器基础

伪类用单冒号 `:`，表示元素**某种状态或位置**，而不是新标签。

### `:hover`

鼠标指针悬停时。

```css
a:hover {
  color: red;
}
```

### `:active`

激活瞬间（如鼠标按下未松开）。

### `:focus`

获得焦点时（键盘 Tab 或点击输入框）。

```css
input:focus {
  outline: 2px solid #2563eb;
  border-color: #2563eb;
}
```

### `:first-child`

作为父元素第一个子元素时。

### `:last-child`

作为父元素最后一个子元素时。

### `:nth-child(n)`

第 n 个子元素。常用：

- `nth-child(2)`：第 2 个
- `nth-child(odd)`：奇数
- `nth-child(even)`：偶数
- `nth-child(3n)`：每 3 个一组的第 1 个

### 完整可运行示例：表格斑马纹

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  table { border-collapse: collapse; width: 100%; }
  th, td { border: 1px solid #e2e8f0; padding: 8px; }
  tr:nth-child(even) { background: #f8fafc; }
  tr:hover { background: #e0f2fe; }
  a { color: #2563eb; text-decoration: none; }
  a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <table>
    <tr><th>姓名</th><th>分数</th></tr>
    <tr><td>小明</td><td>90</td></tr>
    <tr><td>小红</td><td>85</td></tr>
    <tr><td>小刚</td><td>92</td></tr>
  </table>
</body>
</html>
```

### 为什么重要

伪类让页面「可交互」的视觉反馈不需要 JavaScript：悬停变色、焦点高亮、隔行变色。

### 常见错误

- `:hover` 写在移动端为主的项目里并非所有设备都有「悬停」
- `nth-child` 数错：子元素包含文本节点外的标签，序号按所有子元素算

### 调试技巧

`:hover` 在开发者工具可强制状态：Styles → `:hov` → 勾选 `:hover`

### 思维导图式小结

```
伪类 :
├── 状态：hover / active / focus
├── 结构：first-child / nth-child
└── 不增加 DOM 节点
```

---

## 15. 属性选择器

根据元素的 HTML **属性**选中，表单、链接、国际化场景很常见。

### 常用形式

| 选择器 | 含义 |
|--------|------|
| `[type]` | 有 type 属性 |
| `[type="text"]` | type 等于 text |
| `[href^="https"]` | href 以 https 开头 |
| `[href$=".pdf"]` | href 以 .pdf 结尾 |
| `[class*="btn"]` | class 包含 btn |
| `[data-id="42"]` | 自定义 data 属性 |

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  input[type="text"] {
    border: 1px solid #cbd5e1;
    padding: 8px;
  }
  input[type="email"] {
    border: 1px solid #2563eb;
  }
  a[href^="https"]::after {
    content: " ↗";
    font-size: 12px;
  }
  a[download] {
    color: #059669;
  }
  </style>
</head>
<body>
  <input type="text" placeholder="普通输入" />
  <input type="email" placeholder="邮箱" />
  <a href="https://example.com">外链</a>
  <a href="./file.pdf" download>下载 PDF</a>
</body>
</html>
```

### 为什么重要

不用给每个输入框加额外 class，也能按 `type` 区分样式；外链标识、禁用状态 `[disabled]` 都可纯 CSS 处理。

### 常见错误

- 属性值区分大小写（HTML 属性一般小写）
- 忘记引号：`[type=text]` 在含特殊字符时可能失效，应写 `[type="text"]`

### 思维导图式小结

```
属性选择器 [attr]
├── = 完全匹配
├── ^= 开头  $= 结尾  *= 包含
└── 表单、链接场景利器
```

---

## 16. 伪元素 ::before 与 ::after

伪元素用双冒号 `::`（单冒号 `:` 旧写法也常兼容），在元素**内部最前或最后**生成一个虚拟子节点，必须有 `content` 属性（可为空字符串）。

### 基本用法

```css
.required::after {
  content: " *";
  color: red;
}
.tag::before {
  content: "【";
}
.tag::after {
  content: "】";
}
```

### 为什么重要

- 加装饰性内容而不改 HTML（图标、引号、必填星号）
- 配合 `content` 做简单图标
- 做清除浮动、遮罩等（进阶）

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  .quote {
    position: relative;
    padding: 16px 24px;
    background: #f1f5f9;
    border-radius: 8px;
  }
  .quote::before {
    content: "\201C";
    font-size: 48px;
    color: #94a3b8;
    position: absolute;
    left: 8px;
    top: -8px;
  }
  .link-arrow::after {
    content: " →";
  }
  .icon-dot::before {
    content: "";
    display: inline-block;
    width: 8px;
    height: 8px;
    background: #22c55e;
    border-radius: 50%;
    margin-right: 6px;
  }
  </style>
</head>
<body>
  <p class="quote">这是一段引用文字。</p>
  <a class="link-arrow" href="#">了解更多</a>
  <p class="icon-dot">在线状态</p>
</body>
</html>
```

### ::before / ::after 注意点

1. 默认是 `inline`，需要宽高时要改 `display: block` 或 `inline-block`
2. `content` 不能为空（不写 `content` 伪元素不显示）
3. 无法选中伪元素里的文字做复制（装饰性内容）
4. 不是真实 DOM，JS `querySelector` 默认选不到

### 常见错误

- 用伪元素写关键内容（SEO、无障碍会丢信息）
- 忘记 `content: ""` 想做纯色块装饰

### 思维导图式小结

```
伪元素 ::
├── ::before 内容前
├── ::after  内容后
├── 必须有 content
└── 装饰用，关键信息放 HTML
```

---

## 17. 颜色表示方式

### 17.1 颜色英文名

```css
color: red;
```

约 140 个英文名，开发中较少用（除 `transparent`、`white` 等）。

### 17.2 十六进制

```css
color: #ff0000;
color: #f00;        /* 简写 */
color: #ff000080;   /* 带透明度 8 位 */
```

### 17.3 rgb

```css
color: rgb(255, 0, 0);
```

### 17.4 rgba

```css
color: rgba(255, 0, 0, 0.5);
```

最后一个值表示透明度（0 全透，1 不透）。

### 17.5 现代：rgb 带斜杠透明度

```css
color: rgb(255 0 0 / 50%);
background: hsl(220 90% 56% / 0.1);
```

### 17.6 hsl / hsla

```css
color: hsl(0, 100%, 50%);
/* 色相 饱和度 亮度 */
```

调亮色暗色比 rgb 直观。

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  .hex { color: #2563eb; }
  .rgb { color: rgb(37, 99, 235); }
  .rgba-bg {
    background: rgba(37, 99, 235, 0.15);
    padding: 12px;
  }
  .hsl { color: hsl(220 90% 45%); }
  </style>
</head>
<body>
  <p class="hex">十六进制蓝</p>
  <p class="rgb">RGB 蓝</p>
  <p class="rgba-bg">半透明蓝底</p>
  <p class="hsl">HSL 蓝</p>
</body>
</html>
```

### 为什么重要

颜色贯穿全文样式；半透明 `rgba` 做遮罩、阴影特别常见。

### 常见错误

- `#fff` 写成 `#ffff` 
- `rgba` 第四个参数超过 1
- 用颜色名 `darkgray` 在不同显示器上不一致

### 思维导图式小结

```
颜色
├── 名 / #hex / rgb(a) / hsl(a)
├── 透明度：rgba 或 rgb(... / 50%)
└── 设计稿常用 hex
```

---

## 18. CSS 单位详解（px / rem / em / % / vw / vh）

### 18.1 px（像素）

绝对单位，屏幕上的一个点（严格说是 CSS 像素，与物理像素可能不同）。

```css
font-size: 16px;
width: 320px;
```

优点：直观。缺点：改根字号时，所有 px 都要手动改。

### 18.2 em

相对于**当前元素**的 `font-size`（若用于 font-size 则相对父元素）。

```css
.parent { font-size: 16px; }
.child { font-size: 1.5em; } /* 24px */
.child { padding: 1em; }       /* 相对自身 font-size */
```

问题：嵌套多层 em 会**累积放大**，容易算晕。

### 18.3 rem（root em）

相对于**根元素** `html` 的 `font-size`。

```css
html { font-size: 16px; }
.title { font-size: 1.5rem; }  /* 24px */
.body { font-size: 1rem; }     /* 16px */
```

现代布局**首选**配合 rem 做字号和间距。

### 18.4 %（百分比）

相对**父元素**同一属性（width 相对父 width，font-size 相对父 font-size）。

```css
.parent { width: 400px; }
.child { width: 50%; }  /* 200px */
```

### 18.5 vw / vh（视口宽度 / 高度）

- `1vw` = 视口宽度的 1%
- `1vh` = 视口高度的 1%
- `100vw` ≈ 整屏宽（注意可能含滚动条问题）

```css
.hero-title {
  font-size: 5vw;
}
.full-screen {
  height: 100vh;
}
```

### 18.6 还有哪些单位（了解）

| 单位 | 说明 |
|------|------|
| `ch` | 数字 "0" 宽度，常用于等宽输入 |
| `vmin` / `vmax` | vw 和 vh 中较小/较大者 |
| `fr` | Grid 布局专用 |

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  html { font-size: 16px; }
  .box-px { width: 200px; font-size: 14px; }
  .box-rem { font-size: 1.25rem; padding: 1rem; }
  .box-em-parent { font-size: 20px; }
  .box-em-parent .inner { font-size: 1.5em; padding: 1em; }
  .half-parent {
    width: 50%;
    background: #e0f2fe;
    padding: 8px;
  }
  .parent-fixed { width: 400px; border: 1px solid #ccc; }
  .vw-demo { font-size: 4vw; }
  </style>
</head>
<body>
  <p class="box-px">200px 宽，14px 字</p>
  <p class="box-rem">1.25rem 字，1rem 内边距</p>
  <div class="box-em-parent">
    <span class="inner">父 20px，子 1.5em 字 + 1em 内边距</span>
  </div>
  <div class="parent-fixed">
    <div class="half-parent">父 400px，我 50% 宽</div>
  </div>
  <p class="vw-demo">视口 4vw 大字</p>
</body>
</html>
```

### 怎么选单位（实用建议）

| 场景 | 推荐 |
|------|------|
| 字号、间距 | rem |
| 边框 1px | px |
| 宽度占父容器比例 | % |
| 全屏横幅高度 | vh |
| 组件内相对字号 | em（慎用嵌套） |

### 常见错误

- 给 `html` 设 `font-size: 62.5%` 把 1rem 当 10px 心算，但第三方组件可能不兼容
- `%` 高度无效：父元素没有明确 height 时，`height: 50%` 常无效
- `vw` 字号在小屏过大、大屏过小，需配合 `clamp()`

### 思维导图式小结

```
单位
├── px：固定
├── rem：相对根，推荐字号间距
├── em：相对自身/父字号，易嵌套失控
├── %：相对父元素
└── vw/vh：相对视口
```

---

## 19. 文本相关样式

### `color`

设置文字颜色。

### `font-size`

设置字体大小。

```css
p {
  font-size: 16px;
}
```

### `font-weight`

设置字重。

```css
font-weight: bold;    /* 或 700 */
font-weight: normal;  /* 400 */
font-weight: 600;     /* 半粗，标题常用 */
```

### `font-style`

```css
font-style: italic;   /* 斜体 */
font-style: normal;
```

### `line-height`

行高，无单位数字时相对自身字号：

```css
line-height: 1.6;   /* 推荐：1.5~1.8 正文易读 */
line-height: 24px;  /* 固定像素 */
```

### `text-align`

```css
text-align: left;    /* 默认 */
text-align: center;
text-align: right;
text-align: justify; /* 两端对齐，中文慎用 */
```

### 其他常用文本属性

```css
letter-spacing: 0.05em;   /* 字间距 */
word-spacing: 2px;          /* 词间距（英文） */
white-space: nowrap;        /* 不换行 */
text-overflow: ellipsis;    /* 溢出省略号，需配合 overflow:hidden */
```

### 完整可运行示例：文章排版

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  article {
    max-width: 680px;
    margin: 0 auto;
    padding: 24px;
  }
  article h1 {
    font-size: 2rem;
    font-weight: 700;
    line-height: 1.25;
    color: #0f172a;
  }
  article p {
    font-size: 1rem;
    line-height: 1.75;
    color: #334155;
    margin: 1em 0;
  }
  article .lead {
    font-size: 1.125rem;
    color: #475569;
  }
  </style>
</head>
<body>
  <article>
    <h1>文章标题</h1>
    <p class="lead">导语：略大的引言段落。</p>
    <p>正文第一段，行高 1.75 更适合长文阅读。</p>
    <p>正文第二段。</p>
  </article>
</body>
</html>
```

### 为什么重要

网页 80% 是文字。字号、行高、颜色对比度直接决定可读性和专业感。

### 常见错误

- `line-height` 太小（1.0）正文挤在一起
- `text-align: center` 用于长段落（大段居中难读）
- 浅灰字 `#ddd` 在白底上对比度不足（无障碍问题）

### 思维导图式小结

```
文本样式
├── 字号 font-size
├── 行高 line-height（正文 1.5~1.8）
├── 字重 font-weight
├── 对齐 text-align
└── 颜色 color
```

---

## 20. 字体相关样式

```css
body {
  font-family: "Microsoft YaHei", "PingFang SC", sans-serif;
}
```

### 常见字体族（通用族）

- `serif`：衬线，如 Times，传统印刷感
- `sans-serif`：无衬线，UI 最常见
- `monospace`：等宽，代码

### 字体栈

浏览器从左到右找第一个已安装的字体：

```css
font-family: "Helvetica Neue", Arial, sans-serif;
```

最后必须写通用族，防止全部未命中。

### `font` 简写

```css
font: italic bold 16px/1.5 "Microsoft YaHei", sans-serif;
/* style weight size/line-height family */
```

### Web 字体（了解）

```css
@font-face {
  font-family: "MyFont";
  src: url("./myfont.woff2") format("woff2");
}
```

Google Fonts 等可外链字体，注意加载性能。

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  body {
    font-family: system-ui, -apple-system, "Segoe UI", "Microsoft YaHei", sans-serif;
  }
  code, pre {
    font-family: Consolas, "Courier New", monospace;
    background: #f1f5f9;
    padding: 2px 6px;
    border-radius: 4px;
  }
  .serif-demo {
    font-family: Georgia, "Times New Roman", serif;
  }
  </style>
</head>
<body>
  <p>默认无衬线 UI 字体。</p>
  <p class="serif-demo">衬线字体示例。</p>
  <p>行内代码：<code>console.log()</code></p>
</body>
</html>
```

你现在重点理解：

- 字体可以设置候选列表
- 不同系统未必都有同一字体
- `system-ui` 可跟随系统默认 UI 字体

### 思维导图式小结

```
字体
├── font-family 字体栈
├── 结尾写 serif/sans-serif/monospace
└── 代码用 monospace
```

---

## 21. 文本装饰与大小写

### `text-decoration`

```css
a {
  text-decoration: none;      /* 去掉下划线 */
}
.strike {
  text-decoration: line-through;
}
.underline {
  text-decoration: underline;
}
```

可拆：`text-decoration-line`、`color`、`style`、`thickness`

### `text-transform`

```css
text-transform: uppercase;   /* 全大写 AB */
text-transform: lowercase;   /* 全小写 */
text-transform: capitalize;  /* 首字母大写 */
```

只影响**显示**，不改 HTML 源码文字。

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  a { color: #2563eb; text-decoration: none; }
  a:hover { text-decoration: underline; }
  .tag-upper { text-transform: uppercase; font-size: 12px; letter-spacing: 0.1em; }
  .old-price { text-decoration: line-through; color: #94a3b8; }
  </style>
</head>
<body>
  <a href="#">无下划线链接，悬停显示下划线</a>
  <p class="tag-upper">featured</p>
  <p><span class="old-price">¥199</span> ¥99</p>
</body>
</html>
```

### 思维导图式小结

```
文本装饰
├── text-decoration 线：无/下划线/删除线
├── text-transform 大小写显示
└── 链接：none + hover underline 很常见
```

---

## 22. 背景样式

### 背景色

```css
background-color: #f5f5f5;
```

### 背景图片

```css
background-image: url("./bg.png");
```

路径相对于 CSS 文件位置（不是 HTML）。

### 背景重复

```css
background-repeat: no-repeat;
/* repeat | repeat-x | repeat-y | no-repeat */
```

### 背景位置

```css
background-position: center;
/* top left | 50% 50% | 20px 10px */
```

### 背景尺寸

```css
background-size: cover;   /* 铺满，可能裁切 */
background-size: contain; /* 完整显示，可能留白 */
```

### 简写

```css
background: #f5f5f5 url("./bg.png") no-repeat center / cover;
/* color image repeat position / size */
```

### 完整可运行示例：英雄区

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  .hero {
    height: 40vh;
    background-color: #1e293b;
    background-image: linear-gradient(rgba(0,0,0,0.4), rgba(0,0,0,0.4)),
                      url("https://picsum.photos/1200/600");
    background-size: cover;
    background-position: center;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-size: 2rem;
  }
  </style>
</head>
<body>
  <section class="hero">欢迎访问</section>
</body>
</html>
```

### 常见错误

- 图片路径 404（检查相对路径）
- 只设 `background-image` 不设尺寸，大图撑破布局
- `background-size: cover` 在小屏裁切重要内容

### 思维导图式小结

```
背景
├── color / image / repeat / position / size
├── cover vs contain
└── 路径相对 CSS 文件
```

---

## 23. 边框

```css
border: 1px solid #ddd;
```

可以拆成：

- `border-width`：粗细
- `border-style`：`solid` `dashed` `dotted` `none`
- `border-color`：颜色

### 单边边框

```css
border-bottom: 2px solid #2563eb;
```

### `outline` 与 `border` 区别

`outline` 不占用布局空间，常用于 `:focus` 外框。

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  .card {
    border: 1px solid #e2e8f0;
    padding: 16px;
  }
  .card-accent {
    border-left: 4px solid #2563eb;
    padding-left: 12px;
  }
  input:focus {
    outline: 2px solid #93c5fd;
    outline-offset: 2px;
  }
  </style>
</head>
<body>
  <div class="card">四边边框卡片</div>
  <div class="card-accent">左侧强调线</div>
  <input type="text" placeholder="聚焦看 outline" />
</body>
</html>
```

### 思维导图式小结

```
边框
├── border: 宽 样式 色
├── 单边：border-top 等
└── outline：不占位，focus 常用
```

---

## 24. 圆角

```css
border-radius: 8px;
```

### 更多写法

```css
border-radius: 50%;           /* 正圆（元素宽高相等时） */
border-radius: 8px 16px;       /* 左上右下 / 右上左下 */
border-radius: 12px 12px 0 0; /* 仅上圆角 */
```

常见用于：

- 按钮
- 卡片
- 头像图片

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  .btn { border-radius: 6px; padding: 8px 16px; border: none; background: #2563eb; color: white; }
  .avatar {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    object-fit: cover;
  }
  .card-top-round {
    border-radius: 12px 12px 0 0;
    background: #f1f5f9;
    padding: 16px;
  }
  </style>
</head>
<body>
  <button class="btn">圆角按钮</button>
  <img class="avatar" src="https://picsum.photos/64" alt="头像" />
  <div class="card-top-round">仅上方圆角</div>
</body>
</html>
```

### 思维导图式小结

```
圆角 border-radius
├── 像素值 / 50% 圆形
├── 四值顺序：左上 右上 右下 左下
```

---

## 25. 阴影

```css
box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
```

参数大致为：水平偏移 垂直偏移 模糊 扩散 颜色（可 inset 内阴影）。

文字阴影：

```css
text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.3);
```

作用：

- 增强层次感
- 卡片「浮起」效果
- 悬停时加深阴影做反馈

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  .card {
    padding: 24px;
    border-radius: 12px;
    background: white;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    transition: box-shadow 0.2s;
  }
  .card:hover {
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  }
  .title {
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
    color: white;
    background: #334155;
    padding: 12px;
  }
  </style>
</head>
<body>
  <div class="card">悬停加深阴影</div>
  <h2 class="title">文字阴影标题</h2>
</body>
</html>
```

### 思维导图式小结

```
阴影
├── box-shadow 盒子
├── text-shadow 文字
└── 半透明黑 rgba 更自然
```

---

## 26. 列表样式

```css
ul {
  list-style: none;
  padding: 0;
  margin: 0;
}
```

### 其他 list 属性

```css
list-style-type: disc;   /* 圆点 */
list-style-type: decimal; /* 数字，用于 ol */
list-style-position: inside;
```

实际开发中经常把导航列表的默认圆点去掉，再用 Flex 排版。

### 完整可运行示例：导航列表

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  .nav {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    gap: 16px;
  }
  .nav a {
    text-decoration: none;
    color: #334155;
  }
  .nav a:hover { color: #2563eb; }
  </style>
</head>
<body>
  <ul class="nav">
    <li><a href="#">首页</a></li>
    <li><a href="#">产品</a></li>
    <li><a href="#">关于</a></li>
  </ul>
</body>
</html>
```

### 思维导图式小结

```
列表
├── list-style: none 去默认符号
├── 导航常配合 flex + gap
└── 记得清 padding/margin
```

---

## 27. 链接样式建议

```css
a {
  text-decoration: none;
  color: inherit;
}
```

但要注意：

- 不是所有链接都该彻底失去「可点击感」
- 至少保留 `:hover`、`:focus` 变化
- 正文链接可与周围文字区分颜色

### 推荐模式

```css
a { color: #2563eb; text-decoration: none; }
a:hover { text-decoration: underline; }
a:focus-visible {
  outline: 2px solid #2563eb;
  outline-offset: 2px;
}
```

### 思维导图式小结

```
链接样式
├── 去掉默认下划线很常见
├── 保留 hover/focus 反馈
└── color: inherit 用于导航父级控制颜色
```

---

## 28. !important 与优先级计算

CSS 不是后写就一定赢，它由**优先级（特异性 specificity）**和**来源顺序**共同决定。

### 优先级权重（心算版）

| 选择器类型 | 权重示意 |
|------------|----------|
| 内联 style | 1000（最高之一） |
| `#id` | 100 |
| `.class`、`:hover`、`::before`、[attr] | 10 |
| `div`、`p` | 1 |
| `*` | 0 |

实际计算：比较 (id 个数, class 个数, 元素个数)，**不把它们加成千进制数字**（旧教材 1000 算法是简化记忆）。

示例：

```css
p { color: black; }           /* 0,0,1 */
.intro { color: blue; }        /* 0,1,0 赢 */
#main .intro { color: green; } /* 1,1,0 再赢 */
```

### `!important`

```css
p {
  color: red !important;
}
```

作用：在同来源内几乎压过一切普通声明。

### 为什么不要轻易用 !important

- 以后想覆盖只能再加 `!important`，陷入恶性循环
- 第三方库、主题切换难以维护
- 调试时 Styles 面板一片黄标

### 合法使用场景

- 工具类 `.hidden { display: none !important; }`（争议但常见）
- 覆盖第三方内联样式（临时）
- 打印样式 `@media print`

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  p { color: gray; }
  .text { color: blue; }
  #special { color: green; }
  p.text { color: orange; }  /* 0,1,1 比 .text 更具体 */
  </style>
</head>
<body>
  <p>灰色</p>
  <p class="text">橙色的 0,1,1</p>
  <p class="text" id="special">绿色 id 赢</p>
</body>
</html>
```

### 层叠顺序（简化）

1. 重要性：`!important` 普通 > 普通 `!important` > 普通规则（同属性时）
2. 特异性：高者优先
3. 顺序：后写覆盖先写

### 调试技巧

开发者工具中，被划掉的规则是「输了」；看 specificity 可安装浏览器扩展或观察规则顺序。

### 常见错误

- 认为「写在后面的 class 一定赢过前面的 id」（id 通常仍赢）
- 到处 `!important` 修 bug

### 思维导图式小结

```
优先级
├── 内联 > id > class > 元素
├── !important 慎用
├── 后写覆盖先写（同权重时）
└── 用更合理的选择器，而非 !important
```

---

## 29. 继承基础认知

有些 CSS 属性会被子元素**继承**，例如：

- `color`
- `font-family`、`font-size`
- `line-height`
- `text-align`（块级子元素）

### 不会继承的典型属性

- `margin`、`padding`、`border`
- `background`
- `width`、`height`

### 强制继承

```css
.parent {
  border: 1px solid red;
}
.child {
  border: inherit; /* 少见 */
}
```

### 控制继承

```css
color: red;
all: unset;    /* 重置几乎所有属性 */
all: revert;   /* 回浏览器默认 */
```

### 完整可运行示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  .parent {
    color: #2563eb;
    font-size: 20px;
    border: 2px solid #e2e8f0;
    padding: 16px;
  }
  .child {
    /* color 和 font-size 继承，border 不继承 */
  }
  </style>
</head>
<body>
  <div class="parent">
    父元素
    <p class="child">子段落：蓝色、20px，无边框</p>
  </div>
</body>
</html>
```

### 为什么重要

给 `body` 设好字体和颜色，全站大部分文字自动统一；不必每个 `p`、`span` 都写一遍。

### 思维导图式小结

```
继承
├── 文字类常继承
├── 盒模型背景常不继承
└── body 设字体=color 做全局基准
```

---

## 30. CSS 变量入门（自定义属性）

CSS 变量也叫自定义属性，在 `:root` 或任意选择器上定义，用 `var()` 读取。

```css
:root {
  --color-primary: #2563eb;
  --spacing-md: 16px;
  --font-base: 16px;
}

.button {
  background: var(--color-primary);
  padding: var(--spacing-md);
  font-size: var(--font-base);
}
```

### 默认值

```css
color: var(--text-color, #333); /* 未定义时用 #333 */
```

### 为什么重要

- 改一处主题色，全站按钮、链接一起变
- 配合 `prefers-color-scheme` 做暗色模式
- 比 Sass 变量简单，浏览器原生支持

### 与 Sass 变量区别

CSS 变量是**运行时**的，可随媒体查询、class 切换改变；Sass 在编译时固定。

### 完整可运行示例：主题切换思路

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
  :root {
    --bg: #ffffff;
    --text: #1e293b;
    --primary: #2563eb;
  }
  .theme-dark {
    --bg: #0f172a;
    --text: #f1f5f9;
    --primary: #60a5fa;
  }
  body {
    background: var(--bg);
    color: var(--text);
    font-family: sans-serif;
    padding: 24px;
  }
  .btn {
    background: var(--primary);
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 6px;
    cursor: pointer;
  }
  </style>
</head>
<body>
  <button class="btn" onclick="document.body.classList.toggle('theme-dark')">
    切换暗色
  </button>
  <p>点击按钮，CSS 变量改变，颜色跟着变。</p>
</body>
</html>
```

### 常见错误

- 忘记 `--` 前缀
- `var(color-primary)` 少了 `var()` 和 `--`
- 在 IE 不支持（现代项目可忽略）

### 思维导图式小结

```
CSS 变量
├── 定义：--name: value
├── 使用：var(--name, 默认值)
├── 常放 :root
└── 主题色、间距统一管理
```

---

## 31. 完整实战示例一：个人卡片

将 HTML 结构与 CSS 样式分离，做出可复用的个人资料卡。

### HTML

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>个人卡片</title>
  <link rel="stylesheet" href="profile-card.css" />
</head>
<body>
  <main class="page">
    <article class="profile-card">
      <img class="profile-card__avatar" src="https://picsum.photos/120" alt="张三的头像" />
      <h1 class="profile-card__name">张三</h1>
      <p class="profile-card__role">前端学习者</p>
      <p class="profile-card__bio">正在系统学习 HTML、CSS 和 JavaScript，目标是做出干净好用的网页。</p>
      <ul class="profile-card__tags">
        <li>HTML</li>
        <li>CSS</li>
        <li>JavaScript</li>
      </ul>
      <div class="profile-card__actions">
        <a class="btn btn--primary" href="#">关注我</a>
        <a class="btn btn--ghost" href="#">发消息</a>
      </div>
    </article>
  </main>
</body>
</html>
```

### profile-card.css

```css
:root {
  --card-bg: #ffffff;
  --text-main: #0f172a;
  --text-muted: #64748b;
  --primary: #2563eb;
  --border: #e2e8f0;
  --radius: 16px;
  --shadow: 0 8px 24px rgba(15, 23, 42, 0.08);
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family: system-ui, "Microsoft YaHei", sans-serif;
  background: #f1f5f9;
  color: var(--text-main);
}

.page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.profile-card {
  max-width: 360px;
  width: 100%;
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  padding: 32px 24px;
  text-align: center;
}

.profile-card__avatar {
  width: 96px;
  height: 96px;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid #e0f2fe;
}

.profile-card__name {
  margin: 16px 0 4px;
  font-size: 1.5rem;
  font-weight: 700;
}

.profile-card__role {
  margin: 0;
  color: var(--primary);
  font-weight: 600;
  font-size: 0.95rem;
}

.profile-card__bio {
  margin: 16px 0;
  color: var(--text-muted);
  line-height: 1.6;
  font-size: 0.95rem;
}

.profile-card__tags {
  list-style: none;
  padding: 0;
  margin: 0 0 24px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
}

.profile-card__tags li {
  background: #f1f5f9;
  color: #475569;
  padding: 4px 12px;
  border-radius: 999px;
  font-size: 0.85rem;
}

.profile-card__actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.btn {
  display: inline-block;
  padding: 10px 20px;
  border-radius: 8px;
  text-decoration: none;
  font-size: 0.9rem;
  font-weight: 600;
  transition: background 0.2s, color 0.2s;
}

.btn--primary {
  background: var(--primary);
  color: white;
}

.btn--primary:hover {
  background: #1d4ed8;
}

.btn--ghost {
  border: 1px solid var(--border);
  color: var(--text-main);
}

.btn--ghost:hover {
  background: #f8fafc;
}
```

### 本示例用到的知识点

- 外部 CSS、CSS 变量、class 命名
- 圆角头像、`box-shadow`、标签 pills
- `:hover`、Flex 布局（下一份文档会深入 Flex）
- `rem`、颜色、行高、列表去样式

### §31 个人卡片 CSS 逐行读

| 行号/片段 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `:root { --primary: #2563eb; ... }` | 全局 CSS 变量，全页可 `var()` 引用 | 变量名漏 `--` 则无效 |
| `* { box-sizing: border-box; }` | 宽高含 padding/border，避免撑破布局 | 不设时 `width:100%`+padding 可能溢出 |
| `.page { min-height: 100vh; display: flex; ... }` | 卡片在视口垂直水平居中 | 仅 `margin:auto` 有时达不到垂直居中 |
| `.profile-card { max-width: 360px; box-shadow: ... }` | 限制卡片宽度 + 浮起阴影 | 无 max-width 宽屏时卡片过宽 |
| `.profile-card__avatar { border-radius: 50%; object-fit: cover; }` | 圆形头像，图片裁切填满 | 缺 object-fit 非正方形图会变形 |
| `.profile-card__tags li { border-radius: 999px; }` | 大圆角 = 胶囊形标签 | 小圆角则呈圆角矩形 |
| `.btn--primary { background: var(--primary); }` | 主按钮用变量色，改变量全站同步 | 写死色值则改主题要搜多处 |
| `.btn--primary:hover { background: #1d4ed8; }` | 悬停反馈，无需 JavaScript | 无 hover 时交互感弱 |

---

## 32. 完整实战示例二：导航栏

### HTML

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>导航栏</title>
  <link rel="stylesheet" href="navbar.css" />
</head>
<body>
  <header class="site-header">
    <div class="site-header__inner">
      <a class="site-logo" href="#">MySite</a>
      <nav class="site-nav" aria-label="主导航">
        <ul class="site-nav__list">
          <li><a class="site-nav__link is-active" href="#">首页</a></li>
          <li><a class="site-nav__link" href="#">文档</a></li>
          <li><a class="site-nav__link" href="#">博客</a></li>
          <li><a class="site-nav__link" href="#">关于</a></li>
        </ul>
      </nav>
      <a class="site-header__cta" href="#">登录</a>
    </div>
  </header>
  <main class="demo-content">
    <p>下方是页面内容区，导航栏固定在顶部。</p>
  </main>
</body>
</html>
```

### navbar.css

```css
:root {
  --header-bg: #0f172a;
  --header-text: #f8fafc;
  --header-muted: #94a3b8;
  --accent: #38bdf8;
  --header-height: 64px;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family: system-ui, "Microsoft YaHei", sans-serif;
}

.site-header {
  background: var(--header-bg);
  color: var(--header-text);
  position: sticky;
  top: 0;
  z-index: 100;
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.06);
}

.site-header__inner {
  max-width: 1100px;
  margin: 0 auto;
  padding: 0 24px;
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.site-logo {
  color: var(--header-text);
  text-decoration: none;
  font-weight: 700;
  font-size: 1.25rem;
}

.site-logo:hover {
  color: var(--accent);
}

.site-nav__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  gap: 8px;
}

.site-nav__link {
  display: block;
  padding: 8px 14px;
  color: var(--header-muted);
  text-decoration: none;
  border-radius: 6px;
  font-size: 0.95rem;
  transition: color 0.2s, background 0.2s;
}

.site-nav__link:hover {
  color: var(--header-text);
  background: rgba(255, 255, 255, 0.06);
}

.site-nav__link.is-active {
  color: var(--header-text);
  background: rgba(56, 189, 248, 0.15);
}

.site-nav__link.is-active::after {
  content: "";
  /* 当前页指示可仅用背景色，此处留空扩展 */
}

.site-header__cta {
  padding: 8px 16px;
  background: var(--accent);
  color: #0f172a;
  text-decoration: none;
  border-radius: 6px;
  font-weight: 600;
  font-size: 0.9rem;
}

.site-header__cta:hover {
  filter: brightness(1.1);
}

.demo-content {
  padding: 48px 24px;
  max-width: 1100px;
  margin: 0 auto;
}
```

### 调试与扩展建议

- 缩小窗口看导航是否挤在一起（移动端汉堡菜单留到响应式章节）
- 用 Tab 键切换链接，检查 `:focus` 是否可见
- 尝试改 `--accent` 看全栏强调色变化

### §32 导航栏 CSS 逐行读

| 行号/片段 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `.site-header { position: sticky; top: 0; z-index: 100; }` | 滚动时导航吸顶 | 无 sticky 则滚出视口 |
| `--header-height: 64px` + `height: var(--header-height)` | 统一导航高度，便于对齐 | 高度不一致时 logo 与链接垂直不齐 |
| `.site-header__inner { display: flex; justify-content: space-between; }` | 左 logo、中导航、右按钮分布 | 不用 flex 则靠 float/空格难维护 |
| `.site-nav__list { list-style: none; display: flex; gap: 8px; }` | 去圆点 + 横向菜单 + 间距 | 忘记 `list-style:none` 会有黑点 |
| `.site-nav__link:hover { background: rgba(255,255,255,0.06); }` | 半透明悬停底，深色导航常用 | 实色块可能过于抢眼 |
| `.site-nav__link.is-active { background: rgba(56,189,248,0.15); }` | 当前页高亮（配合 HTML class） | 缺 is-active 用户不知在哪一页 |
| `.site-header__cta { filter: brightness(1.1); }` on hover | 悬停略提亮，不改布局 | 用 scale 有时更好，filter 更简单 |

---

## 33. 初学者常见错误（扩充）

### 33.1 结构和样式混着写

建议尽量使用外部 CSS 文件。行内样式只适合临时调试。

### 33.2 class 命名混乱

`div1`、`red-box`、`a` 这类名字后面很难维护。用 `profile-card`、`nav-link` 等语义化命名。

### 33.3 只会靠空格和 `<br>` 排版

这是 HTML 初学者转 CSS 时最常见的问题。间距应使用 `margin`/`padding`，布局用 Flex/Grid（后续学）。

### 33.4 不理解选择器范围

「为什么页脚链接也变红了？」——因为你写了 `a { color: red }` 而不是 `.article a`。

### 33.5 样式「不生效」排查清单

1. 选择器是否选中了元素？（Elements 面板看）
2. 是否被更高优先级覆盖？（看划掉的规则）
3. 属性名/值是否写错？
4. CSS 文件是否加载成功？（Network）
5. 是否写在错误的文件或 `<style>` 块里？

### 33.6 忘记盒模型

设置了 `width: 100%` 又加 `padding`，可能撑出父容器。可先设 `box-sizing: border-box`（下一章详讲）。

### 33.7 滥用绝对定位

应用 `position: absolute` 摆一切，响应式会崩。优先正常文档流 + Flex/Grid。

### 调试技巧总汇

| 问题 | 工具操作 |
|------|----------|
| 谁覆盖了我的样式 | F12 → Styles，看划线规则 |
| 元素尺寸不对 | Computed → box model 图 |
| 悬停样式 | Styles → :hov |
| 文件没加载 | Network 筛选 CSS |

---

## 34. 分级练习题

### 基础级（必做）

1. 用**外部 CSS** 把页面 `body` 背景设为 `#f8fafc`，正文字号 `16px`，颜色 `#334155`。
2. 写选择器：只把 `class="warning"` 的 `p` 设为橙色，不影响其他 `p`。
3. 去掉列表圆点，做横向三个链接（首页、关于、联系）。
4. 给按钮加 `:hover` 时背景变深。
5. 用 `border-radius` 做圆角卡片 + 浅灰边框。

### 进阶级

1. 用 `nth-child(odd)` 给表格奇数行加背景色。
2. 用属性选择器：`input[type="password"]` 与 `input[type="text"]` 不同边框色。
3. 用 `::before` 给必填项标签加红色 `*`（`class="required"`）。
4. 用 `:root` 定义 `--primary`，至少 3 个组件引用 `var(--primary)`。
5. 解释：`.nav a` 和 `.nav > a` 在有嵌套菜单时区别是什么？写 HTML 验证。

### 挑战级

1. 不做个人卡片照抄，自己设计一版资料卡（至少含头像、姓名、简介、两个按钮），用 CSS 变量管理颜色。
2. 导航栏：当前页 `class="is-active"` 高亮；悬停与焦点样式分开写（`:hover` 与 `:focus-visible`）。
3. 写一段 HTML + CSS，故意制造优先级冲突，再用更合理的选择器修复，**不要用** `!important`。
4. 英雄区：`height: 50vh`，背景图 `cover`，标题白色 + `text-shadow`，副标题 `rgba` 半透明。
5. 阅读自己写的 CSS，列出每条规则的选择器权重 `(id, class, element)`。

### 练习自检标准

- 改 HTML 结构后样式仍合理（class 没绑死标签名）
- 开发者工具无大片 invalid 属性
- 链接能 Tab 聚焦且看得见焦点

---

## 35. FAQ 常见问题

### Q1：CSS 写在哪里最好？

**答**：实际项目用外部 `.css` 文件；学习 demo 可用 `<style>`；行内仅调试。

### Q2：class 和 id 怎么选？

**答**：样式几乎总是 class；id 留给 JS 锚点或唯一元素，且不宜过多。

### Q3：样式没生效怎么办？

**答**：F12 看元素是否匹配选择器、规则是否被覆盖、CSS 是否 404。

### Q4：px 还是 rem？

**答**：初学 px 直观；习惯后字号间距推荐 rem，边框 1px 仍用 px。

### Q5：`:` 和 `::` 有什么区别？

**答**：`:` 伪类（状态）；`::` 伪元素（虚拟节点）。旧代码 `:before` 也常能用。

### Q6：什么时候用 `!important`？

**答**：尽量少用。优先改选择器结构；工具类或覆盖第三方时是少数例外。

### Q7：为什么设置了 `height: 50%` 无效？

**答**：百分比高度依赖父元素有明确 height；可试 `50vh` 或给父级设高度。

### Q8：中文用什么 `font-family`？

**答**：`"Microsoft YaHei"`、`"PingFang SC"`、`system-ui`、sans-serif 组合。

### Q9：怎么去掉 a 下划线还不像链接？

**答**：保留颜色变化 + `:hover` 下划线或 `:focus-visible` 轮廓。

### Q10：CSS 变量和 Sass 变量选哪个？

**答**：浏览器原生用 CSS 变量；大型项目可能 Sass + CSS 变量并存，先掌握 CSS 变量。

### Q11：一张图做背景模糊/变暗？

**答**：用 `linear-gradient` 叠在 `background-image` 上层（见第 22 节英雄区示例）。

### Q12：学 CSS 要背多少属性？

**答**：不必背表。本章 30+ 个常用属性熟练后，其余查 [MDN](https://developer.mozilla.org/zh-CN/docs/Web/CSS) 即可。

---

## 36. 练习建议

建议你自己做：

1. 给上一篇 HTML 页面加样式
2. 做一个简单个人卡片（可参考第 31 节，鼓励自己改设计）
3. 做一个基础导航栏（可参考第 32 节）
4. 做一个文章页面的排版（限制宽度、行高、标题层级）
5. 完成第 34 节分级练习至少「基础 + 进阶」各 3 题

---

## 37. 学完标准

如果你能做到这些，这一份就掌握得不错：

- 会写基础 CSS 语法，三种引入方式知道何时用哪种
- 会使用常见选择器：元素、class、id、后代/子代、并集/交集、伪类、伪元素、属性选择器
- 理解 px / rem / em / % / vw / vh 并会选用
- 会控制颜色、字体、背景、边框、圆角、阴影、文本样式
- 知道 class 和 id 的区别，以及优先级、继承、CSS 变量的基本用法
- 能独立完成个人卡片和导航栏静态样式
- 会用开发者工具排查「样式不生效」

---

## 39. 选择器组合实战练习（10题）

### 题目

用以下 HTML 结构回答问题：

```html
<div class="page">
  <header id="site-header">
    <h1 class="logo">MySite</h1>
    <nav class="main-nav">
      <ul>
        <li class="active"><a href="/">首页</a></li>
        <li><a href="/blog">博客</a></li>
        <li><a href="/about">关于</a></li>
      </ul>
    </nav>
  </header>
  <main>
    <article class="post featured">
      <h2><a href="/post/1">文章标题</a></h2>
      <p class="excerpt">摘要文字</p>
      <a href="/post/1" class="read-more" title="阅读全文">阅读更多</a>
    </article>
    <article class="post">
      <h2><a href="/post/2">另一篇</a></h2>
      <p class="excerpt">另一篇摘要</p>
      <a href="/post/2" class="read-more">阅读更多</a>
    </article>
  </main>
  <aside class="sidebar">
    <div class="widget" data-type="tags">
      <h3>标签云</h3>
    </div>
  </aside>
  <footer>
    <p>&copy; 2026</p>
  </footer>
</div>
```

| # | 需求 | 答案 |
|---|------|------|
| 1 | 选中所有文章的标题链接 | `.post h2 a` |
| 2 | 只选中精选文章的标题链接 | `.post.featured h2 a` |
| 3 | 选中导航栏中所有链接 | `.main-nav a` |
| 4 | 只选中当前激活的导航项 | `.main-nav .active` 或 `.main-nav li:first-child` |
| 5 | 选中所有 `href` 以 `/post/` 开头的链接 | `a[href^="/post/"]` |
| 6 | 选中带 `title` 属性的链接 | `a[title]` |
| 7 | 选中第一篇文章的摘要 | `.post:first-child .excerpt` 或 `article:first-of-type .excerpt` |
| 8 | 选中 `data-type="tags"` 的 widget | `[data-type="tags"]` |
| 9 | 选中 id 为 site-header 的直接子元素 h1 | `#site-header > h1` |
| 10 | 选中页面第一个 `article` 后面的所有 `article` | `article + article` 或 `article ~ article` |

### 练习建议

在浏览器里用 `document.querySelectorAll("你的选择器")` 验证结果，看是否选中了预期的元素。

---

## 40. `clamp()` 函数 — 流体响应式

无需媒体查询即可实现字号随屏幕平滑变化：

```css
/* clamp(最小值, 理想值（通常是vw）, 最大值) */
.hero-title {
  font-size: clamp(1.5rem, 5vw, 3rem);
  /* 最小 1.5rem，最大 3rem，中间按 5% 视口宽度平滑变化 */
}

.card {
  width: clamp(280px, 80%, 400px);
  /* 最小 280px，最大 400px，中间 80% */
  padding: clamp(1rem, 3vw, 2rem);
}
```

实用场景：大标题在手机上缩小、桌面放大，且不需要写媒体查询。

---

## 41. `accent-color` 等现代 CSS 属性

```css
/* 统一表单控件的主题色 */
input[type="checkbox"],
input[type="radio"],
input[type="range"],
progress {
  accent-color: #6366f1;
}

/* 选中文字的颜色 */
::selection {
  background: #6366f1;
  color: white;
}

/* 平滑滚动（一个属性搞定全局平滑滚动） */
html {
  scroll-behavior: smooth;
}

/* 图片适应容器 */
img {
  object-fit: cover;  /* 裁切填充 */
  aspect-ratio: 16/9; /* 固定宽高比 */
}

/* 暗色模式适配 */
@media (prefers-color-scheme: dark) {
  body {
    background: #1e293b;
    color: #f1f5f9;
  }
}
```

---

## 42. 自定义滚动条样式

```css
/* 全局自定义滚动条（Chrome/Edge/Safari） */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
::-webkit-scrollbar-track {
  background: #f1f5f9;
  border-radius: 4px;
}
::-webkit-scrollbar-thumb {
  background: #94a3b8;
  border-radius: 4px;
}
::-webkit-scrollbar-thumb:hover {
  background: #64748b;
}

/* Firefox */
* {
  scrollbar-width: thin;
  scrollbar-color: #94a3b8 #f1f5f9;
}
```

---

## 43. 常见 CSS bug 案例分析

### 案例 1：图片下方 3px 空白

```html
<div><img src="photo.jpg" /></div>
<!-- div 比图片高约 3px -->
```

**原因**：图片默认是 `inline` 元素，基线对齐会在下方留空隙。

**解决**：
```css
img { display: block; }
/* 或 */
img { vertical-align: top; }
```

### 案例 2：margin 穿透到父元素上面

```html
<div class="card">
  <h2>标题</h2>
</div>
```
```css
.card { /* 没有 border/padding */ }
.card h2 { margin-top: 30px; }
/* 30px 的间距出现在 .card 上方而不是 h2 上方! */
```

**原因**：父子 margin 合并。

**解决**：给 `.card` 加 `padding-top: 1px` 或 `overflow: hidden` 或 `display: flow-root`。

### 案例 3：width:100% + padding 导致溢出

```css
.box {
  width: 100%;
  padding: 20px;
  /* 实际宽度 = 100% + 40px，可能溢出父容器 */
}
```

**解决**：全站设 `* { box-sizing: border-box; }`。

### 案例 4：inline-block 元素之间有空白

```html
<span class="tag">HTML</span>
<span class="tag">CSS</span>
<span class="tag">JS</span>
<!-- 渲染出来: HTML CSS JS — 标签间有空格！ -->
```

**原因**：HTML 中的换行符被渲染为空格。

**解决**：父元素 `font-size: 0` 然后子元素恢复字号；或用 Flex 布局。

### 案例 5：z-index 9999 还是被挡住

**原因**：元素不在同一个层叠上下文中比较 z-index。

**解决**：检查父级是否有 `position` + `z-index`、`opacity < 1`、`transform` 等创建了新层叠上下文。提升父级 z-index 或调整 DOM 结构。

---

## 44. Elements 面板八步实操（CSS 调试）

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 打开 §31 个人卡片 HTML + 外部 CSS，F12 → **Elements** | 选中 `.profile-card` 能看到 Matched Rules | Network 看 CSS 是否 404 |
| 2 | 右侧 **Styles** 勾选 `:hov` → `:hover` 在 `.btn--primary` 上 | 悬停背景色变深规则生效 | 选择器是否写错 `.btn-primary` |
| 3 | 在 Styles 里临时改 `--primary` 的值 | 主按钮和 role 文字色一起变 | 变量是否定义在 `:root` |
| 4 | 点击规则左侧 checkbox 取消某条 `color` | 文字色立刻变化/恢复 | 学会开关单条声明 |
| 5 | 看被**划掉**的灰色规则 | 上面有更具体或 `!important` 的规则赢了 | 对照 §28 优先级 |
| 6 | 切到 **Computed** 面板，搜 `font-size` | 看到最终生效值及来源文件行号 | 继承自 body 还是自身 class |
| 7 | 打开 §32 导航栏，Tab 键聚焦 `.site-nav__link` | 应能看到 `:focus` 或 `:focus-visible` 轮廓 | 若 `outline:none` 且无替代则无障碍不合格 |
| 8 | **Computed** → box model 图点 padding/margin | 改数值实时看布局变化 | 排查「多出来的空白」 |

### 样式不生效排查表（与 §33.5 配合）

| 症状 | Elements 里先查 |
|------|-------------------|
| 完全没样式 | Network 中 CSS 是否 200；`<link rel="stylesheet">` 是否完整 |
| 部分元素没变 | 选择器是否匹配（Elements 里右键 Copy → selector） |
| 改了没反应 | 是否被更高优先级覆盖（看划掉规则） |
| 颜色对了布局不对 | Computed 里 width/height/margin/padding |

---

## 45. 闭卷自测

1. CSS 全称是什么？「层叠」指的是什么？
2. 三种引入 CSS 的方式中，项目开发推荐哪种？为什么？
3. 元素选择器、class 选择器、id 选择器的权重谁高谁低（简化记忆）？
4. `.article p` 和 `.article > p` 有什么区别？各举一个适用场景。
5. `:hover` 和 `::before` 分别属于伪类还是伪元素？伪元素为什么必须写 `content`？
6. `rem` 相对谁？`em` 用在 `padding` 上相对谁？
7. 正文段落推荐 `line-height` 大约多少？为什么长段落不宜 `text-align: center`？
8. `border` 和 `outline` 在是否占据布局空间上有什么区别？
9. 什么属性子元素通常会继承？`margin` 会继承吗？
10. **动手**：写 CSS 只把 `class="warning"` 的 `p` 设为橙色，不影响其他 `p`。
11. **动手**：用 `:root` 定义 `--primary`，让 `.btn` 背景和 `a` 链接色都引用它。
12. **综合**：导航栏链接悬停有效但键盘 Tab 聚焦看不见，你会改哪条 CSS？

### 45.1 自测参考答案

1. Cascading Style Sheets 层叠样式表；多条规则冲突时的裁决机制（优先级+顺序）。
2. 外部 `.css`；结构样式分离、可缓存、多页复用、易维护。
3. id > class > 元素（同权重时后写覆盖先写）。
4. 后代：`.article` 内任意深度 `p`；子代：仅直接子 `p`。侧栏深层段落 vs 仅文章直接段落。
5. `:hover` 伪类；`::before` 伪元素；无 `content` 则不生成伪元素盒。
6. `rem` 相对 `html` 根字号；`padding: 1em` 相对**当前元素自身** font-size。
7. 约 1.5～1.8；大段居中每行长短不一，阅读费力。
8. `border` 占布局空间；`outline` 不占位，画在边框外。
9. 常继承：color、font-family、font-size、line-height 等；margin **不**继承。
10. `p.warning { color: orange; }` 或 `.warning` 若只在 p 上则 `.warning { color: orange; }`。
11. `:root { --primary: #2563eb; }` `.btn { background: var(--primary); }` `a { color: var(--primary); }`。
12. 补 `:focus-visible { outline: 2px solid ...; outline-offset: 2px; }`，勿只写 `:hover`。

---

## 46. 费曼检验

请在不看资料的情况下，用 3 分钟向朋友解释本章核心：

1. **CSS = 装修**：HTML 是毛坯房结构，CSS 选「刷哪面墙、用什么漆」——选择器决定范围，属性决定样子。
2. **三种引入方式**：装修图纸最好单独一本（外部 CSS），不要写每块砖上（行内 style）。
3. **层叠与变量**：不是后写一定赢，要看谁更「具体」；CSS 变量像全屋色卡，改一处按钮链接一起变。

> 下一章：[04 — CSS 布局 Flex 与 Grid](./05-CSS布局FlexGrid响应式与动画.md)。会选元素、会配色之后，学怎么「摆放家具」。

---

## 47. 与 02 章表单联动：属性选择器实战

把 [02 章表单](./02-HTML表单表格多媒体与语义化.md) 的 HTML 与本章 CSS 连起来——**不用给每个 input 加 class**，也能区分样式：

```html
<!-- login-styled.html：结构来自 02 章 §31，样式如下 -->
<form class="login-form" action="/api/login" method="post">
  <p>
    <label for="email">邮箱</label>
    <input id="email" name="email" type="email" required autocomplete="email" />
  </p>
  <p>
    <label for="password">密码</label>
    <input id="password" name="password" type="password" required minlength="6" />
  </p>
  <button type="submit">登录</button>
</form>
```

```css
/* login-form.css */
.login-form {
  max-width: 360px;
  margin: 2rem auto;
  padding: 24px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.login-form label {
  display: block;
  font-weight: 600;
  margin-bottom: 6px;
}

/* 属性选择器：按 type 区分，无需额外 class */
.login-form input[type="email"],
.login-form input[type="password"] {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 1rem;
}

.login-form input:focus {
  outline: 2px solid #93c5fd;
  border-color: #2563eb;
}

.login-form input:invalid:not(:placeholder-shown) {
  border-color: #ef4444; /* 有内容但不合法时红框 */
}

.login-form button[type="submit"] {
  width: 100%;
  margin-top: 16px;
  padding: 12px;
  background: #2563eb;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  cursor: pointer;
}

.login-form button[type="submit"]:hover {
  background: #1d4ed8;
}
```

### 联动练习逐行读

| 选择器 | 含义 | 改错会怎样 |
|--------|------|------------|
| `.login-form input[type="email"]` | 只选表单内邮箱框 | 写成 `input[type=email]` 缺引号在部分场景有问题 |
| `:focus` | 键盘/鼠标聚焦时 | 去掉则 Tab 用户不知焦点在哪 |
| `:invalid:not(:placeholder-shown)` | 已输入但不合法 | 空字段不会误标红（避免一打开就红） |
| `button[type="submit"]` | 只样式提交钮 | 若页内还有其他 button 需更具体 |

**练习**：在 §31 登录页上挂此外部 CSS，用 Elements 分别聚焦邮箱、输入错误格式、悬停按钮，确认三条状态样式都生效。
