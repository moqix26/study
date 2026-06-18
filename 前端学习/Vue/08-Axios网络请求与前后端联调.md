# Axios 网络请求与前后端联调

## 本章与上一章的关系

07 章 Pinia 管好了登录态（token）和购物车，但 `ProductList.vue` 里的商品仍是**写死的假数组**。后端 [Java 04-SpringBoot核心开发](../../后端学习/Java/04-SpringBoot核心开发.md) 或 [Python 04-FastAPI核心开发](../../后端学习/Python/04-FastAPI核心开发.md) 已经教你搭建了 `demo` / `demo-api` 项目，提供：

- `GET /api/users` — 用户列表
- `GET /api/users/{id}` — 单个用户
- `POST /api/users` — 新增用户
- `POST /api/login` — JWT 登录（04 章挑战练习）
- 统一返回 `Result<T>`：`{ code, message, data }`

这一章用 **Axios** 调这些接口，完成 **前后端联调**。商城场景下，我们暂时用 `/api/users` 数据**模拟商品列表**（id → 商品 id，name → 商品名），08 章重点是打通「请求 → 解析 → 渲染 → 登录带 token」全链路；05 章后端接 MySQL 后可换成真实 `/api/products`。

```mermaid
flowchart LR
    subgraph 前端["shop-vue (5173)"]
        V[Vue 组件]
        A[Axios 实例]
        P[Pinia userStore]
    end
    subgraph 后端["Spring Boot demo (8080)"]
        C[Controller]
        S[Service]
    end
    V --> A
    A --> P
    A -->|HTTP JSON| C
    C --> S
```

**这是 Vue 学习路线和后端学习路线的交汇点。**

---

## 1. 为什么用 Axios 而不是 fetch？

| 能力 | fetch | Axios |
|------|-------|-------|
| JSON 自动解析 | 需手动 `.json()` | 自动 |
| 请求/响应拦截器 | 无 | ✅ 统一 token、错误 |
| 超时控制 | 需 AbortController | `timeout` 配置 |
| 取消请求 | AbortController | CancelToken / AbortController |
| 浏览器/Node 同 API | 部分差异 | 一致 |
| 上传进度 | 有限 | `onUploadProgress` |
| 并发 helper | 无 | `axios.all` / 自行 Promise.all |

**生产项目几乎都用 Axios 或封装 fetch  mimicking Axios。** 本章用 Axios Industry 标准做法。

---

## 2. 联调前检查清单

### 2.1 后端 demo 已启动

```bash
# 在 IDEA 运行 DemoApplication，或：
cd demo
./mvnw spring-boot:run

# 预期控制台：
Started DemoApplication in 2.xxx seconds
```

验证：

```bash
curl http://localhost:8080/api/users
# 预期：{"code":0,"message":"success","data":[...]}
```

### 2.2 后端 CORS 已配置

参考 [04 章 §47](../../后端学习/Java/04-SpringBoot核心开发.md)：

```java
@Configuration
public class CorsConfig implements WebMvcConfigurer {
    @Override
    public void addCorsMappings(CorsRegistry registry) {
        registry.addMapping("/api/**")
            .allowedOriginPatterns("*")
            .allowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
            .allowedHeaders("*")
            .allowCredentials(true);
    }
}
```

### 2.3 统一返回结构 Result

04 章 demo 约定 **`code === 0` 表示成功**：

```json
{
  "code": 0,
  "message": "success",
  "data": { "id": 1, "name": "张三", "age": 20 }
}
```

业务失败：

```json
{
  "code": 1,
  "message": "用户名不能为空",
  "data": null
}
```

前端拦截器必须按此约定解析。

---

## 3. 安装 Axios

```bash
cd shop-vue
npm install axios
```

---

## 4. Vite 开发代理（推荐）

开发环境前端 `http://localhost:5173`，后端 `http://localhost:8080`，浏览器视为**跨域**。两种方案：

| 方案 | 原理 | 适用 |
|------|------|------|
| **Vite proxy** | 开发服务器转发 `/api` | 本地开发首选 |
| **CORS** | 后端允许跨域 Header | 直连后端、生产 |

**`vite.config.js`**：

```js
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // 可选：打印代理日志
        // configure: (proxy) => proxy.on('proxyReq', (_, req) => console.log('代理:', req.url)),
      },
    },
  },
})
```

代理后，前端 `baseURL` 设为空字符串，请求 `/api/users` 会被 Vite 转发到 `http://localhost:8080/api/users`，**浏览器无跨域问题**。

---

## 5. 环境变量

**`.env.development`**：

```env
VITE_API_BASE_URL=
```

**`.env.production`**：

```env
VITE_API_BASE_URL=/api
```

生产环境由 Nginx 把 `/api` 反代到 Spring Boot（10 章）。

---

## 6. 手把手：封装 Axios 实例 `src/api/request.js`

```js
import axios from 'axios'
import { useUserStore } from '@/stores/user'
import router from '@/router'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// ========== 请求拦截器 ==========
request.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// ========== 响应拦截器 ==========
request.interceptors.response.use(
  (response) => {
    const res = response.data

    // 与 Spring Boot Result 约定：code === 0 成功
    if (res.code !== 0) {
      return Promise.reject(new Error(res.message || '业务请求失败'))
    }

    // 直接返回 data 字段，组件里少写一层 .data
    return res.data
  },
  (error) => {
    const status = error.response?.status
    const message = error.response?.data?.message || error.message

    if (status === 401) {
      const userStore = useUserStore()
      userStore.logout()
      router.push({
        name: 'login',
        query: { redirect: router.currentRoute.value.fullPath },
      })
      return Promise.reject(new Error('登录已过期，请重新登录'))
    }

    if (status === 403) {
      return Promise.reject(new Error('没有权限'))
    }

    if (status === 404) {
      return Promise.reject(new Error('接口不存在'))
    }

    if (status >= 500) {
      return Promise.reject(new Error('服务器错误，请稍后重试'))
    }

    if (error.code === 'ECONNABORTED') {
      return Promise.reject(new Error('请求超时'))
    }

    if (!error.response) {
      return Promise.reject(new Error('网络错误，请检查后端是否启动'))
    }

    return Promise.reject(new Error(message))
  }
)

export default request
```

```mermaid
sequenceDiagram
    participant C as 组件
    participant I as 拦截器
    participant B as Spring Boot

    C->>I: getProductList()
    I->>I: 请求拦截：加 Authorization
    I->>B: GET /api/users
    B-->>I: { code:0, data:[...] }
    I->>I: 响应拦截：code!==0 则 reject
    I-->>C: 返回 data（已解包）
```

---

## 7. API 模块拆分

### 7.1 `src/api/auth.js`

```js
import request from './request'

/** 登录 — 对应 04 章 LoginController */
export function login(data) {
  return request.post('/api/login', data)
  // data: { username, password }
  // 返回 data: { token: 'xxx' }
}

/** 注册（若后端实现了 /api/register） */
export function register(data) {
  return request.post('/api/register', data)
}
```

### 7.2 `src/api/product.js`

```js
import request from './request'

/**
 * 商品列表 — 本章用 /api/users 模拟
 * 后端返回 UserVO[]，前端映射为商品结构
 */
export function getProductList(params = {}) {
  return request.get('/api/users', { params })
}

/** 商品详情 — 用 /api/users/{id} 模拟 */
export function getProductById(id) {
  return request.get(`/api/users/${id}`)
}

/** 05 章后端有真实 Product 接口时可替换为： */
// export function getProductList(params) {
//   return request.get('/api/products', { params })
// }
```

### 7.3 `src/api/user.js`

```js
import request from './request'

export function createUser(data) {
  return request.post('/api/users', data)
}

export function deleteUser(id) {
  return request.delete(`/api/users/${id}`)
}
```

### 7.4 `src/api/index.js`（统一导出）

```js
export * from './auth'
export * from './product'
export * from './user'
```

---

## 8. 数据适配：UserVO → 商品展示

后端 `UserVO` 结构：

```json
{ "id": 1, "name": "张三", "age": 20 }
```

前端商品卡片需要 `{ id, name, price }`。在 composable 里做适配：

**`src/composables/useProductAdapter.js`**：

```js
/** 将 UserVO 映射为商品展示结构（联调模拟用） */
export function userToProduct(user) {
  return {
    id: user.id,
    name: user.name,
    price: (user.age || 1) * 10,  // 用 age 模拟价格，仅演示
    category: user.age > 25 ? 'premium' : 'normal',
    desc: `年龄 ${user.age} 岁`,
  }
}

export function usersToProducts(users) {
  return (users || []).map(userToProduct)
}
```

---

## 9. 更新 ProductList.vue（完整四态）

**`src/views/ProductList.vue`**：

```vue
<script setup>
import { ref, onMounted } from 'vue'
import { getProductList } from '@/api/product'
import { usersToProducts } from '@/composables/useProductAdapter'
import ProductCard from '@/components/ProductCard.vue'

const list = ref([])
const loading = ref(false)
const error = ref('')
const isEmpty = ref(false)

async function loadData() {
  loading.value = true
  error.value = ''
  isEmpty.value = false
  try {
    const data = await getProductList()
    // 拦截器已解包，data 即 UserVO[]
    const products = usersToProducts(data)
    list.value = products
    isEmpty.value = products.length === 0
  } catch (e) {
    error.value = e.message || '加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <section>
    <h2>商品列表</h2>

    <!-- 加载态 -->
    <div v-if="loading" class="state-box">
      <p>加载中...</p>
    </div>

    <!-- 错误态 -->
    <div v-else-if="error" class="state-box error">
      <p>{{ error }}</p>
      <button @click="loadData">重试</button>
    </div>

    <!-- 空态 -->
    <div v-else-if="isEmpty" class="state-box">
      <p>暂无商品，请先在后台添加用户数据</p>
      <p class="hint">curl -X POST http://localhost:8080/api/users -H "Content-Type: application/json" -d "{\"name\":\"Vue教程\",\"age\":59}"</p>
    </div>

    <!-- 正常态 -->
    <div v-else class="grid">
      <ProductCard v-for="p in list" :key="p.id" :product="p" />
    </div>
  </section>
</template>

<style scoped>
h2 { margin-bottom: 20px; }
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}
.state-box { padding: 40px; text-align: center; color: #666; }
.state-box.error { color: #e74c3c; }
.hint { font-size: 12px; margin-top: 8px; word-break: break-all; }
button { margin-top: 12px; padding: 8px 16px; cursor: pointer; }
</style>
```

---

## 10. 更新 ProductDetail.vue

**`src/views/ProductDetail.vue`**：

```vue
<script setup>
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getProductById } from '@/api/product'
import { userToProduct } from '@/composables/useProductAdapter'
import { useCartStore } from '@/stores/cart'

const props = defineProps({ id: { type: String, required: true } })
const route = useRoute()
const router = useRouter()
const cartStore = useCartStore()

const product = ref(null)
const loading = ref(false)
const error = ref('')

async function loadDetail(id) {
  loading.value = true
  error.value = ''
  product.value = null
  try {
    const data = await getProductById(id)
    product.value = userToProduct(data)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

watch(() => props.id, (newId) => loadDetail(newId), { immediate: true })
</script>

<template>
  <section>
    <button class="back" @click="router.back()">← 返回</button>

    <div v-if="loading">加载中...</div>
    <div v-else-if="error" class="err">{{ error }}</div>
    <div v-else-if="product">
      <h2>{{ product.name }}</h2>
      <p class="price">¥ {{ product.price }}</p>
      <p>{{ product.desc }}</p>
      <button @click="cartStore.add(product)">加入购物车</button>
    </div>
  </section>
</template>

<style scoped>
.back { margin-bottom: 16px; cursor: pointer; }
.price { color: #e74c3c; font-size: 1.5rem; margin: 12px 0; }
.err { color: #e74c3c; }
button { padding: 8px 16px; background: #42b983; color: #fff; border: none; cursor: pointer; }
</style>
```

---

## 11. 更新 LoginView.vue（真实接口）

**`src/views/LoginView.vue`**：

```vue
<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { login } from '@/api/auth'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const form = reactive({ username: '', password: '' })
const loading = ref(false)
const errMsg = ref('')

async function onSubmit() {
  loading.value = true
  errMsg.value = ''
  try {
    const data = await login(form)
    // data: { token: '...' }
    userStore.setLogin({
      token: data.token,
      username: form.username,
    })
    router.push(route.query.redirect || '/')
  } catch (e) {
    errMsg.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section class="login">
    <h2>用户登录</h2>
    <form @submit.prevent="onSubmit">
      <label>用户名 <input v-model="form.username" required /></label>
      <label>密码 <input v-model="form.password" type="password" required /></label>
      <p v-if="errMsg" class="err">{{ errMsg }}</p>
      <button type="submit" :disabled="loading">
        {{ loading ? '登录中...' : '登录' }}
      </button>
    </form>
  </section>
</template>

<style scoped>
.login { max-width: 360px; margin: 40px auto; }
label { display: block; margin-bottom: 12px; }
input { width: 100%; padding: 8px; margin-top: 4px; box-sizing: border-box; }
.err { color: #e74c3c; }
button { width: 100%; padding: 10px; background: #42b983; color: #fff; border: none; cursor: pointer; }
</style>
```

---

## 12. 后端 Login 接口（联调必备）

> 以下示例为 **Spring Boot（Java）**；若你走 Python 路线，见 [Python 04 §8 JWT 入门](../../后端学习/Python/04-FastAPI核心开发.md) 与 [Python 10 项目实战](../../后端学习/Python/10-后端项目实战与面试准备.md)，接口契约（`Result<T>`、`/api/login`）保持一致即可。

若你尚未实现 04 章 JWT 挑战，可先加**简化版登录**（无 JWT，返回假 token）：

**`dto/LoginDTO.java`**：

```java
package com.example.demo.dto;

import jakarta.validation.constraints.NotBlank;

public class LoginDTO {
    @NotBlank(message = "用户名不能为空")
    private String username;
    @NotBlank(message = "密码不能为空")
    private String password;

    public String getUsername() { return username; }
    public void setUsername(String username) { this.username = username; }
    public String getPassword() { return password; }
    public void setPassword(String password) { this.password = password; }
}
```

**`controller/LoginController.java`**：

```java
package com.example.demo.controller;

import com.example.demo.common.Result;
import com.example.demo.dto.LoginDTO;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.*;

import java.util.Map;
import java.util.UUID;

@RestController
public class LoginController {

    @PostMapping("/api/login")
    public Result<Map<String, String>> login(@Valid @RequestBody LoginDTO dto) {
        // 演示：任意非空用户名密码均通过；生产应查库比对
        if ("admin".equals(dto.getUsername()) && "123456".equals(dto.getPassword())) {
            String token = UUID.randomUUID().toString();
            return Result.success(Map.of("token", token));
        }
        return Result.fail("用户名或密码错误");
    }
}
```

测试：

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"123456\"}"
# 预期：{"code":0,"message":"success","data":{"token":"..."}}
```

完整 JWT 版见 [04 章 §49 挑战参考答案](../../后端学习/Java/04-SpringBoot核心开发.md)。

---

## 13. 完整联调流程

```mermaid
sequenceDiagram
    participant Dev as 开发者
    participant Vue as shop-vue :5173
    participant Boot as Spring Boot :8080
    participant DB as 内存List/MySQL

    Dev->>Boot: 启动 DemoApplication
    Dev->>Vue: npm run dev
    Dev->>Boot: POST /api/users 造数据
    Vue->>Boot: GET /api/users (经 Vite 代理)
    Boot->>DB: findAll()
    DB-->>Boot: List UserVO
    Boot-->>Vue: Result success
    Vue->>Vue: usersToProducts → 渲染列表

    Dev->>Vue: 登录 admin/123456
    Vue->>Boot: POST /api/login
    Boot-->>Vue: { token }
    Vue->>Vue: userStore.setLogin
    Vue->>Boot: GET /api/users (Header: Bearer token)
```

### 13.1 逐步验证

```bash
# 终端 1：后端
cd demo && ./mvnw spring-boot:run

# 终端 2：前端
cd shop-vue && npm run dev

# 终端 3：造数据
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Vue3教程\",\"age\":59}"
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"机械键盘\",\"age\":39}"
```

浏览器：

1. 打开 `http://localhost:5173/products` → 应看到商品卡片
2. F12 → Network → 看到 `/api/users` 状态 200
3. 登录 `admin` / `123456` → 应跳首页，localStorage 有 token

---

## 14. 页面四态设计规范

每个依赖接口的页面必须处理：

| 状态 | 变量 | UI 建议 |
|------|------|---------|
| loading | `loading === true` | 骨架屏、`v-loading`（09 章） |
| error | `error !== ''` | 错误文案 + 重试按钮 |
| empty | `list.length === 0` | 插图 + 引导文案 |
| success | 以上皆否 | 正常业务 UI |

```mermaid
stateDiagram-v2
    [*] --> Loading: 发起请求
    Loading --> Success: code===0 有数据
    Loading --> Empty: code===0 无数据
    Loading --> Error: 网络/业务失败
    Error --> Loading: 点击重试
```

---

## 15. composable 封装请求逻辑

**`src/composables/useRequest.js`**：

```js
import { ref } from 'vue'

export function useRequest(asyncFn) {
  const data = ref(null)
  const loading = ref(false)
  const error = ref('')

  async function execute(...args) {
    loading.value = true
    error.value = ''
    try {
      data.value = await asyncFn(...args)
      return data.value
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  return { data, loading, error, execute }
}
```

使用：

```js
const { data, loading, error, execute } = useRequest(getProductList)
onMounted(() => execute())
```

---

## 16. 并发请求与依赖请求

```js
import { getProductById } from '@/api/product'

// 并发
const [detail, recommend] = await Promise.all([
  getProductById(id),
  getProductList({ category: 'hot' }),
])

// 依赖：先登录再拉个人信息
await login(form)
await fetchProfile()
```

---

## 17. 取消重复请求（防抖）

快速切换 Tab 时，旧请求后返回覆盖新数据：

```js
import axios from 'axios'

let cancelFn = null

async function loadDetail(id) {
  if (cancelFn) cancelFn('取消上一次请求')
  const source = axios.CancelToken.source()
  cancelFn = source.cancel

  const data = await request.get(`/api/users/${id}`, {
    cancelToken: source.token,
  })
  // ...
}
```

或使用 AbortController（Axios 1.x 推荐）。

---

## 18. 文件上传（了解）

```js
export function uploadAvatar(file) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/api/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
```

---

## 19. 生产级案例：统一错误 Toast

09 章用 Element Plus `ElMessage`，08 章可先用 `alert` 或在拦截器：

```js
// 响应拦截 error 分支
console.error('[API Error]', message)
// 09 章：ElMessage.error(message)
```

---

## 20. 生产级案例：Token 刷新（思路）

```mermaid
sequenceDiagram
    participant A as Axios
    participant B as Backend

    A->>B: 请求 + accessToken
    B-->>A: 401 token 过期
    A->>B: POST /api/refresh + refreshToken
    B-->>A: 新 accessToken
    A->>B: 重试原请求
```

实现要点：401 时判断能否 refresh，避免 logout 死循环；用队列暂存并发请求。

---

## 21. 生产级案例：接口 Mock（后端未完成时）

```js
// vite.config.js
export default defineConfig({
  plugins: [
    {
      name: 'mock-api',
      configureServer(server) {
        server.middlewares.use('/api/users', (req, res) => {
          res.setHeader('Content-Type', 'application/json')
          res.end(JSON.stringify({
            code: 0,
            message: 'success',
            data: [{ id: 1, name: 'Mock商品', age: 99 }],
          }))
        })
      },
    },
  ],
})
```

或使用 MSW（Mock Service Worker）。

---

## 22. RESTful 与后端约定对照

| 前端方法 | 后端注解 | 示例 |
|----------|----------|------|
| GET | `@GetMapping` | 列表、详情 |
| POST | `@PostMapping` | 创建、登录 |
| PUT | `@PutMapping` | 全量更新 |
| PATCH | `@PatchMapping` | 部分更新 |
| DELETE | `@DeleteMapping` | 删除 |

详见 [04 章 §51 RESTful 规范](../../后端学习/Java/04-SpringBoot核心开发.md)。

---

## 23. 学完标准

- [ ] 会封装 Axios 实例 + 请求/响应拦截器
- [ ] 能对接 Spring Boot `Result` 结构（code === 0）
- [ ] 会配置 Vite 代理，理解 CORS 原理
- [ ] 登录流程：POST /api/login → token → Pinia → 后续请求带 Bearer
- [ ] 列表/详情页具备 loading/error/empty/success 四态
- [ ] 401 自动 logout 并跳登录页

---

## 24. 分级练习

### 24.1 基础：GET 列表并 v-for 渲染

**参考答案**：见 §9 ProductList.vue。

---

### 24.2 进阶：POST 登录，存 token

**参考答案**：见 §11 LoginView + §6 请求拦截器。

验证：登录后 Network 里后续请求 Header 含 `Authorization: Bearer xxx`。

---

### 24.3 挑战：401 自动跳登录页

**参考答案**：见 §6 响应拦截器 `status === 401` 分支。

测试：手动 `userStore.token = 'invalid'`，刷新访问需鉴权页，应跳登录。

---

### 24.4 挑战+：新增用户 API 对接

```vue
<script setup>
import { createUser } from '@/api/user'
async function handleAdd() {
  await createUser({ name: '新商品', age: 88 })
  await loadData()
}
</script>
```

---

## 25. 常见报错与排查

| 报错信息 | 可能原因 | 排查步骤 | 解决方案 |
|---------|---------|---------|---------|
| `Network Error` | 后端未启动或端口错 | `curl localhost:8080/api/users` | 启动 DemoApplication |
| CORS policy blocked | 没用代理且后端无 CORS | 看 Console 红字 | 配 Vite proxy 或 CorsConfig |
| `404 Not Found` on /api/xxx | 路径或方法不对 | 对比 Controller 注解 | 修正 URL；GET 勿用 POST |
| `401 Unauthorized` | token 无效/过期/未带 | 看 Request Headers | 重新登录；检查 Bearer 格式 |
| 数据是 undefined | 拦截器解包层级错 | 看 Network Response 原始 JSON | 对齐 `res.data.data` 或改拦截器 |
| `code !== 0` 业务失败 | 参数校验失败 | 看 message 字段 | 对齐 DTO 字段名 |
| `timeout of 15000ms exceeded` | 后端慢或死锁 | 看后端日志 | 调大 timeout；修后端 |
| 代理不生效 | vite.config 改完未重启 | 重启 dev server | `npm run dev` |
| `Required request body is missing` | POST 没带 JSON body | 看 Request Payload | 设 Content-Type；传对象 |
| 登录成功但立刻 401 | token 格式与后端不一致 | 看后端 Interceptor | 对齐 JWT 解析逻辑 |
| 刷新后请求不带 token | Pinia 未持久化 | Application localStorage | userStore setLogin 写 storage |
| `Whitelabel Error Page` | 访问了 8080 根路径 | 正常，/ 无映射 | 只访问 /api/** |
| OPTIONS 预检失败 | CORS 未允许 OPTIONS | Network 里 OPTIONS 红 | CorsConfig 加 OPTIONS |

---

## 26. 常见问题 FAQ

### Q1：开发用 proxy，生产怎么办？

生产 Nginx：`location /api { proxy_pass http://backend:8080; }`，前端 `VITE_API_BASE_URL=/api`，同域无跨域。

### Q2：为什么拦截器 return res.data 而不是整个 res？

减少组件里 `.data.data` 重复书写；业务失败已在拦截器 reject。

### Q3：GET 请求 params 怎么传？

```js
request.get('/api/users', { params: { pageNum: 1, pageSize: 10 } })
// 实际 URL: /api/users?pageNum=1&pageSize=10
```

### Q4：前后端字段名不一致怎么办？

 adapter 层映射（见 §8），或后端 `@JsonProperty`，或统一命名规范。

### Q5：能否在 SSR/Nuxt 里用同样封装？

可以，但 `localStorage` 和 `useUserStore` 需防 Node 端 undefined，token 改从 Cookie 读。

### Q6：axios 和 fetch 能混用吗？

能但不推荐；统一实例便于拦截器治理。

---

## 27. 本章小结

```mermaid
mindmap
  root((前后端联调))
    Axios
      实例封装
      拦截器
    后端
      Spring Boot
      Result 约定
      CORS
    开发
      Vite proxy
      环境变量
    业务
      登录 token
      四态 UI
      数据适配
```

接口通了，但手写 HTML/CSS 做表格、表单、分页太慢。下一章（09 Element Plus）接入主流 UI 组件库，快速搭出专业界面。

---

## 下一章预告

08 章联调成功后，`shop-vue` 已具备真实数据流。下一章用 **Element Plus** 把登录表单、商品表格、分页、消息提示换成企业级 UI，并介绍按需引入与布局工程化。

---

*下一章：09 Element Plus 与 UI 工程化*
