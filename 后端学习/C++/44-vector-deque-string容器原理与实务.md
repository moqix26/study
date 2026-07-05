# vector、deque、string 容器原理与实务

> **文件编码**：UTF-8。
> **定位**：C++ Primer Plus 风格深潜——扩容、迭代器、emplace、shrink_to_fit、小对象优化。

> **交叉阅读**：[03 面向对象与类设计](03-面向对象与类设计.md)、[04 STL 标准库容器与算法](04-STL标准库容器与算法.md)、[28 手写 STL 容器面试专题](28-手写STL容器面试专题.md)。

> **章节链**：[43 继承与多态模型完全指南](43-继承与多态模型完全指南.md) → **本章（44）** → [45 关联容器与哈希容器完全指南](45-关联容器与哈希容器完全指南.md)

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**vector/deque/string** = C++ 最常用的三种序列容器——懂扩容、SSO、失效规则才能写出高性能且正确的代码。

### 0.2 你需要提前知道什么

- [04 章](04-STL标准库容器与算法.md) 基本 API
- [28 章](28-手写STL容器面试专题.md) 手写 vector/string 骨架
- [03 章](03-面向对象与类设计.md) 三五法则（元素类型）

### 0.3 本章知识地图（☐→☑）

- [ ] 解释 vector 三指针/扩容均摊 O(1)
- [ ] 说清 deque 分段与迭代器
- [ ] 理解 string SSO 与 C++11 起的小对象优化
- [ ] 正确使用 emplace / shrink_to_fit
- [ ] 列迭代器失效场景
- [ ] 闭卷自测 ≥8/10

---

## 本章与上一章的关系

[43 章](43-继承与多态模型完全指南.md) 讨论对象层次；本章回到 **值语义容器**——[04 章](04-STL标准库容器与算法.md) 会用，[28 章](28-手写STL容器面试专题.md) 会造，本章 **贯通原理与实务**。

---

## 1. 这份文档学什么

- `vector` 内存模型、扩容因子、强异常保证
- `deque` 分段连续、头尾 O(1)
- `string` SSO、COW 历史与现状
- `emplace_back` vs `push_back`
- `shrink_to_fit` 请求释放多余 capacity
- 迭代器失效完整表

---

## 2. vector 原理

### 2.1 典型实现

三个指针：`begin`, `end`, `cap`（或 `data`, `size`, `capacity`）。

```cpp
std::vector<int> v;
v.reserve(100);
v.push_back(1);
std::cout << "size=" << v.size() << " cap=" << v.capacity() << '\n';
// 扩容通常 1.5x 或 2x（实现定义）
```


### 2.2 扩容与均摊分析

插入 n 个元素，倍乘扩容总拷贝 O(n)，单次 **均摊 O(1)**。[28 章](28-手写STL容器面试专题.md) 手撕 `reallocate`。

### 2.3 迭代器失效

| 操作 | 失效 |
|------|------|
| `push_back`/`emplace_back` 超 cap | 全部 |
| `insert` 中间 | insert 及之后 |
| `erase` | erase 及之后 |
| `reserve` 仅当 reallocate | 全部 |

---

## 3. deque 原理

中央 **map** 指向多个固定大小 buffer；随机访问 O(1)，头尾插入 O(1)，中间插入 O(n)。

```cpp
std::deque<int> dq;
dq.push_front(1);
dq.push_back(2);
// 迭代器比 vector 复杂：跨 chunk
```


[28 章](28-手写STL容器面试专题.md) 面试常问结构；不要求完整手撕。

---

## 4. string 与 SSO

**小字符串优化**：短串存对象内部 buffer，无堆分配。长度阈值实现定义（常见 15～22 字节）。

```cpp
std::string s = "hello";
// 可能无 malloc；长串堆分配
std::string long_s(1000, 'x');
```


C++11 后 **禁止 COW** 字符串（多线程安全）。

---

## 5. emplace 系列

```cpp
struct Foo { Foo(int, double); };
std::vector<Foo> v;
v.emplace_back(1, 2.0);  // 原地构造
// v.push_back(Foo(1, 2.0));  // 多一次移动/拷贝
```


[05 章](05-现代C++新特性.md) 完美转发基础。

---

## 6. shrink_to_fit

```cpp
v.clear();
v.shrink_to_fit();  // 非强制；C++11 请求释放 capacity
```

不保证一定缩小；`swap` 技巧：`vector<T>(v).swap(v)`  C++11 前惯用法。

---

## 7. 性能实务

- 预知大小 **`reserve`**
- 避免在循环中 **`push_back` 未 reserve**
- 大对象用 **`emplace`**
- 删除用 **`erase-remove`** 见 [46 章](46-迭代器分类与算法库完全指南.md)

---

## 8. 与 [03 章](03-面向对象与类设计.md) 元素类型

容器存 **可拷贝/可移动** 类型；含指针成员需 Rule of Five。[28 章](28-手写STL容器面试专题.md) 详解。


## 9.1 vector 实务案例 1

**案例**：构建索引 1 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_1() {
    std::vector<std::string> lines;
    lines.reserve(10);
    for (int k = 0; k < 5; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 1**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 1**：短键 `1` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.2 vector 实务案例 2

**案例**：构建索引 2 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_2() {
    std::vector<std::string> lines;
    lines.reserve(20);
    for (int k = 0; k < 10; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 2**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 2**：短键 `2` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.3 vector 实务案例 3

**案例**：构建索引 3 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_3() {
    std::vector<std::string> lines;
    lines.reserve(30);
    for (int k = 0; k < 15; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 3**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 3**：短键 `3` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.4 vector 实务案例 4

**案例**：构建索引 4 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_4() {
    std::vector<std::string> lines;
    lines.reserve(40);
    for (int k = 0; k < 20; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 4**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 4**：短键 `4` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.5 vector 实务案例 5

**案例**：构建索引 5 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_5() {
    std::vector<std::string> lines;
    lines.reserve(50);
    for (int k = 0; k < 25; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 5**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 5**：短键 `5` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.6 vector 实务案例 6

**案例**：构建索引 6 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_6() {
    std::vector<std::string> lines;
    lines.reserve(60);
    for (int k = 0; k < 30; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 6**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 6**：短键 `6` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.7 vector 实务案例 7

**案例**：构建索引 7 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_7() {
    std::vector<std::string> lines;
    lines.reserve(70);
    for (int k = 0; k < 35; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 7**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 7**：短键 `7` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.8 vector 实务案例 8

**案例**：构建索引 8 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_8() {
    std::vector<std::string> lines;
    lines.reserve(80);
    for (int k = 0; k < 40; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 8**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 8**：短键 `8` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.9 vector 实务案例 9

**案例**：构建索引 9 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_9() {
    std::vector<std::string> lines;
    lines.reserve(90);
    for (int k = 0; k < 45; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 9**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 9**：短键 `9` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.10 vector 实务案例 10

**案例**：构建索引 10 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_10() {
    std::vector<std::string> lines;
    lines.reserve(100);
    for (int k = 0; k < 50; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 10**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 10**：短键 `10` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.11 vector 实务案例 11

**案例**：构建索引 11 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_11() {
    std::vector<std::string> lines;
    lines.reserve(110);
    for (int k = 0; k < 55; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 11**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 11**：短键 `11` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.12 vector 实务案例 12

**案例**：构建索引 12 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_12() {
    std::vector<std::string> lines;
    lines.reserve(120);
    for (int k = 0; k < 60; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 12**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 12**：短键 `12` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.13 vector 实务案例 13

**案例**：构建索引 13 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_13() {
    std::vector<std::string> lines;
    lines.reserve(130);
    for (int k = 0; k < 65; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 13**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 13**：短键 `13` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.14 vector 实务案例 14

**案例**：构建索引 14 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_14() {
    std::vector<std::string> lines;
    lines.reserve(140);
    for (int k = 0; k < 70; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 14**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 14**：短键 `14` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.15 vector 实务案例 15

**案例**：构建索引 15 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_15() {
    std::vector<std::string> lines;
    lines.reserve(150);
    for (int k = 0; k < 75; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 15**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 15**：短键 `15` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.16 vector 实务案例 16

**案例**：构建索引 16 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_16() {
    std::vector<std::string> lines;
    lines.reserve(160);
    for (int k = 0; k < 80; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 16**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 16**：短键 `16` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.17 vector 实务案例 17

**案例**：构建索引 17 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_17() {
    std::vector<std::string> lines;
    lines.reserve(170);
    for (int k = 0; k < 85; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 17**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 17**：短键 `17` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.18 vector 实务案例 18

**案例**：构建索引 18 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_18() {
    std::vector<std::string> lines;
    lines.reserve(180);
    for (int k = 0; k < 90; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 18**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 18**：短键 `18` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.19 vector 实务案例 19

**案例**：构建索引 19 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_19() {
    std::vector<std::string> lines;
    lines.reserve(190);
    for (int k = 0; k < 95; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 19**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 19**：短键 `19` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.20 vector 实务案例 20

**案例**：构建索引 20 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_20() {
    std::vector<std::string> lines;
    lines.reserve(200);
    for (int k = 0; k < 100; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 20**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 20**：短键 `20` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.21 vector 实务案例 21

**案例**：构建索引 21 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_21() {
    std::vector<std::string> lines;
    lines.reserve(210);
    for (int k = 0; k < 105; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 21**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 21**：短键 `21` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.22 vector 实务案例 22

**案例**：构建索引 22 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_22() {
    std::vector<std::string> lines;
    lines.reserve(220);
    for (int k = 0; k < 110; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 22**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 22**：短键 `22` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.23 vector 实务案例 23

**案例**：构建索引 23 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_23() {
    std::vector<std::string> lines;
    lines.reserve(230);
    for (int k = 0; k < 115; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 23**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 23**：短键 `23` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.24 vector 实务案例 24

**案例**：构建索引 24 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_24() {
    std::vector<std::string> lines;
    lines.reserve(240);
    for (int k = 0; k < 120; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 24**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 24**：短键 `24` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.25 vector 实务案例 25

**案例**：构建索引 25 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_25() {
    std::vector<std::string> lines;
    lines.reserve(250);
    for (int k = 0; k < 125; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 25**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 25**：短键 `25` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.26 vector 实务案例 26

**案例**：构建索引 26 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_26() {
    std::vector<std::string> lines;
    lines.reserve(260);
    for (int k = 0; k < 130; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 26**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 26**：短键 `26` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.27 vector 实务案例 27

**案例**：构建索引 27 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_27() {
    std::vector<std::string> lines;
    lines.reserve(270);
    for (int k = 0; k < 135; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 27**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 27**：短键 `27` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.28 vector 实务案例 28

**案例**：构建索引 28 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_28() {
    std::vector<std::string> lines;
    lines.reserve(280);
    for (int k = 0; k < 140; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 28**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 28**：短键 `28` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.29 vector 实务案例 29

**案例**：构建索引 29 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_29() {
    std::vector<std::string> lines;
    lines.reserve(290);
    for (int k = 0; k < 145; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 29**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 29**：短键 `29` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.30 vector 实务案例 30

**案例**：构建索引 30 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_30() {
    std::vector<std::string> lines;
    lines.reserve(300);
    for (int k = 0; k < 150; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 30**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 30**：短键 `30` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.31 vector 实务案例 31

**案例**：构建索引 31 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_31() {
    std::vector<std::string> lines;
    lines.reserve(310);
    for (int k = 0; k < 155; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 31**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 31**：短键 `31` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 9.32 vector 实务案例 32

**案例**：构建索引 32 的缓冲区管理。

```cpp
#include <vector>
#include <string>
#include <iostream>

void case_32() {
    std::vector<std::string> lines;
    lines.reserve(320);
    for (int k = 0; k < 160; ++k)
        lines.emplace_back("line-" + std::to_string(k));
    std::cout << "cap=" << lines.capacity() << " size=" << lines.size() << '\n';
    lines.shrink_to_fit();
}
```

**要点**：`reserve` 避免反复扩容；[04 章](04-STL标准库容器与算法.md) 词频统计同样模式。

**deque 对比 32**：需头尾插入队列时用 `deque`；仅需尾插用 `vector` cache 更友好。

**string SSO 32**：短键 `32` 字符内联；[45 章](45-关联容器与哈希容器完全指南.md) map key 长度影响内存。


## 附录 A：高频面试问答

### A.1 vector 扩容因子？

实现定义，常见 2 或 1.5；GCC libstdc++ 2，部分 MSVC 1.5。

### A.2 deque 为何不用整块连续？

头尾 O(1) 无需 reallocate 全表；牺牲局部性。

### A.3 SSO 阈值？

实现定义；可用 `sizeof(string)` 与堆分配探测（仅测试）。

### A.4 shrink_to_fit 保证？

不保证；QoI 实现可能忽略。

### A.5 emplace 何时必须用？

构造昂贵或不可拷贝时；简单类型差别小。

### A.6 vector<bool> 特化？

代理引用，非真正容器；别传 `vector<bool>&`。

### A.7 迭代器与指针？

vector 连续存储时 `&v[0]` 曾不安全（空 vector）；用 `data()`。

### A.8 与 28 章关系？

28 手撕；44 对照标准库行为与工程调优。

## 22. 闭卷自测

1. vector 扩容均摊复杂度？
2. push_back 何时迭代器失效？
3. deque 头插复杂度？
4. SSO 目的？
5. emplace_back 优势？
6. shrink_to_fit 语义？
7. string COW 现状？
8. 空 vector 能取首元素地址吗？
9. 选 deque 而非 vector？
10. 与 03/04/28 关联？

<details>
<summary>自测参考答案</summary>

1. **O(1) 均摊**。
2. reallocate 时 **全部失效**。
3. **O(1)**。
4. 短串 **避免堆分配**。
5. **原地构造**，少临时对象。
6. **请求** 释放 capacity，非强制。
7. C++11 后标准 string **无 COW**。
8. C++11 前 UB；用 data() 或判空。
9. 需 **头尾** 频繁插入且不需整块连续。
10. **03** 元素三五法则；**04** API；**28** 实现。

</details>

---

## 下一章预告

[45 关联容器与哈希容器完全指南](45-关联容器与哈希容器完全指南.md)：map/set/multimap、unordered 系列、自定义比较与哈希、性能权衡。

---

*第 45-关联容器与哈希容器完全指南.md 章 · 建议对照 [04 STL](04-STL标准库容器与算法.md) 与 [28 手写 STL](28-手写STL容器面试专题.md) 复习*
