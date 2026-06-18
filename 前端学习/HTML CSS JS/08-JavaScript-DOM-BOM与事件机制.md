# JavaScript DOM、BOM 与事件机制

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

---

## 38. 分级练习

**基础**：点击按钮切换段落显示/隐藏  
**进阶**：Tab 切换 + CSS 过渡  
**挑战**：事件委托版待办（增删改）+ localStorage 持久化

---

## 34. FAQ

**Q：`getElementById` 和 `querySelector` 用哪个？**  
现代开发优先 `querySelector` / `querySelectorAll`，统一用 CSS 选择器语法。

**Q：`innerHTML` 和 `textContent`？**  
插 HTML 结构用前者；只显示文字或防 XSS 用后者。

**Q：移除事件监听？**  
`removeEventListener` 需传入与 `add` 时**同一个函数引用**；匿名函数无法移除。

---

## 35. 练习建议

1. 点击按钮切换文字与 class
2. Tab 切换组件
3. 动态添加/删除列表项（事件委托）
4. 表单实时校验 + 提交拦截
5. 简易待办清单（为 12 篇实战打基础）

---

## 36. 学完标准

- 理解 DOM 树，会用多种方式选中、创建、插入、删除元素
- 会用 `classList` 管理状态，知道 `textContent` 与 `innerHTML` 区别
- 熟练 `addEventListener`，理解冒泡、委托、`preventDefault`
- 知道脚本加载时机，避免 DOM 未就绪
- 能独立完成 Tab 切换或待办列表交互
