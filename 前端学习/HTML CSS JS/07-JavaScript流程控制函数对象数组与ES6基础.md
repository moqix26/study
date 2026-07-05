# JavaScript 流程控制、函数、对象、数组与 ES6 基础

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0 导读、购物车逐行读、DevTools、FAQ 12 题、闭卷自测、费曼检验；链 Vue 08 / 计网 04 -->

## 0. 读前导读（零基础也能跟上）

> **读者假设**：你已学完 [06-JS 基础语法](./06-JavaScript基础语法与数据类型.md)，能在 Console 里跑 `map`/`filter`。本章把「会写几行 JS」升级为「能组织真实业务逻辑」——Vue 的 composable、Axios 数据处理都建立在本章能力之上。

### 0.1 用一句话弄懂本章

**一句话**：用**函数**封装逻辑、用**数组高阶方法**处理列表、用 **ES6 解构/展开/可选链** 写更少更安全的代码——把商品列表变成购物车总价，就是本章要练的手艺。

**生活类比**：

| 本章概念 | 类比 | 后面章节对应 |
|---------|------|-------------|
| **函数 + 默认参数** | 菜谱「盐少许，默认 3g」 | Vue 组件 props 默认值 |
| **map/filter/reduce 链** | 流水线：筛选→改包装→汇总 | Vue `computed` 过滤商品 |
| **解构 / 展开** | 拆快递、合并两箱货 | [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 拆 API 响应 |
| **JSON.stringify/parse** | 前后端「标准装箱单」 | 计网 [04 HTTP](../计算机网络/04-HTTP协议深入.md) 的 JSON body |
| **ES Module** | 把函数分到不同 `.js` 文件 | Vite 项目 `import` 组件 |

**为什么重要**：不会本章的 `reduce` 和对象解构，读 [Vue 08 Axios 联调](../Vue/08-Axios网络请求与前后端联调.md) 时看到 `const { data } = await axios.get(...)` 会懵；后端 [Java 04](../../后端学习/Java/04-SpringBoot核心开发.md) 返回的 JSON 数组，要靠本章方法清洗后再渲染 DOM（[08 章](./08-JavaScript-DOM-BOM与事件机制.md)）或交给 Vue。

---

### 0.2 你需要提前知道什么

| 水平 | 建议 |
|------|------|
| 完全零基础 | 先完成 [06 章](./06-JavaScript基础语法与数据类型.md) §0～§13，至少能在 Console 跑通 `filter` |
| 已会 06 章 | 从 §3 函数参数开始；§5 箭头函数 this 是重点 |
| 目标 Vue 路线 | 本章 + [08 DOM](./08-JavaScript-DOM-BOM与事件机制.md) → 可并行开 Vue 06；联调前必过 [09 异步](./09-JavaScript异步编程网络请求与本地存储.md) + [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) |
| 想搞懂 HTTP/JSON | 本章 §12 JSON + [10 浏览器网络](./10-浏览器HTTP网络与Web基础.md)；深入见 [计网 04](../计算机网络/04-HTTP协议深入.md) |

---

### 0.3 本章知识地图（☐→☑）

- [ ] 会写默认参数、剩余参数、参数解构
- [ ] 能说明箭头函数 vs 普通函数在 `this` 上的差异
- [ ] 熟练 `map/filter/reduce` 链式调用（至少 3 步链）
- [ ] 会用对象/数组解构、展开、`?.`、`??`
- [ ] 会用 `Set` 去重、`Map` 做词频/缓存
- [ ] 会 `JSON.stringify/parse` 与 `safeParse`
- [ ] 完成 §14 购物车 Console 实战
- [ ] 闭卷自测 ≥ 8/10

---

### 0.4 建议学习时长与节奏

| 阶段 | 时间 | 验收 |
|------|------|------|
| 函数 §3～§5 | 2.5 小时 | 手写 `range(1,5)` 和箭头/普通 this 对比 |
| 对象/Set/Map §6～§7 | 2 小时 | `Map` 统计词频 |
| 数组管道 §8 | 2 小时 | 三步链：filter→map→reduce |
| ES6 语法 §9～§11 | 1.5 小时 | 解构 API 假数据 `{ user: { name } }` |
| 购物车实战 §14 | 1.5 小时 | Console 输出正确总价与优惠 |
| 自测 §20 | 0.5 小时 | ≥ 8/10 |

---

### 0.5 学完本章你能做什么（可验证）

1. 给定 `[{name,price,category}]`，用链式调用算出「数码类商品折后总价」（满 500 九折）。
2. 把 Java 04 风格的 JSON 字符串 `parse` 成对象，用解构取出 `data.list[0].name`。
3. 向朋友解释：为什么对象方法不要用箭头函数写 `sayHi: () => {}`。
4. 为 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 做好准备：能独立写 `async function loadProducts()` 里的数据处理（不含 DOM）。

---

### 0.6 核心术语三件套

**术语（解构 Destructuring）**：从对象或数组里「按名字/位置」一次性取出多个值赋给变量。
**生活类比**：快递单上写「收件人：小明，电话：138…」——解构就是一次撕下姓名和电话，不用拆两次。
**为什么重要**：Axios 响应常写 `const { data } = await axios.get(...)`；Vue 组合式里 `const { count, increment } = useCounter()` 同原理。
**本章用到的地方**：§9 解构；[Vue 08 §4](../Vue/08-Axios网络请求与前后端联调.md) 响应处理。

**术语（reduce 归约）**：把数组「折叠」成一个值（数字、对象、Map 等）；必须提供初始值避免空数组 bug。
**生活类比**：购物车小票——从第一件商品一直加到最后，得到总价；也可按类别分组成小票堆。
**为什么重要**：分组统计、管道汇总、Vue 里算购物车总价本质都是 reduce 思维。
**本章用到的地方**：§8 数组管道；§14 购物车。

**术语（ES Module 模块）**：用 `export`/`import` 把函数拆到独立文件，浏览器通过 `type="module"` 加载。
**生活类比**：把菜谱按「冷菜/热菜」分到不同笔记本，做菜时按需翻开——不是一整本粘在一起。
**为什么重要**：Vite + Vue 项目 100% 用 ESM；与 CommonJS `require`（Node 老写法）不同。
**本章用到的地方**：§15 后模块化小节；Vue 01 `main.js` 预习。

---

## 1. 这一份文档的定位

上一份讲了 JavaScript 最基础的语法。

这一份继续深入，把你最常用的 JavaScript 核心能力补厚：

- 流程控制深入（循环选择、性能对比）
- 函数深入（箭头函数 vs 普通函数、this 场景）
- 对象深入（方法简写、计算属性、遍历模式）
- 数组深入（高阶方法链式调用、reduce 模式）
- ES6+ 常用语法（解构、展开、Set/Map、可选链）
- 模块化入门
- 综合实战：购物车、数据处理管道

学完这份后，加上 06 篇，你应该能独立处理大部分 JS 逻辑问题。

---

## 2. 函数为什么这么重要

JavaScript 不是只靠一堆语句堆起来的。真正可维护的代码要靠函数来：

- 复用逻辑（写一次，到处用）
- 拆分职责（一个函数只做一件事）
- 提高可读性（好函数名 = 自带注释）

### 函数的核心心态

不要写成这样：

```js
// ❌ 意大利面条 — 所有逻辑堆在主流程
const cart = [...];
let total = 0;
for (...) { total += ... ; }
if (total > 500) { total *= 0.9; }
console.log("总计：" + total);
```

应该拆成：

```js
// ✅ 关注点分离
function calcTotal(items) { ... }
function applyDiscount(total) { ... }
function formatPrice(price) { ... }

const total = calcTotal(cart);
const final = applyDiscount(total);
console.log(formatPrice(final));
```

---

## 3. 函数参数（深入）

### 3.1 基本参数

```js
function greet(name) {
  console.log("你好，" + name);
}
greet("小明"); // "你好，小明"
greet();       // "你好，undefined" — 没传参数就是 undefined
```

### 3.2 默认参数

```js
function greet(name = "游客") {
  console.log(`你好，${name}`);
}
greet();       // "你好，游客"
greet("小明"); // "你好，小明"

// 默认值可以使用前面的参数
function createPrice(price, tax = price * 0.13) {
  return price + tax;
}
console.log(createPrice(100)); // 113
```

### 3.3 剩余参数 `...`

```js
function sum(...nums) {
  return nums.reduce((total, n) => total + n, 0);
}
console.log(sum(1, 2, 3, 4, 5)); // 15

// 前几个参数单独命名，剩余打包
function log(prefix, ...messages) {
  for (const msg of messages) {
    console.log(`[${prefix}] ${msg}`);
  }
}
log("INFO", "启动", "连接中", "完成");
// [INFO] 启动
// [INFO] 连接中
// [INFO] 完成
```

### 3.4 参数解构

```js
// 传对象参数（常见于配置型函数）
function createCard({ title, desc = "", color = "blue" }) {
  console.log(`创建${color}卡片：${title} - ${desc}`);
}

createCard({ title: "新品", color: "red" });
// 创建red卡片：新品 -

// 与默认空对象配合，可安全不传参
function safeCreate({ title = "默认" } = {}) {
  console.log(title);
}
safeCreate(); // "默认"
```

---

## 4. return 返回值（深入）

```js
function square(num) {
  return num * num;
}

// 没有 return → 返回 undefined
function noReturn() {
  const x = 1;
}
console.log(noReturn()); // undefined

// return 后面不跟值也返回 undefined
function earlyReturn(condition) {
  if (!condition) return;  // undefined
  return "继续";
}

// 函数执行到 return 就结束
function demo() {
  console.log("这行会执行");
  return;
  console.log("这行永远不会执行"); // 不可达代码
}
```

### 返回值的灵活运用

```js
// 返回对象
function createUser(name, age) {
  return { name, age, createdAt: new Date() };
}

// 返回数组
function minMax(arr) {
  return [Math.min(...arr), Math.max(...arr)];
}
const [min, max] = minMax([3, 1, 4, 1, 5]); // min=1, max=5

// 返回函数（高阶函数）
function multiply(factor) {
  return (n) => n * factor;
}
const double = multiply(2);
console.log(double(5)); // 10
```

---

## 5. 箭头函数 vs 普通函数（完整对比）

这是面试必问题，也是实际开发中做选择的依据。

### 5.1 写法对比

```js
// 普通函数声明
function add(a, b) {
  return a + b;
}

// 函数表达式
const add2 = function(a, b) {
  return a + b;
};

// 箭头函数
const add3 = (a, b) => a + b;

// 只有一个参数的简写
const square = n => n * n;

// 没有参数
const sayHi = () => "Hi";

// 返回对象 — 必须包 ()
const createUser = (name, age) => ({ name, age });

// 多行体
const calc = (a, b) => {
  const result = a + b;
  return result;
};
```

### 5.2 核心区别对比表

| 对比维度 | 普通函数 | 箭头函数 |
|----------|---------|----------|
| `this` 绑定 | 动态（调用时决定） | 词法（定义时捕获外层 this） |
| `arguments` 对象 | ✅ 有 | ❌ 没有（用剩余参数代替） |
| `new` 调用 | ✅ 可以作为构造函数 | ❌ 不能用 new |
| `prototype` 属性 | ✅ 有 | ❌ 没有 |
| 方法简写 | `obj.fn = function(){}` | 不适合做对象方法（this 指向问题） |
| 回调函数 | `arr.map(function(x){})` | `arr.map(x => x * 2)`（首选） |

### 5.3 `this` 行为差异（关键！）

```js
// 普通函数：this 看调用方式
const user1 = {
  name: "小明",
  sayHi: function() {
    console.log(this.name);
  }
};
user1.sayHi(); // "小明" — this 指向 user1

// 将方法赋值给变量 — this 指向丢失！
const fn = user1.sayHi;
// fn(); // undefined（非严格模式下 this 是 window）

// 箭头函数：this 是定义时外层上下文
const user2 = {
  name: "小红",
  sayHi: () => {
    console.log(this.name);
  }
};
// user2.sayHi(); // undefined！箭头函数 this 不是 user2

// 定时器中的对比
const user3 = {
  name: "小刚",
  // ✅ 箭头函数：this 是外层（user3）的 this
  greetLater() {
    setTimeout(() => {
      console.log(`你好，${this.name}`); // "你好，小刚"
    }, 100);
  },
  // ❌ 普通函数：this 丢失
  greetLaterBad() {
    setTimeout(function() {
      console.log(this.name); // undefined
    }, 100);
  }
};
user3.greetLater();
```

**记忆口诀**：
- 对象方法、需要动态 this → 普通函数（或方法简写）
- 回调、数组方法、定时器 → 箭头函数
- 不写 `new` 的场合，优先考虑箭头函数（简洁 + 安全）

---

## 6. 对象再深入

### 6.1 属性简写与方法简写

```js
const name = "小明";
const age = 18;

// ES6 属性简写
const user = {
  name,     // 等同于 name: name
  age,      // 等同于 age: age

  // ES6 方法简写
  sayHi() {
    console.log(`你好，我是${this.name}`);
  },

  // 老式写法
  // sayHi: function() { ... }
};

console.log(user); // { name: "小明", age: 18, sayHi: f }
```

### 6.2 计算属性名

```js
const key = "favoriteColor";
const method = "getData";

const obj = {
  [key]: "blue",         // favoriteColor: "blue"
  [method]() {           // getData() { ... }
    return "data";
  },
  [`prefix_${key}`]: "value", // prefix_favoriteColor: "value"
};

console.log(obj.favoriteColor); // "blue"
console.log(obj.getData());     // "data"
```

### 6.3 对象方法中的 `this`

```js
const product = {
  name: "键盘",
  price: 199,
  show() {
    console.log(`${this.name}：¥${this.price}`);
  },
  discount(rate) {
    this.price = this.price * (1 - rate);
    return this; // 返回 this 支持链式调用
  }
};

product.show();      // "键盘：¥199"
product.discount(0.1);
product.show();      // "键盘：¥179.1"
```

### 6.4 对象合并与拷贝

```js
const defaults = { theme: "light", lang: "zh" };
const userPrefs = { theme: "dark" };

// 合并：后者覆盖前者
const settings = { ...defaults, ...userPrefs };
// { theme: "dark", lang: "zh" }

// Object.assign（原地修改第一个参数）
Object.assign(defaults, userPrefs);
// defaults 被修改了！为避免：Object.assign({}, defaults, userPrefs)
```

---

## 7. Set 和 Map（ES6 新数据结构）

### 7.1 Set — 值的集合（自动去重）

```js
// 创建
const set = new Set([1, 2, 3, 2, 1]); // Set { 1, 2, 3 } — 自动去重！

// 常用方法
set.add(4);          // 添加
set.delete(2);       // 删除
set.has(1);          // 是否有 → true
set.size;            // 元素数量 → 3
set.clear();         // 清空

// 遍历
for (const item of set) {
  console.log(item);
}

// 实战用法
// 1. 数组去重（一行搞定）
const arr = [1, 2, 2, 3, 3, 3];
const unique = [...new Set(arr)]; // [1, 2, 3]

// 2. 判断是否存在
const visited = new Set();
visited.add("page1");
if (!visited.has("page2")) {
  console.log("page2 没被访问过");
}
```

### 7.2 Map — 键值对集合（键可以是任意类型）

```js
// 创建
const map = new Map();

// 增删改查
map.set("name", "小明");
map.set(1, "数字键");
map.set({ id: 1 }, "对象键"); // ← 对象也可以做键！
map.set(true, "布尔键");

map.get("name");      // "小明"
map.has("name");      // true
map.size;             // 4
map.delete(1);        // 删除
map.clear();          // 清空

// 遍历
for (const [key, value] of map) {
  console.log(key, value);
}

// 实战用法
// 1. 缓存计算结果
const cache = new Map();
function expensiveCalc(n) {
  if (cache.has(n)) return cache.get(n);
  const result = n * n; // 假装很昂贵
  cache.set(n, result);
  return result;
}

// 2. 记录元素出现次数
const words = ["a", "b", "a", "c", "a"];
const count = new Map();
for (const w of words) {
  count.set(w, (count.get(w) || 0) + 1);
}
// Map { "a" => 3, "b" => 1, "c" => 1 }
```

### 7.3 Object vs Map 选择指南

| | Object | Map |
|---|--------|-----|
| 键的类型 | 只能是字符串/Symbol | 任意类型（对象、函数等） |
| 遍历顺序 | 有顺序但不保证稳定 | 严格按插入顺序 |
| 获取大小 | `Object.keys(obj).length` | `map.size`（O(1)） |
| 性能（频繁增删） | 一般 | 更快（针对频繁增删优化） |
| 默认原型键 | 有（toString 等可能冲突） | 无 |
| 何时用 | 固定结构、配置对象 | 动态键、需要迭代顺序、键非字符串 |

---

## 8. Array 高阶方法链式调用（数据管道）

### 8.1 链式调用模式

```js
const products = [
  { name: "键盘", price: 199, category: "数码", stock: 10 },
  { name: "鼠标", price: 89, category: "数码", stock: 0 },
  { name: "书架", price: 299, category: "家居", stock: 5 },
  { name: "台灯", price: 159, category: "家居", stock: 8 },
  { name: "耳机", price: 399, category: "数码", stock: 3 },
];

// 需求：数码类、有货、按价格排序、只要名称和价格
const result = products
  .filter(p => p.category === "数码")   // 1. 筛选
  .filter(p => p.stock > 0)            // 2. 有货
  .sort((a, b) => a.price - b.price)   // 3. 排序
  .map(p => `${p.name}：¥${p.price}`); // 4. 格式化

console.log(result);
// ["鼠标：¥89", "键盘：¥199", "耳机：¥399"]
```

#### 8.1.1 链式管道逐行读

| 行号/代码 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `.filter(p => p.category === "数码")` | 第一步缩小集合，返回新数组 | 条件写错 → 结果空 |
| 连续两个 `.filter` | 可读性优于一个复杂 && | 合并成一个 filter 也行 |
| `.sort((a,b)=>a.price-b.price)` | 原地排序当前链上数组 | 忘比较函数 → 字符串序 |
| `.map(p => \`${p.name}：¥${p.price}\`)` | 最后映射为展示字符串 | 箭头函数 `{}` 无 return → undefined |
| 整条链起点是 `products` | 原数组不被 map/filter 改变 | sort 会改中间数组副本的序 |

此管道与 [Vue 02 computed](../Vue/02-模板语法与响应式原理.md) 过滤商品列表同构；数据源来自 API 时见 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md)，HTTP 传输见 [计网 04](../计算机网络/04-HTTP协议深入.md)。

### 8.2 reduce 进阶模式

```js
// 模式 1：分组（group by）
const byCategory = products.reduce((groups, p) => {
  const key = p.category;
  if (!groups[key]) groups[key] = [];
  groups[key].push(p);
  return groups;
}, {});
// { 数码: [...], 家居: [...] }

// 模式 2：转 Map
const productMap = products.reduce((map, p) => {
  map.set(p.name, p);
  return map;
}, new Map());

// 模式 3：多维统计
const stats = products.reduce((s, p) => ({
  totalPrice: s.totalPrice + p.price * p.stock,
  totalStock: s.totalStock + p.stock,
  count: s.count + 1,
}), { totalPrice: 0, totalStock: 0, count: 0 });
// { totalPrice: ..., totalStock: 26, count: 5 }

// 模式 4：管道（组合多个转换）
const pipe = (...fns) => (x) => fns.reduce((v, fn) => fn(v), x);
const double = n => n * 2;
const addOne = n => n + 1;
const process = pipe(double, addOne, double);
console.log(process(3)); // double(3)=6 → add(6)=7 → double(7)=14
```

---

## 9. 解构进阶（函数参数、嵌套、默认值）

```js
// 1. 跳过元素
const [first, , third] = [1, 2, 3];
// first=1, third=3

// 2. 剩余元素
const [head, ...tail] = [1, 2, 3, 4];
// head=1, tail=[2,3,4]

// 3. 嵌套解构
const user = {
  name: "小明",
  address: { city: "北京", district: "海淀" }
};
const { address: { city, district } } = user;
// city="北京", district="海淀"

// 4. 解构 + 重命名
const { name: userName, age: userAge = 18 } = { name: "Tom" };
// userName="Tom", userAge=18

// 5. 交换变量
let a = 1, b = 2;
[a, b] = [b, a]; // a=2, b=1

// 6. 函数参数解构（React/Vue 超常见）
function UserCard({ name, age, city = "未知" }) {
  console.log(`${name}, ${age}, ${city}`);
}
UserCard({ name: "小明", age: 20 }); // "小明, 20, 未知"
```

---

## 10. 可选链 `?.` 与空值合并 `??` 深入

```js
// 可选链：安全访问深层属性
const user = { profile: null };

// ❌ 报错
// user.profile.nickname;

// ✅ 安全——遇到 null/undefined 就返回 undefined 而不是报错
console.log(user?.profile?.nickname);  // undefined
console.log(user?.profile?.nickname ?? "匿名用户"); // "匿名用户"

// 可选链也可用于函数调用
const obj = {};
// obj.method(); // ❌ 报错
obj.method?.();   // ✅ 不报错，返回 undefined

// 可选链用于数组索引
const arr = null;
console.log(arr?.[0]); // undefined（不报错）

// ?? vs || 的区别（重要！）
console.log(0 || "默认");    // "默认" — 0 是假值
console.log(0 ?? "默认");    // 0 — 0 不是 null/undefined
console.log("" || "默认");  // "默认"
console.log("" ?? "默认");  // "" — 空字符串不是 null/undefined
console.log(null ?? "默认");// "默认"
```

---

## 11. Date 基础操作

```js
// 获取当前时间
const now = new Date();
console.log(now); // 当前完整时间

// 创建指定时间
new Date("2026-06-18");            // ISO 字符串
new Date(2026, 5, 18);             // 年, 月(0-11), 日
new Date(2026, 5, 18, 14, 30, 0); // 精确到时/分/秒

// 获取各部分
now.getFullYear();   // 年
now.getMonth();      // 月（0-11！注意加 1）
now.getDate();       // 日
now.getDay();        // 星期（0=周日）
now.getHours();      // 时
now.getMinutes();    // 分
now.getSeconds();    // 秒

// 时间戳（毫秒）
now.getTime();       // 距 1970-1-1 的毫秒数
Date.now();          // 当前时间戳（静态方法）

// 格式化
now.toLocaleString("zh-CN");              // "2026/6/18 14:30:00"
now.toLocaleDateString("zh-CN");          // "2026/6/18"
now.toISOString();                        // "2026-06-18T06:30:00.000Z"（UTC）

// 日期运算
const tomorrow = new Date(now);
tomorrow.setDate(tomorrow.getDate() + 1); // 加一天

const diff = tomorrow - now;               // 毫秒差
console.log(diff / 1000 / 60 / 60);       // 转为小时
```

---

## 12. JSON 深入

```js
// 对象/数组 → JSON 字符串
const data = { name: "小明", scores: [90, 85] };
const jsonStr = JSON.stringify(data);
console.log(jsonStr); // '{"name":"小明","scores":[90,85]}'

// 美化缩进
JSON.stringify(data, null, 2);

// JSON 字符串 → 对象
const parsed = JSON.parse('{"name":"小明","age":18}');

// JSON 不能表示的内容
const bad = {
  fn: function() {},   // ❌ 会被忽略
  undef: undefined,    // ❌ 会被忽略
  date: new Date(),    // ⚠️ 转为 ISO 字符串
  nan: NaN,            // ⚠️ 转为 null
  infinity: Infinity,  // ⚠️ 转为 null
};
console.log(JSON.stringify(bad)); // {"date":"...","nan":null,"infinity":null}

// 安全的 JSON.parse（防止解析失败）
function safeParse(str, fallback = null) {
  try {
    return JSON.parse(str);
  } catch {
    return fallback;
  }
}
```

---

## 13. 异常处理进阶

```js
// 基础
try {
  const result = riskyOperation();
} catch (error) {
  console.error("操作失败:", error.message);
} finally {
  console.log("无论成功失败，我都会执行"); // 清理逻辑放这里
}

// 自定义抛出
function divide(a, b) {
  if (b === 0) throw new Error("除数不能为 0");
  if (typeof a !== "number" || typeof b !== "number") {
    throw new TypeError("参数必须是数字");
  }
  return a / b;
}

// 安全的函数包装
function safeRun(fn, fallback) {
  try {
    return fn();
  } catch (err) {
    console.warn("执行出错:", err.message);
    return fallback;
  }
}
```

---

## 14. 综合实战：购物车完整系统

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <title>购物车系统（Console 练习）</title>
</head>
<body>
  <h1>打开控制台看完整输出</h1>
  <script>
    console.log("========== 购物车系统 ==========\n");

    // ===== 数据 =====
    const products = [
      { id: 1, name: "机械键盘", price: 299, category: "数码" },
      { id: 2, name: "无线鼠标", price: 89, category: "数码" },
      { id: 3, name: "显示器支架", price: 159, category: "办公" },
      { id: 4, name: "笔记本散热架", price: 49, category: "办公" },
      { id: 5, name: "Type-C 数据线", price: 29, category: "配件" },
    ];

    const cart = [
      { productId: 1, quantity: 1 },
      { productId: 2, quantity: 2 },
      { productId: 5, quantity: 3 },
    ];

    const COUPONS = {
      "SAVE10": { discount: 10, minTotal: 100 },
      "SAVE30": { discount: 30, minTotal: 200 },
      "VIP50": { discount: 50, minTotal: 300 },
    };

    // ===== 工具函数 =====
    const formatPrice = (n) => `¥${n.toFixed(2)}`;

    function getCartItems(cart, products) {
      return cart.map(item => {
        const product = products.find(p => p.id === item.productId);
        return product
          ? { ...product, quantity: item.quantity, subtotal: product.price * item.quantity }
          : null;
      }).filter(Boolean);
    }

    function calcTotal(items) {
      return items.reduce((sum, item) => sum + item.subtotal, 0);
    }

    function getBestCoupon(total, coupons) {
      const available = Object.entries(coupons)
        .filter(([_, c]) => total >= c.minTotal)
        .sort((a, b) => b[1].discount - a[1].discount);
      return available.length > 0 ? available[0] : null;
    }

    // ===== 执行 =====
    const items = getCartItems(cart, products);
    const total = calcTotal(items);

    console.log("🛒 购物清单：");
    console.table(items.map(i => ({
      商品: i.name,
      单价: formatPrice(i.price),
      数量: i.quantity,
      小计: formatPrice(i.subtotal),
    })));

    console.log(`\n💰 小计: ${formatPrice(total)}`);

    // 优惠券
    const best = getBestCoupon(total, COUPONS);
    if (best) {
      const [code, { discount }] = best;
      const final = total - discount;
      console.log(`🎫 已用券: ${code}（减${discount}元）`);
      console.log(`✅ 实付: ${formatPrice(final)}`);
    } else {
      console.log("📭 无可用优惠券");
      console.log(`✅ 实付: ${formatPrice(total)}`);
    }

    // 分类统计
    const byCategory = items.reduce((acc, i) => {
      acc[i.category] = (acc[i.category] || 0) + i.subtotal;
      return acc;
    }, {});
    console.log("\n📊 分类消费：");
    for (const [cat, amount] of Object.entries(byCategory)) {
      console.log(`  ${cat}: ${formatPrice(amount)}`);
    }

    console.log("\n✅ 购物车系统运行完毕！");
  </script>
</body>
</html>
```

### 14.1 购物车核心逻辑逐行读

| 行号/代码 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `getCartItems` 里 `find` | 按 `productId` 找商品详情 | 写错 id 字段 → 整行变 `null` 被 filter 掉 |
| `{ ...product, quantity, subtotal }` | 展开商品字段并追加数量与小计 | 不用展开只改引用 → 污染原 `products` |
| `filter(Boolean)` | 去掉 `find` 失败产生的 `null` | 忘记 filter → 后面 reduce 遇到 null 报错 |
| `calcTotal` 初始值 `0` | reduce 必须从数字起算 | 省略初始值且 cart 空 → 报错 |
| `Object.entries(coupons)` | 把券对象变成 `[code, rule]` 数组 | 直接 for...in 也能做，但 entries 更清晰 |
| `sort((a,b)=>b[1].discount-a[1].discount)` | 选折扣最大的券 | 忘记 sort → 可能用到小额券 |
| `const [code, { discount }] = best` | 数组+对象解构一次取出 | 不解构也能写，但可读性差 |
| `byCategory` reduce | 按分类累加金额 | 初始 `{}` 写成 `[]` → 逻辑全错 |

### 14.2 与后端 JSON 的衔接（预习 Vue 08 / 计网 04）

Java [04 Spring Boot](../../后端学习/Java/04-SpringBoot核心开发.md) 典型响应：

```json
{ "code": 200, "data": { "list": [{ "id": 1, "name": "键盘", "price": 299 }] } }
```

本章练法（Console）：

```js
const raw = '{"code":200,"data":{"list":[{"id":1,"name":"键盘","price":299}]}}';
const { data: { list } } = JSON.parse(raw);
const names = list.map(p => p.name);
```

HTTP 里 JSON 如何传输、Content-Type 是什么，见 [计网 04](../计算机网络/04-HTTP协议深入.md)；在 Vue 里用 Axios 封装，见 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md)。

---

## 15. 初学者常见错误（扩展版）

### 15.1 `map` 没 return

```js
// ❌ 箭头函数用了 {} 但没写 return
nums.map((n) => { n * 2 });         // [undefined, undefined, ...]

// ✅
nums.map((n) => n * 2);              // 简写，自动 return
nums.map((n) => { return n * 2; });  // 显式 return
```

### 15.2 解构 undefined

```js
const { x } = undefined; // ❌ 报错
const { x } = undefined ?? {}; // ✅ 安全
const { x } = undefined || {}; // ✅ 也行（但 ?? 更精确）
```

### 15.3 `??` 和 `||` 混淆

```js
const count = 0;
console.log(count || 10);   // 10 — 0 是假值！
console.log(count ?? 10);   // 0 — 只有 null/undefined 才用备用值
```

### 15.4 `sort` 默认是字符串排序

```js
[10, 2, 1, 20].sort();        // [1, 10, 2, 20] — 不是期望的数字序！
[10, 2, 1, 20].sort((a,b) => a - b); // [1, 2, 10, 20] — 数字升序
```

### 15.5 展开运算符是浅拷贝

```js
const original = { a: 1, inner: { b: 2 } };
const copy = { ...original };
copy.inner.b = 99;
console.log(original.inner.b); // 99 — 被改了！浅拷贝只复制第一层
```

### 15.6 箭头函数做对象方法时 this 错误

```js
const obj = {
  name: "test",
  fn: () => console.log(this.name), // this 不是 obj！
};
obj.fn(); // undefined
```

### 15.7 ES Module 拆分（预习 Vue / Vite）

把 §14 工具函数拆到 `utils/cart.js`：

```js
export const formatPrice = (n) => `¥${n.toFixed(2)}`;
export function calcTotal(items) {
  return items.reduce((sum, item) => sum + item.subtotal, 0);
}
```

HTML 引用：

```html
<script type="module">
  import { formatPrice, calcTotal } from "./utils/cart.js";
  console.log(formatPrice(calcTotal([{ subtotal: 100 }])));
</script>
```

| 行号/字段 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `export` | 模块对外接口 | 无 export → import 失败 |
| `import { ... } from './utils/cart.js'` | 命名导入；路径含 `.js` | 路径错 → 404 |
| `type="module"` | 启用 ESM，默认 defer | 普通 script 不能用 import |

[Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 的 `import axios from 'axios'` 同体系；模块里请求 API 的 HTTP 规则见 [计网 04](../计算机网络/04-HTTP协议深入.md)。

---

## 16. 分级练习

### 基础（必做）
1. 用 `filter` 筛出 ≥60 分的成绩，用 `map` 转成 `"姓名：分数"` 格式
2. 用 `reduce` 求数组 `[1,2,3,4,5]` 的乘积
3. 用 `Set` 对数组去重，用 `Map` 统计元素出现次数
4. 实现函数 `range(start, end)` 返回 `[start, ..., end]` 的数组

### 进阶
1. 完成第 14 节购物车系统，加"最贵商品查找"和"总商品数统计"
2. 用 `reduce` 实现 `groupBy(arr, key)` 通用分组函数
3. 用 `Date` 实现 `getDaysBetween(date1, date2)` 计算两个日期的间隔天数

### 挑战
1. 用 `reduce` + Map 实现 `countBy(arr)` 统计词频（返回 Map）
2. 实现函数 `chunk(arr, size)` 将数组分块（如 `[1,2,3,4,5]`, 2 → `[[1,2],[3,4],[5]]`）
3. 用箭头函数 + 链式调用完成完整数据管道（过滤 → 排序 → 分组 → 格式化）

---

## 17. FAQ

**Q1：`sort` 为什么结果不对？**  
默认按字符串 Unicode 排。`[10, 2, 1].sort()` → `[1, 10, 2]`。数字排序必须 `.sort((a,b) => a - b)`。

**Q2：`forEach` 和 `map` 什么时候用？**  
需要**新数组**用 `map`；只做副作用（打印、改 DOM）用 `forEach`。`map` 返回值常被忽略是常见 bug。

**Q3：展开运算符会深拷贝吗？**  
只拷贝一层。嵌套对象仍是引用；深拷贝用 `structuredClone(obj)`（现代浏览器）或 JSON 法（丢函数/Date）。

**Q4：Set/Map 什么时候用？**  
数组去重用 `Set`；任意类型键、词频、LRU 缓存用 `Map`。普通「配置对象」仍用 Object。

**Q5：可选链能连用吗？**  
能。`obj?.a?.b?.c`；配合 `??` 设默认值：`obj?.price ?? 0`。

**Q6：箭头函数能当构造函数吗？**  
不能，没有 `prototype`，`new` 会报错。类组件、需要 `new` 的场景用普通函数或 class。

**Q7：`reduce` 不加初始值会怎样？**  
空数组直接报错；非空则把第一项当初始值——类型不对时难排查。**永远写初始值**。

**Q8：JSON.parse 失败怎么防？**  
用 §12 的 `safeParse(str, fallback)`；localStorage 脏数据尤其需要。

**Q9：解构能设默认值吗？**  
能。`const { name = "游客", age = 0 } = user`；数组 `[a=1, b=2] = arr`。

**Q10：和 Vue 02 的 computed 什么关系？**  
`computed(() => products.filter(...))` 就是响应式版 filter；本章在纯 JS 里练熟，进 Vue 只多一层「自动重算」。

**Q11：ES Module 和 `<script>` 普通脚本区别？**  
`type="module"` 自动严格模式、顶层 `import/export`、默认 defer；变量不污染全局。

**Q12：本章算学完能直接联调后端吗？**  
数据处理够了；还要 [09 异步 fetch](./09-JavaScript异步编程网络请求与本地存储.md) + [10 网络/CORS](./10-浏览器HTTP网络与Web基础.md) + [Vue 08 Axios](../Vue/08-Axios网络请求与前后端联调.md)。

---

## 18. 学完标准

- [ ] 能熟练写函数声明、表达式、箭头函数，知道各自的 this 行为
- [ ] 会用默认参数、剩余参数、参数解构
- [ ] 掌握 `map/filter/find/reduce/some/every`，能链式调用
- [ ] 会用 `reduce` 实现分组、统计、管道等模式
- [ ] 会对象/数组解构、展开运算符、可选链、空值合并
- [ ] 了解 Set（去重）和 Map（键值对）的使用场景
- [ ] 会用 Date 处理日期，JSON 做序列化/反序列化
- [ ] 了解 ES Module 的 `import`/`export` 基本写法
- [ ] 能独立完成购物车系统等综合实战

---

## 19. 本章思维导图

```
JS 核心能力全景
│
├── 函数
│   ├── 声明 / 表达式 / 箭头（三种形式）
│   ├── 默认参数、剩余参数、参数解构
│   ├── return 规则与返回值模式
│   └── 箭头 vs 普通：this / arguments / new
│
├── 数据结构
│   ├── Object：简写、计算属性、合并拷贝
│   ├── Array：高阶方法链式调用
│   ├── Set：去重、存在判断
│   ├── Map：任意键、缓存、词频统计
│   └── Date：创建、运算、格式化
│
├── ES6+ 语法
│   ├── 解构（数组 + 对象 + 嵌套）
│   ├── 展开运算符（合并、拷贝、函数参数）
│   ├── 可选链 ?. + 空值合并 ??
│   └── 模板字符串
│
├── 数据操作
│   ├── JSON：stringify / parse
│   ├── 异常处理：try/catch/finally
│   └── reduce 模式：分组/统计/管道
│
├── 模块化
│   └── ES Module：export / import / type="module"
│
└── 实战：购物车系统（筛选→计算→优惠→统计）
```

---

## 19.1 DevTools Console 七步练习

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | F12 → Console，粘贴 §14 购物车 HTML 用 Live Server 打开 | 页面标题 + 控制台无红色报错 | 检查 script 是否闭合 |
| 2 | 在 Console 输入 `items`（若已暴露）或复制 `getCartItems` 单独跑 | 含 `subtotal` 的对象数组 | 检查 `productId` 是否匹配 |
| 3 | 改 `cart` 里数量为 0 再刷新 | 小计减少 | quantity 为 0 时是否应过滤（可自行加 filter） |
| 4 | 运行 `[10,2,1].sort()` 与 `.sort((a,b)=>a-b)` 对比 | 两种结果不同 | 复习 §15.4 |
| 5 | 试 `const {x} = undefined` 与 `const {x} = {}` | 前者报错后者 x 为 undefined | 用 `?? {}` 兜底 |
| 6 | `JSON.stringify({a:1,fn:()=>{}})` | fn 字段消失 | 理解 JSON 不能传函数 |
| 7 | 用链式：`products.filter(p=>p.price>100).map(p=>p.name)` | 字符串数组 | 箭头函数 `{}` 是否漏 return |

---

## 20. 闭卷自测

1. 箭头函数与普通函数在 `this` 上最核心的区别是什么？
2. `map` 和 `reduce` 各返回什么？空数组 `reduce` 不写初始值会怎样？
3. `0 || 10` 与 `0 ?? 10` 结果分别是什么？为什么？
4. 展开运算符 `{ ...a, ...b }` 若属性同名，谁覆盖谁？
5. `Set` 与 `Array.from(new Set(arr))` 的作用？
6. JSON 里为什么不能直接存 `undefined` 和函数？
7. 对象方法为什么推荐 `sayHi() {}` 而不是 `sayHi: () => {}`？
8. **动手**：写 `groupBy([{type:'a'},{type:'b'},{type:'a'}],'type')` 返回 `{ a:[...], b:[...] }`。
9. **动手**：解构 `const resp = { data: { list: [{ name: 'X' }] } }` 取出第一个 name（一行）。
10. **综合**：说明本章购物车逻辑在 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 里会出现在哪一层（提示：数据处理 vs 模板渲染）。

### 20.1 自测参考答案

1. 箭头函数 `this` 词法绑定（定义处外层）；普通函数 `this` 由调用方式决定。
2. `map` 返回同长度新数组；`reduce` 返回累积的单个值；空数组 reduce 无初始值 → TypeError。
3. 前者 `10`（0 是假值）；后者 `0`（?? 只认 null/undefined）。
4. 后面的 `b` 覆盖前面的 `a`。
5. Set 去重；`Array.from` 把 Set 转回数组。
6. JSON 规范只有 null/对象/数组/字符串/数字/布尔；undefined/函数无法序列化。
7. 方法简写里 `this` 指向调用对象；箭头函数 `this` 不是该对象。
8. `arr.reduce((acc,o)=>{(acc[o[k]]??=[]).push(o);return acc},{})` 或 for 循环版。
9. `const { data: { list: [{ name }] } } = resp` 或 `resp.data.list[0].name`。
10. Axios 拿到 JSON 后用 map/filter/reduce 清洗；渲染交给 Vue 模板或 08 章 DOM。

---

## 21. 费曼检验

3 分钟向没学过编程的朋友解释「本章在练什么」：

1. **函数是乐高块**：把「算总价」「用优惠券」拆成小块，拼起来维护简单——Vue 组件里也是小块组合。
2. **数组方法是流水线**：先 filter 再 map 再 reduce，像工厂质检——后端 JSON 列表到页面展示必经这条路。
3. **ES6 语法是安全带**：解构少写重复代码，可选链防「读取 undefined 的属性」崩溃——联调接口字段缺失时特别有用。

> 下一章：[08 — DOM、BOM 与事件机制](./08-JavaScript-DOM-BOM与事件机制.md)。把本章算好的数据挂到页面上。联调路线：本章 → 08/09 → [10 网络](./10-浏览器HTTP网络与Web基础.md) + [计网 04](../计算机网络/04-HTTP协议深入.md) → [Vue 08 Axios](../Vue/08-Axios网络请求与前后端联调.md)。
