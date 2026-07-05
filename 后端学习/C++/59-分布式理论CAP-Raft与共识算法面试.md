# 分布式理论 CAP、Raft 与共识算法面试

> **文件编码**：UTF-8。CAP、BASE、Paxos vs Raft、Leader 选举、日志复制、Quorum、etcd/ZK、NuRaft/braft、CP 系统
> **交叉阅读**：[56 系统设计](56-系统设计案例库RPC-KV与限流秒杀.md) · [57 Kafka](57-消息队列Kafka与中间件面试专题.md) · [35 KV-Store](35-项目实战高性能KV-Store.md) · [33 八股总表](33-C++Infra面试八股总表.md)

---

## 本章与前后章的关系

| 上一章 | 本章 | 下一章 |
|--------|------|--------|
| [58 模拟面试](58-模拟面试完整流程与压测数据模板.md) | **本章** | [60 抓包排障](60-抓包与网络排障Wireshark实战.md) |

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

分布式面试 = **CAP/BASE 取舍 → 共识算法（Raft/Paxos）→ 工程落地（etcd/ZK/braft）→ CP 系统怎么讲 trade-off**；C++ 岗重点 braft/NuRaft 与 [56 章](56-系统设计案例库RPC-KV与限流秒杀.md) 分布式 KV。

### §0.2 你需要提前知道什么

| 状态 | 动作 |
|------|------|
| 只会用不会讲 | 每节 Q&A 限时 2min 口述 |
| C++ 后端岗 | 必串 [08 多线程](08-多线程与并发编程.md) [10 网络](10-网络编程与简易HTTP服务.md) [23 IO](23-IO多路复用与高性能Server.md) |
| 前置章节 | [58 模拟面试](58-模拟面试完整流程与压测数据模板.md) 系统设计收尾 |
| 后续章节 | [60 抓包](60-抓包与网络排障Wireshark实战.md) 网络层验证 |

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

## §1 CAP 与 BASE 理论

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | CAP 三要素分别指什么？ | Consistency 所有节点同一时刻看到相同数据；Availability 每个请求都能得到响应（不保证最新）；Partition tolerance 网络分区时系统仍能继续工作。 | 分布式系统必须接受 P，实际在 C 与 A 间权衡 |
| Q2 | 为什么说 CAP 里 P 必选？ | 网络不可靠、消息可能丢失/延迟，分区不可避免；拒绝 P 等于假设网络永远可靠，不现实。 | CP vs AP 举例 |
| Q3 | CP 系统典型例子？ | etcd、ZooKeeper、HBase；分区时宁可拒绝写/读旧数据也要保证一致。 | 与 [51 MySQL](51-MySQL原理与索引事务面试专章.md) 主从对比 |
| Q4 | AP 系统典型例子？ | Cassandra、DynamoDB、 Eureka；分区时允许短暂不一致，保证可用。 | 最终一致窗口 |
| Q5 | BASE 与 ACID 对比？ | Basically Available + Soft state + Eventually consistent；牺牲强一致换可用与性能。 | [51 章](51-MySQL原理与索引事务面试专章.md) ACID |
| Q6 | 2PC 两阶段提交问题？ | 协调者单点、阻塞、同步等待；分区时可能脑裂与数据不一致。 | 3PC、TCC 延伸 |
| Q7 | 3PC 改进点？ | 增加 CanCommit 阶段减少阻塞，但仍无法完美解决分区。 | 工程少用 |
| Q8 | TCC 是什么？ | Try-Confirm-Cancel 业务层补偿；适合微服务分布式事务。 | 与 Saga 对比 |
| Q9 | Saga 模式？ | 长事务拆本地事务+补偿；正向执行失败则反向补偿。 | 编排 vs 协调 |
| Q10 | 分布式锁与共识关系？ | 锁需互斥+容错+可释放；ZK/etcd 基于共识实现临时顺序节点锁。 | [63 章](63-JWT认证与接口幂等性实战.md) SetNX |

## §2 Raft 核心机制

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | Raft 三大子问题？ | Leader 选举、日志复制、安全性（已提交日志不丢失）。 | Paxos 更难理解 |
| Q2 | Raft 节点状态？ | Follower、Candidate、Leader；任期 term 单调递增。 | term 作用 |
| Q3 | Leader 选举触发条件？ | Follower 选举超时未收到 Leader 心跳则变 Candidate 发起投票。 | 随机超时防分裂 |
| Q4 | 选举 quorum 规则？ | 获得多数派（⌊N/2⌋+1）选票且日志至少一样新则当选。 | 日志比较规则 |
| Q5 | 日志复制流程？ | Client→Leader 追加本地 log→并行 AppendEntries 到 Follower→多数派 match 则 commit→apply 状态机。 | 异步复制延迟 |
| Q6 | Raft 如何保证安全性？ | Leader Completeness：已提交条目必出现在新 Leader；选举限制旧日志。 | 5.4.3 论文 |
| Q7 | 脑裂如何避免？ | 多数派原则；旧 Leader 分区后无法获得多数写确认。 | fence token |
| Q8 | Raft vs Multi-Paxos？ | Raft 强 Leader、易实现；Multi-Paxos 去中心化但复杂。 | 面试优先讲 Raft |
| Q9 | 日志 compaction？ | Snapshot 压缩历史+log 增量；InstallSnapshot RPC。 | 状态机快照 |
| Q10 | 成员变更 Joint Consensus？ | Cold/Hot 配置切换经 joint 状态防双主。 | etcd learner |

## §3 Paxos 对比与 Quorum

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | Basic Paxos 两阶段？ | Prepare（promise）+ Accept（learn）；多数派才算通过。 | 活锁问题 |
| Q2 | Multi-Paxos 优化？ | 选 Proposer 当 Leader 串行化；减少 Prepare 阶段。 | 与 Raft 映射 |
| Q3 | Paxos 为何难工程化？ | 无明确 Leader 语义、日志空洞、实现细节多。 | Raft 论文动机 |
| Q4 | Leslie Lamport 经典比喻？ | 议会立法：议员提案、表决、法定人数通过。 | 口述加分 |
| Q5 | ZAB 与 Paxos？ | ZK 用 ZAB（类似 Paxos+顺序广播）；Leader _epoch+事务 id。 | Watch 机制 |
| Q6 | Raft 工业实现列表？ | etcd/raft、Consul、TiKV PD、braft、NuRaft。 | C++ 选 braft |
| Q7 | 共识算法应用场景？ | 元数据存储、配置中心、Leader 选举、复制状态机。 | [56 KV](56-系统设计案例库RPC-KV与限流秒杀.md) |
| Q8 | 拜占庭容错 BFT？ | 允许恶意节点；PBFT；区块链常用；Raft 只容错 crash。 | 性能开销 |
| Q9 | Quorum 读写公式？ | R+W>N 可保证读最新写；常见 N=3 W=2 R=2。 | 滑动窗口 quorum |
| Q10 | Linearizable vs Sequential？ | Linearizable 全局实时序；Sequential 单客户端序。 | ZK 默认哪种 |

## §4 etcd 与 ZooKeeper

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | etcd 架构组件？ | Raft 共识 + boltdb 持久化 + gRPC API + Watch/Lease。 | v3 API 前缀 |
| Q2 | etcd 典型用途？ | K8s 存储、服务发现、分布式锁、配置中心。 | [62 章](62-Docker与Kubernetes入门面试.md) |
| Q3 | Lease 与 TTL？ | 租约自动过期删 key；KeepAlive 续期；Session 锁基础。 | 锁续期失败 |
| Q4 | Watch 机制？ | 长连接推送变更；revision 单调；compact 后需 resync。 | 背压 |
| Q5 | etcd 性能瓶颈？ | 磁盘 fsync、大 value、过多 Watch；SSD+batch。 | defrag |
| Q6 | ZooKeeper 数据模型？ | 层次 znode；持久/临时/顺序节点。 | ACL |
| Q7 | ZK 会话 Session？ | 客户端心跳维持；Session 超时临时节点删除触发回调。 | 惊群 |
| Q8 | ZK vs etcd 选型？ | ZK Java 生态老；etcd Go/cloud native；K8s 默认 etcd。 | CP 都强一致 |
| Q9 | Curator 框架？ | Java 封装 ZK 锁/Leader/Barrier；C++ 用原生或 braft。 | C++ 路径 |
| Q10 | etcd 事务 Txn？ | Compare-And-Swap 多 key 原子；MVCC revision。 | 乐观锁 |

## §5 braft / NuRaft C++ 工程

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | braft 是什么？ | 百度开源 C++ Raft 库，brpc 生态；NuRaft 是 eBay 的 C++11 Raft。 | 选型对比 |
| Q2 | braft 核心类？ | Node、LogEntry、Snapshot、Closure 回调；StateMachine 接口。 | 实现状态机 |
| Q3 | StateMachine on_apply？ | Leader commit 后按 index 顺序 apply 到业务；必须幂等。 | [63 幂等](63-JWT认证与接口幂等性实战.md) |
| Q4 | braft 快照流程？ | SaveSnapshot/on_snapshot_load；InstallSnapshot 追平落后节点。 | 大状态机 |
| Q5 | braft 与 brpc 集成？ | Raft 复制走 brpc channel；业务 RPC 与共识分离。 | [19 gRPC](19-gRPC与Protobuf工程化.md) |
| Q6 | NuRaft 特点？ | header-only 倾向、插件化日志存储、async snapshot。 | 新项目可评估 |
| Q7 | C++ 实现注意点？ | apply 线程模型、Closure 生命周期、日志磁盘顺序写。 | [08 并发](08-多线程与并发编程.md) |
| Q8 | Raft 单元测试？ | Mock 网络分区、延迟、丢包；Jepsen 混沌测试。 | 工程成熟度 |
| Q9 | 配置三节点最小？ | N=3 容忍 1 故障；生产 5 节点容忍 2。 | 跨 AZ 部署 |
| Q10 | Read Index / Lease Read？ | 避免 stale read：ReadIndex 或 Leader lease 本地读。 | Follower 读 |

## §6 CP 系统与 trade-off

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 什么是 CP 系统？ | 分区时牺牲可用保一致；写可能失败，读可能阻塞等同步。 | 与 AP 对比 |
| Q2 | Redis Cluster CP 吗？ | 默认 AP 倾向；RedLock 争议；主从异步复制可能丢。 | [52 Redis](52-Redis数据结构与缓存面试专章.md) |
| Q3 | MySQL 同步复制 CP？ | 半同步/组复制增强 CP；异步主从偏 AP。 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Q4 | 配置中心 CP 需求？ | 配置不能分叉；etcd/ZK 强一致；本地缓存+Watch 最终一致视图。 | 启动依赖 |
| Q5 | 分布式 ID CP？ | Snowflake 本地生成非共识；DB/Redis/etcd 序号需一致策略。 | 时钟回拨 |
| Q6 | Split-brain 案例？ | 网络分区两 Leader 同时写；fence+quorum 防双写。 | STONITH |
| Q7 | Fencing Token？ | 单调 token 写存储；旧 Leader  token 被拒绝。 | 共享存储 |
| Q8 | Raft 与 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)？ | Kafka ISR 非严格共识；offset 由 broker 管理；不同问题域。 | 顺序与复制 |
| Q9 | 面试如何讲 trade-off？ | 业务能否接受 stale read、写失败重试、延迟；量化 RPO/RTO。 | [58 模拟](58-模拟面试完整流程与压测数据模板.md) |
| Q10 | CAP 过时论？ | PACELC 延伸：分区时 C/A，正常时 Latency/Consistency；更细。 | Latency 维度 |

## §7 STAR 案例

### §7.1 STAR 案例 1

**S（情境）**：生产 etcd 集群频繁选主，K8s API 超时。

**T（任务）**：30 分钟内恢复控制面稳定。

**A（行动）**：查 slow disk、defrag、网络 jitter；调 election timeout；换 SSD；限制大 Watch。

**R（结果）**：选主频率降 95%，P99 API 延迟从 2s 到 200ms。

**连环追问**：
- 为何 random election timeout？
- Learner 节点作用？
- 如何验证 Raft 日志一致？

### §7.2 STAR 案例 2

**S（情境）**：自研 C++ 配置服务用 braft，Follower 读返回旧值被投诉。

**T（任务）**：提供线性一致读或明确 staleness SLA。

**A（行动）**：默认走 ReadIndex；热点 key 本地 cache+版本号；文档标注 max staleness 500ms。

**R（结果）**：误读工单归零；读 QPS 提升 40%。

**连环追问**：
- Lease Read 风险？
- apply 线程阻塞影响？
- 与 [52 Redis](52-Redis数据结构与缓存面试专章.md) 缓存穿透关系？
## §8 C++ 工程代码示例

```cpp
// braft StateMachine 伪代码（面试白板）
class ConfigSM : public braft::StateMachine {
public:
    void on_apply(braft::Iterator& iter) {
        for (; iter.valid(); iter.next()) {
            auto* done = iter.done();  // 用户 Closure
            std::string cmd(iter.data().to_string());
            apply_cmd(cmd);            // 必须幂等，见 63 章
            if (done) done->Run(Status::OK());
        }
    }
    void on_snapshot_save(braft::SnapshotWriter* writer, braft::Closure* done) {
        // 序列化状态到 snapshot
        done->Run(Status::OK());
    }
};
```

## §9 口述模板（2 分钟版）

### §9.1 CAP 30 秒

分布式必接受分区；我们在 **元数据/锁** 选 CP（etcd），**缓存/会话** 选 AP（Redis）；MySQL 半同步折中。结合 [51](51-MySQL原理与索引事务面试专章.md)[52](52-Redis数据结构与缓存面试专章.md) 讲存储层。

### §9.2 Raft 60 秒

三子问题：选举、复制、安全。Leader 写 log→多数派 ack→commit→apply 状态机。C++ 用 braft StateMachine；读用 ReadIndex。对比 Paxos 更易实现。

### §9.3 工程落地 30 秒

K8s 用 etcd；ZK 老系统；自研选 braft+brpc。注意快照、成员变更、混沌测试。下一章 [60](60-抓包与网络排障Wireshark实战.md) 用抓包验证 RPC 超时。
## §10 闭卷自测清单

- [ ] 能白板画 Raft 选举与日志复制时序图
- [ ] 能白板画 Raft 选举与日志复制时序图
- [ ] 能白板画 Raft 选举与日志复制时序图
- [ ] 能白板画 Raft 选举与日志复制时序图
- [ ] 能白板画 Raft 选举与日志复制时序图
- [ ] 能解释 etcd Lease 锁与 ZK 临时节点差异
- [ ] 能对比 braft 与 NuRaft 集成成本
- [ ] 能结合 [56 章](56-系统设计案例库RPC-KV与限流秒杀.md) 讲分布式 KV 元数据
- [ ] 能口述 PACELC 与 CAP 区别
- [ ] 能列举 CP 系统在分区时的失败模式
## §11 与 51～58 章交叉索引

| 本章考点 | 关联章节 | 说明 |
|----------|----------|------|
| 分布式 KV | [56 系统设计](56-系统设计案例库RPC-KV与限流秒杀.md) | 元数据与分片 |
| 消息顺序 | [57 Kafka](57-消息队列Kafka与中间件面试专题.md) | 非共识复制 |
| 事务一致 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) | 单机 ACID vs 分布式 |
| 缓存一致 | [52 Redis](52-Redis数据结构与缓存面试专章.md) | AP 与 RedLock |
| 模拟面试 | [58 模拟](58-模拟面试完整流程与压测数据模板.md) | 系统设计追问 |

---

**下一章**：[60-抓包与网络排障Wireshark实战.md](60-抓包与网络排障Wireshark实战.md)
### 附录 A.1 深度追问：Raft 场景 1

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.2 深度追问：Raft 场景 2

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.3 深度追问：Raft 场景 3

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.4 深度追问：Raft 场景 4

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.5 深度追问：Raft 场景 5

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.6 深度追问：Raft 场景 6

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.7 深度追问：Raft 场景 7

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.8 深度追问：Raft 场景 8

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.9 深度追问：Raft 场景 9

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.10 深度追问：Raft 场景 10

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.11 深度追问：Raft 场景 11

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.12 深度追问：Raft 场景 12

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.13 深度追问：Raft 场景 13

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.14 深度追问：Raft 场景 14

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.15 深度追问：Raft 场景 15

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.16 深度追问：Raft 场景 16

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.17 深度追问：Raft 场景 17

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.18 深度追问：Raft 场景 18

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.19 深度追问：Raft 场景 19

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.20 深度追问：Raft 场景 20

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.21 深度追问：Raft 场景 21

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.22 深度追问：Raft 场景 22

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.23 深度追问：Raft 场景 23

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.24 深度追问：Raft 场景 24

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.25 深度追问：Raft 场景 25

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.26 深度追问：Raft 场景 26

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.27 深度追问：Raft 场景 27

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.28 深度追问：Raft 场景 28

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.29 深度追问：Raft 场景 29

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.30 深度追问：Raft 场景 30

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.31 深度追问：Raft 场景 31

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.32 深度追问：Raft 场景 32

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.33 深度追问：Raft 场景 33

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.34 深度追问：Raft 场景 34

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.35 深度追问：Raft 场景 35

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.36 深度追问：Raft 场景 36

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.37 深度追问：Raft 场景 37

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.38 深度追问：Raft 场景 38

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.39 深度追问：Raft 场景 39

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。

### 附录 A.40 深度追问：Raft 场景 40

**面试官**：若 Follower 落后 Leader 1000 条日志，如何追平？

**答**：AppendEntries 连续复制；若 prevLogIndex 不匹配则递减 nextIndex；仍不行则 InstallSnapshot。工程上控制日志大小与 snapshot 频率。关联 [61 章](61-线上故障排查与性能诊断实战.md) 磁盘与网络排查。

**面试官**：与 [54 章 TCP](54-计算机网络TCP与HTTP面试深度专章.md) 超时有何关系？

**答**：Raft RPC 超时触发选举；需区分网络抖动与真宕机；配合 [60 章](60-抓包与网络排障Wireshark实战.md) 抓包看 RTT 与重传。


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

