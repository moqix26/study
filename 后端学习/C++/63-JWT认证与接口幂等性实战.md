# JWT 认证与接口幂等性实战

> **文件编码**：UTF-8。JWT 结构、Refresh Token、幂等 vs 分布式锁、Token 机制、唯一索引、状态机、Redis SetNX
> **交叉阅读**：[10 网络 HTTP](10-网络编程与简易HTTP服务.md) · [56 系统设计](56-系统设计案例库RPC-KV与限流秒杀.md) · [52 Redis](52-Redis数据结构与缓存面试专章.md) · [51 MySQL](51-MySQL原理与索引事务面试专章.md)

---

## 本章与前后章的关系

| 上一章 | 本章 | 下一章 |
|--------|------|--------|
| [62 K8s](62-Docker与Kubernetes入门面试.md) | **本章** | （系列 63 章末，回链 [58 模拟](58-模拟面试完整流程与压测数据模板.md)） |

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

本章 = **JWT 无状态认证 + Refresh 轮转 + 写接口幂等（Token/唯一索引/状态机/Redis SetNX）**；与 [62 K8s](62-Docker与Kubernetes入门面试.md) 网关部署、[56 秒杀](56-系统设计案例库RPC-KV与限流秒杀.md) 重复提交、[59 锁](59-分布式理论CAP-Raft与共识算法面试.md) 对比衔接。

### §0.2 你需要提前知道什么

| 状态 | 动作 |
|------|------|
| 只会用不会讲 | 每节 Q&A 限时 2min 口述 |
| C++ 后端岗 | 必串 [08 多线程](08-多线程与并发编程.md) [10 网络](10-网络编程与简易HTTP服务.md) [23 IO](23-IO多路复用与高性能Server.md) |
| 前置章节 | [62 K8s](62-Docker与Kubernetes入门面试.md) 网关与 Secret |
| 后续章节 | [58 模拟面试](58-模拟面试完整流程与压测数据模板.md) 综合演练 |

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

## §1 JWT 结构与验签

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | JWT 三部分？ | Header.Algorithm + Payload.Claims + Signature；Base64URL。 | 非加密仅签名 |
| Q2 | 常见 Claims？ | sub、exp、iat、iss、aud、jti。 | exp 必校验 |
| Q3 | HS256 vs RS256？ | HS 对称密钥共享难；RS 私钥签公钥验适合多服务。 | KMS 管钥 |
| Q4 | JWT 放哪？ | Authorization Bearer；或 HttpOnly Secure Cookie 防 XSS。 | CSRF 配合 |
| Q5 | 无状态优劣？ | 扩展性好；难即时吊销需黑名单/短 exp+Refresh。 | logout 设计 |
| Q6 | Refresh Token 设计？ | 长寿命存 HttpOnly Cookie/DB；轮换一次一用；泄露检测。 | reuse 检测 |
| Q7 | Access Token 寿命？ | 5～15min；减 stolen 窗口。 | 业务容忍 |
| Q8 | jti 用途？ | 唯一 ID 配合 Redis 黑名单防重放。 | 幂等也可复用思路 |
| Q9 | 网关验签？ | Nginx lua / 自研 C++ middleware；缓存 JWKS。 | [10 HTTP](10-网络编程与简易HTTP服务.md) |
| Q10 | 与 Session 对比？ | Session 服务端状态；JWT 客户端带 claims。 | [52 Redis session](52-Redis数据结构与缓存面试专章.md) |

## §2 Refresh Token 与安全

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | Refresh 轮转流程？ | login 发 access+refresh；refresh 端点验 refresh→发新对→旧 refresh 作废。 | 双 token |
| Q2 | reuse 攻击？ | 旧 refresh 再用说明可能泄露；吊销全家桶。 | 告警 |
| Q3 | 存储 Refresh？ | DB/Redis 哈希+device id；勿 localStorage。 | 绑定设备 |
| Q4 | OAuth2 关系？ | JWT 常作 access token 格式；OIDC 标准化 claims。 | 面试延伸 |
| Q5 | C++ 库？ | jwt-cpp、openssl 验 RS256；注意 clock skew。 | 异常处理 |
| Q6 | 密钥轮换？ | kid header 多公钥；平滑切换。 | 零 downtime |
| Q7 | mTLS + JWT？ | 双层；JWT 管用户身份。 | 零信任 |
| Q8 | 性能？ | 验签 CPU；缓存解析结果短 TTL。 | 边缘网关 |
| Q9 | 跨域？ | CORS + Cookie SameSite=None Secure。 | 前后端分离 |
| Q10 | 测试？ | 固定 clock；过期/篡改 signature case。 | [27 GTest](27-Google-Test与单元测试工程.md) |

## §3 幂等 vs 分布式锁

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 幂等定义？ | 同一业务请求执行多次效果等同一次。 | HTTP 方法 |
| Q2 | HTTP 哪些天然幂等？ | GET/PUT/DELETE 设计应幂等；POST 非幂等需机制。 | REST 语义 |
| Q3 | 幂等 vs 分布式锁？ | 幂等防重复结果；锁防并发互斥；支付常两者结合。 | [59 锁](59-分布式理论CAP-Raft与共识算法面试.md) |
| Q4 | Idempotency-Key 头？ | 客户端生成 UUID；服务端去重表。 | Stripe 模式 |
| Q5 | 唯一索引？ | INSERT 业务单号 UNIQUE；冲突则返回原结果。 | [51 MySQL](51-MySQL原理与索引事务面试专章.md) |
| Q6 | 状态机？ | 订单 created→paid→shipped；非法迁移拒绝。 | 乐观锁 version |
| Q7 | Redis SetNX？ | SET key token NX EX 300 占坑；失败说明处理中或已完成。 | [52 Redis](52-Redis数据结构与缓存面试专章.md) |
| Q8 | Token 机制？ | 先占坑返回 token；完成后 mark done；重试用同 token 查结果。 | 异步任务 |
| Q9 | 去重表结构？ | idempotency_key PK, user_id, response_body, status, expire_at。 | 清理策略 |
| Q10 | 与 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)？ | 消费者幂等；at-least-once+业务去重。 | offset 提交 |

## §4 Token / 唯一索引 / 状态机 / SetNX

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | 支付重复点击？ | 前端 debounce+后端 Idempotency-Key+DB 唯一约束。 | 用户感知 |
| Q2 | 秒杀下单？ | [56 章](56-系统设计案例库RPC-KV与限流秒杀.md) 预扣+幂等单号。 | 超卖 vs 重复 |
| Q3 | 分布式事务幂等？ | TCC Try 幂等；Confirm/Cancel 也幂等。 | [59 TCC](59-分布式理论CAP-Raft与共识算法面试.md) |
| Q4 | C++ 实现骨架？ | Middleware 读 key→查 Redis/DB→短路返回 cache response。 | [21 模式](21-设计模式与Infra工程实践.md) |
| Q5 | 并发重复 POST？ | 两请求同时到：唯一索引一个成功一个冲突读结果。 | 事务隔离 |
| Q6 | 过期 key 清理？ | TTL+定时扫；防表膨胀。 | 归档 |
| Q7 | 返回码约定？ | 冲突 200 带原 id vs 409；需 API 文档统一。 | 客户端逻辑 |
| Q8 | 日志审计？ | 记录 key+首次/重复；风控。 | [32 日志](32-fmt-spdlog与可观测性工程.md) |
| Q9 | braft apply 幂等？ | 重复 index apply 跳过或 version 校验。 | [59 braft](59-分布式理论CAP-Raft与共识算法面试.md) |
| Q10 | 压测验证？ | wrk 同 key 并发；断言 DB 单行。 | [58 压测](58-模拟面试完整流程与压测数据模板.md) |

## §5 STAR 案例

### §5.1 STAR 案例 1

**S（情境）**：支付接口重复扣款客诉。

**T（任务）**：零重复扣款。

**A（行动）**：引入 Idempotency-Key+UNIQUE(order_no)+状态机；Redis SetNX 挡并发；返回原单号。

**R（结果）**：投诉归零；接口 P99 不变。

**连环追问**：
- 与锁区别？
- 唯一索引失败处理？
- MySQL 死锁？

### §5.2 STAR 案例 2

**S（情境）**：JWT 泄露疑似。

**T（任务）**：吊销并轮转密钥。

**A（行动）**：缩短 exp；Refresh reuse 检测；Redis jti 黑名单；强制 re-login。

**R（结果）**：2h 内全量换 token。

**连环追问**：
- HS256 风险？
- K8s Secret 管理？
- 见 [62 章](62-Docker与Kubernetes入门面试.md)
## §6 C++ 工程示例

```cpp
// C++ 伪代码：Redis SetNX + MySQL 唯一索引 双保险
bool handle_pay(const PayReq& req, PayResp& out) {
    const std::string key = "idem:" + req.idempotency_key;
    if (!redis.set(key, "processing", NX, EX(300)))
        return load_cached_response(req.idempotency_key, out);

    try {
        db.exec("INSERT INTO orders(idempotency_key, ...) VALUES (?, ...)", req.idempotency_key);
        auto result = do_pay(req);
        save_response(key, result);
        out = result;
        return true;
    } catch (const UniqueViolation&) {
        return load_cached_response(req.idempotency_key, out);
    }
}
```

## §9 口述模板（2 分钟版）

### §9.1 JWT 45 秒

Header.Payload.Signature；RS256 公钥验；短 access+Refresh 轮转；jti 黑名单。

### §9.2 幂等 60 秒

POST 用 Idempotency-Key；DB UNIQUE+状态机；Redis SetNX 挡并发；与分布式锁分工。

### §9.3 串联 30 秒

部署 [62 K8s]；排障 [61]；共识 [59]；模拟 [58] 收官。
## §10 闭卷自测清单

- [ ] 能画 JWT 验签流程
- [ ] 能画 JWT 验签流程
- [ ] 能写 SetNX 伪代码
- [ ] 能设计去重表
- [ ] 能对比 Session
- [ ] 能串 51-62 章节
## §11 全系列 51→63 总结

| 章 | 主题 | 与本章关系 |
|----|------|------------|
| 51 | MySQL | 唯一索引幂等 |
| 52 | Redis | SetNX/Session |
| 56 | 系统设计 | 秒杀幂等 |
| 59 | 分布式 | 锁 vs 幂等 |
| 62 | K8s | 网关部署 |
| 58 | 模拟 | 综合面试 |

**系列回链**：[58-模拟面试完整流程与压测数据模板.md](58-模拟面试完整流程与压测数据模板.md)
### 附录 E.1 幂等场景题 1

**场景 1**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.2 幂等场景题 2

**场景 2**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.3 幂等场景题 3

**场景 3**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.4 幂等场景题 4

**场景 4**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.5 幂等场景题 5

**场景 5**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.6 幂等场景题 6

**场景 6**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.7 幂等场景题 7

**场景 7**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.8 幂等场景题 8

**场景 8**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.9 幂等场景题 9

**场景 9**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.10 幂等场景题 10

**场景 10**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.11 幂等场景题 11

**场景 11**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.12 幂等场景题 12

**场景 12**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.13 幂等场景题 13

**场景 13**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.14 幂等场景题 14

**场景 14**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.15 幂等场景题 15

**场景 15**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.16 幂等场景题 16

**场景 16**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.17 幂等场景题 17

**场景 17**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.18 幂等场景题 18

**场景 18**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.19 幂等场景题 19

**场景 19**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.20 幂等场景题 20

**场景 20**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.21 幂等场景题 21

**场景 21**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.22 幂等场景题 22

**场景 22**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.23 幂等场景题 23

**场景 23**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.24 幂等场景题 24

**场景 24**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.25 幂等场景题 25

**场景 25**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.26 幂等场景题 26

**场景 26**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.27 幂等场景题 27

**场景 27**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.28 幂等场景题 28

**场景 28**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.29 幂等场景题 29

**场景 29**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.30 幂等场景题 30

**场景 30**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.31 幂等场景题 31

**场景 31**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.32 幂等场景题 32

**场景 32**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.33 幂等场景题 33

**场景 33**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.34 幂等场景题 34

**场景 34**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.35 幂等场景题 35

**场景 35**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.36 幂等场景题 36

**场景 36**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.37 幂等场景题 37

**场景 37**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.38 幂等场景题 38

**场景 38**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.39 幂等场景题 39

**场景 39**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.40 幂等场景题 40

**场景 40**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.41 幂等场景题 41

**场景 41**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.42 幂等场景题 42

**场景 42**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.43 幂等场景题 43

**场景 43**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.44 幂等场景题 44

**场景 44**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.45 幂等场景题 45

**场景 45**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.46 幂等场景题 46

**场景 46**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。

### 附录 E.47 幂等场景题 47

**场景 47**：客户端超时重试 POST /api/order，Body 相同。

**期望**：仅创建一单；重试返回相同 order_id。

**方案**：Idempotency-Key header + [51](51-MySQL原理与索引事务面试专章.md) UNIQUE + [52](52-Redis数据结构与缓存面试专章.md) SetNX。


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

