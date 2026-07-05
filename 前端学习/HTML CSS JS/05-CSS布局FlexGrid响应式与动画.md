# CSS 布局、Flex、Grid、响应式与动画

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0 导读（Flex=弹簧排排站、Grid=Excel表格）、DevTools Layout overlay、FAQ≥12、闭卷自测、费曼检验、衔接 Vue 09 -->

> **文件编码**：UTF-8。写 HTML 时请在 `<head>` 中加入 `<meta charset="UTF-8" />`。

---

## 0. 读前导读（零基础也能跟上）

> **读者假设**：你已读完 [04-盒模型与定位](./04-CSS盒模型浮动定位与显示模式.md)，知道元素如何占位、角标如何用 `absolute`。本章进入**现代布局主战场**：Flex 一行排、Grid 整张表、响应式适配与动效。

### 0.1 用一句话弄懂本章

**一句话**：页面主结构用 **Flex**（一维弹簧排排站）或 **Grid**（二维 Excel 表格）分配空间，用**媒体查询**让手机到桌面都好用，用 **transition / animation** 让交互更顺滑。

**生活类比**：

| CSS 概念 | 生活类比 | 本章位置 |
|----------|----------|----------|
| **Flex 容器** | 一排小朋友拉手站队，队长说「靠左/居中/均分」 | §3～§8 |
| **flex-grow** | 弹簧：剩余空地谁更能「长高变宽」谁多占 | §7.1 |
| **justify-content** | 队长喊「排头对齐 / 两头对齐 / 中间散开」 | §6.3 |
| **align-items** | 队长喊「站直齐肩 / 脚底对齐 / 头顶对齐」 | §6.4 |
| **Grid 容器** | Excel 表格：先画几行几列，再往格子里填内容 | §9～§12 |
| **grid-template-areas** | 给区域起名：A1 写「header」，B2 写「sidebar」 | §11.3 |
| **fr 单位** | 分蛋糕：1fr 2fr = 你一份我两份 | §11.1 |
| **@media** | 换房间规则：客厅宽摆三沙发，卧室窄摆一张床 | §14 |
| **transition** | 开关灯渐变，不是啪一下亮灭 | §16 |
| **@keyframes** | 翻页动画脚本：第 0 帧在哪、第 100 帧在哪 | §18 |

**为什么重要**：[Vue 09](../Vue/09-Element-Plus与UI工程化.md) 的 `el-row`/`el-col`、`el-container` 后台布局、商品卡片列表，底层全是 Flex/Grid 思维；不懂本章，只会调组件 props 不会改布局。

---

### 0.2 你需要提前知道什么

| 水平 | 建议 |
|------|------|
| 刚学完 04 章 | 从 §3 开启 Flex 开始，每改一个属性就看 DevTools Flex overlay |
| 只会 float 分栏 | 用 §19 圣杯对比 §19.4 Flex 五行替代，感受代码量差异 |
| 已会一点 Bootstrap | 重点 `min-width: 0`、`auto-fit` vs `auto-fill`、移动优先 §14.6 |
| 目标 Vue 路线 | 本章 + [06 JS 基础](./06-JavaScript基础语法与数据类型.md) 后开 Vue 01；09 章 AdminLayout 直接对应 §11.3 Grid areas |

---

### 0.3 本章知识地图（☐→☑）

- [ ] 说清主轴 / 交叉轴，`justify-*` 管主轴，`align-*` 管交叉轴
- [ ] 会写 `display:flex` + `justify-content` + `align-items` 水平垂直居中
- [ ] 会用 `flex: 1`、`flex: 0 0 240px` 做侧栏 + 自适应主区
- [ ] 会用 `grid-template-areas` 做四区域页面
- [ ] 会用 `repeat(auto-fit, minmax())` 做响应式卡片/图片墙
- [ ] 采用移动优先写至少两个 `@media (min-width: …)` 断点
- [ ] 会写 transition hover 与 `@keyframes` 入场动画
- [ ] 知道动画优先 `transform` + `opacity`（§33）
- [ ] 会用 DevTools Flex/Grid overlay 调试对齐
- [ ] 闭卷自测 ≥ 8/10

---

### 0.4 建议学习时长

| 阶段 | 时间 |
|------|------|
| Flex 容器 + 子项 §3～§8 | 3 小时 |
| Grid 区域 + 图片墙 §9～§13 | 2.5 小时 |
| 响应式 + 移动优先 §14～§15 | 2 小时 |
| transition + animation §16～§18 | 1.5 小时 |
| 综合实战 §21 + DevTools + 自测 | 2 小时 |

---

### 0.5 可验证成果

1. 不查文档写出 Flex 水平垂直居中三行核心 CSS。
2. 用 Grid `grid-template-areas` 搭 header / sidebar / main / footer，768px 以下变单列。
3. 产品卡片列表窄屏单列、宽屏三列，无横向滚动条。
4. 卡片 hover 有 `transform` + `box-shadow` 过渡，入场用 `animation-delay` 依次出现。

---

### 0.6 核心术语三件套

**术语（Flex 弹性盒）**：一维布局：父容器沿主轴排列子项，可伸缩、对齐、换行。
**生活类比**：弹簧排排站——队长一声令下，大家靠左、居中或按比例拉长占满空位。
**为什么重要**：导航栏、工具栏、表单行、Vue `el-row` 的本质；比 float 直观十倍。
**本章用到的地方**：§3～§8；§20 速查表。

**术语（Grid 网格）**：二维布局：同时定义行和列，子项放进格子或跨格。
**生活类比**：Excel 表格——先画表结构，再往 A1、B2 填数据；`grid-area` 像合并单元格命名。
**为什么重要**：整页骨架、仪表盘、图片墙；Element `el-container` 可理解为 Flex，复杂分区用 Grid 更清晰。
**本章用到的地方**：§9～§13；§21 实战。

**术语（移动优先 Mobile First）**：默认写手机样式，用 `min-width` 媒体查询逐步增强大屏。
**生活类比**：先保证小背包能装下，再换大箱子加东西；而不是先装满大箱再硬塞小包。
**为什么重要**：与 Tailwind / Bootstrap 断点一致；Vue 09 §24 响应式侧栏折叠同思路。
**本章用到的地方**：§14.6；§15 产品页案例。

---

## 1. 为什么这一份非常关键

如果说前一份（04-盒模型、浮动、定位）解决的是：

- 盒子怎么长
- 元素怎么摆

那么这一份解决的是：

- 现代页面怎么布局
- 手机和电脑怎么适配
- 动效怎么做

真正写页面时，这一份的使用频率会非常高。Flex 和 Grid 几乎出现在每一个现代页面里；响应式是移动端时代的必备技能；过渡与动画则让页面从「能用」变成「好用、好看」。

### 1.1 学完本章你应该能做到

- 独立用 Flex 完成导航栏、卡片列表、居中、表单布局
- 用 Grid 完成图片墙、后台仪表盘、复杂页面分区
- 采用移动优先策略写响应式页面
- 写出平滑的 hover 过渡和简单入场动画
- 理解圣杯布局/双飞翼布局的历史思路，并知道现代替代方案

---

## 2. 现代布局为什么更常用 Flex 和 Grid

过去很多页面靠：

- `float` 浮动
- `inline-block`
- 甚至 `table` 表格布局
- 大量 `position: absolute` 硬推

但现代开发更常用：

- **Flex** — 一维布局（一行或一列）
- **Grid** — 二维布局（行 + 列同时控制）

原因：

| 对比项 | 传统方案 | Flex / Grid |
|--------|----------|-------------|
| 对齐方式 | 靠 margin 推算 | 属性直接控制 |
| 等高列 | 很难 | Flex / Grid 天然支持 |
| 响应式 | 改 margin 很痛苦 | `flex-wrap` / `grid-template` 一行搞定 |
| 可读性 | 布局意图隐藏在细节里 | 声明式，意图清晰 |
| 维护性 | 改一处崩一片 | 改容器属性即可 |

**重要认知**：Float 和定位并没有「淘汰」，它们仍有特定用途（文字环绕、弹层、固定导航），但**页面主结构**应优先 Flex / Grid。

---

## 3. Flex 是什么

Flex（Flexible Box，弹性盒）是一维布局系统。

你可以把它理解为：

- 父容器（Flex Container）和子项（Flex Item）之间的布局协议
- 更方便地控制一行或一列里子元素的排列、对齐、伸缩、换行

### 3.1 一维 vs 二维

```
Flex（一维）                    Grid（二维）
┌───┬───┬───┐                  ┌───┬───┬───┐
│ A │ B │ C │  ← 主轴方向       │ A │ B │ C │
└───┴───┴───┘                  ├───┼───┼───┤
                               │ D │ E │ F │
                               └───┴───┴───┘
```

---

## 4. 开启 Flex

```css
.container {
  display: flex; /* 块级 Flex 容器 */
}

/* 或 */
.container {
  display: inline-flex; /* 行内 Flex 容器，容器本身不占满一行 */
}
```

当父元素设为 `flex` 后，**直接子元素**会进入 Flex 布局上下文，成为 Flex Item。

```html
<div class="container">
  <div class="item">1</div>  <!-- Flex Item -->
  <div class="item">2</div>  <!-- Flex Item -->
  <span>不是 Flex Item</span> <!-- 孙元素不受影响 -->
</div>
```

---

## 5. 主轴和交叉轴（必背概念）

学习 Flex 一定要理解这两个概念，否则后面所有属性都会混淆。

### 5.1 默认情况（flex-direction: row）

```
交叉轴（Cross Axis）↓
                    ┌─────────────────────────────────────┐
                    │  Item1    Item2    Item3             │
                    └─────────────────────────────────────┘
                    ←────────── 主轴（Main Axis）──────────→
```

- **主轴（Main Axis）**：子元素主要排列的方向
- **交叉轴（Cross Axis）**：与主轴垂直的方向

### 5.2 flex-direction: column 时

```
主轴 ↓
     ┌──────┐
     │Item1 │
     ├──────┤
     │Item2 │  ← 交叉轴 →
     ├──────┤
     │Item3 │
     └──────┘
```

**记忆口诀**：

- `justify-*` 管主轴
- `align-*` 管交叉轴
- 轴的方向由 `flex-direction` 决定，不是固定的「水平/垂直」

---

## 6. 容器属性详解（Flex Container）

以下属性写在**父元素**上。

### 6.1 `flex-direction` — 主轴方向

```css
.container {
  flex-direction: row; /* 默认值：从左到右 */
}
```

| 值 | 效果 | 图示（→ 为排列顺序） |
|----|------|----------------------|
| `row` | 水平，起点在左 | `[1][2][3] →` |
| `row-reverse` | 水平，起点在右 | `← [3][2][1]` |
| `column` | 垂直，起点在上 | `[1]` ↓ `[2]` ↓ `[3]` |
| `column-reverse` | 垂直，起点在下 | `[3]` ↑ `[2]` ↑ `[1]` |

```css
/* 移动端竖排导航示例 */
.nav {
  display: flex;
  flex-direction: column;
}

@media (min-width: 768px) {
  .nav {
    flex-direction: row;
  }
}
```

---

### 6.2 `flex-wrap` — 是否换行

```css
.container {
  flex-wrap: nowrap; /* 默认：不换行，子项会被压缩 */
  flex-wrap: wrap;   /* 换行，多行排列 */
  flex-wrap: wrap-reverse; /* 换行，新行在上方 */
}
```

**图示（wrap）**：

```
容器宽度不够时：

nowrap（默认，可能溢出或压缩）:
┌────────────────────────┐
│ [1][2][3][4][5]溢出→   │
└────────────────────────┘

wrap（自动换行）:
┌────────────────────────┐
│ [1] [2] [3]            │
│ [4] [5]                │
└────────────────────────┘
```

**简写** `flex-flow`：

```css
flex-flow: row wrap; /* = flex-direction + flex-wrap */
```

---

### 6.3 `justify-content` — 主轴对齐

```css
.container {
  justify-content: flex-start; /* 默认：起点对齐 */
}
```

| 值 | 效果图示（□ = 子项，空白 = 剩余空间） |
|----|--------------------------------------|
| `flex-start` | `[□][□][□]········` |
| `flex-end` | `········[□][□][□]` |
| `center` | `····[□][□][□]····` |
| `space-between` | `[□]····[□]····[□]` |
| `space-around` | `·[□]··[□]··[□]·` |
| `space-evenly` | `··[□]··[□]··[□]··` |

```css
/* 导航栏：logo 左，菜单右 */
.navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
```

---

### 6.4 `align-items` — 交叉轴对齐（单行）

```css
.container {
  align-items: stretch; /* 默认：拉伸填满交叉轴 */
}
```

| 值 | 效果（假设容器高度 100px） |
|----|---------------------------|
| `stretch` | 子项高度被拉伸至 100px |
| `flex-start` | 子项靠交叉轴起点（顶部） |
| `flex-end` | 子项靠交叉轴终点（底部） |
| `center` | 子项在交叉轴居中 |
| `baseline` | 按文字基线对齐 |

```
align-items: center 示意：

┌────────────────────────── 100px ──────────────────────────┐
│                                                           │
│              [ Item 高度 40px ]                           │
│                                                           │
└───────────────────────────────────────────────────────────┘
```

---

### 6.5 `align-content` — 交叉轴对齐（多行时）

仅当 `flex-wrap: wrap` 且存在**多行**时生效。控制的是**行与行之间**在交叉轴上的分布。

| 值 | 多行分布示意 |
|----|-------------|
| `flex-start` | 所有行挤在顶部 |
| `flex-end` | 所有行挤在底部 |
| `center` | 所有行整体居中 |
| `space-between` | 第一行顶、最后一行底，中间均分 |
| `space-around` | 每行上下都有间距 |
| `stretch` | 行被拉伸填满（默认） |

```
align-content: space-between：

┌─────────────────────┐
│ Row1: [1][2][3]     │  ← 第一行贴顶
│                     │
│                     │  ← 中间空白均分
│                     │
│ Row2: [4][5]        │  ← 最后一行贴底
└─────────────────────┘
```

---

### 6.6 `gap` / `row-gap` / `column-gap` — 间距

```css
.container {
  gap: 16px;              /* 行间距 + 列间距都是 16px */
  gap: 12px 24px;         /* row-gap column-gap */
  row-gap: 12px;
  column-gap: 24px;
}
```

```
gap: 16px 示意：

┌──────┐ 16px ┌──────┐ 16px ┌──────┐
│  1   │ ←→  │  2   │ ←→  │  3   │
└──────┘      └──────┘      └──────┘
```

比给每个子元素写 `margin-right` 更整洁，且**最后一项不会多出多余外边距**。

---

## 7. 子项属性详解（Flex Item）

以下属性写在**子元素**上。

### 7.1 `flex-grow` — 放大比例

```css
.item {
  flex-grow: 0; /* 默认：不放大 */
  flex-grow: 1; /* 有剩余空间时参与分配 */
}
```

```
容器宽 600px，三个子项各 100px，剩余 300px：

flex-grow: 0 → [100][100][100]  剩余空间闲置

item1: flex-grow: 1
item2: flex-grow: 1
item3: flex-grow: 2
→ 按 1:1:2 分配 300px → [150][150][200]
```

---

### 7.2 `flex-shrink` — 缩小比例

```css
.item {
  flex-shrink: 1; /* 默认：空间不足时等比缩小 */
  flex-shrink: 0; /* 不缩小，可能溢出 */
}
```

当容器宽度小于子项总宽度时，`flex-shrink > 0` 的子项会被压缩。

---

### 7.3 `flex-basis` — 初始主轴尺寸

```css
.item {
  flex-basis: auto;  /* 默认：由 width/height 或内容决定 */
  flex-basis: 200px; /* 初始占 200px，再 grow/shrink */
  flex-basis: 0;     /* 忽略内容宽度，纯按比例分配 */
}
```

---

### 7.4 `flex` — 简写（最常用）

```css
.item {
  flex: 1;           /* 最常见：flex: 1 1 0% 的简写（浏览器实现略有差异） */
  flex: auto;        /* flex: 1 1 auto */
  flex: none;        /* flex: 0 0 auto，不伸缩 */
  flex: 2 1 200px;   /* grow shrink basis */
}
```

| 写法 | 含义 | 典型场景 |
|------|------|----------|
| `flex: 1` | 等分剩余空间 | 三列等宽布局 |
| `flex: 0 0 200px` | 固定 200px 不伸缩 | 侧边栏固定宽度 |
| `flex: none` | 完全不伸缩 | 图标、按钮固定尺寸 |

```css
/* 经典：左侧固定 240px，右侧自适应 */
.layout {
  display: flex;
}
.sidebar {
  flex: 0 0 240px;
}
.main {
  flex: 1;
}
```

---

### 7.5 `align-self` — 单个子项覆盖对齐

```css
.item-special {
  align-self: flex-end; /* 仅此一项靠底部，不影响兄弟 */
}
```

可选值与 `align-items` 相同：`auto | flex-start | flex-end | center | baseline | stretch`

---

### 7.6 `order` — 视觉顺序

```css
.item-a { order: 2; }
.item-b { order: 1; } /* B 会显示在 A 前面 */
```

- 默认 `order: 0`
- 数值越小越靠前
- **注意**：只改变视觉顺序，不影响 DOM 顺序和 Tab 焦点顺序，无障碍场景慎用

---

## 8. Flex 完整小案例：响应式卡片列表

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Flex 卡片列表示例</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }

    .card-list {
      display: flex;
      flex-wrap: wrap;
      gap: 20px;
      padding: 20px;
      max-width: 1200px;
      margin: 0 auto;
    }

    .card {
      flex: 1 1 280px; /* 最小 280px，可放大可缩小，自动换行 */
      background: #fff;
      border-radius: 8px;
      box-shadow: 0 2px 8px rgba(0,0,0,.08);
      padding: 20px;
    }

    .card h3 { margin-bottom: 8px; }
    .card p { color: #666; line-height: 1.6; }
  </style>
</head>
<body>
  <div class="card-list">
    <article class="card"><h3>卡片 1</h3><p>flex: 1 1 280px 让卡片自动换行。</p></article>
    <article class="card"><h3>卡片 2</h3><p>窗口变窄时自动变为单列。</p></article>
    <article class="card"><h3>卡片 3</h3><p>无需写媒体查询也能自适应。</p></article>
  </div>
</body>
</html>
```

### 8.1 主示例逐行读（Flex 卡片列表）

| 行号/选择器 | 含义 | 改错会怎样 |
|-------------|------|------------|
| `.card-list { display: flex }` | 父级成为 Flex 容器，子 `article` 变 flex item | 不写则 `flex: 1 1 280px` 完全无效 |
| `flex-wrap: wrap` | 一行放不下时换行 | 默认 nowrap 时卡片被压扁或溢出横向滚动 |
| `gap: 20px` | 卡片间距，首尾无多余外边距 | 改用 `margin-right` 最后一列常对不齐 |
| `.card { flex: 1 1 280px }` | 最小 280px，可伸可缩，自动参与换行 | 只写 `width:280px` 窄屏不会自动变单列 |
| `max-width: 1200px; margin: 0 auto` | 列表居中且不超过 1200px | 去掉 max-width 超宽屏卡片被拉得过长 |

---

## 9. Grid 是什么

Grid 是**二维**布局系统：同时控制行和列。

如果 Flex 更适合：

- 导航栏（一行）
- 表单行（一行多个字段）
- 工具栏

那么 Grid 更适合：

- 整页布局（头/侧/主/脚）
- 图片墙（多行多列）
- 仪表盘（不规则区域）

---

## 10. 开启 Grid

```css
.grid {
  display: grid;
}

/* 或 */
.grid {
  display: inline-grid;
}
```

---

## 11. Grid 容器属性详解

### 11.1 `grid-template-columns` — 列定义

```css
.grid {
  grid-template-columns: 200px 1fr 1fr;
  /* 第1列固定200px，第2、3列均分剩余 */
}
```

**`fr` 单位**：fraction，剩余空间的比例份额。

```css
grid-template-columns: 1fr 2fr;
/* 第2列宽度是第1列的2倍 */
```

**`repeat()` 函数**：

```css
grid-template-columns: repeat(3, 1fr);        /* 三列等宽 */
grid-template-columns: repeat(4, 100px);      /* 四列各 100px */
grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
/* 自动填充：每列最小 200px，能放几列放几列 */
```

**`minmax()` 函数**：

```css
grid-template-columns: repeat(3, minmax(150px, 1fr));
/* 每列最小 150px，最大 1fr */
```

---

### 11.2 `grid-template-rows` — 行定义

```css
.grid {
  grid-template-rows: 80px auto 60px;
  /* 头 80px，中间自适应，脚 60px */
}
```

---

### 11.3 `grid-template-areas` — 区域命名（重点）

通过给每个子项命名，用字符串直观描述布局。

```css
.page {
  display: grid;
  grid-template-columns: 240px 1fr;
  grid-template-rows: 60px 1fr 50px;
  grid-template-areas:
    "header header"
    "sidebar main"
    "footer footer";
  min-height: 100vh;
  gap: 0;
}

.header  { grid-area: header;  background: #333; color: #fff; }
.sidebar { grid-area: sidebar; background: #f5f5f5; }
.main    { grid-area: main;    background: #fff; }
.footer  { grid-area: footer;  background: #eee; }
```

**图示**：

```
┌─────────────────────────────────┐
│            header               │  60px
├──────────┬──────────────────────┤
│ sidebar  │        main          │  1fr
│  240px   │                      │
├──────────┴──────────────────────┤
│            footer               │  50px
└─────────────────────────────────┘
```

**区域命名规则**：

- 每个字符串代表一行，每个「单词」代表一列
- 相同名字合并为同一区域
- 用 `.` 表示空单元格

```css
/* 中间留空的示例 */
grid-template-areas:
  "logo nav nav"
  ". content aside"
  "footer footer footer";
```

**响应式改造**：

```css
@media (max-width: 768px) {
  .page {
    grid-template-columns: 1fr;
    grid-template-areas:
      "header"
      "main"
      "sidebar"
      "footer";
  }
}
```

---

### 11.4 `grid-template` — 简写

```css
grid-template:
  "head head" 60px
  "nav  main" 1fr
  / 200px 1fr;
/* areas + rows / columns */
```

初学阶段建议拆开写，可读性更好。

---

### 11.5 `gap` / `row-gap` / `column-gap`

与 Flex 相同，Grid 也支持 gap。

```css
.grid {
  gap: 20px;
  gap: 16px 24px; /* 行 列 */
}
```

---

### 11.6 `justify-items` / `align-items` — 单元格内对齐

```css
.grid {
  justify-items: center; /* 单元格内水平居中 */
  align-items: center;   /* 单元格内垂直居中 */
}
```

---

### 11.7 `justify-content` / `align-content` — 网格整体对齐

当网格总尺寸小于容器时，控制整个网格在容器中的位置。

```css
.grid {
  justify-content: center;
  align-content: center;
}
```

---

## 12. Grid 子项属性详解

### 12.1 `grid-column` / `grid-row` — 跨越行列

```css
.item {
  grid-column: 1 / 3;  /* 从第1条列线到第3条列线，跨2列 */
  grid-row: 1 / 2;     /* 占第1行 */
}

/* 简写 */
grid-column: span 2; /* 跨 2 列 */
grid-row: span 3;    /* 跨 3 行 */
```

**Grid 线编号**（从 1 开始）：

```
列线:  1    2    3    4
      ┌────┬────┬────┐
      │ A  │ B  │ C  │  行线 1
      ├────┼────┼────┤
      │ D  │ E  │ F  │  行线 2
      └────┴────┴────┘
                        行线 3
```

`grid-column: 1 / 3` 表示从列线 1 到列线 3，即占据 A 和 B 两列。

---

### 12.2 `grid-area` — 区域名或行列简写

```css
/* 方式一：使用命名区域 */
.item { grid-area: header; }

/* 方式二：grid-row-start / column-start / row-end / column-end */
.item { grid-area: 1 / 1 / 2 / 3; }
/* = row-start / col-start / row-end / col-end */
```

---

### 12.3 `justify-self` / `align-self`

单个网格项在单元格内的对齐，覆盖容器的 `justify-items` / `align-items`。

---

## 13. Grid 图片墙完整示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Grid 图片墙</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }

    .gallery {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
      gap: 12px;
      padding: 20px;
    }

    .gallery img {
      width: 100%;
      aspect-ratio: 1 / 1;
      object-fit: cover;
      border-radius: 8px;
      transition: transform 0.3s ease;
    }

    .gallery img:hover {
      transform: scale(1.05);
    }

    /* 大图占两格 */
    .gallery .featured {
      grid-column: span 2;
      grid-row: span 2;
    }

    .gallery .featured img {
      aspect-ratio: auto;
      height: 100%;
    }
  </style>
</head>
<body>
  <div class="gallery">
    <div class="featured"><img src="https://picsum.photos/400/400?1" alt="精选" /></div>
    <div><img src="https://picsum.photos/200/200?2" alt="" /></div>
    <div><img src="https://picsum.photos/200/200?3" alt="" /></div>
    <div><img src="https://picsum.photos/200/200?4" alt="" /></div>
    <div><img src="https://picsum.photos/200/200?5" alt="" /></div>
    <div><img src="https://picsum.photos/200/200?6" alt="" /></div>
  </div>
</body>
</html>
```

---

## 14. 响应式设计

### 14.1 什么是响应式设计

响应式设计的目标是：

- 让页面在不同屏幕宽度下都尽量好看、好用
- 同一份 HTML，用 CSS 适配多种设备

常见设备：

| 类型 | 大致宽度 | 布局特点 |
|------|----------|----------|
| 小屏手机 | < 480px | 单列、大按钮、汉堡菜单 |
| 大屏手机 | 480–768px | 单列或双列卡片 |
| 平板 | 768–1024px | 双列、可显示部分侧栏 |
| 桌面 | > 1024px | 多列、完整导航 |

---

### 14.2 视口 viewport（必写）

```html
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
```

没有这行，移动端浏览器会把页面当作桌面宽度（通常 980px）渲染后再缩小，导致字体极小。

---

### 14.3 媒体查询 `@media`

```css
/* 最大宽度：屏幕 ≤ 768px 时生效（常用于桌面优先） */
@media (max-width: 768px) {
  .sidebar { display: none; }
}

/* 最小宽度：屏幕 ≥ 768px 时生效（移动优先） */
@media (min-width: 768px) {
  .nav-links { display: flex; }
}
```

**逻辑运算符**：

```css
@media (min-width: 768px) and (max-width: 1024px) { /* 平板区间 */ }
@media (min-width: 768px), (orientation: landscape) { /* 或 */ }
```

---

### 14.4 常见断点参考

没有绝对标准，以下为常见参考：

```css
/* 移动优先常用断点 */
/* 默认：手机 */
/* sm: ≥ 576px  大手机 */
/* md: ≥ 768px  平板 */
/* lg: ≥ 992px  小桌面 */
/* xl: ≥ 1200px 大桌面 */
```

你现在先理解「根据宽度切样式」即可，不必死记数字。

---

### 14.5 响应式单位

| 单位 | 相对谁 | 典型用途 |
|------|--------|----------|
| `%` | 父元素 | 宽度百分比布局 |
| `vw` / `vh` | 视口宽/高 | 全屏区块、大标题 |
| `rem` | 根元素 `font-size` | 字体、间距（推荐） |
| `em` | 当前元素 `font-size` | 组件内部相对尺寸 |
| `fr` | Grid 剩余空间 | Grid 列宽 |

```css
html { font-size: 16px; } /* 1rem = 16px */
.title { font-size: 2rem; } /* 32px */
.section { padding: 1.5rem; } /* 24px */
```

---

### 14.6 桌面优先 vs 移动优先

**桌面优先**（max-width）：

```css
.container { display: flex; }
@media (max-width: 768px) {
  .container { flex-direction: column; }
}
```

**移动优先**（min-width，现代推荐）：

```css
.container {
  display: flex;
  flex-direction: column; /* 默认手机竖排 */
}
@media (min-width: 768px) {
  .container { flex-direction: row; }
}
```

移动优先的优势：

- 默认样式更轻，手机加载更快
- 渐进增强，大屏加功能而非小屏删功能
- 与主流框架（Bootstrap、Tailwind）断点思路一致

---

## 15. 移动优先完整案例：产品展示页

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>移动优先产品页</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

    :root {
      --color-primary: #2563eb;
      --color-text: #1f2937;
      --color-muted: #6b7280;
      --space: 1rem;
      --radius: 8px;
    }

    body {
      font-family: system-ui, sans-serif;
      color: var(--color-text);
      line-height: 1.6;
    }

    /* ===== 手机默认样式 ===== */
    .header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: var(--space);
      border-bottom: 1px solid #e5e7eb;
    }

    .menu-btn {
      background: none;
      border: none;
      font-size: 1.5rem;
      cursor: pointer;
    }

    .nav { display: none; } /* 手机隐藏导航，显示汉堡按钮 */

    .hero {
      padding: 2rem var(--space);
      text-align: center;
      background: linear-gradient(135deg, #eff6ff, #dbeafe);
    }

    .hero h1 { font-size: 1.75rem; margin-bottom: 0.5rem; }
    .hero p { color: var(--color-muted); margin-bottom: 1.5rem; }

    .btn {
      display: inline-block;
      padding: 0.75rem 1.5rem;
      background: var(--color-primary);
      color: #fff;
      border-radius: var(--radius);
      text-decoration: none;
      transition: background 0.2s, transform 0.2s;
    }
    .btn:hover { background: #1d4ed8; transform: translateY(-2px); }

    .products {
      display: grid;
      grid-template-columns: 1fr; /* 手机单列 */
      gap: var(--space);
      padding: var(--space);
      max-width: 1200px;
      margin: 0 auto;
    }

    .product-card {
      border: 1px solid #e5e7eb;
      border-radius: var(--radius);
      overflow: hidden;
      transition: box-shadow 0.3s;
    }
    .product-card:hover { box-shadow: 0 8px 24px rgba(0,0,0,.1); }

    .product-card img { width: 100%; aspect-ratio: 4/3; object-fit: cover; }
    .product-card .info { padding: var(--space); }
    .product-card h3 { font-size: 1.1rem; margin-bottom: 0.25rem; }
    .product-card .price { color: var(--color-primary); font-weight: bold; }

    .footer {
      text-align: center;
      padding: 2rem var(--space);
      background: #f9fafb;
      color: var(--color-muted);
      margin-top: 2rem;
    }

    /* ===== 平板 ≥ 768px ===== */
    @media (min-width: 768px) {
      .menu-btn { display: none; }

      .nav {
        display: flex;
        gap: 1.5rem;
        list-style: none;
      }
      .nav a {
        text-decoration: none;
        color: var(--color-text);
        transition: color 0.2s;
      }
      .nav a:hover { color: var(--color-primary); }

      .hero h1 { font-size: 2.5rem; }

      .products {
        grid-template-columns: repeat(2, 1fr); /* 双列 */
      }
    }

    /* ===== 桌面 ≥ 1024px ===== */
    @media (min-width: 1024px) {
      .hero { padding: 4rem var(--space); }

      .products {
        grid-template-columns: repeat(3, 1fr); /* 三列 */
      }
    }
  </style>
</head>
<body>
  <header class="header">
    <div class="logo">MyShop</div>
    <button class="menu-btn" aria-label="打开菜单">☰</button>
    <ul class="nav">
      <li><a href="#">首页</a></li>
      <li><a href="#">产品</a></li>
      <li><a href="#">关于</a></li>
      <li><a href="#">联系</a></li>
    </ul>
  </header>

  <section class="hero">
    <h1>发现好物，从这里开始</h1>
    <p>移动优先设计，任何设备都有好体验</p>
    <a href="#" class="btn">立即选购</a>
  </section>

  <main class="products">
    <article class="product-card">
      <img src="https://picsum.photos/400/300?p1" alt="产品1" />
      <div class="info"><h3>无线耳机</h3><p class="price">¥299</p></div>
    </article>
    <article class="product-card">
      <img src="https://picsum.photos/400/300?p2" alt="产品2" />
      <div class="info"><h3>机械键盘</h3><p class="price">¥499</p></div>
    </article>
    <article class="product-card">
      <img src="https://picsum.photos/400/300?p3" alt="产品3" />
      <div class="info"><h3>显示器</h3><p class="price">¥1299</p></div>
    </article>
  </main>

  <footer class="footer">
    <p>&copy; 2026 MyShop. 本站仅供学习演示。</p>
  </footer>
</body>
</html>
```

---

## 16. 过渡 transition

过渡用于让**样式变化**更平滑（需要触发条件，如 `:hover`、类名切换）。

### 16.1 语法

```css
transition: property duration timing-function delay;
```

| 部分 | 说明 | 示例 |
|------|------|------|
| property | 要过渡的属性 | `opacity`, `transform`, `all` |
| duration | 持续时间 | `0.3s`, `300ms` |
| timing-function | 速度曲线 | `ease`, `linear`, `ease-in-out` |
| delay | 延迟 | `0.1s` |

```css
button {
  background: #2563eb;
  color: #fff;
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  transition: background 0.3s ease, transform 0.2s ease, box-shadow 0.3s ease;
}

button:hover {
  background: #1d4ed8;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.4);
}
```

### 16.2 可过渡属性

常见可过渡属性：`opacity`、`transform`、`background-color`、`color`、`width`、`height`、`box-shadow`、`border-color` 等。

**不可过渡**：`display`（改用 `opacity` + `visibility` 或 `grid-template-rows` 动画）。

### 16.3 完整卡片 hover 过渡

```css
.card {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0,0,0,.06);
  transition:
    transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1),
    box-shadow 0.3s ease;
}

.card:hover {
  transform: translateY(-8px);
  box-shadow: 0 12px 32px rgba(0,0,0,.12);
}

.card img {
  transition: transform 0.5s ease;
}
.card:hover img {
  transform: scale(1.08);
}
```

---

## 17. 变形 transform

`transform` 在**不影响文档流**的情况下对元素进行位移、缩放、旋转、倾斜。

### 17.1 常用函数

| 函数 | 作用 | 示例 |
|------|------|------|
| `translate(x, y)` | 位移 | `translate(10px, -5px)` |
| `translateX()` / `translateY()` | 单轴位移 | `translateY(-4px)` |
| `scale(x, y)` | 缩放 | `scale(1.05)` |
| `rotate(angle)` | 旋转 | `rotate(45deg)` |
| `skew(x, y)` | 倾斜 | `skew(5deg, 0)` |

### 17.2 变换原点 `transform-origin`

```css
.icon {
  transition: transform 0.3s;
  transform-origin: center center; /* 默认 */
}
.icon:hover {
  transform: rotate(180deg);
}

/* 从左上角缩放 */
.zoom-box {
  transform-origin: top left;
  transform: scale(1.2);
}
```

### 17.3 2D vs 3D

```css
/* 3D 透视（父元素） */
.scene {
  perspective: 800px;
}
.card-3d {
  transform: rotateY(15deg);
  transition: transform 0.5s;
}
.card-3d:hover {
  transform: rotateY(0deg);
}
```

### 17.4 性能提示

优先动画 `transform` 和 `opacity`，它们可由 GPU 合成层处理，不易触发重排（reflow）。

---

## 18. 动画 animation

动画不需要用户交互即可自动播放，或循环播放。

### 18.1 定义关键帧 `@keyframes`

```css
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 等价于 */
@keyframes fadeInUp {
  0%   { opacity: 0; transform: translateY(30px); }
  100% { opacity: 1; transform: translateY(0); }
}
```

### 18.2 应用动画

```css
.hero-title {
  animation: fadeInUp 0.8s ease forwards;
}
```

| 属性 | 说明 |
|------|------|
| `animation-name` | 关键帧名称 |
| `animation-duration` | 持续时间 |
| `animation-timing-function` | 速度曲线 |
| `animation-delay` | 延迟 |
| `animation-iteration-count` | 次数（`infinite` 无限） |
| `animation-direction` | `normal` / `reverse` / `alternate` |
| `animation-fill-mode` | `forwards` 保持结束状态 |
| `animation-play-state` | `running` / `paused` |

**简写**：

```css
animation: fadeInUp 0.8s ease 0.2s 1 forwards;
/* name duration timing delay count fill-mode */
```

### 18.3 完整动画示例集

```css
/* 1. 脉冲提示 */
@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50%      { transform: scale(1.05); }
}
.badge {
  animation: pulse 2s ease-in-out infinite;
}

/* 2. 骨架屏闪烁 */
@keyframes shimmer {
  0%   { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}
.skeleton {
  background: linear-gradient(90deg, #eee 25%, #f5f5f5 50%, #eee 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

/* 3. 旋转加载 */
@keyframes spin {
  to { transform: rotate(360deg); }
}
.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #e5e7eb;
  border-top-color: #2563eb;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

/* 4. 依次入场（配合 animation-delay） */
.item:nth-child(1) { animation: fadeInUp 0.5s ease forwards; animation-delay: 0.1s; opacity: 0; }
.item:nth-child(2) { animation: fadeInUp 0.5s ease forwards; animation-delay: 0.2s; opacity: 0; }
.item:nth-child(3) { animation: fadeInUp 0.5s ease forwards; animation-delay: 0.3s; opacity: 0; }
```

### 18.4 无障碍：尊重用户偏好

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## 19. 圣杯布局与双飞翼布局简介

这是 Flex/Grid 出现之前的经典三栏布局方案，了解有助于读老代码和理解布局演进。

### 19.1 目标结构

```
┌──────────────────────────────────────┐
│              Header                  │
├────────┬─────────────────┬───────────┤
│ Left   │     Main        │   Right   │
│ 200px  │   (自适应)       │   200px   │
├────────┴─────────────────┴───────────┤
│              Footer                  │
└──────────────────────────────────────┘
```

要求：中间主栏先渲染（SEO 友好），三栏等高，左右固定宽。

### 19.2 圣杯布局（Holy Grail Layout）

核心思路：

1. 中间栏 `width: 100%` 先写，左右栏用**负 margin** 拉回来
2. 父容器 `padding` 留出左右栏空间
3. 左右栏 `position: relative` 微调位置

```css
/* 简化示意，现代项目请用 Flex/Grid 替代 */
.container {
  padding: 0 200px; /* 为左右栏留空 */
}
.main { width: 100%; float: left; }
.left {
  width: 200px;
  float: left;
  margin-left: -100%; /* 拉到最左 */
  position: relative;
  right: 200px;
}
.right {
  width: 200px;
  float: left;
  margin-right: -200px;
}
```

### 19.3 双飞翼布局（Double Wing Layout）

与圣杯类似，区别是：

- 圣杯：左中右三个 div 同级，靠 padding + 负 margin
- 双飞翼：main 内再包一层 `.inner`，margin 写在 inner 上，**不需要 relative 定位**

```html
<div class="container">
  <div class="main">
    <div class="main-inner">主内容</div>
  </div>
  <div class="left">左栏</div>
  <div class="right">右栏</div>
</div>
```

### 19.4 现代替代方案（推荐）

```css
/* Flex 三栏 — 5 行搞定 */
.layout {
  display: flex;
}
.left  { flex: 0 0 200px; }
.main  { flex: 1; }
.right { flex: 0 0 200px; }

/* Grid 三栏 — 更清晰 */
.layout {
  display: grid;
  grid-template-columns: 200px 1fr 200px;
}
```

**结论**：理解圣杯/双飞翼的「中间优先 + 两侧固定」思想即可，新项目直接用 Flex 或 Grid。

---

## 20. 常见布局模式速查

### 20.1 水平垂直居中

**方法一：Flex（最推荐）**

```css
.center-box {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
}
```

**方法二：Grid**

```css
.center-box {
  display: grid;
  place-items: center;
  min-height: 100vh;
}
```

**方法三：绝对定位 + transform（已知宽高时）**

```css
.parent { position: relative; }
.child {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}
```

---

### 20.2 等高列

**Flex 默认 stretch**：

```css
.columns {
  display: flex;
  align-items: stretch; /* 默认值 */
}
.col { /* 无需设高度，自动等高 */ }
```

**Grid**：

```css
.columns {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  align-items: stretch;
}
```

---

### 20.3 粘性页脚（Sticky Footer）

页脚始终在页面底部；内容少时在视口底，内容多时在内容后。

**方法一：Flex 列布局（推荐）**

```css
body {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
main { flex: 1; } /* 撑开中间，把 footer 推到底 */
```

**方法二：Grid**

```css
body {
  min-height: 100vh;
  display: grid;
  grid-template-rows: auto 1fr auto;
}
```

```
┌─────────────┐
│   Header    │  auto
├─────────────┤
│             │
│    Main     │  1fr（占据剩余空间）
│             │
├─────────────┤
│   Footer    │  auto
└─────────────┘
```

---

### 20.4 水平均分导航

```css
.nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
```

---

### 20.5 左固定 + 右自适应

```css
.wrapper { display: flex; }
.sidebar { flex: 0 0 260px; }
.content { flex: 1; min-width: 0; } /* min-width: 0 防止 flex 子项溢出 */
```

---

### 20.6 瀑布流（基础 Grid 近似）

```css
.masonry {
  columns: 3 280px;
  column-gap: 16px;
}
.masonry-item {
  break-inside: avoid;
  margin-bottom: 16px;
}
```

真正的 Masonry 可用 `grid-template-rows: masonry`（Firefox 支持）或 JS 库。

---

## 21. 完整实战：响应式导航 + 三列卡片 + 图片墙

以下是一个可运行的综合练习页面，整合本章所有知识点。

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>综合实战：导航 + 卡片 + 图片墙</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

    :root {
      --primary: #6366f1;
      --bg: #f8fafc;
      --text: #0f172a;
      --muted: #64748b;
      --radius: 10px;
      --shadow: 0 4px 16px rgba(0,0,0,.08);
    }

    body {
      font-family: "Segoe UI", system-ui, sans-serif;
      background: var(--bg);
      color: var(--text);
      line-height: 1.6;
      min-height: 100vh;
      display: flex;
      flex-direction: column;
    }

    /* ===== 粘性页脚结构 ===== */
    main { flex: 1; }

    /* ===== 导航栏 ===== */
    .navbar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 1rem 1.5rem;
      background: #fff;
      box-shadow: 0 1px 4px rgba(0,0,0,.06);
      position: sticky;
      top: 0;
      z-index: 100;
    }

    .logo {
      font-size: 1.25rem;
      font-weight: 700;
      color: var(--primary);
    }

    .nav-toggle {
      display: block;
      background: none;
      border: none;
      font-size: 1.5rem;
      cursor: pointer;
    }

    .nav-menu {
      display: none;
      list-style: none;
      flex-direction: column;
      gap: 0.5rem;
      position: absolute;
      top: 100%;
      left: 0;
      right: 0;
      background: #fff;
      padding: 1rem;
      box-shadow: var(--shadow);
    }

    .nav-menu.is-open { display: flex; }

    .nav-menu a {
      text-decoration: none;
      color: var(--text);
      padding: 0.5rem 1rem;
      border-radius: 6px;
      transition: background 0.2s, color 0.2s;
    }
    .nav-menu a:hover { background: #eef2ff; color: var(--primary); }

    @media (min-width: 768px) {
      .nav-toggle { display: none; }
      .nav-menu {
        display: flex;
        flex-direction: row;
        position: static;
        box-shadow: none;
        padding: 0;
        gap: 0.25rem;
      }
    }

    /* ===== 页面区块 ===== */
    .section {
      max-width: 1200px;
      margin: 0 auto;
      padding: 2rem 1.5rem;
    }

    .section-title {
      font-size: 1.5rem;
      margin-bottom: 1.5rem;
      animation: fadeIn 0.6s ease;
    }

    @keyframes fadeIn {
      from { opacity: 0; transform: translateY(10px); }
      to   { opacity: 1; transform: translateY(0); }
    }

    /* ===== 三列卡片（移动优先） ===== */
    .cards {
      display: grid;
      grid-template-columns: 1fr;
      gap: 1.25rem;
    }

    @media (min-width: 640px) {
      .cards { grid-template-columns: repeat(2, 1fr); }
    }
    @media (min-width: 960px) {
      .cards { grid-template-columns: repeat(3, 1fr); }
    }

    .card {
      background: #fff;
      border-radius: var(--radius);
      padding: 1.5rem;
      box-shadow: var(--shadow);
      transition: transform 0.3s ease, box-shadow 0.3s ease;
    }
    .card:hover {
      transform: translateY(-6px);
      box-shadow: 0 12px 28px rgba(99, 102, 241, 0.15);
    }

    .card-icon {
      width: 48px;
      height: 48px;
      background: #eef2ff;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 1.5rem;
      margin-bottom: 1rem;
    }

    .card h3 { margin-bottom: 0.5rem; }
    .card p { color: var(--muted); font-size: 0.95rem; }

    /* ===== 图片墙 ===== */
    .gallery {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
      gap: 10px;
    }

    .gallery-item {
      overflow: hidden;
      border-radius: var(--radius);
      aspect-ratio: 1;
    }

    .gallery-item img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      transition: transform 0.4s ease;
    }
    .gallery-item:hover img { transform: scale(1.1); }

    .gallery-item.wide {
      grid-column: span 2;
      aspect-ratio: 2 / 1;
    }

    /* ===== 页脚 ===== */
    footer {
      text-align: center;
      padding: 1.5rem;
      color: var(--muted);
      background: #fff;
      border-top: 1px solid #e2e8f0;
    }
  </style>
</head>
<body>
  <nav class="navbar">
    <div class="logo">LayoutLab</div>
    <button class="nav-toggle" id="navToggle" aria-label="切换菜单">☰</button>
    <ul class="nav-menu" id="navMenu">
      <li><a href="#cards">卡片</a></li>
      <li><a href="#gallery">图片墙</a></li>
      <li><a href="#">关于</a></li>
      <li><a href="#">联系</a></li>
    </ul>
  </nav>

  <main>
    <section class="section" id="cards">
      <h2 class="section-title">三列特性卡片</h2>
      <div class="cards">
        <article class="card">
          <div class="card-icon">⚡</div>
          <h3>Flex 弹性布局</h3>
          <p>一维排列，轻松实现导航、居中、等分。</p>
        </article>
        <article class="card">
          <div class="card-icon">📐</div>
          <h3>Grid 网格布局</h3>
          <p>二维控制，区域命名，图片墙与仪表盘首选。</p>
        </article>
        <article class="card">
          <div class="card-icon">📱</div>
          <h3>响应式适配</h3>
          <p>移动优先，一套 HTML 适配手机到桌面。</p>
        </article>
      </div>
    </section>

    <section class="section" id="gallery">
      <h2 class="section-title">Grid 图片墙</h2>
      <div class="gallery">
        <div class="gallery-item wide"><img src="https://picsum.photos/400/200?g1" alt="" /></div>
        <div class="gallery-item"><img src="https://picsum.photos/200/200?g2" alt="" /></div>
        <div class="gallery-item"><img src="https://picsum.photos/200/200?g3" alt="" /></div>
        <div class="gallery-item"><img src="https://picsum.photos/200/200?g4" alt="" /></div>
        <div class="gallery-item"><img src="https://photos/200/200?g5" alt="" onerror="this.src='https://picsum.photos/200/200?g5'" /></div>
        <div class="gallery-item"><img src="https://picsum.photos/200/200?g6" alt="" /></div>
      </div>
    </section>
  </main>

  <footer>
    <p>综合实战示例 &copy; 2026 — 仅供学习</p>
  </footer>

  <script>
    document.getElementById('navToggle').addEventListener('click', function () {
      document.getElementById('navMenu').classList.toggle('is-open');
    });
  </script>
</body>
</html>
```

---

## 22. 分级练习

### 22.1 入门级（建议 1–2 天）

1. **Flex 居中**：创建一个 300×300 的盒子，用 Flex 把文字水平垂直居中
2. **导航栏**：Logo 左对齐，三个链接右对齐，用 `space-between`
3. **等分三列**：三个 div 用 `flex: 1` 等宽排列
4. **transition 按钮**：hover 时背景色和圆角平滑变化

**自检标准**：不查文档写出 `display: flex; justify-content: center; align-items: center;`

---

### 22.2 进阶级（建议 3–5 天）

1. **响应式卡片列表**：`flex: 1 1 280px` + `flex-wrap: wrap`，窄屏自动单列
2. **Grid 仪表盘**：用 `grid-template-areas` 做 header / sidebar / main / footer 布局
3. **移动优先产品页**：默认单列，768px 双列，1024px 三列
4. **图片墙**：`repeat(auto-fill, minmax(180px, 1fr))`，大图 `grid-column: span 2`
5. **粘性页脚**：内容不足一屏时 footer 贴底

**自检标准**：能在 DevTools 里切换设备宽度，布局不崩、无横向滚动条

---

### 22.3 挑战级（建议 1 周）

1. **完整落地页**：导航（含移动端汉堡菜单）+ Hero + 三列特性 + 图片墙 + 页脚
2. **纯 CSS 入场动画**：页面加载时标题和卡片依次 fadeInUp（`animation-delay`）
3. **Grid + Flex 混用**：Grid 管页面大结构，Flex 管组件内部（如卡片内 icon + 文字）
4. **还原一个真实网站首页**（如 GitHub 首页简化版）的布局结构
5. **把圣杯布局用 Flex 重写**，并对比代码量

**自检标准**：HTML 语义化、CSS 有变量、移动端可用、动画不卡顿、代码可读

---

## 23. FAQ 常见问题

### Q1：Flex 和 Grid 什么时候用哪个？

**A**：一维排列（一行或一列）用 Flex；二维区域（同时管行和列）用 Grid。也可以混用：Grid 划分页面大区域，Flex 处理组件内部。没有绝对界限，选让你代码最简洁的方案。

---

### Q2：`justify-content` 和 `align-items` 老是搞混怎么办？

**A**：先确定主轴方向（看 `flex-direction`），然后 `justify-*` 永远管主轴，`align-*` 永远管交叉轴。默认 row 时：justify = 水平，align = 垂直。

---

### Q3：`flex: 1` 到底是什么意思？

**A**：简写，等价于 `flex: 1 1 0%`（多数浏览器）。含义：可以放大（grow: 1）、可以缩小（shrink: 1）、初始基准为 0（basis: 0），即忽略内容宽度、纯按比例分剩余空间。

---

### Q4：为什么 Flex 子项内容溢出容器？

**A**：常见原因：子项缺少 `min-width: 0` 或 `overflow: hidden`。Flex 子项默认 `min-width: auto`，不会缩小到比内容更窄。解决：

```css
.flex-child {
  min-width: 0; /* 或 overflow: hidden; */
}
```

---

### Q5：`fr` 和 `%` 有什么区别？

**A**：`%` 相对父容器总宽度，嵌套时可能计算复杂；`fr` 相对 Grid 容器**剩余空间**的比例，只用于 Grid（和 Flex 的 grow 类似）。`1fr 1fr 1fr` 比 `33.33% 33.33% 33.33%` 更简洁且不受 gap 影响。

---

### Q6：媒体查询用 max-width 还是 min-width？

**A**：现代推荐**移动优先**（min-width）：默认写手机样式，逐步增强大屏。老项目可能是桌面优先（max-width），读代码时注意断点方向。

---

### Q7：transition 写了但没效果？

**A**：检查：① 是否有触发（`:hover`、类名变化）；② 属性是否可过渡（`display` 不行）；③ 是否写了 `transition` 在**初始状态**而非 hover 上；④ 持续时间是否为 0。

---

### Q8：动画太多导致卡顿怎么办？

**A**：只动画 `transform` 和 `opacity`；减少同时运行的 infinite 动画；用 `will-change: transform`  sparingly；尊重 `prefers-reduced-motion`。

---

### Q9：Grid 的 `grid-template-areas` 名字可以乱取吗？

**A**：可以，但要有语义（如 `header`、`sidebar`）。注意：区域必须是**矩形**，不能出现 L 形：

```css
/* 非法：L 形区域 */
grid-template-areas:
  "a a"
  "a b";  /* 'a' 不是矩形，无效 */
```

---

### Q10：还需不需要学 float？

**A**：需要知道 float 的存在和基本行为（文字环绕图片），但页面主布局不要再用 float。读老项目、维护 legacy 代码时会遇到。

---

### Q11：`place-items: center` 和 Flex 居中有什么区别？

**A**：`place-items: center` 是 Grid 简写（`align-items` + `justify-items`），在**单元格内**居中；Flex 的 `justify-content` + `align-items` 在**主轴/交叉轴**上居中子项。二者都能做全屏居中：Grid 用 `display:grid; place-items:center; min-height:100vh`，Flex 用 `display:flex; justify-content:center; align-items:center`。选你记得住的即可。

---

### Q12：`gap` 和给子元素写 `margin` 有什么优势？

**A**：`gap` 只加在**项目之间**，首尾不会多出外侧空白；`margin` 最后一项右侧/底侧常要多写负 margin 或 `:last-child` 清零。Flex 和 Grid 都推荐用 `gap` 控制间距（Vue Element 的 `el-row :gutter` 类似思路）。

---

### Q13：为什么动画要优先 `transform` 而不是 `top`/`left`？

**A**：改 `top`/`left`/`width` 会触发**重排（layout）**，浏览器要重新算整页几何，易掉帧；`transform` 和 `opacity` 多在**合成层**完成，GPU 友好。见 §33 性能表。

---

### Q14：Element Plus 的 `el-row` / `el-col` 和本章什么关系？

**A**：`el-row` 本质是 `display:flex` 的封装，`el-col` 的 `:span` 控制子项占主轴比例（类似 `flex` 或百分比宽）。学完本章 Flex，读 [Vue 09](../Vue/09-Element-Plus与UI工程化.md) 布局代码就是在调「弹簧排排站」的参数，而不是魔法。

---

## 24. 初学者常见错误

### 24.1 所有布局都靠 margin 硬推

```css
/* 不推荐 */
.item { margin-left: 200px; margin-top: 50px; }

/* 推荐 */
.container { display: flex; gap: 16px; }
```

维护会很差，改一个元素全部要重算。

---

### 24.2 页面只在自己电脑上看着正常

没有做响应式适配。务必用 DevTools 设备模拟器测试 375px、768px、1280px。

---

### 24.3 一上来就用定位做所有布局

`position: absolute` 脱离文档流，不适合大部分正常页面结构。定位适合：弹窗、下拉菜单、角标、固定导航。

---

### 24.4 动画过多过花

每个元素都在动会让用户疲劳。动效应**有目的**：引导注意力、反馈交互、表达层级。

---

### 24.5 Flex 容器忘了设 `display: flex`

子元素的 `flex: 1` 等属性不会生效。

---

### 24.6 Grid 区域命名后忘记给子项设 `grid-area`

```css
/* 定义了 areas 但子项没对应 */
.header { /* 缺少 grid-area: header; */ }
```

---

## 25. Flex vs Grid 速查对照表

| 场景 | 推荐 | 关键代码 |
|------|------|----------|
| 水平垂直居中 | Flex / Grid | `place-items: center` |
| 导航栏 | Flex | `justify-content: space-between` |
| 等分多列 | Flex / Grid | `flex: 1` 或 `repeat(n, 1fr)` |
| 固定侧栏 + 自适应主区 | Flex / Grid | `flex: 0 0 240px` + `flex: 1` |
| 整页布局 | Grid | `grid-template-areas` |
| 图片墙 | Grid | `auto-fill, minmax()` |
| 卡片自动换行 | Flex | `flex-wrap: wrap; flex: 1 1 280px` |
| 粘性页脚 | Flex / Grid | `flex: 1` on main |
| 表单行内字段 | Flex | `display: flex; gap: 12px` |

---

## 26. 学完标准

如果你能做到这些，这一份就掌握得不错：

- [ ] 不查文档写出 Flex 水平垂直居中
- [ ] 解释 `justify-content` 和 `align-items` 的区别
- [ ] 用 `flex: 1 1 280px` 做自适应卡片列表
- [ ] 用 `grid-template-areas` 做四区域页面布局
- [ ] 用 `repeat(auto-fill, minmax())` 做图片墙
- [ ] 采用移动优先写至少两个断点的响应式页面
- [ ] 写 transition hover 效果和 `@keyframes` 入场动画
- [ ] 知道圣杯布局的思路，并能用 Flex 替代
- [ ] 完成综合实战：导航 + 卡片 + 图片墙

---

## 28. Flex 子项等比例缩放组合示例

### 场景一：固定 + 自适应 + 固定（三栏不等宽）

```css
.container { display: flex; }
.sidebar-left  { flex: 0 0 200px; } /* 永远 200px */
.main-content  { flex: 1; min-width: 0; } /* 拿走所有剩余空间 */
.sidebar-right { flex: 0 0 300px; } /* 永远 300px */
```

### 场景二：部分固定 + 部分等分

```css
.container { display: flex; }
.icon { flex: 0 0 40px; }    /* 图标固定 */
.text { flex: 1; }           /* 文字自适应 */
.badge { flex: 0 0 auto; }   /* 徽章根据内容 */
/* flex: 0 0 auto = 不放大/不缩小/看内容宽度 */
```

### 场景三：所有子项等分但有一个要更大

```css
.container { display: flex; }
.item { flex: 1; }          /* 等分 */
.item.featured { flex: 2; }  /* 这个占 2 份，其他各占 1 份 */
/* 3 个普通 + 1 个 featured → 总份数 = 1+1+1+2 = 5 → featured 占 2/5 */
```

---

## 29. Grid `auto-fit` vs `auto-fill` 对比

这两个极易混淆，但区别很关键：

```html
<style>
  .demo {
    display: grid;
    gap: 12px;
    margin-bottom: 20px;
  }
  .demo > div {
    background: #6366f1; color: #fff; padding: 20px;
    border-radius: 8px; text-align: center;
  }
  .auto-fill {
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  }
  .auto-fit {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  }
</style>

<div class="demo auto-fill">
  <div>1</div><div>2</div><div>3</div>
</div>
<div class="demo auto-fit">
  <div>1</div><div>2</div><div>3</div>
</div>
```

| | `auto-fill` | `auto-fit` |
|---|------------|-----------|
| 子项数量 < 可容纳列数 | 保留空列轨道（有空白间隙） | 子项拉伸填满（无空列） |
| 子项数量 ≥ 可容纳列数 | 行为相同 | 行为相同 |
| 典型场景 | 固定宽度网格 | 自适应卡片列表 |

**记忆**：`fill` = 填轨道（可能留白），`fit` = 填内容（拉伸填满）。

### 29.1 手把手对比 auto-fill 与 auto-fit

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 把 §29 两段 HTML 存为 `grid-demo.html`，浏览器打开 | 上下两个三格紫条区域 | 检查 `display:grid` 是否生效 |
| 2 | 拉宽浏览器到 >600px，观察 **auto-fill** 区域 | 3 个格子左侧排列，**右侧可能有空白列轨道** | 子项少时 fill 保留空轨道 |
| 3 | 看 **auto-fit** 区域 | 3 个格子**拉伸填满**整行，右侧无空列 | fit 会折叠空轨道 |
| 4 | DevTools 选中 `.auto-fill`，开 Grid overlay | 可见未使用的列轨道线 | 与 auto-fit 对比轨道数量 |
| 5 | 把 `minmax(150px, 1fr)` 改成 `minmax(200px, 1fr)` 再试 | 能放的列数变少，更早换行 | min 过大时窄屏只剩一列 |

---

## 30. 响应式图片专题

### 30.1 `srcset` + `sizes` — 根据不同屏幕加载不同图

```html
<img
  src="photo-800.jpg"
  srcset="photo-400.jpg 400w,
          photo-800.jpg 800w,
          photo-1200.jpg 1200w"
  sizes="(max-width: 600px) 100vw,
         (max-width: 1024px) 50vw,
         33vw"
  alt="响应式图片示例"
/>
```

- `srcset`：告诉浏览器有哪些尺寸的图
- `sizes`：告诉浏览器在不同条件下图片显示多大
- 浏览器自动选择最合适的图片下载

### 30.2 `<picture>` — 根据条件切换完全不同的图

```html
<picture>
  <!-- 移动端竖版 -->
  <source media="(max-width: 768px)" srcset="hero-mobile.jpg" />
  <!-- 平板横版 -->
  <source media="(max-width: 1200px)" srcset="hero-tablet.jpg" />
  <!-- 桌面大图 -->
  <img src="hero-desktop.jpg" alt="Hero 图片" />
</picture>
```

`<picture>` 适合：不同屏幕用完全不同的图片（如移动端竖版 vs 桌面横版、不同格式如 WebP vs JPEG）。

```html
<picture>
  <!-- 支持 WebP 的浏览器用 WebP -->
  <source type="image/webp" srcset="photo.webp" />
  <!-- 不支持 WebP 的回退到 JPEG -->
  <img src="photo.jpg" alt="照片" />
</picture>
```

### 30.3 实践建议

- 普通内容图片：用 `img` + `srcset` + `sizes`
- 需要艺术指导（不同裁剪比例）：用 `<picture>`
- 格式回退（WebP → JPEG）：用 `<picture>` + `type`
- 所有图片都应设置 `alt`，都有 `loading="lazy"` 更好

---

## 31. CSS 容器查询 `@container` 入门

媒体查询（`@media`）根据**视口**宽度适配。容器查询根据**父容器**宽度适配——组件放在不同宽度的容器里自动适配。

```css
/* 1. 定义容器 */
.card-wrapper {
  container-type: inline-size; /* 或 container: card-wrapper / inline-size */
  container-name: card-wrapper;
}

/* 2. 根据容器宽度写样式 */
@container card-wrapper (min-width: 400px) {
  .card {
    display: flex;
    gap: 16px;
  }
}

@container card-wrapper (max-width: 399px) {
  .card {
    display: block;
  }
}
```

**容器查询 vs 媒体查询**：
- 媒体查询：响应"整个视口"大小
- 容器查询：响应"父容器"大小——组件在侧边栏（窄）和主内容区（宽）自动适配不同布局

**浏览器兼容**：2024 年后主流浏览器全部支持。

---

## 32. 常见动画模式集合

### 32.1 骨架屏闪烁

```css
@keyframes shimmer {
  0%   { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}

.skeleton {
  height: 16px; border-radius: 4px;
  background: linear-gradient(90deg, #e2e8f0 25%, #f1f5f9 50%, #e2e8f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}
```

### 32.2 弹入弹出

```css
@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes fadeOut {
  from { opacity: 1; }
  to   { opacity: 0; }
}

.modal-enter { animation: slideUp 0.3s cubic-bezier(0.34, 1.56, 0.64, 1) forwards; }
.modal-exit  { animation: fadeOut 0.2s ease forwards; }
```

### 32.3 打字效果

```css
@keyframes typing {
  from { width: 0; }
  to   { width: 100%; }
}

@keyframes blink {
  0%, 100% { border-color: transparent; }
  50%      { border-color: #333; }
}

.typewriter {
  overflow: hidden; white-space: nowrap;
  border-right: 3px solid #333;
  width: fit-content;
  animation: typing 2s steps(20) forwards, blink 0.8s step-end infinite;
}
```

### 32.4 脉冲提示（新消息小红点）

```css
@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50%      { transform: scale(1.3); opacity: 0.7; }
}

.badge-notification {
  animation: pulse 2s ease-in-out infinite;
}
```

---

## 33. `will-change` 与动画性能优化

### 核心原则

1. **动画只用 `transform` 和 `opacity`**：这两个属性只触发合成（Composite），不触发重排（Layout）或重绘（Paint），性能最好
2. **用 `will-change` 预告即将变化的属性**：让浏览器提前准备 GPU 加速层
3. **不要在太多元素上用 `will-change`**：每个加速层都消耗 GPU 内存

```css
/* 正确用法：动画前告知 */
.animated-card {
  will-change: transform;
  transition: transform 0.3s ease;
}
.animated-card:hover {
  transform: translateY(-4px);
}

/* 动画结束后移除（JS 辅助） */
/*
card.addEventListener('animationend', () => {
  card.style.willChange = 'auto';
});
*/
```

### 各属性触发的渲染阶段

| 触发阶段 | CSS 属性 | 性能 |
|----------|----------|------|
| 仅合成（最佳） | `transform`, `opacity` | ⭐⭐⭐⭐⭐ |
| 重绘 | `color`, `background`, `box-shadow`, `border-color` | ⭐⭐⭐ |
| 重排（最差） | `width`, `height`, `left`, `top`, `margin`, `padding` | ⭐ |

**记忆口诀**：动画用 transform + opacity = 60fps 不卡顿。改 width/height/left/top = 掉帧。

---

## 34. 完整实战：带入场动画的产品展示

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>动画产品展示</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: system-ui, sans-serif; background: #f8fafc;
      padding: 40px 20px;
    }
    .page-title {
      text-align: center; font-size: 2rem; margin-bottom: 32px;
      animation: slideUp 0.6s ease;
    }

    .products {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
      gap: 20px; max-width: 1100px; margin: 0 auto;
    }

    .product-card {
      background: #fff; border-radius: 16px; overflow: hidden;
      box-shadow: 0 4px 16px rgba(0,0,0,.06);
      opacity: 0; /* 初始透明，动画到显示 */
      animation: fadeInUp 0.6s ease forwards;
      transition: transform 0.3s ease, box-shadow 0.3s ease;
    }
    /* 每个卡片依次延迟 0.1s 入场 */
    .product-card:nth-child(1) { animation-delay: 0.1s; }
    .product-card:nth-child(2) { animation-delay: 0.2s; }
    .product-card:nth-child(3) { animation-delay: 0.3s; }
    .product-card:nth-child(4) { animation-delay: 0.4s; }
    .product-card:nth-child(5) { animation-delay: 0.5s; }
    .product-card:nth-child(6) { animation-delay: 0.6s; }

    .product-card:hover {
      transform: translateY(-8px);
      box-shadow: 0 16px 36px rgba(0,0,0,.1);
    }

    .product-card img {
      width: 100%; aspect-ratio: 4/3; object-fit: cover;
      transition: transform 0.5s ease;
    }
    .product-card:hover img { transform: scale(1.08); }

    .product-info { padding: 16px; }
    .product-info h3 { margin-bottom: 4px; }
    .product-info .price {
      font-size: 1.25rem; font-weight: 700; color: #6366f1;
    }
    .product-info .badge {
      display: inline-block; padding: 2px 8px;
      background: #fef3c7; color: #d97706; border-radius: 4px;
      font-size: 0.8rem; font-weight: 600; margin-left: 8px;
    }

    .spinner {
      width: 40px; height: 40px; margin: 40px auto;
      border: 3px solid #e5e7eb;
      border-top-color: #6366f1; border-radius: 50%;
      animation: spin 0.8s linear infinite;
    }

    @keyframes fadeInUp {
      from { opacity: 0; transform: translateY(30px); }
      to   { opacity: 1; transform: translateY(0); }
    }
    @keyframes slideUp {
      from { opacity: 0; transform: translateY(20px); }
      to   { opacity: 1; transform: translateY(0); }
    }
    @keyframes spin {
      to { transform: rotate(360deg); }
    }
  </style>
</head>
<body>
  <h1 class="page-title">✨ 精选产品</h1>

  <!-- 加载动画（Demo 用，实际按需显隐） -->
  <div class="spinner" id="loading" style="display:none;"></div>

  <div class="products">
    <article class="product-card">
      <img src="https://picsum.photos/400/300?p=1" alt="产品 1" />
      <div class="product-info">
        <h3>机械键盘</h3>
        <span class="price">¥299</span>
        <span class="badge">热卖</span>
      </div>
    </article>
    <article class="product-card">
      <img src="https://picsum.photos/400/300?p=2" alt="产品 2" />
      <div class="product-info">
        <h3>无线鼠标</h3>
        <span class="price">¥89</span>
      </div>
    </article>
    <article class="product-card">
      <img src="https://picsum.photos/400/300?p=3" alt="产品 3" />
      <div class="product-info">
        <h3>显示器</h3>
        <span class="price">¥1299</span>
      </div>
    </article>
    <article class="product-card">
      <img src="https://picsum.photos/400/300?p=4" alt="产品 4" />
      <div class="product-info">
        <h3>桌面台灯</h3>
        <span class="price">¥159</span>
      </div>
    </article>
    <article class="product-card">
      <img src="https://picsum.photos/400/300?p=5" alt="产品 5" />
      <div class="product-info">
        <h3>笔记本支架</h3>
        <span class="price">¥199</span>
        <span class="badge">新品</span>
      </div>
    </article>
    <article class="product-card">
      <img src="https://picsum.photos/400/300?p=6" alt="产品 6" />
      <div class="product-info">
        <h3>Type-C 扩展坞</h3>
        <span class="price">¥249</span>
      </div>
    </article>
  </div>
</body>
</html>
```

**本示例用到的知识点**：
- Grid `auto-fit + minmax` 自适应列数
- `animation-delay` 实现依次入场
- `transform` + `box-shadow` 悬停效果
- 图片 `object-fit: cover` + `aspect-ratio`
- `will-change` 可以配合使用（见第 33 节）
- 所有动画只用 `transform` 和 `opacity`（最佳性能）

---

## 36. DevTools Flex / Grid Overlay 八步实操

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 打开 §21 综合实战 HTML，F12 → **Elements**，选中 `.navbar` | Styles 里 `display: flex` | 没有 flex 则子项 `space-between` 不生效 |
| 2 | 在 Styles 面板 `flex` 旁找 **flex** 徽章或 **Layout** 标签，点开 Flex overlay | 页面上出现**虚线主轴**、子项之间 gap/对齐线 | Chrome 119+：选中 flex 容器自动提示；旧版在 Layout 侧边栏 |
| 3 | 选中 `.cards`（Grid 容器），开启 **Grid overlay** | 每列每格有编号线，看清 `1fr` 列如何分配 | 确认 `display: grid` 和 `grid-template-columns` |
| 4 | 在 Styles 里临时改 `.cards` 的 `justify-content: center`（若有多余空间） | 网格整体在容器内水平居中 | Grid 的 justify 管**整网格**，不是单元格内 |
| 5 | 设备工具栏 **Ctrl+Shift+M**，宽度切到 375px | 导航变汉堡菜单、卡片变单列 | 若无变化，检查 `@media` 是否 `min-width` 写反 |
| 6 | 宽度切到 1280px | 卡片三列、导航横排 | 横向滚动条出现 → 查 `min-width:0`、图片 `max-width:100%` |
| 7 | 选中 `.card:hover` 规则，用 **:hov** 强制 `:hover` | 看到 `transform` 与 `box-shadow` 过渡 | 过渡写在初始 `.card` 上，不是只写在 hover |
| 8 | **Performance** 录一段 hover 动画（可选） | 主线程以 Composite 为主，少 Layout 紫块 | 若大量 Layout，检查是否动画了 `width`/`height` |

### Flex / Grid 调试速查

| 症状 | DevTools 先查 |
|------|---------------|
| `flex:1` 不生效 | 父级是否 `display:flex`？ |
| 文字把 flex 子项撑破 | 子项加 `min-width: 0` 或 `overflow: hidden` |
| Grid 区域名不生效 | 子项是否写了对应 `grid-area`？区域是否矩形？ |
| 断点不触发 | `@media (min-width:768px)` 是 ≥768 才生效 |
| 动画卡顿 | 是否只动画 `transform`/`opacity`（§33） |

---

## 37. 与 Vue 09 Element 布局的衔接

| 本章能力 | Element Plus 对应 | Vue 09 章节 |
|----------|-------------------|-------------|
| Flex 列布局 + `justify-content: space-between` | `el-header` 内 logo 与用户信息两端对齐 | §6 AdminLayout |
| 固定侧栏宽 + 主区自适应 | `el-aside width="220px"` + `el-main` 占满剩余 | §6 |
| `grid-template-areas` 四区域 | 与 `el-container` 嵌套等价思路；复杂仪表盘可原生 Grid | §6、§24 |
| `el-row` / `el-col` | Flex 封装，`:gutter` ≈ `gap` | §5 表单行、列表 |
| 响应式断点 | §24 侧栏折叠、`display` 切换 | §24 |
| transition 反馈 | 按钮 hover、`el-card` 阴影（可自定义） | 全文组件 |

**推荐练习**：用纯 CSS Flex/Grid 先复刻 [Vue 09 §6 AdminLayout](../Vue/09-Element-Plus与UI工程化.md) 的 Header + Aside + Main 结构（不用 Vue），再对照 Element 源码 class，你会认出 `display:flex` 和 `min-height:100vh`。

**与 04 章分工**：04 负责盒模型、角标 `absolute`、弹窗 `fixed` + `z-index`；本章负责**页面主骨架**；09 章把骨架组件化。

### 布局选型决策树（30 秒）

```
要排几个东西？
├─ 一行或一列（导航、表单行、工具栏）→ Flex
├─ 同时要管行和列（整页、仪表盘、图片墙）→ Grid
├─ 叠在某一区域上（角标、弹窗、下拉）→ 04 章 position
└─ 文字绕图片 → float（少见）或 Grid 子区域
```

Vue 09 的 `el-container` 属于 **Flex 列 + 嵌套 Flex 行**；商品列表 `el-row`/`el-col` 属于 **Flex 行 + 比例 span**。

---

## 38. 闭卷自测

1. Flex 的**主轴**和**交叉轴**由什么属性决定？默认 row 时 justify 管哪方向？
2. `justify-content: space-between` 和 `space-around` 在剩余空间分配上差在哪？
3. `flex: 1` 常见含义是什么（grow / shrink / basis）？
4. `grid-template-columns: repeat(3, 1fr)` 表示什么？`fr` 和 `%` 有何不同？
5. `auto-fill` 和 `auto-fit` 在子项少于列数时行为差在哪？
6. 移动优先为什么推荐 `min-width` 而不是 `max-width`？
7. `transition` 要生效，通常要满足哪 3 个条件？
8. 为什么入场动画推荐 `transform` + `opacity`？
9. **动手**：Flex 实现左侧 200px 侧栏 + 右侧自适应主内容（各一行关键 CSS）。
10. **动手**：Grid 两行两列，`grid-template-areas` 命名 header/main/footer，手机单列。
11. **综合**：§21 导航在 768px 以下变汉堡菜单，写出核心 `@media` 与 `flex-direction` 变化。
12. **综合**：说明 Vue 09 `el-container` 布局对应本章哪几个 Flex/Grid 概念。

### 38.1 自测参考答案

1. `flex-direction`；默认 row 时 justify 管**水平**（主轴），align 管垂直（交叉轴）。
2. `space-between` 首尾贴边、中间均分间隙；`space-around` 每项两侧都有半份间隙，首尾也有留白。
3. 常等价 `flex: 1 1 0%`：可放大、可缩小、初始基准 0，按比例分剩余空间。
4. 三列等分剩余宽；`fr` 分的是 Grid **剩余空间**，不受 gap 影响；`%` 相对容器总宽，嵌套时易算乱。
5. `auto-fill` 保留空列轨道；`auto-fit` 合并空列，子项拉伸填满。
6. 默认轻量手机样式，大屏渐进增强；与主流框架一致，避免小屏删桌面样式。
7. 有触发（hover/类名变化）；属性可过渡；`transition` 写在**初始状态**；duration > 0。
8. 只触发合成层，不触发重排，60fps 更稳（§33）。
9. `.wrap{display:flex}` `.side{flex:0 0 200px}` `.main{flex:1;min-width:0}`。
10. 桌面 `grid-template-areas:"header header" "main main" "footer footer"`；移动 `@media(max-width:767px){ grid-template-columns:1fr; areas:"header" "main" "footer"; }`（或移动优先反写）。
11. 默认 `.nav-menu{display:none}` 或列排；`@media(min-width:768px){ .nav-toggle{display:none}; .nav-menu{display:flex;flex-direction:row} }`。
12. 外层 `el-container` ≈ `min-height:100vh` + Flex 列；内层 `el-header` + 嵌套 `el-container` ≈ 顶栏 + 横向 Flex（aside + main）；见 09 §6。

---

## 39. 费曼检验

请在不看资料的情况下，用 3 分钟向没学过编程的朋友解释本章核心：

1. **Flex = 弹簧排排站**：一行（或一列）小朋友，队长喊靠左、居中、还是按比例拉长占满空位；导航栏、表单行都这么排。
2. **Grid = Excel 表格**：先画几行几列、给区域起名，再往格子里填内容；整页布局、图片墙像填表一样直观。
3. **响应式 + 动效**：手机默认简单布局，屏幕变宽再加列；按钮 hover 用渐变过渡，入场用关键帧——但别乱动宽高，用「挪动和透明度」才流畅。

> 下一章：[06 — JavaScript 基础语法与数据类型](./06-JavaScript基础语法与数据类型.md)。布局搭好后，用 JS 让页面「动起来」；Vue 路线可在 06～07 后开 [Vue 01](../Vue/01-Vue入门与环境搭建.md)，布局到 [Vue 09](../Vue/09-Element-Plus与UI工程化.md) 用 Element 组件封装。

---

*文档版本：扩充版 v2 | 建议学习时长：6~8 小时（含综合实战与 DevTools）*

---

## 35. 学完标准（扩充版）

- [ ] 不查文档写出 Flex 水平垂直居中
- [ ] 解释 `justify-content`（主轴）和 `align-items`（交叉轴）的区别
- [ ] 用 `flex: 1 1 280px` 做自适应卡片列表
- [ ] 用 `grid-template-areas` 做四区域页面布局
- [ ] 用 `repeat(auto-fit, minmax())` 做图片墙
- [ ] 采用移动优先写至少两个断点的响应式页面
- [ ] 写 transition hover 效果和 `@keyframes` 入场动画
- [ ] 知道动画优先 `transform` 和 `opacity`（性能原因）
- [ ] 了解 `srcset` + `<picture>` 响应式图片
- [ ] 了解 `@container` 容器查询的基本概念
- [ ] 知道圣杯布局思路，并能用 Flex 替代
- [ ] 完成综合实战：导航 + 卡片 + 图片墙 + 入场动画
