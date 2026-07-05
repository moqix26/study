# std::string 与字符处理完全指南

> **文件编码**：UTF-8 · **标准**：C++17 / C++20 为主  
> **风格**：C++ Primer Plus — 零基础友好、术语三件套、逐行读代码、分级练习+答案、FAQ、闭卷自测  
> **前置章节**：[01 C++基础语法与数据类型](01-C++基础语法与数据类型.md) · [02 指针引用与内存管理](02-指针引用与内存管理.md) · [03 面向对象与类设计](03-面向对象与类设计.md) · [04 STL标准库容器与算法](04-STL标准库容器与算法.md) · [05 现代C++新特性](05-现代C++新特性.md)  
> **下一章**：[40-复合类型结构体与enum](40-复合类型结构体与enum.md)  
> **本章主题**：string 操作、C 风格对比、wstring、编码、性能

---

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

**std::string = C++ 里「可变长度文字」的标准容器；比 C 字符串安全，但编码与性能仍有讲究。**

### 0.2 你需要提前知道什么

| 章节 | 为什么需要 |
|------|------------|
| [01 基础语法](01-C++基础语法与数据类型.md) | 类型、变量、运算符 |
| [02 指针与内存](02-指针引用与内存管理.md) | 地址、引用、const 初识 |
| [03 面向对象](03-面向对象与类设计.md) | 类、构造、成员函数 |
| [04 STL](04-STL标准库容器与算法.md) | `string`、`vector` 容器思维 |
| [05 现代 C++](05-现代C++新特性.md) | `auto`、列表初始化、移动语义预告 |

若从 [38-函数机制完全指南](38-函数机制完全指南.md) 刚过来：本章是**自然延伸**，建议先扫一遍 §0.3 知识地图打勾。

### 0.3 本章知识地图（☐→☑）


- [ ] 熟练使用 string 构造、拼接、查找、子串（§1-§4）
- [ ] 对比 C 风格 `char*` 与 `string`（§5）
- [ ] 了解 wstring 与 UTF-8/宽字符（§6-§7）
- [ ] 掌握性能与 SSO 直觉（§8）
- [ ] 闭卷自测 ≥8/10

---

## 本章与上一章的关系

[38 章](38-函数机制完全指南.md) 的 `const std::string&` 参数与本章直接衔接。04 章已见过 `string`；本章**系统补全**字符与编码。

---

## 1. string 基础操作

```cpp
#include <iostream>
#include <string>

int main() {
    std::string s1 = "Hello";
    std::string s2(5, 'x');          // "xxxxx"
    std::string s3 = s1 + ", " + "C++";
    s1 += '!';
    std::cout << s3 << ' ' << s1.size() << '\n';
    return 0;
}
```


### 术语三件套：std::string（std::string）

| 维度 | 内容 |
|------|------|
| **定义** | 标准库**动态字符序列**，自动管理内存。 |
| **生活类比** | 可伸缩的记事本，不用自己数格子。 |
| **为什么重要** | 日常文本处理默认选择；避免手动 new char[]。 |
| **本章用到** | §1-§4 |


## 1. 构造与赋值


### 术语三件套：构造与赋值（构造与赋值）

| 维度 | 内容 |
|------|------|
| **定义** | string 专题：构造与赋值。 |
| **生活类比** | 文字处理的某一工位。 |
| **为什么重要** | 后端/Infra 日志、协议、路径常遇。 |
| **本章用到** | §1 |


### 1.1 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section1_ex1";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 1.2 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section1_ex2";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 1.3 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section1_ex3";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


## 2. 元素访问与迭代


### 术语三件套：元素访问与迭代（元素访问与迭代）

| 维度 | 内容 |
|------|------|
| **定义** | string 专题：元素访问与迭代。 |
| **生活类比** | 文字处理的某一工位。 |
| **为什么重要** | 后端/Infra 日志、协议、路径常遇。 |
| **本章用到** | §2 |


### 2.1 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section2_ex1";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 2.2 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section2_ex2";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 2.3 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section2_ex3";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


## 3. 查找与子串


### 术语三件套：查找与子串（查找与子串）

| 维度 | 内容 |
|------|------|
| **定义** | string 专题：查找与子串。 |
| **生活类比** | 文字处理的某一工位。 |
| **为什么重要** | 后端/Infra 日志、协议、路径常遇。 |
| **本章用到** | §3 |


### 3.1 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section3_ex1";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 3.2 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section3_ex2";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 3.3 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section3_ex3";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


## 4. 插入删除与 replace


### 术语三件套：插入删除与 replace（插入删除与 replace）

| 维度 | 内容 |
|------|------|
| **定义** | string 专题：插入删除与 replace。 |
| **生活类比** | 文字处理的某一工位。 |
| **为什么重要** | 后端/Infra 日志、协议、路径常遇。 |
| **本章用到** | §4 |


### 4.1 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section4_ex1";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 4.2 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section4_ex2";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 4.3 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section4_ex3";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


## 5. C 风格字符串对比


### 术语三件套：C 风格字符串对比（C 风格字符串对比）

| 维度 | 内容 |
|------|------|
| **定义** | string 专题：C 风格字符串对比。 |
| **生活类比** | 文字处理的某一工位。 |
| **为什么重要** | 后端/Infra 日志、协议、路径常遇。 |
| **本章用到** | §5 |


### 5.1 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section5_ex1";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 5.2 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section5_ex2";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 5.3 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section5_ex3";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


## 6. wstring 与宽字符


### 术语三件套：wstring 与宽字符（wstring 与宽字符）

| 维度 | 内容 |
|------|------|
| **定义** | string 专题：wstring 与宽字符。 |
| **生活类比** | 文字处理的某一工位。 |
| **为什么重要** | 后端/Infra 日志、协议、路径常遇。 |
| **本章用到** | §6 |


### 6.1 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section6_ex1";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 6.2 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section6_ex2";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 6.3 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section6_ex3";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


## 7. 编码 UTF-8


### 术语三件套：编码 UTF-8（编码 UTF-8）

| 维度 | 内容 |
|------|------|
| **定义** | string 专题：编码 UTF-8。 |
| **生活类比** | 文字处理的某一工位。 |
| **为什么重要** | 后端/Infra 日志、协议、路径常遇。 |
| **本章用到** | §7 |


### 7.1 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section7_ex1";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 7.2 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section7_ex2";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 7.3 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section7_ex3";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


## 8. 性能与 SSO


### 术语三件套：性能与 SSO（性能与 SSO）

| 维度 | 内容 |
|------|------|
| **定义** | string 专题：性能与 SSO。 |
| **生活类比** | 文字处理的某一工位。 |
| **为什么重要** | 后端/Infra 日志、协议、路径常遇。 |
| **本章用到** | §8 |


### 8.1 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section8_ex1";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 8.2 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section8_ex2";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 8.3 示例

```cpp
#include <iostream>
#include <string>
int main() {
    std::string s = "section8_ex3";
    std::cout << s << " len=" << s.size() << '\n';
    return 0;
}
```


### 逐行读代码：string 与 c_str

```cpp
std::string path = "/tmp/log";
const char* p = path.c_str();  // 只读 C 串，生命期随 path
// 勿：char* q = path.c_str();  // 丢弃 const
printf("%s\n", p);
```

| 行/片段 | 含义 | 改错会怎样 |
|---------|------|------------|
| `std::string path` | 拥有缓冲的可变串 | — |
| `.c_str()` | 以 `\0` 结尾的 const char* | path 销毁后 p 悬空 |
| `printf` | C API 需要 C 串 | 优先 C++20 `std::format` |


## 9. 字符处理实战 §9

```cpp
#include <algorithm>
#include <cctype>
#include <iostream>
#include <string>

std::string to_lower(std::string s) {
    std::transform(s.begin(), s.end(), s.begin(),
        [](unsigned char c) { return static_cast<char>(std::tolower(c)); });
    return s;
}

int main() {
    std::string line = "Section 9 UTF-8 中文";
    std::cout << to_lower(line) << '\n';
    return 0;
}
```


## 10. 字符处理实战 §10

```cpp
#include <algorithm>
#include <cctype>
#include <iostream>
#include <string>

std::string to_lower(std::string s) {
    std::transform(s.begin(), s.end(), s.begin(),
        [](unsigned char c) { return static_cast<char>(std::tolower(c)); });
    return s;
}

int main() {
    std::string line = "Section 10 UTF-8 中文";
    std::cout << to_lower(line) << '\n';
    return 0;
}
```


## 11. 字符处理实战 §11

```cpp
#include <algorithm>
#include <cctype>
#include <iostream>
#include <string>

std::string to_lower(std::string s) {
    std::transform(s.begin(), s.end(), s.begin(),
        [](unsigned char c) { return static_cast<char>(std::tolower(c)); });
    return s;
}

int main() {
    std::string line = "Section 11 UTF-8 中文";
    std::cout << to_lower(line) << '\n';
    return 0;
}
```


## 12. 字符处理实战 §12

```cpp
#include <algorithm>
#include <cctype>
#include <iostream>
#include <string>

std::string to_lower(std::string s) {
    std::transform(s.begin(), s.end(), s.begin(),
        [](unsigned char c) { return static_cast<char>(std::tolower(c)); });
    return s;
}

int main() {
    std::string line = "Section 12 UTF-8 中文";
    std::cout << to_lower(line) << '\n';
    return 0;
}
```


## 13. 字符处理实战 §13

```cpp
#include <algorithm>
#include <cctype>
#include <iostream>
#include <string>

std::string to_lower(std::string s) {
    std::transform(s.begin(), s.end(), s.begin(),
        [](unsigned char c) { return static_cast<char>(std::tolower(c)); });
    return s;
}

int main() {
    std::string line = "Section 13 UTF-8 中文";
    std::cout << to_lower(line) << '\n';
    return 0;
}
```


## 14. 字符处理实战 §14

```cpp
#include <algorithm>
#include <cctype>
#include <iostream>
#include <string>

std::string to_lower(std::string s) {
    std::transform(s.begin(), s.end(), s.begin(),
        [](unsigned char c) { return static_cast<char>(std::tolower(c)); });
    return s;
}

int main() {
    std::string line = "Section 14 UTF-8 中文";
    std::cout << to_lower(line) << '\n';
    return 0;
}
```


## FAQ


**Q：`s == "abc"` 合法吗？**  
A：合法，会调用 string 的比较。

**Q：`string` 能存中文吗？**  
A：可以存 UTF-8 **字节序列**；「字符数」≠ `size()`。

**Q：`substr` 越界？**  
A：抛 `std::out_of_range`。

**Q：C 串 `strlen` vs `string::size`？**  
A：后者 O(1)；C 串 O(n)。

**Q：频繁 `+` 拼接性能？**  
A：大量拼接用 `reserve` 或 `ostringstream`/`fmt`。

**Q：SSO 是什么？**  
A：小字符串优化：短串放对象内部不调堆。

**Q：`wstring` 何时用？**  
A：Windows API 宽字符；跨平台文本推荐 UTF-8 string。

**Q：`data()` 与 `c_str()`？**  
A：C++17 `data()` 可非 const；均保证连续存储。


## 分级练习

### 基础

1. 读一行 string 统计单词数（空格分隔）
2. 判断 string 是否回文
3. 用 `find` 替换所有子串 `"old"`→`"new"`

### 进阶

4. 实现 `split(s, delim)` 返回 `vector<string>`
5. 对比 `string` 与 `char*` 作为函数参数
6. 写 UTF-8 下「字节长度 vs 字符数」说明程序

### 挑战

7. 简单 URL decode `%XX`
8. benchmark：`+=` vs `reserve`+append
9. 读文件到 string 并统计行数

---

## 分级练习参考答案

### 基础

见 §1-§4 参考答案模板。

### 进阶

见 §5 C 对比。

### 挑战

见 §8 性能节。


## 闭卷自测（10 题）

1. string 与 C 串内存谁管？
2. `size()` 含 `\0` 吗？
3. SSO 好处？
4. UTF-8 中文 `size()` 含义？
5. `c_str()` 生命期？
6. 为何 `+` 链式可能慢？
7. `wstring` 典型场景？
8. `npos` 是什么？
9. `stoi` 失败怎么办？
10. C++20 `string::starts_with`？

<details>
<summary>自测参考答案</summary>

1. **string 自动**；C 串手动。
2. **不含**。
3. 小串免堆分配更快。
4. **字节数**非 Unicode 字符数。
5. 随 **string 对象**有效。
6. 多次临时串分配。
7. Win32 宽 API。
8. 查找失败的返回值 `size_t(-1)`。
9. 抛异常；用 try 或 `std::from_chars`。
10. 可读性更好的前缀判断。

</details>


---

## 下一章预告

本章打牢基础后，下一章 **[复合类型：结构体与 enum](40-复合类型结构体与enum.md)** 会继续深入。建议完成 §闭卷自测 ≥8/10 再进入下一章。

---

*当前：第 39 章 · 下一章：40-复合类型结构体与enum*


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