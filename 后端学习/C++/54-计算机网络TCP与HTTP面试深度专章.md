# 计算机网络 TCP 与 HTTP 面试深度专章

> **文件编码**：UTF-8。握手挥手、拥塞、HTTP/1/2/3、HTTPS、粘包、零拷贝、TIME_WAIT。[10/23/26] + [计网系列](../../前端学习/计算机网络/00-学习路线图与说明.md)

---


## 本章与前后章的关系

| 上一章 | 本章 | 下一章 |
|--------|------|--------|
| 53 OS | **54** | — |


**学习链**：50 → [51 MySQL](51-MySQL原理与索引事务面试专章.md) → [52 Redis](52-Redis数据结构与缓存面试专章.md) → [53 OS](53-操作系统面试八股与口述模板.md) → [54 计网](54-计算机网络TCP与HTTP面试深度专章.md)


**交叉阅读**：[10 网络](10-网络编程与简易HTTP服务.md)、[23 IO](23-IO多路复用与高性能Server.md)、[26 Asio](26-Boost.Asio异步网络编程.md)

---


## §0 读前导读


### §0.1 用一句话弄懂本章

TCP 提供 **可靠有序字节流**，HTTP 在其上 **语义化请求/响应**；C++ 服务端必须掌握粘包、连接状态机、零拷贝、TLS 与 TIME_WAIT/CLOSE_WAIT 排查，并与 [23 epoll](23-IO多路复用与高性能Server.md) 配合。


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
| Q1 | [1] 三次握手 | SYN SYN+ACK ACK | SYN 队列 |
| Q2 | [2] 四次挥手 | FIN ACK FIN ACK | TIME_WAIT |
| Q3 | [3] 滑动窗口 | rwnd 流控 | 零窗口 |
| Q4 | [4] 拥塞控制 | 慢启动快重传 | cwnd |
| Q5 | [5] HTTP/1.1 | 持久连接 HOL | chunked |
| Q6 | [6] HTTP/2 | 多路复用 HPACK | TCP HOL |
| Q7 | [7] HTTP/3 QUIC | UDP 独立流 | 防火墙 |
| Q8 | [8] HTTPS | TLS on TCP | [计网05](../../前端学习/计算机网络/05-HTTPS与TLS加密.md) |
| Q9 | [9] 粘包 | Content-Length/长度前缀 | [10章](10-网络编程与简易HTTP服务.md) |
| Q10 | [10] sendfile | 内核零拷贝 | [18章](18-高性能C++与内存对齐.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [11] TIME_WAIT | 2MSL 主动关 | reuse 长连接 |
| Q2 | [12] CLOSE_WAIT | 应用未 close | fd 泄漏 |
| Q3 | [13] 502/504 | 上游坏/超时 | nginx |
| Q4 | [14] Nagle | TCP_NODELAY | 低延迟 |
| Q5 | [15] epoll 驱动 | socket 可读 | [23章](23-IO多路复用与高性能Server.md) |
| Q6 | [16] curl 计时 | connect/TTFB | 分段 |
| Q7 | [17] DNS | 递归迭代 | [计网03](../../前端学习/计算机网络/03-IP地址与DNS解析.md) |
| Q8 | [18] MTU MSS | 1500/1460 | 分片 |
| Q9 | [19] WebSocket | Upgrade | 全双工 |
| Q10 | [20] gRPC HTTP/2 | Protobuf | [19章](19-gRPC与Protobuf工程化.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [21] 三次握手 | SYN SYN+ACK ACK | SYN 队列 |
| Q2 | [22] 四次挥手 | FIN ACK FIN ACK | TIME_WAIT |
| Q3 | [23] 滑动窗口 | rwnd 流控 | 零窗口 |
| Q4 | [24] 拥塞控制 | 慢启动快重传 | cwnd |
| Q5 | [25] HTTP/1.1 | 持久连接 HOL | chunked |
| Q6 | [26] HTTP/2 | 多路复用 HPACK | TCP HOL |
| Q7 | [27] HTTP/3 QUIC | UDP 独立流 | 防火墙 |
| Q8 | [28] HTTPS | TLS on TCP | [计网05](../../前端学习/计算机网络/05-HTTPS与TLS加密.md) |
| Q9 | [29] 粘包 | Content-Length/长度前缀 | [10章](10-网络编程与简易HTTP服务.md) |
| Q10 | [30] sendfile | 内核零拷贝 | [18章](18-高性能C++与内存对齐.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [31] TIME_WAIT | 2MSL 主动关 | reuse 长连接 |
| Q2 | [32] CLOSE_WAIT | 应用未 close | fd 泄漏 |
| Q3 | [33] 502/504 | 上游坏/超时 | nginx |
| Q4 | [34] Nagle | TCP_NODELAY | 低延迟 |
| Q5 | [35] epoll 驱动 | socket 可读 | [23章](23-IO多路复用与高性能Server.md) |
| Q6 | [36] curl 计时 | connect/TTFB | 分段 |
| Q7 | [37] DNS | 递归迭代 | [计网03](../../前端学习/计算机网络/03-IP地址与DNS解析.md) |
| Q8 | [38] MTU MSS | 1500/1460 | 分片 |
| Q9 | [39] WebSocket | Upgrade | 全双工 |
| Q10 | [40] gRPC HTTP/2 | Protobuf | [19章](19-gRPC与Protobuf工程化.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [41] 三次握手 | SYN SYN+ACK ACK | SYN 队列 |
| Q2 | [42] 四次挥手 | FIN ACK FIN ACK | TIME_WAIT |
| Q3 | [43] 滑动窗口 | rwnd 流控 | 零窗口 |
| Q4 | [44] 拥塞控制 | 慢启动快重传 | cwnd |
| Q5 | [45] HTTP/1.1 | 持久连接 HOL | chunked |
| Q6 | [46] HTTP/2 | 多路复用 HPACK | TCP HOL |
| Q7 | [47] HTTP/3 QUIC | UDP 独立流 | 防火墙 |
| Q8 | [48] HTTPS | TLS on TCP | [计网05](../../前端学习/计算机网络/05-HTTPS与TLS加密.md) |
| Q9 | [49] 粘包 | Content-Length/长度前缀 | [10章](10-网络编程与简易HTTP服务.md) |
| Q10 | [50] sendfile | 内核零拷贝 | [18章](18-高性能C++与内存对齐.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [51] TIME_WAIT | 2MSL 主动关 | reuse 长连接 |
| Q2 | [52] CLOSE_WAIT | 应用未 close | fd 泄漏 |
| Q3 | [53] 502/504 | 上游坏/超时 | nginx |
| Q4 | [54] Nagle | TCP_NODELAY | 低延迟 |
| Q5 | [55] epoll 驱动 | socket 可读 | [23章](23-IO多路复用与高性能Server.md) |
| Q6 | [56] curl 计时 | connect/TTFB | 分段 |
| Q7 | [57] DNS | 递归迭代 | [计网03](../../前端学习/计算机网络/03-IP地址与DNS解析.md) |
| Q8 | [58] MTU MSS | 1500/1460 | 分片 |
| Q9 | [59] WebSocket | Upgrade | 全双工 |
| Q10 | [60] gRPC HTTP/2 | Protobuf | [19章](19-gRPC与Protobuf工程化.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [61] 三次握手 | SYN SYN+ACK ACK | SYN 队列 |
| Q2 | [62] 四次挥手 | FIN ACK FIN ACK | TIME_WAIT |
| Q3 | [63] 滑动窗口 | rwnd 流控 | 零窗口 |
| Q4 | [64] 拥塞控制 | 慢启动快重传 | cwnd |
| Q5 | [65] HTTP/1.1 | 持久连接 HOL | chunked |
| Q6 | [66] HTTP/2 | 多路复用 HPACK | TCP HOL |
| Q7 | [67] HTTP/3 QUIC | UDP 独立流 | 防火墙 |
| Q8 | [68] HTTPS | TLS on TCP | [计网05](../../前端学习/计算机网络/05-HTTPS与TLS加密.md) |
| Q9 | [69] 粘包 | Content-Length/长度前缀 | [10章](10-网络编程与简易HTTP服务.md) |
| Q10 | [70] sendfile | 内核零拷贝 | [18章](18-高性能C++与内存对齐.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [71] TIME_WAIT | 2MSL 主动关 | reuse 长连接 |
| Q2 | [72] CLOSE_WAIT | 应用未 close | fd 泄漏 |
| Q3 | [73] 502/504 | 上游坏/超时 | nginx |
| Q4 | [74] Nagle | TCP_NODELAY | 低延迟 |
| Q5 | [75] epoll 驱动 | socket 可读 | [23章](23-IO多路复用与高性能Server.md) |
| Q6 | [76] curl 计时 | connect/TTFB | 分段 |
| Q7 | [77] DNS | 递归迭代 | [计网03](../../前端学习/计算机网络/03-IP地址与DNS解析.md) |
| Q8 | [78] MTU MSS | 1500/1460 | 分片 |
| Q9 | [79] WebSocket | Upgrade | 全双工 |
| Q10 | [80] gRPC HTTP/2 | Protobuf | [19章](19-gRPC与Protobuf工程化.md) |

## §2 核心面试 Q&A 组 2


| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [1] 三次握手 | SYN SYN+ACK ACK | SYN 队列 |
| Q2 | [2] 四次挥手 | FIN ACK FIN ACK | TIME_WAIT |
| Q3 | [3] 滑动窗口 | rwnd 流控 | 零窗口 |
| Q4 | [4] 拥塞控制 | 慢启动快重传 | cwnd |
| Q5 | [5] HTTP/1.1 | 持久连接 HOL | chunked |
| Q6 | [6] HTTP/2 | 多路复用 HPACK | TCP HOL |
| Q7 | [7] HTTP/3 QUIC | UDP 独立流 | 防火墙 |
| Q8 | [8] HTTPS | TLS on TCP | [计网05](../../前端学习/计算机网络/05-HTTPS与TLS加密.md) |
| Q9 | [9] 粘包 | Content-Length/长度前缀 | [10章](10-网络编程与简易HTTP服务.md) |
| Q10 | [10] sendfile | 内核零拷贝 | [18章](18-高性能C++与内存对齐.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [11] TIME_WAIT | 2MSL 主动关 | reuse 长连接 |
| Q2 | [12] CLOSE_WAIT | 应用未 close | fd 泄漏 |
| Q3 | [13] 502/504 | 上游坏/超时 | nginx |
| Q4 | [14] Nagle | TCP_NODELAY | 低延迟 |
| Q5 | [15] epoll 驱动 | socket 可读 | [23章](23-IO多路复用与高性能Server.md) |
| Q6 | [16] curl 计时 | connect/TTFB | 分段 |
| Q7 | [17] DNS | 递归迭代 | [计网03](../../前端学习/计算机网络/03-IP地址与DNS解析.md) |
| Q8 | [18] MTU MSS | 1500/1460 | 分片 |
| Q9 | [19] WebSocket | Upgrade | 全双工 |
| Q10 | [20] gRPC HTTP/2 | Protobuf | [19章](19-gRPC与Protobuf工程化.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [21] 三次握手 | SYN SYN+ACK ACK | SYN 队列 |
| Q2 | [22] 四次挥手 | FIN ACK FIN ACK | TIME_WAIT |
| Q3 | [23] 滑动窗口 | rwnd 流控 | 零窗口 |
| Q4 | [24] 拥塞控制 | 慢启动快重传 | cwnd |
| Q5 | [25] HTTP/1.1 | 持久连接 HOL | chunked |
| Q6 | [26] HTTP/2 | 多路复用 HPACK | TCP HOL |
| Q7 | [27] HTTP/3 QUIC | UDP 独立流 | 防火墙 |
| Q8 | [28] HTTPS | TLS on TCP | [计网05](../../前端学习/计算机网络/05-HTTPS与TLS加密.md) |
| Q9 | [29] 粘包 | Content-Length/长度前缀 | [10章](10-网络编程与简易HTTP服务.md) |
| Q10 | [30] sendfile | 内核零拷贝 | [18章](18-高性能C++与内存对齐.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [31] TIME_WAIT | 2MSL 主动关 | reuse 长连接 |
| Q2 | [32] CLOSE_WAIT | 应用未 close | fd 泄漏 |
| Q3 | [33] 502/504 | 上游坏/超时 | nginx |
| Q4 | [34] Nagle | TCP_NODELAY | 低延迟 |
| Q5 | [35] epoll 驱动 | socket 可读 | [23章](23-IO多路复用与高性能Server.md) |
| Q6 | [36] curl 计时 | connect/TTFB | 分段 |
| Q7 | [37] DNS | 递归迭代 | [计网03](../../前端学习/计算机网络/03-IP地址与DNS解析.md) |
| Q8 | [38] MTU MSS | 1500/1460 | 分片 |
| Q9 | [39] WebSocket | Upgrade | 全双工 |
| Q10 | [40] gRPC HTTP/2 | Protobuf | [19章](19-gRPC与Protobuf工程化.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [41] 三次握手 | SYN SYN+ACK ACK | SYN 队列 |
| Q2 | [42] 四次挥手 | FIN ACK FIN ACK | TIME_WAIT |
| Q3 | [43] 滑动窗口 | rwnd 流控 | 零窗口 |
| Q4 | [44] 拥塞控制 | 慢启动快重传 | cwnd |
| Q5 | [45] HTTP/1.1 | 持久连接 HOL | chunked |
| Q6 | [46] HTTP/2 | 多路复用 HPACK | TCP HOL |
| Q7 | [47] HTTP/3 QUIC | UDP 独立流 | 防火墙 |
| Q8 | [48] HTTPS | TLS on TCP | [计网05](../../前端学习/计算机网络/05-HTTPS与TLS加密.md) |
| Q9 | [49] 粘包 | Content-Length/长度前缀 | [10章](10-网络编程与简易HTTP服务.md) |
| Q10 | [50] sendfile | 内核零拷贝 | [18章](18-高性能C++与内存对齐.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [51] TIME_WAIT | 2MSL 主动关 | reuse 长连接 |
| Q2 | [52] CLOSE_WAIT | 应用未 close | fd 泄漏 |
| Q3 | [53] 502/504 | 上游坏/超时 | nginx |
| Q4 | [54] Nagle | TCP_NODELAY | 低延迟 |
| Q5 | [55] epoll 驱动 | socket 可读 | [23章](23-IO多路复用与高性能Server.md) |
| Q6 | [56] curl 计时 | connect/TTFB | 分段 |
| Q7 | [57] DNS | 递归迭代 | [计网03](../../前端学习/计算机网络/03-IP地址与DNS解析.md) |
| Q8 | [58] MTU MSS | 1500/1460 | 分片 |
| Q9 | [59] WebSocket | Upgrade | 全双工 |
| Q10 | [60] gRPC HTTP/2 | Protobuf | [19章](19-gRPC与Protobuf工程化.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [61] 三次握手 | SYN SYN+ACK ACK | SYN 队列 |
| Q2 | [62] 四次挥手 | FIN ACK FIN ACK | TIME_WAIT |
| Q3 | [63] 滑动窗口 | rwnd 流控 | 零窗口 |
| Q4 | [64] 拥塞控制 | 慢启动快重传 | cwnd |
| Q5 | [65] HTTP/1.1 | 持久连接 HOL | chunked |
| Q6 | [66] HTTP/2 | 多路复用 HPACK | TCP HOL |
| Q7 | [67] HTTP/3 QUIC | UDP 独立流 | 防火墙 |
| Q8 | [68] HTTPS | TLS on TCP | [计网05](../../前端学习/计算机网络/05-HTTPS与TLS加密.md) |
| Q9 | [69] 粘包 | Content-Length/长度前缀 | [10章](10-网络编程与简易HTTP服务.md) |
| Q10 | [70] sendfile | 内核零拷贝 | [18章](18-高性能C++与内存对齐.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [71] TIME_WAIT | 2MSL 主动关 | reuse 长连接 |
| Q2 | [72] CLOSE_WAIT | 应用未 close | fd 泄漏 |
| Q3 | [73] 502/504 | 上游坏/超时 | nginx |
| Q4 | [74] Nagle | TCP_NODELAY | 低延迟 |
| Q5 | [75] epoll 驱动 | socket 可读 | [23章](23-IO多路复用与高性能Server.md) |
| Q6 | [76] curl 计时 | connect/TTFB | 分段 |
| Q7 | [77] DNS | 递归迭代 | [计网03](../../前端学习/计算机网络/03-IP地址与DNS解析.md) |
| Q8 | [78] MTU MSS | 1500/1460 | 分片 |
| Q9 | [79] WebSocket | Upgrade | 全双工 |
| Q10 | [80] gRPC HTTP/2 | Protobuf | [19章](19-gRPC与Protobuf工程化.md) |

## §3 核心面试 Q&A 组 3


| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [1] 三次握手 | SYN SYN+ACK ACK | SYN 队列 |
| Q2 | [2] 四次挥手 | FIN ACK FIN ACK | TIME_WAIT |
| Q3 | [3] 滑动窗口 | rwnd 流控 | 零窗口 |
| Q4 | [4] 拥塞控制 | 慢启动快重传 | cwnd |
| Q5 | [5] HTTP/1.1 | 持久连接 HOL | chunked |
| Q6 | [6] HTTP/2 | 多路复用 HPACK | TCP HOL |
| Q7 | [7] HTTP/3 QUIC | UDP 独立流 | 防火墙 |
| Q8 | [8] HTTPS | TLS on TCP | [计网05](../../前端学习/计算机网络/05-HTTPS与TLS加密.md) |
| Q9 | [9] 粘包 | Content-Length/长度前缀 | [10章](10-网络编程与简易HTTP服务.md) |
| Q10 | [10] sendfile | 内核零拷贝 | [18章](18-高性能C++与内存对齐.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [11] TIME_WAIT | 2MSL 主动关 | reuse 长连接 |
| Q2 | [12] CLOSE_WAIT | 应用未 close | fd 泄漏 |
| Q3 | [13] 502/504 | 上游坏/超时 | nginx |
| Q4 | [14] Nagle | TCP_NODELAY | 低延迟 |
| Q5 | [15] epoll 驱动 | socket 可读 | [23章](23-IO多路复用与高性能Server.md) |
| Q6 | [16] curl 计时 | connect/TTFB | 分段 |
| Q7 | [17] DNS | 递归迭代 | [计网03](../../前端学习/计算机网络/03-IP地址与DNS解析.md) |
| Q8 | [18] MTU MSS | 1500/1460 | 分片 |
| Q9 | [19] WebSocket | Upgrade | 全双工 |
| Q10 | [20] gRPC HTTP/2 | Protobuf | [19章](19-gRPC与Protobuf工程化.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [21] 三次握手 | SYN SYN+ACK ACK | SYN 队列 |
| Q2 | [22] 四次挥手 | FIN ACK FIN ACK | TIME_WAIT |
| Q3 | [23] 滑动窗口 | rwnd 流控 | 零窗口 |
| Q4 | [24] 拥塞控制 | 慢启动快重传 | cwnd |
| Q5 | [25] HTTP/1.1 | 持久连接 HOL | chunked |
| Q6 | [26] HTTP/2 | 多路复用 HPACK | TCP HOL |
| Q7 | [27] HTTP/3 QUIC | UDP 独立流 | 防火墙 |
| Q8 | [28] HTTPS | TLS on TCP | [计网05](../../前端学习/计算机网络/05-HTTPS与TLS加密.md) |
| Q9 | [29] 粘包 | Content-Length/长度前缀 | [10章](10-网络编程与简易HTTP服务.md) |
| Q10 | [30] sendfile | 内核零拷贝 | [18章](18-高性能C++与内存对齐.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [31] TIME_WAIT | 2MSL 主动关 | reuse 长连接 |
| Q2 | [32] CLOSE_WAIT | 应用未 close | fd 泄漏 |
| Q3 | [33] 502/504 | 上游坏/超时 | nginx |
| Q4 | [34] Nagle | TCP_NODELAY | 低延迟 |
| Q5 | [35] epoll 驱动 | socket 可读 | [23章](23-IO多路复用与高性能Server.md) |
| Q6 | [36] curl 计时 | connect/TTFB | 分段 |
| Q7 | [37] DNS | 递归迭代 | [计网03](../../前端学习/计算机网络/03-IP地址与DNS解析.md) |
| Q8 | [38] MTU MSS | 1500/1460 | 分片 |
| Q9 | [39] WebSocket | Upgrade | 全双工 |
| Q10 | [40] gRPC HTTP/2 | Protobuf | [19章](19-gRPC与Protobuf工程化.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [41] 三次握手 | SYN SYN+ACK ACK | SYN 队列 |
| Q2 | [42] 四次挥手 | FIN ACK FIN ACK | TIME_WAIT |
| Q3 | [43] 滑动窗口 | rwnd 流控 | 零窗口 |
| Q4 | [44] 拥塞控制 | 慢启动快重传 | cwnd |
| Q5 | [45] HTTP/1.1 | 持久连接 HOL | chunked |
| Q6 | [46] HTTP/2 | 多路复用 HPACK | TCP HOL |
| Q7 | [47] HTTP/3 QUIC | UDP 独立流 | 防火墙 |
| Q8 | [48] HTTPS | TLS on TCP | [计网05](../../前端学习/计算机网络/05-HTTPS与TLS加密.md) |
| Q9 | [49] 粘包 | Content-Length/长度前缀 | [10章](10-网络编程与简易HTTP服务.md) |
| Q10 | [50] sendfile | 内核零拷贝 | [18章](18-高性能C++与内存对齐.md) |

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | [51] TIME_WAIT | 2MSL 主动关 | reuse 长连接 |
| Q2 | [52] CLOSE_WAIT | 应用未 close | fd 泄漏 |
| Q3 | [53] 502/504 | 上游坏/超时 | nginx |
| Q4 | [54] Nagle | TCP_NODELAY | 低延迟 |
| Q5 | [55] epoll 驱动 | socket 可读 | [23章](23-IO多路复用与高性能Server.md) |
| Q6 | [56] curl 计时 | connect/TTFB | 分段 |
| Q7 | [57] DNS | 递归迭代 | [计网03](../../前端学习/计算机网络/03-IP地址与DNS解析.md) |
| Q8 | [58] MTU MSS | 1500/1460 | 分片 |
| Q9 | [59] WebSocket | Upgrade | 全双工 |
| Q10 | [60] gRPC HTTP/2 | Protobuf | [19章](19-gRPC与Protobuf工程化.md) |

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
| SYN | TCP/HTTP 协议栈面试术语 1 | [10 网络](10-网络编程与简易HTTP服务.md) |
| ACK | TCP/HTTP 协议栈面试术语 2 | [23 IO](23-IO多路复用与高性能Server.md) |
| FIN | TCP/HTTP 协议栈面试术语 3 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| TIME_WAIT | TCP/HTTP 协议栈面试术语 4 | [10 网络](10-网络编程与简易HTTP服务.md) |
| CLOSE_WAIT | TCP/HTTP 协议栈面试术语 5 | [23 IO](23-IO多路复用与高性能Server.md) |
| rwnd | TCP/HTTP 协议栈面试术语 6 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| cwnd | TCP/HTTP 协议栈面试术语 7 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SACK | TCP/HTTP 协议栈面试术语 8 | [23 IO](23-IO多路复用与高性能Server.md) |
| RTO | TCP/HTTP 协议栈面试术语 9 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| Nagle | TCP/HTTP 协议栈面试术语 10 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HTTP/1.1 | TCP/HTTP 协议栈面试术语 11 | [23 IO](23-IO多路复用与高性能Server.md) |
| HTTP/2 | TCP/HTTP 协议栈面试术语 12 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| HTTP/3 | TCP/HTTP 协议栈面试术语 13 | [10 网络](10-网络编程与简易HTTP服务.md) |
| QUIC | TCP/HTTP 协议栈面试术语 14 | [23 IO](23-IO多路复用与高性能Server.md) |
| TLS | TCP/HTTP 协议栈面试术语 15 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| HPACK | TCP/HTTP 协议栈面试术语 16 | [10 网络](10-网络编程与简易HTTP服务.md) |
| 粘包 | TCP/HTTP 协议栈面试术语 17 | [23 IO](23-IO多路复用与高性能Server.md) |
| Content-Length | TCP/HTTP 协议栈面试术语 18 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| chunked | TCP/HTTP 协议栈面试术语 19 | [10 网络](10-网络编程与简易HTTP服务.md) |
| 502 | TCP/HTTP 协议栈面试术语 20 | [23 IO](23-IO多路复用与高性能Server.md) |
| 504 | TCP/HTTP 协议栈面试术语 21 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| CORS | TCP/HTTP 协议栈面试术语 22 | [10 网络](10-网络编程与简易HTTP服务.md) |
| Cookie | TCP/HTTP 协议栈面试术语 23 | [23 IO](23-IO多路复用与高性能Server.md) |
| CDN | TCP/HTTP 协议栈面试术语 24 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| sendfile | TCP/HTTP 协议栈面试术语 25 | [10 网络](10-网络编程与简易HTTP服务.md) |
| splice | TCP/HTTP 协议栈面试术语 26 | [23 IO](23-IO多路复用与高性能Server.md) |
| DMA | TCP/HTTP 协议栈面试术语 27 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| WebSocket | TCP/HTTP 协议栈面试术语 28 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SSE | TCP/HTTP 协议栈面试术语 29 | [23 IO](23-IO多路复用与高性能Server.md) |
| ALPN | TCP/HTTP 协议栈面试术语 30 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| SNI | TCP/HTTP 协议栈面试术语 31 | [10 网络](10-网络编程与简易HTTP服务.md) |
| mTLS | TCP/HTTP 协议栈面试术语 32 | [23 IO](23-IO多路复用与高性能Server.md) |
| HSTS | TCP/HTTP 协议栈面试术语 33 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| TCP Fast Open | TCP/HTTP 协议栈面试术语 34 | [10 网络](10-网络编程与简易HTTP服务.md) |
| BBR | TCP/HTTP 协议栈面试术语 35 | [23 IO](23-IO多路复用与高性能Server.md) |
| SO_REUSEPORT | TCP/HTTP 协议栈面试术语 36 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| backlog | TCP/HTTP 协议栈面试术语 37 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SYN-37 | TCP/HTTP 协议栈面试术语 38 | [23 IO](23-IO多路复用与高性能Server.md) |
| ACK-38 | TCP/HTTP 协议栈面试术语 39 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| FIN-39 | TCP/HTTP 协议栈面试术语 40 | [10 网络](10-网络编程与简易HTTP服务.md) |
| TIME_WAIT-40 | TCP/HTTP 协议栈面试术语 41 | [23 IO](23-IO多路复用与高性能Server.md) |
| CLOSE_WAIT-41 | TCP/HTTP 协议栈面试术语 42 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| rwnd-42 | TCP/HTTP 协议栈面试术语 43 | [10 网络](10-网络编程与简易HTTP服务.md) |
| cwnd-43 | TCP/HTTP 协议栈面试术语 44 | [23 IO](23-IO多路复用与高性能Server.md) |
| SACK-44 | TCP/HTTP 协议栈面试术语 45 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| RTO-45 | TCP/HTTP 协议栈面试术语 46 | [10 网络](10-网络编程与简易HTTP服务.md) |
| Nagle-46 | TCP/HTTP 协议栈面试术语 47 | [23 IO](23-IO多路复用与高性能Server.md) |
| HTTP/1.1-47 | TCP/HTTP 协议栈面试术语 48 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| HTTP/2-48 | TCP/HTTP 协议栈面试术语 49 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HTTP/3-49 | TCP/HTTP 协议栈面试术语 50 | [23 IO](23-IO多路复用与高性能Server.md) |
| QUIC-50 | TCP/HTTP 协议栈面试术语 51 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| TLS-51 | TCP/HTTP 协议栈面试术语 52 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HPACK-52 | TCP/HTTP 协议栈面试术语 53 | [23 IO](23-IO多路复用与高性能Server.md) |
| 粘包-53 | TCP/HTTP 协议栈面试术语 54 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| Content-Length-54 | TCP/HTTP 协议栈面试术语 55 | [10 网络](10-网络编程与简易HTTP服务.md) |
| chunked-55 | TCP/HTTP 协议栈面试术语 56 | [23 IO](23-IO多路复用与高性能Server.md) |
| 502-56 | TCP/HTTP 协议栈面试术语 57 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| 504-57 | TCP/HTTP 协议栈面试术语 58 | [10 网络](10-网络编程与简易HTTP服务.md) |
| CORS-58 | TCP/HTTP 协议栈面试术语 59 | [23 IO](23-IO多路复用与高性能Server.md) |
| Cookie-59 | TCP/HTTP 协议栈面试术语 60 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| CDN-60 | TCP/HTTP 协议栈面试术语 61 | [10 网络](10-网络编程与简易HTTP服务.md) |
| sendfile-61 | TCP/HTTP 协议栈面试术语 62 | [23 IO](23-IO多路复用与高性能Server.md) |
| splice-62 | TCP/HTTP 协议栈面试术语 63 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| DMA-63 | TCP/HTTP 协议栈面试术语 64 | [10 网络](10-网络编程与简易HTTP服务.md) |
| WebSocket-64 | TCP/HTTP 协议栈面试术语 65 | [23 IO](23-IO多路复用与高性能Server.md) |
| SSE-65 | TCP/HTTP 协议栈面试术语 66 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| ALPN-66 | TCP/HTTP 协议栈面试术语 67 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SNI-67 | TCP/HTTP 协议栈面试术语 68 | [23 IO](23-IO多路复用与高性能Server.md) |
| mTLS-68 | TCP/HTTP 协议栈面试术语 69 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| HSTS-69 | TCP/HTTP 协议栈面试术语 70 | [10 网络](10-网络编程与简易HTTP服务.md) |
| TCP Fast Open-70 | TCP/HTTP 协议栈面试术语 71 | [23 IO](23-IO多路复用与高性能Server.md) |
| BBR-71 | TCP/HTTP 协议栈面试术语 72 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| SO_REUSEPORT-72 | TCP/HTTP 协议栈面试术语 73 | [10 网络](10-网络编程与简易HTTP服务.md) |
| backlog-73 | TCP/HTTP 协议栈面试术语 74 | [23 IO](23-IO多路复用与高性能Server.md) |
| SYN-74 | TCP/HTTP 协议栈面试术语 75 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| ACK-75 | TCP/HTTP 协议栈面试术语 76 | [10 网络](10-网络编程与简易HTTP服务.md) |
| FIN-76 | TCP/HTTP 协议栈面试术语 77 | [23 IO](23-IO多路复用与高性能Server.md) |
| TIME_WAIT-77 | TCP/HTTP 协议栈面试术语 78 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| CLOSE_WAIT-78 | TCP/HTTP 协议栈面试术语 79 | [10 网络](10-网络编程与简易HTTP服务.md) |
| rwnd-79 | TCP/HTTP 协议栈面试术语 80 | [23 IO](23-IO多路复用与高性能Server.md) |
| cwnd-80 | TCP/HTTP 协议栈面试术语 81 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| SACK-81 | TCP/HTTP 协议栈面试术语 82 | [10 网络](10-网络编程与简易HTTP服务.md) |
| RTO-82 | TCP/HTTP 协议栈面试术语 83 | [23 IO](23-IO多路复用与高性能Server.md) |
| Nagle-83 | TCP/HTTP 协议栈面试术语 84 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| HTTP/1.1-84 | TCP/HTTP 协议栈面试术语 85 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HTTP/2-85 | TCP/HTTP 协议栈面试术语 86 | [23 IO](23-IO多路复用与高性能Server.md) |
| HTTP/3-86 | TCP/HTTP 协议栈面试术语 87 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| QUIC-87 | TCP/HTTP 协议栈面试术语 88 | [10 网络](10-网络编程与简易HTTP服务.md) |
| TLS-88 | TCP/HTTP 协议栈面试术语 89 | [23 IO](23-IO多路复用与高性能Server.md) |
| HPACK-89 | TCP/HTTP 协议栈面试术语 90 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| 粘包-90 | TCP/HTTP 协议栈面试术语 91 | [10 网络](10-网络编程与简易HTTP服务.md) |
| Content-Length-91 | TCP/HTTP 协议栈面试术语 92 | [23 IO](23-IO多路复用与高性能Server.md) |
| chunked-92 | TCP/HTTP 协议栈面试术语 93 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| 502-93 | TCP/HTTP 协议栈面试术语 94 | [10 网络](10-网络编程与简易HTTP服务.md) |
| 504-94 | TCP/HTTP 协议栈面试术语 95 | [23 IO](23-IO多路复用与高性能Server.md) |
| CORS-95 | TCP/HTTP 协议栈面试术语 96 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| Cookie-96 | TCP/HTTP 协议栈面试术语 97 | [10 网络](10-网络编程与简易HTTP服务.md) |
| CDN-97 | TCP/HTTP 协议栈面试术语 98 | [23 IO](23-IO多路复用与高性能Server.md) |
| sendfile-98 | TCP/HTTP 协议栈面试术语 99 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| splice-99 | TCP/HTTP 协议栈面试术语 100 | [10 网络](10-网络编程与简易HTTP服务.md) |
| DMA-100 | TCP/HTTP 协议栈面试术语 101 | [23 IO](23-IO多路复用与高性能Server.md) |
| WebSocket-101 | TCP/HTTP 协议栈面试术语 102 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| SSE-102 | TCP/HTTP 协议栈面试术语 103 | [10 网络](10-网络编程与简易HTTP服务.md) |
| ALPN-103 | TCP/HTTP 协议栈面试术语 104 | [23 IO](23-IO多路复用与高性能Server.md) |
| SNI-104 | TCP/HTTP 协议栈面试术语 105 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| mTLS-105 | TCP/HTTP 协议栈面试术语 106 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HSTS-106 | TCP/HTTP 协议栈面试术语 107 | [23 IO](23-IO多路复用与高性能Server.md) |
| TCP Fast Open-107 | TCP/HTTP 协议栈面试术语 108 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| BBR-108 | TCP/HTTP 协议栈面试术语 109 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SO_REUSEPORT-109 | TCP/HTTP 协议栈面试术语 110 | [23 IO](23-IO多路复用与高性能Server.md) |
| backlog-110 | TCP/HTTP 协议栈面试术语 111 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| SYN-111 | TCP/HTTP 协议栈面试术语 112 | [10 网络](10-网络编程与简易HTTP服务.md) |
| ACK-112 | TCP/HTTP 协议栈面试术语 113 | [23 IO](23-IO多路复用与高性能Server.md) |
| FIN-113 | TCP/HTTP 协议栈面试术语 114 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| TIME_WAIT-114 | TCP/HTTP 协议栈面试术语 115 | [10 网络](10-网络编程与简易HTTP服务.md) |
| CLOSE_WAIT-115 | TCP/HTTP 协议栈面试术语 116 | [23 IO](23-IO多路复用与高性能Server.md) |
| rwnd-116 | TCP/HTTP 协议栈面试术语 117 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| cwnd-117 | TCP/HTTP 协议栈面试术语 118 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SACK-118 | TCP/HTTP 协议栈面试术语 119 | [23 IO](23-IO多路复用与高性能Server.md) |
| RTO-119 | TCP/HTTP 协议栈面试术语 120 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

## §6 术语速查表 2


| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| SYN | TCP/HTTP 协议栈面试术语 1 | [10 网络](10-网络编程与简易HTTP服务.md) |
| ACK | TCP/HTTP 协议栈面试术语 2 | [23 IO](23-IO多路复用与高性能Server.md) |
| FIN | TCP/HTTP 协议栈面试术语 3 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| TIME_WAIT | TCP/HTTP 协议栈面试术语 4 | [10 网络](10-网络编程与简易HTTP服务.md) |
| CLOSE_WAIT | TCP/HTTP 协议栈面试术语 5 | [23 IO](23-IO多路复用与高性能Server.md) |
| rwnd | TCP/HTTP 协议栈面试术语 6 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| cwnd | TCP/HTTP 协议栈面试术语 7 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SACK | TCP/HTTP 协议栈面试术语 8 | [23 IO](23-IO多路复用与高性能Server.md) |
| RTO | TCP/HTTP 协议栈面试术语 9 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| Nagle | TCP/HTTP 协议栈面试术语 10 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HTTP/1.1 | TCP/HTTP 协议栈面试术语 11 | [23 IO](23-IO多路复用与高性能Server.md) |
| HTTP/2 | TCP/HTTP 协议栈面试术语 12 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| HTTP/3 | TCP/HTTP 协议栈面试术语 13 | [10 网络](10-网络编程与简易HTTP服务.md) |
| QUIC | TCP/HTTP 协议栈面试术语 14 | [23 IO](23-IO多路复用与高性能Server.md) |
| TLS | TCP/HTTP 协议栈面试术语 15 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| HPACK | TCP/HTTP 协议栈面试术语 16 | [10 网络](10-网络编程与简易HTTP服务.md) |
| 粘包 | TCP/HTTP 协议栈面试术语 17 | [23 IO](23-IO多路复用与高性能Server.md) |
| Content-Length | TCP/HTTP 协议栈面试术语 18 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| chunked | TCP/HTTP 协议栈面试术语 19 | [10 网络](10-网络编程与简易HTTP服务.md) |
| 502 | TCP/HTTP 协议栈面试术语 20 | [23 IO](23-IO多路复用与高性能Server.md) |
| 504 | TCP/HTTP 协议栈面试术语 21 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| CORS | TCP/HTTP 协议栈面试术语 22 | [10 网络](10-网络编程与简易HTTP服务.md) |
| Cookie | TCP/HTTP 协议栈面试术语 23 | [23 IO](23-IO多路复用与高性能Server.md) |
| CDN | TCP/HTTP 协议栈面试术语 24 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| sendfile | TCP/HTTP 协议栈面试术语 25 | [10 网络](10-网络编程与简易HTTP服务.md) |
| splice | TCP/HTTP 协议栈面试术语 26 | [23 IO](23-IO多路复用与高性能Server.md) |
| DMA | TCP/HTTP 协议栈面试术语 27 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| WebSocket | TCP/HTTP 协议栈面试术语 28 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SSE | TCP/HTTP 协议栈面试术语 29 | [23 IO](23-IO多路复用与高性能Server.md) |
| ALPN | TCP/HTTP 协议栈面试术语 30 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| SNI | TCP/HTTP 协议栈面试术语 31 | [10 网络](10-网络编程与简易HTTP服务.md) |
| mTLS | TCP/HTTP 协议栈面试术语 32 | [23 IO](23-IO多路复用与高性能Server.md) |
| HSTS | TCP/HTTP 协议栈面试术语 33 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| TCP Fast Open | TCP/HTTP 协议栈面试术语 34 | [10 网络](10-网络编程与简易HTTP服务.md) |
| BBR | TCP/HTTP 协议栈面试术语 35 | [23 IO](23-IO多路复用与高性能Server.md) |
| SO_REUSEPORT | TCP/HTTP 协议栈面试术语 36 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| backlog | TCP/HTTP 协议栈面试术语 37 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SYN-37 | TCP/HTTP 协议栈面试术语 38 | [23 IO](23-IO多路复用与高性能Server.md) |
| ACK-38 | TCP/HTTP 协议栈面试术语 39 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| FIN-39 | TCP/HTTP 协议栈面试术语 40 | [10 网络](10-网络编程与简易HTTP服务.md) |
| TIME_WAIT-40 | TCP/HTTP 协议栈面试术语 41 | [23 IO](23-IO多路复用与高性能Server.md) |
| CLOSE_WAIT-41 | TCP/HTTP 协议栈面试术语 42 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| rwnd-42 | TCP/HTTP 协议栈面试术语 43 | [10 网络](10-网络编程与简易HTTP服务.md) |
| cwnd-43 | TCP/HTTP 协议栈面试术语 44 | [23 IO](23-IO多路复用与高性能Server.md) |
| SACK-44 | TCP/HTTP 协议栈面试术语 45 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| RTO-45 | TCP/HTTP 协议栈面试术语 46 | [10 网络](10-网络编程与简易HTTP服务.md) |
| Nagle-46 | TCP/HTTP 协议栈面试术语 47 | [23 IO](23-IO多路复用与高性能Server.md) |
| HTTP/1.1-47 | TCP/HTTP 协议栈面试术语 48 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| HTTP/2-48 | TCP/HTTP 协议栈面试术语 49 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HTTP/3-49 | TCP/HTTP 协议栈面试术语 50 | [23 IO](23-IO多路复用与高性能Server.md) |
| QUIC-50 | TCP/HTTP 协议栈面试术语 51 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| TLS-51 | TCP/HTTP 协议栈面试术语 52 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HPACK-52 | TCP/HTTP 协议栈面试术语 53 | [23 IO](23-IO多路复用与高性能Server.md) |
| 粘包-53 | TCP/HTTP 协议栈面试术语 54 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| Content-Length-54 | TCP/HTTP 协议栈面试术语 55 | [10 网络](10-网络编程与简易HTTP服务.md) |
| chunked-55 | TCP/HTTP 协议栈面试术语 56 | [23 IO](23-IO多路复用与高性能Server.md) |
| 502-56 | TCP/HTTP 协议栈面试术语 57 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| 504-57 | TCP/HTTP 协议栈面试术语 58 | [10 网络](10-网络编程与简易HTTP服务.md) |
| CORS-58 | TCP/HTTP 协议栈面试术语 59 | [23 IO](23-IO多路复用与高性能Server.md) |
| Cookie-59 | TCP/HTTP 协议栈面试术语 60 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| CDN-60 | TCP/HTTP 协议栈面试术语 61 | [10 网络](10-网络编程与简易HTTP服务.md) |
| sendfile-61 | TCP/HTTP 协议栈面试术语 62 | [23 IO](23-IO多路复用与高性能Server.md) |
| splice-62 | TCP/HTTP 协议栈面试术语 63 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| DMA-63 | TCP/HTTP 协议栈面试术语 64 | [10 网络](10-网络编程与简易HTTP服务.md) |
| WebSocket-64 | TCP/HTTP 协议栈面试术语 65 | [23 IO](23-IO多路复用与高性能Server.md) |
| SSE-65 | TCP/HTTP 协议栈面试术语 66 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| ALPN-66 | TCP/HTTP 协议栈面试术语 67 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SNI-67 | TCP/HTTP 协议栈面试术语 68 | [23 IO](23-IO多路复用与高性能Server.md) |
| mTLS-68 | TCP/HTTP 协议栈面试术语 69 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| HSTS-69 | TCP/HTTP 协议栈面试术语 70 | [10 网络](10-网络编程与简易HTTP服务.md) |
| TCP Fast Open-70 | TCP/HTTP 协议栈面试术语 71 | [23 IO](23-IO多路复用与高性能Server.md) |
| BBR-71 | TCP/HTTP 协议栈面试术语 72 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| SO_REUSEPORT-72 | TCP/HTTP 协议栈面试术语 73 | [10 网络](10-网络编程与简易HTTP服务.md) |
| backlog-73 | TCP/HTTP 协议栈面试术语 74 | [23 IO](23-IO多路复用与高性能Server.md) |
| SYN-74 | TCP/HTTP 协议栈面试术语 75 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| ACK-75 | TCP/HTTP 协议栈面试术语 76 | [10 网络](10-网络编程与简易HTTP服务.md) |
| FIN-76 | TCP/HTTP 协议栈面试术语 77 | [23 IO](23-IO多路复用与高性能Server.md) |
| TIME_WAIT-77 | TCP/HTTP 协议栈面试术语 78 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| CLOSE_WAIT-78 | TCP/HTTP 协议栈面试术语 79 | [10 网络](10-网络编程与简易HTTP服务.md) |
| rwnd-79 | TCP/HTTP 协议栈面试术语 80 | [23 IO](23-IO多路复用与高性能Server.md) |
| cwnd-80 | TCP/HTTP 协议栈面试术语 81 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| SACK-81 | TCP/HTTP 协议栈面试术语 82 | [10 网络](10-网络编程与简易HTTP服务.md) |
| RTO-82 | TCP/HTTP 协议栈面试术语 83 | [23 IO](23-IO多路复用与高性能Server.md) |
| Nagle-83 | TCP/HTTP 协议栈面试术语 84 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| HTTP/1.1-84 | TCP/HTTP 协议栈面试术语 85 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HTTP/2-85 | TCP/HTTP 协议栈面试术语 86 | [23 IO](23-IO多路复用与高性能Server.md) |
| HTTP/3-86 | TCP/HTTP 协议栈面试术语 87 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| QUIC-87 | TCP/HTTP 协议栈面试术语 88 | [10 网络](10-网络编程与简易HTTP服务.md) |
| TLS-88 | TCP/HTTP 协议栈面试术语 89 | [23 IO](23-IO多路复用与高性能Server.md) |
| HPACK-89 | TCP/HTTP 协议栈面试术语 90 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| 粘包-90 | TCP/HTTP 协议栈面试术语 91 | [10 网络](10-网络编程与简易HTTP服务.md) |
| Content-Length-91 | TCP/HTTP 协议栈面试术语 92 | [23 IO](23-IO多路复用与高性能Server.md) |
| chunked-92 | TCP/HTTP 协议栈面试术语 93 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| 502-93 | TCP/HTTP 协议栈面试术语 94 | [10 网络](10-网络编程与简易HTTP服务.md) |
| 504-94 | TCP/HTTP 协议栈面试术语 95 | [23 IO](23-IO多路复用与高性能Server.md) |
| CORS-95 | TCP/HTTP 协议栈面试术语 96 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| Cookie-96 | TCP/HTTP 协议栈面试术语 97 | [10 网络](10-网络编程与简易HTTP服务.md) |
| CDN-97 | TCP/HTTP 协议栈面试术语 98 | [23 IO](23-IO多路复用与高性能Server.md) |
| sendfile-98 | TCP/HTTP 协议栈面试术语 99 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| splice-99 | TCP/HTTP 协议栈面试术语 100 | [10 网络](10-网络编程与简易HTTP服务.md) |
| DMA-100 | TCP/HTTP 协议栈面试术语 101 | [23 IO](23-IO多路复用与高性能Server.md) |
| WebSocket-101 | TCP/HTTP 协议栈面试术语 102 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| SSE-102 | TCP/HTTP 协议栈面试术语 103 | [10 网络](10-网络编程与简易HTTP服务.md) |
| ALPN-103 | TCP/HTTP 协议栈面试术语 104 | [23 IO](23-IO多路复用与高性能Server.md) |
| SNI-104 | TCP/HTTP 协议栈面试术语 105 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| mTLS-105 | TCP/HTTP 协议栈面试术语 106 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HSTS-106 | TCP/HTTP 协议栈面试术语 107 | [23 IO](23-IO多路复用与高性能Server.md) |
| TCP Fast Open-107 | TCP/HTTP 协议栈面试术语 108 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| BBR-108 | TCP/HTTP 协议栈面试术语 109 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SO_REUSEPORT-109 | TCP/HTTP 协议栈面试术语 110 | [23 IO](23-IO多路复用与高性能Server.md) |
| backlog-110 | TCP/HTTP 协议栈面试术语 111 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| SYN-111 | TCP/HTTP 协议栈面试术语 112 | [10 网络](10-网络编程与简易HTTP服务.md) |
| ACK-112 | TCP/HTTP 协议栈面试术语 113 | [23 IO](23-IO多路复用与高性能Server.md) |
| FIN-113 | TCP/HTTP 协议栈面试术语 114 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| TIME_WAIT-114 | TCP/HTTP 协议栈面试术语 115 | [10 网络](10-网络编程与简易HTTP服务.md) |
| CLOSE_WAIT-115 | TCP/HTTP 协议栈面试术语 116 | [23 IO](23-IO多路复用与高性能Server.md) |
| rwnd-116 | TCP/HTTP 协议栈面试术语 117 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| cwnd-117 | TCP/HTTP 协议栈面试术语 118 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SACK-118 | TCP/HTTP 协议栈面试术语 119 | [23 IO](23-IO多路复用与高性能Server.md) |
| RTO-119 | TCP/HTTP 协议栈面试术语 120 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

## §7 术语速查表 3


| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| SYN | TCP/HTTP 协议栈面试术语 1 | [10 网络](10-网络编程与简易HTTP服务.md) |
| ACK | TCP/HTTP 协议栈面试术语 2 | [23 IO](23-IO多路复用与高性能Server.md) |
| FIN | TCP/HTTP 协议栈面试术语 3 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| TIME_WAIT | TCP/HTTP 协议栈面试术语 4 | [10 网络](10-网络编程与简易HTTP服务.md) |
| CLOSE_WAIT | TCP/HTTP 协议栈面试术语 5 | [23 IO](23-IO多路复用与高性能Server.md) |
| rwnd | TCP/HTTP 协议栈面试术语 6 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| cwnd | TCP/HTTP 协议栈面试术语 7 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SACK | TCP/HTTP 协议栈面试术语 8 | [23 IO](23-IO多路复用与高性能Server.md) |
| RTO | TCP/HTTP 协议栈面试术语 9 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| Nagle | TCP/HTTP 协议栈面试术语 10 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HTTP/1.1 | TCP/HTTP 协议栈面试术语 11 | [23 IO](23-IO多路复用与高性能Server.md) |
| HTTP/2 | TCP/HTTP 协议栈面试术语 12 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| HTTP/3 | TCP/HTTP 协议栈面试术语 13 | [10 网络](10-网络编程与简易HTTP服务.md) |
| QUIC | TCP/HTTP 协议栈面试术语 14 | [23 IO](23-IO多路复用与高性能Server.md) |
| TLS | TCP/HTTP 协议栈面试术语 15 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| HPACK | TCP/HTTP 协议栈面试术语 16 | [10 网络](10-网络编程与简易HTTP服务.md) |
| 粘包 | TCP/HTTP 协议栈面试术语 17 | [23 IO](23-IO多路复用与高性能Server.md) |
| Content-Length | TCP/HTTP 协议栈面试术语 18 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| chunked | TCP/HTTP 协议栈面试术语 19 | [10 网络](10-网络编程与简易HTTP服务.md) |
| 502 | TCP/HTTP 协议栈面试术语 20 | [23 IO](23-IO多路复用与高性能Server.md) |
| 504 | TCP/HTTP 协议栈面试术语 21 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| CORS | TCP/HTTP 协议栈面试术语 22 | [10 网络](10-网络编程与简易HTTP服务.md) |
| Cookie | TCP/HTTP 协议栈面试术语 23 | [23 IO](23-IO多路复用与高性能Server.md) |
| CDN | TCP/HTTP 协议栈面试术语 24 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| sendfile | TCP/HTTP 协议栈面试术语 25 | [10 网络](10-网络编程与简易HTTP服务.md) |
| splice | TCP/HTTP 协议栈面试术语 26 | [23 IO](23-IO多路复用与高性能Server.md) |
| DMA | TCP/HTTP 协议栈面试术语 27 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| WebSocket | TCP/HTTP 协议栈面试术语 28 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SSE | TCP/HTTP 协议栈面试术语 29 | [23 IO](23-IO多路复用与高性能Server.md) |
| ALPN | TCP/HTTP 协议栈面试术语 30 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| SNI | TCP/HTTP 协议栈面试术语 31 | [10 网络](10-网络编程与简易HTTP服务.md) |
| mTLS | TCP/HTTP 协议栈面试术语 32 | [23 IO](23-IO多路复用与高性能Server.md) |
| HSTS | TCP/HTTP 协议栈面试术语 33 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| TCP Fast Open | TCP/HTTP 协议栈面试术语 34 | [10 网络](10-网络编程与简易HTTP服务.md) |
| BBR | TCP/HTTP 协议栈面试术语 35 | [23 IO](23-IO多路复用与高性能Server.md) |
| SO_REUSEPORT | TCP/HTTP 协议栈面试术语 36 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| backlog | TCP/HTTP 协议栈面试术语 37 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SYN-37 | TCP/HTTP 协议栈面试术语 38 | [23 IO](23-IO多路复用与高性能Server.md) |
| ACK-38 | TCP/HTTP 协议栈面试术语 39 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| FIN-39 | TCP/HTTP 协议栈面试术语 40 | [10 网络](10-网络编程与简易HTTP服务.md) |
| TIME_WAIT-40 | TCP/HTTP 协议栈面试术语 41 | [23 IO](23-IO多路复用与高性能Server.md) |
| CLOSE_WAIT-41 | TCP/HTTP 协议栈面试术语 42 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| rwnd-42 | TCP/HTTP 协议栈面试术语 43 | [10 网络](10-网络编程与简易HTTP服务.md) |
| cwnd-43 | TCP/HTTP 协议栈面试术语 44 | [23 IO](23-IO多路复用与高性能Server.md) |
| SACK-44 | TCP/HTTP 协议栈面试术语 45 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| RTO-45 | TCP/HTTP 协议栈面试术语 46 | [10 网络](10-网络编程与简易HTTP服务.md) |
| Nagle-46 | TCP/HTTP 协议栈面试术语 47 | [23 IO](23-IO多路复用与高性能Server.md) |
| HTTP/1.1-47 | TCP/HTTP 协议栈面试术语 48 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| HTTP/2-48 | TCP/HTTP 协议栈面试术语 49 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HTTP/3-49 | TCP/HTTP 协议栈面试术语 50 | [23 IO](23-IO多路复用与高性能Server.md) |
| QUIC-50 | TCP/HTTP 协议栈面试术语 51 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| TLS-51 | TCP/HTTP 协议栈面试术语 52 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HPACK-52 | TCP/HTTP 协议栈面试术语 53 | [23 IO](23-IO多路复用与高性能Server.md) |
| 粘包-53 | TCP/HTTP 协议栈面试术语 54 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| Content-Length-54 | TCP/HTTP 协议栈面试术语 55 | [10 网络](10-网络编程与简易HTTP服务.md) |
| chunked-55 | TCP/HTTP 协议栈面试术语 56 | [23 IO](23-IO多路复用与高性能Server.md) |
| 502-56 | TCP/HTTP 协议栈面试术语 57 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| 504-57 | TCP/HTTP 协议栈面试术语 58 | [10 网络](10-网络编程与简易HTTP服务.md) |
| CORS-58 | TCP/HTTP 协议栈面试术语 59 | [23 IO](23-IO多路复用与高性能Server.md) |
| Cookie-59 | TCP/HTTP 协议栈面试术语 60 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| CDN-60 | TCP/HTTP 协议栈面试术语 61 | [10 网络](10-网络编程与简易HTTP服务.md) |
| sendfile-61 | TCP/HTTP 协议栈面试术语 62 | [23 IO](23-IO多路复用与高性能Server.md) |
| splice-62 | TCP/HTTP 协议栈面试术语 63 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| DMA-63 | TCP/HTTP 协议栈面试术语 64 | [10 网络](10-网络编程与简易HTTP服务.md) |
| WebSocket-64 | TCP/HTTP 协议栈面试术语 65 | [23 IO](23-IO多路复用与高性能Server.md) |
| SSE-65 | TCP/HTTP 协议栈面试术语 66 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| ALPN-66 | TCP/HTTP 协议栈面试术语 67 | [10 网络](10-网络编程与简易HTTP服务.md) |
| SNI-67 | TCP/HTTP 协议栈面试术语 68 | [23 IO](23-IO多路复用与高性能Server.md) |
| mTLS-68 | TCP/HTTP 协议栈面试术语 69 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| HSTS-69 | TCP/HTTP 协议栈面试术语 70 | [10 网络](10-网络编程与简易HTTP服务.md) |
| TCP Fast Open-70 | TCP/HTTP 协议栈面试术语 71 | [23 IO](23-IO多路复用与高性能Server.md) |
| BBR-71 | TCP/HTTP 协议栈面试术语 72 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| SO_REUSEPORT-72 | TCP/HTTP 协议栈面试术语 73 | [10 网络](10-网络编程与简易HTTP服务.md) |
| backlog-73 | TCP/HTTP 协议栈面试术语 74 | [23 IO](23-IO多路复用与高性能Server.md) |
| SYN-74 | TCP/HTTP 协议栈面试术语 75 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| ACK-75 | TCP/HTTP 协议栈面试术语 76 | [10 网络](10-网络编程与简易HTTP服务.md) |
| FIN-76 | TCP/HTTP 协议栈面试术语 77 | [23 IO](23-IO多路复用与高性能Server.md) |
| TIME_WAIT-77 | TCP/HTTP 协议栈面试术语 78 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| CLOSE_WAIT-78 | TCP/HTTP 协议栈面试术语 79 | [10 网络](10-网络编程与简易HTTP服务.md) |
| rwnd-79 | TCP/HTTP 协议栈面试术语 80 | [23 IO](23-IO多路复用与高性能Server.md) |
| cwnd-80 | TCP/HTTP 协议栈面试术语 81 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| SACK-81 | TCP/HTTP 协议栈面试术语 82 | [10 网络](10-网络编程与简易HTTP服务.md) |
| RTO-82 | TCP/HTTP 协议栈面试术语 83 | [23 IO](23-IO多路复用与高性能Server.md) |
| Nagle-83 | TCP/HTTP 协议栈面试术语 84 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| HTTP/1.1-84 | TCP/HTTP 协议栈面试术语 85 | [10 网络](10-网络编程与简易HTTP服务.md) |
| HTTP/2-85 | TCP/HTTP 协议栈面试术语 86 | [23 IO](23-IO多路复用与高性能Server.md) |
| HTTP/3-86 | TCP/HTTP 协议栈面试术语 87 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| QUIC-87 | TCP/HTTP 协议栈面试术语 88 | [10 网络](10-网络编程与简易HTTP服务.md) |
| TLS-88 | TCP/HTTP 协议栈面试术语 89 | [23 IO](23-IO多路复用与高性能Server.md) |
| HPACK-89 | TCP/HTTP 协议栈面试术语 90 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |

| 术语 | 一句话 | 关联章节 |
|------|--------|----------|
| 粘包-90 | TCP/HTTP 协议栈面试术语 91 | [10 网络](10-网络编程与简易HTTP服务.md) |
| Content-Length-91 | TCP/HTTP 协议栈面试术语 92 | [23 IO](23-IO多路复用与高性能Server.md) |
| chunked-92 | TCP/HTTP 协议栈面试术语 93 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| 502-93 | TCP/HTTP 协议栈面试术语 94 | [10 网络](10-网络编程与简易HTTP服务.md) |
| 504-94 | TCP/HTTP 协议栈面试术语 95 | [23 IO](23-IO多路复用与高性能Server.md) |
| CORS-95 | TCP/HTTP 协议栈面试术语 96 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| Cookie-96 | TCP/HTTP 协议栈面试术语 97 | [10 网络](10-网络编程与简易HTTP服务.md) |
| CDN-97 | TCP/HTTP 协议栈面试术语 98 | [23 IO](23-IO多路复用与高性能Server.md) |
| sendfile-98 | TCP/HTTP 协议栈面试术语 99 | [计网02](../../前端学习/计算机网络/02-TCP与UDP.md) |
| splice-99 | TCP/HTTP 协议栈面试术语 100 | [10 网络](10-网络编程与简易HTTP服务.md) |

## §8 闭卷自测（10 题 + 参考答案）

| 题号 | 题目 | ☐ |
|------|------|---|
| 1 | TCP 三次握手过程 | ☐ |
| 2 | TCP 四次挥手与 TIME_WAIT | ☐ |
| 3 | 滑动窗口与流量控制 | ☐ |
| 4 | 拥塞控制慢启动/快重传 | ☐ |
| 5 | HTTP/2 相对 HTTP/1.1 改进 | ☐ |
| 6 | HTTPS 协议栈层次 | ☐ |
| 7 | TCP 粘包及 C++ 处理方式 | ☐ |
| 8 | sendfile 零拷贝原理 | ☐ |
| 9 | CLOSE_WAIT 过多如何排查 | ☐ |
| 10 | 502 vs 504 区别 | ☐ |

### 参考答案

1. SYN → SYN+ACK → ACK；同步初始序列号、交换 MSS。  
2. 双方各发 FIN+ACK；主动关闭方进入 TIME_WAIT 等待 2MSL。  
3. 接收方 rwnd 限制发送方在途字节数，防止接收缓冲区溢出。  
4. cwnd 慢启动指数增；3 个 dup ACK 触发快重传；NewReno 快恢复。  
5. 二进制分帧、多路复用、HPACK 头压缩；TCP 层队头阻塞仍在。  
6. HTTP over TLS over TCP；TLS 1.3 可 1-RTT 握手。  
7. TCP 无消息边界；HTTP 用 Content-Length/chunked；自定义协议用长度前缀 [10章](10-网络编程与简易HTTP服务.md)。  
8. 数据在内核 fd→socket 间拷贝，不经用户态 buffer。  
9. ss/netstat 看 CLOSE_WAIT；查应用是否未 close fd（泄漏）。  
10. 502 网关收到上游无效响应；504 网关等待上游超时。


**系列完结** 51→52→53→**54**
