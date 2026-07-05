# JavaScript 异步编程、网络请求与本地存储

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0 导读、fetch 步骤表、FAQ 12 题、闭卷自测、费曼检验；链 计网 04 / Vue 08 -->

## 0. 读前导读（零基础也能跟上）

> **读者假设**：你已会 [07 章](./07-JavaScript流程控制函数对象数组与ES6基础.md) 数据处理与 [08 章](./08-JavaScript-DOM-BOM与事件机制.md) 改 DOM。本章解决「**等服务器回话**」——加载中、失败重试、本地记 token，是 [Vue 08 Axios 联调](../Vue/08-Axios网络请求与前后端联调.md) 的直接前置。

### 0.1 用一句话弄懂本章

**一句话**：JavaScript **单线程**却用 **Promise / async-await** 和**事件循环**处理网络等待；用 **fetch** 对话后端，用 **localStorage** 在浏览器存登录态——和 [计网 04 HTTP](../计算机网络/04-HTTP协议深入.md) 讲的是同一根网线两端。

**生活类比**：

| 本章概念 | 类比 | 对应章节 |
|---------|------|----------|
| **异步** | 点外卖继续做别的事，到了再收 | 事件循环 §7a |
| **Promise** | 外卖订单号（pending→送达/退单） | §4～§6 |
| **async/await** | 用「等」字写订单流程，读起来像同步 | §7 |
| **fetch** | 前台打电话给后厨要菜单 | [计网 04](../计算机网络/04-HTTP协议深入.md) 定协议 |
| **localStorage** | 浏览器抽屉存会员卡号 | Vue 08 存 JWT |

---

### 0.2 你需要提前知道什么

| 水平 | 建议 |
|------|------|
| 刚学完 08 | 从 §2 同步/异步开始；§7 fetch 是核心 |
| 被 CORS 报错卡住 | 先读本章 §7c，再读 [10 章 CORS](./10-浏览器HTTP网络与Web基础.md)，深入 [计网 04](../计算机网络/04-HTTP协议深入.md) |
| 目标 Vue 08 | 本章 fetch 练熟后，Vue 08 把 fetch 换成 axios + 拦截器 |
| 后端 Java 04 | 接口返回 JSON 结构要和本章 `await res.json()` 字段一致 |

---

### 0.3 本章知识地图（☐→☑）

- [ ] 能解释同步 vs 异步、宏任务 vs 微任务（§7a 顺序题）
- [ ] 会写 Promise 链与 `async/await` + `try/catch`
- [ ] 会用 `fetch` GET/POST，检查 `res.ok`
- [ ] 实现加载中 / 成功 / 失败 / 空数据 四态 UI（§24）
- [ ] 会用 localStorage 存读 JSON，知道 sessionStorage 区别
- [ ] 知道防抖节流用途（§23）
- [ ] 闭卷自测 ≥ 8/10

---

### 0.4 建议学习时长与节奏

| 阶段 | 时间 |
|------|------|
| Promise + async/await §4～§7 | 3 小时 |
| 事件循环 §7a | 1 小时 |
| fetch 实战 §7c～§24 | 3 小时 |
| localStorage + 防抖 §7e、§23 | 1.5 小时 |
| 自测 | 0.5 小时 |

---

### 0.5 学完本章你能做什么（可验证）

1. 用 fetch 拉 jsonplaceholder 用户列表并渲染到 [08 章](./08-JavaScript-DOM-BOM与事件机制.md) 的 `#list`。
2. 向朋友解释：`await fetch()` 缺 `await res.json()` 会得到什么。
3. 说明 404 时 fetch **不会**自动 throw，和 [计网 04 状态码](../计算机网络/04-HTTP协议深入.md) 如何对应。

---

### 0.6 核心术语三件套

**术语（Promise 承诺对象）**：代表未来才会知道结果的异步操作；状态 pending / fulfilled / rejected。
**生活类比**：网购订单——下单(pending)、签收(fulfilled)、退货(rejected) 三态。
**为什么重要**：fetch、axios、数据库操作都返回 Promise；不会 Promise 无法写现代前端。
**本章用到的地方**：§4～§7b。

**术语（async/await）**：在 async 函数内用 await 暂停等待 Promise 完成，写法像同步。
**生活类比**：排队取号后可以刷手机，叫号时再办业务——代码顺序仍从上到下读。
**为什么重要**：Vue 08 的 `onMounted(async () => { await axios.get(...) })` 标准写法。
**本章用到的地方**：§7；[Vue 08](../Vue/08-Axios网络请求与前后端联调.md)。

**术语（fetch API）**：浏览器内置发 HTTP 请求的函数；返回 Promise<Response>。
**生活类比**：不装第三方 App，直接用系统电话打给服务器——axios 是「加功能的外挂电话」。
**为什么重要**：理解 fetch 后学 axios 只是加拦截器、baseURL；HTTP 本质不变，见 [计网 04](../计算机网络/04-HTTP协议深入.md)。
**本章用到的地方**：§7c～§24。

---

## 1. 为什么这一份很关键

真正的前端不是只会改页面文字和颜色。

前端经常要做这些事：

- 请求后端接口
- 等待数据返回
- 处理加载状态
- 处理失败重试
- 本地保存一些数据

这些能力离不开：

- 异步编程
- 网络请求
- 本地存储

## 2. 什么是同步和异步

### 同步

代码一行一行按顺序执行，前一个不结束，后一个不开始。

### 异步

有些任务不需要立刻阻塞等待，可以先发起，等结果回来再处理。

例如：

- 网络请求
- 定时器
- 文件读取

## 3. 回调函数基础认知

早期 JavaScript 很多异步逻辑靠回调处理。

```js
setTimeout(() => {
  console.log("延迟执行");
}, 1000);
```

这里传进去的函数就是回调。

## 4. Promise 是什么

Promise 是 JavaScript 处理异步的重要机制。

你可以先把它理解成：

- 对未来某个结果的承诺对象

它有三种常见状态：

- pending
- fulfilled
- rejected

## 5. Promise 基础示例

```js
const p = new Promise((resolve, reject) => {
  const ok = true;
  if (ok) {
    resolve("成功");
  } else {
    reject("失败");
  }
});

p.then((res) => {
  console.log(res);
}).catch((err) => {
  console.log(err);
});
```

## 6. `then`、`catch`、`finally`

### `then`

处理成功结果。

### `catch`

处理失败结果。

### `finally`

不管成功失败都会执行。

## 7. `async/await` — 现代异步核心

这是现代前端处理异步最常用的方式。

### 7.1 基本语法

```js
// async 函数总是返回 Promise
async function getData() {
  return "数据";  // 自动包装成 Promise.resolve("数据")
}

// await 等待 Promise 完成
async function run() {
  const data = await getData();
  console.log(data); // "数据"
}
```

### 7.2 为什么它好用

- 写起来更像同步代码，不需要 `.then()` 链
- 错误处理用 `try/catch`，更自然
- 可读性远超 Promise 链

```js
// Promise 链（回调地狱残余）
fetchUser()
  .then(user => fetchOrders(user.id))
  .then(orders => processOrders(orders))
  .catch(err => console.error(err));

// async/await（清晰如瀑布）
async function load() {
  try {
    const user = await fetchUser();
    const orders = await fetchOrders(user.id);
    const result = await processOrders(orders);
    console.log(result);
  } catch (err) {
    console.error(err);
  }
}
```

### 7.3 async/await 常见错误模式

```js
// ❌ 错误：await 只能用在 async 函数里
// const data = await fetch(); // 顶层直接使用会报错

// ✅ 正确：包在 async 函数中
(async () => {
  const data = await fetch();
})();

// ✅ ES2022+ 顶层 await（模块中）
// const data = await fetch(); // 在 <script type="module"> 中可用

// ❌ 忘记 await
async function bad() {
  const data = fetch("/api");   // data 是 Promise，不是数据！
  console.log(data);            // Promise {<pending>}
}

// ✅ 正确
async function good() {
  const res = await fetch("/api");
  const data = await res.json();
  console.log(data);
}

// ❌ 顺序 await 浪费性能（两个不相关的请求可以并行）
async function slow() {
  const user = await fetchUser();    // 等 500ms
  const posts = await fetchPosts();  // 再等 500ms = 共 1000ms
}

// ✅ Promise.all 并行
async function fast() {
  const [user, posts] = await Promise.all([
    fetchUser(),
    fetchPosts(),
  ]); // 同时发起，共约 500ms
}
```

---

## 7a. 事件循环 Event Loop（理解异步的关键）

JavaScript 是单线程语言，但能处理异步任务，靠的就是事件循环。

### 简化模型

```
调用栈（同步代码）
    ↓ 遇到异步任务
Web API（定时器、网络请求等）
    ↓ 完成后
任务队列（宏任务：setTimeout、fetch回调）
微任务队列（Promise.then、MutationObserver）
    ↓
事件循环：先清空微任务队列 → 取一个宏任务 → 再清空微任务队列 → ...
```

### 执行顺序示例

```js
console.log("1 同步");

setTimeout(() => console.log("2 setTimeout"), 0);

Promise.resolve().then(() => console.log("3 Promise.then"));

console.log("4 同步");

// 输出顺序：1 → 4 → 3 → 2
// 解释：同步代码先执行 → 微任务(Promise.then) → 宏任务(setTimeout)
```

### 宏任务 vs 微任务

| 类型 | 常见来源 | 优先级 |
|------|----------|--------|
| 宏任务 | `setTimeout`、`setInterval`、`fetch` 回调、DOM 事件 | 低（每次取一个） |
| 微任务 | `Promise.then/catch/finally`、`queueMicrotask` | 高（清空为止） |

**记忆口诀**：同步 → 微任务 → 宏任务 → 微任务 → 重渲染 → 循环。

### 7a.1 事件循环自测小题（必做）

```js
console.log("A");
setTimeout(() => console.log("B"), 0);
Promise.resolve().then(() => console.log("C"));
queueMicrotask(() => console.log("D"));
console.log("E");
// 答案：A E C D B
```

| 步骤 | 你的动作 | 预期 | 若不对 |
|------|----------|------|--------|
| 1 | Console 粘贴上段代码 | 顺序 A E C D B | 复习宏/微任务表 |
| 2 | 把 setTimeout 改成 0ms 与 100ms 两个 | 仍先跑完微任务 | Promise.then 优先于 timer |
| 3 | 在 then 里再嵌套 then | 嵌套微任务在同轮清空 | 见 [计网 04](../计算机网络/04-HTTP协议深入.md) 外，JS 引擎章节 |

理解事件循环后，`await fetch` 之后的代码为何「让出」线程就清楚了——[Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 里 `await axios.get` 同理。

---

## 7b. Promise 静态方法

```js
// Promise.all([...]) — 全部成功才成功，一个失败整体失败
const [user, posts] = await Promise.all([
  fetchUser(1),
  fetchPosts(1),
]);

// Promise.allSettled([...]) — 等所有完成，不管成功失败
const results = await Promise.allSettled([
  fetch("/api/user"),
  fetch("/api/backup"), // 可能 404，但不影响其他
]);
results.forEach(r => {
  if (r.status === "fulfilled") console.log("成功:", r.value);
  else console.log("失败:", r.reason);
});

// Promise.race([...]) — 第一个完成的（无论成功失败）就返回
const timeout = new Promise((_, reject) =>
  setTimeout(() => reject(new Error("超时")), 5000)
);
const data = await Promise.race([fetch("/api"), timeout]);

// Promise.any([...]) — 第一个成功的，全失败才 reject
const data = await Promise.any([
  fetch("https://cdn1.example.com/data"),
  fetch("https://cdn2.example.com/data"),
]);

// Promise.resolve / Promise.reject（快速创建 Promise）
const cached = Promise.resolve({ name: "小明" }); // 已完成的 Promise
const failed = Promise.reject(new Error("失败"));  // 已失败的 Promise
```

---

## 7c. fetch 完整配置与错误处理

### fetch 完整参数

```js
const res = await fetch("/api/data", {
  method: "POST",              // GET | POST | PUT | DELETE | PATCH
  headers: {
    "Content-Type": "application/json",
    "Authorization": "Bearer " + token,
  },
  body: JSON.stringify({ a: 1 }), // GET/HEAD 不能有 body
  mode: "cors",                // cors | no-cors | same-origin
  credentials: "include",      // omit | same-origin | include（带 cookie）
  cache: "no-cache",           // default | no-cache | reload | force-cache
  signal: abortController.signal, // 用于取消请求
});
```

### fetch 错误处理（重要！fetch 只 reject 网络错误）

```js
// ❌ 错误做法：以为 !res.ok 会自动 reject
try {
  const res = await fetch("/api");
  const data = await res.json(); // 可能 res 是 404！
} catch {
  // 这里只捕获网络断连等，不捕获 404/500！
}

// ✅ 正确做法：手动检查 res.ok
async function request(url, options) {
  const res = await fetch(url, options);

  if (!res.ok) {
    const errorBody = await res.text().catch(() => "");
    throw new Error(`HTTP ${res.status}: ${errorBody || res.statusText}`);
  }

  return res.json();
}
```

### 请求取消 AbortController

```js
const controller = new AbortController();
const timeoutId = setTimeout(() => controller.abort(), 5000);

try {
  const res = await fetch("/api/slow", {
    signal: controller.signal,
  });
  clearTimeout(timeoutId);
  const data = await res.json();
} catch (err) {
  if (err.name === "AbortError") {
    console.log("请求被取消（超时）");
  }
}
```

---

## 7d. 完整实战：带缓存策略的用户列表页

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <title>用户列表（含缓存）</title>
  <style>
    body { font-family: system-ui, sans-serif; padding: 24px; }
    #app { max-width: 600px; margin: 0 auto; }
    .user-card {
      padding: 16px; margin-bottom: 12px;
      border: 1px solid #e2e8f0; border-radius: 8px;
    }
    .user-card h3 { margin: 0 0 4px; }
    .user-card p { margin: 0; color: #64748b; font-size: 14px; }
    #status { text-align: center; padding: 40px; color: #64748b; }
    .error { color: #ef4444; text-align: center; padding: 40px; }
    button {
      padding: 8px 16px; border: none; border-radius: 6px;
      background: #6366f1; color: #fff; cursor: pointer; margin-bottom: 16px;
    }
  </style>
</head>
<body>
  <div id="app">
    <button id="refresh">🔄 刷新列表</button>
    <div id="content"><div id="status">加载中...</div></div>
  </div>
  <script>
    const content = document.getElementById("content");
    const refreshBtn = document.getElementById("refresh");

    const CACHE_KEY = "user_list_cache";
    const CACHE_TIME_KEY = "user_list_cache_time";
    const CACHE_DURATION = 5 * 60 * 1000; // 5 分钟

    // 渲染函数
    function renderLoading() { content.innerHTML = '<div id="status">加载中...</div>'; }
    function renderError(msg) { content.innerHTML = `<div class="error">${msg}</div>`; }
    function renderEmpty() { content.innerHTML = '<div id="status">暂无数据</div>'; }
    function renderUsers(users) {
      content.innerHTML = users.length === 0
        ? '<div id="status">暂无数据</div>'
        : users.map(u => `
          <div class="user-card">
            <h3>${escapeHtml(u.name)}</h3>
            <p>📧 ${escapeHtml(u.email)} | 📞 ${escapeHtml(u.phone || 'N/A')}</p>
          </div>
        `).join("");
    }

    function escapeHtml(str) {
      const div = document.createElement("div");
      div.textContent = str;
      return div.innerHTML;
    }

    // 缓存读写
    function getCache() {
      const data = localStorage.getItem(CACHE_KEY);
      const time = localStorage.getItem(CACHE_TIME_KEY);
      if (!data || !time) return null;
      if (Date.now() - Number(time) > CACHE_DURATION) return null;
      return JSON.parse(data);
    }

    function setCache(users) {
      localStorage.setItem(CACHE_KEY, JSON.stringify(users));
      localStorage.setItem(CACHE_TIME_KEY, String(Date.now()));
    }

    // 加载逻辑：先读缓存再请求
    async function loadUsers(forceRefresh = false) {
      // 1. 先显示缓存（快速展示）
      if (!forceRefresh) {
        const cached = getCache();
        if (cached) {
          renderUsers(cached);
          console.log("✅ 来自缓存，发送后台请求刷新...");
        } else {
          renderLoading();
        }
      } else {
        renderLoading();
      }

      // 2. 请求接口
      try {
        const res = await fetch("https://jsonplaceholder.typicode.com/users");
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const users = await res.json();

        setCache(users);
        renderUsers(users);
        console.log("✅ 数据已更新，缓存已刷新");
      } catch (err) {
        console.error("请求失败:", err);
        // 如果有缓存就用缓存兜底
        const cached = getCache();
        if (cached) {
          renderUsers(cached);
          console.warn("⚠️ 网络失败，显示过期缓存");
        } else {
          renderError(`加载失败：${err.message}，请重试`);
        }
      }
    }

    // 初始化
    loadUsers();
    refreshBtn.addEventListener("click", () => loadUsers(true));
  </script>
</body>
</html>
```

---

## 7e. localStorage 封装（带过期时间）

```js
const storage = {
  set(key, value, ttlMs = null) {
    const item = {
      value,
      ...(ttlMs ? { expires: Date.now() + ttlMs } : {}),
    };
    localStorage.setItem(key, JSON.stringify(item));
  },

  get(key, fallback = null) {
    const raw = localStorage.getItem(key);
    if (!raw) return fallback;
    try {
      const item = JSON.parse(raw);
      if (item.expires && Date.now() > item.expires) {
        localStorage.removeItem(key);
        return fallback;
      }
      return item.value ?? fallback;
    } catch {
      return fallback;
    }
  },

  remove(key) {
    localStorage.removeItem(key);
  },

  clear() {
    localStorage.clear();
  },
};

// 使用示例
storage.set("user", { name: "小明" }, 30 * 60 * 1000); // 30分钟过期
const user = storage.get("user", { name: "游客" });
```

---

## 7f. 文件上传 fetch + FormData

```js
async function uploadFile(file) {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("description", "用户头像");

  try {
    const res = await fetch("/api/upload", {
      method: "POST",
      body: formData,
      // ⚠️ 不要手动设置 Content-Type！浏览器会自动设 multipart/form-data + boundary
    });
    if (!res.ok) throw new Error(`上传失败: ${res.status}`);
    return await res.json();
  } catch (err) {
    console.error(err);
    throw err;
  }
}

// 配合 input 使用
// <input type="file" id="fileInput" />
document.getElementById("fileInput").addEventListener("change", async (e) => {
  const file = e.target.files[0];
  if (!file) return;
  if (file.size > 5 * 1024 * 1024) {
    alert("文件不能超过 5MB");
    return;
  }
  const result = await uploadFile(file);
  console.log("上传成功:", result);
});

## 23. 防抖与节流完整实现

### 防抖 debounce

```js
function debounce(fn, delay = 300) {
  let timer = null;
  return function (...args) {
    clearTimeout(timer);
    timer = setTimeout(() => fn.apply(this, args), delay);
  };
}

const searchInput = document.querySelector("#search");
searchInput.addEventListener(
  "input",
  debounce((e) => {
    console.log("搜索：", e.target.value);
  }, 500)
);
```

### 节流 throttle

```js
function throttle(fn, interval = 200) {
  let last = 0;
  return function (...args) {
    const now = Date.now();
    if (now - last >= interval) {
      last = now;
      fn.apply(this, args);
    }
  };
}

window.addEventListener(
  "scroll",
  throttle(() => console.log("滚动位置", window.scrollY), 200)
);
```

---

## 24. 完整实战：带状态的列表请求页

```html
<div id="app">
  <p id="status">加载中...</p>
  <ul id="list"></ul>
</div>
```

```js
async function loadUsers() {
  const statusEl = document.querySelector("#status");
  const listEl = document.querySelector("#list");

  try {
    statusEl.textContent = "加载中...";
    const res = await fetch("https://jsonplaceholder.typicode.com/users");
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const users = await res.json();

    if (users.length === 0) {
      statusEl.textContent = "暂无数据";
      return;
    }

    statusEl.textContent = "";
    listEl.innerHTML = users
      .map((u) => `<li>${u.name} - ${u.email}</li>`)
      .join("");
  } catch (err) {
    statusEl.textContent = "加载失败，请稍后重试";
    console.error(err);
  }
}

loadUsers();
```

**公开练习 API**：https://jsonplaceholder.typicode.com

### 24.1 loadUsers 逐行读

| 行号/代码 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `querySelector("#status")` | 取状态 DOM，配合 [08 章](./08-JavaScript-DOM-BOM与事件机制.md) | 选择器错 → null.textContent 报错 |
| `statusEl.textContent = "加载中..."` | 同步更新 UI，用户即时反馈 | 不写 → 白屏不知在等 |
| `await fetch(...)` | 发起 GET，返回 Promise<Response> | 漏 await → res 是 Promise |
| `if (!res.ok) throw ...` | 404/500 主动进 catch | 漏检查 → 把错误页当 JSON parse |
| `await res.json()` | 读 body 流解析为数组 | 只 await 一次 → data 仍是 ReadableStream |
| `users.length === 0` | 空数据四态之一 | 只处理成功有数据 → 空列表 UI 怪异 |
| `list.innerHTML = users.map(...)` | 批量渲染 li | map 漏 return → undefined 列表 |
| `catch` 改 status 文案 | 网络/CORS/parse 统一兜底 | 只 console.error → 用户不知失败 |

### 24.2 fetch POST 登录模拟（对接 Java 04 / Vue 08）

```js
async function login(username, password) {
  const res = await fetch("https://jsonplaceholder.typicode.com/posts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const data = await res.json();
  localStorage.setItem("auth", JSON.stringify({ token: "demo-" + data.id }));
  return data;
}
```

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | Network 勾选 Preserve log，调用 `login('a','b')` | POST 201，Request Payload 有 JSON | method/body 是否遗漏 |
| 2 | 看 Request Headers 的 Content-Type | `application/json` | 与 [计网 04](../计算机网络/04-HTTP协议深入.md) body 格式一致 |
| 3 | Application → localStorage 键 `auth` | 有 token 字符串 | setItem 是否执行 |
| 4 | 对比 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) axios.post 写法 | 字段相同，axios 少写 stringify | 理解语法糖层次 |

---

## 25. 异步编程常见错误

### 25.1 忘记 `await`

```js
async function bad() {
  const data = fetch("/api"); // 这是 Promise，不是数据！
}
async function good() {
  const res = await fetch("/api");
  const data = await res.json();
}
```

### 25.2 不处理错误

必须用 `try/catch` 或 `.catch()`。

### 25.3 只处理成功

空数组、404、500 都要有 UI 反馈。

---

## 26. 分级练习

**基础**：用 `setTimeout` 模拟 2 秒后打印「完成」  
**进阶**：用 `fetch` 请求 jsonplaceholder 的用户列表并渲染  
**挑战**：搜索框 + 防抖 + 请求 + 加载/空/错三态

---

## 27. FAQ

**Q1：Promise 和 async/await 什么关系？**  
`async/await` 是 Promise 的语法糖；每个 async 函数返回 Promise。

**Q2：localStorage 能存对象吗？**  
要 `JSON.stringify` 存、`JSON.parse` 取；容量约 5MB，同源限制。

**Q3：跨域怎么解决？**  
后端 CORS 响应头、开发代理、生产 Nginx 转发。详见 [10 章](./10-浏览器HTTP网络与Web基础.md) 与 [计网 04](../计算机网络/04-HTTP协议深入.md)。

**Q4：fetch 404 会进 catch 吗？**  
**不会**。只有网络错误才 reject；必须 `if (!res.ok) throw ...`（§7c）。

**Q5：为什么要 `await res.json()` 两次 await？**  
第一次等响应头/状态；第二次等 body 流解析成 JS 对象——两步都是异步。

**Q6：事件循环输出顺序怎么记？**  
同步 → 清空微任务(Promise.then) → 一个宏任务(setTimeout) → 再微任务… §7a 例题必背。

**Q7：Promise.all 一个失败会怎样？**  
整体 reject；要部分成功用 `Promise.allSettled`。

**Q8：防抖和节流区别？**  
防抖：停止触发后才执行（搜索框）；节流：固定间隔最多执行一次（滚动）。

**Q9：token 放 localStorage 安全吗？**  
有 XSS 风险可被 JS 读走；HttpOnly Cookie 更安全。Web 安全系列会展开；初学 localStorage 够用。

**Q10：和 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 分工？**  
本章原生 fetch + 手写状态；Vue 08 用 axios、拦截器、组件内 loading 变量。

**Q11：GET 和 POST 在 fetch 里怎么写？**  
GET 默认；POST 设 `method: 'POST'` + `headers` + `body: JSON.stringify(data)`，对齐 [计网 04](../计算机网络/04-HTTP协议深入.md) Content-Type。

**Q12：sessionStorage 何时用？**  
仅当前标签页有效的临时数据（如表单草稿）；关页即清；登录 token 通常 localStorage 或 Cookie。

---

## 28. 练习建议

1. 用 `fetch` 请求公开接口渲染列表
2. 做登录表单 POST（可用 mock 接口）
3. token 存 localStorage，刷新后仍能读取
4. 实现加载中 / 成功 / 失败三态 UI

---

## 28.1 fetch GET 列表手把手步骤

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 08 章 HTML 加 `<p id="status">` 和 `<ul id="list">` | 空列表 | DOM id 是否一致 |
| 2 | 写 `async function loadUsers()`，`status` 设「加载中…」 | 文字变化 | 函数是否被调用 |
| 3 | `const res = await fetch('https://jsonplaceholder.typicode.com/users')` | Network 出现请求 200 | 网络/CORS/URL 拼写 |
| 4 | `if (!res.ok) throw new Error(res.status)` | 故意改错 URL 时进 catch | 404 不会自动 throw |
| 5 | `const users = await res.json()`，Console 打印 users | 10 个用户对象数组 | 是否漏第二 await |
| 6 | `list.innerHTML = users.map(u => \`<li>${u.name}</li>\`).join('')` | 页面出现名字列表 | map 是否 return 字符串 |
| 7 | catch 里 `status.textContent = '加载失败'` | 断网时友好提示 | 是否包 try/catch |
| 8 | 对照 [计网 04](../计算机网络/04-HTTP协议深入.md) 看 Request Method、Status、Content-Type | Headers 与文档一致 | 复习 HTTP 报文结构 |

---

## 29. 学完标准

- 理解同步与异步，能解释事件循环的基本概念
- 会用 Promise 链和 `async/await`
- 会用 `fetch` 发 GET/POST，处理 JSON
- 知道 localStorage / sessionStorage 区别与用法
- 能实现防抖节流，知道跨域是浏览器策略
- 能在 Network 面板对照 [计网 04](../计算机网络/04-HTTP协议深入.md) 读状态码与头

---

## 30. 闭卷自测

1. JavaScript 为何说是单线程仍能处理异步？
2. 下面输出顺序：`console.log(1); setTimeout(()=>console.log(2),0); Promise.resolve().then(()=>console.log(3)); console.log(4);`
3. `async function f(){ return 1 }` 的返回值类型是什么？
4. fetch 返回 500 时，不检查 `res.ok` 会怎样？
5. localStorage 和 sessionStorage 在生命周期上有何区别？
6. 防抖适合搜索框还是滚动监听？节流呢？
7. JSON 请求 POST 时 Content-Type 通常写什么？
8. **动手**：写 fetch GET，把返回数组长度显示在 `#status`。
9. **动手**：把 `{ token: 'abc' }` 存入 localStorage 键 `auth`，刷新后读出。
10. **综合**：从 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 视角，axios 相比 fetch 多哪三层封装（提示：实例、拦截器、自动 JSON）？

### 30.1 自测参考答案

1. 事件循环：同步代码跑完，Web API 完成回调进任务队列再执行。
2. `1 4 3 2`（同步 → 微任务 → 宏任务）。
3. Promise（async 函数总是返回 Promise）。
4. 仍可能 `res.json()` 得到错误页 HTML/JSON，UI 误以为成功。
5. localStorage 持久直到清除；sessionStorage 关标签页即没。
6. 搜索框防抖；滚动节流。
7. `application/json`。
8. `const r=await fetch(url); const d=await r.json(); status.textContent=d.length`（加 ok 检查更好）。
9. `localStorage.setItem('auth', JSON.stringify({token:'abc'}))`；`JSON.parse(localStorage.getItem('auth'))`。
10. baseURL 实例、请求/响应拦截器、默认 transform JSON（了解即可，Vue 08 详讲）。

---

## 31. 费曼检验

3 分钟解释「前端怎么从服务器拿数据」：

1. **fetch 是一通电话**：浏览器按 [HTTP 规则](../计算机网络/04-HTTP协议深入.md) 问服务器要数据，回来是 Promise。
2. **await 是「等对方说完」**：拿到 Response 还要再 parse body 才是 JS 对象。
3. **localStorage 是浏览器记事本**：token 放这儿刷新还在——Vue 08 联调登录常用，但要注意 XSS。

> 下一章：[10 — 浏览器 HTTP 与 Web 基础](./10-浏览器HTTP网络与Web基础.md)（开发者视角 Network 面板）。Vue 项目联调：[Vue 08 Axios](../Vue/08-Axios网络请求与前后端联调.md)。
