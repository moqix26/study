# Vue 进阶特性

> **文件编码**：UTF-8。本章在掌握 01～11 主线后学习；不影响 MVP 交付，但影响面试深度与复杂组件设计能力。

---

## 本章与上一章的关系

11 章商城项目把 **Vue Router、Pinia、Axios、Element Plus** 等主线技术都用上了。日常业务 80% 不需要本章全部特性——但在以下场景会用到：

- 封装 **可复用 UI 组件**（Card、Table 列自定义）→ 插槽
- **列表 ↔ 详情** 来回切换保留滚动位置 → KeepAlive
- **Modal 弹层** DOM 层级与 body 滚动 → Teleport
- **自动聚焦、防抖、权限按钮** → 自定义指令
- 面试追问 **性能优化、Vue 2/3 差异、响应式边界**

本章按「能写 demo → 知道原理 → 面试能讲」组织。

---

## 1. 插槽 slot（默认 / 具名 / 作用域）

### 1.1 为什么需要插槽

**props** 只能传数据；**slot** 让父组件向子组件 **注入模板内容**，实现「结构由父决定、外壳由子提供」——类似 React 的 `children` render props。

典型场景：通用 `BaseCard`、`DataTable` 自定义列、Layout 侧边栏。

### 1.2 默认插槽

```vue
<!-- BaseCard.vue -->
<template>
  <div class="card">
    <div class="card-body">
      <slot>默认内容（父不传时显示）</slot>
    </div>
  </div>
</template>

<!-- 父组件 -->
<BaseCard>
  <p>这是插入的正文</p>
</BaseCard>
```

### 1.3 具名插槽

```vue
<!-- BaseCard.vue -->
<template>
  <div class="card">
    <header v-if="$slots.header" class="card-header">
      <slot name="header" />
    </header>
    <main class="card-main">
      <slot />
    </main>
    <footer v-if="$slots.footer" class="card-footer">
      <slot name="footer" />
    </footer>
  </div>
</template>

<!-- 父组件：v-slot:header 简写 #header -->
<BaseCard>
  <template #header>
    <h3>{{ product.name }}</h3>
  </template>
  <p>{{ product.description }}</p>
  <template #footer>
    <el-button type="primary" @click="addCart">加入购物车</el-button>
  </template>
</BaseCard>
```

**`$slots`**：判断父是否传入某插槽，避免空 header 占高度。

### 1.4 作用域插槽（scoped slot）

子组件 **把内部数据交给父** 决定如何渲染：

```vue
<!-- ProductTable.vue 子 -->
<template>
  <table>
    <tr v-for="row in products" :key="row.id">
      <td><slot name="name" :item="row">{{ row.name }}</slot></td>
      <td><slot :item="row" /></td>
    </tr>
  </table>
</template>

<!-- 父：自定义价格列样式 -->
<ProductTable :products="list">
  <template #default="{ item }">
    <span class="price">¥{{ item.price.toFixed(2) }}</span>
  </template>
</ProductTable>
```

Element Plus 的 `el-table-column` 的 `#default="{ row }"` 就是作用域插槽模式。

### 1.5 插槽与 shop-vue 结合

| 组件 | 插槽设计 |
|------|----------|
| ProductCard | `#footer` 放加购按钮 |
| PageLayout | `#sidebar` / `#default` |
| EmptyState | `#action` 放「去逛逛」按钮 |

---

## 2. KeepAlive 缓存组件实例

### 2.1 问题场景

用户在产品列表滚到第 5 页 → 点进详情 → 返回列表，**滚动位置丢失、重新请求**。

`<KeepAlive>` 缓存不活跃的组件实例，切换路由时不销毁。

### 2.2 基本用法

```vue
<!-- App.vue 或 Layout -->
<router-view v-slot="{ Component, route }">
  <keep-alive :include="cachedViews">
    <component :is="Component" :key="route.fullPath" />
  </keep-alive>
</router-view>
```

```js
// 仅缓存列表页
const cachedViews = ['ProductListView']
```

**script setup 必须声明 name**（KeepAlive 的 include 匹配组件 name）：

```vue
<script setup>
defineOptions({ name: 'ProductListView' })
</script>
```

### 2.3 include / exclude / max

| 属性 | 说明 |
|------|------|
| `include` | 字符串或数组，匹配的 **组件 name** 才缓存 |
| `exclude` | 排除不缓存 |
| `max` | 最多缓存 N 个，LRU 淘汰 |

### 2.4 与路由 meta 配合

```js
{
  path: '/products',
  component: () => import('@/views/ProductListView.vue'),
  meta: { keepAlive: true },
}
```

```vue
<keep-alive>
  <component :is="Component" v-if="route.meta.keepAlive" />
</keep-alive>
<component :is="Component" v-if="!route.meta.keepAlive" />
```

### 2.5 生命周期：activated / deactivated

缓存组件不会反复 `mounted`，而是：

- 进入：`activated`
- 离开：`deactivated`

可在 `activated` 里刷新 stale 数据。

### 2.6 注意点

- 缓存过多占内存；列表页缓存、详情页一般不缓存
- `:key="route.fullPath"` 同一组件不同 id 需不同 key 时慎用全 path

---

## 3. 动态组件 component :is

### 3.1 切换 Tab 不走路由

```vue
<script setup>
import { ref, shallowRef } from 'vue'
import TabProducts from './TabProducts.vue'
import TabOrders from './TabOrders.vue'

const tabs = {
  products: TabProducts,
  orders: TabOrders,
}
const currentTab = ref('products')
</script>

<template>
  <el-radio-group v-model="currentTab">
    <el-radio-button value="products">商品</el-radio-button>
    <el-radio-button value="orders">订单</el-radio-button>
  </el-radio-group>
  <keep-alive>
    <component :is="tabs[currentTab]" />
  </keep-alive>
</template>
```

**`shallowRef`**：组件对象不需深度响应式，用 `shallowRef` 存组件定义更高效。

---

## 4. Teleport 传送门

### 4.1 问题

Modal 写在组件内，可能受父级 `overflow: hidden`、`z-index` 影响，遮罩盖不全。

### 4.2 用法

```vue
<template>
  <button @click="visible = true">打开</button>
  <teleport to="body">
    <div v-if="visible" class="modal-mask" @click.self="visible = false">
      <div class="modal-box">弹窗内容</div>
    </div>
  </teleport>
</template>
```

`to` 可为 CSS 选择器：`#modal-root`、`body`。

Element Plus 的 `el-dialog` 默认 `append-to-body`，内部类似 Teleport。

### 4.3 disabled

`<teleport disabled>` 留在原位置（调试或 SSR 场景）。

---

## 5. 自定义指令

### 5.1 全局注册

```js
// directives/focus.js
export const vFocus = {
  mounted(el) {
    el.focus()
  },
}

// main.js
import { vFocus } from './directives/focus'
app.directive('focus', vFocus)
```

```vue
<input v-focus />
```

### 5.2 钩子函数

| 钩子 | 时机 |
|------|------|
| `created` | 绑定到元素前 |
| `beforeMount` | 元素挂载前 |
| `mounted` | 元素挂载后 |
| `beforeUpdate` | 组件更新前 |
| `updated` | 组件更新后 |
| `beforeUnmount` | 卸载前 |
| `unmounted` | 卸载后 |

### 5.3 v-debounce 点击防抖

```js
// directives/debounce.js
export const vDebounce = {
  mounted(el, binding) {
    const delay = Number(binding.arg) || 300
    let timer = null
    el._debounceHandler = () => {
      clearTimeout(timer)
      timer = setTimeout(() => {
        binding.value?.()
      }, delay)
    }
    el.addEventListener('click', el._debounceHandler)
  },
  unmounted(el) {
    el.removeEventListener('click', el._debounceHandler)
  },
}
```

```vue
<el-button v-debounce:500="submitOrder">提交订单</el-button>
```

### 5.4 v-permission 权限（示例）

```js
export const vPermission = {
  mounted(el, binding) {
    const userStore = useUserStore()
    const roles = binding.value // ['admin']
    if (!roles.includes(userStore.role)) {
      el.parentNode?.removeChild(el)
    }
  },
}
```

### 5.5 Vue 3 与 Vue 2 指令差异

- 钩子名：`inserted` → `mounted`，`bind` → `beforeMount`
- 组件实例：`binding.instance` 访问当前组件

---

## 6. watch vs watchEffect vs computed

### 6.1 watch（明确数据源）

```js
watch(
  () => route.params.id,
  (newId) => fetchDetail(newId),
  { immediate: true }
)

watch(keyword, debounceFn, { flush: 'post' })
```

### 6.2 watchEffect（自动收集依赖）

```js
watchEffect(() => {
  document.title = `${product.value?.name || ''} - Shop`
})
```

组件卸载时自动停止。适合副作用：改标题、同步 localStorage。

### 6.3 选型

| 需求 | 用 |
|------|-----|
| 派生展示值 | computed |
| 异步请求、日志 | watch |
| 多源自动追踪副作用 | watchEffect |

---

## 7. provide / inject 跨层通信

避免 props 层层传递（如主题色、locale）：

```js
// App.vue 或 Layout
provide('shopConfig', { currency: 'CNY', theme: 'light' })

// 深层 ProductPrice.vue
const config = inject('shopConfig')
```

Pinia 能替代的不用 provide；provide 适合 **插件级、不常变** 的配置。

---

## 8. 异步组件与 Suspense

### 8.1 defineAsyncComponent

```js
import { defineAsyncComponent } from 'vue'

const HeavyChart = defineAsyncComponent({
  loader: () => import('./HeavyChart.vue'),
  loadingComponent: PageLoading,
  delay: 200,
  timeout: 10000,
})
```

### 8.2 Suspense（实验性，了解即可）

```vue
<Suspense>
  <template #default>
    <AsyncSetupComponent />
  </template>
  <template #fallback>
    <PageLoading />
  </template>
</Suspense>
```

---

## 9. 响应式进阶：ref / reactive / toRefs / shallow

### 9.1 reactive 解构丢失响应式

```js
const state = reactive({ count: 0 })
const { count } = state  // 丢失响应式
const { count } = toRefs(state)  // count 是 ref
```

### 9.2 shallowRef / shallowReactive

大对象、第三方实例（ECharts）不需深度监听：

```js
const chartOption = shallowRef({ series: [...] })
// 整体替换才触发更新
chartOption.value = newOption
```

### 9.3 readonly / markRaw

```js
const copy = readonly(state)  // 不可改
const foo = markRaw({ huge: data })  // 永不转响应式
```

---

## 10. 模板 ref 与 expose

### 10.1 模板 ref

```vue
<script setup>
import { ref, onMounted } from 'vue'
const inputRef = ref(null)
onMounted(() => inputRef.value.focus())
</script>
<template>
  <input ref="inputRef" />
</template>
```

### 10.2 子组件 expose

```vue
<!-- Child.vue -->
<script setup>
const validate = () => { /* ... */ }
defineExpose({ validate })
</script>

<!-- Parent.vue -->
<script setup>
const childRef = ref(null)
const submit = () => childRef.value.validate()
</script>
<template>
  <Child ref="childRef" />
</template>
```

Element Plus 表单 `formRef.validate()` 即此模式。

---

## 11. 性能优化完整清单

### 11.1 渲染层

| 手段 | 说明 | 示例 |
|------|------|------|
| v-show vs v-if | 频繁切换 show | Tab |
| v-memo | 跳过子树更新（3.2+） | 大列表项 |
| 稳定 :key | 用 id 不用 index | v-for |
| 避免模板复杂表达式 | 用 computed | 过滤排序 |

### 11.2 组件层

| 手段 | 说明 |
|------|------|
| 路由懒加载 | `() => import()` |
| KeepAlive | 列表缓存 |
| 异步组件 | 重组件按需 |
| Functional 组件 | 无状态纯展示（少用） |

### 11.3 数据层

| 手段 | 说明 |
|------|------|
| shallowRef | 大对象 |
| computed 缓存 | 派生数据 |
| 虚拟列表 | vue-virtual-scroller、el-table-v2 |
| 分页 | 不要一次渲染万行 |

### 11.4 构建层（见 10 章）

- manualChunks
- Element Plus 按需
- 图片懒加载 `loading="lazy"`

### 11.5 运行时分析

Vue DevTools → Performance  
Chrome Performance 录制  
`app.config.performance = true`（开发）

---

## 12. Vue 3 vs Vue 2 对比（面试必背）

| 维度 | Vue 2 | Vue 3 |
|------|-------|-------|
| 响应式 | Object.defineProperty | Proxy |
| API 风格 | Options API 为主 | Composition API |
| 状态管理 | Vuex | Pinia |
| 构建工具 | Webpack (Vue CLI) | Vite |
| 根节点 | 单根 | 多根 Fragment |
| 生命周期 | beforeDestroy | beforeUnmount |
| v-model | 组件 model 选项 | modelValue + update:modelValue |
| 全局 API | new Vue() | createApp() |
| TypeScript | 支持一般 | 原生友好 |
| 体积 | 较大 | Tree-shaking 更小 |

### 12.1 Proxy 优势口述

- 监听新增/删除属性
- 监听数组索引和 length
- 惰性递归，性能更好

---

## 13. Composition API vs Options API

| Options | Composition |
|---------|-------------|
| data/methods 分散 | 按逻辑组织 |
| 大组件难维护 | composables 复用 |
| mixin 命名冲突 | 函数组合 |

**本资料默认 script setup**；维护 Vue 2 老项目才需熟练 Options。

---

## 14. 内置特殊元素与属性

| 特性 | 说明 |
|------|------|
| `<Transition>` | 单元素进出场动画 |
| `<TransitionGroup>` | 列表动画 |
| `v-once` | 只渲染一次 |
| `v-pre` | 跳过编译 |
| `v-cloak` | 防闪烁（配合 CSS） |

列表动画示例：

```vue
<TransitionGroup name="list" tag="ul">
  <li v-for="item in items" :key="item.id">{{ item.name }}</li>
</TransitionGroup>
```

```css
.list-enter-active, .list-leave-active { transition: all 0.3s; }
.list-enter-from, .list-leave-to { opacity: 0; transform: translateX(30px); }
```

---

## 15. SSR / Nuxt 入门（扩展阅读）

- **CSR**（本资料）：浏览器下载 JS 再渲染，SEO 弱
- **SSR**：服务端出 HTML，Nuxt 3 基于 Vue 3
- 面试：知道区别即可，初级不强制会 Nuxt

---

## 16. shop-vue 进阶改造建议

| 改造 | 章节技术 | 优先级 |
|------|----------|--------|
| ProductCard 插槽化 footer | §1 | 中 |
| 列表 KeepAlive | §2 | 高 |
| 结算按钮 v-debounce | §5 | 中 |
| 商品大图 Teleport 预览 | §4 | 低 |
| 订单 Tab 动态组件 | §3 | 低 |

---

## 17. 常见报错

| 报错 | 原因 | 解决 |
|------|------|------|
| KeepAlive 不生效 | 无 defineOptions name | 声明 name 与 include 一致 |
| slot 内容不显示 | 拼写 `#header` vs `#head` | 对齐 slot name |
| Teleport 目标不存在 | to 选择器无匹配 | 确保 body 上有节点 |
| 指令 binding.value 非函数 | debounce 未传回调 | `v-debounce="fn"` |

---

## 18. 学完标准

- [ ] 会写默认、具名、作用域插槽
- [ ] 会配置 KeepAlive + defineOptions name
- [ ] 了解 Teleport、动态组件使用场景
- [ ] 能写简单自定义指令（focus、debounce）
- [ ] 能列举 5 条以上性能优化手段
- [ ] 能口述 Vue 2/3 响应式与生态差异
- [ ] 知道 watch / watchEffect / computed 选型

---

## 19. 分级练习

### 基础

写 `BaseCard`：header 插槽 + 默认插槽 + footer 插槽，在 ProductDetail 中使用。

### 进阶

`ProductListView` 加 KeepAlive，滚到中间进详情再返回，滚动位置保留。

### 挑战

实现 `v-debounce` 指令，用于「提交订单」防重复点击；写 `useInfiniteScroll` composable（IntersectionObserver）。

### 参考答案（挑战 v-debounce）

见 §5.3 完整代码；测试：快速连点只触发一次 submit。

---

## 20. FAQ

**Q：插槽和 props 传 VNode 区别？**  
Vue 3 可 props 传 h()  vnode，但 slot 更符合模板语义、作用域插槽更灵活。

**Q：KeepAlive 和 Vuex/Pinia 缓存列表数据？**  
KeepAlive 缓存 **组件状态**（含 DOM）；Pinia 缓存 **数据**。可组合：KeepAlive 保滚动，Pinia 保筛选条件。

**Q：还要学 Vue 2 吗？**  
新项目不用；面试可能问迁移差异，看 §12 表即可。

---

## 下一章预告

进阶特性补完了——下一章（13 高频场景题与面试专题）专门练 **面试官怎么问、你怎么答**：15+ 题带完整回答框架，每题尽量结合 shop-vue。

---

*下一章：13 高频场景题与面试专题*
