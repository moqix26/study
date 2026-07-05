# ABI 与二进制兼容性

> **文件编码**：UTF-8。
> **定位**：对象布局、name mangling、vtable、虚继承、pimpl——跨编译器发布 SDK 的必修课。
> **交叉阅读**：[29 对象模型](29-对象模型与虚函数表深入.md)、[48 编译链接](48-编译预处理与链接原理.md)、[03 面向对象](03-面向对象与类设计.md)、[05 现代 C++](05-现代C++新特性.md)。
> **上一章**：[81 UB 与陷阱](81-未定义行为UB与语言陷阱大全.md) | **下一章**：[83 错误处理哲学](83-C++错误处理哲学与方案抉择.md)

---

## §0 读前导读

### §0.1 一句话
**ABI** = 二进制接线标准；**API** = 源码头文件。

### §0.2 知识地图
- [ ] 解释 mangling 与 extern "C"
- [ ] 画单继承 vtable
- [ ] 说清虚继承指针调整
- [ ] 用 pimpl 稳定 ABI
- [ ] §闭卷自测 ≥8/10

---

## 本章在 81→86 链中的位置

**上一章**：[81 UB 与陷阱](81-未定义行为UB与语言陷阱大全.md)
**下一章**：[83 错误处理哲学](83-C++错误处理哲学与方案抉择.md)


## 1. API 与 ABI 的区别

**API（Application Programming Interface）**：头文件中的函数签名、类声明——源码层契约。

**ABI（Application Binary Interface）**：已编译二进制如何交互——对象布局、成员偏移、调用约定、name mangling、异常传播、vtable 格式。

改 API 不一定破坏 ABI（如仅改内联函数体）；改成员布局一定破坏 ABI。


## 2. Name Mangling

C++ 支持重载、命名空间、模板，链接符号必须编码：

```cpp
namespace ns {
void foo(int);
void foo(double);
}
extern "C" void bar(int);
```

```bash
g++ -c demo.cpp && nm demo.o | c++filt
# 典型：_ZN2ns3fooEi / _ZN2ns3fooEd / bar
```

`extern "C"` 关闭 mangling，用于跨语言边界；见 [48 章](48-编译预处理与链接原理.md)、[19 gRPC](19-gRPC与Protobuf工程化.md)。


## 3. Itanium C++ ABI

GCC、Clang 在 Linux/macOS 上多采用 **Itanium ABI**：

- vtable 放在 **类层次** 相关位置（非每对象完整拷贝）
- `type_info`、虚基类 offset 表
- 异常表 `.gcc_except_table`

MSVC 使用不同 mangling 与对象布局——**不可混用 C++ 标准库对象跨 DLL**。


## 4. vtable 布局（单继承）

```cpp
struct Base {
    virtual void f();
    virtual ~Base();
    int b = 1;
};
struct Derived : Base {
    void f() override;
    int d = 2;
};
```

典型对象布局：`[vptr → vtable][Base::b][Derived::d]` + padding。

vtable 含：指向虚函数的函数指针槽、`type_info*`、虚基类相关项（若有）。

深潜 [29 章](29-对象模型与虚函数表深入.md)、[43 章](43-继承与多态模型完全指南.md)。


## 5. 多重继承与虚继承内存

```cpp
struct VBase { virtual void v(); int x; };
struct Left  : virtual VBase { int l; };
struct Right : virtual VBase { int r; };
struct Bottom : Left, Right { int b; };
```

菱形继承：Bottom 对象含 Left 子对象、Right 子对象、**共享** VBase 子对象；`static_cast`/`dynamic_cast` 可能调整指针（thunk）。

白板题高频：画出 vptr 与 vbase offset 表。


## 6. ABI 稳定性原则

| 改动 | 破坏 ABI？ |
|------|------------|
| 类**末尾**添加非虚数据成员 | 通常否 |
| **中间**插入成员 | 是 |
| 增删**虚函数** | 是 |
| 改 `enum` 底层类型 | 是 |
| 改内联函数体（头文件） | 否* |

* 已编译调用方不重新编译则不感知，但静态库混用旧对象文件仍危险。


## 7. pimpl 解耦

```cpp
// widget.h — 公开头文件稳定
class Widget {
public:
    Widget();
    ~Widget();
    Widget(Widget&&) noexcept;
    Widget& operator=(Widget&&) noexcept;
    void draw();
private:
    struct Impl;
    std::unique_ptr<Impl> p_;
};
```

```cpp
// widget.cpp
struct Widget::Impl {
    std::vector<int> data;  // STL 不进头文件
    void draw() { /* ... */ }
};
Widget::Widget() : p_(std::make_unique<Impl>()) {}
Widget::~Widget() = default;
```

优点：编译防火墙、ABI 稳定。缺点：堆分配、间接访问、需显式三五法则。见 [21 章](21-设计模式与Infra工程实践.md) Bridge。


## 8. 跨编译器与跨版本发布

1. 公开 C ABI（`extern "C"` + 不透明句柄）或 pimpl + 工厂函数。
2. 不跨模块传递 `std::string`/`std::vector`（布局与分配器绑定实现）。
3. MSVC 统一 `/MD` 或 `/MT` 运行时。
4. 符号可见性：`-fvisibility=hidden` + 显式 export。
5. 插件系统几乎总是 C ABI。


## 9. 深化：调用约定

**调用约定**：x64 System V 与 Windows x64 不同；`this` 指针通过寄存器传递。

```cpp
// ABI deep-dive 9: 调用约定
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 10. 深化：RVO/NRVO

**RVO/NRVO**：返回值优化改变是否隐式移动，但不改变公开 API 签名。

```cpp
// ABI deep-dive 10: RVO/NRVO
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 11. 深化：虚函数 thunk

**虚函数 thunk**：多重继承下调换 this 指针的跳板函数。

```cpp
// ABI deep-dive 11: 虚函数 thunk
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 12. 深化：type_info 链接

**type_info 链接**：RTTI 字符串跨 DLL 可能 pointer 比较失败，用 `strcmp` name。

```cpp
// ABI deep-dive 12: type_info 链接
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 13. 深化：COMDAT 折叠

**COMDAT 折叠**：inline 函数多 TU 合并。

```cpp
// ABI deep-dive 13: COMDAT 折叠
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 14. 深化：dllexport

**dllexport**：Windows 导出表与 `__declspec(dllexport)`。

```cpp
// ABI deep-dive 14: dllexport
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 15. 深化：LTO 与 ABI

**LTO 与 ABI**：链接期优化可能内联跨 TU，改变调试符号。

```cpp
// ABI deep-dive 15: LTO 与 ABI
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 16. 深化：alignas 与 ABI

**alignas 与 ABI**：过度对齐类型跨模块传递需约定。

```cpp
// ABI deep-dive 16: alignas 与 ABI
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 17. 深化：std::function 布局

**std::function 布局**：小对象优化实现定义，勿跨 DLL。

```cpp
// ABI deep-dive 17: std::function 布局
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 18. 深化：异常跨 DLL

**异常跨 DLL**：MSVC 需同一编译器异常实现。

```cpp
// ABI deep-dive 18: 异常跨 DLL
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 19. 深化：调用约定

**调用约定**：x64 System V 与 Windows x64 不同；`this` 指针通过寄存器传递。

```cpp
// ABI deep-dive 19: 调用约定
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 20. 深化：RVO/NRVO

**RVO/NRVO**：返回值优化改变是否隐式移动，但不改变公开 API 签名。

```cpp
// ABI deep-dive 20: RVO/NRVO
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 21. 深化：虚函数 thunk

**虚函数 thunk**：多重继承下调换 this 指针的跳板函数。

```cpp
// ABI deep-dive 21: 虚函数 thunk
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 22. 深化：type_info 链接

**type_info 链接**：RTTI 字符串跨 DLL 可能 pointer 比较失败，用 `strcmp` name。

```cpp
// ABI deep-dive 22: type_info 链接
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 23. 深化：COMDAT 折叠

**COMDAT 折叠**：inline 函数多 TU 合并。

```cpp
// ABI deep-dive 23: COMDAT 折叠
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 24. 深化：dllexport

**dllexport**：Windows 导出表与 `__declspec(dllexport)`。

```cpp
// ABI deep-dive 24: dllexport
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 25. 深化：LTO 与 ABI

**LTO 与 ABI**：链接期优化可能内联跨 TU，改变调试符号。

```cpp
// ABI deep-dive 25: LTO 与 ABI
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 26. 深化：alignas 与 ABI

**alignas 与 ABI**：过度对齐类型跨模块传递需约定。

```cpp
// ABI deep-dive 26: alignas 与 ABI
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 27. 深化：std::function 布局

**std::function 布局**：小对象优化实现定义，勿跨 DLL。

```cpp
// ABI deep-dive 27: std::function 布局
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 28. 深化：异常跨 DLL

**异常跨 DLL**：MSVC 需同一编译器异常实现。

```cpp
// ABI deep-dive 28: 异常跨 DLL
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 29. 深化：调用约定

**调用约定**：x64 System V 与 Windows x64 不同；`this` 指针通过寄存器传递。

```cpp
// ABI deep-dive 29: 调用约定
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 30. 深化：RVO/NRVO

**RVO/NRVO**：返回值优化改变是否隐式移动，但不改变公开 API 签名。

```cpp
// ABI deep-dive 30: RVO/NRVO
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 31. 深化：虚函数 thunk

**虚函数 thunk**：多重继承下调换 this 指针的跳板函数。

```cpp
// ABI deep-dive 31: 虚函数 thunk
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 32. 深化：type_info 链接

**type_info 链接**：RTTI 字符串跨 DLL 可能 pointer 比较失败，用 `strcmp` name。

```cpp
// ABI deep-dive 32: type_info 链接
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 33. 深化：COMDAT 折叠

**COMDAT 折叠**：inline 函数多 TU 合并。

```cpp
// ABI deep-dive 33: COMDAT 折叠
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 34. 深化：dllexport

**dllexport**：Windows 导出表与 `__declspec(dllexport)`。

```cpp
// ABI deep-dive 34: dllexport
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 35. 深化：LTO 与 ABI

**LTO 与 ABI**：链接期优化可能内联跨 TU，改变调试符号。

```cpp
// ABI deep-dive 35: LTO 与 ABI
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 36. 深化：alignas 与 ABI

**alignas 与 ABI**：过度对齐类型跨模块传递需约定。

```cpp
// ABI deep-dive 36: alignas 与 ABI
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 37. 深化：std::function 布局

**std::function 布局**：小对象优化实现定义，勿跨 DLL。

```cpp
// ABI deep-dive 37: std::function 布局
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 38. 深化：异常跨 DLL

**异常跨 DLL**：MSVC 需同一编译器异常实现。

```cpp
// ABI deep-dive 38: 异常跨 DLL
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 39. 深化：调用约定

**调用约定**：x64 System V 与 Windows x64 不同；`this` 指针通过寄存器传递。

```cpp
// ABI deep-dive 39: 调用约定
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 40. 深化：RVO/NRVO

**RVO/NRVO**：返回值优化改变是否隐式移动，但不改变公开 API 签名。

```cpp
// ABI deep-dive 40: RVO/NRVO
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 41. 深化：虚函数 thunk

**虚函数 thunk**：多重继承下调换 this 指针的跳板函数。

```cpp
// ABI deep-dive 41: 虚函数 thunk
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 42. 深化：type_info 链接

**type_info 链接**：RTTI 字符串跨 DLL 可能 pointer 比较失败，用 `strcmp` name。

```cpp
// ABI deep-dive 42: type_info 链接
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 43. 深化：COMDAT 折叠

**COMDAT 折叠**：inline 函数多 TU 合并。

```cpp
// ABI deep-dive 43: COMDAT 折叠
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 44. 深化：dllexport

**dllexport**：Windows 导出表与 `__declspec(dllexport)`。

```cpp
// ABI deep-dive 44: dllexport
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 45. 深化：LTO 与 ABI

**LTO 与 ABI**：链接期优化可能内联跨 TU，改变调试符号。

```cpp
// ABI deep-dive 45: LTO 与 ABI
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 46. 深化：alignas 与 ABI

**alignas 与 ABI**：过度对齐类型跨模块传递需约定。

```cpp
// ABI deep-dive 46: alignas 与 ABI
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 47. 深化：std::function 布局

**std::function 布局**：小对象优化实现定义，勿跨 DLL。

```cpp
// ABI deep-dive 47: std::function 布局
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 48. 深化：异常跨 DLL

**异常跨 DLL**：MSVC 需同一编译器异常实现。

```cpp
// ABI deep-dive 48: 异常跨 DLL
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 49. 深化：调用约定

**调用约定**：x64 System V 与 Windows x64 不同；`this` 指针通过寄存器传递。

```cpp
// ABI deep-dive 49: 调用约定
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 50. 深化：RVO/NRVO

**RVO/NRVO**：返回值优化改变是否隐式移动，但不改变公开 API 签名。

```cpp
// ABI deep-dive 50: RVO/NRVO
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 51. 深化：虚函数 thunk

**虚函数 thunk**：多重继承下调换 this 指针的跳板函数。

```cpp
// ABI deep-dive 51: 虚函数 thunk
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 52. 深化：type_info 链接

**type_info 链接**：RTTI 字符串跨 DLL 可能 pointer 比较失败，用 `strcmp` name。

```cpp
// ABI deep-dive 52: type_info 链接
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 53. 深化：COMDAT 折叠

**COMDAT 折叠**：inline 函数多 TU 合并。

```cpp
// ABI deep-dive 53: COMDAT 折叠
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 54. 深化：dllexport

**dllexport**：Windows 导出表与 `__declspec(dllexport)`。

```cpp
// ABI deep-dive 54: dllexport
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 55. 深化：LTO 与 ABI

**LTO 与 ABI**：链接期优化可能内联跨 TU，改变调试符号。

```cpp
// ABI deep-dive 55: LTO 与 ABI
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 56. 深化：alignas 与 ABI

**alignas 与 ABI**：过度对齐类型跨模块传递需约定。

```cpp
// ABI deep-dive 56: alignas 与 ABI
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 57. 深化：std::function 布局

**std::function 布局**：小对象优化实现定义，勿跨 DLL。

```cpp
// ABI deep-dive 57: std::function 布局
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 58. 深化：异常跨 DLL

**异常跨 DLL**：MSVC 需同一编译器异常实现。

```cpp
// ABI deep-dive 58: 异常跨 DLL
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 59. 深化：调用约定

**调用约定**：x64 System V 与 Windows x64 不同；`this` 指针通过寄存器传递。

```cpp
// ABI deep-dive 59: 调用约定
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 60. 深化：RVO/NRVO

**RVO/NRVO**：返回值优化改变是否隐式移动，但不改变公开 API 签名。

```cpp
// ABI deep-dive 60: RVO/NRVO
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 61. 深化：虚函数 thunk

**虚函数 thunk**：多重继承下调换 this 指针的跳板函数。

```cpp
// ABI deep-dive 61: 虚函数 thunk
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 62. 深化：type_info 链接

**type_info 链接**：RTTI 字符串跨 DLL 可能 pointer 比较失败，用 `strcmp` name。

```cpp
// ABI deep-dive 62: type_info 链接
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 63. 深化：COMDAT 折叠

**COMDAT 折叠**：inline 函数多 TU 合并。

```cpp
// ABI deep-dive 63: COMDAT 折叠
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 64. 深化：dllexport

**dllexport**：Windows 导出表与 `__declspec(dllexport)`。

```cpp
// ABI deep-dive 64: dllexport
struct ExportMe {
    virtual ~ExportMe() = default;
};
```

与 [29 对象模型](29-对象模型与虚函数表深入.md)、[48 链接](48-编译预处理与链接原理.md) 对照。


## 练习

1. 对照正文完成最小可编译 demo。
2. 与交叉链接章节各读一节并做笔记。
3. 闭卷自测前先合上书复述知识地图。
4. 将本章术语填入 [15 补充总表](15-补充知识点总表.md) 个人区。
5. 与同学互相出题白板 10 分钟。

## FAQ

**Q：ABI 和 API 哪个更重要？**
发二进制 SDK 看 ABI；只发源码看 API。

**Q：pimpl 缺点？**
堆分配、间接、移动需手写。

**Q：与 [48 ODR](48-编译预处理与链接原理.md)？**
ODR 管源码实体；ABI 管布局。

## 闭卷自测

1. mangling 解决什么问题？
2. extern "C" 局限？
3. 单继承几个 vptr？
4. 虚继承为何调整指针？
5. pimpl 如何稳定 ABI？
6. 改 vtable 为何破坏兼容？
7. Itanium ABI 用于哪些平台？
8. EBO 影响 ABI 吗？
9. 跨编译器传 std::string 风险？
10. RTTI 跨 DLL 注意什么？

<details>
<summary>自测参考答案</summary>

1. 重载/命名空间编码。
2. 无重载、无 C++ 异常统一。
3. 通常一个。
4. 多路径到虚基类。
5. 隐藏实现细节。
6. 槽位顺序变。
7. Linux/macOS GCC/Clang。
8. 影响 sizeof/偏移。
9. 布局/分配器不同。
10. 用 name 比较。

</details>


---

## 下一章预告

[83 错误处理哲学](83-C++错误处理哲学与方案抉择.md)

*下一章：83 错误处理哲学*
