# Redis 数据结构与缓存面试专章

> **文件编码**：UTF-8。五结构、RDB/AOF、集群、穿透击穿雪崩、分布式锁、一致性。[51 MySQL](51-MySQL原理与索引事务面试专章.md) + [08/10/23/25/35]

---


## 本章与前后章的关系

| 上一章 | 本章 | 下一章 |
|--------|------|--------|
| 51 MySQL | **52** | 53 OS |


**学习链**：50 → [51 MySQL](51-MySQL原理与索引事务面试专章.md) → [52 Redis](52-Redis数据结构与缓存面试专章.md) → [53 OS](53-操作系统面试八股与口述模板.md) → [54 计网](54-计算机网络TCP与HTTP面试深度专章.md)


**交叉阅读**：[08](08-多线程与并发编程.md)、[23](23-IO多路复用与高性能Server.md)、[25 无锁](25-无锁编程与内存序.md)、[35 KV-Store](35-项目实战高性能KV-Store.md)

---


## §0 读前导读


### §0.1 用一句话弄懂本章

Redis = **单线程命令执行 + IO 多路复用 + 内存五结构**；面试重点：底层编码、RDB/AOF、集群、穿透/击穿/雪崩、分布式锁与 MySQL 缓存一致性。


### §0.2 你需要提前知道什么

| 状态 | 动作 |
|------|------|
| 只会用不会讲 | 每节 Q&A 限时 2min 口述 |
| C++ 后端岗 | 必串 [08](08-多线程与并发编程.md) [10](10-网络编程与简易HTTP服务.md) [23](23-IO多路复用与高性能Server.md) |
| 计网薄弱 | [计网 02](../../前端学习/计算机网络/02-TCP与UDP.md) 并行 |


### §0.3 本章知识地图（☐→☑）


- ☐ 模块 1 能闭卷口述

- ☐ 模块 2 能闭卷口述

- ☐ 模块 3 能闭卷口述

- ☐ 模块 4 能闭卷口述

- ☐ 模块 5 能闭卷口述

- ☐ 模块 6 能闭卷口述

- ☐ 模块 7 能闭卷口述

- ☐ 模块 8 能闭卷口述


### §0.4 建议节奏

| 阶段 | 时长 | 内容 |
|------|------|------|
| 首轮通读 | 3h | §1～§N Q&A |
| 二轮口述 | 2h | 口述稿录音 |
| 交叉刷 | 2h | C++/计网文档 |
| 闭卷自测 | 1h | ≥8/10 |


```mermaid
flowchart LR
    A[理论八股] --> B[口述 2min]
    B --> C[C++ 工程场景]
    C --> D[闭卷自测]
```

---


## §1 核心面试 Q&A 组 1


| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [1] String SDS | O(1) len 二进制安全 | embstr/raw |
| Q2 | [2] quicklist | 链表+ziplist | List |
| Q3 | [3] Hash dict/ziplist | 渐进 rehash | 小对象压缩 |
| Q4 | [4] ZSet skiplist | O(logN) 范围 | +dict |
| Q5 | [5] 单线程 | 无锁 epoll | 6.0 IO 线程 |
| Q6 | [6] RDB | fork COW 快照 | 丢数据窗口 |
| Q7 | [7] AOF | 追加 rewrite | everysec |
| Q8 | [8] Cluster | 16384 slot | MOVED |
| Q9 | [9] 穿透 | 布隆/空值 | 恶意 |
| Q10 | [10] 击穿 | 互斥/逻辑过期 | 热 key |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [11] 雪崩 | 随机 TTL/集群 | 宕机 |
| Q2 | [12] Cache Aside | 读填写删 | 最常用 |
| Q3 | [13] 分布式锁 | SET NX EX + Lua 删 | 续期 |
| Q4 | [14] Pipeline | 减 RTT | 非原子 |
| Q5 | [15] hiredis 线程 | 每线程连接 | [08章](08-多线程与并发编程.md) |
| Q6 | [16] 与 MySQL 一致 | 先 DB 删缓存 Canal | [51章](51-MySQL原理与索引事务面试专章.md) |
| Q7 | [17] bigkey | UNLINK lazy free | --bigkeys |
| Q8 | [18] 热 key | local cache | 监控 |
| Q9 | [19] Stream | 消费组 ACK | 替代 PubSub |
| Q10 | [20] fork 卡顿 | COW THP | [53 OS](53-操作系统面试八股与口述模板.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [21] String SDS | O(1) len 二进制安全 | embstr/raw |
| Q2 | [22] quicklist | 链表+ziplist | List |
| Q3 | [23] Hash dict/ziplist | 渐进 rehash | 小对象压缩 |
| Q4 | [24] ZSet skiplist | O(logN) 范围 | +dict |
| Q5 | [25] 单线程 | 无锁 epoll | 6.0 IO 线程 |
| Q6 | [26] RDB | fork COW 快照 | 丢数据窗口 |
| Q7 | [27] AOF | 追加 rewrite | everysec |
| Q8 | [28] Cluster | 16384 slot | MOVED |
| Q9 | [29] 穿透 | 布隆/空值 | 恶意 |
| Q10 | [30] 击穿 | 互斥/逻辑过期 | 热 key |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [31] 雪崩 | 随机 TTL/集群 | 宕机 |
| Q2 | [32] Cache Aside | 读填写删 | 最常用 |
| Q3 | [33] 分布式锁 | SET NX EX + Lua 删 | 续期 |
| Q4 | [34] Pipeline | 减 RTT | 非原子 |
| Q5 | [35] hiredis 线程 | 每线程连接 | [08章](08-多线程与并发编程.md) |
| Q6 | [36] 与 MySQL 一致 | 先 DB 删缓存 Canal | [51章](51-MySQL原理与索引事务面试专章.md) |
| Q7 | [37] bigkey | UNLINK lazy free | --bigkeys |
| Q8 | [38] 热 key | local cache | 监控 |
| Q9 | [39] Stream | 消费组 ACK | 替代 PubSub |
| Q10 | [40] fork 卡顿 | COW THP | [53 OS](53-操作系统面试八股与口述模板.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [41] String SDS | O(1) len 二进制安全 | embstr/raw |
| Q2 | [42] quicklist | 链表+ziplist | List |
| Q3 | [43] Hash dict/ziplist | 渐进 rehash | 小对象压缩 |
| Q4 | [44] ZSet skiplist | O(logN) 范围 | +dict |
| Q5 | [45] 单线程 | 无锁 epoll | 6.0 IO 线程 |
| Q6 | [46] RDB | fork COW 快照 | 丢数据窗口 |
| Q7 | [47] AOF | 追加 rewrite | everysec |
| Q8 | [48] Cluster | 16384 slot | MOVED |
| Q9 | [49] 穿透 | 布隆/空值 | 恶意 |
| Q10 | [50] 击穿 | 互斥/逻辑过期 | 热 key |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [51] 雪崩 | 随机 TTL/集群 | 宕机 |
| Q2 | [52] Cache Aside | 读填写删 | 最常用 |
| Q3 | [53] 分布式锁 | SET NX EX + Lua 删 | 续期 |
| Q4 | [54] Pipeline | 减 RTT | 非原子 |
| Q5 | [55] hiredis 线程 | 每线程连接 | [08章](08-多线程与并发编程.md) |
| Q6 | [56] 与 MySQL 一致 | 先 DB 删缓存 Canal | [51章](51-MySQL原理与索引事务面试专章.md) |
| Q7 | [57] bigkey | UNLINK lazy free | --bigkeys |
| Q8 | [58] 热 key | local cache | 监控 |
| Q9 | [59] Stream | 消费组 ACK | 替代 PubSub |
| Q10 | [60] fork 卡顿 | COW THP | [53 OS](53-操作系统面试八股与口述模板.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [61] String SDS | O(1) len 二进制安全 | embstr/raw |
| Q2 | [62] quicklist | 链表+ziplist | List |
| Q3 | [63] Hash dict/ziplist | 渐进 rehash | 小对象压缩 |
| Q4 | [64] ZSet skiplist | O(logN) 范围 | +dict |
| Q5 | [65] 单线程 | 无锁 epoll | 6.0 IO 线程 |
| Q6 | [66] RDB | fork COW 快照 | 丢数据窗口 |
| Q7 | [67] AOF | 追加 rewrite | everysec |
| Q8 | [68] Cluster | 16384 slot | MOVED |
| Q9 | [69] 穿透 | 布隆/空值 | 恶意 |
| Q10 | [70] 击穿 | 互斥/逻辑过期 | 热 key |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [71] 雪崩 | 随机 TTL/集群 | 宕机 |
| Q2 | [72] Cache Aside | 读填写删 | 最常用 |
| Q3 | [73] 分布式锁 | SET NX EX + Lua 删 | 续期 |
| Q4 | [74] Pipeline | 减 RTT | 非原子 |
| Q5 | [75] hiredis 线程 | 每线程连接 | [08章](08-多线程与并发编程.md) |
| Q6 | [76] 与 MySQL 一致 | 先 DB 删缓存 Canal | [51章](51-MySQL原理与索引事务面试专章.md) |
| Q7 | [77] bigkey | UNLINK lazy free | --bigkeys |
| Q8 | [78] 热 key | local cache | 监控 |
| Q9 | [79] Stream | 消费组 ACK | 替代 PubSub |
| Q10 | [80] fork 卡顿 | COW THP | [53 OS](53-操作系统面试八股与口述模板.md) |

## §2 核心面试 Q&A 组 2


| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [1] String SDS | O(1) len 二进制安全 | embstr/raw |
| Q2 | [2] quicklist | 链表+ziplist | List |
| Q3 | [3] Hash dict/ziplist | 渐进 rehash | 小对象压缩 |
| Q4 | [4] ZSet skiplist | O(logN) 范围 | +dict |
| Q5 | [5] 单线程 | 无锁 epoll | 6.0 IO 线程 |
| Q6 | [6] RDB | fork COW 快照 | 丢数据窗口 |
| Q7 | [7] AOF | 追加 rewrite | everysec |
| Q8 | [8] Cluster | 16384 slot | MOVED |
| Q9 | [9] 穿透 | 布隆/空值 | 恶意 |
| Q10 | [10] 击穿 | 互斥/逻辑过期 | 热 key |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [11] 雪崩 | 随机 TTL/集群 | 宕机 |
| Q2 | [12] Cache Aside | 读填写删 | 最常用 |
| Q3 | [13] 分布式锁 | SET NX EX + Lua 删 | 续期 |
| Q4 | [14] Pipeline | 减 RTT | 非原子 |
| Q5 | [15] hiredis 线程 | 每线程连接 | [08章](08-多线程与并发编程.md) |
| Q6 | [16] 与 MySQL 一致 | 先 DB 删缓存 Canal | [51章](51-MySQL原理与索引事务面试专章.md) |
| Q7 | [17] bigkey | UNLINK lazy free | --bigkeys |
| Q8 | [18] 热 key | local cache | 监控 |
| Q9 | [19] Stream | 消费组 ACK | 替代 PubSub |
| Q10 | [20] fork 卡顿 | COW THP | [53 OS](53-操作系统面试八股与口述模板.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [21] String SDS | O(1) len 二进制安全 | embstr/raw |
| Q2 | [22] quicklist | 链表+ziplist | List |
| Q3 | [23] Hash dict/ziplist | 渐进 rehash | 小对象压缩 |
| Q4 | [24] ZSet skiplist | O(logN) 范围 | +dict |
| Q5 | [25] 单线程 | 无锁 epoll | 6.0 IO 线程 |
| Q6 | [26] RDB | fork COW 快照 | 丢数据窗口 |
| Q7 | [27] AOF | 追加 rewrite | everysec |
| Q8 | [28] Cluster | 16384 slot | MOVED |
| Q9 | [29] 穿透 | 布隆/空值 | 恶意 |
| Q10 | [30] 击穿 | 互斥/逻辑过期 | 热 key |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [31] 雪崩 | 随机 TTL/集群 | 宕机 |
| Q2 | [32] Cache Aside | 读填写删 | 最常用 |
| Q3 | [33] 分布式锁 | SET NX EX + Lua 删 | 续期 |
| Q4 | [34] Pipeline | 减 RTT | 非原子 |
| Q5 | [35] hiredis 线程 | 每线程连接 | [08章](08-多线程与并发编程.md) |
| Q6 | [36] 与 MySQL 一致 | 先 DB 删缓存 Canal | [51章](51-MySQL原理与索引事务面试专章.md) |
| Q7 | [37] bigkey | UNLINK lazy free | --bigkeys |
| Q8 | [38] 热 key | local cache | 监控 |
| Q9 | [39] Stream | 消费组 ACK | 替代 PubSub |
| Q10 | [40] fork 卡顿 | COW THP | [53 OS](53-操作系统面试八股与口述模板.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [41] String SDS | O(1) len 二进制安全 | embstr/raw |
| Q2 | [42] quicklist | 链表+ziplist | List |
| Q3 | [43] Hash dict/ziplist | 渐进 rehash | 小对象压缩 |
| Q4 | [44] ZSet skiplist | O(logN) 范围 | +dict |
| Q5 | [45] 单线程 | 无锁 epoll | 6.0 IO 线程 |
| Q6 | [46] RDB | fork COW 快照 | 丢数据窗口 |
| Q7 | [47] AOF | 追加 rewrite | everysec |
| Q8 | [48] Cluster | 16384 slot | MOVED |
| Q9 | [49] 穿透 | 布隆/空值 | 恶意 |
| Q10 | [50] 击穿 | 互斥/逻辑过期 | 热 key |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [51] 雪崩 | 随机 TTL/集群 | 宕机 |
| Q2 | [52] Cache Aside | 读填写删 | 最常用 |
| Q3 | [53] 分布式锁 | SET NX EX + Lua 删 | 续期 |
| Q4 | [54] Pipeline | 减 RTT | 非原子 |
| Q5 | [55] hiredis 线程 | 每线程连接 | [08章](08-多线程与并发编程.md) |
| Q6 | [56] 与 MySQL 一致 | 先 DB 删缓存 Canal | [51章](51-MySQL原理与索引事务面试专章.md) |
| Q7 | [57] bigkey | UNLINK lazy free | --bigkeys |
| Q8 | [58] 热 key | local cache | 监控 |
| Q9 | [59] Stream | 消费组 ACK | 替代 PubSub |
| Q10 | [60] fork 卡顿 | COW THP | [53 OS](53-操作系统面试八股与口述模板.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [61] String SDS | O(1) len 二进制安全 | embstr/raw |
| Q2 | [62] quicklist | 链表+ziplist | List |
| Q3 | [63] Hash dict/ziplist | 渐进 rehash | 小对象压缩 |
| Q4 | [64] ZSet skiplist | O(logN) 范围 | +dict |
| Q5 | [65] 单线程 | 无锁 epoll | 6.0 IO 线程 |
| Q6 | [66] RDB | fork COW 快照 | 丢数据窗口 |
| Q7 | [67] AOF | 追加 rewrite | everysec |
| Q8 | [68] Cluster | 16384 slot | MOVED |
| Q9 | [69] 穿透 | 布隆/空值 | 恶意 |
| Q10 | [70] 击穿 | 互斥/逻辑过期 | 热 key |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [71] 雪崩 | 随机 TTL/集群 | 宕机 |
| Q2 | [72] Cache Aside | 读填写删 | 最常用 |
| Q3 | [73] 分布式锁 | SET NX EX + Lua 删 | 续期 |
| Q4 | [74] Pipeline | 减 RTT | 非原子 |
| Q5 | [75] hiredis 线程 | 每线程连接 | [08章](08-多线程与并发编程.md) |
| Q6 | [76] 与 MySQL 一致 | 先 DB 删缓存 Canal | [51章](51-MySQL原理与索引事务面试专章.md) |
| Q7 | [77] bigkey | UNLINK lazy free | --bigkeys |
| Q8 | [78] 热 key | local cache | 监控 |
| Q9 | [79] Stream | 消费组 ACK | 替代 PubSub |
| Q10 | [80] fork 卡顿 | COW THP | [53 OS](53-操作系统面试八股与口述模板.md) |

## §3 核心面试 Q&A 组 3


| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [1] String SDS | O(1) len 二进制安全 | embstr/raw |
| Q2 | [2] quicklist | 链表+ziplist | List |
| Q3 | [3] Hash dict/ziplist | 渐进 rehash | 小对象压缩 |
| Q4 | [4] ZSet skiplist | O(logN) 范围 | +dict |
| Q5 | [5] 单线程 | 无锁 epoll | 6.0 IO 线程 |
| Q6 | [6] RDB | fork COW 快照 | 丢数据窗口 |
| Q7 | [7] AOF | 追加 rewrite | everysec |
| Q8 | [8] Cluster | 16384 slot | MOVED |
| Q9 | [9] 穿透 | 布隆/空值 | 恶意 |
| Q10 | [10] 击穿 | 互斥/逻辑过期 | 热 key |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [11] 雪崩 | 随机 TTL/集群 | 宕机 |
| Q2 | [12] Cache Aside | 读填写删 | 最常用 |
| Q3 | [13] 分布式锁 | SET NX EX + Lua 删 | 续期 |
| Q4 | [14] Pipeline | 减 RTT | 非原子 |
| Q5 | [15] hiredis 线程 | 每线程连接 | [08章](08-多线程与并发编程.md) |
| Q6 | [16] 与 MySQL 一致 | 先 DB 删缓存 Canal | [51章](51-MySQL原理与索引事务面试专章.md) |
| Q7 | [17] bigkey | UNLINK lazy free | --bigkeys |
| Q8 | [18] 热 key | local cache | 监控 |
| Q9 | [19] Stream | 消费组 ACK | 替代 PubSub |
| Q10 | [20] fork 卡顿 | COW THP | [53 OS](53-操作系统面试八股与口述模板.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [21] String SDS | O(1) len 二进制安全 | embstr/raw |
| Q2 | [22] quicklist | 链表+ziplist | List |
| Q3 | [23] Hash dict/ziplist | 渐进 rehash | 小对象压缩 |
| Q4 | [24] ZSet skiplist | O(logN) 范围 | +dict |
| Q5 | [25] 单线程 | 无锁 epoll | 6.0 IO 线程 |
| Q6 | [26] RDB | fork COW 快照 | 丢数据窗口 |
| Q7 | [27] AOF | 追加 rewrite | everysec |
| Q8 | [28] Cluster | 16384 slot | MOVED |
| Q9 | [29] 穿透 | 布隆/空值 | 恶意 |
| Q10 | [30] 击穿 | 互斥/逻辑过期 | 热 key |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [31] 雪崩 | 随机 TTL/集群 | 宕机 |
| Q2 | [32] Cache Aside | 读填写删 | 最常用 |
| Q3 | [33] 分布式锁 | SET NX EX + Lua 删 | 续期 |
| Q4 | [34] Pipeline | 减 RTT | 非原子 |
| Q5 | [35] hiredis 线程 | 每线程连接 | [08章](08-多线程与并发编程.md) |
| Q6 | [36] 与 MySQL 一致 | 先 DB 删缓存 Canal | [51章](51-MySQL原理与索引事务面试专章.md) |
| Q7 | [37] bigkey | UNLINK lazy free | --bigkeys |
| Q8 | [38] 热 key | local cache | 监控 |
| Q9 | [39] Stream | 消费组 ACK | 替代 PubSub |
| Q10 | [40] fork 卡顿 | COW THP | [53 OS](53-操作系统面试八股与口述模板.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [41] String SDS | O(1) len 二进制安全 | embstr/raw |
| Q2 | [42] quicklist | 链表+ziplist | List |
| Q3 | [43] Hash dict/ziplist | 渐进 rehash | 小对象压缩 |
| Q4 | [44] ZSet skiplist | O(logN) 范围 | +dict |
| Q5 | [45] 单线程 | 无锁 epoll | 6.0 IO 线程 |
| Q6 | [46] RDB | fork COW 快照 | 丢数据窗口 |
| Q7 | [47] AOF | 追加 rewrite | everysec |
| Q8 | [48] Cluster | 16384 slot | MOVED |
| Q9 | [49] 穿透 | 布隆/空值 | 恶意 |
| Q10 | [50] 击穿 | 互斥/逻辑过期 | 热 key |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [51] 雪崩 | 随机 TTL/集群 | 宕机 |
| Q2 | [52] Cache Aside | 读填写删 | 最常用 |
| Q3 | [53] 分布式锁 | SET NX EX + Lua 删 | 续期 |
| Q4 | [54] Pipeline | 减 RTT | 非原子 |
| Q5 | [55] hiredis 线程 | 每线程连接 | [08章](08-多线程与并发编程.md) |
| Q6 | [56] 与 MySQL 一致 | 先 DB 删缓存 Canal | [51章](51-MySQL原理与索引事务面试专章.md) |
| Q7 | [57] bigkey | UNLINK lazy free | --bigkeys |
| Q8 | [58] 热 key | local cache | 监控 |
| Q9 | [59] Stream | 消费组 ACK | 替代 PubSub |
| Q10 | [60] fork 卡顿 | COW THP | [53 OS](53-操作系统面试八股与口述模板.md) |

## §4 大厂口述模板（STAR / 四段式）


### 口述稿：索引优化订单查询

> **结论**：联合索引覆盖 WHERE+ORDER BY

**原理要点**：

  - 最左前缀匹配 WHERE 列
  - 覆盖索引避免回表
  - EXPLAIN 确认 type=ref/range

**工程实践**：C++ 服务用 PreparedStatement + 连接池 [23章](23-IO多路复用与高性能Server.md)

**结果/风险**：P99 从 800ms 到 50ms


### 口述稿：RR 下防幻读

> **结论**：Next-Key Lock + MVCC

**原理要点**：

  - 快照读不加锁
  - 当前读 FOR UPDATE 锁 gap
  - 短事务减锁持有

**工程实践**：库存扣减 SELECT FOR UPDATE 走主键

**结果/风险**：无超卖


### 口述稿：Cache Aside 一致性

> **结论**：先更 MySQL 再删 Redis

**原理要点**：

  - 延迟双删
  - Canal 订阅 binlog
  - 读走主或容忍延迟

**工程实践**：与 [52 Redis](52-Redis数据结构与缓存面试专章.md) 配合

**结果/风险**：最终一致


### 口述稿：epoll 改造 HTTP 服务

> **结论**：Reactor + 线程池

**原理要点**：

  - LT 模式安全
  - accept 与 worker 分离
  - DB 查询不在 IO 线程

**工程实践**：[10 mini-http](10-网络编程与简易HTTP服务.md) 演进 [23章](23-IO多路复用与高性能Server.md)

**结果/风险**：并发 1w 连接


### 口述稿：TIME_WAIT 排查

> **结论**：短连接改长连接+连接池

**原理要点**：

  - 客户端 reuse
  - 服务端避免主动 close
  - ss 看状态分布

**工程实践**：C++ HTTP client 池化

**结果/风险**：端口耗尽解决


## §5 术语速查表 1


| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| SDS | Redis 内存 KV 面试术语 1 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| quicklist | Redis 内存 KV 面试术语 2 | [23 IO](23-IO多路复用与高性能Server.md) |
| skiplist | Redis 内存 KV 面试术语 3 | [53 OS](53-操作系统面试八股与口述模板.md) |
| intset | Redis 内存 KV 面试术语 4 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| dict | Redis 内存 KV 面试术语 5 | [23 IO](23-IO多路复用与高性能Server.md) |
| rehash | Redis 内存 KV 面试术语 6 | [53 OS](53-操作系统面试八股与口述模板.md) |
| RDB | Redis 内存 KV 面试术语 7 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| AOF | Redis 内存 KV 面试术语 8 | [23 IO](23-IO多路复用与高性能Server.md) |
| 混合持久化 | Redis 内存 KV 面试术语 9 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 主从 | Redis 内存 KV 面试术语 10 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 哨兵 | Redis 内存 KV 面试术语 11 | [23 IO](23-IO多路复用与高性能Server.md) |
| Cluster | Redis 内存 KV 面试术语 12 | [53 OS](53-操作系统面试八股与口述模板.md) |
| slot | Redis 内存 KV 面试术语 13 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| MOVED | Redis 内存 KV 面试术语 14 | [23 IO](23-IO多路复用与高性能Server.md) |
| ASK | Redis 内存 KV 面试术语 15 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| 穿透 | Redis 内存 KV 面试术语 16 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 击穿 | Redis 内存 KV 面试术语 17 | [23 IO](23-IO多路复用与高性能Server.md) |
| 雪崩 | Redis 内存 KV 面试术语 18 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 布隆 | Redis 内存 KV 面试术语 19 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| CacheAside | Redis 内存 KV 面试术语 20 | [23 IO](23-IO多路复用与高性能Server.md) |
| 分布式锁 | Redis 内存 KV 面试术语 21 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Redlock | Redis 内存 KV 面试术语 22 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Pipeline | Redis 内存 KV 面试术语 23 | [23 IO](23-IO多路复用与高性能Server.md) |
| MULTI | Redis 内存 KV 面试术语 24 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Lua | Redis 内存 KV 面试术语 25 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Stream | Redis 内存 KV 面试术语 26 | [23 IO](23-IO多路复用与高性能Server.md) |
| PubSub | Redis 内存 KV 面试术语 27 | [53 OS](53-操作系统面试八股与口述模板.md) |
| lazyfree | Redis 内存 KV 面试术语 28 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| bigkey | Redis 内存 KV 面试术语 29 | [23 IO](23-IO多路复用与高性能Server.md) |
| 热key | Redis 内存 KV 面试术语 30 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| maxmemory | Redis 内存 KV 面试术语 31 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| LRU | Redis 内存 KV 面试术语 32 | [23 IO](23-IO多路复用与高性能Server.md) |
| LFU | Redis 内存 KV 面试术语 33 | [53 OS](53-操作系统面试八股与口述模板.md) |
| jemalloc | Redis 内存 KV 面试术语 34 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| fork | Redis 内存 KV 面试术语 35 | [23 IO](23-IO多路复用与高性能Server.md) |
| COW | Redis 内存 KV 面试术语 36 | [53 OS](53-操作系统面试八股与口述模板.md) |
| RESP | Redis 内存 KV 面试术语 37 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| IO线程 | Redis 内存 KV 面试术语 38 | [23 IO](23-IO多路复用与高性能Server.md) |
| ACL | Redis 内存 KV 面试术语 39 | [53 OS](53-操作系统面试八股与口述模板.md) |
| TLS | Redis 内存 KV 面试术语 40 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| SDS-40 | Redis 内存 KV 面试术语 41 | [23 IO](23-IO多路复用与高性能Server.md) |
| quicklist-41 | Redis 内存 KV 面试术语 42 | [53 OS](53-操作系统面试八股与口述模板.md) |
| skiplist-42 | Redis 内存 KV 面试术语 43 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| intset-43 | Redis 内存 KV 面试术语 44 | [23 IO](23-IO多路复用与高性能Server.md) |
| dict-44 | Redis 内存 KV 面试术语 45 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| rehash-45 | Redis 内存 KV 面试术语 46 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| RDB-46 | Redis 内存 KV 面试术语 47 | [23 IO](23-IO多路复用与高性能Server.md) |
| AOF-47 | Redis 内存 KV 面试术语 48 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 混合持久化-48 | Redis 内存 KV 面试术语 49 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 主从-49 | Redis 内存 KV 面试术语 50 | [23 IO](23-IO多路复用与高性能Server.md) |
| 哨兵-50 | Redis 内存 KV 面试术语 51 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Cluster-51 | Redis 内存 KV 面试术语 52 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| slot-52 | Redis 内存 KV 面试术语 53 | [23 IO](23-IO多路复用与高性能Server.md) |
| MOVED-53 | Redis 内存 KV 面试术语 54 | [53 OS](53-操作系统面试八股与口述模板.md) |
| ASK-54 | Redis 内存 KV 面试术语 55 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 穿透-55 | Redis 内存 KV 面试术语 56 | [23 IO](23-IO多路复用与高性能Server.md) |
| 击穿-56 | Redis 内存 KV 面试术语 57 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 雪崩-57 | Redis 内存 KV 面试术语 58 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 布隆-58 | Redis 内存 KV 面试术语 59 | [23 IO](23-IO多路复用与高性能Server.md) |
| CacheAside-59 | Redis 内存 KV 面试术语 60 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| 分布式锁-60 | Redis 内存 KV 面试术语 61 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Redlock-61 | Redis 内存 KV 面试术语 62 | [23 IO](23-IO多路复用与高性能Server.md) |
| Pipeline-62 | Redis 内存 KV 面试术语 63 | [53 OS](53-操作系统面试八股与口述模板.md) |
| MULTI-63 | Redis 内存 KV 面试术语 64 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Lua-64 | Redis 内存 KV 面试术语 65 | [23 IO](23-IO多路复用与高性能Server.md) |
| Stream-65 | Redis 内存 KV 面试术语 66 | [53 OS](53-操作系统面试八股与口述模板.md) |
| PubSub-66 | Redis 内存 KV 面试术语 67 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| lazyfree-67 | Redis 内存 KV 面试术语 68 | [23 IO](23-IO多路复用与高性能Server.md) |
| bigkey-68 | Redis 内存 KV 面试术语 69 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 热key-69 | Redis 内存 KV 面试术语 70 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| maxmemory-70 | Redis 内存 KV 面试术语 71 | [23 IO](23-IO多路复用与高性能Server.md) |
| LRU-71 | Redis 内存 KV 面试术语 72 | [53 OS](53-操作系统面试八股与口述模板.md) |
| LFU-72 | Redis 内存 KV 面试术语 73 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| jemalloc-73 | Redis 内存 KV 面试术语 74 | [23 IO](23-IO多路复用与高性能Server.md) |
| fork-74 | Redis 内存 KV 面试术语 75 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| COW-75 | Redis 内存 KV 面试术语 76 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| RESP-76 | Redis 内存 KV 面试术语 77 | [23 IO](23-IO多路复用与高性能Server.md) |
| IO线程-77 | Redis 内存 KV 面试术语 78 | [53 OS](53-操作系统面试八股与口述模板.md) |
| ACL-78 | Redis 内存 KV 面试术语 79 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| TLS-79 | Redis 内存 KV 面试术语 80 | [23 IO](23-IO多路复用与高性能Server.md) |
| SDS-80 | Redis 内存 KV 面试术语 81 | [53 OS](53-操作系统面试八股与口述模板.md) |
| quicklist-81 | Redis 内存 KV 面试术语 82 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| skiplist-82 | Redis 内存 KV 面试术语 83 | [23 IO](23-IO多路复用与高性能Server.md) |
| intset-83 | Redis 内存 KV 面试术语 84 | [53 OS](53-操作系统面试八股与口述模板.md) |
| dict-84 | Redis 内存 KV 面试术语 85 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| rehash-85 | Redis 内存 KV 面试术语 86 | [23 IO](23-IO多路复用与高性能Server.md) |
| RDB-86 | Redis 内存 KV 面试术语 87 | [53 OS](53-操作系统面试八股与口述模板.md) |
| AOF-87 | Redis 内存 KV 面试术语 88 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 混合持久化-88 | Redis 内存 KV 面试术语 89 | [23 IO](23-IO多路复用与高性能Server.md) |
| 主从-89 | Redis 内存 KV 面试术语 90 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| 哨兵-90 | Redis 内存 KV 面试术语 91 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Cluster-91 | Redis 内存 KV 面试术语 92 | [23 IO](23-IO多路复用与高性能Server.md) |
| slot-92 | Redis 内存 KV 面试术语 93 | [53 OS](53-操作系统面试八股与口述模板.md) |
| MOVED-93 | Redis 内存 KV 面试术语 94 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| ASK-94 | Redis 内存 KV 面试术语 95 | [23 IO](23-IO多路复用与高性能Server.md) |
| 穿透-95 | Redis 内存 KV 面试术语 96 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 击穿-96 | Redis 内存 KV 面试术语 97 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 雪崩-97 | Redis 内存 KV 面试术语 98 | [23 IO](23-IO多路复用与高性能Server.md) |
| 布隆-98 | Redis 内存 KV 面试术语 99 | [53 OS](53-操作系统面试八股与口述模板.md) |
| CacheAside-99 | Redis 内存 KV 面试术语 100 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 分布式锁-100 | Redis 内存 KV 面试术语 101 | [23 IO](23-IO多路复用与高性能Server.md) |
| Redlock-101 | Redis 内存 KV 面试术语 102 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Pipeline-102 | Redis 内存 KV 面试术语 103 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| MULTI-103 | Redis 内存 KV 面试术语 104 | [23 IO](23-IO多路复用与高性能Server.md) |
| Lua-104 | Redis 内存 KV 面试术语 105 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| Stream-105 | Redis 内存 KV 面试术语 106 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| PubSub-106 | Redis 内存 KV 面试术语 107 | [23 IO](23-IO多路复用与高性能Server.md) |
| lazyfree-107 | Redis 内存 KV 面试术语 108 | [53 OS](53-操作系统面试八股与口述模板.md) |
| bigkey-108 | Redis 内存 KV 面试术语 109 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 热key-109 | Redis 内存 KV 面试术语 110 | [23 IO](23-IO多路复用与高性能Server.md) |
| maxmemory-110 | Redis 内存 KV 面试术语 111 | [53 OS](53-操作系统面试八股与口述模板.md) |
| LRU-111 | Redis 内存 KV 面试术语 112 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| LFU-112 | Redis 内存 KV 面试术语 113 | [23 IO](23-IO多路复用与高性能Server.md) |
| jemalloc-113 | Redis 内存 KV 面试术语 114 | [53 OS](53-操作系统面试八股与口述模板.md) |
| fork-114 | Redis 内存 KV 面试术语 115 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| COW-115 | Redis 内存 KV 面试术语 116 | [23 IO](23-IO多路复用与高性能Server.md) |
| RESP-116 | Redis 内存 KV 面试术语 117 | [53 OS](53-操作系统面试八股与口述模板.md) |
| IO线程-117 | Redis 内存 KV 面试术语 118 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| ACL-118 | Redis 内存 KV 面试术语 119 | [23 IO](23-IO多路复用与高性能Server.md) |
| TLS-119 | Redis 内存 KV 面试术语 120 | [53 OS](53-操作系统面试八股与口述模板.md) |

## §6 术语速查表 2


| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| SDS | Redis 内存 KV 面试术语 1 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| quicklist | Redis 内存 KV 面试术语 2 | [23 IO](23-IO多路复用与高性能Server.md) |
| skiplist | Redis 内存 KV 面试术语 3 | [53 OS](53-操作系统面试八股与口述模板.md) |
| intset | Redis 内存 KV 面试术语 4 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| dict | Redis 内存 KV 面试术语 5 | [23 IO](23-IO多路复用与高性能Server.md) |
| rehash | Redis 内存 KV 面试术语 6 | [53 OS](53-操作系统面试八股与口述模板.md) |
| RDB | Redis 内存 KV 面试术语 7 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| AOF | Redis 内存 KV 面试术语 8 | [23 IO](23-IO多路复用与高性能Server.md) |
| 混合持久化 | Redis 内存 KV 面试术语 9 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 主从 | Redis 内存 KV 面试术语 10 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 哨兵 | Redis 内存 KV 面试术语 11 | [23 IO](23-IO多路复用与高性能Server.md) |
| Cluster | Redis 内存 KV 面试术语 12 | [53 OS](53-操作系统面试八股与口述模板.md) |
| slot | Redis 内存 KV 面试术语 13 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| MOVED | Redis 内存 KV 面试术语 14 | [23 IO](23-IO多路复用与高性能Server.md) |
| ASK | Redis 内存 KV 面试术语 15 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| 穿透 | Redis 内存 KV 面试术语 16 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 击穿 | Redis 内存 KV 面试术语 17 | [23 IO](23-IO多路复用与高性能Server.md) |
| 雪崩 | Redis 内存 KV 面试术语 18 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 布隆 | Redis 内存 KV 面试术语 19 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| CacheAside | Redis 内存 KV 面试术语 20 | [23 IO](23-IO多路复用与高性能Server.md) |
| 分布式锁 | Redis 内存 KV 面试术语 21 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Redlock | Redis 内存 KV 面试术语 22 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Pipeline | Redis 内存 KV 面试术语 23 | [23 IO](23-IO多路复用与高性能Server.md) |
| MULTI | Redis 内存 KV 面试术语 24 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Lua | Redis 内存 KV 面试术语 25 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Stream | Redis 内存 KV 面试术语 26 | [23 IO](23-IO多路复用与高性能Server.md) |
| PubSub | Redis 内存 KV 面试术语 27 | [53 OS](53-操作系统面试八股与口述模板.md) |
| lazyfree | Redis 内存 KV 面试术语 28 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| bigkey | Redis 内存 KV 面试术语 29 | [23 IO](23-IO多路复用与高性能Server.md) |
| 热key | Redis 内存 KV 面试术语 30 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| maxmemory | Redis 内存 KV 面试术语 31 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| LRU | Redis 内存 KV 面试术语 32 | [23 IO](23-IO多路复用与高性能Server.md) |
| LFU | Redis 内存 KV 面试术语 33 | [53 OS](53-操作系统面试八股与口述模板.md) |
| jemalloc | Redis 内存 KV 面试术语 34 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| fork | Redis 内存 KV 面试术语 35 | [23 IO](23-IO多路复用与高性能Server.md) |
| COW | Redis 内存 KV 面试术语 36 | [53 OS](53-操作系统面试八股与口述模板.md) |
| RESP | Redis 内存 KV 面试术语 37 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| IO线程 | Redis 内存 KV 面试术语 38 | [23 IO](23-IO多路复用与高性能Server.md) |
| ACL | Redis 内存 KV 面试术语 39 | [53 OS](53-操作系统面试八股与口述模板.md) |
| TLS | Redis 内存 KV 面试术语 40 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| SDS-40 | Redis 内存 KV 面试术语 41 | [23 IO](23-IO多路复用与高性能Server.md) |
| quicklist-41 | Redis 内存 KV 面试术语 42 | [53 OS](53-操作系统面试八股与口述模板.md) |
| skiplist-42 | Redis 内存 KV 面试术语 43 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| intset-43 | Redis 内存 KV 面试术语 44 | [23 IO](23-IO多路复用与高性能Server.md) |
| dict-44 | Redis 内存 KV 面试术语 45 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| rehash-45 | Redis 内存 KV 面试术语 46 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| RDB-46 | Redis 内存 KV 面试术语 47 | [23 IO](23-IO多路复用与高性能Server.md) |
| AOF-47 | Redis 内存 KV 面试术语 48 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 混合持久化-48 | Redis 内存 KV 面试术语 49 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 主从-49 | Redis 内存 KV 面试术语 50 | [23 IO](23-IO多路复用与高性能Server.md) |
| 哨兵-50 | Redis 内存 KV 面试术语 51 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Cluster-51 | Redis 内存 KV 面试术语 52 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| slot-52 | Redis 内存 KV 面试术语 53 | [23 IO](23-IO多路复用与高性能Server.md) |
| MOVED-53 | Redis 内存 KV 面试术语 54 | [53 OS](53-操作系统面试八股与口述模板.md) |
| ASK-54 | Redis 内存 KV 面试术语 55 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 穿透-55 | Redis 内存 KV 面试术语 56 | [23 IO](23-IO多路复用与高性能Server.md) |
| 击穿-56 | Redis 内存 KV 面试术语 57 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 雪崩-57 | Redis 内存 KV 面试术语 58 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 布隆-58 | Redis 内存 KV 面试术语 59 | [23 IO](23-IO多路复用与高性能Server.md) |
| CacheAside-59 | Redis 内存 KV 面试术语 60 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| 分布式锁-60 | Redis 内存 KV 面试术语 61 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Redlock-61 | Redis 内存 KV 面试术语 62 | [23 IO](23-IO多路复用与高性能Server.md) |
| Pipeline-62 | Redis 内存 KV 面试术语 63 | [53 OS](53-操作系统面试八股与口述模板.md) |
| MULTI-63 | Redis 内存 KV 面试术语 64 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Lua-64 | Redis 内存 KV 面试术语 65 | [23 IO](23-IO多路复用与高性能Server.md) |
| Stream-65 | Redis 内存 KV 面试术语 66 | [53 OS](53-操作系统面试八股与口述模板.md) |
| PubSub-66 | Redis 内存 KV 面试术语 67 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| lazyfree-67 | Redis 内存 KV 面试术语 68 | [23 IO](23-IO多路复用与高性能Server.md) |
| bigkey-68 | Redis 内存 KV 面试术语 69 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 热key-69 | Redis 内存 KV 面试术语 70 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| maxmemory-70 | Redis 内存 KV 面试术语 71 | [23 IO](23-IO多路复用与高性能Server.md) |
| LRU-71 | Redis 内存 KV 面试术语 72 | [53 OS](53-操作系统面试八股与口述模板.md) |
| LFU-72 | Redis 内存 KV 面试术语 73 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| jemalloc-73 | Redis 内存 KV 面试术语 74 | [23 IO](23-IO多路复用与高性能Server.md) |
| fork-74 | Redis 内存 KV 面试术语 75 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| COW-75 | Redis 内存 KV 面试术语 76 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| RESP-76 | Redis 内存 KV 面试术语 77 | [23 IO](23-IO多路复用与高性能Server.md) |
| IO线程-77 | Redis 内存 KV 面试术语 78 | [53 OS](53-操作系统面试八股与口述模板.md) |
| ACL-78 | Redis 内存 KV 面试术语 79 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| TLS-79 | Redis 内存 KV 面试术语 80 | [23 IO](23-IO多路复用与高性能Server.md) |
| SDS-80 | Redis 内存 KV 面试术语 81 | [53 OS](53-操作系统面试八股与口述模板.md) |
| quicklist-81 | Redis 内存 KV 面试术语 82 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| skiplist-82 | Redis 内存 KV 面试术语 83 | [23 IO](23-IO多路复用与高性能Server.md) |
| intset-83 | Redis 内存 KV 面试术语 84 | [53 OS](53-操作系统面试八股与口述模板.md) |
| dict-84 | Redis 内存 KV 面试术语 85 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| rehash-85 | Redis 内存 KV 面试术语 86 | [23 IO](23-IO多路复用与高性能Server.md) |
| RDB-86 | Redis 内存 KV 面试术语 87 | [53 OS](53-操作系统面试八股与口述模板.md) |
| AOF-87 | Redis 内存 KV 面试术语 88 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 混合持久化-88 | Redis 内存 KV 面试术语 89 | [23 IO](23-IO多路复用与高性能Server.md) |
| 主从-89 | Redis 内存 KV 面试术语 90 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| 哨兵-90 | Redis 内存 KV 面试术语 91 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Cluster-91 | Redis 内存 KV 面试术语 92 | [23 IO](23-IO多路复用与高性能Server.md) |
| slot-92 | Redis 内存 KV 面试术语 93 | [53 OS](53-操作系统面试八股与口述模板.md) |
| MOVED-93 | Redis 内存 KV 面试术语 94 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| ASK-94 | Redis 内存 KV 面试术语 95 | [23 IO](23-IO多路复用与高性能Server.md) |
| 穿透-95 | Redis 内存 KV 面试术语 96 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 击穿-96 | Redis 内存 KV 面试术语 97 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 雪崩-97 | Redis 内存 KV 面试术语 98 | [23 IO](23-IO多路复用与高性能Server.md) |
| 布隆-98 | Redis 内存 KV 面试术语 99 | [53 OS](53-操作系统面试八股与口述模板.md) |
| CacheAside-99 | Redis 内存 KV 面试术语 100 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 分布式锁-100 | Redis 内存 KV 面试术语 101 | [23 IO](23-IO多路复用与高性能Server.md) |
| Redlock-101 | Redis 内存 KV 面试术语 102 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Pipeline-102 | Redis 内存 KV 面试术语 103 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| MULTI-103 | Redis 内存 KV 面试术语 104 | [23 IO](23-IO多路复用与高性能Server.md) |
| Lua-104 | Redis 内存 KV 面试术语 105 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| Stream-105 | Redis 内存 KV 面试术语 106 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| PubSub-106 | Redis 内存 KV 面试术语 107 | [23 IO](23-IO多路复用与高性能Server.md) |
| lazyfree-107 | Redis 内存 KV 面试术语 108 | [53 OS](53-操作系统面试八股与口述模板.md) |
| bigkey-108 | Redis 内存 KV 面试术语 109 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 热key-109 | Redis 内存 KV 面试术语 110 | [23 IO](23-IO多路复用与高性能Server.md) |
| maxmemory-110 | Redis 内存 KV 面试术语 111 | [53 OS](53-操作系统面试八股与口述模板.md) |
| LRU-111 | Redis 内存 KV 面试术语 112 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| LFU-112 | Redis 内存 KV 面试术语 113 | [23 IO](23-IO多路复用与高性能Server.md) |
| jemalloc-113 | Redis 内存 KV 面试术语 114 | [53 OS](53-操作系统面试八股与口述模板.md) |
| fork-114 | Redis 内存 KV 面试术语 115 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| COW-115 | Redis 内存 KV 面试术语 116 | [23 IO](23-IO多路复用与高性能Server.md) |
| RESP-116 | Redis 内存 KV 面试术语 117 | [53 OS](53-操作系统面试八股与口述模板.md) |
| IO线程-117 | Redis 内存 KV 面试术语 118 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| ACL-118 | Redis 内存 KV 面试术语 119 | [23 IO](23-IO多路复用与高性能Server.md) |
| TLS-119 | Redis 内存 KV 面试术语 120 | [53 OS](53-操作系统面试八股与口述模板.md) |

## §7 术语速查表 3


| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| SDS | Redis 内存 KV 面试术语 1 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| quicklist | Redis 内存 KV 面试术语 2 | [23 IO](23-IO多路复用与高性能Server.md) |
| skiplist | Redis 内存 KV 面试术语 3 | [53 OS](53-操作系统面试八股与口述模板.md) |
| intset | Redis 内存 KV 面试术语 4 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| dict | Redis 内存 KV 面试术语 5 | [23 IO](23-IO多路复用与高性能Server.md) |
| rehash | Redis 内存 KV 面试术语 6 | [53 OS](53-操作系统面试八股与口述模板.md) |
| RDB | Redis 内存 KV 面试术语 7 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| AOF | Redis 内存 KV 面试术语 8 | [23 IO](23-IO多路复用与高性能Server.md) |
| 混合持久化 | Redis 内存 KV 面试术语 9 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 主从 | Redis 内存 KV 面试术语 10 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 哨兵 | Redis 内存 KV 面试术语 11 | [23 IO](23-IO多路复用与高性能Server.md) |
| Cluster | Redis 内存 KV 面试术语 12 | [53 OS](53-操作系统面试八股与口述模板.md) |
| slot | Redis 内存 KV 面试术语 13 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| MOVED | Redis 内存 KV 面试术语 14 | [23 IO](23-IO多路复用与高性能Server.md) |
| ASK | Redis 内存 KV 面试术语 15 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| 穿透 | Redis 内存 KV 面试术语 16 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 击穿 | Redis 内存 KV 面试术语 17 | [23 IO](23-IO多路复用与高性能Server.md) |
| 雪崩 | Redis 内存 KV 面试术语 18 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 布隆 | Redis 内存 KV 面试术语 19 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| CacheAside | Redis 内存 KV 面试术语 20 | [23 IO](23-IO多路复用与高性能Server.md) |
| 分布式锁 | Redis 内存 KV 面试术语 21 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Redlock | Redis 内存 KV 面试术语 22 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Pipeline | Redis 内存 KV 面试术语 23 | [23 IO](23-IO多路复用与高性能Server.md) |
| MULTI | Redis 内存 KV 面试术语 24 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Lua | Redis 内存 KV 面试术语 25 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Stream | Redis 内存 KV 面试术语 26 | [23 IO](23-IO多路复用与高性能Server.md) |
| PubSub | Redis 内存 KV 面试术语 27 | [53 OS](53-操作系统面试八股与口述模板.md) |
| lazyfree | Redis 内存 KV 面试术语 28 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| bigkey | Redis 内存 KV 面试术语 29 | [23 IO](23-IO多路复用与高性能Server.md) |
| 热key | Redis 内存 KV 面试术语 30 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| maxmemory | Redis 内存 KV 面试术语 31 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| LRU | Redis 内存 KV 面试术语 32 | [23 IO](23-IO多路复用与高性能Server.md) |
| LFU | Redis 内存 KV 面试术语 33 | [53 OS](53-操作系统面试八股与口述模板.md) |
| jemalloc | Redis 内存 KV 面试术语 34 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| fork | Redis 内存 KV 面试术语 35 | [23 IO](23-IO多路复用与高性能Server.md) |
| COW | Redis 内存 KV 面试术语 36 | [53 OS](53-操作系统面试八股与口述模板.md) |
| RESP | Redis 内存 KV 面试术语 37 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| IO线程 | Redis 内存 KV 面试术语 38 | [23 IO](23-IO多路复用与高性能Server.md) |
| ACL | Redis 内存 KV 面试术语 39 | [53 OS](53-操作系统面试八股与口述模板.md) |
| TLS | Redis 内存 KV 面试术语 40 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| SDS-40 | Redis 内存 KV 面试术语 41 | [23 IO](23-IO多路复用与高性能Server.md) |
| quicklist-41 | Redis 内存 KV 面试术语 42 | [53 OS](53-操作系统面试八股与口述模板.md) |
| skiplist-42 | Redis 内存 KV 面试术语 43 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| intset-43 | Redis 内存 KV 面试术语 44 | [23 IO](23-IO多路复用与高性能Server.md) |
| dict-44 | Redis 内存 KV 面试术语 45 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| rehash-45 | Redis 内存 KV 面试术语 46 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| RDB-46 | Redis 内存 KV 面试术语 47 | [23 IO](23-IO多路复用与高性能Server.md) |
| AOF-47 | Redis 内存 KV 面试术语 48 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 混合持久化-48 | Redis 内存 KV 面试术语 49 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 主从-49 | Redis 内存 KV 面试术语 50 | [23 IO](23-IO多路复用与高性能Server.md) |
| 哨兵-50 | Redis 内存 KV 面试术语 51 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Cluster-51 | Redis 内存 KV 面试术语 52 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| slot-52 | Redis 内存 KV 面试术语 53 | [23 IO](23-IO多路复用与高性能Server.md) |
| MOVED-53 | Redis 内存 KV 面试术语 54 | [53 OS](53-操作系统面试八股与口述模板.md) |
| ASK-54 | Redis 内存 KV 面试术语 55 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 穿透-55 | Redis 内存 KV 面试术语 56 | [23 IO](23-IO多路复用与高性能Server.md) |
| 击穿-56 | Redis 内存 KV 面试术语 57 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 雪崩-57 | Redis 内存 KV 面试术语 58 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 布隆-58 | Redis 内存 KV 面试术语 59 | [23 IO](23-IO多路复用与高性能Server.md) |
| CacheAside-59 | Redis 内存 KV 面试术语 60 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| 分布式锁-60 | Redis 内存 KV 面试术语 61 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Redlock-61 | Redis 内存 KV 面试术语 62 | [23 IO](23-IO多路复用与高性能Server.md) |
| Pipeline-62 | Redis 内存 KV 面试术语 63 | [53 OS](53-操作系统面试八股与口述模板.md) |
| MULTI-63 | Redis 内存 KV 面试术语 64 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Lua-64 | Redis 内存 KV 面试术语 65 | [23 IO](23-IO多路复用与高性能Server.md) |
| Stream-65 | Redis 内存 KV 面试术语 66 | [53 OS](53-操作系统面试八股与口述模板.md) |
| PubSub-66 | Redis 内存 KV 面试术语 67 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| lazyfree-67 | Redis 内存 KV 面试术语 68 | [23 IO](23-IO多路复用与高性能Server.md) |
| bigkey-68 | Redis 内存 KV 面试术语 69 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 热key-69 | Redis 内存 KV 面试术语 70 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| maxmemory-70 | Redis 内存 KV 面试术语 71 | [23 IO](23-IO多路复用与高性能Server.md) |
| LRU-71 | Redis 内存 KV 面试术语 72 | [53 OS](53-操作系统面试八股与口述模板.md) |
| LFU-72 | Redis 内存 KV 面试术语 73 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| jemalloc-73 | Redis 内存 KV 面试术语 74 | [23 IO](23-IO多路复用与高性能Server.md) |
| fork-74 | Redis 内存 KV 面试术语 75 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| COW-75 | Redis 内存 KV 面试术语 76 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| RESP-76 | Redis 内存 KV 面试术语 77 | [23 IO](23-IO多路复用与高性能Server.md) |
| IO线程-77 | Redis 内存 KV 面试术语 78 | [53 OS](53-操作系统面试八股与口述模板.md) |
| ACL-78 | Redis 内存 KV 面试术语 79 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| TLS-79 | Redis 内存 KV 面试术语 80 | [23 IO](23-IO多路复用与高性能Server.md) |
| SDS-80 | Redis 内存 KV 面试术语 81 | [53 OS](53-操作系统面试八股与口述模板.md) |
| quicklist-81 | Redis 内存 KV 面试术语 82 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| skiplist-82 | Redis 内存 KV 面试术语 83 | [23 IO](23-IO多路复用与高性能Server.md) |
| intset-83 | Redis 内存 KV 面试术语 84 | [53 OS](53-操作系统面试八股与口述模板.md) |
| dict-84 | Redis 内存 KV 面试术语 85 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| rehash-85 | Redis 内存 KV 面试术语 86 | [23 IO](23-IO多路复用与高性能Server.md) |
| RDB-86 | Redis 内存 KV 面试术语 87 | [53 OS](53-操作系统面试八股与口述模板.md) |
| AOF-87 | Redis 内存 KV 面试术语 88 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 混合持久化-88 | Redis 内存 KV 面试术语 89 | [23 IO](23-IO多路复用与高性能Server.md) |
| 主从-89 | Redis 内存 KV 面试术语 90 | [53 OS](53-操作系统面试八股与口述模板.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| 哨兵-90 | Redis 内存 KV 面试术语 91 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Cluster-91 | Redis 内存 KV 面试术语 92 | [23 IO](23-IO多路复用与高性能Server.md) |
| slot-92 | Redis 内存 KV 面试术语 93 | [53 OS](53-操作系统面试八股与口述模板.md) |
| MOVED-93 | Redis 内存 KV 面试术语 94 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| ASK-94 | Redis 内存 KV 面试术语 95 | [23 IO](23-IO多路复用与高性能Server.md) |
| 穿透-95 | Redis 内存 KV 面试术语 96 | [53 OS](53-操作系统面试八股与口述模板.md) |
| 击穿-96 | Redis 内存 KV 面试术语 97 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| 雪崩-97 | Redis 内存 KV 面试术语 98 | [23 IO](23-IO多路复用与高性能Server.md) |
| 布隆-98 | Redis 内存 KV 面试术语 99 | [53 OS](53-操作系统面试八股与口述模板.md) |
| CacheAside-99 | Redis 内存 KV 面试术语 100 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |

## §8 闭卷自测（10 题 + 参考答案）

| 题号 | 题目 | ☐ |
|------|------|---|
| 1 | 五种结构及典型底层编码 | ☐ |
| 2 | RDB vs AOF 选型 | ☐ |
| 3 | 过期删除策略 | ☐ |
| 4 | maxmemory 淘汰策略 | ☐ |
| 5 | Cluster 16384 slot 路由 | ☐ |
| 6 | 穿透、击穿、雪崩区别与方案 | ☐ |
| 7 | Cache Aside 读写流程 | ☐ |
| 8 | 分布式锁如何安全释放 | ☐ |
| 9 | 主从复制大致过程 | ☐ |
| 10 | 单线程为何仍高性能 | ☐ |

### 参考答案

1. String→SDS；List→quicklist；Hash→ziplist/dict；Set→intset/dict；ZSet→skiplist+dict。  
2. RDB 恢复快但可能丢窗口数据；AOF 丢少但文件大；生产常用 AOF everysec 或混合持久化。  
3. 惰性删除（访问时）+ 定期抽样删除。  
4. volatile-lru/allkeys-lru、LFU、TTL、random、noeviction 等。  
5. CRC16(key)%16384 得 slot；MOVED 永久重定向，ASK 迁移中临时重定向。  
6. 穿透：布隆/空值；击穿：互斥锁/逻辑过期；雪崩：随机 TTL/集群/限流。  
7. 读 miss 读 DB 写缓存；写 DB 后删缓存（非双写缓存）。  
8. Lua 脚本比较 value 再 DEL，防误删他人锁；或 Redisson 看门狗续期。  
9. 全量 RDB 快照 + repl_backlog 增量命令流。  
10. 内存操作快、单线程无锁无切换、IO 多路复用；6.0 起 IO 线程并行读写 socket。


→ [53-OS](53-操作系统面试八股与口述模板.md)
