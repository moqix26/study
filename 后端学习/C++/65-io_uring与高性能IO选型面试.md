# io_uring 与高性能 IO 选型面试

> **文件编码**：UTF-8。  
> **定位**：io_uring vs epoll trade-off、liburing、SQPOLL、C++20 协程+io_uring、选型决策树——承接 [64 定时器](64-定时器与时间轮延时队列设计.md)  
> **交叉阅读**：[23 IO](23-IO多路复用与高性能Server.md) · [31 协程](31-协程C++20-coroutine.md) · [64 定时器](64-定时器与时间轮延时队列设计.md)

---

## 本章与前后章的关系

| 上一章 | 本章 | 下一章 |
|--------|------|--------|
| [64 定时器](64-定时器与时间轮延时队列设计.md) | **本章** | [66 面经](66-大厂面经按公司分类精讲.md) |

```mermaid
flowchart LR
  P[64 定时器] --> C[本章]
  C --> N[66 面经]
```

---

## §0 读前导读

### §0.1 用一句话弄懂本章

io_uring vs epoll trade-off、liburing、SQPOLL、C++20 协程+io_uring、选型决策树

### §0.2 你需要提前知道什么

| 前置 | 说明 |
|------|------|
| 基础 C++ | [01～09 语言基础](01-C++基础语法与数据类型.md) |
| Linux 系统编程 | [11 章](11-Linux与系统编程入门.md) |
| 多线程 | [08 章](08-多线程与并发编程.md) |

### §0.3 本章知识地图（☐→☑）

- ☐ 核心概念能 2 分钟口述
- ☐ 能画架构/时序图
- ☐ 能写 C++ 骨架代码
- ☐ 连环追问 ≥5 题不卡壳
- ☐ 闭卷自测 ≥8/10

### §0.4 建议节奏

| 阶段 | 时长 | 内容 |
|------|------|------|
| 首轮通读 | 4h | 全文 + 代码 |
| 二轮口述 | 2h | Q&A 录音 |
| 工程实验 | 3h | 本地 demo |
| 闭卷自测 | 1h | 文末清单 |

---


## §1 io_uring 概述

SQ/CQ 环减少 syscall；支持 read/write/accept/poll 等。

| | epoll | io_uring |
|---|-------|----------|
| 模型 | Reactor | 可 Proactor |
| syscall | 每次 IO | 批量 submit |
| 内核 | 老 | ≥5.1，生产 ≥5.10 |


## §2 io_uring 核心

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 与 epoll 区别？ | epoll 就绪通知；uring 可完成 IO | POLL vs READ |
| Q2 | liburing？ | queue_init/submit/wait_cqe | 官方封装 |
| Q3 | SQPOLL？ | 内核轮询 SQ；需权限 | CPU 换延迟 |
| Q4 | IOPOLL？ | 轮询完成；O_DIRECT | 裸盘 |
| Q5 | registered buf？ | 固定内存零拷贝 | iovec |
| Q6 | 混用 epoll？ | OP_POLL_ADD | 迁移 |
| Q7 | C++20 协程？ | co_await wait_cqe | [31章](31-协程C++20-coroutine.md) |
| Q8 | nginx 仍 epoll？ | 成熟跨版本 | realism |
| Q9 | 何时 uring？ | 极高 IOPS NVMe | 压测 |
| Q10 | 常见坑？ | CQ 溢出、buf 生命周期 | pin 内核 |

### liburing 读文件

```cpp
#include <liburing.h>
io_uring ring; io_uring_queue_init(32, &ring, 0);
io_uring_sqe* sqe = io_uring_get_sqe(&ring);
io_uring_prep_read(sqe, fd, buf, sizeof(buf), 0);
io_uring_submit(&ring);
io_uring_cqe* cqe; io_uring_wait_cqe(&ring, &cqe);
io_uring_cqe_seen(&ring, cqe);
```


## §3 选型决策树

```
Linux? → 否: Asio
  → 是: 内核≥5.10? → 否: epoll+timerfd(64章)
              → 是: 极限IOPS? → 否: epoll
                          → 是: io_uring+压测
```


## §4 trade-off

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | LT vs ET？ | LT 简单 ET 读空 | [23章](23-IO多路复用与高性能Server.md) |
| Q2 | 延迟 vs 吞吐？ | 批量高吞吐小包或不如 epoll | benchmark |
| Q3 | 线程模型？ | 单环 vs 每线程环 | 无锁 SQ |
| Q4 | 与 64 timerfd？ | POLL timerfd | 统一 wait |
| Q5 | Docker 内核？ | 宿主决定 | K8s |
| Q6 | 回退？ | feature flag epoll | 灰度 |
| Q7 | 面试怎么说？ | 默认 epoll 有数据再 uring | 诚实 |
| Q8 | PG 为何 uring？ | 随机读 | 可选 |
| Q9 | Redis io-threads？ | 非 uring | 区分 |
| Q10 | 趋势？ | uring+eBPF | 关注 |

## §5 选型场景（组 1）

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 场景0：小包 RPC？ | epoll 足够 | 延迟 |
| Q2 | 场景1：NVMe 随机读？ | uring+registered | IOPS |
| Q3 | 场景2：网关？ | epoll+协程 | [31章](31-协程C++20-coroutine.md) |
| Q4 | 场景3：存储？ | uring+IOpoll | 对齐 |
| Q5 | 场景4：混合？ | epoll 接入 uring 盘 | 分层 |
| Q6 | 场景5：Windows？ | IOCP/Asio | 跨平台 |
| Q7 | 场景6：macOS？ | kqueue/Asio | 无 uring |
| Q8 | 场景7：内核 bug？ | 版本 pin | LKML |
| Q9 | 场景8：安全？ | 固定 buf 防替换 | 生命周期 |
| Q10 | 场景9：监控？ | cqe 错误率 | [32章](32-fmt-spdlog与可观测性工程.md) |

## §6 选型场景（组 2）

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 场景10：小包 RPC？ | epoll 足够 | 延迟 |
| Q2 | 场景11：NVMe 随机读？ | uring+registered | IOPS |
| Q3 | 场景12：网关？ | epoll+协程 | [31章](31-协程C++20-coroutine.md) |
| Q4 | 场景13：存储？ | uring+IOpoll | 对齐 |
| Q5 | 场景14：混合？ | epoll 接入 uring 盘 | 分层 |
| Q6 | 场景15：Windows？ | IOCP/Asio | 跨平台 |
| Q7 | 场景16：macOS？ | kqueue/Asio | 无 uring |
| Q8 | 场景17：内核 bug？ | 版本 pin | LKML |
| Q9 | 场景18：安全？ | 固定 buf 防替换 | 生命周期 |
| Q10 | 场景19：监控？ | cqe 错误率 | [32章](32-fmt-spdlog与可观测性工程.md) |

## §7 选型场景（组 3）

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 场景20：小包 RPC？ | epoll 足够 | 延迟 |
| Q2 | 场景21：NVMe 随机读？ | uring+registered | IOPS |
| Q3 | 场景22：网关？ | epoll+协程 | [31章](31-协程C++20-coroutine.md) |
| Q4 | 场景23：存储？ | uring+IOpoll | 对齐 |
| Q5 | 场景24：混合？ | epoll 接入 uring 盘 | 分层 |
| Q6 | 场景25：Windows？ | IOCP/Asio | 跨平台 |
| Q7 | 场景26：macOS？ | kqueue/Asio | 无 uring |
| Q8 | 场景27：内核 bug？ | 版本 pin | LKML |
| Q9 | 场景28：安全？ | 固定 buf 防替换 | 生命周期 |
| Q10 | 场景29：监控？ | cqe 错误率 | [32章](32-fmt-spdlog与可观测性工程.md) |

## §8 选型场景（组 4）

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 场景30：小包 RPC？ | epoll 足够 | 延迟 |
| Q2 | 场景31：NVMe 随机读？ | uring+registered | IOPS |
| Q3 | 场景32：网关？ | epoll+协程 | [31章](31-协程C++20-coroutine.md) |
| Q4 | 场景33：存储？ | uring+IOpoll | 对齐 |
| Q5 | 场景34：混合？ | epoll 接入 uring 盘 | 分层 |
| Q6 | 场景35：Windows？ | IOCP/Asio | 跨平台 |
| Q7 | 场景36：macOS？ | kqueue/Asio | 无 uring |
| Q8 | 场景37：内核 bug？ | 版本 pin | LKML |
| Q9 | 场景38：安全？ | 固定 buf 防替换 | 生命周期 |
| Q10 | 场景39：监控？ | cqe 错误率 | [32章](32-fmt-spdlog与可观测性工程.md) |

## §9 选型场景（组 5）

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 场景40：小包 RPC？ | epoll 足够 | 延迟 |
| Q2 | 场景41：NVMe 随机读？ | uring+registered | IOPS |
| Q3 | 场景42：网关？ | epoll+协程 | [31章](31-协程C++20-coroutine.md) |
| Q4 | 场景43：存储？ | uring+IOpoll | 对齐 |
| Q5 | 场景44：混合？ | epoll 接入 uring 盘 | 分层 |
| Q6 | 场景45：Windows？ | IOCP/Asio | 跨平台 |
| Q7 | 场景46：macOS？ | kqueue/Asio | 无 uring |
| Q8 | 场景47：内核 bug？ | 版本 pin | LKML |
| Q9 | 场景48：安全？ | 固定 buf 防替换 | 生命周期 |
| Q10 | 场景49：监控？ | cqe 错误率 | [32章](32-fmt-spdlog与可观测性工程.md) |

## §10 选型场景（组 6）

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 场景50：小包 RPC？ | epoll 足够 | 延迟 |
| Q2 | 场景51：NVMe 随机读？ | uring+registered | IOPS |
| Q3 | 场景52：网关？ | epoll+协程 | [31章](31-协程C++20-coroutine.md) |
| Q4 | 场景53：存储？ | uring+IOpoll | 对齐 |
| Q5 | 场景54：混合？ | epoll 接入 uring 盘 | 分层 |
| Q6 | 场景55：Windows？ | IOCP/Asio | 跨平台 |
| Q7 | 场景56：macOS？ | kqueue/Asio | 无 uring |
| Q8 | 场景57：内核 bug？ | 版本 pin | LKML |
| Q9 | 场景58：安全？ | 固定 buf 防替换 | 生命周期 |
| Q10 | 场景59：监控？ | cqe 错误率 | [32章](32-fmt-spdlog与可观测性工程.md) |

## §11 选型场景（组 7）

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 场景60：小包 RPC？ | epoll 足够 | 延迟 |
| Q2 | 场景61：NVMe 随机读？ | uring+registered | IOPS |
| Q3 | 场景62：网关？ | epoll+协程 | [31章](31-协程C++20-coroutine.md) |
| Q4 | 场景63：存储？ | uring+IOpoll | 对齐 |
| Q5 | 场景64：混合？ | epoll 接入 uring 盘 | 分层 |
| Q6 | 场景65：Windows？ | IOCP/Asio | 跨平台 |
| Q7 | 场景66：macOS？ | kqueue/Asio | 无 uring |
| Q8 | 场景67：内核 bug？ | 版本 pin | LKML |
| Q9 | 场景68：安全？ | 固定 buf 防替换 | 生命周期 |
| Q10 | 场景69：监控？ | cqe 错误率 | [32章](32-fmt-spdlog与可观测性工程.md) |

## §12 C++20 协程 + io_uring

```cpp
task<size_t> async_read(int fd, span<char> buf) {
    co_return co_await UringRead{fd, buf};
}
```

CQE 到达 resume；strand 保 conn 串行。

---

## §13 下一章

[66 大厂面经按公司分类精讲](66-大厂面经按公司分类精讲.md)

---

## §A1 io_uring 对比附录 1

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A2 io_uring 对比附录 2

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A3 io_uring 对比附录 3

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A4 io_uring 对比附录 4

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A5 io_uring 对比附录 5

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A6 io_uring 对比附录 6

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A7 io_uring 对比附录 7

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A8 io_uring 对比附录 8

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A9 io_uring 对比附录 9

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A10 io_uring 对比附录 10

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A11 io_uring 对比附录 11

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A12 io_uring 对比附录 12

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A13 io_uring 对比附录 13

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A14 io_uring 对比附录 14

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A15 io_uring 对比附录 15

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A16 io_uring 对比附录 16

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A17 io_uring 对比附录 17

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A18 io_uring 对比附录 18

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A19 io_uring 对比附录 19

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A20 io_uring 对比附录 20

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A21 io_uring 对比附录 21

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A22 io_uring 对比附录 22

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A23 io_uring 对比附录 23

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A24 io_uring 对比附录 24

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A25 io_uring 对比附录 25

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A26 io_uring 对比附录 26

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A27 io_uring 对比附录 27

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A28 io_uring 对比附录 28

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A29 io_uring 对比附录 29

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A30 io_uring 对比附录 30

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A31 io_uring 对比附录 31

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A32 io_uring 对比附录 32

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A33 io_uring 对比附录 33

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A34 io_uring 对比附录 34

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A35 io_uring 对比附录 35

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A36 io_uring 对比附录 36

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
- `io_uring_queue_exit`

## §A37 io_uring 对比附录 37

| 维度 | epoll | io_uring |
|------|-------|----------|
| 延迟 | epoll 低 | uring 批量高吞吐 |
| CPU | SQPOLL 占核 | epoll 事件驱动 |
| 兼容 | 全内核 | ≥5.10 |
| 网络 | 成熟 | 渐进 |
| 存储 | sendfile | registered buf |
| 协程 | Asio | co_await CQE |
| 迁移 | A/B 压测 | feature flag |
| 面试 | 默认 epoll | 有数据再 uring |

### liburing API

- `io_uring_queue_init`
- `io_uring_get_sqe`
- `io_uring_prep_read`
- `io_uring_submit`
- `io_uring_wait_cqe`
- `io_uring_cqe_seen`
