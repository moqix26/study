# Vue 3 与 TypeScript

## 本章衔接

[06-模块声明文件与三方库](./06-模块声明文件与三方库.md) 里，你已经在 `src/types/` 定义了 `Product`、`ApiResult`，并配置好了 `@/` 别名与 `vite-env.d.ts`。但这些类型尚未进入 **Vue 单文件组件（SFC）** 和 **Pinia Store**。

若你按 [Vue 学习路线](../Vue/00-学习路线图与说明.md) 推进，04～07 章的 `shop-vue` 组件与 Store 目前多半是 **JavaScript** 写法：

- `defineProps({ product: { type: Object } })` — 运行时校验，无编译期类型
- `const cartItems = ref([])` — 推断为 `never[]` 或 `any[]`
- Pinia Store 无泛型，重构时 IDE 无法追踪

本章目标：**把 shop-vue 核心文件改为 TypeScript**，获得与后端 [Java 04](../../后端学习/Java/04-SpringBoot核心开发.md) 联调时「字段拼错即报错」的体验。

```mermaid
flowchart LR
    TS06["TS 06：types + @/"] --> TS07["TS 07：Vue SFC + Pinia"]
    Vue05["Vue 05：script setup"] --> TS07
    Vue07["Vue 07：Pinia"] --> TS07
    TS07 --> TS10["TS 10：全项目迁移"]
```

**前置检查**：

- 完成 TS 01～06
- 完成 [Vue 05-组合式 API 与 script setup](../Vue/05-组合式API与script-setup.md)
- 建议已读 [Vue 07-Pinia 状态管理](../Vue/07-Pinia状态管理.md)（本章会加类型层）
- `shop-vue` 可 `npm run dev`；推荐安装 `vue-tsc` 做类型检查

**读法建议**：Vue 主线同学 **精读本章**；React 主线同学 **浏览一遍** 即可，重点看「类型思维」如何映射到 08 章。

---

## 1. 为什么在 Vue 3 里用 TypeScript

### 1.1 JS 写法的隐性成本

[Vue 04](../Vue/04-组件基础与组件通信.md) 的 `ProductCard` 用运行时 props：

```js
defineProps({
  product: { type: Object, required: true },
})
```

问题：

| 问题 | 后果 |
|------|------|
| `product` 是 `Object` | `product.pric` 拼错无提示 |
| emit 事件名手写字符串 | 父组件 `@add-crat` 静默失效 |
| Store 的 `items` 无类型 | `item.qty` 写成 `item.quantity` 运行才错 |
| 重构改字段名 | 靠全局搜索，易漏 |

### 1.2 TS 写法的收益

| 场景 | TS 表现 |
|------|---------|
| props 传错类型 | 模板与脚本均报红 |
| emit 载荷类型 | `defineEmits` 泛型约束 payload |
| API 响应 | `ref<Product[]>` 自动推导 |
| Pinia | 完整 state/getter/action 推断 |
| 与后端对齐 | `Product` 接口一处定义，全项目复用 |

```mermaid
flowchart TB
    API[Spring Boot JSON] --> Types["@/types Product"]
    Types --> SFC[ProductCard.vue]
    Types --> Store[cartStore]
    Types --> API2[productApi.ts]
```

---

## 2. 启用 TypeScript：项目与 SFC 配置

### 2.1 创建或迁移到 vue-ts 模板

```bash
npm create vite@latest shop-vue -- --template vue-ts
cd shop-vue
npm install
npm install pinia vue-router
npm install -D vue-tsc
```

**`package.json` scripts 补充**：

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc -b && vite build",
    "type-check": "vue-tsc --noEmit"
  }
}
```

**`npm run type-check` 预期**：无类型错误时静默退出，code 0。

### 2.2 单文件组件启用 TS

```vue
<script setup lang="ts">
// 此处代码按 TypeScript 解析
</script>
```

| 属性 | 含义 |
|------|------|
| `setup` | 组合式 API 语法糖 |
| `lang="ts"` | script 块使用 TypeScript（默认是 JS） |

**注意**：`<template>` 和 `<style>` 不写字面量 `lang="ts"`；模板类型检查由 `vue-tsc` 基于 script 中的 props/emits 推断。

### 2.3 `tsconfig` 与 Vue 插件

Vite Vue-TS 模板通常包含：

```json
// tsconfig.app.json
{
  "compilerOptions": {
    "jsx": "preserve",
    "jsxImportSource": "vue",
    "moduleResolution": "bundler",
    "strict": true,
    "paths": { "@/*": ["./src/*"] }
  },
  "include": ["src/**/*.ts", "src/**/*.tsx", "src/**/*.vue"]
}
```

`vue-tsc` 会读取 `.vue` 文件并做类型检查。

---

## 3. `defineProps` 与泛型写法

### 3.1 类型声明式（推荐）

Vue 3.3+ 支持**基于泛型的类型声明**，编译期检查，无运行时 props 对象：

```vue
<script setup lang="ts">
import type { Product } from '@/types'

const props = defineProps<{
  product: Product
  showPrice?: boolean
}>()

// props.product.price 有完整提示
</script>
```

带默认值需配合 `withDefaults`：

```vue
<script setup lang="ts">
import type { Product } from '@/types'

const props = withDefaults(
  defineProps<{
    product: Product
    showPrice?: boolean
  }>(),
  {
    showPrice: true,
  }
)
</script>
```

### 3.2 运行时声明 + `PropType`（过渡方案）

从 JS 迁移时常见：

```vue
<script setup lang="ts">
import type { PropType } from 'vue'
import type { Product } from '@/types'

defineProps({
  product: {
    type: Object as PropType<Product>,
    required: true,
  },
  showPrice: {
    type: Boolean,
    default: true,
  },
})
</script>
```

| 方式 | 优点 | 缺点 |
|------|------|------|
| `defineProps<{...}>()` | 简洁、类型即文档 | 默认值需 `withDefaults` |
| `PropType` + 运行时 | 保留运行时校验 | 样板代码多 |

**新项目优先泛型写法**。

### 3.3 不要用解构丢失响应式

```vue
<script setup lang="ts">
import type { Product } from '@/types'

// ❌ 直接解构 props 会失去响应式（除非用 toRefs）
const { product } = defineProps<{ product: Product }>()

// ✅ 模板里用 props.product，或：
const props = defineProps<{ product: Product }>()
</script>

<template>
  <p>{{ props.product.name }}</p>
</template>
```

Vue 3.5+ 对解构有改进，初学阶段 **模板中直接用 `props.xxx` 最稳妥**。

---

## 4. `defineEmits` 类型

### 4.1 元组语法（推荐）

```vue
<script setup lang="ts">
import type { Product } from '@/types'

const emit = defineEmits<{
  'add-cart': [product: Product]
  'view-detail': [id: number]
}>()

function handleAdd(product: Product) {
  emit('add-cart', product)
}
</script>
```

父组件：

```vue
<ProductCard
  :product="p"
  @add-cart="onAddCart"
/>
```

```typescript
function onAddCart(product: Product) {
  // product 类型自动推断
}
```

### 4.2 事件名与载荷错误示例

```typescript
emit('add-cart', { id: '1', name: 'x', price: 1 })
// ❌ id 应为 number

emit('add-crat', product)
// ❌ 事件名拼写错误，编译报错
```

### 4.3 与 Pinia 协作时

[ProductCard](../Vue/07-Pinia状态管理.md) 在 07 章改为直接调 `cartStore.add(product)`，可 **不再 emit**。类型重点落在 **Store 的 `add` 参数** 上（§8）。

---

## 5. `ref`、`reactive` 与 `computed` 类型

### 5.1 `ref` 显式泛型

```typescript
import { ref } from 'vue'
import type { Product } from '@/types'

// 空数组必须标注，否则是 never[]
const list = ref<Product[]>([])

// 简单值可自动推断
const keyword = ref('')        // Ref<string>
const loading = ref(false)     // Ref<boolean>

// 可空对象
const current = ref<Product | null>(null)
```

访问时记得 `.value`：

```typescript
list.value.push({ id: 1, name: 'T恤', price: 99 })
if (current.value) {
  console.log(current.value.name)
}
```

### 5.2 `reactive` 与类型

```typescript
import { reactive } from 'vue'
import type { Product } from '@/types'

interface FilterState {
  keyword: string
  category: string
  minPrice: number
}

const filter = reactive<FilterState>({
  keyword: '',
  category: 'all',
  minPrice: 0,
})

// 整对象替换会丢响应式 — 应改字段
filter.keyword = '手机'
```

**经验**：表单/筛选状态用 `reactive<Interface>`；列表、可空实体用 `ref<T>`。

### 5.3 `computed` 推断

```typescript
import { computed } from 'vue'

const filteredList = computed(() => {
  return list.value.filter((p) =>
    p.name.includes(keyword.value)
  )
})
// ComputedRef<Product[]>

const total = computed(() =>
  filteredList.value.reduce((sum, p) => sum + p.price, 0)
)
// ComputedRef<number>
```

显式标注（复杂逻辑时）：

```typescript
const stats = computed<{ count: number; total: number }>(() => ({
  count: filteredList.value.length,
  total: filteredList.value.reduce((s, p) => s + p.price, 0),
}))
```

### 5.4 模板中的类型收窄

```vue
<template>
  <p v-if="current">{{ current.name }}</p>
</template>
```

`v-if="current"` 后，模板内 `current` 收窄为 `Product`（非 null）。

---

## 6. 组合式函数（Composables）类型

与 [Vue 05](../Vue/05-组合式API与script-setup.md) 的 `useProducts` 对应，加上类型：

**`src/composables/useProducts.ts`**：

```typescript
import { ref, computed } from 'vue'
import type { Product } from '@/types'
import { fetchProducts } from '@/api/productApi'

export function useProducts() {
  const products = ref<Product[]>([])
  const keyword = ref('')
  const loading = ref(false)
  const error = ref<string | null>(null)

  const filteredProducts = computed(() =>
    products.value.filter((p) =>
      p.name.toLowerCase().includes(keyword.value.toLowerCase())
    )
  )

  async function load() {
    loading.value = true
    error.value = null
    try {
      const res = await fetchProducts()
      if (res.code === 0) {
        products.value = res.data
      } else {
        error.value = res.message
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : '未知错误'
    } finally {
      loading.value = false
    }
  }

  return {
    products,
    keyword,
    loading,
    error,
    filteredProducts,
    load,
  }
}
```

组件中导入即有完整返回类型推断。

---

## 7. Pinia Store 类型化

对照 [Vue 07-Pinia](../Vue/07-Pinia状态管理.md)，将 `.js` Store 改为 `.ts`。

### 7.1 Setup Store（推荐）

**`src/stores/cart.ts`**：

```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Product, CartItem } from '@/types'

export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>([])

  const totalCount = computed(() =>
    items.value.reduce((sum, item) => sum + item.qty, 0)
  )

  const totalPrice = computed(() =>
    items.value.reduce((sum, item) => sum + item.price * item.qty, 0)
  )

  const isEmpty = computed(() => items.value.length === 0)

  function add(product: Product, qty = 1) {
    const exist = items.value.find((i) => i.id === product.id)
    if (exist) {
      exist.qty += qty
    } else {
      items.value.push({ ...product, qty })
    }
  }

  function remove(id: number) {
    items.value = items.value.filter((i) => i.id !== id)
  }

  function updateQty(id: number, qty: number) {
    const item = items.value.find((i) => i.id === id)
    if (!item) return
    if (qty <= 0) {
      remove(id)
    } else {
      item.qty = qty
    }
  }

  function clear() {
    items.value = []
  }

  return {
    items,
    totalCount,
    totalPrice,
    isEmpty,
    add,
    remove,
    updateQty,
    clear,
  }
})
```

### 7.2 `storeToRefs` 保持响应式

```vue
<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useCartStore } from '@/stores/cart'

const cartStore = useCartStore()
const { items, totalCount, isEmpty } = storeToRefs(cartStore)
// items 是 Ref<CartItem[]>，模板可直接用
</script>
```

```typescript
// ❌ 直接解构丢响应式
const { items } = useCartStore()
```

### 7.3 userStore 类型示例

**`src/stores/user.ts`**：

```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

const TOKEN_KEY = 'shop_token'
const USERNAME_KEY = 'shop_username'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem(TOKEN_KEY) || '')
  const username = ref(localStorage.getItem(USERNAME_KEY) || '')

  const isLoggedIn = computed(() => !!token.value)
  const displayName = computed(() => username.value || '游客')

  function setLogin(newToken: string, name: string) {
    token.value = newToken
    username.value = name
    localStorage.setItem(TOKEN_KEY, newToken)
    localStorage.setItem(USERNAME_KEY, name)
  }

  function logout() {
    token.value = ''
    username.value = ''
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USERNAME_KEY)
  }

  return { token, username, isLoggedIn, displayName, setLogin, logout }
})
```

### 7.4 Options Store 类型（了解）

```typescript
import { defineStore } from 'pinia'
import type { CartItem } from '@/types'

interface CartState {
  items: CartItem[]
}

export const useCartStore = defineStore('cart', {
  state: (): CartState => ({
    items: [],
  }),
  getters: {
    totalCount(state): number {
      return state.items.reduce((s, i) => s + i.qty, 0)
    },
  },
  actions: {
    add(item: CartItem) {
      this.items.push(item)
    },
  },
})
```

新项目仍推荐 **Setup Store**，与 `<script setup lang="ts">` 一致。

---

## 8. 手把手：ProductCard 完整 TS 版

对照 [Vue 07 §11.1](../Vue/07-Pinia状态管理.md)，将 `ProductCard.vue` 类型化。

### 8.1 类型定义（若 06 章未完成）

**`src/types/product.ts`**：

```typescript
export interface Product {
  id: number
  name: string
  price: number
  stock?: number
  img?: string
  category?: string
}

export interface CartItem extends Product {
  qty: number
}
```

### 8.2 ProductCard.vue

**`src/components/ProductCard.vue`**：

```vue
<script setup lang="ts">
import { computed } from 'vue'
import { useCartStore } from '@/stores/cart'
import type { Product } from '@/types'

const props = withDefaults(
  defineProps<{
    product: Product
    showPrice?: boolean
  }>(),
  { showPrice: true }
)

const cartStore = useCartStore()

const isSoldOut = computed(() => props.product.stock === 0)

const formattedPrice = computed(() =>
  `¥ ${props.product.price.toFixed(2)}`
)

function handleAdd() {
  if (isSoldOut.value) return
  cartStore.add(props.product)
}
</script>

<template>
  <article class="card" :class="{ 'card--sold': isSoldOut }">
    <img
      v-if="product.img"
      :src="product.img"
      :alt="product.name"
      class="thumb"
    />
    <h3>{{ product.name }}</h3>
    <p v-if="showPrice" class="price">{{ formattedPrice }}</p>
    <p v-if="isSoldOut" class="badge">售罄</p>
    <div class="actions">
      <router-link :to="`/products/${product.id}`">详情</router-link>
      <button type="button" :disabled="isSoldOut" @click="handleAdd">
        加入购物车
      </button>
    </div>
  </article>
</template>

<style scoped>
.card {
  background: #fff;
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 16px;
}
.card--sold { opacity: 0.6; }
.thumb { width: 100%; height: 120px; object-fit: cover; border-radius: 4px; }
.price { color: #e74c3c; margin: 8px 0; }
.badge { color: #999; font-size: 12px; }
.actions { display: flex; gap: 12px; align-items: center; }
button {
  padding: 6px 12px;
  background: #42b983;
  color: #fff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
button:disabled { background: #ccc; cursor: not-allowed; }
</style>
```

### 8.3 父组件 ProductList / HomeView

```vue
<script setup lang="ts">
import { onMounted } from 'vue'
import ProductCard from '@/components/ProductCard.vue'
import { useProducts } from '@/composables/useProducts'

const { filteredProducts, keyword, loading, load } = useProducts()

onMounted(() => {
  load()
})
</script>

<template>
  <section>
    <input v-model="keyword" placeholder="搜索商品" />
    <p v-if="loading">加载中...</p>
    <div v-else class="grid">
      <ProductCard
        v-for="p in filteredProducts"
        :key="p.id"
        :product="p"
      />
    </div>
  </section>
</template>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}
</style>
```

### 8.4 验证清单

| 步骤 | 预期 |
|------|------|
| `:product="{ id: 1, name: 'x' }"` 缺 `price` | `vue-tsc` 报错 |
| `cartStore.add` 传入非 Product | 报红 |
| 点击加购 | Pinia DevTools 可见 items |
| `npm run type-check` | exit code 0 |

---

## 9. `ComponentPublicInstance` 与模板 ref

### 9.1 何时需要

[Vue 04 §12](../Vue/04-组件基础与组件通信.md) 用父组件 `ref` 调子组件 `expose` 的方法。TS 需标注 ref 类型。

子组件 **SearchBar.vue**：

```vue
<script setup lang="ts">
import { ref } from 'vue'

const inputRef = ref<HTMLInputElement | null>(null)

function focus() {
  inputRef.value?.focus()
}

defineExpose({ focus })
</script>

<template>
  <input ref="inputRef" />
</template>
```

父组件：

```vue
<script setup lang="ts">
import { ref } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import SearchBar from '@/components/SearchBar.vue'

// 方式 1：自定义 expose 接口（推荐）
interface SearchBarExpose {
  focus: () => void
}

const searchBarRef = ref<SearchBarExpose | null>(null)

// 方式 2：ComponentPublicInstance（较宽，少约束）
// const searchBarRef = ref<ComponentPublicInstance | null>(null)

function onPageShow() {
  searchBarRef.value?.focus()
}
</script>

<template>
  <SearchBar ref="searchBarRef" />
</template>
```

### 9.2 `ComponentPublicInstance` 说明

`ComponentPublicInstance` 是 Vue 3 对「组件实例」的公共类型，包含 `$el`、`$emit` 等。日常更推荐 **为 `defineExpose` 声明专用接口**，类型更精确。

### 9.3 DOM ref 类型

```typescript
const el = ref<HTMLDivElement | null>(null)
const btn = ref<HTMLButtonElement | null>(null)
```

模板：`<div ref="el">` — `vue-tsc` 会校验 ref 与元素匹配。

---

## 10. 路由与 Router 类型（简要）

```typescript
// src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  { path: '/', component: () => import('@/views/HomeView.vue') },
  { path: '/products/:id', component: () => import('@/views/ProductDetailView.vue') },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
```

组件内：

```typescript
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const productId = Number(route.params.id) // params 默认 string | string[]

const router = useRouter()
router.push({ name: 'cart' })
```

详情见 [Vue 06-Router](../Vue/06-Vue-Router路由管理.md)；TS 层重点是 **`RouteRecordRaw`** 与 **`route.params` 收窄**。

---

## 11. 与 Axios 联调的类型衔接

[Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 将对接后端。提前准备：

**`src/api/request.ts`**：

```typescript
import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig } from 'axios'
import type { ApiResult } from '@/types'

const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api',
  timeout: 10000,
})

request.interceptors.response.use((response) => {
  return response.data as ApiResult<unknown>
})

export default request

export async function get<T>(url: string): Promise<ApiResult<T>> {
  return request.get(url)
}
```

确保 JSON 字段与 [Java DTO](../../后端学习/Java/04-SpringBoot核心开发.md) 一致，否则 TS 再严也无法阻止运行时字段缺失——**类型是前后端契约的文档**。

---

## 12. 常见报错与排查

| 报错信息 | 可能原因 | 排查步骤 | 解决方案 |
|---------|---------|---------|---------|
| 类型 `{ id: string }` 不能赋给 `Product` | API 返回 id 为字符串 | 看 Network 响应 | 统一类型或 `Number(id)` 转换 |
| `Property 'value' does not exist` | 忘了 ref 的 .value | 看变量是否 Ref | 脚本里用 `.value` |
| `getActivePinia was called with no active Pinia` | main.ts 未 use pinia | 查 main.ts 顺序 | `app.use(pinia)` 在 mount 前 |
| 解构 store 后 UI 不更新 | 未用 storeToRefs | 对比 §7.2 | `storeToRefs(cartStore)` |
| `defineProps` 默认值不生效 | 泛型 props 未用 withDefaults | 查 script | 包一层 `withDefaults` |
| 模板中 `product` 可能为 undefined | 可选 prop 未收窄 | 加 v-if | `v-if="product"` 或必填 |
| `vue-tsc` 找不到模块 `@/types` | paths 未配 | tsconfig | 见 TS 06 章 |
| `emit` 参数类型不匹配 | 元组载荷写错 | 看 defineEmits | 对齐 `[product: Product]` |
| `ComponentPublicInstance` 上无 `focus` | expose 未声明类型 | 子组件 defineExpose | 父组件用自定义 Expose 接口 |
| `never[]` 不能 push | ref 未标注泛型 | `ref([])` | `ref<Product[]>([])` |
| 导入 `.vue` 无默认导出 | vite-env.d.ts 缺失 | 查声明文件 | 补 `declare module '*.vue'` |
| Pinia action 内 `this` 类型错误 | Options Store 混用 | 改 Setup Store | `defineStore('id', () => {...})` |

---

## 13. Vue 3 + TS 最佳实践清单

| 实践 | 说明 |
|------|------|
| 统一 `@/types` | Product、User、ApiResult 一处维护 |
| SFC 一律 `lang="ts"` | 新组件不写 JS |
| props 用泛型 + withDefaults | 少写 PropType 样板 |
| 列表/实体用 `ref<T>` | 避免 never[] |
| Store 用 Setup Store + `.ts` | 与 composable 风格统一 |
| 提交前跑 `vue-tsc --noEmit` | CI 可接同一命令 |
| 模板 ref 声明 Expose 接口 | 优于宽泛的 ComponentPublicInstance |

```mermaid
mindmap
  root((Vue3 + TS))
    SFC
      script setup lang ts
      defineProps 泛型
      defineEmits 元组
    响应式
      ref 泛型
      reactive 接口
      computed 推断
    状态
      Pinia Setup Store
      storeToRefs
    组件
      ProductCard 示例
      defineExpose
    工程
      vue-tsc
      vite-env.d.ts
```

---

## 14. 常见问题 FAQ

### Q1：`<script setup>` 里能用 `interface` 吗？

可以。`interface Props {}` 与 `type Props = {}` 均可用于 `defineProps<Props>()`。

### Q2：模板里需要写类型吗？

不需要。模板类型由 `vue-tsc` 根据 script 推断；错误会标在模板或 script 引用处。

### Q3：Pinia 和 composable 如何选择？

跨路由、全局态 → Pinia（见 [Vue 07](../Vue/07-Pinia状态管理.md)）；单页或局部复用逻辑 → composable。

### Q4：能不能部分组件 TS、部分 JS？

可以（`allowJs: true`），但 shop 迁移目标应是 **全 TS**（[10 章](./10-项目实战JS到TS迁移.md)）。

### Q5：和 React 08 章的关系？

类型定义（`Product`、`ApiResult`）**两边可共用概念**；组件写法不同，但 `@/types` 思路一致。

---

## 15. 本章小结

本章把 [06 章](./06-模块声明文件与三方库.md) 的类型基础设施接入了 Vue 3 生态：`defineProps` / `defineEmits` 泛型、`ref`/`computed` 标注、Pinia Setup Store，以及 **shop-vue ProductCard** 的完整 TS 实现。完成后，你的商城组件具备 **编译期 props 校验** 与 **Store 重构安全网**。

---

## 练习建议

### 基础题

1. 将 `CartBadge.vue` 改为 `lang="ts"`，`defineProps<{ count: number }>()`。
2. 为 `LoginForm.vue` 定义 `defineEmits<{ success: [username: string] }>()`。
3. 在 `useProducts.ts` 中为 `keyword` 和 `filteredProducts` 确认类型推断无误。

### 进阶题

4. 实现 `SearchBar.vue`：`v-model:keyword` 用 `defineModel<string>()`（Vue 3.4+）或 props + emit，并写全类型。
5. `ProductDetailView.vue` 从 `route.params.id` 取 id，请求详情，使用 `ref<Product | null>(null)`。

### 挑战题

6. 为 `cartStore.add` 增加库存校验：`product.stock` 不足时 `throw` 或返回 `false`，调用方类型化处理；补全 `vue-tsc` 通过。

---

## 练习参考答案

### 基础题 1：CartBadge.vue

```vue
<script setup lang="ts">
defineProps<{
  count: number
}>()
</script>

<template>
  <span class="badge">🛒 {{ count }}</span>
</template>
```

### 基础题 2：LoginForm.vue emit

```vue
<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{
  success: [username: string]
}>()

const username = ref('')
const password = ref('')

function handleSubmit() {
  if (username.value && password.value) {
    emit('success', username.value)
  }
}
</script>
```

### 进阶题 4：defineModel 示例

```vue
<script setup lang="ts">
const keyword = defineModel<string>('keyword', { default: '' })
</script>

<template>
  <input v-model="keyword" placeholder="搜索" />
</template>
```

### 挑战题 6：库存校验

```typescript
function add(product: Product, qty = 1): boolean {
  if (product.stock !== undefined && product.stock <= 0) return false
  const exist = items.value.find((i) => i.id === product.id)
  const nextQty = (exist?.qty ?? 0) + qty
  if (product.stock !== undefined && nextQty > product.stock) return false
  // ... 原有逻辑
  return true
}
```

```vue
function handleAdd() {
  const ok = cartStore.add(props.product)
  if (!ok) alert('库存不足')
}
```

---

## 学完标准

| # | 能力 | 自检方式 |
|---|------|----------|
| 1 | 新建 `.vue` 默认 `script setup lang="ts"` | 无 JS 新组件 |
| 2 | `defineProps` 泛型 + `withDefaults` | ProductCard 通过 type-check |
| 3 | `defineEmits` 元组约束事件载荷 | 故意错事件名会报错 |
| 4 | `ref<Product[]>` 等显式泛型 | 无 never[] 问题 |
| 5 | Pinia Setup Store 全 `.ts` | cart/user store 有完整推断 |
| 6 | `storeToRefs` 正确使用 | 购物车页数据响应式 |
| 7 | 模板 ref + `defineExpose` 类型 | 能调子组件 focus |
| 8 | `npm run type-check` 通过 | exit code 0 |

---

## 下一章预告

Vue 路线同学可继续把列表页、路由守卫改为 TS，并阅读 [09-工程化与 tsconfig 深入](./09-工程化与tsconfig深入.md) 开启 `strict` 全量修错。

**React 主线**请转 [08 React 与 TypeScript](./08-React与TypeScript.md)：同一套 `Product` 类型，在 `.tsx`、`useState` 泛型、Zustand 中的写法对照学习。两章 **类型文件可复用**，组件层二选一精读即可。

---

*下一章：08 React 与 TypeScript（React 主线）/ 09 工程化与 tsconfig 深入（Vue 续）*
