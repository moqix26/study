# Vue 进阶特性

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0 导读、术语三件套、FAQ≥10、闭卷自测、费曼检验 -->

> **文件编码**：UTF-8。本章在掌握 01～11 主线后学习；不影响 MVP 交付，但影响面试深度与复杂组件设计能力。

---

## 0. 读前导读（零基础也能跟上）

> **读者假设**：01～11 章 shop-vue MVP 能跑；本章是**锦上添花**——日常 80% 业务用不到全部特性，但面试和封装通用组件时会考。

### 0.1 用一句话弄懂本章

**一句话**：Vue 在「写页面」之外还提供 **插槽、缓存、传送门、自定义指令** 等能力，用来封装可复用 UI、优化性能、处理弹层层级。

**生活类比**：

| 概念 | 类比 |
|------|------|
| **插槽 slot** | 相框（子组件）+ 你自选的照片（父组件内容） |
| **KeepAlive** | 浏览器标签页：切走不关闭，回来还在原位置 |
| **Teleport** | 把弹窗「瞬移」到 body，不被父盒子裁剪 |
| **自定义指令** | 给元素贴「自动聚焦」「防抖点击」便利贴 |
| **shallowRef** | 只盯整箱货物换没换，不逐件清点 |

---

### 0.2 你需要提前知道什么

| 水平 | 建议 |
|------|------|
| 01～11 未完成 | 先完成 [11 项目实战](./11-Vue项目实战与面试准备.md)，至少路由+Pinia+Axios 能跑 |
| 只会写单页 | 先复习 [04 组件通信](./04-组件基础与组件通信.md) 的 props/emit |
| 目标面试 | 本章 + [13 场景题](./13-高频场景题与面试专题.md) 一起过 |
| 零基础 | 先完成 [HTML CSS JS 06～08](../HTML%20CSS%20JS/06-JavaScript基础语法与数据类型.md) 再学 Vue |

---

### 0.3 本章知识地图（☐→☑）

- [ ] 会写默认、具名、作用域插槽
- [ ] 会配置 KeepAlive + defineOptions name
- [ ] 会用 Teleport 解决弹层 z-index
- [ ] 能写 v-focus、v-debounce 自定义指令
- [ ] 知道 watch / watchEffect / computed 选型
- [ ] 能列举 5 条以上性能优化手段
- [ ] 能口述 Vue 2/3 响应式差异
- [ ] 闭卷自测 ≥ 8/10

---

### 0.4 建议学习时长

| 阶段 | 时间 |
|------|------|
| §1 插槽 + §2 KeepAlive | 2 小时 |
| §4 Teleport + §5 指令 | 1.5 小时 |
| §6～§10 响应式进阶 | 2 小时 |
| §11 性能 + §12 Vue2/3 对比 | 1.5 小时 |
| shop-vue 改造 + 自测 | 2 小时 |

---

### 0.5 可验证成果

1. 封装 `BaseCard` 三插槽，在 ProductDetail 使用。
2. ProductListView 加 KeepAlive，进详情再返回滚动位置保留。
3. 实现 `v-debounce:500` 用于提交订单按钮。
4. 能向朋友 3 分钟讲清插槽 vs props 的区别。

---

### 0.6 核心术语三件套

**术语（插槽 Slot）**：父组件向子组件 **注入模板片段** 的占位机制。
**生活类比**：子组件是「标准相框」，父组件决定框里放什么画。
**为什么重要**：封装 `BaseCard`、`DataTable` 列自定义离不开插槽。
**本章用到的地方**：§1。

**术语（KeepAlive）**：Vue 内置组件，**缓存不活跃的路由/组件实例**，切换时不销毁。
**生活类比**：手机 App 切后台不杀进程，回来还在原页面。
**为什么重要**：列表↔详情来回跳时保留滚动、筛选条件。
**本章用到的地方**：§2。

**术语（Teleport 传送门）**：把组件 DOM **渲染到指定容器**（通常是 body），脱离父级 overflow 限制。
**生活类比**：把投影仪画面投到另一块幕布上。
**为什么重要**：Modal、Toast 不被父级 `overflow:hidden` 裁切。
**本章用到的地方**：§4。

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

### 12.2 diff 算法口述（面试了解）

Vue 3 同层比较 + key 优化：

1. 比较同一层级的节点类型
2. **key** 相同则复用 DOM，只更新 props/children
3. 不同 type 则替换整棵子树

**生活类比**：班级点名用学号（key）不用座位号（index），换座不乱人。

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

**Q1：插槽和 props 传 VNode 区别？**  
Vue 3 可 props 传 h() vnode，但 slot 更符合模板语义、作用域插槽更灵活。

**Q2：KeepAlive 和 Pinia 缓存列表数据？**  
KeepAlive 缓存 **组件状态**（含 DOM）；Pinia 缓存 **数据**。可组合：KeepAlive 保滚动，Pinia 保筛选条件。

**Q3：还要学 Vue 2 吗？**  
新项目不用；面试可能问迁移差异，看 §12 表即可。

**Q4：作用域插槽 `#default="{ item }"` 里的 item 从哪来？**  
子组件 `<slot :item="row" />` 通过 **props 形式** 传给父组件插槽内容；父组件用解构接收。

**Q5：KeepAlive 的 include 写组件名还是路由名？**  
写 **组件的 name**（`defineOptions({ name: 'Xxx' })`），不是路由 path。

**Q6：Teleport 和 Element Plus append-to-body 一样吗？**  
原理相同：把 DOM 挂到 body；`el-dialog` 默认已做，手写 Modal 才需 Teleport。

**Q7：自定义指令还能用吗？Vue 3 推荐吗？**  
能用；DOM 操作用指令，业务逻辑用 composables。指令适合 **横切关注点**（focus、权限、防抖）。

**Q8：shallowRef 和 ref 什么时候换？**  
大对象、第三方实例（ECharts、地图）用 shallowRef；需要深度监听嵌套属性时用 ref/reactive。

**Q9：watchEffect 和 watch 哪个先执行？**  
同 tick 内 watch 更可控；watchEffect 立即执行一次并自动收集依赖。组件卸载两者都会停止。

**Q10：v-memo 和 computed 区别？**  
v-memo 跳过 **子树 re-render**（模板层）；computed 缓存 **计算结果**（数据层）。大列表项可组合使用。

**Q11：Transition 动画卡顿怎么排查？**  
优先动画 `transform`/`opacity`，避免 `width`/`height`；Chrome Performance 看 Layout/Paint 耗时。

**Q12：进阶特性要全用在 shop-vue 吗？**  
不必。MVP 优先：KeepAlive 列表 + 插槽封装；Teleport/指令/异步组件按需加。

---

## 21. KeepAlive 手把手配置（shop-vue）

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | `ProductListView.vue` 加 `defineOptions({ name: 'ProductListView' })` | 无报错 | 检查 Vue 3.3+ |
| 2 | `App.vue` router-view 外包 `<keep-alive :include="['ProductListView']">` | 切换路由组件被缓存 | include 拼写与 name 一致 |
| 3 | 列表滚到中间 → 进详情 → 浏览器后退 | 滚动位置大致保留 | 检查是否整页 remount |
| 4 | 详情页 **不要** 放进 include | 详情每次 fresh 请求 | exclude 或 v-if 分支 |
| 5 | `activated` 里可选刷新 stale 数据 | 返回列表时 badge 更新 | 对比 mounted 只执行一次 |

---

## 22. 插槽 BaseCard 逐行读（>10 行示例）

```vue
<template>
  <div class="card">
    <header v-if="$slots.header"><slot name="header" /></header>
    <main><slot /></main>
    <footer v-if="$slots.footer"><slot name="footer" /></footer>
  </div>
</template>
```

| 行号/字段 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `v-if="$slots.header"` | 父没传 header 时不渲染空条 | 去掉会留空白 header 占高度 |
| `<slot name="header" />` | 具名插槽占位 | name 拼错父组件 `#header` 不显示 |
| `<slot />` | 默认插槽（匿名） | 父组件直接写子内容即可 |
| `$slots.footer` | 运行时插槽对象，判断有无内容 | 与 `v-if` 配合避免空 footer |

---

## 23. 闭卷自测

1. 默认插槽、具名插槽、作用域插槽各解决什么问题？
2. KeepAlive 的 include 匹配什么？script setup 如何声明 name？
3. Teleport 典型使用场景是什么？
4. 自定义指令 mounted 和 unmounted 为什么要成对解绑事件？
5. ref 解构 reactive 为什么会丢响应式？怎么解决？
6. shallowRef 适合什么类型的数据？
7. watch、watchEffect、computed 各适合什么场景？（各举一例）
8. **动手**：写三插槽 BaseCard 并在父组件填充 header/footer。
9. **动手**：给 ProductListView 加 KeepAlive，验证滚动保留。
10. **综合**：列举 5 条 Vue 3 性能优化手段并说明 shop-vue 已用哪几条。

### 23.1 自测参考答案

1. 默认：父注入正文；具名：多区域（header/footer）；作用域：子传数据给父自定义渲染。
2. 匹配 **组件 name**；`defineOptions({ name: 'ProductListView' })`。
3. Modal/Drawer 挂 body，避免父级 overflow/z-index 问题。
4. 防止内存泄漏和重复绑定；unmounted 必须 removeEventListener。
5. 解构取的是快照；用 `toRefs(state)` 或 `storeToRefs`。
6. 大对象、第三方实例、整体替换才更新的数据（ECharts option）。
7. computed：totalPrice；watch：route.params.id 拉详情；watchEffect：document.title。
8. 见 §1.3、§1.5 与 §21 步骤。
9. 见 §21 手把手表。
10. 路由懒加载、KeepAlive、computed 缓存、Element 按需、图片 lazy 等；shop-vue 至少 lazy+拦截器+分页。

---

## 24. 费曼检验

请 3 分钟向没学过 Vue 的朋友解释「为什么列表页需要 KeepAlive」：

1. **问题**：SPA 切路由会销毁组件，滚动位置和表单输入丢失。
2. **方案**：KeepAlive 像「标签页不关闭」，缓存组件实例和 DOM 状态。
3. **注意**：要声明组件 name；详情页一般不要缓存；可与 Pinia 存筛选条件配合。

---

## 下一章预告

进阶特性补完了——下一章（13 高频场景题与面试专题）专门练 **面试官怎么问、你怎么答**：15+ 题带完整回答框架，每题尽量结合 shop-vue。

---

*下一章：13 高频场景题与面试专题*
