# 现代 C++ 惯用法（Idiom）集

> **文件编码**：UTF-8。
> **定位**：pimpl、CRTP、EBO、SFINAE、tag dispatch、ADL、type erasure、copy-and-swap。
> **交叉阅读**：[05](05-现代C++新特性.md)、[06 模板](06-模板与泛型编程.md)、[29 对象模型](29-对象模型与虚函数表深入.md)、[82 pimpl](82-ABI与二进制兼容性.md)。
> **上一章**：[84 内存模型](84-C++内存模型与原子操作深入.md) | **下一章**：[86 标准库全景](86-C++20-26标准库组件全景.md)

---

## §0 读前导读

### §0.1 一句话
惯用法是读懂 Boost/Folly/Abseil 的钥匙。

### §0.2 知识地图
- [ ] 每种 idiom 写一最小例子
- [ ] 对比 CRTP 与虚函数
- [ ] §闭卷自测 ≥8/10

---

## 本章在 81→86 链中的位置

**上一章**：[84 内存模型](84-C++内存模型与原子操作深入.md)
**下一章**：[86 标准库全景](86-C++20-26标准库组件全景.md)


## 1. pimpl

见 [82 章](82-ABI与二进制兼容性.md)。

```cpp
class Widget {
    struct Impl;
    std::unique_ptr<Impl> p;
public:
    Widget();
    ~Widget();
    void draw();
};
```


## 2. CRTP

静态多态，无虚表。

```cpp
template<class D>
struct Base {
    void f() { static_cast<D*>(this)->impl(); }
};
struct Derived : Base<Derived> {
    void impl() { /* ... */ }
};
```

热路径可内联；[29 章](29-对象模型与虚函数表深入.md) 对比虚函数。


## 3. EBO

空基类不占子对象空间。

```cpp
struct E {};
struct X : E { int x; };
// sizeof(X) 常为 4
```

`std::function` 小对象优化类比。


## 4. SFINAE

[06 章](06-模板与泛型编程.md)；C++20 用 concepts 替代部分。

```cpp
template<class T>
std::enable_if_t<std::is_integral_v<T>, T>
abs(T x) { return x < 0 ? -x : x; }
```


## 5. Tag Dispatch

```cpp
struct fast_t {};
struct safe_t {};
inline constexpr fast_t fast;
inline constexpr safe_t safe;
void sort(fast_t);
void sort(safe_t);
```


## 6. ADL

```cpp
using std::swap;
swap(a, b);  // ADL 找自定义 swap
```


## 7. Type Erasure

```cpp
class Drawable {
    struct Concept { virtual void draw() = 0; virtual ~Concept() = default; };
    template<class T>
    struct Model : Concept {
        T obj;
        void draw() override { obj.draw(); }
    };
    std::unique_ptr<Concept> p;
public:
    template<class T> Drawable(T x) : p(std::make_unique<Model<T>>(std::move(x))) {}
    void draw() { p->draw(); }
};
```

类似 `std::function`。


## 8. Copy-and-Swap

[07 章](07-异常处理与RAII.md) 强保证赋值。

```cpp
Friend& operator=(Friend other) noexcept {
    swap(*this, other);
    return *this;
}
```


## 9. 选型表

| 需求 | idiom |
|------|-------|
| 稳定 ABI | pimpl |
| 热路径多态 | CRTP |
| 小对象 | EBO |
| 约束模板 | concepts/SFINAE |
| 策略 | tag dispatch |
| 自定义 swap | ADL |
| 非继承多态 | type erasure |
| 强保证赋值 | copy-and-swap |


## 10. 变体：compressed_pair

**compressed_pair** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 11. 变体：scope_guard

**scope_guard** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 12. 变体：iterator facade

**iterator facade** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 13. 变体：named ctor

**named ctor** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 14. 变体：factory create

**factory create** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 15. 变体：policy design

**policy design** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 16. 变体：expression templates

**expression templates** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 17. 变体：SBO

**SBO** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 18. 变体：attorney-client

**attorney-client** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 19. 变体：handle-body

**handle-body** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 20. 变体：mixin

**mixin** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 21. 变体：detector

**detector** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 22. 变体：void_t

**void_t** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 23. 变体：requires

**requires** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 24. 变体：enum bitmask

**enum bitmask** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 25. 变体：string SSO

**string SSO** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 26. 变体：small vector

**small vector** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 27. 变体：intrusive list

**intrusive list** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 28. 变体：object pool

**object pool** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 29. 变体：flyweight

**flyweight** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 30. 变体：state machine

**state machine** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 31. 变体：overload pattern

**overload pattern** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 32. 变体：currying

**currying** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 33. 变体：BNF parser combinator

**BNF parser combinator** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 34. 变体：typed erasure

**typed erasure** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 35. 变体：Y combinator

**Y combinator** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 36. 变体：MVC passive

**MVC passive** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 37. 变体：RAII mutex

**RAII mutex** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 38. 变体：unique_any

**unique_any** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 39. 变体：callable traits

**callable traits** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 40. 变体：compressed_pair

**compressed_pair** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 41. 变体：scope_guard

**scope_guard** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 42. 变体：iterator facade

**iterator facade** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 43. 变体：named ctor

**named ctor** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 44. 变体：factory create

**factory create** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 45. 变体：policy design

**policy design** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 46. 变体：expression templates

**expression templates** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 47. 变体：SBO

**SBO** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 48. 变体：attorney-client

**attorney-client** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 49. 变体：handle-body

**handle-body** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 50. 变体：mixin

**mixin** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 51. 变体：detector

**detector** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 52. 变体：void_t

**void_t** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 53. 变体：requires

**requires** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 54. 变体：enum bitmask

**enum bitmask** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 55. 变体：string SSO

**string SSO** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 56. 变体：small vector

**small vector** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 57. 变体：intrusive list

**intrusive list** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 58. 变体：object pool

**object pool** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 59. 变体：flyweight

**flyweight** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 60. 变体：state machine

**state machine** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 61. 变体：overload pattern

**overload pattern** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 62. 变体：currying

**currying** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 63. 变体：BNF parser combinator

**BNF parser combinator** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 64. 变体：typed erasure

**typed erasure** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 65. 变体：Y combinator

**Y combinator** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 66. 变体：MVC passive

**MVC passive** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 67. 变体：RAII mutex

**RAII mutex** 与八大 idiom 的组合应用。

```cpp
template<class Policy>
struct Algo : Policy {
    void run() { this->policy_impl(); }
};
```

阅读 [06 模板](06-模板与泛型编程.md)、[21 设计模式](21-设计模式与Infra工程实践.md)。


## 练习

1. 对照正文完成最小可编译 demo。
2. 与交叉链接章节各读一节并做笔记。
3. 闭卷自测前先合上书复述知识地图。
4. 将本章术语填入 [15 补充总表](15-补充知识点总表.md) 个人区。
5. 与同学互相出题白板 10 分钟。

## FAQ

**Q：CRTP vs 虚函数？**
CRTP 编译期、可内联；虚函数运行时多态。

**Q：type erasure 成本？**
通常堆分配 + 虚调；小对象可 SBO。

## 闭卷自测

1. pimpl 优点？
2. CRTP 模式？
3. EBO 条件？
4. SFINAE 用途？
5. tag dispatch？
6. ADL 经典？
7. type erasure vs 继承？
8. copy-and-swap 保证？
9. 与 06 关系？
10. 与 82 ABI？

<details>
<summary>自测参考答案</summary>

1. 隐藏实现稳定 ABI。
2. Derived: Base<Derived>。
3. 空基类零大小。
4. 模板约束。
5. 编译期策略。
6. swap/adl。
7. 无 vtable 运行时类型。
8. 强异常安全。
9. 模板技巧集。
10. pimpl 减 ABI 破坏。

</details>


---

## 下一章预告

[86 标准库全景](86-C++20-26标准库组件全景.md)

*下一章：86 标准库全景*
