# JavaScript 流程控制、函数、对象、数组与 ES6 基础

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

**Q：`sort` 为什么结果不对？**
默认按字符串 Unicode 排。`[10, 2, 1].sort()` → `[1, 10, 2]`。必须传比较函数 `.sort((a,b) => a - b)`。

**Q：`forEach` 和 `map` 什么时候用？**
需要返回新数组用 `map`；只是执行副作用（打印、改 DOM）用 `forEach`。

**Q：展开运算符会深拷贝吗？**
只拷贝一层。嵌套对象仍是引用，需要深拷贝用 `structuredClone()` 或 `JSON.parse(JSON.stringify())`（有局限）。

**Q：Set/Map 什么时候用？**
数组去重用 Set；统计词频/缓存用 Map；Object 的键只能是字符串。

**Q：可选链能连用吗？**
能。`obj?.a?.b?.c`，遇到任何一个 null/undefined 就返回 undefined。

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

> 下一章：08 — DOM、BOM 与事件机制。真正的页面交互从这里开始。
- 会用对象和数组解构、展开运算符、可选链、空值合并
- 能完成商品列表数据处理实战
- 了解 ES Module 的 `import`/`export` 基本写法
