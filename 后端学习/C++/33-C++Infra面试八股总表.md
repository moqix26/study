# C++ Infra 面试八股总表

> **文件编码**：UTF-8。Infra/基建/游戏/量化岗 **200+ 知识点索引** + **2 分钟口述模板**；配合 [14 高频面试](14-高频面试专题与场景题.md)、[15 补充总表](15-补充知识点总表.md)、[35 KV-Store 项目](35-项目实战高性能KV-Store.md)。

---

## 本章与前后章的关系

| 上一章（32/15） | 本章（33） | 下一章（34） |
|-----------------|------------|--------------|
| 知识点总表/扩展 | **面试索引 + 口述** | 手撕 TOP50 |
| 章节定位 | 200+ 条速查 | 白板代码 |

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

这是 **C++ Infra 面试八股的「电话簿」**：每条 1 行索引 + 口述模板，面试前 30～60 分钟按模块过一遍，不会的点跳回 01～23 对应章。

### 0.2 你需要提前知道什么

| 状态 | 动作 |
|------|------|
| 14 章 Q&A 没刷 | 先 [14 章](14-高频面试专题与场景题.md) 再回本表勾选 |
| 只会背不会讲 | 用 §8 口述模板，每题限时 2 min |
| 投 Infra/游戏岗 | 重点 §2～§4 + §6 网络 + §7 体系结构 |
| 投 LLM Infra | 叠加 [LLMInfra 20](../LLMInfra/20-面试专题与系统设计.md) |

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

- ☐ 内存模块 ≥35 条能 ⬜→🔶
- ☐ OOP 模块 ≥30 条能口述 vtable/三五法则
- ☐ STL 模块能对比 vector/map/unordered_map
- ☐ 并发模块能讲死锁四条件 + memory order 入门
- ☐ 网络模块能讲 epoll LT/ET + TCP 状态
- ☐ 体系结构能讲 cache line / false sharing
- ☐ §8 口述模板能闭卷讲 ≥15 条
- ☐ §11 闭卷自测 ≥8/10

### 0.4 节奏与验证

首轮 2 h 通读 §2～§7 标 ⬜；二轮 3 h 用 §8 口述录音；面试前 45 min 弱项+网络+体系结构；闭卷 §11。  
**图例**：⬜ 知道 · 🔶 会用 · ✅ 会讲

---

## 1. 本表怎么用

1. **查漏补缺**：按模块扫表，⬜ 跳回对应章节
2. **口述训练**：用 §8「结论→原理→代码→项目」四段式
3. **与 14 章分工**：14 章是 **深度 Q&A**；本章是 **广度索引**
4. **与 15 章分工**：15 章覆盖 01～14 学习进度；本章 **面向 Infra 面试加料**

### 1.1 口述四段式（通用模板）

```text
【结论】一句话（≤15 字）
【原理】2～3 个要点（带复杂度/边界）
【代码/图】可选：1 行代码或 ASCII 图
【项目】mini-http / KV-Store / 线程池踩坑一句
```

### 1.2 模块与文档对照

| 模块 | 主文档 | 扩展 |
|------|--------|------|
| 内存 | 02、05、07、18 | Valgrind/ASan（12） |
| OOP | 03、21 | 设计模式 |
| STL | 04、13 | 算法题 |
| 并发 | 08、23 | 线程池 |
| 网络 | 10、23、19 | gRPC |
| 体系结构 | 22、18 | perf/cache |

---

## 2. 内存与资源管理（M01～M42，42 条）

| 编号 | 知识点 | 文档 | 掌握 | 一句话 |
|------|--------|------|------|--------|
| M01 | 栈 vs 堆分配 | 02 | ⬜ | 栈自动快小；堆 new/智能指针 |
| M02 | 栈帧与调用栈 | 02 | ⬜ | 局部变量、返回地址在栈 |
| M03 | 堆碎片 | 02 | ⬜ | 频繁 new/delete 可能碎片 |
| M04 | new/delete vs malloc/free | 02/14-Q3 | ⬜ | new 调构造；malloc 只字节 |
| M05 | new[]/delete[] 配对 | 02 | ⬜ | 数组必须 delete[] |
| M06 | 内存泄漏定义 | 02/14-Q2 | ⬜ | 丢失指针未释放 |
| M07 | fd/socket 泄漏 | 07/12 | ⬜ | OS 资源，RAII 封装 |
| M08 | RAII 原则 | 07 | ⬜ | 构造获取析构释放 |
| M09 | Rule of Zero | 05/14-Q12 | ⬜ | 成员皆 RAII 则不手写五函数 |
| M10 | Rule of Five | 05/14-Q12 | ⬜ | 析构/拷贝/移动共五个 |
| M11 | unique_ptr 独占 | 05 | ⬜ | 不可拷贝可 move |
| M12 | shared_ptr 引用计数 | 05 | ⬜ | 控制块+对象，make_shared 优 |
| M13 | weak_ptr 破环 | 05 | ⬜ | 不增计数，lock 变 shared |
| M14 | 循环引用泄漏 | 05/14 | ⬜ | shared 互持→weak 打破 |
| M15 | make_unique/make_shared | 05 | ⬜ | 异常安全+减少分配 |
| M16 | 野指针 | 02/14-Q5 | ⬜ | 未初始化或释放后用 |
| M17 | 悬空引用/dangling | 05/14 | ⬜ | string_view 指向已销毁 string |
| M18 | 左值/右值 | 05/14-Q11 | ⬜ | 有名可取址 vs 临时量 |
| M19 | 右值引用 T&& | 05 | ⬜ | 绑定临时，延长生命周期规则 |
| M20 | std::move 语义 | 05/14-Q13 | ⬜ | 仅 cast，不移动 |
| M21 | 移动构造/赋值 | 05 | ⬜ | 偷资源，源变 valid 未指定 |
| M22 | 完美转发 forward | 05/14-Q14 | ⬜ | 保持值类别 |
| M23 | NRVO/RVO | 05 | ⬜ | 返回值优化，少拷贝 |
| M24 | 拷贝 elision | 05 | ⬜ | 编译器省略拷贝 |
| M25 | 对齐 alignof/alignas | 18/14-Q29 | ⬜ | 硬件访问对齐更快 |
| M26 | 结构体 padding | 18 | ⬜ | 成员顺序影响大小 |
| M27 | 内存对齐与 false sharing | 18/22 | ⬜ | 不同核写同 cache line |
| M28 | placement new | 18 | ⬜ | 已分配内存上构造 |
| M29 | 自定义 allocator | 04/18 | ⬜ | STL 容器可换分配器 |
| M30 | 内存池动机 | 18/34 | ⬜ | 减少 syscall、碎片 |
| M31 | tcmalloc/jemalloc 概念 | 18 | ⬜ | 多线程友好分配器 |
| M32 | Valgrind memcheck | 12 | ⬜ | 堆泄漏检测 |
| M33 | ASan AddressSanitizer | 12 | ⬜ | 编译期插桩越界/泄漏 |
| M34 | UAF use-after-free | 12 | ⬜ | 释放后使用，UB |
| M35 | double free | 02 | ⬜ | 重复 delete，UB |
| M36 | 未定义行为 UB | 14/08 | ⬜ | 标准不保证，Release 难查 |
| M37 | 实现定义/ unspecified | 02 | ⬜ | 编译器决定但需文档 |
| M38 | 对象模型：非 POD 布局 | 03/18 | ⬜ | vptr、成员顺序 |
| M39 | 空基类优化 EBO | 06 | ⬜ | 空基类不占子对象空间 |
| M40 | 位域 bit-field | 02 | ⬜ | 节省空间，实现相关 |
| M41 | volatile 非线程同步 | 08 | ⬜ | 别用 volatile 当 atomic |
| M42 | 资源获取即初始化 FdGuard | 07/10 | ⬜ | close 在析构 |

---

## 3. OOP 与语言机制（O01～O38，38 条）

| 编号 | 知识点 | 文档 | 掌握 | 一句话 |
|------|--------|------|------|--------|
| O01 | 封装/继承/多态 | 03 | ⬜ | 三大特性 |
| O02 | public/protected/private | 03 | ⬜ | 访问控制 |
| O03 | 构造/析构顺序 | 03 | ⬜ | 基→派生建；派→基毁 |
| O04 | 虚函数与动态绑定 | 03/14-Q6 | ⬜ | 运行时查 vtable |
| O05 | vptr/vtable 布局 | 03/14 | ⬜ | 对象头存 vptr |
| O06 | 虚析构必要性 | 03/14-Q7 | ⬜ | 基指针 delete 派生 |
| O07 | override/final | 03/14-Q8 | ⬜ | 编译期检查/禁止重写 |
| O08 | 纯虚函数/抽象类 | 03/14-Q9 | ⬜ | =0 不可实例化 |
| O09 | 接口 vs 抽象类 | 03/21 | ⬜ | 纯虚集合 vs 部分实现 |
| O10 | 静态多态 CRTP | 06 | ⬜ | 模板编译期多态 |
| O11 | 运行时 vs 编译期多态 | 03/14-Q10 | ⬜ | virtual vs template |
| O12 | 虚函数开销 | 03 | ⬜ | 间接调用+难内联 |
| O13 | 虚继承/菱形问题 | 03 | ⬜ | 共享基类子对象 |
| O14 | 默认构造/删除函数 =default/=delete | 05 | ⬜ | 显式控制特殊成员 |
| O15 | 三五法则与拷贝交换 | 03/14 | ⬜ | copy-and-swap idiom |
| O16 | const 成员函数 | 03 | ⬜ | 不修改 *this 逻辑状态 |
| O17 | mutable 成员 | 03 | ⬜ | const 函数内可改 |
| O18 | static 成员/函数 | 03 | ⬜ | 类级别，无 this |
| O19 | 友元 friend | 03 | ⬜ | 打破封装，谨慎 |
| O20 | 运算符重载 | 03 | ⬜ | 保持语义一致 |
| O21 | 显式转换 explicit | 14-Q28 | ⬜ | 防隐式窄化 |
| O22 | 四种 cast | 14-Q30 | ⬜ | static/dynamic/reinterpret/const |
| O23 | dynamic_cast 与 RTTI | 03 | ⬜ | 多态类型安全向下转 |
| O24 | typeid | 03 | ⬜ | RTTI 类型信息 |
| O25 | 重载 vs 重写 vs 隐藏 | 03 | ⬜ | overload/override/hide |
| O26 | 构造函数初始化列表 | 03 | ⬜ | 成员/const/引用必用 |
| O27 | 委托构造 | 05 | ⬜ | 构造函数链 |
| O28 | PIMPL 惯用法 | 21 | ⬜ | 隐藏实现减编译依赖 |
| O29 | 单例模式线程安全 | 21/08 | ⬜ | Meyers singleton / once_flag |
| O30 | 工厂模式 | 21 | ⬜ | 创建与使用解耦 |
| O31 | 策略/观察者（游戏常考） | 21 | ⬜ | 行为可替换/事件订阅 |
| O32 | 值语义 vs 指针语义 | 03/05 | ⬜ | C++ 默认值拷贝 |
| O33 | 切片 slicing | 03 | ⬜ | 派生按值赋给基类丢部分 |
| O34 | 对象池与 placement | 21/34 | ⬜ | 复用对象降分配 |
| O35 | enum class | 05 | ⬜ | 强类型枚举 |
| O36 | std::variant/optional | 05 | ⬜ | 类型安全联合体/可空 |
| O37 | 异常安全等级 | 07 | ⬜ | basic/strong/nothrow |
| O38 | noexcept 与 move | 07/05 | ⬜ | move 常 noexcept |

---

## 4. STL 与算法（S01～S40，40 条）

| 编号 | 知识点 | 文档 | 掌握 | 一句话 |
|------|--------|------|------|--------|
| S01 | vector 连续内存 | 04/14-Q15 | ⬜ | 随机访问 O(1) |
| S02 | vector 扩容 2 倍 | 04 | ⬜ | reallocate 均摊 O(1) |
| S03 | vector 迭代器失效 | 04/14-Q18 | ⬜ | push_back 可能全失效 |
| S04 | list 双向链表 | 04 | ⬜ | 插删 O(1)无随机访问 |
| S05 | deque 双端队列 | 04 | ⬜ | 分段连续 |
| S06 | array 固定栈数组 | 04 | ⬜ | 编译期大小 |
| S07 | map 红黑树 | 04/14-Q17 | ⬜ | 有序 O(log n) |
| S08 | set 唯一键 | 04 | ⬜ | 有序集合 |
| S09 | unordered_map 哈希 | 04 | ⬜ | 均摊 O(1)无序 |
| S10 | 哈希冲突与 rehash | 04 | ⬜ | 负载因子触发扩容 |
| S11 | 自定义 hash/比较 | 04 | ⬜ | 结构体作 key |
| S12 | multimap/multiset | 04 | ⬜ | 允许重复键 |
| S13 | priority_queue 堆 | 04/13 | ⬜ | 默认大顶堆 |
| S14 | stack/queue 适配器 | 04 | ⬜ | 容器适配 |
| S15 | emplace vs push | 04/14-Q19 | ⬜ | 原地构造 |
| S16 | shrink_to_fit | 04 | ⬜ | 请求释放多余 capacity |
| S17 | reserve 预分配 | 04 | ⬜ | 避免多次扩容 |
| S18 | 迭代器 category | 04 | ⬜ | random/bidirectional/forward |
| S19 | 范围 for 与迭代器 | 04 | ⬜ | 语法糖 |
| S20 | sort 复杂度 O(n log n) | 04/13 | ⬜ |  introsort |
| S21 | stable_sort | 04 | ⬜ | 稳定 O(n log n) |
| S22 | lower_bound/upper_bound | 04/13 | ⬜ | 有序二分 |
| S23 | std::find vs map::find | 04 | ⬜ | 线性 vs 对数 |
| S24 | remove-erase 惯用法 | 04 | ⬜ | vector 删元素 |
| S25 | 算法复杂度选型 | 13 | ⬜ | 面试报复杂度 |
| S26 | string SSO 小字符串优化 | 04 | ⬜ | 短串栈上存 |
| S27 | string_view 非拥有 | 05 | ⬜ | 注意生命周期 |
| S28 | pair/tuple structured binding | 05 | ⬜ | C++17 解构 |
| S29 | algorithm 头常用 | 04 | ⬜ | copy, transform, accumulate |
| S30 | 函数对象与 lambda | 05/14 | ⬜ | 谓词/回调 |
| S31 | std::function 类型擦除 | 05 | ⬜ | 有开销 |
| S32 | 容器线程安全 | 08 | ⬜ | STL 容器本身非线程安全 |
| S33 | concurrent 容器概念 | 08 | ⬜ | 需外部锁或 TBB |
| S34 | 迭代器 invalidation 汇总 | 04/14 | ⬜ | 各容器不同 |
| S35 | map node handle(C++17) | 04 | ⬜ | 提取节点 |
| S36 | flat_map 概念 | 04 | ⬜ | 有序 vector+二分 |
| S37 | small vector 概念 | 18 | ⬜ | 小容量栈上 |
| S38 | LRU = list + unordered_map | 13/34 | ⬜ | O(1) get/put |
| S39 | 并查集 | 13 | ⬜ | 路径压缩 |
| S40 |  TopK 堆 | 13 | ⬜ | 维护 size k 堆 |

---

## 5. 并发与多线程（C01～C40，40 条）

| 编号 | 知识点 | 文档 | 掌握 | 一句话 |
|------|--------|------|------|--------|
| C01 | std::thread 启动/join | 08 | ⬜ | join 或 detach 二选一 |
| C02 | 线程 detach 风险 | 08 | ⬜ | 生命周期难控 |
| C03 | mutex 互斥锁 | 08/14-Q20 | ⬜ | 保护临界区 |
| C04 | lock_guard vs unique_lock | 08 | ⬜ | 后者可 unlock 配 cv |
| C05 | scoped_lock(C++17) | 08 | ⬜ | 多锁防死锁 |
| C06 | recursive_mutex | 08 | ⬜ | 同线程可重入 |
| C07 | condition_variable | 08/14-Q22 | ⬜ | wait 配谓词防虚假唤醒 |
| C08 | 虚假唤醒 spurious wakeup | 08 | ⬜ | 必须用 while 条件 |
| C09 | 生产者-消费者模型 | 08/34 | ⬜ | queue+mutex+cv |
| C10 | 线程安全队列 | 08/14-Q41 | ⬜ | 阻塞 pop/push |
| C11 | 线程池动机 | 08/23 | ⬜ | 减创建销毁开销 |
| C12 | 任务队列+worker | 08 | ⬜ | 固定 N 线程取任务 |
| C13 | future/promise/async | 08 | ⬜ | 异步结果 |
| C14 | packaged_task | 08 | ⬜ | 任务封装 |
| C15 | atomic 无锁计数 | 08/14-Q20 | ⬜ | 单变量 CAS |
| C16 | memory_order 六种 | 08/14-Q35 | ⬜ | seq_cst/acquire/release… |
| C17 | happens-before | 08 | ⬜ | 可见性规则 |
| C18 | 数据竞争=UB | 08/14 | ⬜ | 无同步并发写 |
| C19 | 死锁四条件 | 08/14-Q21 | ⬜ | 互斥/占有等待/不可剥夺/循环 |
| C20 | 固定加锁顺序 | 08 | ⬜ | 破坏循环等待 |
| C21 | std::lock 同时加锁 | 08 | ⬜ | 避免 AB-BA |
| C22 | 读写锁 shared_mutex | 08 | ⬜ | 读多写少 |
| C23 | 自旋锁适用场景 | 08/14 | ⬜ | 临界区极短 |
| C24 | 线程本地存储 TLS | 08 | ⬜ | thread_local |
| C25 | 线程池大小估算 | 08/23 | ⬜ | CPU 密集≈核数；IO 可更多 |
| C26 | 伪并行 vs 并行 | 08 | ⬜ | 单核时间片 |
| C27 | 协程 vs 线程(概念) | 08 | ⬜ | 用户态调度 |
| C28 | C++20 jthread | 05 | ⬜ | RAII join |
| C29 | stop_token 协作取消 | 05 | ⬜ | 优雅停止 |
| C30 | 屏障 barrier/latch | 08 | ⬜ | 同步点 |
| C31 | semaphore(C++20) | 08 | ⬜ | 计数信号量 |
| C32 | SPSC/MPSC 队列 | 21/23 | ⬜ | 无锁队列场景 |
| C33 | ABA 问题 | 08 | ⬜ | CAS 经典坑 |
| C34 | 双重检查锁定 DCL | 08/21 | ⬜ | 需 memory barrier |
| C35 | 线程池拒绝策略 | 08 | ⬜ | 抛异常/丢弃/CallerRuns |
| C36 | 定时器与 IO 线程分离 | 23 | ⬜ | Reactor 模型 |
| C37 | 并发 bug 复现难 | 12 | ⬜ | TSan/压测 |
| C38 | TSan ThreadSanitizer | 12 | ⬜ | 数据竞争检测 |
| C39 | 锁粒度 | 08 | ⬜ | 粗锁简单细锁吞吐 |
| C40 | 无锁≠无 bug | 08 | ⬜ | 正确性更难证 |

---

## 6. 网络与 IO（N01～N35，35 条）

| 编号 | 知识点 | 文档 | 掌握 | 一句话 |
|------|--------|------|------|--------|
| N01 | TCP 三次握手 | 10/计网 | ⬜ | SYN-SYN/ACK-ACK |
| N02 | TCP 四次挥手 | 10 | ⬜ | FIN 半关闭 |
| N03 | TIME_WAIT 作用 | 10 | ⬜ | 防旧包、可靠关闭 |
| N04 | TCP 粘包/拆包 | 10 | ⬜ | 流式需定界 |
| N05 | socket 基本 API | 10 | ⬜ | socket/bind/listen/accept |
| N06 |阻塞 vs 非阻塞 | 10/23 | ⬜ | fcntl O_NONBLOCK |
| N07 | select 缺点 | 23 | ⬜ | O(n)扫描+1024 限制 |
| N08 | poll vs select | 23 | ⬜ | pollfd 数组无上限 |
| N09 | epoll 创建/ctl/wait | 23/14-Q40 | ⬜ | O(1) 就绪通知 |
| N10 | epoll LT 水平触发 | 23 | ⬜ | 未读完仍通知 |
| N11 | epoll ET 边缘触发 | 23 | ⬜ | 需循环读到 EAGAIN |
| N12 | Reactor 模型 | 23/21 | ⬜ | 等事件→分发 |
| N13 | 线程池+epoll 架构 | 23/35 | ⬜ | accept/IO 与计算分离 |
| N14 | HTTP 请求行/头/体 | 10 | ⬜ | GET/POST 解析 |
| N15 | HTTP Keep-Alive | 10 | ⬜ | 复用 TCP |
| N16 | HTTP/1.1 vs HTTP/2 | 10 | ⬜ | 多路复用/HPACK |
| N17 | 零拷贝 sendfile | 23 | ⬜ | 内核态传文件 |
| N18 | mmap 文件映射 | 11/35 | ⬜ | 持久化/WAL |
| N19 | UDP vs TCP | 10 | ⬜ | 无连接不可靠 |
| N20 | DNS 解析(概念) | 计网 | ⬜ | 域名→IP |
| N21 | 端口与五元组 | 10 | ⬜ | 连接标识 |
| N22 | SO_REUSEADDR | 10 | ⬜ | 重启 bind |
| N23 | TCP_NODELAY 禁用 Nagle | 10 | ⬜ | 低延迟小包 |
| N24 |  backlog 半连接队列 | 10 | ⬜ | listen 参数 |
| N25 | 惊群 thundering herd | 23 | ⬜ | 多进程 accept |
| N26 | io_uring 概念 | 23 | ⬜ | 异步 syscall 批量 |
| N27 | Boost.Asio io_context | 23 | ⬜ | Proactor 风格 |
| N28 | gRPC over HTTP/2 | 19 | ⬜ | Infra RPC 常用 |
| N29 | Protobuf 序列化 | 19 | ⬜ | 二进制 schema |
| N30 | 长连接 SSE/WebSocket | 10 | ⬜ | LLM 流式输出 |
| N31 | 反向代理 nginx 角色 | 23 | ⬜ | 接客+负载均衡 |
| N32 | 连接限流与超时 | 10/35 | ⬜ | 防慢连接攻击 |
| N33 | 字节序 htons/ntoh | 10 | ⬜ | 网络大端 |
| N34 | 自定义二进制协议 | 35 | ⬜ | length-prefix |
| N35 | WAL 预写日志 | 35/11 | ⬜ | 先日志后刷盘 |

---

## 7. 计算机体系结构（A01～A30，30 条）

| 编号 | 知识点 | 文档 | 掌握 | 一句话 |
|------|--------|------|------|--------|
| A01 | CPU 缓存层次 L1/L2/L3 | 22 | ⬜ | 越小越快越贵 |
| A02 | cache line 典型 64B | 22/18 | ⬜ | 最小缓存单位 |
| A03 | 时间/空间局部性 | 22 | ⬜ | 刚用过/相邻会再用 |
| A04 | false sharing | 18/22 | ⬜ | 独立变量同 line 互 invalid |
| A05 | 缓存对齐优化 | 18 | ⬜ | alignas 64 分离热数据 |
| A06 | NUMA 概念 | 22 | ⬜ | 多 socket 本地内存快 |
| A07 | 分支预测 | 22 | ⬜ | 预测错流水线清空 |
| A08 | 指令流水线 | 22 | ⬜ | 取指译码执行重叠 |
| A09 | 乱序执行 | 22 | ⬜ | CPU 重排保 as-if |
| A10 | 内存屏障与 atomic | 08/22 | ⬜ | 编译/硬件重排约束 |
| A11 | TLB | 22 | ⬜ | 页表缓存 |
| A12 | 页 fault | 11/22 | ⬜ | 缺页中断加载 |
| A13 | 大页 huge page | 22 | ⬜ | 减 TLB miss |
| A14 | SIMD SSE/AVX 概念 | 22 | ⬜ | 向量指令 |
| A15 | perf 采样 | 12 | ⬜ | top 热点函数 |
| A16 | cache miss 分析 | 12/22 | ⬜ | perf stat cache-misses |
| A17 | context switch 成本 | 08/22 | ⬜ | 寄存器+缓存污染 |
| A18 | 系统调用开销 | 11/22 | ⬜ | 用户态→内核态 |
| A19 | 磁盘顺序/随机 IO | 11/35 | ⬜ | 随机慢几个数量级 |
| A20 | fsync 持久性 | 11/35 | ⬜ | 刷盘保证 |
| A21 | AIO 概念 | 11 | ⬜ | 异步磁盘 IO |
| A22 | CPU 亲和性 affinity | 08 | ⬜ | 绑核减迁移 |
| A23 | 超线程 HT | 22 | ⬜ | 逻辑核≠物理核 |
| A24 | Little 定律(系统) | 22 | ⬜ | L=λW 排队 |
| A25 | 吞吐 vs 延迟 | 12/22 | ⬜ | Infra 常两者同调 |
| A26 | P99 延迟 | 12 | ⬜ | 尾延迟比均值重要 |
| A27 | 编译优化 -O2/-O3 | 09/12 | ⬜ | 内联/矢量化 |
| A28 | LTO 链接优化 | 09 | ⬜ | 跨 TU 内联 |
| A29 | prefault 预分配 | 18 | ⬜ | 启动时 touch 页面 |
| A30 | DPDK/内核 bypass 概念 | 22/23 | ⬜ | 极致网络可选 |

---

## 8. 高频口述模板（15 题 × 2 分钟）

每题按 §1.1 **四段式**口述。完整 Q&A 见 [14 章](14-高频面试专题与场景题.md)。

| 题 | 结论 | 原理要点 | 项目钩子 |
|----|------|----------|----------|
| T01 栈堆+RAII | 栈自动；堆 RAII | LIFO；new/delete；析构释放 | FdGuard（12） |
| T02 智能指针 | unique/shared/weak | 控制块；weak 破环 | KV 缓存（35） |
| T03 vtable | vptr 查表多态 | virtual→vptr；override | 游戏 Component（21） |
| T04 move+三五 | move 偷资源 | Rule of zero 优先 | HTTP vector 移动（10） |
| T05 vector/umap | 顺序 vs 键值 | O(1) 下标 vs 哈希 | LRU 双结构（34） |
| T06 迭代器失效 | vector 全失效 | reallocate 搬迁 | HTTP 解析（10） |
| T07 mutex/atomic | 复合 mutex | 单变量 atomic | 线程池（08） |
| T08 死锁 | 破坏四条件 | 固定锁序 std::lock | KV 模块锁序（35） |
| T09 condition_variable | mutex+谓词 wait | 防虚假唤醒 | worker 等任务（08） |
| T10 线程池 | worker+队列 | 减创建开销 | epoll 提交任务（35） |
| T11 epoll | O(1) 就绪 | vs select 扫描 | RPS 提升（23） |
| T12 TCP 握手 | SYN 同步序号 | 三次防旧连接 | TIME_WAIT（35） |
| T13 false sharing | 同 line 互 invalid | alignas 64 | thread_local 统计 |
| T14 sendfile | 内核直传 | 少用户态拷贝 | 静态文件（23） |
| T15 WAL | 先日志后内存 | replay 恢复 | KV Put 流程（35） |

---

## 9. 模块自评汇总表

| 模块 | 条数 | ⬜ | 🔶 | ✅ | 目标 |
|------|------|----|----|-----|------|
| §2 内存 | 42 | | | | ≥30 🔶 |
| §3 OOP | 38 | | | | ≥28 🔶 |
| §4 STL | 40 | | | | ≥30 🔶 |
| §5 并发 | 40 | | | | ≥30 🔶 |
| §6 网络 | 35 | | | | ≥25 🔶 |
| §7 体系结构 | 30 | | | | ≥20 🔶 |
| **合计** | **225** | | | | **≥163 🔶** |

---

## 10. 常见问题 FAQ

1. **225 条都要背吗？** 不必；⬜ 跳章重学，🔶 面试前口述，✅ 能带项目。
2. **和 14 章重复怎么办？** 14 深度；33 广度索引；时间紧先 14 再 33 勾选。
3. **Infra 岗最重哪块？** 并发+网络+体系结构；游戏加 OOP+内存。
4. **LLM Infra 还要什么？** 本章 + LLMInfra CUDA/推理/调度。
5. **口述卡壳怎么办？** 用 §8 四段式；结论先说。
6. **体系结构要学到哪？** A01～A05 必会；A14+ 了解即可。
7. **八股和手撕比例？** 面经约 40% 八股 + 40% 手撕 + 20% 项目。
8. **如何验证会讲？** 录音 2 min，无「嗯啊」且能画一张图。
9. **项目没做过怎么办？** 跟做 [35 KV-Store](35-项目实战高性能KV-Store.md)。
10. **Java 背景怎么转？** 对照 §8 T07/T11；强调 RAII 与 UB。

---

## 11. 闭卷自测

1. 本表共多少条索引？分几个模块？
2. 口述四段式哪四段？
3. M11 与 M13 区别？各一句场景。
4. O05 画对象含 vptr 的 ASCII 布局。
5. S38 LRU 两个 STL 容器是什么？各 O(?) 操作。
6. C19 死锁四条件？破坏「循环等待」常用法？
7. N10 与 N11 LT/ET 读 socket 差异？
8. A04 false sharing 一句话 + 一条优化。
9. T11 epoll 相对 select 两点优势。
10. 综合：2 分钟口述 T15 WAL（用 STAR 结尾）。

### 自测参考答案

1. 225 条；6 模块（§2～§7）。
2. 结论→原理→代码/图→项目。
3. unique 独占（文件句柄）；weak 不增计数（缓存父指针）。
4. 例：`+0 vptr +8 data...`（实现相关，说清 vptr 在对象前）。
5. `list` 维护顺序 + `unordered_map` 查迭代器；get/put 均摊 O(1)。
6. 互斥/占有且等待/不可剥夺/循环等待；固定锁序或 std::lock。
7. LT 未读完继续通知；ET 只通知一次需读到 EAGAIN。
8. 多核写同 cache line；alignas/padding/thread_local 分离。
9. O(1) 就绪、无 1024 上限（示例两点即可）。
10. 结论 WAL 先日志；原理 crash 重放；项目 KV-Store Put 流程；结果可恢复。

---

## 12. 学完标准

- [ ] 225 条中 ≥163 条达到 🔶
- [ ] §8 十五模板能闭卷讲 ≥10 题
- [ ] 能 5 分钟画 epoll+线程池+WAL 数据流
- [ ] 面试前按 §10 速过一遍
- [ ] 弱项全部定位到 01～23 具体章
- [ ] 结合 35 章能讲 1 个 KV-Store STAR
- [ ] §12 闭卷自测 ≥8/10

---

## 13. 下一章

- 手撕代码：[34-手撕代码TOP50与白板专题.md](34-手撕代码TOP50与白板专题.md)
- 综合项目：[35-项目实战高性能KV-Store.md](35-项目实战高性能KV-Store.md)
- STAR 简历：[36-面试STAR表达与简历手册.md](36-面试STAR表达与简历手册.md)
