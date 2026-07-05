# RabbitMQ 与消息队列实战

<!-- 修改说明: 新增本章与上一章的关系 -->

## 本章与上一章的关系

07 章你用 Redis 解决了「读快」——商品详情走缓存，数据库压力下来了。但还有一类场景：用户下单成功后要发短信、记日志、同步搜索索引，这些**附属操作**不应该让用户等着。

这一章引入 **RabbitMQ 消息队列**：主流程写库 + 发消息，消费者异步处理。07 章解决热点读，08 章解决写后异步和解耦。学完后你的 demo 项目就具备「Spring Boot + MySQL + Redis + MQ」的完整骨架。

---

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

用户下单后，发短信、记日志、同步搜索索引这些**附属活**不该让用户干等——**消息队列（MQ）像留言板**：主流程写完订单就在留言板上贴一条「订单 xxx 已创建」，后台同事（消费者）有空再去看留言慢慢处理。

### 0.2 你需要提前知道什么（真不会就先跳到哪一章）

| 你已会 | 可以直接学本章 |
|--------|----------------|
| 04～07 章：Spring Boot、MySQL、Redis | ✅ 本章 |
| 会跑 Docker、改 `application.yml` | ✅ 本章 |
| 没学过 Spring Boot | 先学 **04 Spring Boot 核心** |
| 没学过 Redis | 建议先学 **07 Redis**（本章幂等示例用到） |

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

- ☐ 能说出 MQ 三大价值：异步、解耦、削峰
- ☐ 能画出 Producer → Exchange → Queue → Consumer 流转
- ☐ 会用 Docker 启动 RabbitMQ 并登录管理台
- ☐ 能在 Spring Boot 里配置 Exchange、Queue、Binding
- ☐ 会写 `RabbitTemplate.convertAndSend` 和 `@RabbitListener`
- ☐ 理解手动 ACK、消息重复消费、幂等去重思路
- ☐ 知道消息丢失的三个环节及基本对策
- ☐ 听说过死信队列、延迟队列、prefetch
- ☐ 能完成 demo 下单后发 MQ 并联调验证

### 0.4 建议学习时长与节奏

| 阶段 | 内容 | 建议时长 |
|------|------|----------|
| 第 1 天 | §1～§4 概念 + Docker 启动 + 管理台浏览 | 2 小时 |
| 第 2 天 | §4.1 手把手接入 demo（九步） | 3 小时 |
| 第 3 天 | §9～§12 可靠性 + §39～§40 死信 | 2 小时 |
| 复盘 | 闭卷自测 + 费曼检验 | 30 分钟 |

**节奏建议**：§4.1 每一步都要看到「预期输出」再继续；管理台 Queues 页面是验证消费是否成功的关键。

### 0.5 学完本章你能做什么（可验证的具体动作）

1. `docker run` 启动 RabbitMQ，浏览器打开 `http://localhost:15672` 登录
2. 在 demo 项目下单后，控制台看到「异步处理订单：…」
3. 管理台 `order.created.queue` 的 Ready = 0（消息已被消费）
4. 解释为什么下单接口不直接发短信（同步慢、耦合重）
5. 用 Redis `SETNX` 实现消费幂等，重复消息不重复发短信

---

## 1. 为什么需要消息队列

很多业务逻辑不一定要在主流程里同步做完。

比如下单之后可能还要：

- 发短信
- 发邮件
- 记录日志
- 通知其他系统

如果这些都同步做，主流程会变慢。

这时候消息队列就能发挥作用。

## 2. 消息队列的核心价值

**消息队列（Message Queue, MQ）**：生产者发消息、Broker 暂存、消费者异步拉取处理的中间件。
**生活类比**：**留言板异步**——前台（生产者）办完业务就在留言板贴条「请处理订单 123」；后台（消费者）不用站在前台等，有空看留言再处理；高峰期留言堆着，慢慢消化（削峰）。
**为什么重要**：解耦主流程与附属流程；提升接口响应速度；保护下游不被瞬时流量打垮。
**本章用到的地方**：§2～§4、§4.1 demo 全流程

### 2.1 异步

把一些没必要同步完成的事情放到后面处理。

### 2.2 解耦

下单服务不需要直接依赖所有后续服务。

### 2.3 削峰

高峰期先把任务放进队列，慢慢消费。

## 3. RabbitMQ 的基本概念

**RabbitMQ**：基于 AMQP 协议的开源消息 Broker，支持多种交换机路由模式。
**生活类比**：**带分拣中心的留言板系统**——Producer 把信交给 Exchange（分拣员），按 RoutingKey（地址标签）分到不同 Queue（信箱），Consumer（收信人）从信箱取信处理。
**为什么重要**：Java/Spring 生态里最常上手的企业级 MQ 之一；路由灵活、管理台可视化。
**本章用到的地方**：§3～§6、§4.1 配置类

你先掌握这几个核心角色：

- **Producer（生产者）**：发消息的一方，像在前台贴留言的人
- **Consumer（消费者）**：收消息并处理的一方，像后台看留言的同事
- **Queue（队列）**：消息的缓冲区，留言条实际存放的板子
- **Exchange（交换机）**：路由中枢，决定留言贴到哪块板子上
- **RoutingKey（路由键）**：路由规则，像留言分类标签「order.created」

| 角色 | 生活类比（留言板） | 代码里常见类/注解 |
|------|-------------------|-------------------|
| Producer | 前台贴留言 | `RabbitTemplate.convertAndSend` |
| Exchange | 分拣员 | `DirectExchange` / `@Bean` |
| Queue | 留言板 | `Queue` / `@RabbitListener(queues=...)` |
| Consumer | 后台处理员 | `@RabbitListener` 方法 |
| RoutingKey | 留言分类 | `"order.created"` 字符串 |

## 4. 一条消息大致怎么流动

<!-- 修改说明: 新增 RabbitMQ 消息流转 Mermaid 图 -->

```mermaid
flowchart LR
    P[Producer 生产者] -->|routing key| E[Exchange 交换机]
    E -->|binding| Q1[Queue 订单队列]
    E -->|binding| Q2[Queue 通知队列]
    Q1 --> C1[Consumer 订单消费者]
    Q2 --> C2[Consumer 通知消费者]
```

1. 生产者发送消息到交换机
2. 交换机根据规则把消息路由到队列
3. 消费者监听队列并消费消息

---

<!-- 修改说明: 新增手把手接入 demo 项目 -->

## 4.1 手把手：demo 项目接入 RabbitMQ

### 第一步：Docker 启动 RabbitMQ

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 确保 Docker Desktop 已启动 | 托盘图标正常 | Windows 先开 Docker |
| 2 | 执行下方 `docker run` 命令 | 输出一长串容器 ID | 5672 被占用则改 `-p 5673:5672` |
| 3 | `docker ps` 查看 | STATUS 为 Up，端口 5672/15672 | `docker logs study-rabbitmq` 查日志 |
| 4 | 浏览器打开 `http://localhost:15672` | 登录页 | guest 仅限 localhost |
| 5 | 用 guest/guest 登录 | Overview 页显示版本信息 | 见 §38.1 ACCESS_REFUSED |

```powershell
docker run -d --name study-rabbitmq -p 5672:5672 -p 15672:15672 -e RABBITMQ_DEFAULT_USER=guest -e RABBITMQ_DEFAULT_PASS=guest rabbitmq:3-management
```

```bash
docker ps
# 预期输出：
# CONTAINER ID   IMAGE                   STATUS    PORTS
# xxxx           rabbitmq:3-management   Up ...    0.0.0.0:5672->5672/tcp, 0.0.0.0:15672->15672/tcp
```

**管理台验证**：浏览器打开 `http://localhost:15672`，账号 `guest` / `guest`

```text
# 预期：进入 Overview 页面，显示 RabbitMQ 版本和节点信息
# 左侧菜单：Connections / Channels / Exchanges / Queues
```

### 第二步：pom.xml 追加依赖

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 在 `<dependencies>` 内追加 `spring-boot-starter-amqp` | Maven 重新下载依赖 | 检查 Spring Boot 父 POM 版本 |
| 2 | IDE 刷新 Maven 项目 | 无红色 import | `mvn dependency:tree` 看 amqp |
| 3 | 确认与现有 starter 无冲突 | 编译通过 | 版本由 BOM 管理即可 |

```xml
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-amqp</artifactId>
</dependency>
```

### 第三步：application.yml 配置

```yaml
spring:
  rabbitmq:
    host: localhost
    port: 5672
    username: guest
    password: guest
    listener:
      simple:
        acknowledge-mode: manual
        prefetch: 1
    publisher-confirm-type: correlated
```

### 第四步：声明 Exchange、Queue、Binding

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 新建 `config/RabbitMQConfig.java` | 类上有 `@Configuration` | 包名与项目一致 |
| 2 | 启动项目 | 日志无 `PRECONDITION_FAILED` | 队列参数变了需删旧队列 |
| 3 | 管理台 → Exchanges | 出现 `order.exchange` | 检查 `@Bean` 是否被扫描 |
| 4 | 管理台 → Queues | 出现 `order.created.queue` | 检查 Queue 名称常量 |
| 5 | 管理台 → Bindings | exchange 绑到 queue，routing key 正确 | 见 §38.1 NOT_FOUND |

`config/RabbitMQConfig.java`：

```java
package com.example.demo.config;

import org.springframework.amqp.core.Binding;
import org.springframework.amqp.core.BindingBuilder;
import org.springframework.amqp.core.DirectExchange;
import org.springframework.amqp.core.Queue;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class RabbitMQConfig {

    public static final String ORDER_EXCHANGE = "order.exchange";
    public static final String ORDER_QUEUE = "order.created.queue";
    public static final String ORDER_ROUTING_KEY = "order.created";

    @Bean
    public DirectExchange orderExchange() {
        return new DirectExchange(ORDER_EXCHANGE, true, false);
    }

    @Bean
    public Queue orderQueue() {
        return new Queue(ORDER_QUEUE, true);
    }

    @Bean
    public Binding orderBinding() {
        return BindingBuilder.bind(orderQueue())
                .to(orderExchange())
                .with(ORDER_ROUTING_KEY);
    }
}
```

**逐行读 `RabbitMQConfig`**：

| 行号/代码 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `public static final String ORDER_EXCHANGE` | 交换机名常量，全局统一引用 | 拼写不一致导致 NOT_FOUND |
| `new DirectExchange(..., true, false)` | 直连交换机；durable 持久化；不自动删除 | 第二个 false=autoDelete，true 则无绑定时删交换机 |
| `new Queue(ORDER_QUEUE, true)` | 持久化队列 | false 则 Broker 重启队列消失 |
| `BindingBuilder.bind(queue).to(exchange).with(KEY)` | 绑定：该 routing key 的消息进此队列 | key 与 Producer send 时不一致则消息路由不到 |
| 三个 `@Bean` | Spring 容器启动时自动声明 AMQP 对象 | 漏 `@Configuration` 则 Bean 不加载 |

### 第五步：消息体 DTO

```java
package com.example.demo.dto;

import java.io.Serializable;

public class OrderMessage implements Serializable {
    private Long orderId;
    private Long userId;
    private String orderNo;

    public OrderMessage() {}

    public OrderMessage(Long orderId, Long userId, String orderNo) {
        this.orderId = orderId;
        this.userId = userId;
        this.orderNo = orderNo;
    }

    public Long getOrderId() { return orderId; }
    public void setOrderId(Long orderId) { this.orderId = orderId; }
    public Long getUserId() { return userId; }
    public void setUserId(Long userId) { this.userId = userId; }
    public String getOrderNo() { return orderNo; }
    public void setOrderNo(String orderNo) { this.orderNo = orderNo; }
}
```

### 第六步：生产者

```java
package com.example.demo.service;

import com.example.demo.config.RabbitMQConfig;
import com.example.demo.dto.OrderMessage;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.stereotype.Service;

@Service
public class OrderMessageProducer {

    private final RabbitTemplate rabbitTemplate;

    public OrderMessageProducer(RabbitTemplate rabbitTemplate) {
        this.rabbitTemplate = rabbitTemplate;
    }

    public void sendOrderCreated(OrderMessage message) {
        rabbitTemplate.convertAndSend(
                RabbitMQConfig.ORDER_EXCHANGE,
                RabbitMQConfig.ORDER_ROUTING_KEY,
                message
        );
    }
}
```

**逐行读 `OrderMessageProducer.sendOrderCreated`**：

| 行号/代码 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `@Service` | Spring 管理 Bean，可注入到 OrderService | 漏注解则 @Autowired 失败 |
| `RabbitTemplate rabbitTemplate` | Spring AMQP 发送模板 | 未配 spring.rabbitmq 则连接失败 |
| `convertAndSend(exchange, routingKey, message)` | 序列化消息体并路由发送 | 三参数顺序不能颠倒 |
| 使用 `RabbitMQConfig` 常量 | 与 Config/Binding 名称一致 | 硬编码拼错则 NOT_FOUND |

### 第七步：消费者（手动 ACK + 幂等）

```java
package com.example.demo.consumer;

import com.example.demo.dto.OrderMessage;
import com.rabbitmq.client.Channel;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.amqp.support.AmqpHeaders;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.messaging.handler.annotation.Header;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.time.Duration;

@Component
public class OrderMessageConsumer {

    private final StringRedisTemplate redis;

    public OrderMessageConsumer(StringRedisTemplate redis) {
        this.redis = redis;
    }

    @RabbitListener(queues = "order.created.queue")
    public void handle(OrderMessage msg, Channel channel,
                       @Header(AmqpHeaders.DELIVERY_TAG) long tag) throws IOException {
        String dedupeKey = "mq:consumed:" + msg.getOrderNo();
        try {
            Boolean first = redis.opsForValue()
                    .setIfAbsent(dedupeKey, "1", Duration.ofDays(1));
            if (Boolean.FALSE.equals(first)) {
                System.out.println("重复消息，跳过：" + msg.getOrderNo());
                channel.basicAck(tag, false);
                return;
            }
            // 模拟异步通知：发短信、写日志等
            System.out.println("异步处理订单：" + msg.getOrderId() + "，用户：" + msg.getUserId());
            channel.basicAck(tag, false);
        } catch (Exception e) {
            channel.basicNack(tag, false, true);
        }
    }
}
```

**逐行读 `OrderMessageConsumer.handle`**：

| 行号/代码 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `@RabbitListener(queues = "order.created.queue")` | 监听指定队列，消息到达时调用本方法 | 队列名与 Config 不一致则收不到 |
| `Channel channel` + `DELIVERY_TAG` | 手动 ACK 需要 channel 和投递标签 | 自动 ACK 模式下 channel 可能为 null |
| `setIfAbsent(dedupeKey, "1", Duration.ofDays(1))` | Redis 幂等：同一 orderNo 只处理一次 | 无 Redis 则重复消息会重复发短信 |
| `Boolean.FALSE.equals(first)` | 已处理过则跳过业务 | 用 `== false` 可能对 null NPE |
| `channel.basicAck(tag, false)` | 告诉 Broker 消费成功，可删消息 | 不 ACK 则消息一直 Unacked |
| `channel.basicNack(tag, false, true)` | 失败且 requeue=true，消息回队列重试 | 无限重试可能打爆消费者 |

### 第八步：下单后发送消息

在 `OrderService.createOrder` 事务提交成功后：

```java
orderMessageProducer.sendOrderCreated(
        new OrderMessage(order.getId(), order.getUserId(), order.getOrderNo())
);
```

### 第九步：运行验证

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 启动 Redis、RabbitMQ、demo 应用 | Spring Boot 启动无报错 | 见 §38.1 Connection refused |
| 2 | 调用下单 API（Postman/curl） | HTTP 200，订单创建成功 | 先确认 MySQL 正常 |
| 3 | 看应用控制台 | `异步处理订单：1，用户：1001` | 检查 @RabbitListener 是否扫描到 |
| 4 | 管理台 → Queues → `order.created.queue` | Ready=0，Total 递增 | Ready>0 说明消费者未 ack 或未启动 |
| 5 | 再下一单相同 orderNo（测幂等） | `重复消息，跳过：xxx` | Redis 未启动则幂等失效 |

1. 启动 demo 项目（确保 Redis、RabbitMQ 已运行）
2. 调用下单接口
3. 控制台预期输出：

```text
异步处理订单：1，用户：1001
```

4. 打开 RabbitMQ 管理台 → **Queues** → `order.created.queue`

```text
# 预期：Ready 为 0（消息已被消费），Message rates 有 publish/consume 曲线
```

---

## 5. 为什么不直接让生产者发给消费者

因为交换机和队列的存在让系统更灵活：

- 可以一对一
- 可以一对多
- 可以按规则分发

## 6. 常见交换机类型

| 类型 | 路由规则 | 生活类比（留言板） | 典型场景 |
|------|----------|-------------------|----------|
| **Direct** | routing key 完全匹配 | 按精确分类贴到指定板 | 订单创建、支付成功各一条队列 |
| **Topic** | 通配符 `*`、`#` 匹配 | 「order.*」收所有订单类留言 | 按业务前缀订阅 |
| **Fanout** | 忽略 routing key，广播 | 喇叭广播，所有板都贴一份 | 系统公告、缓存失效通知 |

### 6.1 Direct

按精确 routing key 路由。

### 6.2 Topic

按通配规则路由，更灵活。

### 6.3 Fanout

广播到所有绑定队列。

## 7. Spring Boot 中的基础使用思路

### 7.1 发送消息

```java
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.stereotype.Service;

@Service
public class OrderMessageService {

    private final RabbitTemplate rabbitTemplate;

    public OrderMessageService(RabbitTemplate rabbitTemplate) {
        this.rabbitTemplate = rabbitTemplate;
    }

    public void sendOrderMessage(Long orderId) {
        rabbitTemplate.convertAndSend("order.exchange", "order.create", orderId);
    }
}
```

### 7.2 接收消息

```java
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.stereotype.Component;

@Component
public class OrderConsumer {

    @RabbitListener(queues = "order.queue")
    public void handle(Long orderId) {
        System.out.println("收到订单消息：" + orderId);
    }
}
```

## 8. 真实项目里怎么用 RabbitMQ

比较常见的场景：

- 用户下单后异步发通知
- 注册后异步发欢迎邮件
- 订单创建后异步扣减积分
- 任务异步执行

## 9. 消息确认

为什么要确认：

- 防止消息丢了却没人知道

你要有这个基础认知：

- 生产者发送后最好有确认机制
- 消费者消费后也要确认处理成功

## 10. 消息重复消费

这在真实系统里很常见。

为什么可能重复：

- 网络抖动
- 消费确认异常
- 服务重试

解决思路：

- 幂等设计
- 唯一业务 ID 去重

比如订单消息消费时，可以先判断这个订单是否已经处理过。

## 11. 消息丢失

常见问题方向有三个：

- 生产者发丢了
- 队列存储丢了
- 消费者处理时丢了

你现阶段先知道：

- 需要可靠投递思路
- 需要消费确认
- 重要消息要考虑持久化

## 12. 消息积压

当生产速度远大于消费速度时，就可能积压。

常见原因：

- 消费者处理太慢
- 队列峰值太高
- 消费者实例不够

常见应对方向：

- 提升消费者并发
- 优化消费逻辑
- 拆分队列

## 13. 和 Redis 队列有什么区别

这类问题面试经常会问。

你可以这样理解：

- Redis 也能做简单消息结构
- RabbitMQ 更像专业消息队列
- 在可靠性、路由能力、确认机制上更适合业务消息

## 14. 什么时候适合用 RabbitMQ

适合：

- 业务异步通知
- 中小型系统消息流程
- 需要相对完善消息机制的场景

## 15. 这一章的项目建议

你最好在项目里落地一个真实场景，比如：

### 下单异步通知

流程：

1. 用户下单
2. 订单写库成功
3. 发送消息
4. 消费者收到消息后发短信或写通知表

这样你在面试里就能讲：

- 为什么要异步
- 为什么用 MQ
- 如何避免重复消费

## 16. 这一章的练习建议

建议你自己完成：

1. 一个最基础的发送和消费 demo
2. 一个下单异步消息 demo
3. 一个幂等消费示例

## 17. 学完标准

如果你能做到下面这些，就说明这一章过关了：

- 知道为什么需要消息队列
- 知道 RabbitMQ 的核心角色
- 能写基础的发送和消费代码
- 知道消息重复消费和消息丢失的基本应对方向

## 18. 死信队列基础认知

死信队列通常用于处理：

- 消费失败无法正常处理的消息
- 过期消息
- 被拒绝的消息

它的价值是：

- 防止问题消息直接丢失
- 方便后续人工或程序补偿

## 19. 重试机制

消费者失败后，有时需要重试。

但要注意：

- 不能无限重试
- 否则可能把系统拖垮

所以你要逐步建立这个认知：

- 重试要有次数控制
- 失败要能落日志或进死信队列

## 20. 消息顺序问题

有些业务对顺序敏感，比如：

- 同一个订单的状态变更

这时就要考虑：

- 同一个业务对象的消息尽量按顺序处理

## 21. 消费幂等为什么重要

因为消息系统里“至少一次投递”很常见。

所以业务系统应该默认接受这样一个事实：

- 同一条消息可能收到多次

常见解决方向：

- 唯一业务 ID
- 状态判断
- 去重表

## 22. RabbitMQ 和 Kafka 的简单比较

你现在可以先这样理解：

- RabbitMQ 更偏业务消息
- Kafka 更偏高吞吐日志/流式场景

RabbitMQ 更适合你当前阶段上手和做项目。

## 23. 这一章的进一步知识点

后面你还可以继续学习：

- 延迟队列
- 死信交换机
- 消息堆积治理
- 消费者并发控制
- 顺序消息

## 24. 队列、交换机、绑定关系再细一点

很多初学者容易把 RabbitMQ 理解成：

- 生产者直接把消息放进队列

但更准确地说，常见流程是：

1. 生产者发给交换机
2. 交换机根据绑定关系路由到队列
3. 消费者从队列消费

其中：

- 交换机负责路由
- 队列负责存放消息

## 25. Topic 路由示例

假设有路由键：

- `order.create`
- `order.pay`
- `user.register`

如果某个队列绑定：

- `order.*`

那么它可以接收到：

- `order.create`
- `order.pay`

这就是 Topic 的灵活性所在。

## 26. 消息持久化基础认知

如果希望 RabbitMQ 重启后消息尽量还在，就要有持久化思路。

你现在先知道三个层面：

- 队列是否持久化
- 交换机是否持久化
- 消息是否持久化

## 27. 手动 ACK 基础认知

为什么很多项目会关注 ACK：

- 因为消费成功和消费失败要有明确反馈

你现在可以先这样理解：

- 自动确认更简单
- 手动确认更可控

对于重要业务消息，手动确认往往更稳妥。

## 28. Prefetch 基础认知

消费者一次不要无限拿消息。

Prefetch 的作用可以粗略理解为：

- 控制消费者一次最多预取多少消息

这有助于避免：

- 单个消费者积压过多未处理消息

## 29. 延迟队列基础认知

延迟队列很适合这些场景：

- 订单超时取消
- 一段时间后执行通知

你现在先知道它是“不是立刻消费，而是延后处理”的消息方案即可。

## 30. 为什么 MQ 不能代替数据库

消息队列的核心职责是：

- 传递消息
- 削峰异步

它不是用来长期稳定保存业务主数据的。

所以要分清：

- 业务主数据：数据库
- 异步事件流转：MQ

## 31. 消费失败后的常见处理思路

通常可以有这些方向：

1. 记录日志
2. 有限次数重试
3. 进入死信队列
4. 人工补偿

这也是为什么真实 MQ 方案比“发个消息”复杂得多。

## 32. 业务中哪些功能适合 MQ

适合：

- 短信通知
- 邮件通知
- 日志异步写入
- 积分发放
- 下单后的非核心流程

不太适合：

- 极强一致且必须立刻完成的主流程核心写库

## 33. MQ 这一章的高频知识点总清单

建议整理这些点：

- 为什么用 MQ
- 异步、解耦、削峰
- Producer、Consumer
- Exchange、Queue、RoutingKey
- Direct、Topic、Fanout
- ACK
- 消息持久化
- 重复消费
- 消息丢失
- 消息积压
- 死信队列
- 延迟队列

---

## 34. Docker 启动 RabbitMQ

```bash
docker run -d --name study-rabbitmq -p 5672:5672 -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=guest -e RABBITMQ_DEFAULT_PASS=guest \
  rabbitmq:3-management
```

```bash
# 预期输出：一行容器 ID
# a1b2c3d4e5f6...

curl -s -o /dev/null -w "%{http_code}" -u guest:guest http://localhost:15672/api/overview
# 预期输出：200
```

管理台：`http://localhost:15672`（guest / guest）

**管理台创建队列（可选手动练习）**：

1. 登录 → **Queues and Streams** → **Add a new queue**
2. Name 填 `test.queue` → **Add queue**
3. 预期：队列列表出现 `test.queue`，Messages 为 0

---

## 35. Spring Boot 生产者 / 消费者

```java
// 发送
rabbitTemplate.convertAndSend("order.exchange", "order.created", orderId);

// 消费
@RabbitListener(queues = "order.created")
public void onMessage(String orderId, Channel ch, @Header(AmqpHeaders.DELIVERY_TAG) long tag)
    throws IOException {
    try {
        // 业务处理
        ch.basicAck(tag, false);
    } catch (Exception e) {
        ch.basicNack(tag, false, true);
    }
}
```

`acknowledge-mode: manual` 开启手动 ACK。

---

## 36. 可靠性：生产 confirm、队列 durable、消费幂等后 ACK

幂等：`SETNX mq:consumed:{msgId}` 已存在则跳过。

---

## 37. 学完标准

- 说清解耦/异步/削峰；Exchange-Queue 模型
- Spring AMQP 发收消息；ACK 与重复消费对策

---

## 38. 分级练习

**基础**：管理台看队列  
**进阶**：下单后发 MQ  
**挑战**：死信队列配置

<!-- 修改说明: 新增分级练习参考答案 -->

### 参考答案

#### 基础：管理台看队列

1. 启动 RabbitMQ（§4.1 或 §34）
2. 运行 demo 项目，触发一次下单
3. 管理台 → **Queues** → 找到 `order.created.queue`
4. 检查：**Ready = 0**、**Total** 递增说明有消息流过

#### 进阶：下单后发 MQ

4.1 节完整代码即标准答案。验证清单：

- [ ] 下单接口返回成功
- [ ] 消费者控制台打印「异步处理订单」
- [ ] 管理台 Queue 无积压

#### 挑战：死信队列（思路 + 配置）

**场景**：消费失败 3 次后进入死信队列，人工排查。

`RabbitMQConfig` 中把原 `orderQueue` 改为带死信参数（需 `import java.util.HashMap; import java.util.Map;`）：

```java
@Bean
public Queue orderQueue() {
    Map<String, Object> args = new HashMap<>();
    args.put("x-dead-letter-exchange", ORDER_DLX);
    args.put("x-dead-letter-routing-key", "order.dead");
    return new Queue(ORDER_QUEUE, true, false, false, args);
}
```

消费者 `basicNack(tag, false, false)` 且不 requeue 时，消息进入 DLQ。管理台查看 `order.dead.queue` 即可。

---

<!-- 修改说明: 新增常见报错与排查 -->

## 38.1 常见报错与排查

| 报错信息（关键词） | 可能原因 | 解决方案 |
|-------------------|---------|---------|
| `Connection refused: localhost:5672` | RabbitMQ 未启动 | `docker start study-rabbitmq` |
| `ACCESS_REFUSED` | 用户名密码错 | 检查 `application.yml` 与 Docker 环境变量一致 |
| `NOT_FOUND - no exchange 'xxx'` | Exchange 未声明或名称不一致 | 确认 `RabbitMQConfig` 和业务代码中的名称相同 |
| 消息发送成功但无人消费 | 消费者未启动或队列名不匹配 | 检查 `@RabbitListener(queues=...)` 与 Binding |
| Queue 消息一直堆积 | 消费太慢或消费者挂了 | 看消费者日志；增加消费者实例；优化消费逻辑 |
| `PRECONDITION_FAILED` | 队列参数与已有队列冲突 | 删旧队列重建，或换队列名 |

---

## 39. 消息可靠性保证（生产者到消费者的完整链路）

### 39.0 手把手：验证 Publisher Confirm

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | `application.yml` 设 `publisher-confirm-type: correlated` | 启动无报错 | 拼写必须是 correlated |
| 2 | 在 Producer 里 `setConfirmCallback` 打印 ack | 发送后日志 `消息确认到达交换机` | 回调未注册则静默 |
| 3 | 故意 send 到不存在的 exchange 名 | ack=false 或 return 回调 | 需 `publisher-returns: true` |
| 4 | 改回正确 exchange 再发 | ack=true | 对照 Config 常量名 |
| 5 | 管理台 Queues 看消息是否入队 | Ready 或已被消费 | Binding/routing key 错误则不入队 |

### 39.1 发送端确认（Publisher Confirm）

```yaml
spring:
  rabbitmq:
    publisher-confirm-type: correlated   # 开启发送确认
    publisher-returns: true              # 路由失败退回
```

```java
@Component
public class OrderProducer {

    private final RabbitTemplate rabbitTemplate;

    public void sendOrder(OrderMessage msg) {
        // 设置确认回调
        rabbitTemplate.setConfirmCallback((correlationData, ack, cause) -> {
            if (ack) {
                log.info("消息确认到达交换机: {}", correlationData.getId());
            } else {
                log.error("消息未到达交换机: {}，原因: {}", correlationData.getId(), cause);
                // 补偿：重发或记 DB 等定时任务重试
            }
        });

        CorrelationData data = new CorrelationData(msg.getOrderNo());
        rabbitTemplate.convertAndSend("order.exchange", "order.create", msg, data);
    }
}
```

### 39.2 消费端确认（Manual Ack）

```yaml
spring:
  rabbitmq:
    listener:
      simple:
        acknowledge-mode: manual   # 手动确认
```

```java
@RabbitListener(queues = "order.queue")
public void handleOrder(OrderMessage msg, Channel channel,
                        @Header(AmqpHeaders.DELIVERY_TAG) long tag) {
    try {
        // 处理业务逻辑
        orderService.process(msg);
        channel.basicAck(tag, false);  // 确认消费成功
    } catch (Exception e) {
        log.error("消费失败", e);
        // basicNack(tag, false, true)：重新入队（重试）
        // basicNack(tag, false, false)：不重新入队，进入死信队列
    }
}
```

### 39.3 消息持久化

```java
// 队列持久化
@Bean
public Queue orderQueue() {
    return QueueBuilder.durable("order.queue").build();
}

// 消息持久化（默认已开启）
rabbitTemplate.convertAndSend("order.exchange", "order.create", msg, m -> {
    m.getMessageProperties().setDeliveryMode(MessageDeliveryMode.PERSISTENT);
    return m;
});
```

### 39.4 完整可靠性链路

```
发送端                     Broker                  消费端
  │                          │                        │
  ├─ ConfirmCallback ───────→│←── 交换机确认           │
  │                          │                        │
  │                          ├── 消息持久化到磁盘       │
  │                          │                        │
  │                          ├────────→ 消费者接收      │
  │                          │                        │
  │                          │                        ├─ 业务处理
  │                          │←── basicAck ──────────┤
  │                          │                        │
  │                          │  如果 basicNack：       │
  │                          │  → 重新入队（重试）     │
  │                          │  → 或进入死信队列      │
```

---

## 40. 死信队列（DLQ）详解

### 40.1 什么时候进死信

1. 消息被消费者拒绝（`basicNack` 且 `requeue=false`）
2. 消息 TTL 过期
3. 队列达到最大长度

### 40.2 完整配置

```java
@Configuration
public class DeadLetterConfig {

    public static final String ORDER_QUEUE = "order.queue";
    public static final String ORDER_DLX = "order.dlx";
    public static final String ORDER_DLQ = "order.dlq";

    // 死信交换机
    @Bean
    public DirectExchange deadLetterExchange() {
        return new DirectExchange(ORDER_DLX);
    }

    // 死信队列
    @Bean
    public Queue deadLetterQueue() {
        return new Queue(ORDER_DLQ);
    }

    // 死信队列绑定
    @Bean
    public Binding deadLetterBinding() {
        return BindingBuilder.bind(deadLetterQueue())
                .to(deadLetterExchange()).with("order.dead");
    }

    // 普通队列（带死信参数）
    @Bean
    public Queue orderQueue() {
        Map<String, Object> args = new HashMap<>();
        args.put("x-dead-letter-exchange", ORDER_DLX);
        args.put("x-dead-letter-routing-key", "order.dead");
        args.put("x-message-ttl", 60000);  // 消息 60s 未消费进死信
        return QueueBuilder.durable(ORDER_QUEUE).withArguments(args).build();
    }
}
```

---

## 41. 延迟消息（RabbitMQ 实现）

### 41.1 使用场景

- 下单后 30 分钟未支付自动取消
- 消息发送后 N 秒检查状态

### 41.2 TTL + 死信队列实现延迟

```
普通队列（无消费者，等 TTL 过期）
  → 消息过期 → 进入死信交换机
  → 死信队列（有消费者处理）
```

```java
// 延迟队列配置
@Bean
public Queue delayQueue() {
    Map<String, Object> args = new HashMap<>();
    args.put("x-dead-letter-exchange", "order.exchange");
    args.put("x-dead-letter-routing-key", "order.process");
    args.put("x-message-ttl", 30 * 60 * 1000);  // 30 分钟
    return QueueBuilder.durable("order.delay.queue").withArguments(args).build();
}

// 发送到延迟队列（不绑定消费者，等 TTL 后自动转到死信 → 被普通队列消费者处理）
```

---

## 42. 削峰填谷实战

```java
// 限流消费：每次只处理 N 条
spring:
  rabbitmq:
    listener:
      simple:
        prefetch: 50   # 每个消费者一次最多取 50 条
        concurrency: 5 # 消费者线程数
```

```
高并发下单 → MQ 缓冲（队列可堆积百万条）→ 消费者按自己节奏处理 → DB 压力平稳

          10000/s 写入
          ─────────
          │  MQ 队列 │
          ─────────
               ↓ 100/s 消费（平稳落库）
          ┌──────┐
          │ MySQL │
          └──────┘
```

---

## 43. 学完标准（扩充版）

- [ ] 理解 MQ 三大作用：异步、解耦、削峰
- [ ] 会用 Spring AMQP 发送/消费消息，配置交换机/队列/绑定
- [ ] 知道消息可靠性怎么保证（发送确认 + 手动 Ack + 持久化）
- [ ] 理解死信队列：什么时候进、怎么配置
- [ ] 会用 TTL + 死信实现延迟消息（订单超时取消场景）
- [ ] 知道消费者 `prefetch` 和并发配置对削峰的影响
- [ ] 能说出"消息丢失、重复消费、顺序消息"三类问题的基本对策

---

## 44. FAQ

**Q1：MQ 和 Redis List 做队列有什么区别？**  
Redis List 适合简单、低可靠场景；RabbitMQ 有 Exchange 路由、持久化、ACK、死信等专业机制，适合业务消息。

**Q2：消息会不会丢？**  
可能。三个环节：发送端（开 publisher confirm）、Broker（队列/消息持久化）、消费端（手动 ACK 成功后再 ack）。

**Q3：同一条消息为什么会消费两次？**  
网络抖动、消费后 ACK 前宕机、requeue 重试——MQ 常见「至少一次」语义，业务必须幂等。

**Q4：Exchange 和 Queue 必须都有吗？**  
常规 AMQP 模型：Producer → Exchange →（Binding）→ Queue → Consumer；Producer 不直接写 Queue。

**Q5：Direct 和 Fanout 区别？**  
Direct 按 routing key 精确匹配；Fanout 广播到所有绑定队列，忽略 routing key。

**Q6：`acknowledge-mode: manual` 是什么意思？**  
消费者必须显式 `basicAck`/`basicNack`；处理成功再 ack，失败可 nack 重试或进死信。

**Q7：prefetch 设 1 和设 50 区别？**  
prefetch=1 公平分发、慢消费者不会 hoard 消息；prefetch 大吞吐高但可能堆积在慢消费者本地。

**Q8：死信队列干什么用？**  
消费多次失败、TTL 过期、队列满被拒绝的消息进入 DLQ，便于人工排查和补偿，防静默丢失。

**Q9：下单和扣库存能用 MQ 吗？**  
扣库存是核心一致性逻辑，应和下单在同一事务或可靠同步流程；发短信、记日志适合 MQ 异步。

**Q10：RabbitMQ 和 Kafka 怎么选（入门阶段）？**  
RabbitMQ 适合业务通知、任务分发；Kafka 适合日志采集、超高吞吐流式。本阶段 demo 用 RabbitMQ 即可。

**Q11：guest 账号生产能用吗？**  
不能。guest 默认只允许 localhost；生产必须改强密码、限制网络、开 TLS。

**Q12：`NOT_FOUND - no exchange` 怎么排查？**  
检查 `RabbitMQConfig` 是否 `@Configuration` 被扫描；Exchange 名称与 `convertAndSend` 第一个参数一致。

---

## 45. 闭卷自测

1. **概念** MQ 三大作用（异步、解耦、削峰）各举一例。
2. **概念** Producer、Exchange、Queue、Consumer、RoutingKey 各是什么？
3. **概念** 为什么下单后主流程只发消息而不直接发短信？
4. **概念** 手动 ACK 和自动 ACK 哪个更适合重要业务？为什么？
5. **概念** 消息重复消费时，如何用 Redis 做幂等？
6. **概念** 死信队列在什么情况下收到消息？
7. **动手** 写出 `rabbitTemplate.convertAndSend` 的三个参数分别是什么。
8. **动手** 消费成功后应调用 `basicAck` 还是 `basicNack`？参数 `requeue` 填什么？
9. **综合** 队列 Ready 持续增长说明什么？可从哪几方面排查？
10. **综合** 画出「下单 → 写库 → 发 MQ → 消费者发短信」时序（文字步骤即可）。

### 45.1 自测参考答案

1. 异步：下单后发短信；解耦：订单服务不依赖短信服务接口；削峰：秒杀请求先堆积在队列，消费者按能力处理。
2. Producer 发消息；Exchange 路由；Queue 存储；Consumer 消费；RoutingKey 路由键。
3. 发短信慢且非核心，同步会拖长接口 RT；MQ 让主流程快速返回。
4. 手动 ACK；业务处理完再 ack，失败可 nack，防消息静默丢失。
5. `SETNX mq:consumed:{orderNo}` 成功才执行业务，已存在则直接 ack 跳过。
6. nack 且 requeue=false、TTL 过期、队列超长等。
7. exchange 名、routingKey、消息体对象。
8. `basicAck(tag, false)`；成功时不需要 nack。
9. 消费慢或消费者挂了；查消费者日志、增实例、优化逻辑、看 prefetch/concurrency。
10. ①用户下单 ②OrderService 事务写库 ③commit 后 Producer send ④Exchange 路由 ⑤Queue 暂存 ⑥Consumer 监听 ⑦幂等检查 ⑧发短信 ⑨basicAck。

---

## 46. 费曼检验

请在不看资料的情况下，用 3 分钟向朋友解释本章核心。

**对照提纲**：

1. **留言板比喻**：下单就像前台接待——登记完在留言板贴条「订单好了」，后台同事有空再处理，不用客户干等。
2. **四个角色**：发信人（Producer）→ 分拣中心（Exchange）→ 信箱（Queue）→ 收信人（Consumer）；RoutingKey 是地址标签。
3. **可靠性三件套**：发送确认、持久化、手动 ACK；重复消息靠幂等（如 Redis 去重），失败进死信别丢。

---

## 47. 本章核心速记卡

| 概念 | 一句话 | 类比 |
|------|--------|------|
| MQ | 异步传递消息 | 留言板 |
| Producer | 发消息 | 前台贴留言 |
| Exchange | 路由 | 分拣员 |
| Queue | 存消息 | 留言板/信箱 |
| Consumer | 处理消息 | 后台同事 |
| RoutingKey | 路由标签 | 留言分类 |
| ACK | 消费确认 | 处理后划掉留言 |
| DLQ | 死信队列 | 问题留言箱 |

**可靠性口诀**：发→confirm；存→durable；收→manual ack + 幂等；失败→有限重试 + 死信；积压→加消费者/优化逻辑/prefetch。

---

<!-- 修改说明: 新增下一章预告 -->

## 下一章预告

这一章你的 demo 在本地已经跑通：接口、MySQL、Redis、MQ 都有了——但还只在你的电脑上。怎么部署到服务器？怎么一条命令起全部中间件？怎么让 Nginx 把前端和后端串起来？

下一章（09 Linux、Docker、Nginx 部署基础）就是「上线入门」：`mvn package` 打 jar、`docker compose` 一键起环境、Nginx 反向代理 `/api`。08 章是「业务能异步」，09 章是「服务能对外跑」。

---

*下一章：09 Linux、Docker、Nginx 部署基础*
