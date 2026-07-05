# JavaScript DOM、BOM 与事件机制

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0 导读、Elements DevTools、待办逐行读、FAQ 12 题、闭卷自测、费曼检验；链 Vue 08 / 计网 04 -->

## 0. 读前导读（零基础也能跟上）

> **读者假设**：你已会 [07 章](./07-JavaScript流程控制函数对象数组与ES6基础.md) 的函数与数组方法，能在 Console 处理 JSON。本章让 JavaScript **真正碰到网页**——点按钮、改文字、存待办，是 Vue 模板编译成 DOM 操作之前的「裸机版」。

### 0.1 用一句话弄懂本章

**一句话**：浏览器把 HTML 变成 **DOM 树**，JavaScript 用 `querySelector` 选中节点、用 `addEventListener` 听用户操作、用 **事件委托** 管动态列表——待办清单就是迷你版 Vue 页面。

**生活类比**：

| 本章概念 | 类比 | Vue / 后端对应 |
|---------|------|----------------|
| **DOM 树** | 网页的「家族族谱」 | Vue 最终也渲染成 DOM |
| **querySelector** | 按门牌号找人 | 类似 `document.querySelector('#app')` 再 mount |
| **classList.toggle** | 电灯开关 | `v-bind:class` 的底层 |
| **事件委托** | 班长统一收作业 | `v-on` 在父级监听子项 |
| **localStorage** | 浏览器小抽屉 | [09 章](./09-JavaScript异步编程网络请求与本地存储.md) 详讲；Vue 08 存 token 同源 |

**为什么重要**：[Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 联调失败时，DevTools **Elements** 看 DOM 是否更新、**Network** 看请求——Network 原理在 [10 章](./10-浏览器HTTP网络与Web基础.md) 与 [计网 04](../计算机网络/04-HTTP协议深入.md)。

---

### 0.2 你需要提前知道什么

| 水平 | 建议 |
|------|------|
| 不会 JS | 先 [06](./06-JavaScript基础语法与数据类型.md) + [07](./07-JavaScript流程控制函数对象数组与ES6基础.md) |
| 已会 07 | 从 §3 选手元素；§17 事件委托是核心 |
| 目标 Vue | 本章 + [09 异步](./09-JavaScript异步编程网络请求与本地存储.md) 后可开 [Vue 06](../Vue/05-组合式API与script-setup.md) |
| 排查接口页面空白 | 本章 DOM + [10 Network](./10-浏览器HTTP网络与Web基础.md) + [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) |

---

### 0.3 本章知识地图（☐→☑）

- [ ] 会用 `querySelector` / `querySelectorAll` 选中元素
- [ ] 区分 `textContent` 与 `innerHTML`（含 XSS 意识）
- [ ] 会用 `classList` 切换样式状态
- [ ] 熟练 `addEventListener`，会 `preventDefault` / `stopPropagation`
- [ ] 能写事件委托版动态列表（§28 / §37）
- [ ] 知道 `DOMContentLoaded` 与 `defer` 脚本时机
- [ ] 完成待办清单或 Tab 切换实战
- [ ] 闭卷自测 ≥ 8/10

---

### 0.4 建议学习时长与节奏

| 阶段 | 时间 |
|------|------|
| DOM 查找与修改 §3～§8 | 2 小时 |
| 事件基础 §9～§17 | 2.5 小时 |
| BOM + 加载时机 §18～§26 | 1.5 小时 |
| Tab / 表单 / 待办实战 §29～§37 | 3 小时 |
| DevTools + 自测 | 1 小时 |

---

### 0.5 学完本章你能做什么（可验证）

1. 不依赖框架，写一个可增删改、刷新仍在的待办清单（localStorage）。
2. 解释：为什么给 100 个 `<li>` 分别绑 click 不如给 `<ul>` 绑一次委托。
3. 在 Elements 面板找到 `#list` 节点，确认 `fetch` 渲染后子节点数量变对（配合 [09 章 fetch](./09-JavaScript异步编程网络请求与本地存储.md)）。

---

### 0.6 核心术语三件套

**术语（DOM Document Object Model）**：浏览器把 HTML 解析成的可编程对象树；每个标签是一个节点。
**生活类比**：网页是建筑图纸，DOM 是带编号的实物模型——JS 按编号改模型，屏幕跟着变。
**为什么重要**：所有前端框架最终都操作 DOM（或虚拟 DOM 再 diff）；不会 DOM 无法调试「数据对了页面没动」。
**本章用到的地方**：§2～§8。

**术语（事件冒泡 Event Bubbling）**：子元素触发的事件会向上传到父元素。
**生活类比**：小孩房间按门铃，全家都能听见——父级 `ul` 能收到里面 `button` 的点击。
**为什么重要**：事件委托的基础；Vue 的 `@click` 在组件根上有时也利用类似机制。
**本章用到的地方**：§15～§17。

**术语（事件委托 Event Delegation）**：把监听器绑在父元素，用 `event.target` 判断实际点击的子元素。
**生活类比**：前台统一收件，看信封上名字分发——不用给每个员工门口装邮箱。
**为什么重要**：动态列表（Ajax 加载、待办增删）性能与代码量都更优；Vue `v-for` 列表点击本质思路相同。
**本章用到的地方**：§17、§28、§37。

---

## 1. 为什么这一份极其重要

很多人学完 JavaScript 基础后，还是不会真正做网页交互。

原因通常是不会这几件事：

- 选中页面元素
- 修改页面内容
- 监听用户操作
- 处理点击、输入、滚动等事件

这些能力都在：

- DOM
- BOM
- 事件机制

里。

## 2. DOM 是什么

DOM 全称是：

- Document Object Model

你可以先把它理解成：

- 浏览器把 HTML 页面解析成一棵对象树

JavaScript 可以通过 DOM：

- 查找元素
- 修改元素
- 创建元素
- 删除元素

## 3. 选中元素

### 3.1 `getElementById`

```js
const title = document.getElementById("title");
```

### 3.2 `querySelector`

```js
const box = document.querySelector(".box");
```

### 3.3 `querySelectorAll`

```js
const items = document.querySelectorAll(".item");
```

这是现代开发中非常常见的方式。

## 4. 修改文本内容

### `textContent`

```js
title.textContent = "新的标题";
```

更适合纯文本。

### `innerHTML`

```js
box.innerHTML = "<strong>加粗内容</strong>";
```

可以插入 HTML 结构。

但要注意：

- 不要随便插入不可信内容

## 5. 修改属性

```js
const img = document.querySelector("img");
img.src = "./new.png";
img.alt = "新图片";
```

也可以用：

```js
img.setAttribute("title", "提示");
```

## 6. 修改样式

### 直接改 style

```js
box.style.color = "red";
box.style.backgroundColor = "black";
```

### 更推荐的方式：切换 class

```js
box.classList.add("active");
box.classList.remove("hidden");
box.classList.toggle("open");
```

这通常比直接改一堆 style 更清晰。

## 7. 创建和插入元素

```js
const li = document.createElement("li");
li.textContent = "新的列表项";

const list = document.querySelector("ul");
list.appendChild(li);
```

## 8. 删除元素

```js
li.remove();
```

或者通过父元素移除子元素。

## 9. 事件是什么

事件就是：

- 用户或浏览器发生的某种行为

常见事件：

- 点击
- 输入
- 提交
- 鼠标移入移出
- 键盘按下
- 页面加载

## 10. 事件监听

```js
const btn = document.querySelector("button");

btn.addEventListener("click", function () {
  console.log("按钮被点击了");
});
```

这是最推荐的现代写法。

## 11. 常见事件类型

### `click`

点击事件。

### `input`

输入框内容变化。

### `change`

值变化并确认后触发。

### `submit`

表单提交。

### `keydown`

键盘按下。

### `mouseover` / `mouseout`

鼠标移入移出。

## 12. 事件对象

事件回调通常会收到一个事件对象。

```js
btn.addEventListener("click", function (event) {
  console.log(event);
});
```

常见用途：

- 获取点击目标
- 阻止默认行为
- 阻止冒泡

## 13. `event.target`

表示真正触发事件的元素。

```js
list.addEventListener("click", function (event) {
  console.log(event.target);
});
```

## 14. 阻止默认行为

例如点击链接默认会跳转，表单默认会提交。

```js
form.addEventListener("submit", function (event) {
  event.preventDefault();
});
```

### 14.1 阻止表单默认提交手把手

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 写 `<form id="f"><input name="q"><button>搜</button></form>` 无 preventDefault | 点击提交 URL 出现 `?q=...`，页面刷新 | 这就是默认行为 |
| 2 | JS 里 `f.addEventListener('submit', e => { e.preventDefault(); console.log('拦截') })` | 不刷新，Console 打印 | 监听绑在 form 不是 button |
| 3 | 在 handler 里 `fetch` 或读 FormData | Network 出现请求 | 忘记 async/await 见 [09 章](./09-JavaScript异步编程网络请求与本地存储.md) |
| 4 | 对照 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) `@submit.prevent` | 语法糖等价 preventDefault | Vue 表单联调同一逻辑 |

## 15. 事件冒泡

当子元素触发事件时，事件可能向父级一层层冒泡。

```html
<div class="parent">
  <button class="child">点击</button>
</div>
```

如果给父子都绑定点击事件，点按钮时父级也可能收到。

## 16. 阻止冒泡

```js
btn.addEventListener("click", function (event) {
  event.stopPropagation();
});
```

## 17. 事件委托

这是非常实用的技巧。

思路：

- 不给每个子元素单独绑事件
- 而是给父元素绑事件，再通过 `event.target` 判断是谁触发的

适合：

- 动态列表
- 大量重复元素

## 18. BOM 是什么

BOM 全称一般叫：

- Browser Object Model

可以理解为：

- 浏览器环境提供的一些对象和能力

常见对象：

- `window`
- `location`
- `history`
- `navigator`

## 19. `window`

浏览器中的全局对象。

很多全局方法和变量都挂在它上面。

## 20. `location`

用于获取或修改当前地址信息。

```js
console.log(location.href);
```

## 21. `history`

用于浏览器历史记录控制。

```js
history.back();
history.forward();
```

## 22. 定时器

### `setTimeout`

延迟执行一次。

```js
setTimeout(() => {
  console.log("1秒后执行");
}, 1000);
```

### `setInterval`

周期执行。

```js
const timer = setInterval(() => {
  console.log("每秒执行一次");
}, 1000);
```

停止：

```js
clearInterval(timer);
```

## 23. 页面加载事件

```js
window.addEventListener("load", function () {
  console.log("页面资源加载完成");
});
```

## 24. 表单交互常见场景

你以后会不断写这些逻辑：

- 获取输入框值
- 校验是否为空
- 实时显示提示
- 提交前拦截

```js
const input = document.querySelector("#username");
console.log(input.value);
```

## 25. DOM 树与节点类型

浏览器把 HTML 解析成树形结构：

```text
document
└── html
    ├── head
    │   ├── meta
    │   └── title
    └── body
        ├── header
        ├── main
        │   └── p
        └── script
```

常见节点类型（了解即可）：

- **元素节点**：`<div>`、`<p>` 等
- **文本节点**：标签内的文字
- **属性**：`id`、`class` 等（在 DOM 里以不同方式访问）

```js
const p = document.querySelector("p");
p.parentElement;      // 父元素
p.children;           // 子元素集合
p.nextElementSibling; // 下一个兄弟元素
p.closest(".card");   // 向上找最近的匹配祖先
```

---

## 26. 脚本加载时机（非常重要）

### 问题：为什么 `querySelector` 得到 `null`？

```html
<!-- 错误：脚本在 body 内容之前执行 -->
<head>
  <script>
    const btn = document.querySelector("#btn"); // null
  </script>
</head>
<body>
  <button id="btn">点我</button>
</body>
```

### 三种正确做法

**做法 1**：`<script>` 放在 `</body>` 前（初学推荐）

```html
<body>
  <button id="btn">点我</button>
  <script src="./main.js"></script>
</body>
```

**做法 2**：`DOMContentLoaded`

```js
document.addEventListener("DOMContentLoaded", () => {
  const btn = document.querySelector("#btn");
  // 此时 DOM 已解析完，可以安全操作
});
```

**做法 3**：`<script defer src="...">` 放在 head（以后学）

---

## 27. 事件流：捕获 → 目标 → 冒泡

```text
点击 button 时事件传播顺序：
window → document → html → body → div → button（目标）→ 再冒泡回去
```

```js
div.addEventListener("click", () => console.log("捕获"), true);  // 第三参数 true
div.addEventListener("click", () => console.log("冒泡"));
// 点击内部 button：先「捕获」再「冒泡」
```

日常开发 **99% 用冒泡**（默认 false）即可；知道有捕获阶段即可。

---

## 28. 事件委托完整示例：待办列表

```html
<ul id="todo-list">
  <li data-id="1">任务一 <button class="del">删</button></li>
  <li data-id="2">任务二 <button class="del">删</button></li>
</ul>
<button id="add">添加</button>
```

```js
const list = document.querySelector("#todo-list");
let nextId = 3;

// 只绑一次：删除 + 以后动态添加的项也能删
list.addEventListener("click", (e) => {
  if (e.target.classList.contains("del")) {
    e.target.closest("li").remove();
  }
});

document.querySelector("#add").addEventListener("click", () => {
  const li = document.createElement("li");
  li.dataset.id = nextId++;
  li.innerHTML = `新任务 <button class="del">删</button>`;
  list.appendChild(li);
});
```

---

## 29. 完整实战：Tab 切换

```html
<div class="tabs">
  <button class="tab active" data-tab="0">首页</button>
  <button class="tab" data-tab="1">产品</button>
  <button class="tab" data-tab="2">关于</button>
</div>
<div class="panel active" data-panel="0">首页内容</div>
<div class="panel" data-panel="1">产品内容</div>
<div class="panel" data-panel="2">关于内容</div>
```

```css
.panel { display: none; }
.panel.active { display: block; }
.tab.active { font-weight: bold; border-bottom: 2px solid blue; }
```

```js
const tabs = document.querySelectorAll(".tab");
const panels = document.querySelectorAll(".panel");

tabs.forEach((tab) => {
  tab.addEventListener("click", () => {
    const index = tab.dataset.tab;
    tabs.forEach((t) => t.classList.remove("active"));
    panels.forEach((p) => p.classList.remove("active"));
    tab.classList.add("active");
    document.querySelector(`[data-panel="${index}"]`).classList.add("active");
  });
});
```

### 29.1 Tab 切换逻辑逐行读

| 行号/代码 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `data-tab` / `data-panel` | 自定义 data 属性存索引 | 不一致 → 面板对不上 |
| `tab.dataset.tab` | 读 data-tab 字符串 | 应用 Number 若做算术 |
| `tabs.forEach` 移除 active | 先清空所有 Tab 高亮 | 只 add 不 remove → 多个 active |
| `querySelector(\`[data-panel="${index}"]\`)` | 找对应面板 | 模板字符串拼错 → null |
| `classList.add/remove("active")` | 用 CSS 控制显隐 | 直接改 style.display 也可，难维护 |

Vue 里同样模式：`v-for` 渲染 Tab，`@click` 改 `activeIndex`——见 [Vue 04 组件通信](../Vue/04-组件基础与组件通信.md)。接口驱动的 Tab 内容在 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 拉数据后渲染。

---

## 30. 表单校验完整示例

```html
<form id="login-form">
  <input id="username" placeholder="用户名" />
  <span id="user-error" class="error"></span>
  <input id="password" type="password" placeholder="密码" />
  <button type="submit">登录</button>
</form>
```

```js
const form = document.querySelector("#login-form");
const userInput = document.querySelector("#username");
const userError = document.querySelector("#user-error");

userInput.addEventListener("input", () => {
  if (userInput.value.trim().length < 2) {
    userError.textContent = "用户名至少 2 个字符";
  } else {
    userError.textContent = "";
  }
});

form.addEventListener("submit", (e) => {
  e.preventDefault();
  if (userInput.value.trim().length < 2) {
    userError.textContent = "请填写有效用户名";
    return;
  }
  console.log("提交", { user: userInput.value });
});
```

---

## 31. `input` vs `change` vs `submit`

| 事件 | 触发时机 |
|------|----------|
| `input` | 输入框内容每次变化（实时） |
| `change` | 失焦且值改变；select、checkbox 选中变化 |
| `submit` | 表单提交（点 submit 按钮或回车） |

搜索建议、字数统计用 `input`；下拉框选完用 `change`。

---

## 32. 初学者常见错误

### 32.1 脚本执行时 DOM 未就绪

见第 26 节，用底部 script 或 `DOMContentLoaded`。

### 32.2 用 `innerHTML` 拼接用户输入（XSS 风险）

```js
// 危险：用户输入 <img src=x onerror=alert(1)>
box.innerHTML = userInput;

// 安全：纯文本用 textContent
box.textContent = userInput;
```

### 32.3 `querySelectorAll` 不是数组

需转数组或用 `forEach`（NodeList 现代浏览器也支持 forEach）。

### 32.4 循环里给每个按钮绑事件却用错闭包

```js
// 错：var i 导致全是最后一个
for (var i = 0; i < btns.length; i++) {
  btns[i].onclick = () => console.log(i);
}
// 对：let 或 data-index
```

---

## 33. 完整实战：可拖拽排序列表

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <title>拖拽排序</title>
  <style>
    .list { list-style: none; padding: 0; max-width: 400px; margin: 20px auto; }
    .list-item {
      padding: 12px 16px; margin-bottom: 8px;
      background: #fff; border: 1px solid #e2e8f0; border-radius: 8px;
      cursor: grab; display: flex; align-items: center; gap: 12px;
      transition: box-shadow 0.2s, transform 0.2s;
    }
    .list-item:active { cursor: grabbing; }
    .list-item.dragging { opacity: 0.5; box-shadow: 0 4px 16px rgba(0,0,0,0.1); }
    .list-item.drag-over { border-color: #6366f1; background: #eef2ff; transform: scale(1.02); }
    .handle { color: #94a3b8; font-size: 20px; cursor: grab; user-select: none; }
  </style>
</head>
<body>
  <ul class="list" id="sortable">
    <li class="list-item" draggable="true"><span class="handle">⋮⋮</span> 任务一</li>
    <li class="list-item" draggable="true"><span class="handle">⋮⋮</span> 任务二</li>
    <li class="list-item" draggable="true"><span class="handle">⋮⋮</span> 任务三</li>
    <li class="list-item" draggable="true"><span class="handle">⋮⋮</span> 任务四</li>
  </ul>

  <script>
    const list = document.getElementById("sortable");
    let draggedItem = null;

    list.addEventListener("dragstart", (e) => {
      draggedItem = e.target.closest(".list-item");
      if (!draggedItem) return;
      draggedItem.classList.add("dragging");
      e.dataTransfer.effectAllowed = "move";
    });

    list.addEventListener("dragend", (e) => {
      draggedItem?.classList.remove("dragging");
      document.querySelectorAll(".drag-over").forEach(el => el.classList.remove("drag-over"));
      draggedItem = null;
    });

    list.addEventListener("dragover", (e) => {
      e.preventDefault();
      const target = e.target.closest(".list-item");
      if (!target || target === draggedItem) return;
      document.querySelectorAll(".drag-over").forEach(el => el.classList.remove("drag-over"));
      target.classList.add("drag-over");
    });

    list.addEventListener("drop", (e) => {
      e.preventDefault();
      const target = e.target.closest(".list-item");
      if (!target || target === draggedItem || !draggedItem) return;
      const children = [...list.children];
      const targetIndex = children.indexOf(target);
      const draggedIndex = children.indexOf(draggedItem);
      if (targetIndex < draggedIndex) {
        list.insertBefore(draggedItem, target);
      } else {
        list.insertBefore(draggedItem, target.nextSibling);
      }
    });
  </script>
</body>
</html>
```

---

## 34. 完整实战：下拉菜单组件

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <title>下拉菜单</title>
  <style>
    .dropdown { position: relative; display: inline-block; }
    .dropdown-btn {
      padding: 8px 16px; border: 1px solid #cbd5e1; border-radius: 6px;
      background: #fff; cursor: pointer; font-size: 14px;
    }
    .dropdown-menu {
      display: none; position: absolute; top: 100%; left: 0;
      min-width: 160px; background: #fff; border: 1px solid #e2e8f0;
      border-radius: 8px; box-shadow: 0 8px 24px rgba(0,0,0,0.1);
      margin-top: 4px; z-index: 100;
    }
    .dropdown-menu.show { display: block; }
    .dropdown-menu a {
      display: block; padding: 10px 16px; color: #334155;
      text-decoration: none; font-size: 14px;
    }
    .dropdown-menu a:hover { background: #f1f5f9; }
  </style>
</head>
<body>
  <div class="dropdown">
    <button class="dropdown-btn" id="dropdownBtn">操作 ▼</button>
    <div class="dropdown-menu" id="dropdownMenu">
      <a href="#">编辑</a>
      <a href="#">分享</a>
      <a href="#">删除</a>
    </div>
  </div>

  <script>
    const btn = document.getElementById("dropdownBtn");
    const menu = document.getElementById("dropdownMenu");

    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      menu.classList.toggle("show");
    });

    // 点击菜单外部关闭
    document.addEventListener("click", (e) => {
      if (!btn.contains(e.target) && !menu.contains(e.target)) {
        menu.classList.remove("show");
      }
    });

    // ESC 关闭
    document.addEventListener("keydown", (e) => {
      if (e.key === "Escape") menu.classList.remove("show");
    });

    // 点击菜单项
    menu.addEventListener("click", (e) => {
      if (e.target.tagName === "A") {
        e.preventDefault();
        console.log("点击了:", e.target.textContent);
        menu.classList.remove("show");
      }
    });
  </script>
</body>
</html>
```

---

## 35. Intersection Observer 入门（懒加载、无限滚动）

```js
// 1. 图片懒加载
const images = document.querySelectorAll("img[data-src]");
const imgObserver = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      const img = entry.target;
      img.src = img.dataset.src;
      img.removeAttribute("data-src");
      imgObserver.unobserve(img); // 已经加载了就停止观察
    }
  });
}, { rootMargin: "100px" }); // 提前 100px 开始加载

images.forEach(img => imgObserver.observe(img));

// 2. 无限滚动
const sentinel = document.getElementById("sentinel");
const list = document.getElementById("list");
let page = 1;

const scrollObserver = new IntersectionObserver((entries) => {
  if (entries[0].isIntersecting) {
    page++;
    // fetchMoreData(page).then(data => renderList(data));
    console.log(`触发加载第 ${page} 页`);
  }
});

scrollObserver.observe(sentinel);
```

HTML 中观察哨兵元素：
```html
<ul id="list">...</ul>
<div id="sentinel" style="height:1px"></div>
```

---

## 36. BOM 完整 API 速查

### window — 全局对象

```js
window.innerWidth;      // 视口宽度（含滚动条）
window.innerHeight;     // 视口高度
window.scrollX;         // 水平滚动位置
window.scrollY;         // 垂直滚动位置
window.scrollTo({ top: 0, behavior: "smooth" }); // 平滑滚动
window.open(url, "_blank"); // 打开新窗口
window.print();         // 打印
```

### location — 地址信息

```js
location.href;          // 完整 URL
location.host;          // 域名 + 端口
location.pathname;      // 路径
location.search;        // ?参数
location.hash;          // #哈希
location.reload();      // 刷新
location.href = "/new"; // 跳转（会留下历史记录）
location.replace("/new"); // 替换（不留历史记录）
```

### history — 历史控制

```js
history.back();         // 后退
history.forward();      // 前进
history.go(-2);         // 后退 2 页
history.pushState({ id: 1 }, "", "/page1"); // 添加历史记录（SPA 路由基础）
```

### navigator — 设备信息

```js
navigator.userAgent;    // 浏览器标识
navigator.language;     // 浏览器语言
navigator.onLine;       // 是否联网
navigator.clipboard.writeText("复制的内容"); // 剪贴板
```

### screen — 屏幕信息

```js
screen.width;           // 屏幕宽度
screen.height;          // 屏幕高度
```

---

## 37. 完整实战：待办清单（综合所有 DOM 技能）

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>待办清单</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: system-ui, sans-serif;
      background: #f1f5f9;
      min-height: 100vh; display: flex; justify-content: center; padding-top: 60px;
    }
    .app {
      width: 100%; max-width: 480px; padding: 0 16px;
    }
    h1 { font-size: 1.5rem; margin-bottom: 20px; text-align: center; }
    .input-row { display: flex; gap: 8px; margin-bottom: 20px; }
    .input-row input {
      flex: 1; padding: 10px 12px; border: 1px solid #cbd5e1;
      border-radius: 8px; font-size: 15px; outline: none;
    }
    .input-row input:focus { border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,0.1); }
    .input-row button {
      padding: 10px 20px; border: none; border-radius: 8px;
      background: #6366f1; color: #fff; cursor: pointer; font-size: 15px;
    }
    .input-row button:hover { background: #4f46e5; }
    .todo-list { list-style: none; }
    .todo-item {
      display: flex; align-items: center; gap: 10px;
      padding: 12px 16px; margin-bottom: 8px;
      background: #fff; border-radius: 8px; border: 1px solid #e2e8f0;
      transition: all 0.2s;
    }
    .todo-item.done { opacity: 0.6; }
    .todo-item.done .todo-text { text-decoration: line-through; color: #94a3b8; }
    .todo-check { width: 20px; height: 20px; cursor: pointer; accent-color: #6366f1; }
    .todo-text { flex: 1; font-size: 15px; }
    .todo-del {
      background: none; border: none; color: #ef4444; cursor: pointer;
      font-size: 18px; padding: 4px 8px; border-radius: 4px;
    }
    .todo-del:hover { background: #fef2f2; }
    .footer { display: flex; justify-content: space-between; align-items: center; margin-top: 16px; font-size: 14px; color: #64748b; }
    .footer button { background: none; border: 1px solid #e2e8f0; padding: 6px 12px; border-radius: 6px; cursor: pointer; font-size: 13px; }
    .empty { text-align: center; color: #94a3b8; padding: 40px 0; }
  </style>
</head>
<body>
  <div class="app">
    <h1>📝 待办清单</h1>
    <div class="input-row">
      <input type="text" id="todoInput" placeholder="输入任务，回车添加..." />
      <button id="addBtn">添加</button>
    </div>
    <ul class="todo-list" id="todoList"></ul>
    <div class="footer">
      <span id="count">0 项</span>
      <button id="clearDone">清除已完成</button>
    </div>
  </div>

  <script>
    const input = document.getElementById("todoInput");
    const addBtn = document.getElementById("addBtn");
    const list = document.getElementById("todoList");
    const countEl = document.getElementById("count");
    const clearBtn = document.getElementById("clearDone");

    // 从 localStorage 读取
    let todos = JSON.parse(localStorage.getItem("todos") || "[]");
    render();

    // 添加任务
    function addTodo(text) {
      if (!text.trim()) return;
      todos.push({ id: Date.now(), text: text.trim(), done: false });
      save();
      render();
    }

    addBtn.addEventListener("click", () => { addTodo(input.value); input.value = ""; input.focus(); });
    input.addEventListener("keydown", (e) => { if (e.key === "Enter") { addTodo(input.value); input.value = ""; } });

    // 事件委托：勾选和删除
    list.addEventListener("click", (e) => {
      const li = e.target.closest(".todo-item");
      if (!li) return;
      const id = Number(li.dataset.id);

      if (e.target.classList.contains("todo-del")) {
        todos = todos.filter(t => t.id !== id);
      } else if (e.target.classList.contains("todo-check")) {
        const todo = todos.find(t => t.id === id);
        if (todo) todo.done = e.target.checked;
      }
      save();
      render();
    });

    // 清除已完成
    clearBtn.addEventListener("click", () => {
      todos = todos.filter(t => !t.done);
      save();
      render();
    });

    function save() {
      localStorage.setItem("todos", JSON.stringify(todos));
    }

    function render() {
      if (todos.length === 0) {
        list.innerHTML = '<li class="empty">暂无任务，添加一个吧 ✨</li>';
      } else {
        list.innerHTML = todos.map(t => `
          <li class="todo-item${t.done ? ' done' : ''}" data-id="${t.id}">
            <input type="checkbox" class="todo-check" ${t.done ? 'checked' : ''} />
            <span class="todo-text">${escapeHtml(t.text)}</span>
            <button class="todo-del">×</button>
          </li>
        `).join("");
      }
      countEl.textContent = `${todos.filter(t => !t.done).length} 项未完成`;
    }

    // 防 XSS：转义 HTML 特殊字符
    function escapeHtml(str) {
      const div = document.createElement("div");
      div.textContent = str;
      return div.innerHTML;
    }
  </script>
</body>
</html>
```

### 37.1 待办清单核心逻辑逐行读

| 行号/代码 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `JSON.parse(localStorage.getItem("todos") \|\| "[]")` | 读本地 JSON，无数据用空数组 | 不设 `\|\| "[]"` → parse null 报错 |
| `addTodo` 里 `Date.now()` 作 id | 简易唯一 id | 极快连点可能重复（生产用 uuid） |
| `list.addEventListener("click", ...)` | 父级委托处理勾选/删除 | 绑在每个 `li` → 动态项要反复绑 |
| `e.target.closest(".todo-item")` | 向上找带 data-id 的行 | 点空白处 li 为 null，应 return |
| `Number(li.dataset.id)` | 读 HTML `data-id` 属性 | 忘写 data-id → NaN 找不到项 |
| `todos.filter(t => t.id !== id)` | 不可变删除 | 直接 splice 也行，但易与渲染不同步 |
| `innerHTML = todos.map(...).join("")` | 整表重绘 | 大数据量应文档片段优化；初学够用 |
| `escapeHtml(t.text)` | 防 XSS，用户输入当文本 | 直接插 `${text}` → 可注入 `<script>` |

### 37.2 与 Vue 08 / 计网 04 的衔接

- **Vue 08**：Axios 拉列表后同样要渲染 DOM 或交给 Vue 模板；本章 `render()` 相当于手写 `v-for`。
- **计网 04**：待办存 localStorage 不走网络；若改成 REST API，GET/POST 见 [HTTP 协议深入](../计算机网络/04-HTTP协议深入.md)，fetch 见 [09 章](./09-JavaScript异步编程网络请求与本地存储.md)。

---

## 38. 分级练习

**基础**：点击按钮切换段落显示/隐藏  
**进阶**：Tab 切换 + CSS 过渡  
**挑战**：事件委托版待办（增删改）+ localStorage 持久化

---

## 39. Elements 面板八步实操

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 用 Live Server 打开 §37 待办 HTML，F12 → **Elements** | 看到 `<ul id="list">` | script 是否在 body 底或 defer |
| 2 | 添加一条任务 | `ul` 下出现新 `li.todo-item` | Console 是否有报错 |
| 3 | 右键 `#list` → Break on → subtree modifications | 再添加时 JS 断在 `render` | 学会观察重绘时机 |
| 4 | 选中 `.todo-item`，右侧 Styles 改 `color` | 仅该行变色 | 选择器是否选错节点 |
| 5 | 勾选 checkbox，看 `li` 是否加 `done` class | class 列表变化 | 委托逻辑是否走到 |
| 6 | 刷新页面 | 任务仍在 | localStorage 键名 `todos` 是否存在 Application 面板 |
| 7 | Application → Local Storage 手动改 JSON | 刷新后界面跟着变 | JSON 格式非法会 parse 失败 |
| 8 | 配合 [09 fetch](./09-JavaScript异步编程网络请求与本地存储.md) 把列表改成接口数据 | Network 有 XHR | CORS 见 [10 章](./10-浏览器HTTP网络与Web基础.md) |

---

## 40. FAQ

**Q1：`getElementById` 和 `querySelector` 用哪个？**  
现代优先 `querySelector` / `querySelectorAll`，与 CSS 选择器一致；id 极快但 API 不统一。

**Q2：`innerHTML` 和 `textContent`？**  
要 HTML 结构用前者；纯文本或防 XSS 用后者。用户输入**永远**先转义再插 HTML。

**Q3：如何移除事件监听？**  
`removeEventListener` 必须传**同一函数引用**；匿名函数无法移除，需命名函数或 AbortController。

**Q4：为什么点按钮没反应？**  
常见：DOM 未加载就 querySelector 得到 null；script 在 head 且无 defer；选择器写错。

**Q5：事件委托为什么能管「后添加」的元素？**  
监听绑在父级，子元素后来插入也会冒泡到父级——单独绑则新元素无监听。

**Q6：`preventDefault` 和 `stopPropagation` 区别？**  
前者阻止默认行为（如表单提交、链接跳转）；后者阻止冒泡，父级收不到事件。

**Q7：DOM 和 Virtual DOM 什么关系？**  
Vue/React 先在内存里 diff，再批量改真实 DOM；本章是直接改真实 DOM 的「手动挡」。

**Q8：`DOMContentLoaded` 和 `load`？**  
前者 HTML 解析完即可操作 DOM；后者还要等图片等资源。脚本用 `defer` 类似 DOMContentLoaded 后执行。

**Q9：修改样式用 style 还是 class？**  
多状态、响应式、动画优先 **class**；单次微调可用 style。Vue 用 `:class` 绑定。

**Q10：localStorage 和 sessionStorage？**  
localStorage 持久；sessionStorage 关标签页即清。详见 [09 章](./09-JavaScript异步编程网络请求与本地存储.md)。

**Q11：和 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 分工？**  
本章管「数据到了怎么画页面」；Vue 08 管「数据从哪来（Axios/HTTP）」。

**Q12：Network 200 但页面空白？**  
用 Elements 看节点是否插入；Console 看 JS 报错；响应 JSON 结构是否和 `map` 字段一致（对照 [计网 04](../计算机网络/04-HTTP协议深入.md)）。

---

## 41. 练习建议

1. 点击按钮切换文字与 class
2. Tab 切换组件
3. 动态添加/删除列表项（事件委托）
4. 表单实时校验 + 提交拦截
5. 简易待办清单（为 Vue 项目打基础）
6. 把待办改成从 jsonplaceholder 拉用户名为标题（预习 09 + Vue 08）

---

## 42. 学完标准

- 理解 DOM 树，会用多种方式选中、创建、插入、删除元素
- 会用 `classList` 管理状态，知道 `textContent` 与 `innerHTML` 区别
- 熟练 `addEventListener`，理解冒泡、委托、`preventDefault`
- 知道脚本加载时机，避免 DOM 未就绪
- 能独立完成 Tab 切换或待办列表交互
- 会用 Elements 面板验证渲染结果

---

## 43. 闭卷自测

1. DOM 是什么？JavaScript 通过 DOM 能对页面做哪四类事？
2. `querySelector(".item")` 与 `querySelectorAll(".item")` 返回值有何不同？
3. 事件冒泡是什么？委托利用了什么特性？
4. 表单提交时为什么要 `event.preventDefault()`？
5. `script` 放 head 且无 defer/async 会有什么问题？
6. 动态列表为什么推荐事件委托？
7. `textContent` 赋值能否防 XSS？`innerHTML` 呢？
8. **动手**：写三行 JS 给 `#btn` 绑 click，点击后 `#msg` 文字变「已点击」。
9. **动手**：`ul` 委托，点 `.del` 删除对应 `li`（可用 `closest`）。
10. **综合**：Vue 08 请求成功后页面不更新，你会用 Elements 和 Console 各查什么？

### 43.1 自测参考答案

1. 文档对象模型；查找、修改、创建、删除节点。
2. 前者返回第一个匹配 Element 或 null；后者返回 NodeList（类数组，可 forEach）。
3. 事件从目标向上传；委托在父级统一监听，靠 target 识别子元素。
4. 阻止浏览器默认整页刷新提交，改由 JS 发 fetch/Ajax。
5. 脚本执行时 DOM 可能未解析，querySelector 得到 null。
6. 性能更好、动态节点免重复绑定、代码更集中。
7. textContent 当纯文本安全；innerHTML 解析 HTML，不可信内容危险。
8. `btn.addEventListener('click',()=>{msg.textContent='已点击'})`（先取元素）。
9. `ul.addEventListener('click',e=>{if(e.target.matches('.del'))e.target.closest('li')?.remove()})`。
10. Elements 看列表 DOM 是否增加；Console 看渲染函数是否报错、数据结构是否匹配。

---

## 44. 费曼检验

3 分钟向朋友解释「网页怎么响应点击」：

1. **HTML 变成 DOM 树**：标签变成 JS 能摸到的对象，改对象页面就变。
2. **监听 = 登记门铃**：`addEventListener` 告诉浏览器「这件事发生时叫我」。
3. **委托 = 前台统一接待**：列表再长也只绑一个父级——Vue 列表、Axios 拉数据后渲染，思路一样。

> 下一章：[09 — 异步编程与网络请求](./09-JavaScript异步编程网络请求与本地存储.md)。DOM 会动了，接下来等数据从服务器来。深入 HTTP 见 [10 章](./10-浏览器HTTP网络与Web基础.md) 与 [计网 04](../计算机网络/04-HTTP协议深入.md)；Vue 封装见 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md)。
