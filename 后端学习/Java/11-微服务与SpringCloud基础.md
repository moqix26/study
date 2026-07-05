# 微服务与 Spring Cloud 基础

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0、FAQ、闭卷自测、费曼检验 -->

## 0. 读前导读（零基础也能跟上）

> **读者假设**：10 章单体商城 MVP 做透了——本章建立**架构视野**：什么时候该开分店、分店之间怎么打电话、总部门怎么接客。

### 0.1 用一句话弄懂本章

**一句话**：**微服务**像把一家大店拆成多家**分店**（用户店、商品店、订单店），各店独立装修升级；**Spring Cloud** 是连锁总部的**规章制度**（注册中心、网关、远程调用、熔断）。

**生活类比——微服务 = 分店连锁**：

| 概念 | 组件 | 生活类比 |
|------|------|----------|
| **单体** | 一个 demo.jar | **一家总店**全包：前台、仓库、收银在一起 |
| **微服务** | user/product/order-service | **分店**：用户店只管会员，商品店只管货架 |
| **注册中心 Nacos** | 服务发现 | **总部花名册**——各店开业登记电话，打电话先查花名册 |
| **Gateway 网关** | 统一入口 | **集团总前台**——客人只进总门，前台分流到各分店 |
| **OpenFeign** | 声明式 HTTP 调用 | **分店之间内线**——订单店打电话给商品店查库存 |
| **Sentinel 熔断** | 限流降级 | 某分店装修暂停营业，**别拖垮整个集团** |
| **09 章 Nginx** | 反代 | 单店时的前台；Gateway 是**多店时代的加强版总前台** |

**为什么重要**：面试常问「单体和微服务区别」「为什么拆」——本章让你**会说、会画**，而不是只会背组件名。

**本章用到的地方**：§3 拆分、§7 概念、§28 组件详解、§31 FAQ。

---

### 0.2 你需要提前知道什么（真不会就先跳到哪一章）

| 你现在的水平 | 建议动作 |
|--------------|----------|
| 单体 demo 还没做透 | **先回 10 章**；没总店别急着开分店 |
| HTTP/REST 不熟 | 回 [04 Spring Boot](./04-SpringBoot核心开发.md) |
| 只会背 Nacos/Gateway 名词 | 本章重概念与链路图，动手拆可后补 |
| 想了解分布式事务 | 读 §29，知道「最终一致性+MQ」即可 |
| 准备面试架构题 | §23 决策表 + §48 自测 + §49 费曼 |

**最低门槛**：10 章能画单体架构图；知道 HTTP 调用和本地方法调用的区别。

---

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

- [ ] 说出单体 vs 微服务优缺点各 3 条
- [ ] 用**分店**类比解释为什么要注册中心
- [ ] 画出：用户 → Gateway → order → Feign → product/user
- [ ] 说出 Gateway、Feign、Nacos 各解决一个问题
- [ ] 知道拆分原则：**按业务域**，不是按表拆
- [ ] 了解熔断：下游挂了不能拖死上游
- [ ] 了解分布式事务：互联网多用最终一致性 + MQ
- [ ] 能读 §28 Gateway 路由 YAML 和 Feign 接口
- [ ] 知道何时**不该**拆微服务（小团队 MVP）
- [ ] 闭卷自测 10 题正确 ≥ 8 题

---

### 0.4 建议学习时长与节奏

| 阶段 | 建议时间 | 做什么 |
|------|----------|--------|
| §0 + §2～§5 单体/微服务 | 2 小时 | 决策表、优缺点 |
| §7～§11 基础设施概念 | 2 小时 | 注册/网关/配置/链路 |
| §28 组件详解 | 2 小时 | YAML + Feign + Sentinel 扫读 |
| §29 分布式事务 | 1 小时 | 最终一致性为主 |
| 画链路图 + 自测 | 1 小时 | 能默画 sequence 图 |

---

### 0.5 学完本章你能做什么（可验证的具体动作）

1. **5 分钟**白板画出 § 开头 Mermaid 等价图并讲解每一跳。
2. **口述**：把 10 章单体商城拆成 user/product/order 三服务，各管哪些表、哪些接口。
3. **解释** Feign 调用与单体里 `productService.getById()` 的差异（网络、超时、失败）。
4. **答题**：「你们为什么还不拆？」——小团队、MVP、运维成本（§23 表）。
5. **闭卷自测** ≥ 8/10。

---

### 0.6 手把手总览：从单体到「概念上的分店」

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 列出 10 章 demo 模块与表 | 用户/商品/订单边界清晰 | 按业务域而非 tech 层拆 |
| 2 | 画单体架构图 | 一个 jar 连三个中间件 | 对照 10 章 Mermaid |
| 3 | 画拆分后图 + Gateway | 三个服务 + 总前台 | 对照本章 sequence 图 |
| 4 | 标出 2 次 Feign 调用 | 下单调 user、调 product | 路径与服务名一致 |
| 5 | 口述一次失败场景 | 商品店挂了→熔断/降级 | §20、§28.4 |
| 6 | 费曼 3 分钟 | 朋友听懂「分店+总前台」 | §49 提纲 |

---

## 本章与上一章的关系

10 章你把单体商城 MVP 做透了——用户、商品、订单、缓存、MQ 都在一个 jar 里。当用户量上来、团队变大，一个 jar 会越来越难维护：改订单逻辑要重新部署整个系统，商品模块想扩容得把整个应用都扩。

这一章不急着动手拆，先建立**微服务概念**：为什么要拆、Gateway/Feign/Nacos 各干什么。11 章是「架构视野」，12 章是「高并发与分布式 deeper」——单体扎实后再看这些，才不会只会背组件名。

### 微服务调用链路图

```mermaid
sequenceDiagram
    participant User as 用户
    participant GW as Gateway 网关
    participant Order as order-service
    participant Product as product-service
    participant UserSvc as user-service
    participant DB as MySQL
    participant MQ as RabbitMQ

    User->>GW: POST /api/orders
    GW->>GW: JWT 鉴权
    GW->>Order: 路由转发
    Order->>UserSvc: Feign 查用户
    UserSvc-->>Order: 用户信息
    Order->>Product: Feign 查库存
    Product-->>Order: 库存 OK
    Order->>DB: 写订单 + 扣库存
    Order->>MQ: 发异步通知
    Order-->>GW: 订单号
    GW-->>User: JSON 响应
```

---

## 1. 为什么要学这一章

当你单体项目做得越来越大时，会逐渐遇到这些问题：

- 一个项目太庞大
- 模块耦合太重
- 发布一次影响全部
- 不同模块扩容需求不同

这时就会引出微服务架构。

## 2. 什么是单体架构

单体架构就是：

- 所有模块都在一个项目里
- 一起开发
- 一起部署

优点：

- 开发简单
- 部署直接
- 初期成本低

缺点：

- 项目变大后维护压力上升
- 模块边界不清时容易互相影响

## 3. 什么是微服务架构

**微服务（Microservices）**：把大系统按**业务边界**拆成多个小服务，各自独立开发、部署、扩缩容。

**生活类比——微服务 = 分店连锁**：

- **用户服务** = 会员分店（注册、登录、资料）
- **商品服务** = 商品分店（上架、库存、搜索）
- **订单服务** = 订单分店（下单、支付状态）
- 各店有**自己的仓库**（独立数据库），不能随意翻别人仓库；要货就**打电话**（Feign/HTTP）

**为什么重要**：10 章单体是「总店全包」——简单好管；用户量/团队变大后，改订单逻辑不必重装整个总店。

**本章用到的地方**：§25 调用链、§28 Feign、§29 跨库事务。

微服务可以理解为：

- 把一个大系统拆成多个小服务
- 每个服务独立开发、独立部署

例如商城系统可以拆成：

- 用户服务
- 商品服务
- 订单服务
- 支付服务

## 4. 微服务的优点

- 服务边界更清晰
- 可以独立扩容
- 团队协作更灵活

## 5. 微服务的缺点

微服务不是没有代价的，它会带来更多复杂度：

- 服务调用变复杂
- 需要注册发现
- 需要网关
- 需要配置中心
- 需要链路追踪
- 需要更强的运维能力

所以你要理解：

- 微服务不是“更高级就一定更好”
- 它是为了解决单体在特定阶段的问题

## 6. Spring Cloud 是干什么的

Spring Cloud 提供的是一整套微服务常见问题的解决方案。

你可以把它理解成：

- 微服务开发的一组基础设施工具

## 7. 你应该先掌握哪些概念

### 7.1 服务注册与发现

为什么需要：

- 服务拆多了，调用方要知道被调用方在哪

### 7.2 配置中心

为什么需要：

- 多个服务的配置不适合到处散落

### 7.3 网关

为什么需要：

- 统一入口
- 路由转发
- 鉴权控制

### 7.4 负载均衡

为什么需要：

- 同一个服务可能部署多个实例

### 7.5 熔断和降级

为什么需要：

- 某个下游服务异常时，不能把整个系统拖死

## 8. 微服务调用的基本认知

单体里通常是方法调用。

微服务里更多是：

- HTTP 调用
- RPC 调用
- 消息队列异步调用

## 9. 网关的角色

**Gateway（API 网关）**：微服务体系的**统一入口**，负责路由、鉴权、限流等。

**生活类比**：10 章 **Nginx 前台分流**管单店；Gateway 是**连锁集团总前台**——客人只认一个域名，总前台把 `/api/users` 分到用户分店、`/api/orders` 分到订单分店，还能统一验会员证（JWT）。

**为什么重要**：没有总前台，浏览器要记多个端口；鉴权逻辑也会散落各店。

**本章用到的地方**：§28.2 路由 YAML。

网关常做的事：

- 路由转发
- 统一鉴权
- 限流
- 日志
- 跨域处理

你可以把它理解成微服务系统的“统一门口”。

## 10. 配置中心的角色

多服务场景下，数据库地址、Redis 地址、MQ 地址等配置很多。

配置中心的作用是：

- 集中管理配置
- 动态更新配置

## 11. 链路追踪基础认知

服务拆多之后，一个请求可能要经过很多服务。

这时问题就来了：

- 请求到底卡在哪个服务
- 哪个环节最慢

所以要有链路追踪能力。

## 12. 分布式事务基础认知

单体系统一个数据库事务还能比较直接地处理。

微服务里如果跨多个服务、多个库，就会遇到分布式事务问题。

你现在先知道这是个难点即可，不用一开始深挖。

## 13. 初学微服务的正确顺序

建议顺序：

1. 先把单体项目做好
2. 再理解为什么要拆服务
3. 再学微服务基础设施

不要一上来就学一堆 Spring Cloud 组件名字。

## 14. 这一章学到什么程度够用

对于当前阶段，你至少要做到：

- 知道单体和微服务的区别
- 知道微服务为什么会引入注册中心、网关、配置中心
- 知道微服务会带来什么复杂度

这就已经足够支撑初级面试中的基础问答。

## 15. 服务注册中心再细一点

服务注册中心要解决的问题是：

- 服务地址不是写死的
- 实例可能会增减

常见思路：

1. 服务启动后把自己注册进去
2. 调用方去注册中心发现可用实例

你后面会看到的名字可能包括：

- Eureka
- Nacos

## 16. 服务调用基础认知

微服务之间不会像单体里那样直接方法调用。

常见方式：

- HTTP 调用
- OpenFeign 等声明式调用

你现在先知道：

- 服务调用会引入网络开销和失败风险

## 17. 网关为什么重要

如果没有网关，外部请求可能直接面对多个服务，管理会很乱。

网关的价值：

- 统一入口
- 路由转发
- 权限控制
- 限流
- 日志

## 18. 配置中心为什么重要

微服务多了以后，这些配置会越来越多：

- 数据库地址
- Redis 地址
- MQ 地址
- 第三方密钥

如果每个服务自己管，维护成本会很高。

## 19. OpenFeign 基础认知

这是微服务中很常见的调用工具。

你可以先简单理解：

- 用接口方式声明调用另一个服务

它让服务调用写法更清晰。

## 20. 熔断和降级再细一点

为什么需要熔断：

- 下游服务挂了，不能把上游一起拖死

为什么需要降级：

- 系统压力太大时，优先保证核心功能可用

## 21. 微服务拆分不是越细越好

这是非常重要的认知。

拆得太细会带来：

- 调用链更长
- 运维更复杂
- 故障面更多

所以拆分要有业务边界，而不是为了“高级”而拆。

## 22. 微服务这一章的高频知识点总清单

建议整理这些点：

- 单体架构
- 微服务架构
- 注册中心
- 服务发现
- 网关
- 配置中心
- 服务调用
- 熔断
- 降级
- 负载均衡
- 链路追踪
- 分布式事务基础

---

## 23. 单体 vs 微服务决策表

| 维度 | 单体 | 微服务 |
|------|------|--------|
| 团队规模 | 小团队 | 多团队并行 |
| 部署 | 一个 jar | 多个服务独立发布 |
| 技术栈 | 统一 | 可按服务选型 |
| 复杂度 | 低 | 高（治理、链路） |
| 适用 | 创业、MVP、学习 | 大流量、大组织 |

**学习建议**：第一个项目用 **单体 Spring Boot** 做完；11 篇理解概念即可。

---

## 24. Spring Cloud 常见组件速查

| 组件 | 作用 | 备注 |
|------|------|------|
| Nacos / Eureka | 注册与发现 | 服务名调用 |
| Gateway / Zuul | API 网关 | 路由、鉴权、限流 |
| OpenFeign | 声明式 HTTP 调用 | 像调本地方法 |
| Sentinel | 限流熔断 | 阿里系常用 |
| Config / Nacos Config | 配置中心 | 动态刷新 |
| Sleuth + Zipkin | 链路追踪 | 查慢调用 |

---

## 25. 一次微服务调用链示例

```text
用户 → Gateway（鉴权、路由）
     → order-service（下单）
         → Feign 调 product-service（查库存）
         → Feign 调 user-service（查用户）
         → 写 order-db
         → 发 RabbitMQ
```

你要能说出：**每一跳都可能超时、失败，需要熔断和超时配置**。

---

## 26. 学完标准

- 能画单体与微服务架构对比图
- 说出注册中心、网关、Feign 各解决什么问题
- 知道拆分原则：按业务域，不是按表
- 了解分布式事务 Seata/TCC 「听说过」级别

---

## 28. Spring Cloud 核心组件详解

| 组件 | 作用 | Spring Cloud 实现 |
|------|------|-------------------|
| 服务注册/发现 | 服务上线后自动注册，调用者自动发现 | Nacos / Eureka |
| 配置中心 | 配置统一管理、动态刷新 | Nacos Config |
| 网关 | 统一入口、鉴权、路由、限流 | Spring Cloud Gateway |
| 远程调用 | 服务间 HTTP 调用像本地方法 | OpenFeign |
| 负载均衡 | 调用者侧负载均衡 | LoadBalancer |
| 熔断降级 | 下游故障时快速失败 | Sentinel / Resilience4j |
| 链路追踪 | 追踪请求在各服务间的路径 | Micrometer + Zipkin/SkyWalking |

### 28.1 服务注册中心工作原理

```
启动时：order-service → 注册到 Nacos（"我叫 order-service，IP 192.168.1.10:8080"）
心跳：每 5 秒向 Nacos 发送心跳包
发现：user-service 调用时 → 从 Nacos 拿到 order-service 的实例列表
下线：order-service 停止 → Nacos 30 秒没收到心跳 → 踢掉实例
```

### 28.2 Gateway 网关路由配置

```yaml
spring:
  cloud:
    gateway:
      routes:
        - id: user-service
          uri: lb://user-service   # lb = 负载均衡，自动从注册中心发现
          predicates:
            - Path=/api/users/**
          filters:
            - StripPrefix=0
        - id: order-service
          uri: lb://order-service
          predicates:
            - Path=/api/orders/**
```

### 28.3 OpenFeign 声明式调用

```java
@FeignClient(name = "product-service")  // 通过注册中心自动发现
public interface ProductClient {

    @GetMapping("/api/products/{id}")
    Result<ProductVO> getProduct(@PathVariable Long id);

    @PutMapping("/api/products/{id}/stock")
    Result<Void> deductStock(@PathVariable Long id, @RequestBody StockDTO dto);
}
```

```java
@Service
public class OrderService {

    private final ProductClient productClient;  // 像调本地方法一样

    public void createOrder(CreateOrderDTO dto) {
        Result<ProductVO> result = productClient.getProduct(dto.getProductId());
        // ...
    }
}
```

### 28.4 Sentinel 熔断降级

```java
// 方式一：注解
@SentinelResource(value = "createOrder",
                  fallback = "createOrderFallback",
                  blockHandler = "createOrderBlockHandler")
public Result<Long> createOrder(CreateOrderDTO dto) {
    // 业务逻辑
}

// 降级方法（业务异常时走）
public Result<Long> createOrderFallback(CreateOrderDTO dto, Throwable e) {
    return Result.fail("下单服务暂不可用，请稍后重试");
}

// 限流方法（被 Sentinel 拦截时走）
public Result<Long> createOrderBlockHandler(CreateOrderDTO dto, BlockException e) {
    return Result.fail("系统繁忙，请稍后重试");
}
```

```yaml
# Sentinel 控制台规则示例（也可在控制台界面配置）
# 资源: createOrder
# 流控规则: QPS=100 → 超过后快速失败
# 降级规则: 慢调用比例 > 50% 且 RT > 200ms → 熔断 10 秒
# 热点规则: 参数 productId=0 → QPS 限制 10
```

---

## 29. 分布式事务概念

微服务环境下，一个业务可能跨多个数据库：

```
下单流程：
order-service 写订单库 → product-service 扣库存库 → account-service 扣款库

问题：扣库存成功了，扣款失败了—— 跨服务的"要么全成功要么全失败"怎么做？
```

| 方案 | 原理 | 适用场景 |
|------|------|----------|
| 2PC（XA） | 两阶段提交，协调者决定提交还是回滚 | 强一致但性能差，基本不用 |
| TCC | Try-Confirm-Cancel：预留资源 → 确认 → 取消 | 金融类强一致场景 |
| 最终一致性 | 本地事务 + MQ 异步补偿 | **互联网最常用** |
| Seata AT | 自动回滚 undo_log | 不想写补偿逻辑时的折中方案 |

```java
// 最终一致性 + MQ（最常用）
@Transactional
public void createOrder(CreateOrderDTO dto) {
    // 1. 本地事务：写订单
    orderMapper.insert(order);

    // 2. 发送 MQ 消息（扣库存、发通知等异步处理）
    rabbitTemplate.convertAndSend("order.created", order);

    // 如果 MQ 发送失败 → 事务回滚，订单也没创建
    // 如果 MQ 发送成功但消费者失败 → MQ 重试 / 死信补偿
}
```

---

## 30. 学完标准（扩充版）

- [ ] 能画出单体架构 vs 微服务架构对比图
- [ ] 说出注册中心（Nacos）、网关（Gateway）、远程调用（Feign）各解决什么问题
- [ ] 理解微服务拆分原则：**按业务域拆，不是按表拆**
- [ ] 能写一个 Feign 接口，通过服务名调用另一个服务
- [ ] 知道熔断降级（Sentinel）的基本用法：流控 + 降级 + 热点
- [ ] 了解分布式事务：最终一致性 + MQ 是互联网主流

---

## 31. FAQ

### Q1：必须学 Spring Cloud Alibaba 吗？

**A**：国内 Nacos + Sentinel 组合很普遍；原理和 Netflix 系基本相通。面试能说清职责比纠结版本更重要。

### Q2：和 Dubbo 区别？

**A**：Feign 是 HTTP REST（文本协议，调试方便）；Dubbo 是 RPC（二进制协议，性能更高）。初学从 REST 开始，Dubbo 工作中遇到再学。

### Q3：微服务初学要不要搭完整环境？

**A**：不用。先在脑子里把单体拆成几个**分店**的架构图画清楚，理解组件职责。单体项目做透再动手拆。

### Q4：怎么判断要不要拆微服务？

**A**：团队小（<10人）、业务不复杂 → 单体做透。拆分有成本——分布式事务、网络延迟、运维复杂度都会增加。

### Q5：注册中心 Eureka 和 Nacos 选哪个？

**A**：新项目国内多用 **Nacos**（注册+配置一体）；老项目可能还有 Eureka。初学记 Nacos 即可。

### Q6：Gateway 和 09 章 Nginx 什么关系？

**A**：Nginx 是通用 Web/反代服务器；**Gateway** 是 Spring 生态里的 API 网关，和服务发现、负载均衡集成更深。单店用 Nginx，多店常用 Gateway 当**总前台**。

### Q7：Feign 调用失败常见原因？

**A**：服务名错、路径错、下游未注册、超时太短、网络不通——见 §27.1 排查表。

### Q8：按数据库表拆服务行吗？

**A**：**不行**。应按**业务域**拆（订单域、商品域）；按表拆会导致一个用户故事跨太多店。

### Q9：分布式事务一定要 Seata 吗？

**A**：互联网多数用 **最终一致性 + MQ**（§29）；Seata AT 是折中。初级「听说过 TCC/Seata」即可。

### Q10：熔断和降级一句话区别？

**A**：**熔断**：下游故障多了，暂时别打了（断路器打开）。**降级**：返回兜底结果（如「请稍后再试」），保核心体验。

### Q11：链路追踪解决什么问题？

**A**：一个请求经 Gateway→order→product，慢在哪一段？Sleuth/Zipkin/SkyWalking 给每段打 traceId。

### Q12：OpenFeign 和 RestTemplate 区别？

**A**：Feign **声明式**——写接口像调本地方法；RestTemplate 手写 URL。微服务里 Feign 更常见。

---

<!-- 修改说明: 新增常见报错与排查 + 下一章预告 -->

## 27.1 微服务入门常见坑（排查表）

| 现象 | 可能原因 | 解决方案 |
|------|---------|---------|
| Feign 调用 404 | 服务名或路径与 Provider 不一致 | 核对 `@FeignClient(name=...)` 和 Controller 路径 |
| 注册中心看不到服务 | 未启动 Nacos/Eureka 或 namespace 错 | 看注册中心控制台；检查 `spring.cloud.nacos.discovery` |
| Gateway 路由不通 | `uri` 或 `Path` _predicate 配错 | 看 Gateway 日志；用 `lb://service-name` |
| 调用超时 | 默认超时太短或下游慢 | 配置 `feign.client.config.default.readTimeout` |
| 循环依赖 | A 调 B、B 又调 A | 重新划边界；改 MQ 异步解耦 |

---

## 48. 闭卷自测

> 先遮住答案，逐题口述或默写。

### 概念题（6 道）

1. 用**分店**类比说明单体和微服务各是什么组织形态。
2. 注册中心解决什么问题？服务下线后调用方怎么知道？
3. Gateway 和 09 章 Nginx **前台分流**的相同点与不同点（各 1 条）。
4. OpenFeign 与单体里直接调 Service 方法比，多出了哪些风险？
5. 熔断和降级分别解决什么？
6. 为什么说「按表拆服务」往往是错的？

### 动手题（2 道）

7. 写 Gateway 路由：`/api/orders/**` → `lb://order-service`（YAML 片段）。
8. 写 Feign 接口：`ProductClient.getProduct(Long id)` 调 `product-service`。

### 综合题（2 道）

9. 画出下单请求经 Gateway、order-service、product-service、user-service 的调用顺序（文字或简图）。
10. 10 人团队 MVP 阶段，你会选单体还是微服务？用 §23 决策表说 3 条理由。

### 自测参考答案

1. 单体=一家总店全包；微服务=用户/商品/订单等分店，独立装修升级。
2. 实例地址动态变化；启动注册、心跳保活、超时剔除，调用方查花名册拿可用实例。
3. 相同：统一入口、路由转发。不同：Gateway 集成服务发现/负载均衡，面向微服务治理。
4. 网络超时、部分失败、序列化、需要熔断降级；单体方法调用无这些网络因素。
5. 熔断防下游故障扩散；降级在压力大或失败时返回兜底，保核心功能。
6. 业务故事跨表——应按域拆，否则一个功能要打多家店，复杂度高。
7. 见 §28.2：`Path=/api/orders/**`, `uri: lb://order-service`。
8. `@FeignClient(name="product-service")` + `@GetMapping("/api/products/{id}")`。
9. User→GW→order→Feign user→Feign product→写 order-db→MQ（与章首 sequence 图一致）。
10. 选单体：复杂度低、部署一个 jar、无分布式事务、团队小并行成本低；等流量/组织再拆。

---

## 49. 费曼检验

**任务**：请在不看资料的情况下，用 **3 分钟** 向朋友解释「什么是微服务，为什么不要一上来就拆」。

**对照提纲**：

1. **总店 vs 分店**：10 章一个 jar 是总店；生意大了把会员、商品、订单分成**分店**各自扩。
2. **总前台 + 花名册**：Gateway 接客，Nacos 记各店电话；Feign 是分店之间打电话。
3. **代价**：每次打电话可能占线、说不同步——所以要熔断、分布式事务、链路追踪。
4. **顺序**：先把总店（单体 MVP）做扎实；小团队别为了「高级」硬拆。

若朋友能说出「按业务拆、有统一入口、拆分有成本」，本章核心已掌握。

---

## 50. 本章与前后章节衔接速查

| 本章学会 | 10 章关系 | 12 章预告 |
|----------|-----------|-----------|
| 单体 vs 微服务 | 10 章是单体 MVP 目标 | 高并发下的限流熔断深入 |
| Gateway | 09 Nginx 的单店版前台 | 更大流量治理 |
| Feign 远程调用 | 10 章本地 Service 调用 | 跨服务一致性 |
| 最终一致性+MQ | 10 章下单发 MQ | 分布式场景放大 |

**动手验收清单**：

- [ ] 默画 sequence 图（章首）
- [ ] 口述 10 章单体如何拆三服务
- [ ] FAQ 任抽 3 题能答
- [ ] 闭卷自测 ≥ 8/10

---

## 51.1 Gateway 路由 YAML 逐行读（§28.2 延伸）

| 行/字段 | 含义 | 改错会怎样 |
|---------|------|------------|
| `id: order-service` | 路由规则名，唯一即可 | 重复 id 可能覆盖或启动失败 |
| `uri: lb://order-service` | **lb**=负载均衡，从 Nacos 找 order 实例 | 写死 IP 则失去扩缩容；服务名错则 503 |
| `Path=/api/orders/**` | 路径匹配 predicate | 少写 `**` 可能子路径不匹配 |
| `StripPrefix=0` | 转发时是否剥前缀 | 与下游 Controller 路径要对齐 |

### 51.2 10 章单体 → 三「分店」拆分对照

| 单体模块（10 章 jar） | 拆分后服务 | 独立库（示意） | 对外接口示例 |
|----------------------|------------|----------------|--------------|
| UserController + UserService | user-service | user_db | `/api/users`, `/api/login` |
| ProductController + ProductService | product-service | product_db | `/api/products`, 扣库存内部 API |
| OrderController + OrderService | order-service | order_db | `/api/orders`；Feign 调 user/product |

**注意**：拆库后「下单扣库存」跨两个库，不能再靠单个 `@Transactional`——需 §29 最终一致性或 Seata（了解）。

### 51.3 常见学习弯路

| 弯路 | 表现 | 纠正 |
|------|------|------|
| 单体没做透就拆 | 三个空壳服务 | 先 10 章 MVP 跑通 |
| 按 Controller 层拆 | user-controller-service | 按**业务域**拆 |
| 忽略 Feign 超时 | 偶发 504 | 配置 readTimeout/connectTimeout |
| 网关鉴权与业务重复 | 两套 JWT 逻辑 | 网关验 token，下游信任或再验 |
| 背组件不会画图 | 面试卡壳 | 章首 sequence 默画 3 遍 |

---

## 下一章预告

11 章你知道微服务「是什么、为什么」了——但面试官还会问：「QPS 10 万怎么办？」「缓存和 DB 不一致怎么权衡？」

下一章（12 高并发与分布式系统基础）讲限流、熔断、CAP、秒杀思路——把 07 Redis、08 MQ 的知识放到更大架构里理解。

---

*下一章：12 高并发与分布式系统基础*
