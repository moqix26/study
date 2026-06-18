# JavaScript 异步编程、网络请求与本地存储

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

**Q：Promise 和 async/await 什么关系？**  
`async/await` 是 Promise 的语法糖，更易读。

**Q：localStorage 能存对象吗？**  
要 `JSON.stringify` 存、`JSON.parse` 取。

**Q：跨域怎么解决？**  
初学阶段用后端配置 CORS；开发时可用代理。详见 10 篇。

---

## 28. 练习建议

1. 用 `fetch` 请求公开接口渲染列表
2. 做登录表单 POST（可用 mock 接口）
3. token 存 localStorage，刷新后仍能读取
4. 实现加载中 / 成功 / 失败三态 UI

---

## 29. 学完标准

- 理解同步与异步，能解释事件循环的基本概念
- 会用 Promise 链和 `async/await`
- 会用 `fetch` 发 GET/POST，处理 JSON
- 知道 localStorage / sessionStorage 区别与用法
- 能实现防抖节流，知道跨域是浏览器策略
