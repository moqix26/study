# JavaScript 基础语法与数据类型

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0 导读（Vue/01 前置）、DevTools、FAQ、闭卷自测、费曼检验 -->

## 0. 读前导读（零基础也能跟上）

> **读者假设**：你会打开浏览器、能复制粘贴。本章是 **Vue 01～05 与 JavaScript 全路线** 的地基——不会本章的变量、数组、函数，学 Vue 的 `ref` 和 `v-for` 会很痛苦。

### 0.1 用一句话弄懂本章

**一句话**：JavaScript 是网页的「行为层」——用变量存数据、用函数封装逻辑、用对象和数组组织信息；Vue 框架里的所有代码本质上都是 JavaScript。

**生活类比**：

| JS 概念 | 类比 | Vue 里对应 |
|---------|------|------------|
| **变量 let/const** | 贴标签的储物盒 | `const count = ref(0)` 里的值 |
| **数组** | 购物清单 | `products` 商品列表 |
| **对象** | 一张名片 | `{ id, name, price }` 商品 |
| **函数** | 菜谱步骤 | `@click` 调用的 handler |
| **Console** | 厨房试吃台 | 调试 ref 和接口数据 |

**为什么重要**：[Vue 01](../Vue/01-Vue入门与环境搭建.md) 起每个 `.vue` 的 `<script setup>` 都是 JS；08 章调 [Java 04 REST API](../../后端学习/Java/04-SpringBoot核心开发.md) 时要处理 JSON 对象和数组。

---

### 0.2 你需要提前知道什么

| 水平 | 建议 |
|------|------|
| 完全零基础 | 先过 [HTML 01](./01-HTML基础结构与常用标签.md) 和 [03 CSS 基础](./03-CSS基础语法选择器与文本样式.md)（05 Flex 可在 Vue 前补） |
| 只会 HTML/CSS | **从 §2 第一个示例开始，每段代码在 Console 跑一遍** |
| 已会其他语言 | 重点看 §4 变量、§16 ===、§13 数组高阶方法 |
| 目标 Vue 路线 | 本章 + [07-JS 流程与 ES6](../HTML%20CSS%20JS/07-JavaScript流程控制函数对象数组与ES6基础.md) 后再开 Vue 01 |

---

### 0.3 本章知识地图（☐→☑）

- [ ] 会用 `let`/`const`，不用 `var`
- [ ] 知道 7 种原始类型 + object，会用 `typeof`（含陷阱）
- [ ] 熟练 `===` 与 15 个对照案例
- [ ] 会操作对象 CRUD、数组 push/map/filter/find/reduce
- [ ] 会写函数声明、表达式、箭头函数
- [ ] 会用 for / for...of / for...in
- [ ] 能在浏览器 Console 或 `.js` 文件运行代码
- [ ] 完成学生成绩管理系统 §21
- [ ] 闭卷自测 ≥ 8/10

---

### 0.4 建议学习时长

| 阶段 | 时间 |
|------|------|
| 变量 + 类型 §4～§11 | 3 小时 |
| 对象 + 数组 §12～§13 | 3 小时 |
| 运算符 + 流程 §14～§19 | 2 小时 |
| 函数 + 实战 §20～§21 | 2 小时 |
| Console 练习 §22 + 自测 | 1 小时 |

---

### 0.5 可验证成果

1. Console 输入 `typeof []`、`typeof null`、`0.1+0.2` 能解释结果。
2. 用 `filter` 从商品数组筛出价格 >100 的项（为 Vue computed 做准备）。
3. 写一个 `formatPrice(n)` 返回 `¥xx.xx`（Vue 模板会调用类似函数）。

---

### 0.6 核心术语三件套

**术语（let / const 块级变量）**：ES6 变量声明；`const` 默认首选，仅重新赋值时用 `let`。
**生活类比**：`const` 是贴死标签的盒子——不能换盒子，盒里东西可换；`let` 标签可撕了重贴。
**为什么重要**：Vue script 里 99% 用 `const` 声明 ref；与 `var` 混用会出作用域 bug。
**本章用到的地方**：§4 变量。

**术语（数组高阶方法 map/filter）**：不改动原数组（通常）返回新数组；map 映射、filter 筛选。
**生活类比**：流水线质检——map 给每个零件喷漆；filter 只留合格品。
**为什么重要**：Vue `computed` 过滤商品列表本质就是 `products.filter(...)` 的响应式版。
**本章用到的地方**：§13 数组；Vue [02 §15](../Vue/02-模板语法与响应式原理.md) 列表。

**术语（JSON 数据交换格式）**：前后端传数据的文本格式；键必须双引号。
**生活类比**：快递单标准格式——Java 04 后端填单，前端 `JSON.parse` 或 axios 自动拆包。
**为什么重要**：08 章联调 Java [04 Spring Boot API](../../后端学习/Java/04-SpringBoot核心开发.md) 全是 JSON。
**本章用到的地方**：§12 对象；07 章 fetch 预习。

---

## 1. JavaScript 是什么

JavaScript 是网页中的脚本语言。

它负责：

- 交互
- 逻辑
- 数据处理
- 与页面元素联动

前端三件套里：

- HTML 管结构
- CSS 管样式
- JavaScript 管行为

## 2. 第一个 JavaScript 示例

```html
<script>
  console.log("Hello JavaScript");
</script>
```

### `console.log`

作用：

- 在浏览器控制台输出信息

这是学习 JavaScript 最重要的调试入口之一。

## 3. JavaScript 的引入方式

### 3.1 内部脚本

```html
<script>
  alert("你好");
</script>
```

### 3.2 外部脚本

```html
<script src=”./main.js”></script>
```

实际开发更推荐外部文件。

### 3.3 `<script>` 放哪里？（重要）

很多初学者会遇到 `querySelector` 拿到 `null` 的问题，原因就是脚本执行时 DOM 还没加载。

**推荐做法（初学阶段）**：把 `<script>` 放在 `</body>` 前面。

```html
<body>
  <h1>标题</h1>
  <p>内容</p>
  <script src=”./main.js”></script>
</body>
```

这样浏览器会先解析 HTML，再执行 JS，保你操作 DOM 不出错。

### 3.4 完整可运行示例：三种引入方式对比

```html
<!DOCTYPE html>
<html lang=”zh-CN”>
<head>
  <meta charset=”UTF-8” />
  <title>JS 引入方式</title>
  <!-- 方式一：外部脚本（推荐） -->
  <script src=”external.js” defer></script>
</head>
<body>
  <h1>JS 引入方式演示</h1>
  
  <!-- 方式二：内部脚本 -->
  <script>
    console.log(“内部脚本：写在 HTML 里的 JS”);
  </script>
  
  <!-- 方式三：内联事件处理（不推荐，仅了解） -->
  <button onclick=”alert('行内事件')”>点我</button>
</body>
</html>
```

**小结**：
- 外部 `.js` 文件 → 项目开发首选
- 内部 `<script>` → 小 demo、学习示例
- 行内 `onclick` → 了解即可，不要用于项目

---

## 4. 变量

变量就是用来存储数据的”盒子”。JavaScript 中声明变量有三种方式：

- `var` — 旧写法，有坑
- `let` — 可修改的变量
- `const` — 不可重新赋值的常量

### 4.1 `let` — 最常用的变量声明

```js
let age = 18;
age = 19;     // ✅ 可以重新赋值
console.log(age); // 19
```

特点：
- 有**块级作用域**（`{}` 内声明的变量外面访问不到）
- 可以重新赋值
- 不会变量提升到全局（有暂时性死区 TDZ）

### 4.2 `const` — 常量声明（首选）

```js
const name = “Tom”;
// name = “Jerry”;  // ❌ 报错！const 不能重新赋值
```

特点：
- 有块级作用域
- 声明时必须初始化（不能只写 `const x;`）
- **对象和数组的内容可以修改**（只是绑定不能变）

```js
const user = { name: “小明” };
user.name = “小红”;    // ✅ 可以修改对象属性
user.age = 18;         // ✅ 可以添加属性
// user = {};           // ❌ 报错！不能重新赋值整个对象

const arr = [1, 2, 3];
arr.push(4);           // ✅ 可以修改数组内容
// arr = [5, 6];       // ❌ 报错！
```

**实践建议**：默认用 `const`，当你确定变量需要重新赋值时再用 `let`。

### 4.3 `var` — 旧写法（了解即可，尽量不用）

```js
var age = 18;
var age = 20;  // ⚠️ 可以重复声明！不会报错（容易出 bug）
```

`var` 的三个主要问题：

1. **没有块级作用域**：`if` / `for` 里的 `var` 会泄露到外面
2. **可以重复声明**：不小心重复声明同名变量不报错
3. **变量提升容易产生困惑**：var 声明的变量会被提升，赋值前访问得到 `undefined`

```js
// var 的问题演示：块级作用域缺失
if (true) {
  var x = 10;
}
console.log(x); // 10 —— 泄漏到外面了！

// let 安全
if (true) {
  let y = 10;
}
// console.log(y); // ❌ 报错！y 不存在
```

### 4.4 暂时性死区（TDZ）— let/const 的重要特性

`let` 和 `const` 在声明之前不能使用，这个区域叫”暂时性死区”：

```js
// console.log(a); // ❌ 报错：Cannot access 'a' before initialization
let a = 1;

console.log(b);    // undefined（var 提升，不报错但有隐患）
var b = 2;
```

### 4.5 `let`、`const`、`var` 完整对比表

| 特性 | `var` | `let` | `const` |
|------|-------|-------|---------|
| 能否重新赋值 | 能 | 能 | 不能（绑定不变，对象内容可改） |
| 块级作用域 | 无（函数级） | 有 | 有 |
| 重复声明 | 允许 | 不允许 | 不允许 |
| 变量提升 | 提升且初始化为 undefined | 提升但进入 TDZ（不可访问） | 提升但进入 TDZ |
| 全局声明时成为 window 属性 | 是 | 否 | 否 |
| 声明时是否需要初始化 | 否 | 否 | **是** |
| 现代项目推荐度 | ⭐（避免使用） | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐（首选） |

### 4.6 变量命名规范

```js
// ✅ 推荐命名
let userName = “小明”;        // 驼峰命名 camelCase
const MAX_COUNT = 100;        // 常量用全大写 + 下划线
let isLoading = true;         // 布尔值用 is/has/can 开头
const API_BASE_URL = “/api”;  // 配置常量全大写

// ❌ 不推荐命名
let a = 18;                   // 无意义
let shuz = 100;               // 拼音
let UserName = “小明”;        // 类名才用 PascalCase
let 1stItem = “first”;       // 不能以数字开头
```

### 4.7 完整可运行示例：变量实验

```html
<!DOCTYPE html>
<html lang=”zh-CN”>
<head>
  <meta charset=”UTF-8” />
  <title>变量声明实验</title>
</head>
<body>
  <h1>打开控制台（F12）看结果</h1>
  <script>
    // 实验 1：块级作用域
    console.log(“=== 实验 1：块级作用域 ===”);
    {
      var a = “var 变量”;
      let b = “let 变量”;
      const c = “const 变量”;
    }
    console.log(“var 在外面:”, a);        // 能访问
    // console.log(b); // 报错
    // console.log(c); // 报错
    
    // 实验 2：const 对象内容可变
    console.log(“=== 实验 2：const 对象 ===”);
    const user = { name: “小明”, age: 18 };
    user.age = 19;
    user.city = “北京”;
    console.log(“修改后的 user:”, user);   // { name: “小明”, age: 19, city: “北京” }
    
    // 实验 3：for 循环中的 var vs let
    console.log(“=== 实验 3：for 循环 ===”);
    for (var i = 0; i < 3; i++) {
      // 循环体
    }
    console.log(“var i 在外面:”, i);       // 3（泄漏了！）
    
    for (let j = 0; j < 3; j++) {
      // 循环体
    }
    // console.log(j); // 报错：j is not defined
    
    console.log(“✅ 所有实验完成，仔细看上面的输出！”);
  </script>
</body>
</html>
```

---

## 5. 数据类型

JavaScript 的数据类型分为两大类：**原始类型（Primitive）** 和 **引用类型（Reference）**。

### 5.1 七种原始类型

| 类型 | 说明 | 示例 |
|------|------|------|
| `number` | 整数和浮点数 | `42`, `3.14`, `Infinity`, `NaN` |
| `string` | 文本字符串 | `”hello”`, `'world'`, `` `模板` `` |
| `boolean` | 逻辑值 | `true`, `false` |
| `undefined` | 未定义（变量已声明未赋值） | `let x;` → `undefined` |
| `null` | 空值（故意的”无”） | `let user = null;` |
| `symbol` | 唯一标识符（ES6，进阶了解） | `Symbol(“id”)` |
| `bigint` | 超大整数（ES2020，进阶了解） | `9007199254740993n` |

### 5.2 引用类型：Object

`object` 是引用类型的基础。数组、函数、Date、RegExp 等本质上都是对象。

```js
const obj = { name: “小明” };   // 普通对象
const arr = [1, 2, 3];          // 数组（特殊的对象）
function fn() {}                // 函数（也是对象）
```

### 5.3 原始类型 vs 引用类型的核心区别

```js
// 原始类型：按值比较/复制
let a = 10;
let b = a;       // b 得到的是 10 这个值的副本
a = 20;
console.log(b);  // 10 —— 不受 a 修改的影响

// 引用类型：按引用比较/复制（共享同一份数据）
let obj1 = { value: 10 };
let obj2 = obj1;    // obj2 指向同一块内存
obj1.value = 20;
console.log(obj2.value); // 20 —— 受影响了！
```

### 5.4 typeof 运算符与陷阱

```js
console.log(typeof 123);        // “number”
console.log(typeof “hello”);    // “string”
console.log(typeof true);       // “boolean”
console.log(typeof undefined);  // “undefined”
console.log(typeof null);       // “object” ← 著名历史 bug！
console.log(typeof {});         // “object”
console.log(typeof []);         // “object” ← 数组也是 object
console.log(typeof function(){});// “function”
console.log(typeof Symbol());   // “symbol”
console.log(typeof 123n);       // “bigint”
```

**typeof 陷阱速查**：
| 表达式 | 结果 | 注意 |
|--------|------|------|
| `typeof null` | `”object”` | 语言历史 bug，记住即可 |
| `typeof []` | `”object”` | 判断数组用 `Array.isArray()` |
| `typeof NaN` | `”number”` | NaN 是数字类型 |

---

## 6. number — 数字类型详解

JavaScript 中所有数字（整数、小数、负数）都是 `number` 类型。内部使用 64 位浮点数存储（IEEE 754）。

### 6.1 基本数字

```js
let price = 99.9;     // 小数
let count = 10;       // 整数
let negative = -5;    // 负数
let hex = 0xff;       // 十六进制（255）
let binary = 0b1010;  // 二进制（10）
let octal = 0o77;     // 八进制（63）
let scientific = 1e6; // 科学计数法（1000000）
```

### 6.2 特殊数值

```js
// NaN — Not a Number（不是一个数字，但类型是 number）
console.log(0 / 0);              // NaN
console.log(parseInt("hello"));  // NaN
console.log(typeof NaN);         // "number" ← 注意！
console.log(NaN === NaN);        // false ← NaN 不等于任何值，包括自己！

// 判断 NaN 的正确方式：
console.log(Number.isNaN(NaN));  // true
console.log(Number.isNaN("hi")); // false（不会做类型转换）

// Infinity — 无穷大
console.log(1 / 0);              // Infinity
console.log(-1 / 0);             // -Infinity

// 安全整数范围（超过会不精确）
console.log(Number.MAX_SAFE_INTEGER); // 9007199254740991
console.log(Number.MIN_SAFE_INTEGER); // -9007199254740991
// 超出安全范围用 BigInt：9007199254740993n
```

### 6.3 浮点数精度问题（重要！必须知道）

```js
console.log(0.1 + 0.2);             // 0.30000000000000004 ← 不是 0.3！
console.log(0.1 + 0.2 === 0.3);     // false

// 解决方式：比较时用容差值
const result = 0.1 + 0.2;
console.log(Math.abs(result - 0.3) < 0.000001); // true

// 显示时保留小数位
console.log((0.1 + 0.2).toFixed(1));  // "0.3"
```

### 6.4 数字常用方法

```js
const num = 3.14159;

num.toFixed(2);         // "3.14"（四舍五入，返回字符串）
num.toFixed(0);         // "3"

Number.isInteger(10);   // true
Number.isInteger(10.5); // false

parseInt("123abc");     // 123（解析整数部分）
parseFloat("3.14px");   // 3.14（解析浮点数）
Number("123");          // 123（严格转换，"123abc" → NaN）

// 进制转换
(255).toString(16);     // "ff"
parseInt("ff", 16);     // 255
```

### 6.5 Math 对象常用方法

```js
Math.round(3.6);     // 4（四舍五入取整）
Math.floor(3.9);     // 3（向下取整）
Math.ceil(3.1);      // 4（向上取整）
Math.abs(-5);        // 5（绝对值）
Math.max(1, 3, 2);   // 3（取最大值）
Math.min(1, 3, 2);   // 1（取最小值）
Math.pow(2, 3);      // 8（2 的 3 次方），也可用 2 ** 3
Math.sqrt(16);       // 4（平方根）
Math.random();       // 0~1 的随机小数（不含 1）

// 随机整数：min~max（含两端）
function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}
console.log(randomInt(1, 10)); // 1~10 之间的整数
```

---

## 7. string — 字符串详解

### 7.1 三种写法

```js
let s1 = '单引号';
let s2 = "双引号";
let s3 = `反引号（模板字符串）`;

// 引号嵌套
let s4 = "It's a book";     // 双引号内可用单引号
let s5 = '他说："你好"';     // 单引号内可用双引号
let s6 = 'It\'s a book';    // 或用 \ 转义
```

### 7.2 转义字符

| 转义 | 含义 |
|------|------|
| `\'` | 单引号 |
| `\"` | 双引号 |
| `\\` | 反斜杠 |
| `\n` | 换行 |
| `\t` | Tab |

```js
console.log("第一行\n第二行");
// 输出两行
```

### 7.3 字符串不可变

任何字符串方法都返回**新字符串**，不修改原字符串。

```js
let s = "hello";
s.toUpperCase();       // 返回 "HELLO"
console.log(s);        // 仍是 "hello"！
s = s.toUpperCase();   // 必须重新赋值
```

### 7.4 常用方法速查

```js
const str = "Hello World";

// 长度
str.length;                       // 11

// 查找与判断
str.includes("World");            // true
str.startsWith("Hello");          // true
str.endsWith("ld");               // true
str.indexOf("o");                 // 4（第一个位置，找不到返回 -1）
str.lastIndexOf("o");             // 7（最后一个位置）

// 提取
str.slice(0, 5);                  // "Hello"（从 0 到 5，不含 5）
str.slice(6);                     // "World"（从 6 到末尾）
str.slice(-5);                    // "World"（负数从末尾倒数）

// 变换
str.toUpperCase();                // "HELLO WORLD"
str.toLowerCase();                // "hello world"
"  hi  ".trim();                  // "hi"（去首尾空格）
str.replace("World", "JS");       // "Hello JS"（替换第一个）

// 分割与拼接
str.split(" ");                   // ["Hello", "World"]
"a,b,c".split(",");              // ["a", "b", "c"]
"hello".split("");               // ["h","e","l","l","o"]

// 补全
"5".padStart(3, "0");             // "005"
"x".padEnd(5, "-");               // "x----"

// 重复
"Hi".repeat(3);                   // "HiHiHi"
```

### 7.5 字符串实操示例

```js
// 手机号脱敏
const phone = "13812345678";
phone.slice(0, 3) + "****" + phone.slice(-4); // "138****5678"

// 判断文件类型
const filename = "photo.jpg";
filename.endsWith(".jpg");   // true
filename.split(".").pop();   // "jpg"
```

---

## 8. 模板字符串（深入）

```js
const name = "小明";
const age = 18;

// 嵌入变量
const text = `你好，我是${name}，今年${age}岁。`;

// 嵌入表达式
console.log(`明年${age + 1}岁`);

// 嵌入函数调用
const price = 199;
console.log(`价格：¥${price.toFixed(2)}`);

// 多行字符串
const html = `
  <div class="card">
    <h2>${name}</h2>
    <p>年龄：${age}</p>
  </div>
`;

// 三元表达式
console.log(`${name}${age >= 18 ? "已成年" : "未成年"}`);
```

模板字符串是前后端交互、动态渲染 HTML 的核心工具，务必练熟。

---

## 9. boolean — 布尔值

```js
let isLogin = true;
let isLoading = false;
```

### 9.1 假值（falsy）— 以下值在 if 判断中等价于 false

`false`、`0`、`-0`、`0n`、`""`（空字符串）、`null`、`undefined`、`NaN`

**除了这 8 个（含 0n、-0），其他值都是 true**（包括空数组 `[]` 和空对象 `{}`）。

```js
if ([]) {}    // true！空数组也是 true
if ({}) {}    // true！空对象也是 true
if ("0") {}   // true！非空字符串
```

### 9.2 快速转布尔

```js
!!0           // false
!!1           // true
!!""          // false
!!"hello"     // true

// 实战：过滤数组中的假值
const arr = ["a", "", "b", null, "c", undefined];
const clean = arr.filter(Boolean); // ["a", "b", "c"]
```

---

## 10. undefined 详解

表示"变量已声明但未赋值"。

```js
let a;
console.log(a);              // undefined

function foo() {}            // 没写 return
console.log(foo());          // undefined

const obj = {};
console.log(obj.notExist);   // undefined（访问不存在的属性）
```

---

## 11. null 详解

表示"故意为空"——开发者主动设置的值。

```js
let user = null; // 还没登录，先设为 null

// typeof 历史 bug
console.log(typeof null); // "object" ← 记住即可

// 正确判断
console.log(user === null); // true
```

### undefined vs null

| | undefined | null |
|---|-----------|------|
| 谁设置的 | JS 引擎自动 | 开发者主动 |
| typeof | `"undefined"` | `"object"`（bug） |
| == null | true | true |
| === null | false | true |
| 转数字 | NaN | 0 |

---

## 12. object — 对象详解

对象是 JavaScript 的核心数据结构，它存储**键值对（key-value）**。

### 12.1 创建对象

```js
// 字面量创建（最常用）
const user = {
  name: "张三",
  age: 18,
  "phone-number": "13800138000", // 特殊属性名需要用引号
};

// 构造函数创建（了解即可）
const user2 = new Object();
user2.name = "李四";
```

### 12.2 访问属性

```js
const user = { name: "张三", age: 18 };

// 点语法（最常用）
console.log(user.name);       // "张三"

// 方括号语法（属性名是变量或含特殊字符时用）
const key = "name";
console.log(user[key]);       // "张三"
console.log(user["phone-number"]); // 属性名有连字符只能用方括号
```

### 12.3 增删改查

```js
const user = { name: "小明" };

// 增 / 改
user.age = 18;            // 添加属性
user.name = "大明";       // 修改属性

// 查
console.log("name" in user);    // true（检查属性是否存在）
console.log(user.hasOwnProperty("age")); // true

// 删
delete user.age;
console.log(user);              // { name: "大明" }
```

### 12.4 对象方法（值是函数的属性）

```js
const user = {
  name: "小明",
  sayHi() {                   // 方法简写
    console.log(`你好，我是${this.name}`);
  }
};

user.sayHi(); // "你好，我是小明"
```

### 12.5 遍历对象

```js
const user = { name: "小明", age: 18, city: "北京" };

// 遍历键名
for (const key of Object.keys(user)) {
  console.log(key);           // "name", "age", "city"
}

// 遍历值
for (const value of Object.values(user)) {
  console.log(value);         // "小明", 18, "北京"
}

// 同时遍历键和值
for (const [key, value] of Object.entries(user)) {
  console.log(`${key}: ${value}`);
}
// 输出：
// name: 小明
// age: 18
// city: 北京
```

### 12.6 对象常用方法

```js
const a = { x: 1 };
const b = { y: 2 };

// 合并对象
const merged = Object.assign({}, a, b); // { x: 1, y: 2 }
// 或使用展开运算符（更常用）
const merged2 = { ...a, ...b };          // { x: 1, y: 2 }
```

### 12.7 完整可运行示例：学生信息管理

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <title>对象练习</title>
</head>
<body>
  <h1>打开控制台（F12）看结果</h1>
  <script>
    console.log("=== 对象操作练习 ===\n");
    
    const student = {
      id: 1,
      name: "小明",
      scores: { math: 90, english: 85, chinese: 92 },
      // 方法：计算平均分
      getAverage() {
        const { math, english, chinese } = this.scores;
        return ((math + english + chinese) / 3).toFixed(1);
      },
      // 方法：生成简介
      getInfo() {
        return `${this.name}（学号：${this.id}），平均分：${this.getAverage()}`;
      }
    };
    
    console.log(student.getInfo());
    // "小明（学号：1），平均分：89.0"
    
    // 增删改查
    student.scores.chinese = 95;  // 改分数
    delete student.id;             // 删属性
    student.grade = "三年级";      // 加属性
    
    console.log("最终对象:", student);
    console.log("所有键:", Object.keys(student));
    console.log("所有值:", Object.values(student));
  </script>
</body>
</html>
```

---

## 13. 数组 Array — 详解

数组是有序的数据集合，通过数字索引（从 0 开始）访问元素。

### 13.1 创建数组

```js
// 字面量（最常用）
const list = ["HTML", "CSS", "JavaScript"];

// Array 构造函数（了解）
const arr = new Array(3); // [empty × 3]，不推荐

// Array.of（推荐，行为一致）
const arr2 = Array.of(1, 2, 3); // [1, 2, 3]
```

### 13.2 访问与修改

```js
const list = ["HTML", "CSS", "JavaScript"];

console.log(list[0]);         // "HTML"（索引从 0 开始）
console.log(list.length);     // 3（数组长度）
console.log(list[list.length - 1]); // "JavaScript"（最后一项）

list[1] = "SCSS";             // 修改指定位置
list[3] = "TypeScript";       // 添加新元素
console.log(list);            // ["HTML", "SCSS", "JavaScript", "TypeScript"]
```

### 13.3 增删改查（会改变原数组的方法标记 ⚠️）

```js
const arr = [1, 2, 3];

// 末尾操作
arr.push(4);          // ⚠️ [1,2,3,4] 尾部添加，返回新长度
arr.pop();            // ⚠️ [1,2,3]   尾部删除，返回被删元素

// 开头操作
arr.unshift(0);       // ⚠️ [0,1,2,3] 头部添加，返回新长度
arr.shift();          // ⚠️ [1,2,3]   头部删除，返回被删元素

// 截取（不改变原数组）
arr.slice(0, 2);      // [1, 2]     从索引 0 截到 2（不含 2）

// 增删改（万能方法，会改变原数组）
arr.splice(1, 1, "x"); // ⚠️ 从索引1删1个，插入"x" → [1, "x", 3]

// 查找
arr.includes(2);      // true
arr.indexOf(2);       // 1（找不到返回 -1）

// 排序
arr.sort((a, b) => a - b); // ⚠️ 升序排列

// 反转
arr.reverse();         // ⚠️ 反转数组

// 转字符串
arr.join(", ");       // "1, 2, 3"
```

### 13.4 数组遍历方法（这些是重点，日常开发大量使用）

```js
const nums = [1, 2, 3, 4, 5];

// forEach：遍历每个元素（无返回值）
nums.forEach((item, index) => {
  console.log(index, item);
});

// map：映射为新数组（返回新数组，长度不变）
const doubled = nums.map(n => n * 2);     // [2, 4, 6, 8, 10]

// filter：过滤（返回新数组，长度可能变）
const even = nums.filter(n => n % 2 === 0); // [2, 4]

// find：找到第一个匹配的元素
const found = nums.find(n => n > 3);       // 4

// findIndex：找到第一个匹配的索引
const idx = nums.findIndex(n => n > 3);    // 3

// some：是否有任意一个满足条件
const hasBig = nums.some(n => n > 10);      // false

// every：是否全部满足条件
const allSmall = nums.every(n => n < 10);   // true

// reduce：累积计算（万能方法）
const sum = nums.reduce((累计, 当前) => 累计 + 当前, 0); // 15
```

### 13.5 完整可运行示例：数组实战

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <title>数组练习</title>
</head>
<body>
  <h1>打开控制台（F12）看结果</h1>
  <script>
    console.log("=== 数组操作练习 ===\n");
    
    // 1. 价格计算
    const prices = [199, 89, 299, 159, 49];
    const total = prices.reduce((sum, p) => sum + p, 0);
    console.log("总价:", total);
    console.log("均价:", (total / prices.length).toFixed(2));
    
    // 2. 筛选 + 映射
    const cheap = prices
      .filter(p => p < 100)
      .map(p => `¥${p}`);
    console.log("低于100的商品:", cheap);    // ["¥89", "¥49"]
    
    // 3. 去重
    const tags = ["前端", "后端", "前端", "设计", "后端", "前端"];
    const unique = [...new Set(tags)];
    console.log("去重后:", unique);            // ["前端", "后端", "设计"]
    
    // 4. 按条件排序
    const students = [
      { name: "小明", score: 90 },
      { name: "小红", score: 95 },
      { name: "小刚", score: 85 },
    ];
    // 按分数降序
    students.sort((a, b) => b.score - a.score);
    console.log("排名:", students.map(s => `${s.name}:${s.score}`).join(", "));
    
    // 5. 分组统计
    const allTags = ["前端", "后端", "前端", "设计"];
    const count = allTags.reduce((acc, tag) => {
      acc[tag] = (acc[tag] || 0) + 1;
      return acc;
    }, {});
    console.log("标签统计:", count);           // { 前端: 2, 后端: 1, 设计: 1 }
  </script>
</body>
</html>
```

---

## 14. `typeof` — 类型检测

```js
console.log(typeof 123);        // "number"
console.log(typeof "abc");      // "string"
console.log(typeof true);       // "boolean"
console.log(typeof undefined);  // "undefined"
console.log(typeof null);       // "object" ← 历史 bug
console.log(typeof {});         // "object"
console.log(typeof []);         // "object" ← 判断数组用 Array.isArray()
console.log(typeof function(){});// "function"
```

### typeof 陷阱速查

| 表达式 | 结果 | 正确判断方式 |
|--------|------|-------------|
| `typeof null` | `"object"` | `x === null` |
| `typeof []` | `"object"` | `Array.isArray(x)` |
| `typeof NaN` | `"number"` | `Number.isNaN(x)` |

---

## 15. 运算符

### 15.1 算术运算符

```js
let a = 10;
let b = 3;

console.log(a + b);  // 13 加
console.log(a - b);  // 7  减
console.log(a * b);  // 30 乘
console.log(a / b);  // 3.333... 除（注意不是整数除法）
console.log(a % b);  // 1  取余
console.log(a ** b); // 1000 指数（ES7）

// 自增/自减
let count = 0;
count++;  // 后置自增（先返回再 +1）
++count;  // 前置自增（先 +1 再返回）
count--;  // 自减
```

### 15.2 赋值运算符

```js
let x = 10;
x += 5;   // x = x + 5  → 15
x -= 3;   // x = x - 3  → 12
x *= 2;   // x = x * 2  → 24
x /= 4;   // x = x / 4  → 6
x %= 4;   // x = x % 4  → 2
```

### 15.3 比较运算符

```js
// 大小比较
3 > 2;    // true
3 < 2;    // false
3 >= 3;   // true
3 <= 2;   // false

// 相等比较（非常关键！）
5 == "5";    // true（宽松相等，会做类型转换）
5 === "5";   // false（严格相等，类型不同）
5 != "5";    // false（宽松不等）
5 !== "5";   // true（严格不等）
```

### 15.4 逻辑运算符

```js
// &&（与）：两边都为 true 才为 true
true && true;     // true
true && false;    // false

// ||（或）：任意一边为 true 就为 true
true || false;    // true
false || false;   // false

// !（非）：取反
!true;            // false
!false;           // true

// 短路运算（实战中非常常见）
const name = inputName || "默认用户名";  // 如果 inputName 是假值就用默认值
const result = isReady && doSomething(); // 只有 isReady 为 true 才执行

// 空值合并 ??（ES2020，更精确的默认值）
0 || "默认";      // "默认" —— 0 是假值被跳过了
0 ?? "默认";      // 0 —— 只有 null/undefined 才用默认值
```

---

## 16. `==` 和 `===` — 彻底搞懂

这是 JavaScript 最容易出 bug 的地方之一。

### 16.1 核心规则

```js
// === 严格相等：类型和值都必须相同
// == 宽松相等：如果类型不同，先转换再比较
```

### 16.2 `==` 的转换规则（了解即可，关键是不要用！）

```js
// 规则 1：null == undefined 为 true（特例）
null == undefined;       // true

// 规则 2：字符串和数字比较时，字符串转数字
5 == "5";               // true（"5" 转成 5）
0 == "";                // true（"" 转成 0）

// 规则 3：布尔值转数字
true == 1;              // true
false == 0;             // true

// 规则 4：对象和原始值比较时，对象转原始值
[1] == 1;               // true（[1] 转成 "1" 再转 1）
```

### 16.3 `==` vs `===` 对照实验（15个必知案例）

```js
//              ==          ===
false == 0;     // true      false
false == "";    // true      false
0 == "";        // true      false
0 == "0";       // true      false
null == undefined; // true   false
" \t\r\n" == 0; // true      false
[] == 0;        // true      false
[] == "";       // true      false
[1] == 1;       // true      false
[1,2] == "1,2"; // true      false
NaN == NaN;     // false     false（NaN 不等于任何值！）
!null == true;  // true      ——（!null → true, true == true）
!!"" == false;  // true      ——（!!"" → false, false == false）
```

**记忆口诀**：写比较用 `===`，不要用 `==`，除非你明确知道自己在做什么。

---

## 17. 类型转换

### 17.1 显式转换（推荐，意图明确）

```js
// 转字符串
String(123);          // "123"
String(true);         // "true"
(123).toString();     // "123"

// 转数字
Number("123");        // 123
Number("123abc");     // NaN
parseInt("123abc");   // 123（解析到非数字为止）
parseFloat("3.14px"); // 3.14
const n = +"123";     // 123（一元 + 快速转数字）

// 转布尔
Boolean(1);           // true
Boolean(0);           // false
Boolean("");          // false
const b = !!123;      // true（双重否定快速转布尔）
```

### 17.2 隐式转换（了解即可，但经常遇到）

```js
// 字符串拼接时，其他类型自动转字符串
"5" + 1;        // "51"（数字 1 转成了 "1"）
"5" + true;     // "5true"
"5" + null;     // "5null"

// 数学运算时，字符串尽量转数字
"5" - 1;        // 4（减号触发了数字转换）
"5" * "2";      // 10
"5" / "2";      // 2.5

// 条件判断时转布尔
if (1) {}       // 1 → true
if (0) {}       // 0 → false
```

---

## 18. 流程控制

### 18.1 `if...else if...else`

```js
const score = 85;

if (score >= 90) {
  console.log("优秀");
} else if (score >= 80) {
  console.log("良好");
} else if (score >= 60) {
  console.log("及格");
} else {
  console.log("不及格");
}
// 输出：良好
```

### 18.2 三元运算符（`if` 的简洁写法）

```js
const age = 16;
const canVote = age >= 18 ? "可以投票" : "不可以投票";
console.log(canVote); // "不可以投票"

// 嵌套（可读性会变差，不建议嵌套太深）
const level = score >= 90 ? "A" : score >= 60 ? "B" : "C";
```

### 18.3 `switch`

```js
const status = 1;
switch (status) {
  case 0:
    console.log("待支付");
    break;
  case 1:
    console.log("已支付");
    break;
  default:
    console.log("未知状态");
}
```

---

## 19. 循环 — 详解

### 19.1 `for` 循环（最常用）

```js
// 标准格式：for (初始化; 条件; 每次循环后执行)
for (let i = 0; i < 5; i++) {
  console.log(i); // 0, 1, 2, 3, 4
}

// 遍历数组
const fruits = ["苹果", "香蕉", "橙子"];
for (let i = 0; i < fruits.length; i++) {
  console.log(`${i}: ${fruits[i]}`);
}
```

### 19.2 `for...of`（遍历数组元素，推荐）

```js
const fruits = ["苹果", "香蕉", "橙子"];

for (const fruit of fruits) {
  console.log(fruit);
}
// 苹果, 香蕉, 橙子 — 直接拿到值，不需要索引

// 如果需要索引
for (const [index, fruit] of fruits.entries()) {
  console.log(`${index}: ${fruit}`);
}
```

### 19.3 `for...in`（遍历对象键名，遍历数组索引）

```js
// 遍历对象
const user = { name: "小明", age: 18 };
for (const key in user) {
  console.log(`${key}: ${user[key]}`);
}

// ⚠️ 不推荐用于数组（拿到的是索引字符串，不是值）
const arr = ["a", "b"];
for (const i in arr) {
  console.log(i);     // "0", "1" — 字符串索引！
  console.log(arr[i]); // "a", "b"
}
```

### 19.4 `while` 和 `do...while`

```js
// while：先判断再执行
let i = 0;
while (i < 5) {
  console.log(i);
  i++;
}

// do...while：先执行一次再判断
let j = 10;
do {
  console.log(j); // 至少执行一次
  j++;
} while (j < 5);  // 条件为 false，不再继续
```

### 19.5 `break` 和 `continue`

```js
// break：终止整个循环
for (let i = 0; i < 10; i++) {
  if (i === 5) break;
  console.log(i); // 0, 1, 2, 3, 4
}

// continue：跳过本次，继续下一次
for (let i = 0; i < 5; i++) {
  if (i === 2) continue;
  console.log(i); // 0, 1, 3, 4
}
```

### 19.6 循环选择指南

| 场景 | 推荐 | 原因 |
|------|------|------|
| 遍历数组，要值 | `for...of` | 简洁，直接拿到值 |
| 遍历数组，要索引 | `for` 传统式 | 控制粒度最细 |
| 遍历数组，要索引+值 | `.entries()` + `for...of` | 解构拿到两者 |
| 遍历对象键名 | `for...in` | 唯一能直接遍历对象键的循环 |
| 遍历对象键值 | `Object.entries()` + `for...of` | 更安全（不含原型链） |
| 已知循环次数 | `for` | 意图最清晰 |
| 未知循环次数（条件） | `while` | 符合语义 |

---

## 20. 函数 — 详解

### 20.1 函数声明

```js
function add(a, b) {
  return a + b;
}

console.log(add(2, 3)); // 5

// 函数声明有"提升"：可以在定义前调用
sayHello(); // ✅ 可以
function sayHello() {
  console.log("Hello");
}
```

### 20.2 参数与默认值

```js
// 基本参数
function greet(name) {
  console.log(`你好，${name}`);
}
greet("小明"); // "你好，小明"
greet();       // "你好，undefined" ← 没传参就是 undefined

// 默认参数
function greet2(name = "游客") {
  console.log(`你好，${name}`);
}
greet2(); // "你好，游客"

// 多个参数 + 默认值
function createUser(name, age = 18, city = "北京") {
  return { name, age, city };
}
console.log(createUser("小明", 20)); // { name: "小明", age: 20, city: "北京" }
```

### 20.3 `return` 返回值

```js
function multiply(a, b) {
  return a * b;             // 返回结果，函数结束
  console.log("这行不会执行"); // 永远不会运行
}

console.log(multiply(3, 4)); // 12

// 没有 return 的函数返回 undefined
function noReturn() {
  const x = 1;
}
console.log(noReturn()); // undefined

// return;（空返回）也返回 undefined
function earlyReturn(condition) {
  if (!condition) return;   // 提前退出
  console.log("继续执行");
}
earlyReturn(false); // 什么都不输出
```

### 20.4 函数表达式

```js
// 把函数赋值给变量
const add = function(a, b) {
  return a + b;
};

console.log(add(2, 3)); // 5

// 函数表达式没有"提升"
// sayHi(); // ❌ 报错！Cannot access before initialization
const sayHi = function() {
  console.log("Hi");
};
```

### 20.5 箭头函数（非常重要）

```js
// 完整写法
const add = (a, b) => {
  return a + b;
};

// 简写：函数体只有一句 return 时，可省略 {} 和 return
const add2 = (a, b) => a + b;

// 只有一个参数时，可省略 ()
const square = n => n * n;

// 没有参数时，必须有 ()
const sayHi = () => console.log("Hi");

// 返回对象字面量——必须包一层 ()
const createUser = (name, age) => ({ name, age });
```

### 20.6 函数声明 vs 函数表达式 vs 箭头函数

| | 函数声明 | 函数表达式 | 箭头函数 |
|---|----------|-----------|----------|
| 写法 | `function f() {}` | `const f = function(){}` | `const f = () => {}` |
| 提升 | ✅ 可以在定义前调用 | ❌ | ❌ |
| `this` 绑定 | 调用时决定 | 调用时决定 | 定义时捕获（无自己的 this） |
| 适合场景 | 顶层工具函数 | 需要提升的场景 | 回调、数组方法（最常用） |

---

## 21. 综合实战：学生成绩管理系统

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <title>学生成绩管理（Console 练习）</title>
</head>
<body>
  <h1>打开控制台（F12）看完整输出</h1>
  <script>
    console.log("========== 学生成绩管理系统 ==========\n");

    // 1. 数据
    const students = [
      { name: "小明", math: 90, english: 85, chinese: 92 },
      { name: "小红", math: 95, english: 88, chinese: 96 },
      { name: "小刚", math: 55, english: 60, chinese: 58 },
      { name: "小丽", math: 78, english: 92, chinese: 88 },
      { name: "小强", math: 42, english: 50, chinese: 45 },
    ];

    // 2. 计算每个学生的平均分和等级
    function getLevel(avg) {
      if (avg >= 90) return "优秀";
      if (avg >= 80) return "良好";
      if (avg >= 60) return "及格";
      return "不及格";
    }

    const processed = students.map(s => {
      const avg = (s.math + s.english + s.chinese) / 3;
      return {
        ...s,
        average: +avg.toFixed(1),  // + 转数字
        level: getLevel(avg),
        total: s.math + s.english + s.chinese,
      };
    });

    console.log("📊 所有学生成绩：");
    console.table(processed.map(s => ({
      姓名: s.name,
      数学: s.math,
      英语: s.english,
      语文: s.chinese,
      平均分: s.average,
      等级: s.level,
    })));

    // 3. 统计数据
    const allAvg = processed.map(s => s.average);
    const classAvg = (allAvg.reduce((a, b) => a + b, 0) / allAvg.length).toFixed(1);

    const passRate = (processed.filter(s => s.level !== "不及格").length / processed.length * 100).toFixed(0);

    const best = processed.reduce((best, s) => s.average > best.average ? s : best);

    const mathAvg = (processed.reduce((sum, s) => sum + s.math, 0) / processed.length).toFixed(1);

    console.log("\n📈 班级统计：");
    console.log(`  班级平均分: ${classAvg}`);
    console.log(`  及格率: ${passRate}%`);
    console.log(`  第一名: ${best.name}（${best.average}分）`);
    console.log(`  数学平均分: ${mathAvg}`);

    // 4. 筛选功能
    console.log("\n🔍 筛选结果：");
    console.log("  优秀学生:", processed.filter(s => s.level === "优秀").map(s => s.name).join("、"));
    console.log("  不及格学生:", processed.filter(s => s.level === "不及格").map(s => s.name).join("、"));
    console.log("  数学不及格:", processed.filter(s => s.math < 60).map(s => `${s.name}(${s.math}分)`).join("、"));
    console.log("  总成绩 > 260:", processed.filter(s => s.total > 260).map(s => `${s.name}(${s.total})`).join("、"));

    console.log("\n✅ 系统运行完毕！建议你把代码复制到本地改改数据玩。");
  </script>
</body>
</html>
```

---

## 22. 在浏览器里怎么练 JS（含 DevTools 步骤表）

| 方法 | 操作 | 适合 |
|------|------|------|
| Console 直接敲 | `F12` → Console → 输入代码回车 | 快速试验、计算 |
| 内部 `<script>` | `<script>你的代码</script>` 放 `</body>` 前 | 小练习 |
| 外部 `.js` 文件 | `<script src="main.js"></script>` 放 `</body>` 前 | 正式练习 |

**建议**：学语法阶段多用 Console；到了 DOM 阶段再写 HTML 文件。

### 22.1 Console 手把手步骤

| 步骤 | 动作 | 预期 | 若不对 |
|------|------|------|--------|
| 1 | 任意网页 F12 → **Console** | 出现 `>` 输入提示 | 切到 Console 标签 |
| 2 | 输入 `console.log('hello')` 回车 | 输出 hello | 检查引号配对 |
| 3 | 输入 `let a = 1; a + 2` | 显示 3 | 不要用中文分号 |
| 4 | 定义 `const arr = [1,2,3]` | 展开可见元素 | — |
| 5 | `arr.map(x => x * 2)` | `[2,4,6]` | 箭头函数语法 |
| 6 | **Sources** 新建 Snippet 贴 §21 代码 | Run 可执行 | 长代码别全挤 Console |
| 7 | Vue 项目 Console 看报错 | 红色栈指向行号 | 配合 [Vue 01 §13.3](../Vue/01-Vue入门与环境搭建.md) |

**与 Vue 的关系**：`.vue` 的 `<script setup>` 就是 JS；`ref`、箭头函数、数组方法本章都要练熟。

**与后端 JSON**：[Java 04](../../后端学习/Java/04-SpringBoot核心开发.md) 返回 JSON 对象/数组，用本章 `.属性`、`map/filter` 解析后再给 Vue 的 ref。

### 22.2 为 Vue 02/03 准备的 map/filter 迷你练习

在 Console 粘贴运行（模拟 08 章接 API 后的数据处理）：

```js
// 模拟 Java 04 GET /api/products 返回的 data 数组
const products = [
  { id: 1, name: 'Java 编程思想', price: 99, stock: 10 },
  { id: 2, name: 'Spring Boot 实战', price: 79, stock: 0 },
  { id: 3, name: 'Redis 设计与实现', price: 89, stock: 3 },
]

// 对应 Vue 03 computed 过滤：名称含 spring 且有货
const filtered = products.filter(
  p => p.name.toLowerCase().includes('spring') && p.stock > 0
)
console.log('filtered', filtered) // []

// 对应 Vue 模板 formatPrice：映射展示价
const prices = products.map(p => `¥${p.price.toFixed(2)}`)
console.log('prices', prices)

// 对应购物车总价 reduce
const total = products
  .filter(p => p.stock > 0)
  .reduce((sum, p) => sum + p.price, 0)
console.log('有货商品标价总和', total) // 188
```

| 方法 | Vue 里谁用 | Java 04 场景 |
|------|------------|--------------|
| `filter` | computed 搜索/有货 | 前端筛选项，后端仍分页 |
| `map` | 列表展示转换 | DTO 数组 → 卡片字段 |
| `find` | 按 id 找商品 | 详情页路由参数 id |
| `reduce` | 购物车总价 | 订单小计预览 |

---

## 23. 初学者常见错误（扩展版）

### 23.1 混用 `==` 和 `===`
一律优先 `===`，除非你明确需要类型转换。

### 23.2 忘记 `return`
```js
// ❌ 错
function double(n) {
  n * 2; // 没写 return，返回 undefined
}
// ✅ 对
function double(n) {
  return n * 2;
}
```

### 23.3 `const` 对象/数组误以为不能改
```js
const user = { name: "小明" };
user.name = "小红"; // ✅ 可以！只是不能 user = {...}

const arr = [1, 2, 3];
arr.push(4);        // ✅ 可以！只是不能 arr = [...]
```

### 23.4 不会看控制台报错
红色报错第一行告诉你了**文件、行号、错误类型**。先看这三个，比猜重要 100 倍。

### 23.5 变量名无意义
```js
// ❌ 差
let a = 18;
let shuz = 100;
// ✅ 好
let age = 18;
let price = 100;
```

### 23.6 `0.1 + 0.2 !== 0.3`
浮点数精度问题，用 `.toFixed()` 或比较时用容差值。

### 23.7 `NaN === NaN` 为 false
判断 NaN 用 `Number.isNaN()`。

### 23.8 把 `for...in` 用在数组上
遍历索引字符串，而不是值。数组应用 `for...of` 或 `forEach`/`map`。

---

## 24. 分级练习

### 基础（必做）
1. 用 `if/else` 判断成绩等级（90+ 优秀，80+ 良好，60+ 及格，否则不及格）
2. 用 `for` 循环打印 1~100 中的所有偶数
3. 定义函数 `calcArea(width, height)` 返回矩形面积
4. 用 `for...of` 遍历 `["HTML", "CSS", "JS"]`，打印 `"第X门：xxx"`

### 进阶
1. 完成第 21 节学生成绩系统，加 `filter` 筛出平均分 > 85 的学生
2. 用模板字符串拼接用户信息卡片文字（含三元判断成年/未成年）
3. 用 `reduce` 统计数组中每个元素出现的次数

### 挑战
1. 写一个 `isPalindrome(str)` 判断字符串是否回文（如 `"上海自来水来自海上"` → true）
2. 只用 `let`/`const`（不用 `var`）重写你之前所有练习
3. 模拟购物车：商品数组，计算总价、应用折扣、格式化输出

---

## 25. FAQ 常见问题

**Q：`<script>` 放 head 还是 body？**  
初学放 `</body>` 前最省心；以后可学 `defer`/`async`。

**Q：单引号还是双引号？**  
团队统一即可。JSON 必须用双引号。模板字符串用反引号。

**Q：数组是对象吗？**  
`typeof []` 返回 `"object"`。数组是特殊的对象，用 `Array.isArray()` 判断。

**Q：`const` 的对象属性为什么可以改？**  
`const` 锁的是"变量绑定"（不能指向新对象），但对象内部内容不受 `const` 限制。

**Q：什么时候用 `let`，什么时候用 `const`？**  
默认用 `const`，只在确实需要重新赋值时用 `let`。永远不用 `var`。

**Q：箭头函数和普通函数有什么区别？**  
箭头函数更短，没有自己的 `this`，不能用 `new`，适合回调和数组方法。

**Q：为什么 Vue 文档里全是 `const`？**  
ref 变量本身不重新赋值，改的是 `.value`；见 [Vue 01](../Vue/01-Vue入门与环境搭建.md)。

**Q：JSON 和 JS 对象一样吗？**  
JSON 是字符串，键须双引号；`JSON.parse` 后才是对象，08 章接 [Java 04 API](../../后端学习/Java/04-SpringBoot核心开发.md) 时常用。

**Q：null 和 undefined 区别？**  
undefined 未赋值；null 故意为空。`typeof null === 'object'` 是历史遗留。

**Q：学到什么程度再开 Vue？**  
本章 + [07 ES6 基础](./07-JavaScript流程控制函数对象数组与ES6基础.md)；至少会变量、函数、map/filter、对象。

**Q：Console 报 SyntaxError？**  
括号/引号不配或中文标点；见 §23 常见错误表。

**Q：和 Java 语法像吗？**  
JS 动态类型；后端 Java 04 用强类型 DTO，前端用对象灵活映射字段。

---

## 26. 学完标准

- [ ] 理解 JavaScript 基础语法，能在 Console 或 `.js` 文件里运行代码
- [ ] 会定义 `let`/`const` 变量，能解释和 `var` 的区别
- [ ] 知道 7 种原始类型 + object，能用 `typeof` 判断，知道 typeof 的陷阱
- [ ] 熟练掌握 `===`、`==` 区别和 15 个对照案例
- [ ] 理解真值假值（8 个 falsy 值），会短路运算和 `??`
- [ ] 会操作对象和数组，能用 `map`/`filter`/`find`/`reduce` 处理数据
- [ ] 会用 `for`/`for...of`/`for...in` 正确遍历
- [ ] 能写三种函数形式，知道各自区别
- [ ] 知道 `0.1 + 0.2 !== 0.3` 的原因和解决方案
- [ ] 会用 `console.log` / `console.table` 查看运行结果
- [ ] 能独立完成学生成绩管理系统和购物车计算等综合练习

---

## 27. 本章思维导图

```
JavaScript 基础全景
│
├── 变量
│   ├── const（默认，不可重新赋值）
│   ├── let（可改，有块作用域）
│   └── var（旧，避免使用）
│
├── 数据类型
│   ├── 原始类型：number string boolean undefined null symbol bigint
│   ├── 引用类型：object（含 array、function）
│   └── typeof（记住陷阱：null、[]、NaN）
│
├── 运算符
│   ├── 算术：+ - * / % **
│   ├── 比较：===（首选）> ==（避免）
│   ├── 逻辑：&& || ! ??
│   └── 短路运算与默认值
│
├── 流程控制
│   ├── if/else if/else
│   ├── switch
│   └── 三元运算符 ?:
│
├── 循环
│   ├── for（已知次数）
│   ├── for...of（遍历数组值）
│   ├── for...in（遍历对象键）
│   ├── while（条件循环）
│   └── break / continue
│
├── 函数
│   ├── 函数声明（有提升）
│   ├── 函数表达式（无提升）
│   └── 箭头函数（短，无 this）
│
├── 核心操作
│   ├── 对象：CRUD、遍历、entries
│   ├── 数组：push/pop/shift/splice/slice/join
│   ├── 高阶：map/filter/find/reduce/some/every
│   └── 字符串：slice/includes/split/replace/trim
│
└── 实战：学生成绩管理 / 购物车计算
```

---

## 28. 闭卷自测

1. `let`、`const`、`var` 在作用域和重新赋值上各有什么不同？
2. 列出 8 个 falsy 值。
3. `typeof null` 和 `typeof []` 分别返回什么？为什么？
4. `===` 和 `==` 在 `0 == false` 和 `0 === false` 上结果各是什么？
5. 数组 `map` 和 `filter` 各返回什么？会改原数组吗？
6. 箭头函数为什么不适合做构造函数？
7. 对象和数组在 Vue 的 `products` 列表里如何配合使用？
8. **动手**：Console 里用 `filter` 筛出价格 >100 的商品数组。
9. **动手**：写 `formatPrice(12.5)` 返回 `"¥12.50"`。
10. **综合**：Java 04 返回 `[{id:1,name:"A",price:99}]`，如何用 JS 取出所有 name（两种写法）？

### 28.1 自测参考答案

1. var 函数作用域可重复声明；let/const 块作用域；const 不能重新绑定，let 可以。
2. `false`、`0`、`''`、`null`、`undefined`、`NaN`、`document.all`（了解）、`0n`。
3. 都 `"object"`；null 历史 bug；数组是特殊对象。
4. `0==false` 为 true（转换）；`0===false` 为 false。
5. map 返回同长度新数组（映射）；filter 返回满足条件的新数组；都不改原数组（除非回调里改元素）。
6. 没有 prototype，不能 `new`。
7. 数组存多个商品对象；`v-for="p in products"` 访问 `p.name` 等。
8. `products.filter(p => p.price > 100)`。
9. `` `¥${n.toFixed(2)}` `` 或字符串拼接。
10. `arr.map(x=>x.name)` 或 `for...of` 循环 push。

---

## 29. 费曼检验

3 分钟向朋友解释「JavaScript 在本章解决了什么」：

1. **变量与类型**：给数据贴标签，知道数字、字符串、对象、数组各是什么。
2. **函数与数组方法**：把重复逻辑打包；map/filter 像流水线处理列表——Vue 的 computed 本质就是这类思维。
3. **Console 调试**：写 Vue 之前先在浏览器试代码；以后 JSON 从 Java 后端来，还是 JS 对象和数组。

> 下一章：07 — JavaScript 流程控制、函数、对象、数组深度与 ES6 基础。建议先把本章练习全部做完再继续；然后可开 [Vue 01](../Vue/01-Vue入门与环境搭建.md)。
