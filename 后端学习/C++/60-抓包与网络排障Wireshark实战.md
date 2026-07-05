# 抓包与网络排障 Wireshark 实战

> **文件编码**：UTF-8。tcpdump、Wireshark 过滤、三次握手、TLS、延迟排查、C++10/23 网络调试
> **交叉阅读**：[54 计网](54-计算机网络TCP与HTTP面试深度专章.md) · [10 网络编程](10-网络编程与简易HTTP服务.md) · [23 IO 多路复用](23-IO多路复用与高性能Server.md) · [12 性能分析](12-性能分析与调试.md)

---

## 本章与前后章的关系

| 上一章 | 本章 | 下一章 |
|--------|------|--------|
| [59 分布式](59-分布式理论CAP-Raft与共识算法面试.md) | **本章** | [61 线上排障](61-线上故障排查与性能诊断实战.md) |

**学习链扩展（51～63）**：

| [51 MySQL](51-MySQL原理与索引事务面试专章.md) | [52 Redis](52-Redis数据结构与缓存面试专章.md) | [53 OS](53-操作系统面试八股与口述模板.md) |
| [54 计网](54-计算机网络TCP与HTTP面试深度专章.md) | [55 笔试](55-大厂C++笔试选择题与代码输出陷阱题集.md) | [56 系统设计](56-系统设计案例库RPC-KV与限流秒杀.md) |
| [57 Kafka](57-消息队列Kafka与中间件面试专题.md) | [58 模拟面试](58-模拟面试完整流程与压测数据模板.md) | [59 分布式](59-分布式理论CAP-Raft与共识算法面试.md) |
| [60 抓包](60-抓包与网络排障Wireshark实战.md) | [61 排障](61-线上故障排查与性能诊断实战.md) | [62 K8s](62-Docker与Kubernetes入门面试.md) |
| [63 JWT 幂等](63-JWT认证与接口幂等性实战.md) | | |

```mermaid
flowchart LR
  A[51 MySQL] --> B[52 Redis]
  B --> C[53 OS]
  C --> D[54 计网]
  D --> E[55 笔试]
  E --> F[56 系统设计]
  F --> G[57 Kafka]
  G --> H[58 模拟]
  H --> I[59 分布式]
  I --> J[60 抓包]
  J --> K[61 排障]
  K --> L[62 K8s]
  L --> M[63 JWT]
```

---

## §0 读前导读

### §0.1 用一句话弄懂本章

抓包排障 = **tcpdump 采集 → Wireshark 过滤定位 → 协议层（TCP/TLS/HTTP）→ 与 C++ 服务（[10/23 章](10-网络编程与简易HTTP服务.md)）指标互证**；是 [59 章](59-分布式理论CAP-Raft与共识算法面试.md) RPC 超时与 [61 章](61-线上故障排查与性能诊断实战.md) 延迟诊断的前置技能。

### §0.2 你需要提前知道什么

| 状态 | 动作 |
|------|------|
| 只会用不会讲 | 每节 Q&A 限时 2min 口述 |
| C++ 后端岗 | 必串 [08 多线程](08-多线程与并发编程.md) [10 网络](10-网络编程与简易HTTP服务.md) [23 IO](23-IO多路复用与高性能Server.md) |
| 前置章节 | [59 分布式](59-分布式理论CAP-Raft与共识算法面试.md) Raft RPC 超时 |
| 后续章节 | [61 线上排障](61-线上故障排查与性能诊断实战.md) 全链路诊断 |

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
| 交叉刷 | 2h | 51～58 + C++ 工程文档 |
| 闭卷自测 | 1h | ≥8/10 |

---

## §1 tcpdump 实战

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | tcpdump 基本语法？ | -i 网卡 -nn 不解析域名端口 -s0 全包 -w 写文件；host/port/tcp 过滤。 | 权限 cap_net_raw |
| Q2 | 生产环境抓包原则？ | 限定时长与包数、镜像口、避免核心交换机过载；先 metadata 后 full payload。 | 合规脱敏 |
| Q3 | -s snaplen 影响？ | 过小截断 TCP payload 导致 TLS/HTTP 解不开；一般 -s0 或 262144。 | 磁盘空间 |
| Q4 | 如何只抓 SYN？ | tcpdump 'tcp[tcpflags] & tcp-syn != 0' | SYN flood 排查 |
| Q5 | 如何抓特定连接？ | host X and host Y and port 443 | 四元组 |
| Q6 | rotate 文件？ | -C 100 -W 5 每 100MB 轮转保留 5 个 | 长时间抓包 |
| Q7 | 与 tshark 关系？ | tshark CLI 版 Wireshark；批处理解析 pcap。 | 脚本化 |
| Q8 | 容器内抓包？ | nsenter 进 net namespace；或 sidecar tcpdump。 | [62 K8s](62-Docker与Kubernetes入门面试.md) |
| Q9 | 丢包如何观察？ | ifconfig drops、ethtool -S、tcpdump 计数 vs 应用收包。 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Q10 | C++ 服务配合？ | 打 correlation id；日志时间戳对齐 pcap。 | [32 可观测](32-fmt-spdlog与可观测性工程.md) |

## §2 Wireshark 过滤与分析

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | Wireshark 显示过滤 vs 捕获过滤？ | 捕获过滤 BPF 减体积；显示过滤强大可后改不改抓包。 | 语法不同 |
| Q2 | 常用显示过滤？ | ip.addr==、tcp.port==、http.request、tls.handshake.type==1 | 组合 and/or |
| Q3 | Follow TCP Stream？ | 重组应用层字节流看 HTTP/Redis 文本协议。 | 二进制需 hex |
| Q4 | Expert Info 颜色？ | 红 Error 重传；黄 Warn 乱序；蓝 Note。 | 快速定位 |
| Q5 | RTT 如何看？ | TCP Stream Graph → Round Trip Time；或 tcp.analysis.ack_rtt | 区分 SYN RTT |
| Q6 | 重传类型？ | Retransmission、Fast Retransmission、Spurious Retransmission。 | 三次 dup ACK |
| Q7 | ZeroWindow？ | 接收端 buffer 满；背压；查应用 read 慢。 | [23 epoll](23-IO多路复用与高性能Server.md) |
| Q8 | TLS 解密前提？ | 有 server/client key log（SSLKEYLOGFILE）或私钥。 | 前向保密限制 |
| Q9 | HTTP/2 分析？ | 帧级 decode；SETTINGS、WINDOW_UPDATE、RST_STREAM。 | [54 HTTP](54-计算机网络TCP与HTTP面试深度专章.md) |
| Q10 | Statistics IO Graph？ | 吞吐随时间；对齐业务高峰。 | 压测 [58 章](58-模拟面试完整流程与压测数据模板.md) |

## §3 三次握手与 TCP 状态

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 三次握手包序？ | SYN → SYN-ACK → ACK；随后可 TLS ClientHello。 | SYN cookie |
| Q2 | 半连接队列？ | SYN_RECV 堆积；syncookies/net.ipv4.tcp_max_syn_backlog。 | SYN flood |
| Q3 | 全连接队列？ | accept 队列满丢 ACK 或发 RST；ss -ltn 看 Recv-Q。 | backlog 调参 |
| Q4 | TIME_WAIT 过多？ | 短连接高频；tcp_tw_reuse（谨慎）；长连接+连接池。 | [51 连接池](51-MySQL原理与索引事务面试专章.md) |
| Q5 | 四次挥手抓包特征？ | FIN/ACK 交错；CLOSE_WAIT 泄漏查代码未 close。 | C++ RAII socket |
| Q6 | MSS 协商？ | SYN 选项 MSS；PMTU 黑洞 DF 位。 | ICMP frag needed |
| Q7 | 窗口缩放 wscale？ | 高 BDP 网络必需；否则吞吐上不去。 | 长 fat 网络 |
| Q8 | SACK 作用？ | 选择性确认减少重传量。 | 乱序场景 |
| Q9 | Keep-Alive？ | TCP 层探测；与应用 heartbeat 不同。 | 中间件超时 |
| Q10 | 抓包看连接失败？ | RST 原因：端口未监听、防火墙、半开超限。 | telnc 对比 |

## §4 TLS 抓包与解密

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | TLS 握手阶段？ | ClientHello→ServerHello/Cert/...→ClientKeyExchange→Finished。 | 1-RTT vs 0-RTT |
| Q2 | TLS 1.3 简化？ | 1-RTT 握手；0-RTT 有重放风险。 | 面试常问 |
| Q3 | 证书链验证失败？ | UNKNOWN CA、hostname mismatch、过期。 | 抓包仍见 Encrypted |
| Q4 | ALPN 协商？ | h2 vs http/1.1 in ClientHello。 | gRPC h2 |
| Q5 | Session Resumption？ | Session ID / Ticket 减少握手 RTT。 | 性能优化 |
| Q6 | mTLS 双向？ | 客户端也带 cert；服务网格常见。 | Istio |
| Q7 | 解密 workflow？ | 设置 (Pre)-Master-Secret log path；Reload。 | 生产 rarely |
| Q8 | TLS 与性能？ | CPU 加解密；AES-NI；session reuse。 | [18 高性能](18-高性能C++与内存对齐.md) |
| Q9 | Application Data 看不到？ | 正常；只能看明文侧或 key log。 | 安全 |
| Q10 | 与 [63 JWT](63-JWT认证与接口幂等性实战.md)？ | JWT 在 HTTP Authorization；TLS 管传输加密。 | 分层 |

## §5 延迟排查方法论

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 延迟拆解四段？ | DNS、TCP/TLS 握手、TTFB、下载 body。 | curl -w 模板 |
| Q2 | 定位高 TTFB？ | 服务端慢 vs 网络；抓包看 request 到 first byte 间隔。 | [61 perf](61-线上故障排查与性能诊断实战.md) |
| Q3 | 重传率指标？ | retrans/sec / total packets；>1% 需查。 | 无线/跨境 |
| Q4 | 乱序原因？ | 多路径、负载均衡 L4；适度可接受。 | 缓冲 |
| Q5 | Nagle 与延迟？ | TCP_NODELAY 禁 Nagle 降小包延迟；trade-off 吞吐。 | C++ setsockopt |
| Q6 | C++ 侧 timestamp？ | steady_clock 打点到日志；与 pcap 时间对齐 UTC。 | chrono |
| Q7 | 跨 AZ 延迟？ | 物理距离+排队；Raft 选主见 [59 章](59-分布式理论CAP-Raft与共识算法面试.md)。 | RTT 预算 |
| Q8 | Redis 延迟抓包？ | RESP 文本协议易读；慢命令 vs 网络。 | [52 Redis](52-Redis数据结构与缓存面试专章.md) |
| Q9 | MySQL 抓包？ | Wireshark mysql dissector；慢查询互证。 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Q10 | 全链路 trace？ | OpenTelemetry span + pcap 单请求；理想态。 | [32 可观测](32-fmt-spdlog与可观测性工程.md) |

## §6 C++10/20/23 网络调试

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | C++10 chrono 打点？ | auto t0=steady_clock::now(); ... duration_cast<microseconds>(t1-t0) | 日志格式 |
| Q2 | C++20 coroutine 网络？ | co_await async_read；超时与 cancel 需抓包验证。 | [31 协程](31-协程C++20-coroutine.md) |
| Q3 | C++23 expected？ | 错误路径可观测；对齐 HTTP status 与 RST。 | 类型安全 |
| Q4 | Boost.Asio 超时？ | deadline_timer + async_op；抓包看是否 half-open。 | [26 Asio](26-Boost.Asio异步网络编程.md) |
| Q5 | setsockopt 常用？ | TCP_NODELAY、SO_KEEPALIVE、SO_RCVBUF/SO_SNDBUF。 | 与窗口 |
| Q6 | 非阻塞 EAGAIN？ | 抓包仍有数据说明未 read；epoll 边沿误用。 | [23 epoll](23-IO多路复用与高性能Server.md) |
| Q7 | 连接池泄漏？ | CLOSE_WAIT 增多；valgrind 不如 pcap 直观。 | [61 章](61-线上故障排查与性能诊断实战.md) |
| Q8 | gRPC 调试？ | grpc trace + tcpdump 443；HTTP/2 RST。 | [19 gRPC](19-gRPC与Protobuf工程化.md) |
| Q9 | 单元测试 mock 网络？ | 不替代集成抓包；本地 loopback pcap。 | [27 GTest](27-Google-Test与单元测试工程.md) |
| Q10 | Instrument 编译选项？ | -g 保留符号；Release 也可抓包不看栈。 | [12 调试](12-性能分析与调试.md) |

## §7 STAR 案例

### §7.1 STAR 案例 1

**S（情境）**：用户报告 API P99 3s，日志显示 handler 50ms。

**T（任务）**：区分网络 vs 服务端。

**A（行动）**：tcpdump 看重传与 RTT；发现 LB 到 pod SYN 重传；调 syn backlog 与 health check。

**R（结果）**：P99 降到 120ms。

**连环追问**：
- 如何证明是网络？
- ZeroWindow 见过吗？
- 与 [59 Raft](59-分布式理论CAP-Raft与共识算法面试.md) 选举超时区别？
## §8 命令速查表

```bash
# 抓 HTTPS 443，60 秒，10 万包
sudo tcpdump -i any -nn -s0 -c 100000 -w api.pcap 'tcp port 443'

# Wireshark 显示过滤：某 trace 对应 IP
tcp.stream eq 3 && ip.addr == 10.0.1.5

# curl 延迟分解
curl -w '@curl-format.txt' -o /dev/null -s https://api.example.com/health
```

```cpp
// C++17：请求级延迟日志，便于与 Wireshark 对齐
#include <chrono>
#include <string>
#include <spdlog/spdlog.h>

struct ScopeLatency {
    std::string trace_id;
    std::chrono::steady_clock::time_point t0{std::chrono::steady_clock::now()};
    ~ScopeLatency() {
        auto us = std::chrono::duration_cast<std::chrono::microseconds>(
            std::chrono::steady_clock::now() - t0).count();
        spdlog::info("trace={} latency_us={}", trace_id, us);
    }
};

// 使用：ScopeLatency _{req.trace_id};  // 见 32 章 spdlog
```

## §9 口述模板（2 分钟版）

### §9.1 抓包三板斧

先 **tcpdump 限定四元组** → Wireshark **Expert Info 看重传** → **Follow Stream** 看应用协议。对齐 C++ 日志 trace id。

### §9.2 三次握手口述

SYN 客户端 seq=x；SYN-ACK seq=y ack=x+1；ACK ack=y+1。队列满表现为 SYN 重传或 RST。

### §9.3 TLS 30 秒

握手 1-RTT(1.3)；业务 JWT 在 HTTP 层；解密需 key log。下一章 [61](61-线上故障排查与性能诊断实战.md) 结合 perf。
## §10 闭卷自测清单

- [ ] 能写出 tcpdump 抓 443 并轮转命令
- [ ] 能解释 Fast Retransmission 与超时重传
- [ ] 能在 Wireshark 过滤 http.request.method==POST
- [ ] 能读 TLS ClientHello 中 SNI
- [ ] 能用 curl -w 拆解延迟
- [ ] 能结合 [54 章](54-计算机网络TCP与HTTP面试深度专章.md) 口述 TCP 状态机
- [ ] 能说明 C++ TCP_NODELAY 场景
- [ ] 能设计 request id 与 pcap 对齐方案
- [ ] 能区分捕获过滤与显示过滤
- [ ] 能口述 TIME_WAIT 与 CLOSE_WAIT 排查
## §11 交叉索引

| 考点 | 章节 |
|------|------|
| TCP 理论 | [54](54-计算机网络TCP与HTTP面试深度专章.md) |
| epoll 服务器 | [23](23-IO多路复用与高性能Server.md) |
| 分布式 RPC | [59](59-分布式理论CAP-Raft与共识算法面试.md) |
| 全链路排障 | [61](61-线上故障排查与性能诊断实战.md) |

**下一章**：[61-线上故障排查与性能诊断实战.md](61-线上故障排查与性能诊断实战.md)
### 附录 B.1 抓包练习 1

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.2 抓包练习 2

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.3 抓包练习 3

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.4 抓包练习 4

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.5 抓包练习 5

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.6 抓包练习 6

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.7 抓包练习 7

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.8 抓包练习 8

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.9 抓包练习 9

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.10 抓包练习 10

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.11 抓包练习 11

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.12 抓包练习 12

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.13 抓包练习 13

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.14 抓包练习 14

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.15 抓包练习 15

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.16 抓包练习 16

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.17 抓包练习 17

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.18 抓包练习 18

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.19 抓包练习 19

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.20 抓包练习 20

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.21 抓包练习 21

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.22 抓包练习 22

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.23 抓包练习 23

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.24 抓包练习 24

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.25 抓包练习 25

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.26 抓包练习 26

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.27 抓包练习 27

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.28 抓包练习 28

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.29 抓包练习 29

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.30 抓包练习 30

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.31 抓包练习 31

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.32 抓包练习 32

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.33 抓包练习 33

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.34 抓包练习 34

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.35 抓包练习 35

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.36 抓包练习 36

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.37 抓包练习 37

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.38 抓包练习 38

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.39 抓包练习 39

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.40 抓包练习 40

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.41 抓包练习 41

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.42 抓包练习 42

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.43 抓包练习 43

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`

### 附录 B.44 抓包练习 44

**场景**：C++ [23 章](23-IO多路复用与高性能Server.md) epoll 服务响应慢。

**步骤**：(1) ss 看 Recv-Q (2) tcpdump 看重传 (3) perf 看 syscall (4) 对照 [61 章](61-线上故障排查与性能诊断实战.md) 火焰图。

**过滤示例**：`tcp.port == 8080 && tcp.analysis.retransmission`


## 附录扩展 Q&A（自测用）

### 自测 1

**Q1**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 1**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 2

**Q2**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 2**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 3

**Q3**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 3**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 4

**Q4**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 4**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 5

**Q5**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 5**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 6

**Q6**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 6**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 7

**Q7**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 7**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 8

**Q8**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 8**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 9

**Q9**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 9**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 10

**Q10**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 10**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 11

**Q11**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 11**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 12

**Q12**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 12**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 13

**Q13**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 13**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 14

**Q14**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 14**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 15

**Q15**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 15**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 16

**Q16**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 16**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 17

**Q17**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 17**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 18

**Q18**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 18**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 19

**Q19**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 19**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 20

**Q20**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 20**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 21

**Q21**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 21**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 22

**Q22**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 22**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 23

**Q23**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 23**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 24

**Q24**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 24**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 25

**Q25**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 25**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 26

**Q26**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 26**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 27

**Q27**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 27**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 28

**Q28**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 28**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 29

**Q29**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 29**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 30

**Q30**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 30**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 31

**Q31**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 31**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 32

**Q32**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 32**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 33

**Q33**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 33**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 34

**Q34**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 34**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 35

**Q35**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 35**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 36

**Q36**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 36**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 37

**Q37**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 37**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 38

**Q38**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 38**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 39

**Q39**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 39**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 40

**Q40**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 40**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 41

**Q41**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 41**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 42

**Q42**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 42**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

