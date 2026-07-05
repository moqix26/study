# string 类与 STL 入门第一课

> **文件编码**：UTF-8 · **标准**：C++17 为主，C++11/14 兼容说明会标注  
> **风格**：C++ Primer Plus 第 6 版 — 零基础友好、术语三件套、生活类比、最小可编译示例、常见错误、循序渐进  
> **前置章节**：[103 章](103-代码重用包含与多重继承入门.md) · [04 STL](../04-STL标准库容器与算法.md) · [39 string](../39-std-string与字符处理完全指南.md)  
> **章节链**：[103-代码重用包含与多重继承入门](103-代码重用包含与多重继承入门.md) → **本章** → [105-IO输入输出与文件入门](105-IO输入输出与文件入门.md)  
> **本章主题**：vector 入门、string 类常用操作、迭代器初识、algorithm 入门(sort/find/for_each)、auto 与范围 for 配合 STL、避免数组

> **学习建议**：每读完一小节，在本地编译运行示例；遇到「常见错误」小节务必亲手复现一次再改对——这比多看十页更有效。

---

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

**STL = 容器 + 迭代器 + 算法；`vector`/`string` 替代 C 数组，`algorithm` 替代手写循环——现代 C++ 日常编码的主战场。**

### 0.2 你需要提前知道什么

| 章节/知识点 | 为什么需要 |
|-------------|------------|
| [103 章](103-代码重用包含与多重继承入门.md) | 代码重用 |
| [101 章](101-类与动态内存分配入门.md) | 理解 vector 管理内存 |

若你刚学完上一章，建议先扫一遍 §0.3 知识地图，勾选已掌握项再开始正文。

### 0.3 本章知识地图（☐→☑）

- [ ] 使用 vector 增删查改
- [ ] 熟练 string 常用 API
- [ ] 用迭代器遍历容器
- [ ] 调用 sort/find/for_each
- [ ] auto 与范围 for
- [ ] 说明为何少用手动数组

### 0.4 建议学习时长

**4～7 天**；每天 1～2 个小节 + 手写示例 + 章末练习至少完成「基础」层。

### 0.5 三条贯穿类比（本章版）

| 类比 | 对应概念 |
|------|----------|
| **遥控器选台** | 基类指针 + 虚函数 |
| **工具箱与零件** | 组合 has-a vs 继承 is-a |


---

## 本章与上一章的关系

[103 章](103-代码重用包含与多重继承入门.md) 的组合思想在 STL 中达到顶峰：`vector<string>` 是 has-a 套娃，且 Rule of Zero。

## 1. 为何避免 C 风格数组

| C 数组 | vector / string |
|--------|-----------------|
| 大小固定或易错 | 动态扩容 |
| 不知长度（衰减为指针） | `.size()` |
| 无边界检查 | `at()` 可选 |
| 难与算法协作 | 迭代器统一接口 |

```cpp
// 不推荐
int arr[100];
// 推荐
#include <vector>
std::vector<int> v;
v.push_back(1);
```

## 2. vector 入门

```cpp
#include <vector>
#include <iostream>

int main() {
    std::vector<int> scores;
    scores.push_back(90);
    scores.push_back(85);
    scores.push_back(88);

    std::cout << "size: " << scores.size() << '\n';
    std::cout << "第一个: " << scores[0] << '\n';
    std::cout << "最后一个: " << scores.back() << '\n';

    for (std::size_t i = 0; i < scores.size(); ++i)
        std::cout << scores[i] << ' ';
    std::cout << '\n';
    return 0;
}
```


### 术语三件套：vector

| 维度 | 内容 |
|------|------|
| **定义** | 动态数组容器，连续存储，随机访问 O(1)，尾部增删均摊 O(1)。 |
| **生活类比** | 可伸缩书架，书多了自动换更大的架子。 |
| **为什么重要** | STL 最常用容器；替代 raw 数组。 |
| **本章用到** | §2 |


## 3. string 常用操作

```cpp
#include <string>
#include <iostream>

int main() {
    std::string s1 = "Hello";
    std::string s2(" World");
    s1 += s2;
    std::cout << s1 << '\n';           // Hello World
    std::cout << s1.size() << '\n';
    std::cout << s1.substr(0, 5) << '\n';  // Hello
    std::cout << (s1.find("World") != std::string::npos) << '\n';
    return 0;
}
```

| 操作 | 说明 |
|------|------|
| `+` / `+=` | 连接 |
| `size()` / `length()` | 长度 |
| `substr(pos, n)` | 子串 |
| `find(str)` | 查找 |
| `empty()` | 是否为空 |
| `clear()` | 清空 |

## 4. 迭代器初识

```cpp
#include <vector>
#include <iostream>

int main() {
    std::vector<int> v{1, 2, 3, 4, 5};
    // 迭代器像「泛型指针」
    for (std::vector<int>::iterator it = v.begin(); it != v.end(); ++it)
        std::cout << *it << ' ';
    std::cout << '\n';

    for (auto it = v.cbegin(); it != v.cend(); ++it)
        std::cout << *it << ' ';
    return 0;
}
```

| 类型 | 含义 |
|------|------|
| `begin()` / `end()` | 可写迭代器区间 [begin, end) |
| `cbegin()` / `cend()` | 只读 |
| `rbegin()` / `rend()` | 反向 |

## 5. algorithm 入门

### 5.1 sort

```cpp
#include <algorithm>
#include <vector>
std::vector<int> v{3, 1, 4, 1, 5};
std::sort(v.begin(), v.end());
```

### 5.2 find

```cpp
#include <algorithm>
auto it = std::find(v.begin(), v.end(), 4);
if (it != v.end()) std::cout << "found at " << (it - v.begin()) << '\n';
```

### 5.3 for_each

```cpp
#include <algorithm>
#include <iostream>
std::for_each(v.begin(), v.end(), [](int x) {
    std::cout << x * 2 << ' ';
});
```

## 6. auto 与范围 for

```cpp
std::vector<std::string> names{"Alice", "Bob", "Carol"};
for (const auto& name : names)  // 只读引用，不拷贝
    std::cout << name << '\n';

for (auto& x : v)  // 可修改元素
    x *= 2;
```


#### 常见错误：范围 for 修改却用 by value

**错误写法：**

```cpp
for (auto x : v) x = 0;  // 只改拷贝
```

**正确写法：**

```cpp
for (auto& x : v) x = 0;
```

**原因**：按值拷贝元素，修改不影响容器。


## 7. vector + string 综合小项目

```cpp
#include <vector>
#include <string>
#include <algorithm>
#include <iostream>

int main() {
    std::vector<std::string> words{"banana", "apple", "cherry"};
    std::sort(words.begin(), words.end());
    for (const auto& w : words)
        std::cout << w << '\n';

    auto it = std::find(words.begin(), words.end(), "apple");
    if (it != words.end()) *it = "green apple";
    return 0;
}
```

## 8. 与 101 章 String 对照

[101 章](101-类与动态内存分配入门.md) 手写 String 为了理解内存；日常 **一律 `std::string` + `std::vector`**，Rule of Zero，算法库直接协作。
## FAQ

**Q：vector 越界？**  
A：`[]` 不检查；`at()` 抛异常。

**Q：string 与 C 字符串？**  
A：`c_str()` 返回 const char*，临时指针别长期保存。

**Q：迭代器失效？**  
A：扩容/erase 可能失效；见 [44 章](../44-vector-deque-string容器原理与实务.md)。

**Q：sort 能排 string？**  
A：可以，字典序。

**Q：list 为何不用？**  
A：初学 vector 足够；链表见 04/44 章。

**Q：auto 推导引用？**  
A：范围 for 用 `const auto&` 读大对象。

**Q：算法头文件？**  
A：`#include <algorithm>`。

**Q：还能用数组吗？**  
A：std::array 固定大小可以；raw 数组避免。
## 章末编程练习

> 建议：先闭卷写，再对照答案；答案在折叠区下方。

### 练习 1（基础）

读入 5 个 int 存 vector，排序后输出。

### 练习 2（基础）

string 查找子串并替换。

### 练习 3（进阶）

用 for_each 输出 vector 平方。

### 练习 4（进阶）

统计一行文本中单词数（stringstream 预告 105）。

### 练习 5（挑战）

合并两个有序 vector 仍有序（不用 merge，练算法）。

### 练习 6（挑战）

读文件每行存 vector<string>（预告 105）。


---

## 练习参考答案

### 练习 1 参考

```cpp
std::sort
```

### 练习 2 参考

```cpp
find + replace
```

### 练习 3 参考

```cpp
for_each lambda
```

### 练习 4 参考

```cpp
split 练习
```

### 练习 5 参考

```cpp
双指针合并
```

### 练习 6 参考

```cpp
fstream 105
```
## 闭卷自测（10 题）

> 合上书，15 分钟内完成；≥8 分再进入下一章。

1. vector 头文件？

2. string 连接操作？

3. 迭代器 end 含义？

4. sort 参数？

5. find 返回值？

6. 范围 for 只读写法？

7. 为何避免 C 数组？

8. push_back 复杂度？

9. cbegin 与 begin？

10. algorithm 命名空间？


---

## 自测参考答案

1. <vector>

2. +/+=

3. 尾后哨兵

4. begin,end

5. 迭代器或 end

6. const auto&

7. 安全/算法协作

8. 均摊 O(1)

9. const 只读

10. std

---


## 附录 D.1 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_1(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.2 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_2(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.3 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_3(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.4 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_4(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.5 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_5(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.6 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_6(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.7 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_7(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.1 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_1(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.2 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_2(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.3 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_3(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.4 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_4(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.5 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_5(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.6 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_6(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.7 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_7(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.1 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_1(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.2 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_2(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.3 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_3(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.4 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_4(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.5 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_5(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.6 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_6(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.7 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_7(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.1 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_1(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.2 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_2(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.3 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_3(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.4 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_4(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.5 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_5(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.6 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_6(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.7 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_7(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.1 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_1(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.2 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_2(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.3 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_3(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.4 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_4(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.5 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_5(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.6 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_6(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.7 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_7(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.1 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_1(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.2 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_2(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.3 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_3(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.4 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_4(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.5 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_5(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.6 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_6(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.7 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_7(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.1 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_1(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.2 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_2(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.3 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_3(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.4 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_4(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.5 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_5(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.6 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_6(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.7 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_7(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.1 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_1(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.2 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_2(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.3 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_3(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.4 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_4(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.5 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_5(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.6 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_6(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.7 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_7(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.1 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_1(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.2 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_2(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.3 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_3(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.4 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_4(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.5 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_5(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.6 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_6(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.7 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_7(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.1 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_1(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.2 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_2(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.3 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_3(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.4 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_4(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.5 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_5(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.6 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_6(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.7 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_7(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.1 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_1(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.2 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_2(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.3 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_3(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.4 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_4(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.5 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_5(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.6 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_6(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.7 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_7(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.1 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_1(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.2 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_2(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.3 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_3(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.4 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_4(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.5 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_5(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```


## 附录 D.6 STL 练习片段

```cpp
#include <vector>
#include <algorithm>
void drill_6(std::vector<int>& v) {
    std::sort(v.begin(), v.end());
}
```

## 下一章

继续阅读 **[105-IO输入输出与文件入门](105-IO输入输出与文件入门.md)**。



---

> **本章完** · 编码 UTF-8 · 如有勘误欢迎在学习笔记中标注页码与小节号。
