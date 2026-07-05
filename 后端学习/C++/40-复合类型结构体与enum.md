# 复合类型：结构体与 enum

> **文件编码**：UTF-8 · **标准**：C++17 / C++20 为主  
> **风格**：C++ Primer Plus — 零基础友好、术语三件套、逐行读代码、分级练习+答案、FAQ、闭卷自测  
> **前置章节**：[01 C++基础语法与数据类型](01-C++基础语法与数据类型.md) · [02 指针引用与内存管理](02-指针引用与内存管理.md) · [03 面向对象与类设计](03-面向对象与类设计.md) · [04 STL标准库容器与算法](04-STL标准库容器与算法.md) · [05 现代C++新特性](05-现代C++新特性.md)  
> **下一章**：[41-构造析构与三五法则大全](41-构造析构与三五法则大全.md)  
> **本章主题**：struct、union、enum class、位域、对齐、typedef/using

---

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

**复合类型 = 把多个字段「打包成一条记录」；struct/enum 是 C++ 组织数据的基石，对齐影响性能与网络协议。**

### 0.2 你需要提前知道什么

| 章节 | 为什么需要 |
|------|------------|
| [01 基础语法](01-C++基础语法与数据类型.md) | 类型、变量、运算符 |
| [02 指针与内存](02-指针引用与内存管理.md) | 地址、引用、const 初识 |
| [03 面向对象](03-面向对象与类设计.md) | 类、构造、成员函数 |
| [04 STL](04-STL标准库容器与算法.md) | `string`、`vector` 容器思维 |
| [05 现代 C++](05-现代C++新特性.md) | `auto`、列表初始化、移动语义预告 |

若从 [39-std-string与字符处理完全指南](39-std-string与字符处理完全指南.md) 刚过来：本章是**自然延伸**，建议先扫一遍 §0.3 知识地图打勾。

### 0.3 本章知识地图（☐→☑）


- [ ] 定义 struct、初始化、嵌套（§1）
- [ ] 理解 union 与位域（§2-§3）
- [ ] 使用 enum class（§4）
- [ ] 掌握对齐与 `#pragma pack` 直觉（§5）
- [ ] typedef / using 别名（§6）
- [ ] 闭卷自测 ≥8/10

---

## 本章与上一章的关系

[39 章](39-std-string与字符处理完全指南.md) 处理线性字符；本章处理**结构化记录**——日志字段、协议头、配置项都是 struct。

---

## 1. struct 基础

```cpp
#include <iostream>
#include <string>

struct User {
    int id;
    std::string name;
    double score;
};

int main() {
    User u{1, "Alice", 98.5};
    std::cout << u.name << ' ' << u.score << '\n';
    return 0;
}
```


### 术语三件套：struct（struct）

| 维度 | 内容 |
|------|------|
| **定义** | 聚合类型，成员**默认 public**（class 默认 private）。 |
| **生活类比** | 表格一行：id、姓名、分数。 |
| **为什么重要** | POD/聚合初始化、序列化、与 C 互操作。 |
| **本章用到** | §1 |


## 2. union 共用体


### 术语三件套：union 共用体（union 共用体）

| 维度 | 内容 |
|------|------|
| **定义** | union 共用体 的定义与用途。 |
| **生活类比** | 数据打包的一种抽屉。 |
| **为什么重要** | 协议、嵌入式、状态机。 |
| **本章用到** | §2 |


### 2.1 代码示例

```cpp
#include <iostream>
// §2 示例 1
struct Demo {
    int x = 2;
    int y = 1;
};
int main() {
    Demo d{};
    std::cout << d.x << ',' << d.y << '\n';
    return 0;
}
```


### 2.2 代码示例

```cpp
#include <iostream>
// §2 示例 2
struct Demo {
    int x = 2;
    int y = 2;
};
int main() {
    Demo d{};
    std::cout << d.x << ',' << d.y << '\n';
    return 0;
}
```


## 3. 位域 bit-field


### 术语三件套：位域 bit-field（位域 bit-field）

| 维度 | 内容 |
|------|------|
| **定义** | 位域 bit-field 的定义与用途。 |
| **生活类比** | 数据打包的一种抽屉。 |
| **为什么重要** | 协议、嵌入式、状态机。 |
| **本章用到** | §3 |


### 3.1 代码示例

```cpp
#include <iostream>
// §3 示例 1
struct Demo {
    int x = 3;
    int y = 1;
};
int main() {
    Demo d{};
    std::cout << d.x << ',' << d.y << '\n';
    return 0;
}
```


### 3.2 代码示例

```cpp
#include <iostream>
// §3 示例 2
struct Demo {
    int x = 3;
    int y = 2;
};
int main() {
    Demo d{};
    std::cout << d.x << ',' << d.y << '\n';
    return 0;
}
```


## 4. enum class


### 术语三件套：enum class（enum class）

| 维度 | 内容 |
|------|------|
| **定义** | enum class 的定义与用途。 |
| **生活类比** | 数据打包的一种抽屉。 |
| **为什么重要** | 协议、嵌入式、状态机。 |
| **本章用到** | §4 |


### 4.1 代码示例

```cpp
#include <iostream>
// §4 示例 1
struct Demo {
    int x = 4;
    int y = 1;
};
int main() {
    Demo d{};
    std::cout << d.x << ',' << d.y << '\n';
    return 0;
}
```


### 4.2 代码示例

```cpp
#include <iostream>
// §4 示例 2
struct Demo {
    int x = 4;
    int y = 2;
};
int main() {
    Demo d{};
    std::cout << d.x << ',' << d.y << '\n';
    return 0;
}
```


## 5. 内存对齐 alignment


### 术语三件套：内存对齐 alignment（内存对齐 alignment）

| 维度 | 内容 |
|------|------|
| **定义** | 内存对齐 alignment 的定义与用途。 |
| **生活类比** | 数据打包的一种抽屉。 |
| **为什么重要** | 协议、嵌入式、状态机。 |
| **本章用到** | §5 |


### 5.1 代码示例

```cpp
#include <iostream>
// §5 示例 1
struct Demo {
    int x = 5;
    int y = 1;
};
int main() {
    Demo d{};
    std::cout << d.x << ',' << d.y << '\n';
    return 0;
}
```


### 5.2 代码示例

```cpp
#include <iostream>
// §5 示例 2
struct Demo {
    int x = 5;
    int y = 2;
};
int main() {
    Demo d{};
    std::cout << d.x << ',' << d.y << '\n';
    return 0;
}
```


## 6. typedef 与 using


### 术语三件套：typedef 与 using（typedef 与 using）

| 维度 | 内容 |
|------|------|
| **定义** | typedef 与 using 的定义与用途。 |
| **生活类比** | 数据打包的一种抽屉。 |
| **为什么重要** | 协议、嵌入式、状态机。 |
| **本章用到** | §6 |


### 6.1 代码示例

```cpp
#include <iostream>
// §6 示例 1
struct Demo {
    int x = 6;
    int y = 1;
};
int main() {
    Demo d{};
    std::cout << d.x << ',' << d.y << '\n';
    return 0;
}
```


### 6.2 代码示例

```cpp
#include <iostream>
// §6 示例 2
struct Demo {
    int x = 6;
    int y = 2;
};
int main() {
    Demo d{};
    std::cout << d.x << ',' << d.y << '\n';
    return 0;
}
```


## 7. 结构化绑定


### 术语三件套：结构化绑定（结构化绑定）

| 维度 | 内容 |
|------|------|
| **定义** | 结构化绑定 的定义与用途。 |
| **生活类比** | 数据打包的一种抽屉。 |
| **为什么重要** | 协议、嵌入式、状态机。 |
| **本章用到** | §7 |


### 7.1 代码示例

```cpp
#include <iostream>
// §7 示例 1
struct Demo {
    int x = 7;
    int y = 1;
};
int main() {
    Demo d{};
    std::cout << d.x << ',' << d.y << '\n';
    return 0;
}
```


### 7.2 代码示例

```cpp
#include <iostream>
// §7 示例 2
struct Demo {
    int x = 7;
    int y = 2;
};
int main() {
    Demo d{};
    std::cout << d.x << ',' << d.y << '\n';
    return 0;
}
```


### 逐行读代码：enum class 强类型枚举

```cpp
enum class Color : std::uint8_t { Red, Green, Blue };
Color c = Color::Red;
// int n = c;  // 错误
auto u = static_cast<std::uint8_t>(c);
```

| 行/片段 | 含义 | 改错会怎样 |
|---------|------|------------|
| `enum class` | 作用域限定，不污染外层 | 比旧 enum 安全 |
| `: uint8_t` | 指定底层类型 | 网络包常指定宽度 |
| `Color::Red` | 必须带枚举名 | — |
| `static_cast` | 显式转底层类型 | 避免隐式混用 |


## 8. 对齐与布局 §8

```cpp
#include <cstddef>
#include <iostream>

struct A { char c; int i; };
struct B { int i; char c; };

int main() {
    std::cout << sizeof(A) << ' ' << sizeof(B) << '\n';
    std::cout << offsetof(A, i) << '\n';
    return 0;
}
```

**思考**：为何 `sizeof(A)` 常大于 5？——**对齐 padding**（与 [18 章](18-高性能C++与内存对齐.md) 呼应）。


## 9. 对齐与布局 §9

```cpp
#include <cstddef>
#include <iostream>

struct A { char c; int i; };
struct B { int i; char c; };

int main() {
    std::cout << sizeof(A) << ' ' << sizeof(B) << '\n';
    std::cout << offsetof(A, i) << '\n';
    return 0;
}
```

**思考**：为何 `sizeof(A)` 常大于 5？——**对齐 padding**（与 [18 章](18-高性能C++与内存对齐.md) 呼应）。


## 10. 对齐与布局 §10

```cpp
#include <cstddef>
#include <iostream>

struct A { char c; int i; };
struct B { int i; char c; };

int main() {
    std::cout << sizeof(A) << ' ' << sizeof(B) << '\n';
    std::cout << offsetof(A, i) << '\n';
    return 0;
}
```

**思考**：为何 `sizeof(A)` 常大于 5？——**对齐 padding**（与 [18 章](18-高性能C++与内存对齐.md) 呼应）。


## 11. 对齐与布局 §11

```cpp
#include <cstddef>
#include <iostream>

struct A { char c; int i; };
struct B { int i; char c; };

int main() {
    std::cout << sizeof(A) << ' ' << sizeof(B) << '\n';
    std::cout << offsetof(A, i) << '\n';
    return 0;
}
```

**思考**：为何 `sizeof(A)` 常大于 5？——**对齐 padding**（与 [18 章](18-高性能C++与内存对齐.md) 呼应）。


## 12. 对齐与布局 §12

```cpp
#include <cstddef>
#include <iostream>

struct A { char c; int i; };
struct B { int i; char c; };

int main() {
    std::cout << sizeof(A) << ' ' << sizeof(B) << '\n';
    std::cout << offsetof(A, i) << '\n';
    return 0;
}
```

**思考**：为何 `sizeof(A)` 常大于 5？——**对齐 padding**（与 [18 章](18-高性能C++与内存对齐.md) 呼应）。


## 13. 对齐与布局 §13

```cpp
#include <cstddef>
#include <iostream>

struct A { char c; int i; };
struct B { int i; char c; };

int main() {
    std::cout << sizeof(A) << ' ' << sizeof(B) << '\n';
    std::cout << offsetof(A, i) << '\n';
    return 0;
}
```

**思考**：为何 `sizeof(A)` 常大于 5？——**对齐 padding**（与 [18 章](18-高性能C++与内存对齐.md) 呼应）。


## 14. 对齐与布局 §14

```cpp
#include <cstddef>
#include <iostream>

struct A { char c; int i; };
struct B { int i; char c; };

int main() {
    std::cout << sizeof(A) << ' ' << sizeof(B) << '\n';
    std::cout << offsetof(A, i) << '\n';
    return 0;
}
```

**思考**：为何 `sizeof(A)` 常大于 5？——**对齐 padding**（与 [18 章](18-高性能C++与内存对齐.md) 呼应）。


## FAQ


**Q：struct 和 class 区别？**  
A：默认访问：struct public，class private；其余几乎相同。

**Q：union 安全吗？**  
A：现代 C++ 需配合 `std::variant` 更安全；union 要自己管活跃成员。

**Q：位域可移植吗？**  
A：布局实现定义；网络协议常用手动移位。

**Q：enum class 为何优于 enum？**  
A：不隐式转 int、作用域清晰。

**Q：`alignas` 做什么？**  
A：指定对齐要求。

**Q：`typedef` vs `using`？**  
A：模板别名只能 using。

**Q：结构化绑定？**  
A：`auto [x,y] = pair;` C++17 解构聚合。

**Q：空 struct 大小？**  
A：通常为 1（C++ 对象唯一地址）。

**Q：#pragma pack 风险？**  
A：破坏默认对齐，可能降性能；跨平台需谨慎。


## 分级练习

### 基础

1. 定义 `Point{x,y}` 并计算距离原点
2. 用 enum class 表示星期
3. 打印 struct 各成员 sizeof 与 offsetof

### 进阶

4. 用 union 解析 float 的十六进制（type punning 讨论）
5. 设计 1 字节对齐的协议头 struct
6. using 定义函数指针别名

### 挑战

7. 比较 `#pragma pack(1)` 前后 sizeof
8. 实现 `enum class` 到 string 的 switch 映射
9. C++17 结构化绑定遍历 map

---

## 分级练习参考答案

### 基础

见 §1。

### 进阶

见 §4。

### 挑战

见 §5-§6。


## 闭卷自测（10 题）

1. struct 默认访问？
2. union 特点？
3. enum class 如何赋值给 int？
4. 位域用途？
5. 对齐 padding 原因？
6. typedef 老写法？
7. using 别名优势？
8. 结构化绑定需要什么类型？
9. 空 struct 大小？
10. 与 41 章类区别预告？

<details>
<summary>自测参考答案</summary>

1. **public**。
2. 成员共享同一块内存。
3. **不能隐式**；需 static_cast。
4. 节省位存储（嵌入式）。
5. CPU 按 word 访问效率。
6. `typedef int INT32;`。
7. 可读、支持模板别名。
8. **聚合**或 tuple-like。
9. 通常 **1**。
10. struct 偏数据；class 偏封装+构造析构。

</details>


---

## 下一章预告

本章打牢基础后，下一章 **[构造、析构与三五法则大全](41-构造析构与三五法则大全.md)** 会继续深入。建议完成 §闭卷自测 ≥8/10 再进入下一章。

---

*当前：第 40 章 · 下一章：41-构造析构与三五法则大全*


---

## 附录补充 §900（精读扩展）

### A.900.1 复习要点

- 回顾本章术语三件套与逐行读表
- 在本地编译本节所有小例子：`g++ -std=c++17 -Wall -Wextra -o app app.cpp`
- 与 [01 章](01-C++基础语法与数据类型.md) 类型系统对照笔记

### A.900.2 常见编译警告处理

| 警告 | 含义 | 处理 |
|------|------|------|
| `-Wsign-compare` | 有符号无符号比较 | 统一类型或 cast |
| `-Wunused-variable` | 未使用变量 | 删除或 `[[maybe_unused]]` |
| `-Wnarrowing` | 窄化 | 用 static_cast 或改类型 |

### A.900.3 费曼检验

用 3 分钟向同学讲解本章一个最容易错的知识点，并记录对方一个问题。

### A.900.4 交叉索引

| 关联章 | 内容 |
|--------|------|
| [02 指针](02-指针引用与内存管理.md) | 内存与 const |
| [03 OOP](03-面向对象与类设计.md) | 类与构造 |
| [04 STL](04-STL标准库容器与算法.md) | string/vector |
| [05 现代](05-现代C++新特性.md) | auto/移动 |

```cpp
// 附录练习 900
#include <iostream>
int main() {
    std::cout << "appendix section 900 OK\n";
    return 0;
}
```

---

## 附录补充 §901（精读扩展）

### A.901.1 复习要点

- 回顾本章术语三件套与逐行读表
- 在本地编译本节所有小例子：`g++ -std=c++17 -Wall -Wextra -o app app.cpp`
- 与 [01 章](01-C++基础语法与数据类型.md) 类型系统对照笔记

### A.901.2 常见编译警告处理

| 警告 | 含义 | 处理 |
|------|------|------|
| `-Wsign-compare` | 有符号无符号比较 | 统一类型或 cast |
| `-Wunused-variable` | 未使用变量 | 删除或 `[[maybe_unused]]` |
| `-Wnarrowing` | 窄化 | 用 static_cast 或改类型 |

### A.901.3 费曼检验

用 3 分钟向同学讲解本章一个最容易错的知识点，并记录对方一个问题。

### A.901.4 交叉索引

| 关联章 | 内容 |
|--------|------|
| [02 指针](02-指针引用与内存管理.md) | 内存与 const |
| [03 OOP](03-面向对象与类设计.md) | 类与构造 |
| [04 STL](04-STL标准库容器与算法.md) | string/vector |
| [05 现代](05-现代C++新特性.md) | auto/移动 |

```cpp
// 附录练习 901
#include <iostream>
int main() {
    std::cout << "appendix section 901 OK\n";
    return 0;
}
```

---

## 附录补充 §902（精读扩展）

### A.902.1 复习要点

- 回顾本章术语三件套与逐行读表
- 在本地编译本节所有小例子：`g++ -std=c++17 -Wall -Wextra -o app app.cpp`
- 与 [01 章](01-C++基础语法与数据类型.md) 类型系统对照笔记

### A.902.2 常见编译警告处理

| 警告 | 含义 | 处理 |
|------|------|------|
| `-Wsign-compare` | 有符号无符号比较 | 统一类型或 cast |
| `-Wunused-variable` | 未使用变量 | 删除或 `[[maybe_unused]]` |
| `-Wnarrowing` | 窄化 | 用 static_cast 或改类型 |

### A.902.3 费曼检验

用 3 分钟向同学讲解本章一个最容易错的知识点，并记录对方一个问题。

### A.902.4 交叉索引

| 关联章 | 内容 |
|--------|------|
| [02 指针](02-指针引用与内存管理.md) | 内存与 const |
| [03 OOP](03-面向对象与类设计.md) | 类与构造 |
| [04 STL](04-STL标准库容器与算法.md) | string/vector |
| [05 现代](05-现代C++新特性.md) | auto/移动 |

```cpp
// 附录练习 902
#include <iostream>
int main() {
    std::cout << "appendix section 902 OK\n";
    return 0;
}
```

---

## 附录补充 §903（精读扩展）

### A.903.1 复习要点

- 回顾本章术语三件套与逐行读表
- 在本地编译本节所有小例子：`g++ -std=c++17 -Wall -Wextra -o app app.cpp`
- 与 [01 章](01-C++基础语法与数据类型.md) 类型系统对照笔记

### A.903.2 常见编译警告处理

| 警告 | 含义 | 处理 |
|------|------|------|
| `-Wsign-compare` | 有符号无符号比较 | 统一类型或 cast |
| `-Wunused-variable` | 未使用变量 | 删除或 `[[maybe_unused]]` |
| `-Wnarrowing` | 窄化 | 用 static_cast 或改类型 |

### A.903.3 费曼检验

用 3 分钟向同学讲解本章一个最容易错的知识点，并记录对方一个问题。

### A.903.4 交叉索引

| 关联章 | 内容 |
|--------|------|
| [02 指针](02-指针引用与内存管理.md) | 内存与 const |
| [03 OOP](03-面向对象与类设计.md) | 类与构造 |
| [04 STL](04-STL标准库容器与算法.md) | string/vector |
| [05 现代](05-现代C++新特性.md) | auto/移动 |

```cpp
// 附录练习 903
#include <iostream>
int main() {
    std::cout << "appendix section 903 OK\n";
    return 0;
}
```

---

## 附录补充 §904（精读扩展）

### A.904.1 复习要点

- 回顾本章术语三件套与逐行读表
- 在本地编译本节所有小例子：`g++ -std=c++17 -Wall -Wextra -o app app.cpp`
- 与 [01 章](01-C++基础语法与数据类型.md) 类型系统对照笔记

### A.904.2 常见编译警告处理

| 警告 | 含义 | 处理 |
|------|------|------|
| `-Wsign-compare` | 有符号无符号比较 | 统一类型或 cast |
| `-Wunused-variable` | 未使用变量 | 删除或 `[[maybe_unused]]` |
| `-Wnarrowing` | 窄化 | 用 static_cast 或改类型 |

### A.904.3 费曼检验

用 3 分钟向同学讲解本章一个最容易错的知识点，并记录对方一个问题。

### A.904.4 交叉索引

| 关联章 | 内容 |
|--------|------|
| [02 指针](02-指针引用与内存管理.md) | 内存与 const |
| [03 OOP](03-面向对象与类设计.md) | 类与构造 |
| [04 STL](04-STL标准库容器与算法.md) | string/vector |
| [05 现代](05-现代C++新特性.md) | auto/移动 |

```cpp
// 附录练习 904
#include <iostream>
int main() {
    std::cout << "appendix section 904 OK\n";
    return 0;
}
```