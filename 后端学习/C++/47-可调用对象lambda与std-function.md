# 可调用对象、Lambda 与 std::function

> **文件编码**：UTF-8。

---

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

**可调用对象** = 像函数一样能被 `()` 调用的东西：普通函数、函数对象、lambda、`std::function`——STL 算法与回调的通用「插头」。

### 0.2 你需要提前知道什么

- [05 章](05-现代C++新特性.md) lambda 初识；[04 章](04-STL标准库容器与算法.md) `sort` 比较器
- [06 章](06-模板与泛型编程.md) 模板与 `operator()`
- [03 章](03-面向对象与类设计.md) 类与成员函数

### 0.3 本章知识地图（☐→☑）

- [ ] 写带 `operator()` 的函数对象
- [ ] 掌握 lambda 四种捕获与 init-capture
- [ ] 理解 `mutable` 与 `constexpr` lambda
- [ ] 会用 `std::function` 存任意可调用体
- [ ] 了解 `std::bind` 与 C++20 `std::bind_front`
- [ ] §25 闭卷自测 ≥8/10

### 0.4 建议学习时长

**3～5 天**；建议边读边写小 demo，编译器报错是最好的老师。

### 0.5 学完你能做什么

给 `std::sort`/`for_each` 写 lambda；用 `std::function` 做策略模式；读懂 Asio 回调签名；避免悬空引用捕获。

### 0.6 与 Java / Python 对照

| 本章 | 对照 |
|------|------|
| lambda / functor | Java `@FunctionalInterface`；Python `lambda`/`def` |
| std::function | Java `Function`；Python 一等函数对象 |
| 捕获生命周期 | Java 闭包 final 变量；Python 闭包 late binding |
| bind | Java `MethodHandle`；Python `functools.partial` |

---

## 本章与上一章的关系

[46 章](46-右值引用与移动语义进阶.md)（若已学）巩固右值与完美转发；本章把「函数当值传递」讲到工程级。

---

## 1. 这份文档学什么

- 函数指针与 typedef/using 别名
- 函数对象（Functor）与 `operator()`
- lambda 语法、捕获列表、泛型 lambda
- `mutable`、立即调用 IIFE
- `std::function` 类型擦除与开销
- `std::bind`/`bind_front` 与占位符
- 可调用对象在 STL 算法中的应用
- 生命周期陷阱与最佳实践

---

## 2. 函数指针与可调用概念

C++ 里「可调用（Callable）」是概念而非单一类型：任何能被 `std::invoke` 调用的实体。

```cpp
#include <iostream>

int add(int a, int b) { return a + b; }

int main() {
    int (*fp)(int, int) = add;          // 函数指针
    std::cout << fp(2, 3) << '\n';     // 5
    auto f = add;                       // 函数名 decay 为指针
    std::cout << f(4, 5) << '\n';
    return 0;
}
```

| 写法 | 含义 |
|------|------|
| `int (*fp)(int,int)` | 指向函数的指针 |
| `using F = int(*)(int,int);` | 类型别名更清晰 |
| `std::function<int(int,int)>` | 类型擦除包装（见 §8） |

**Primer 提示**：函数名在表达式中会 decay 为指针，但函数指针**不能**像普通对象一样随意重新绑定 overload 集——重载集需显式 `static_cast`。

---
## 3. 函数对象（Functor）

函数对象是定义了 `operator()` 的类实例。STL 大量算法默认接受函数对象，因为**内联友好**且无 `std::function` 堆分配。

```cpp
#include <iostream>
#include <vector>
#include <algorithm>

struct MultiplyBy {
    int factor;
    explicit MultiplyBy(int f) : factor(f) {}
    int operator()(int x) const { return x * factor; }
};

int main() {
    std::vector<int> v{1, 2, 3, 4};
    std::transform(v.begin(), v.end(), v.begin(), MultiplyBy{10});
    for (int x : v) std::cout << x << ' ';  // 10 20 30 40
    return 0;
}
```

### 3.1 带状态的函数对象

与无状态 lambda 按值捕获类似，函数对象成员保存状态：

```cpp
struct Counter {
    int count = 0;
    void operator()() { ++count; }
    int get() const { return count; }
};
```

### 3.2 标准库函数对象

`<functional>` 提供 `std::plus<>`, `std::less<>`, `std::negate<>` 等，配合算法：

```cpp
#include <functional>
#include <numeric>
std::vector<int> v{1,2,3};
int s = std::accumulate(v.begin(), v.end(), 0, std::plus<>{});
```

| 函数对象 | 用途 |
|----------|------|
| `std::less<T>` | 默认比较 |
| `std::hash<T>` | 无序容器 |
| `ref(w)`/`cref(w)` | 包装引用供算法修改外部变量 |

---
## 4. Lambda 表达式基础

Lambda 是 C++11 起的**匿名函数对象**语法糖，编译器生成唯一 closure 类型。

```cpp
#include <iostream>
#include <vector>
#include <algorithm>

int main() {
    std::vector<int> v{3, 1, 4, 1, 5};
    std::sort(v.begin(), v.end(), [](int a, int b) {
        return a < b;  // 升序
    });
    int threshold = 3;
    auto n = std::count_if(v.begin(), v.end(),
        [threshold](int x) { return x > threshold; });
    std::cout << "count=" << n << '\n';
    return 0;
}
```

语法：`[capture](params) specifiers -> ret { body }`

| 部分 | 说明 |
|------|------|
| `[]` | 捕获列表 |
| `(params)` | 参数，可省略 |
| `mutable` | 允许修改按值捕获副本 |
| `constexpr` | C++17 起 constexpr lambda |
| `-> ret` | 尾置返回类型，通常可省略 |

---
## 5. 捕获列表详解

| 捕获 | 含义 |
|------|------|
| `[]` | 不捕获 |
| `[=]` | 按值捕获所有使用到的自动变量 |
| `[&]` | 按引用捕获所有使用到的自动变量 |
| `[x]` | 按值捕获 x |
| `[&x]` | 按引用捕获 x |
| `[=, &x]` | 默认按值，x 按引用 |
| `[this]` | 捕获当前对象指针（成员访问） |
| `[*this]` | C++17 按值捕获 *this 副本 |

**C++14 初始化捕获（init-capture）**：

```cpp
auto ptr = std::make_unique<int>(42);
auto fn = [p = std::move(ptr)]() { return *p; };
// ptr 已空，资源在 lambda 内
```

### 5.1 生命周期陷阱（必背）

```cpp
std::function<void()> make_bad() {
    int x = 42;
    return [&x]() { std::cout << x; };  // UB：x 已销毁
}
// 修复：return [x]() { ... }; 或 init-capture
```

**Primer 口诀**：返回 lambda 时，**禁止**默认 `[&]` 捕获栈上局部变量。

---
## 6. mutable 与 constexpr lambda

按值捕获的变量在 lambda 体内默认 const。`mutable` 允许修改**捕获副本**：

```cpp
int n = 0;
auto inc = [n]() mutable { return ++n; };
std::cout << inc() << '\n';  // 1
std::cout << n << '\n';        // 0 未变
```

C++17 `constexpr` lambda 可在编译期求值：

```cpp
constexpr auto sq = [](int x) { return x * x; };
static_assert(sq(5) == 25);
```

C++14 起 lambda 可泛型：

```cpp
auto cmp = [](auto a, auto b) { return a < b; };
```

---
## 7. 立即调用 lambda（IIFE）

```cpp
const int value = []() {
    // 复杂初始化逻辑
    return 42;
}();  // 注意末尾 ()
```

用于 const 局部变量的一次性初始化，避免额外函数或 `std::optional` 中间态。

---
## 8. std::function

`std::function<R(Args...)>` 是**类型擦除**的可调用包装器，可存函数指针、lambda、函数对象。

```cpp
#include <functional>
#include <iostream>

int add(int a, int b) { return a + b; }

int main() {
    std::function<int(int,int)> f = add;
    f = [](int a, int b) { return a * b; };
    std::cout << f(3, 4) << '\n';  // 12
    if (!f) { /* 空 function 为 false */ }
    return 0;
}
```

| 优点 | 缺点 |
|------|------|
| 统一类型存不同 callable | 可能堆分配 + 间接调用 |
| 适合回调接口 | 大对象 functor 拷贝进包装 |

**工程建议**：热路径用模板 + 具体 lambda/functor；接口层（如事件总线）用 `std::function`。

---
## 9. std::bind 与 bind_front

```cpp
#include <functional>
#include <iostream>

void print(int a, int b, int c) {
    std::cout << a << b << c << '\n';
}

int main() {
    using namespace std::placeholders;
    auto f = std::bind(print, 1, _2, _1);
    f(30, 20);  // 1 20 30
    return 0;
}
```

C++20 推荐 `std::bind_front` 替代部分 bind 场景：

```cpp
auto g = std::bind_front(print, 100);
```

| 工具 | 场景 |
|------|------|
| lambda | **首选**，可读性高 |
| `bind_front` | 固定前几个参数 |
| `bind` + placeholders | 老代码/回调适配 |

---
## 10. 与 STL 算法协作

```cpp
#include <algorithm>
#include <vector>
#include <string>

std::vector<std::string> names{"Bob", "Alice", "Charlie"};
std::sort(names.begin(), names.end(),
    [](const std::string& a, const std::string& b) {
        return a.size() < b.size();
    });

std::for_each(names.begin(), names.end(),
    [](const std::string& s) { std::cout << s << '\n'; });
```

配合 `<numeric>`：`transform`, `accumulate`, `reduce`(C++17)。

---
### 11.1 练习与辨析 #1

**题目 1**：下列捕获是否安全？说明理由。

```cpp
void exercise_1() {
    int local = 1;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.2 练习与辨析 #2

**题目 2**：下列捕获是否安全？说明理由。

```cpp
void exercise_2() {
    int local = 2;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.3 练习与辨析 #3

**题目 3**：下列捕获是否安全？说明理由。

```cpp
void exercise_3() {
    int local = 3;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.4 练习与辨析 #4

**题目 4**：下列捕获是否安全？说明理由。

```cpp
void exercise_4() {
    int local = 4;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.5 练习与辨析 #5

**题目 5**：下列捕获是否安全？说明理由。

```cpp
void exercise_5() {
    int local = 5;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.6 练习与辨析 #6

**题目 6**：下列捕获是否安全？说明理由。

```cpp
void exercise_6() {
    int local = 6;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.7 练习与辨析 #7

**题目 7**：下列捕获是否安全？说明理由。

```cpp
void exercise_7() {
    int local = 7;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.8 练习与辨析 #8

**题目 8**：下列捕获是否安全？说明理由。

```cpp
void exercise_8() {
    int local = 8;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.9 练习与辨析 #9

**题目 9**：下列捕获是否安全？说明理由。

```cpp
void exercise_9() {
    int local = 9;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.10 练习与辨析 #10

**题目 10**：下列捕获是否安全？说明理由。

```cpp
void exercise_10() {
    int local = 10;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.11 练习与辨析 #11

**题目 11**：下列捕获是否安全？说明理由。

```cpp
void exercise_11() {
    int local = 11;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.12 练习与辨析 #12

**题目 12**：下列捕获是否安全？说明理由。

```cpp
void exercise_12() {
    int local = 12;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.13 练习与辨析 #13

**题目 13**：下列捕获是否安全？说明理由。

```cpp
void exercise_13() {
    int local = 13;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.14 练习与辨析 #14

**题目 14**：下列捕获是否安全？说明理由。

```cpp
void exercise_14() {
    int local = 14;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.15 练习与辨析 #15

**题目 15**：下列捕获是否安全？说明理由。

```cpp
void exercise_15() {
    int local = 15;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.16 练习与辨析 #16

**题目 16**：下列捕获是否安全？说明理由。

```cpp
void exercise_16() {
    int local = 16;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.17 练习与辨析 #17

**题目 17**：下列捕获是否安全？说明理由。

```cpp
void exercise_17() {
    int local = 17;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.18 练习与辨析 #18

**题目 18**：下列捕获是否安全？说明理由。

```cpp
void exercise_18() {
    int local = 18;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.19 练习与辨析 #19

**题目 19**：下列捕获是否安全？说明理由。

```cpp
void exercise_19() {
    int local = 19;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.20 练习与辨析 #20

**题目 20**：下列捕获是否安全？说明理由。

```cpp
void exercise_20() {
    int local = 20;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.21 练习与辨析 #21

**题目 21**：下列捕获是否安全？说明理由。

```cpp
void exercise_21() {
    int local = 21;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.22 练习与辨析 #22

**题目 22**：下列捕获是否安全？说明理由。

```cpp
void exercise_22() {
    int local = 22;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.23 练习与辨析 #23

**题目 23**：下列捕获是否安全？说明理由。

```cpp
void exercise_23() {
    int local = 23;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.24 练习与辨析 #24

**题目 24**：下列捕获是否安全？说明理由。

```cpp
void exercise_24() {
    int local = 24;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>

### 11.25 练习与辨析 #25

**题目 25**：下列捕获是否安全？说明理由。

```cpp
void exercise_25() {
    int local = 25;
    auto a = [local]() { return local; };
    auto b = [&local]() { return local; };
    // 若将 b 存储到 std::function 并跨作用域调用？
}
```

<details>
<summary>参考思路</summary>

- `a` 按值捕获，跨作用域安全。
- `b` 按引用捕获，仅在 `local` 存活期间安全。
- 返回 `b` 会导致悬空引用（与 §5.1 同型错误）。

</details>


## 12. 常见错误速查

| # | 错误 | 正确做法 |
|---|------|----------|
| 1 | 返回 `[&]` lambda | 按值或 init-capture |
| 2 | 空 `std::function` 直接调用 | 先 `if (f)` |
| 3 | 大 functor 塞进 function 热路径 | 模板参数传递 |
| 4 | `bind` 占位符顺序搞反 | 改 lambda |
| 5 | 捕获 `this` 对象已销毁 | `weak_ptr` 或值捕获必要数据 |
| 6 | 误以为 `mutable` 改外部变量 | 只改捕获副本 |
| 7 | 递归 lambda 未命名 | C++14 用 `auto self = [&](auto&&...){...};` 模式 |
| 8 | `std::function` 与函数指针签名不匹配 | 检查 const、 noexcept |

## 13. FAQ

**Q：lambda 和函数对象性能？**  
优化后通常相同；lambda 即匿名 closure 类型。

**Q：何时不用 std::function？**  
模板回调、性能敏感循环、嵌入式禁止堆分配。

**Q：[=] 还推荐吗？**  
C++20 起更推荐显式列出捕获变量，避免隐式拷贝大对象。

**Q：generic lambda 与模板函数对象？**  
C++14 generic lambda 的 `auto` 参数即模板，简化 functor 写法。

## 14. 闭卷自测

1. 函数对象相对函数指针的两项优势？
2. `[=]` 与 `[&]` 区别？返回 lambda 时为何慎用 `[&]`？
3. init-capture 解决什么问题？
4. `mutable` 修改的是谁？
5. `std::function` 空状态如何判断？
6. `std::bind` 占位符 `_1` 含义？
7. 为何 Asio 回调常用 `std::function`？
8. C++14 generic lambda 语法？
9. `[this]` 与 `[*this]` 区别（C++17）？
10. 47 章与 05 章 lambda 小节关系？

<details>
<summary>自测参考答案</summary>

1. 可带状态；易内联；可模板化 operator()。
2. 按值拷贝 vs 按引用别名；返回时引用捕获栈变量悬空。
3. 移动 unique_ptr 进 lambda；自定义捕获名。
4. 按值捕获的**副本**，不影响外部。
5. `if (!f)` 或 `f == nullptr`（对 function）。
6. 调用时第 1 个实参位置。
7. 统一存储不同 callable 签名的回调。
8. `[](auto x, auto y){ ... }`。
9. `[this]` 捕获指针；`[*this]` 按值捕获对象副本（拷贝构造）。
10. 05 入门；47 系统讲 callable 全家桶与工程陷阱。

</details>

## 15. 费曼检验

3 分钟讲「lambda 是什么」——_compiler 把 lambda 变成带 operator() 的小类，捕获就是类的成员_。

## 16. 术语三件套

**术语（Callable）**：可用 `()` 调用的实体集合。

**生活类比**：统一规格的**电源插头**——函数、lambda、function 都能插进 STL 算法的「插座」。

**为什么重要**：现代 C++ 默认用 lambda 写回调，替代函数指针 + void* 的 C 风格。

---

## 下一章预告

48 章讲宏、头文件守卫、翻译单元与链接器——理解 lambda 生成的符号如何进入 `.o` 文件。

---

*下一章：48 编译预处理与链接原理*

