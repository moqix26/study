# React 进阶特性

> **文件编码**：UTF-8。本章在掌握 01～11 主线后学习；不影响 MVP 交付，但影响面试深度与复杂组件设计能力。

---

## 本章与上一章的关系

11 章商城项目把 **React Router、Zustand/Context、Axios、Ant Design** 等主线技术都用上了。日常业务 80% 不需要本章全部特性——但在以下场景会用到：

- **主题 / 语言 / 用户权限** 跨多层传递 → Context API
- **大列表子组件** 父 state 变就全量重渲染 → `React.memo` + `useCallback`
- **重组件按需加载** → `React.lazy` + `Suspense`
- **Modal 弹层** DOM 层级与 body 滚动 → Portals
- **接口错误、子树崩溃** 不拖垮整页 → Error Boundary
- 面试追问 **性能优化、Hooks 规则、React 18 并发特性**

本章按「能写 demo → 知道原理 → 面试能讲」组织。

---

## 1. Context API 跨层传递

### 1.1 为什么需要 Context

**props 层层传递（prop drilling）** 在深层组件需要主题、语言、登录用户时很痛苦。Context 提供 **「生产者—消费者」** 模式，任意深度的子组件可直接读取，无需中间层转发。

典型场景：主题色、国际化 locale、当前用户信息、购物车摘要（小项目可用 Context，大项目仍推荐 Zustand）。

### 1.2 基本用法

```jsx
// contexts/ThemeContext.jsx
import { createContext, useContext, useState } from 'react'

const ThemeContext = createContext(null)

export function ThemeProvider({ children }) {
  const [theme, setTheme] = useState('light')
  const value = { theme, setTheme, isDark: theme === 'dark' }
  return (
    <ThemeContext.Provider value={value}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useTheme 必须在 ThemeProvider 内使用')
  return ctx
}
```

```jsx
// main.jsx
import { ThemeProvider } from './contexts/ThemeContext'

createRoot(document.getElementById('root')).render(
  <ThemeProvider>
    <App />
  </ThemeProvider>
)
```

```jsx
// 深层 ProductCard.jsx — 无需 props 透传
import { useTheme } from '../contexts/ThemeContext'

function ProductCard({ product }) {
  const { isDark } = useTheme()
  return (
    <div className={isDark ? 'card dark' : 'card'}>
      {product.name}
    </div>
  )
}
```

### 1.3 多个 Context 与性能

多个 Provider 可抽成 `AppProviders` 单文件。Context **value 变化时所有消费者都会 re-render**。

**优化手段**：

| 手段 | 说明 |
|------|------|
| 拆分 Context | `ThemeContext` 与 `UserContext` 分开 |
| `useMemo` 稳定 value | `const value = useMemo(() => ({ theme, setTheme }), [theme])` |
| 状态下沉 | 频繁变的 state 不放 Context，用 Zustand |
| 选择器模式 | 第三方如 `use-context-selector` |

```jsx
const value = useMemo(
  () => ({ theme, setTheme }),
  [theme]
)
```

### 1.4 Context vs Zustand vs props

| 方式 | 适用 |
|------|------|
| props | 父子 1～2 层 |
| Context | 主题、locale、少变配置；中小型全局 |
| Zustand / Redux | 购物车、用户、频繁更新、需 devtools |

**shop-react**：`userStore`、`cartStore` 用 Zustand；`ThemeContext` 可选做暗色模式。

### 1.5 与 Vue provide/inject 对照

| React Context | Vue provide/inject |
|---------------|-------------------|
| `createContext` + `Provider` | `provide('key', value)` |
| `useContext` | `inject('key')` |
| value 变全量订阅者更新 | 非响应式 provide 需 reactive 包装 |

---

## 2. children、Render Props 与组合

### 2.1 children（类似 Vue 默认插槽）

```jsx
function BaseCard({ children, title }) {
  return (
    <div className="card">
      {title && <header className="card-header">{title}</header>}
      <main className="card-body">{children}</main>
    </div>
  )
}

// 使用
<BaseCard title="商品详情">
  <p>这是插入的正文</p>
</BaseCard>
```

### 2.2 Render Props（类似 Vue 作用域插槽）

子组件把 **内部数据交给父** 决定如何渲染：

```jsx
function ProductTable({ products, renderRow }) {
  return (
    <table>
      <tbody>
        {products.map((item) => (
          <tr key={item.id}>
            <td>{renderRow(item)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

// 使用
<ProductTable
  products={list}
  renderRow={(item) => (
    <span className="price">¥{item.price.toFixed(2)}</span>
  )}
/>
```

也可用 **children 作为函数**：

```jsx
function ProductTable({ products, children }) {
  return (
    <table>
      <tbody>
        {products.map((item) => (
          <tr key={item.id}>
            <td>{children(item)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

<ProductTable products={list}>
  {(item) => <span>¥{item.price}</span>}
</ProductTable>
```

Ant Design `Table` 的 `columns.render` 即类似模式。

### 2.3 复合组件模式（了解）

通过 Context 把 `Tabs`、`Tabs.Tab`、`Tabs.Panel` 组合使用（类似 Ant Design Tabs 内部结构），进阶封装时再深入即可。

---

## 3. React.memo 避免无效重渲染

### 3.1 问题场景

父组件 `ProductList` 的 `keyword` 变化时，列表里每个 `ProductCard` 都会 re-render，即使 `product` props 没变。

### 3.2 基本用法

```jsx
import { memo } from 'react'

const ProductCard = memo(function ProductCard({ product, onAddCart }) {
  console.log('render', product.id)
  return (
    <div className="card">
      <h3>{product.name}</h3>
      <button onClick={() => onAddCart(product)}>加入购物车</button>
    </div>
  )
})
```

`memo` 对 props 做 **浅比较**（`Object.is`），全等则跳过 render。

### 3.3 自定义比较函数

```jsx
const ProductCard = memo(
  function ProductCard({ product }) {
    return <div>{product.name}</div>
  },
  (prev, next) => prev.product.id === next.product.id
)
```

第二个参数返回 `true` 表示 **props 相等、跳过更新**（与 `shouldComponentUpdate` 逻辑相反，易混，少用）。

### 3.4 memo 失效的常见原因

| 原因 | 说明 |
|------|------|
| 内联对象/数组 props | 每次父 render 都是新引用 |
| 内联函数 props | `onClick={() => fn()}` 每次都是新函数 |
| Context 消费 | 子组件用了变化的 Context 仍会更新 |
| 子组件自身 state 变 | memo 只挡 props 变化 |

**解决**：父组件用 `useCallback` / `useMemo` 稳定引用（见 §4）。

### 3.5 何时不用 memo

- 组件 render 极轻量（几个文本节点）
- props 几乎每次都变
- 过早优化增加代码复杂度

**原则**：先测量（React DevTools Profiler），再优化。

---

## 4. useCallback 与 useMemo 性能优化

### 4.1 理解 re-render 机制

React 函数组件：**任意 state 或父组件 re-render，默认子组件跟着 render**。`useCallback` / `useMemo` 不是减少计算那么简单，核心是 **稳定引用**，让 `memo` 和依赖数组生效。

### 4.2 useMemo — 缓存计算结果

```jsx
import { useMemo } from 'react'

function ProductList({ products, keyword }) {
  const filtered = useMemo(() => {
    const kw = keyword.trim().toLowerCase()
    if (!kw) return products
    return products.filter((p) =>
      p.name.toLowerCase().includes(kw)
    )
  }, [products, keyword])

  return (
    <ul>
      {filtered.map((p) => (
        <ProductCard key={p.id} product={p} />
      ))}
    </ul>
  )
}
```

**适用**：过滤、排序、复杂派生数据；避免每次 render 重新 `filter` 大数组。

**不适用**：简单 `a + b`；`useMemo` 自身有开销。

### 4.3 useCallback — 缓存函数引用

```jsx
import { useCallback } from 'react'

function ProductList({ products }) {
  const addToCart = useCartStore((s) => s.addItem)

  const handleAddCart = useCallback(
    (product) => {
      addToCart(product)
    },
    [addToCart]
  )

  return products.map((p) => (
    <ProductCard
      key={p.id}
      product={p}
      onAddCart={handleAddCart}
    />
  ))
}
```

没有 `useCallback` 时，每次 `ProductList` render 都会创建新的 `onAddCart`，`memo(ProductCard)` 失效。

### 4.4 与 Vue computed 对照

| React | Vue |
|-------|-----|
| `useMemo` | `computed`（缓存派生值） |
| `useCallback` | 无直接对应；稳定方法引用 |
| 手动声明依赖数组 | 自动依赖追踪 |

React 需 **显式写依赖**；漏写导致 stale closure（闭包陈旧值）是常见 bug。

### 4.5 依赖数组最佳实践

```jsx
// ❌ 漏依赖 — keyword 变了仍用旧值
useEffect(() => {
  fetchList(keyword)
}, [])

// ✅
useEffect(() => {
  fetchList(keyword)
}, [keyword])
```

ESLint `react-hooks/exhaustive-deps` 务必开启。

### 4.6 shop-react 优化清单

| 位置 | 优化 |
|------|------|
| ProductCard | `memo` + 稳定 `onAddCart` |
| Cart 总价 | `useMemo` 或 Zustand selector |
| 搜索过滤 | `useMemo` 本地过滤或防抖请求 |
| Header 购物车数量 | Zustand 细粒度订阅 |

---

## 5. forwardRef 与 useImperativeHandle（了解）

父组件通过 `forwardRef` 把 ref 转给子组件内部 DOM；`useImperativeHandle` 只暴露 `focus`、`validate` 等方法（类似 Vue `defineExpose`）。Ant Design `Form` 的 `form.validateFields()` 即此模式。详情见 05 章。

---

## 6. Error Boundary 错误边界

### 6.1 问题

子组件 throw（接口渲染崩溃、第三方库错误），默认会导致 **整棵组件树卸载、白屏**。

Error Boundary 是 **class 组件**（截至 React 19 仍无 Hook 版官方 API），捕获子树渲染期错误，展示降级 UI。

### 6.2 实现

```jsx
import { Component } from 'react'

class ErrorBoundary extends Component {
  state = { hasError: false, error: null }

  static getDerivedStateFromError(error) {
    return { hasError: true, error }
  }

  componentDidCatch(error, info) {
    console.error('ErrorBoundary caught:', error, info.componentStack)
    // 上报 Sentry 等
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback ?? (
        <div className="error-page">
          <h2>页面出错了</h2>
          <button onClick={() => this.setState({ hasError: false })}>
            重试
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
```

```jsx
// 使用：按路由或模块包裹
<ErrorBoundary fallback={<ProductListError />}>
  <ProductList />
</ErrorBoundary>
```

### 6.3 能捕获 vs 不能捕获

| 能捕获 | 不能捕获 |
|--------|----------|
| 子组件 render 报错 | 事件处理器内错误（用 try/catch） |
| 生命周期报错 | 异步 setTimeout/Promise（需在回调里处理） |
| 子 Boundary 以下 | 自身 Boundary 的 render |

### 6.4 shop-react 建议布局

```text
App
├── ErrorBoundary（全局）
│   ├── Layout
│   │   ├── ErrorBoundary（商品模块）
│   │   │   └── ProductRoutes
│   │   └── CartRoutes
```

**面试话术**：Error Boundary 是 **渲染兜底**；接口错误用 **状态 + 空态组件** 更常见。

### 6.5 与 Vue 对照

Vue 3 无内置 Error Boundary；可用 `app.config.errorHandler` 全局捕获，或 `onErrorCaptured` 组合式钩子做类似降级。

---

## 7. React.lazy 与 Suspense

### 7.1 路由级懒加载

```jsx
import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import PageLoading from './components/PageLoading'

const ProductList = lazy(() => import('./pages/ProductList'))
const ProductDetail = lazy(() => import('./pages/ProductDetail'))
const Cart = lazy(() => import('./pages/Cart'))

function AppRoutes() {
  return (
    <Suspense fallback={<PageLoading />}>
      <Routes>
        <Route path="/products" element={<ProductList />} />
        <Route path="/products/:id" element={<ProductDetail />} />
        <Route path="/cart" element={<Cart />} />
      </Routes>
    </Suspense>
  )
}
```

`lazy` 接收返回 `import()` 的函数；组件首次渲染时触发加载，期间显示 `fallback`。

### 7.2 组件级懒加载

```jsx
const HeavyChart = lazy(() => import('./HeavyChart'))

function Dashboard() {
  const [show, setShow] = useState(false)
  return (
    <div>
      <button onClick={() => setShow(true)}>显示图表</button>
      {show && (
        <Suspense fallback={<Spin />}>
          <HeavyChart />
        </Suspense>
      )}
    </div>
  )
}
```

外层 `PageLoading`、内层 `Skeleton` 可嵌套 Suspense，分别兜底整页与局部。

### 7.3 与 Vue 对照

| React | Vue |
|-------|-----|
| `React.lazy` + `import()` | `() => import()` 路由懒加载 |
| `Suspense` + `fallback` | `defineAsyncComponent` + loading 组件 |
| Suspense 支持数据请求（实验） | Vue Suspense 仍实验性 |

### 7.4 打包注意

Vite / Webpack 自动为每个 `import()` 生成独立 chunk；`build` 后检查 `dist/assets` 体积分布。

---

## 8. Portals 传送门

### 8.1 问题

Modal 写在 `ProductCard` 内，可能受父级 `overflow: hidden`、`z-index`、`transform` 影响，遮罩盖不全。

### 8.2 用法

```jsx
import { createPortal } from 'react-dom'

function ImagePreview({ src, onClose }) {
  if (!src) return null
  return createPortal(
    <div className="modal-mask" onClick={onClose}>
      <div className="modal-box" onClick={(e) => e.stopPropagation()}>
        <img src={src} alt="preview" />
      </div>
    </div>,
    document.body
  )
}
```

事件冒泡仍按 **React 树** 传播（不是 DOM 树），`onClose` 在父组件注册仍可收到。

### 8.3 Ant Design Modal

`Modal` 默认 `getContainer={() => document.body}`，内部即 Portal 思路。

### 8.4 与 Vue Teleport 对照

| React Portals | Vue Teleport |
|---------------|--------------|
| `createPortal(children, domNode)` | `<teleport to="body">` |
| 需手动指定 DOM 节点 | `to` 可为选择器 |

---

## 9. React 18 并发特性入门

### 9.1 什么是并发渲染（Concurrent Rendering）

React 18 引入 **可中断、可优先级调度** 的渲染模式。用户交互（输入、点击）可 **插队** 优先于低优先级更新（大列表渲染、图表），减少卡顿感。

启用方式：根节点用 `createRoot`（非 legacy `ReactDOM.render`）：

```jsx
import { createRoot } from 'react-dom/client'

createRoot(document.getElementById('root')).render(<App />)
```

### 9.2 useTransition — 标记低优先级更新

```jsx
import { useState, useTransition } from 'react'

function ProductSearch({ products }) {
  const [keyword, setKeyword] = useState('')
  const [filtered, setFiltered] = useState(products)
  const [isPending, startTransition] = useTransition()

  const handleChange = (e) => {
    const value = e.target.value
    setKeyword(value) // 高优先级：输入框立即响应
    startTransition(() => {
      // 低优先级：过滤可被打断
      const kw = value.toLowerCase()
      setFiltered(
        products.filter((p) => p.name.toLowerCase().includes(kw))
      )
    })
  }

  return (
    <div>
      <input value={keyword} onChange={handleChange} />
      {isPending && <span>筛选中...</span>}
      <ProductList items={filtered} />
    </div>
  )
}
```

**场景**：搜索框 + 大列表过滤；Tab 切换重组件。

### 9.3 useDeferredValue — 延迟派生值

`const deferredKeyword = useDeferredValue(keyword)` 得到略滞后的值，用于过滤大列表；`keyword !== deferredKeyword` 时可降低 opacity 提示「更新中」。与 `useTransition` 类似，但延迟的是 **值** 而非包裹 setState。

### 9.4 Suspense 与数据获取（了解）

React 18 起 Suspense 可配合支持 Suspense 的数据源（Relay、Next.js App Router 等）。传统 Axios + useEffect 项目 **不必强行上 Suspense 取数**；知道概念即可。

### 9.5 StrictMode 双重调用

开发环境 `StrictMode` 会 **故意双调** `useEffect`、部分生命周期，帮助发现副作用问题。生产环境不会。

```jsx
<StrictMode>
  <App />
</StrictMode>
```

**面试**：说明这是开发期检测机制，不是 bug。

### 9.6 自动批处理（Automatic Batching）

React 18 在 **setTimeout、Promise、原生事件** 里的多次 `setState` 也会批处理为一次 render（17 仅事件处理器内批处理）。

---

## 10. 列表缓存与状态保留（对比 Vue KeepAlive）

React **无内置 KeepAlive**。常见替代：

| 方案 | 说明 |
|------|------|
| 状态提升 / URL 存 scroll | 列表 scroll 存 `sessionStorage` 或 query |
| `react-router` + 自定义 Outlet 缓存 | 第三方 `react-activation` |
| Zustand 存筛选条件 | 返回列表恢复 keyword、page |
| TanStack Query `keepPreviousData` | 分页切换保留上一页 |

离开列表时用 `sessionStorage` 存 `scrollY`，返回时 `window.scrollTo` 恢复；或用 Zustand 存 `keyword`、`page`。

**面试对比**：Vue `<KeepAlive>` 缓存组件实例；React 需组合方案，这是两框架差异点之一。

---

## 11. 性能优化完整清单

### 11.1 渲染层

| 手段 | 说明 | 示例 |
|------|------|------|
| `key` 稳定 | 用 id 不用 index | 商品列表 |
| 状态下沉 | 输入状态放叶子组件 | 搜索框 |
| `memo` | 纯展示子组件 | ProductCard |
| `useMemo` / `useCallback` | 稳定引用与计算 | 过滤列表 |
| 虚拟列表 | `react-window`、`antd Table` 虚拟滚动 | 万行数据 |
| 分页 | 不要一次渲染全部 | 商品列表 |

### 11.2 组件层

| 手段 | 说明 |
|------|------|
| `React.lazy` | 路由、重组件按需 |
| Code Splitting | 动态 import |
| Error Boundary | 局部降级 |
| 避免 Context 大对象 | 拆分 + useMemo value |

### 11.3 数据层

| 手段 | 说明 |
|------|------|
| Zustand selector | 只订阅需要的字段 |
| React Query / SWR | 缓存、去重、后台刷新 |
| 防抖搜索 | 300ms 再请求 |
| AbortController | 取消过时请求 |

### 11.4 构建与运行时

构建：`manualChunks`、Ant Design 按需、图片 lazy。运行：React DevTools Profiler、Chrome Performance。

---

## 12. React 进阶 vs Vue 12 专题对照

| Vue 12 主题 | React 对应 | 差异要点 |
|-------------|------------|----------|
| 插槽 slot | children / render props | React 用函数更灵活 |
| KeepAlive | 无内置，需组合方案 | Vue 更便捷 |
| Teleport | Portals | 概念一致 |
| 自定义指令 | 无；用组件或 Hook | `useEffect` 绑事件 |
| provide/inject | Context API | Context 变更触发渲染 |
| defineAsyncComponent | React.lazy | 都配合 Suspense |
| shallowRef | 无响应式系统 | 靠 memo/useMemo 优化 |
| v-memo | React.memo | 手动优化 |
| Transition | `react-transition-group` / CSS | 需额外库 |
| Vue 2/3 对比 | React 18 并发 | 不同演进路线 |

---

## 13. shop-react 进阶改造建议

| 改造 | 章节技术 | 优先级 |
|------|----------|--------|
| ProductCard memo + useCallback | §3、§4 | 高 |
| 路由 lazy + Suspense | §7 | 高 |
| 商品大图 Portal 预览 | §8 | 中 |
| 商品模块 ErrorBoundary | §6 | 中 |
| ThemeContext 暗色模式 | §1 | 低 |
| 搜索 useTransition | §9 | 低 |

---

## 14. 常见报错

| 报错 / 现象 | 原因 | 解决 |
|-------------|------|------|
| `memo` 无效 | 内联函数 props | useCallback |
| Context 全树刷新 | value 每次新对象 | useMemo 稳定 value |
| lazy 白屏 | 未包 Suspense | 外层加 fallback |
| Portal 事件不触发 | 误解 DOM 冒泡 | 事件仍走 React 树 |
| Hook 顺序报错 | 条件里调 Hook | 提到顶层 |
| useEffect 死循环 | 依赖引用每次变 | 稳定依赖或抽离 |

---

## 15. 练习建议

### 基础

1. 实现 `ThemeContext`：亮/暗切换，ProductCard 读取主题色。  
2. 给 `ProductCard` 加 `memo`，父组件加搜索框，用 Profiler 对比优化前后 render 次数。

### 进阶

3. `React.lazy` 拆分 ProductDetail、Cart 路由，自定义 `PageLoading`。  
4. 商品图片点击用 `createPortal` 全屏预览，点击遮罩关闭。

### 挑战

5. 实现 `ErrorBoundary` + 模拟子组件 throw，展示降级页与重试。  
6. 大列表（1000 条）搜索：对比普通 `setState` 与 `useTransition` 输入流畅度。  
7. 封装 `useDebounce(value, delay)` Hook，用于商品搜索请求。

---

## 16. 学完标准

- [ ] 会创建 Context Provider 并在深层 `useContext` 消费  
- [ ] 能解释 `memo`、`useMemo`、`useCallback` 各自解决什么问题  
- [ ] 会写 `React.lazy` + `Suspense` 路由懒加载  
- [ ] 会用 `createPortal` 实现 Modal 挂到 body  
- [ ] 能实现 class 版 Error Boundary 并知道其局限  
- [ ] 能口述 `useTransition` / `useDeferredValue` 使用场景  
- [ ] 能对比 React 与 Vue 12 章节中 KeepAlive、插槽、Teleport 差异  
- [ ] 能列举 5 条以上 React 性能优化手段  

---

## 17. FAQ

**Q：Context 能替代 Zustand 吗？** 少变配置可以；购物车等频繁更新用 Zustand。  
**Q：每个组件都 memo 好吗？** 不好，先 Profiler 再优化。  
**Q：Error Boundary 能捕接口错误吗？** 不能，用 `catch` + 错误状态展示。

---

## 下一章预告

进阶特性补完了——下一章（13 高频场景题与面试专题）专门练 **面试官怎么问、你怎么答**：25+ 题带完整回答框架，每题尽量结合 shop-react。

---

*下一章：13 高频场景题与面试专题*
