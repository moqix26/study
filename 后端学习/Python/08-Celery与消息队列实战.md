# Celery 与消息队列实战

> **文件编码**：UTF-8。任务 payload、日志输出建议 UTF-8。

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0、FAQ≥12、闭卷自测、费曼检验 -->

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

用户下单后，发短信、写通知、记日志这些**附属活**不该让用户干等——**消息队列（MQ）像留言板**：FastAPI 写完订单就在留言板上贴一条「订单 xxx 已创建」，**Celery Worker**（后台同事）有空再慢慢处理。

### 0.2 你需要提前知道什么（真不会就先跳到哪一章）

| 你已会 | 可以直接学本章 |
|--------|----------------|
| 04～07 章：FastAPI、MySQL、Redis | ✅ 本章 |
| 会跑 Docker、改 `.env` | ✅ 本章 |
| 没学过 FastAPI | 先学 **[04 FastAPI 核心](./04-FastAPI核心开发.md)** |
| 没学过 Redis | 建议先学 **[07 Redis](./07-Redis核心原理与缓存实战.md)**（幂等示例用到 SETNX） |
| Windows 开发 | Worker 用 `--pool=solo`，见 §5.2 |

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

- [ ] 能说出 MQ 三大价值：异步、解耦、削峰
- [ ] 能画出 Producer → Exchange → Queue → Celery Worker 流转
- [ ] 会用 Docker 启动 RabbitMQ 并登录管理台 `:15672`
- [ ] 能配置 `celery_app.py` 与 `@celery_app.task`
- [ ] 会在 FastAPI 里 `send_order_sms.delay(...)` 投递任务
- [ ] 能启动 Worker 并看到 `succeeded` 日志
- [ ] 理解 `task_acks_late`、幂等消费、重复消息对策
- [ ] 知道 Celery 与 asyncio 的分工（03 章对照）
- [ ] 能完成 demo-api 下单异步通知并联调验证

### 0.4 建议学习时长与节奏

| 阶段 | 内容 | 建议时长 |
|------|------|----------|
| 第 1 天 | §1～§3 概念 + Docker 启动 RabbitMQ | 2 小时 |
| 第 2 天 | §4～§5 手把手接入 demo-api | 3 小时 |
| 第 3 天 | §6～§7 可靠性 + 分级练习 | 2 小时 |
| 复盘 | FAQ + 闭卷自测 + 费曼检验 | 30 分钟 |

**节奏建议**：§5 必须开**两个终端**（FastAPI + Worker）；管理台 Queues 页 Ready=0 是消费成功的关键信号。

### 0.5 学完本章你能做什么（可验证的具体动作）

1. `docker run` 启动 RabbitMQ，浏览器打开 `http://localhost:15672` 登录
2. POST 下单后 Worker 打印「异步处理订单 ORD…」
3. 管理台对应队列 **Ready = 0**（消息已被消费）
4. 向面试官解释：为什么下单接口不直接发短信（同步慢、耦合重）
5. 用 Redis `SETNX` 实现消费幂等，重复消息不重复发短信

---

## 本章与上一章的关系

07 章 Redis 让读变快，但用户下单后还要：发短信、写通知表、记操作日志、同步搜索索引——这些若全在 FastAPI 请求里同步做完，接口 RT 会从 50ms 涨到 2s，且任一环节失败可能导致整单回滚。

**消息队列**把「主流程」和「附属流程」拆开：订单写库成功立刻返回，后续任务丢进队列由 **Celery Worker** 异步执行。这一章用 **Celery + RabbitMQ** 落地「下单异步通知」demo。

**与 Java 路线对照**：架构思想与 [Java 08 RabbitMQ 与消息队列实战](../Java/08-RabbitMQ与消息队列实战.md) 一致；Java 用 Spring AMQP，Python 用 Celery。

---

## 本章衔接

| 上一章（07） | 本章（08） | 下一章（09） |
|--------------|------------|--------------|
| Redis 缓存读 | Celery 异步写后任务 | Docker Compose 一键部署 |
| 同步 HTTP 内完成 CRUD | 下单接口 + 后台 Worker | uvicorn + Nginx 反代 |
| SETNX 防重复 | MQ 幂等消费 | 全栈容器化 |

```mermaid
flowchart LR
    User[用户] --> API[FastAPI 下单]
    API --> DB[(MySQL 写订单)]
    API --> MQ[RabbitMQ]
    API --> User2[立即返回 200]
    MQ --> Worker[Celery Worker]
    Worker --> SMS[发短信/写通知]
```

---

## 1. 为什么需要消息队列

### 1.1 三大价值

| 价值 | 说明 | 示例 |
|------|------|------|
| **异步** | 主流程不等待慢操作 | 下单后异步发短信 |
| **解耦** | 生产者和消费者独立演进 | 换短信供应商不改下单接口 |
| **削峰** | 高峰任务进队列慢慢消费 | 秒杀后 10 万通知排队处理 |

### 1.2 MQ 不能代替数据库

- 业务主数据：MySQL
- 异步事件流转：RabbitMQ / Celery

消息可能丢失、重复——数据库才是权威数据源。

### 1.3 深入：同步发短信 vs MQ 异步（小案例）

假设发短信平均 **800ms**，写 MySQL **20ms**：

| 方案 | 用户感知 RT | 短信失败对订单的影响 |
|------|-------------|---------------------|
| 同步：写库 + 发短信 | ~820ms | 可能整单回滚或用户长时间等待 |
| 异步：写库 + `delay()` | ~25ms | 订单已成功；短信失败 Worker 重试 |

**面试一句话**：主链路只保证「订单落库成功」；通知类 SLA 可以更低，用 MQ 隔离失败域。

---

## 2. RabbitMQ 核心概念

**RabbitMQ**：基于 AMQP 协议的开源消息 Broker，负责暂存与路由消息。
**生活类比**：**带分拣中心的留言板**——Producer 把信交给 Exchange（分拣员），按 RoutingKey（地址标签）分到不同 Queue（信箱），Consumer 从信箱取信处理。
**为什么重要**：Celery 默认 Broker；路由灵活、管理台可视化，与 Java Spring AMQP 概念一致。
**本章用到的地方**：§2～§5、§6 可靠性

```mermaid
flowchart TB
    P[Producer 生产者] --> E[Exchange 交换机]
    E -->|Binding + RoutingKey| Q1[Queue order.notify]
    E --> Q2[Queue order.dead]
    Q1 --> C[Celery Worker 消费者]
```

| 概念 | 说明 |
|------|------|
| Producer | 发消息的一方（FastAPI / Celery task.delay） |
| Exchange | 按规则把消息路由到队列 |
| Queue | 存消息 |
| Routing Key | 路由键，决定进哪个队列 |
| Consumer | Celery Worker 进程 |

Celery 默认用 RabbitMQ 作 Broker，也会用 Redis 作 Broker——本路线与 Java 路线对齐，**默认 RabbitMQ**。

---

## 3. Celery 是什么

**Celery**：Python 生态最流行的分布式任务队列框架，把「函数调用」变成「异步后台任务」。
**生活类比**：**外卖平台的骑手系统**——你（FastAPI）下单后只负责把订单推给平台（Broker），骑手（Worker）独立取单配送，你不用站在柜台等骑手回来。
**为什么重要**：附属流程与主 HTTP 请求解耦；可水平加 Worker 削峰。
**本章用到的地方**：§3.1～§5、§8 常用命令

Celery 是 Python 最流行的**分布式任务队列**框架：

- **Broker**：存任务消息（RabbitMQ）
- **Worker**：执行任务的进程
- **Backend**（可选）：存任务结果（Redis / RPC）

与 FastAPI 关系：FastAPI 只负责 `task.delay(...)` 投递；Worker 独立进程跑 `@app.task` 函数。

---

## 3.1 手把手：Docker 启动 RabbitMQ

```powershell
docker run -d --name study-rabbitmq `
  -p 5672:5672 -p 15672:15672 `
  -e RABBITMQ_DEFAULT_USER=guest `
  -e RABBITMQ_DEFAULT_PASS=guest `
  rabbitmq:3-management
```

```powershell
# 预期：一行容器 ID
docker ps --filter name=study-rabbitmq
# 预期：5672、15672 端口映射
```

管理台：`http://localhost:15672`（guest / guest）

验证 API：

```powershell
curl -u guest:guest http://localhost:15672/api/overview
# 预期：JSON 概览信息
```

**管理台创建队列（练习）**：

1. 登录 → **Queues and Streams** → **Add a new queue**
2. Name 填 `celery`（Celery 默认队列名）或自定义
3. 预期：队列列表出现新队列

---

## 4. demo-api 集成 Celery

### 4.1 项目结构

```text
demo-api/
├── app/
│   ├── main.py
│   ├── api/orders.py
│   ├── services/order_service.py
│   └── tasks/
│       ├── __init__.py
│       └── notify.py
├── celery_app.py          ← Celery 入口
├── requirements.txt
└── .env
```

### 4.2 依赖

```text
celery[redis]>=5.3.0
kombu>=5.3.0
```

RabbitMQ 作 Broker 时不需要 `redis` extra，但结果 Backend 常用 Redis：

```powershell
pip install "celery[redis]" redis
```

### 4.3 环境变量

```text
CELERY_BROKER_URL=amqp://guest:guest@localhost:5672//
CELERY_RESULT_BACKEND=redis://127.0.0.1:6379/1
DATABASE_URL=mysql+asyncmy://root:123456@127.0.0.1:3306/study_db
REDIS_URL=redis://127.0.0.1:6379/0
```

### 4.4 celery_app.py

```python
from celery import Celery
import os

celery_app = Celery(
    "demo_api",
    broker=os.getenv("CELERY_BROKER_URL", "amqp://guest:guest@localhost:5672//"),
    backend=os.getenv("CELERY_RESULT_BACKEND", "redis://127.0.0.1:6379/1"),
    include=["app.tasks.notify"],
)

celery_app.conf.update(
    task_serializer="json",
    accept_content=["json"],
    result_serializer="json",
    timezone="Asia/Shanghai",
    enable_utc=True,
    task_acks_late=True,           # 任务执行完再 ACK
    worker_prefetch_multiplier=1,  # 一次只预取 1 条，公平分发
)
```

**逐行读 `celery_app.py`（>10 行必配）**：

| 行/配置项 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `Celery("demo_api", broker=...)` | 应用名 + Broker 地址 | Broker 错则 Worker 连不上 |
| `include=["app.tasks.notify"]` | 注册任务模块 | 漏写则 `NotRegistered` |
| `task_serializer="json"` | 任务参数序列化格式 | 与 accept_content 不一致会拒收 |
| `task_acks_late=True` | 执行完再 ACK | 过早 ACK 可能丢任务 |
| `worker_prefetch_multiplier=1` | 预取条数 | 过大则 Worker 间负载不均 |

### 4.5 异步任务 app/tasks/notify.py

```python
import logging
from celery_app import celery_app

logger = logging.getLogger(__name__)

@celery_app.task(name="notify.send_order_sms", bind=True, max_retries=3)
def send_order_sms(self, order_no: str, phone: str) -> str:
    """模拟：下单成功后发短信 + 写通知"""
    try:
        logger.info("异步处理订单 %s，发送至 %s", order_no, phone)
        # 模拟外部 SMS API
        # requests.post("https://sms.example.com/send", ...)
        return f"OK:{order_no}"
    except Exception as exc:
        raise self.retry(exc=exc, countdown=5)
```

**逐行读 `notify.py`**：

| 行/装饰器 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `@celery_app.task(..., bind=True)` | 绑定 task 实例，可用 `self.retry` | 无 bind 则无法优雅重试 |
| `max_retries=3` | 最多重试 3 次 | 过大可能打爆 SMS 网关 |
| `name="notify.send_order_sms"` | 任务注册名 | 与 include 模块不一致会 NotRegistered |
| `logger.info(...)` | Worker 侧日志 | 生产应配 log 级别与轮转 |
| `self.retry(countdown=5)` | 5 秒后重试 | countdown 过短会加剧下游压力 |

### 4.6 FastAPI 与 Celery 协作边界

```text
FastAPI 进程（uvicorn）          Celery Worker 进程
─────────────────────          ───────────────────
接收 HTTP                         不监听 HTTP
写 MySQL（同步、要 fast）          可慢：SMS、日志、ES
task.delay() 只投递               执行 @task 函数体
不要在这里 requests 调 SMS 2s     这里可以调外部 API + retry
```

**常见误区**：在 FastAPI 里 `await asyncio.to_thread(send_sms)` 仍占用 Web worker 线程；高流量下应 **Celery 独立扩容 Worker**，而不是堆 uvicorn workers。

### 4.7 FastAPI 下单接口

```python
from fastapi import APIRouter, Depends
from pydantic import BaseModel
from app.tasks.notify import send_order_sms

router = APIRouter(prefix="/orders", tags=["orders"])

class OrderCreate(BaseModel):
    product_id: int
    quantity: int = 1

@router.post("")
async def create_order(body: OrderCreate, user_id: int = 1):
    # 1. 同步：写 MySQL（省略具体 ORM，见 05/06 章）
    order_no = f"ORD{user_id}{body.product_id}001"

    # 2. 异步：投递 Celery 任务
    send_order_sms.delay(order_no, "13800000000")

    return {"code": 0, "message": "下单成功", "data": {"order_no": order_no}}
```

---

## 5. 启动 Worker（手把手）

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | `docker start study-rabbitmq`（或 §3.1 新建） | `docker ps` 显示 Up | 见 §13 `Connection refused :5672` |
| 2 | 终端 1：`uvicorn app.main:app --reload` | `Application startup complete` | 检查 venv、端口 8000 |
| 3 | 终端 2：`celery -A celery_app worker --pool=solo -l info` | `celery@... ready` + 任务列表 | Windows 加 `--pool=solo` |
| 4 | `curl -X POST .../orders` | JSON 含 `order_no` | 看 FastAPI 日志 |
| 5 | 看 Worker 终端 | `succeeded in 0.xxs` | 见 §13 `NotRegistered` / 无 Worker |
| 6 | 打开 `:15672` Queues | Ready 短暂 >0 后变 0 | 任务积压查 Worker 是否运行 |

### 5.1 终端 1：FastAPI

```powershell
cd f:\study\demo-api
.\.venv\Scripts\Activate.ps1
uvicorn app.main:app --reload --port 8000
```

```text
# 预期输出：
INFO:     Uvicorn running on http://127.0.0.1:8000
INFO:     Application startup complete.
```

### 5.2 终端 2：Celery Worker

```powershell
cd f:\study\demo-api
.\.venv\Scripts\Activate.ps1
celery -A celery_app worker --loglevel=info --pool=solo
```

> Windows 上建议 `--pool=solo`（单进程）；Linux 生产可用 `--concurrency=4`。

**预期终端输出**：

```text
 -------------- celery@DESKTOP-XXX v5.3.x
--- ***** -----
-- ******* ---- Windows-10-...
- *** --- * ---
- ** ---------- [config]
- ** ---------- .> app:         demo_api:0x...
- ** ---------- .> transport:   amqp://guest:**@localhost:5672//
- ** ---------- .> results:     redis://127.0.0.1:6379/1
- *** --- * --- .> concurrency: 1 (solo)
-- ******* ----
--- ***** -----

[tasks]
  . notify.send_order_sms

[2025-06-18 10:00:00,000: INFO/MainProcess] Connected to amqp://guest:**@127.0.0.1:5672//
[2025-06-18 10:00:00,100: INFO/MainProcess] celery@DESKTOP-XXX ready.
```

### 5.3 触发任务

```powershell
curl -X POST http://127.0.0.1:8000/orders `
  -H "Content-Type: application/json" `
  -d '{"product_id": 1, "quantity": 1}'
```

**Worker 预期输出**：

```text
[2025-06-18 10:01:00,000: INFO/MainProcess] Task notify.send_order_sms[abc-123-def] received
[2025-06-18 10:01:00,050: INFO/MainProcess] 异步处理订单 ORD11001，发送至 13800000000
[2025-06-18 10:01:00,100: INFO/MainProcess] Task notify.send_order_sms[abc-123-def] succeeded in 0.05s: 'OK:ORD11001'
```

RabbitMQ 管理台 → Queues → 对应队列 **Ready = 0**（已消费）。

---

## 6. 消息可靠性

### 6.1 生产端

- Broker 持久化：RabbitMQ 队列 `durable=True`（Celery 默认配置可调）
- 任务 `bind=True` + `retry` 失败重试

### 6.2 消费端

- `task_acks_late=True`：任务执行成功后再 ACK
- 消费逻辑必须**幂等**

### 6.3 幂等消费示例

```python
import redis
from celery_app import celery_app

redis_client = redis.from_url("redis://127.0.0.1:6379/0")

@celery_app.task(name="notify.send_order_sms")
def send_order_sms(order_no: str, phone: str) -> str:
    dedup_key = f"mq:consumed:{order_no}"
    if not redis_client.set(dedup_key, "1", nx=True, ex=86400):
        return "SKIP:duplicate"
    # 真正发短信...
    return f"OK:{order_no}"
```

---

## 7. 重复消费与消息丢失

| 问题 | 原因 | 对策 |
|------|------|------|
| 重复消费 | 至少一次投递、Worker 崩溃重投 | 业务幂等、去重 key |
| 消息丢失 | 未持久化、ACK 过早 | durable 队列、acks_late |
| 消息积压 | 消费慢于生产 | 加 Worker 并发、优化任务逻辑 |

---

## 8. Celery 常用命令

```powershell
# 启动 Worker
celery -A celery_app worker --loglevel=info --pool=solo

# 查看注册任务
celery -A celery_app inspect registered

# 查看活跃任务
celery -A celery_app inspect active

# Flower 监控（可选）
pip install flower
celery -A celery_app flower --port=5555
# 浏览器 http://localhost:5555
```

---

## 9. 定时任务（认知）

```python
from celery.schedules import crontab

celery_app.conf.beat_schedule = {
    "cleanup-every-hour": {
        "task": "notify.cleanup_expired",
        "schedule": crontab(minute=0),
    },
}
```

```powershell
celery -A celery_app beat --loglevel=info
```

Beat 进程负责调度；Worker 负责执行——生产环境可分开部署。

---

## 10. Celery vs 直接 asyncio

| 场景 | 推荐 |
|------|------|
| 请求内等外部 HTTP（短） | asyncio + httpx |
| 下单后发短信、写日志（可延迟） | Celery |
| CPU 密集计算 | Celery + 多 Worker |
| 强实时 (<100ms) | 尽量同步或专用服务 |

03 章 asyncio 与 08 章 Celery 的分工：见 [03 并发编程选型](./03-Python并发编程与asyncio.md) 决策树。

### 10.1 交叉索引（前后章 + 平行系列）

| 主题 | 本章位置 | 延伸阅读 |
|------|----------|----------|
| Redis 幂等 SETNX | §6.3 | [07 §6 锁](./07-Redis核心原理与缓存实战.md)、[12 幂等](./12-高并发与分布式系统基础.md) |
| 下单主流程 | §4.6、§11 | [10 项目 §6](./10-后端项目实战与面试准备.md) |
| docker-compose 起 MQ | 下一章 | [09 §6 compose](./09-LinuxDockerNginx部署基础.md) |
| 延迟关单 | §9 Beat | [14 场景 §7](./14-高频场景设计与面试专题.md) |
| Java 对照 | 全章 | [Java 08 RabbitMQ](../Java/08-RabbitMQ与消息队列实战.md) |
| 系统设计 MQ | 概念 | [系统设计 04](../系统设计/04-消息队列架构设计.md) |

---

## 11. 完整下单通知流程

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as FastAPI
    participant M as MySQL
    participant Q as RabbitMQ
    participant W as Celery Worker
    participant S as 短信网关

    U->>F: POST /orders
    F->>M: INSERT order (事务)
    M-->>F: OK
    F->>Q: send_order_sms.delay()
    F-->>U: 200 下单成功
    Q->>W: 投递任务
    W->>S: 调用短信 API
    W->>M: INSERT notification (可选)
```

---

## 12. 与前端联调

前端 [Vue 08](../../前端学习/Vue/08-Axios网络请求与前后端联调.md) 下单后：

- 接口应**快速返回**（不等待短信）
- 前端可轮询「通知列表」或 WebSocket 推送（10 章扩展）
- 失败重试对用户透明，由 Worker 负责

---

## 13. 常见报错与排查

| 报错信息（关键词） | 可能原因 | 解决方案 |
|-------------------|---------|---------|
| `Connection refused :5672` | RabbitMQ 未启动 | `docker start study-rabbitmq` |
| `ACCESS_REFUSED` | 用户名密码错 | 核对 `CELERY_BROKER_URL` |
| `NotRegistered: notify.send_order_sms` | Worker 未加载任务模块 | `include=` 配置；重启 Worker |
| Worker 无输出 | 任务未投递或队列错 | 检查 `.delay()` 是否执行；管理台看队列 |
| `kombu.exceptions.OperationalError` | Broker 不可达 | Docker 网络；URL 格式 `amqp://...` |
| `Task always eager` 误解 | 配置了 `task_always_eager=True` | 开发调试可开，生产关闭 |
| Windows `PermissionError` | 多进程 pool 问题 | 使用 `--pool=solo` |
| 任务一直 PENDING | 无 Worker 在跑 | 启动 celery worker |
| `Backend unreachable` | Redis 结果 Backend 未起 | `docker start study-redis` |
| 重复发短信 | 未做幂等 | Redis SETNX 去重 |

---

## 14. 分级练习

### 基础

启动 RabbitMQ 管理台，观察下单后队列消息流入与消费。

### 进阶

实现完整 `send_order_sms` + 下单 `delay()`，截图 Worker 成功日志。

### 挑战

任务失败自动 retry 3 次，第 4 次进入死信（需配置 DLX 或 Celery 死信队列）。

---

## 15. 参考答案

### 基础

1. `docker start study-rabbitmq`
2. 打开 `http://localhost:15672`
3. POST 下单 → Queues 页看到消息短暂增加后 Ready=0

### 进阶

§4.4～§5.3 代码即标准答案。验证清单：

- [ ] FastAPI 返回 `order_no`
- [ ] Worker 打印「异步处理订单」
- [ ] 管理台无积压

### 挑战：重试思路

```python
@celery_app.task(bind=True, max_retries=3, default_retry_delay=5)
def send_order_sms(self, order_no: str, phone: str):
    try:
        if simulate_fail():
            raise RuntimeError("SMS API down")
        return "OK"
    except Exception as exc:
        raise self.retry(exc=exc)
```

第 4 次失败会抛 `MaxRetriesExceededError`，可配置 `task_reject_on_worker_lost` 或 RabbitMQ DLX 进死信队列（与 Java 08 死信配置思路相同）。

---

## 16. 高频知识点清单

- 异步 / 解耦 / 削峰
- Exchange / Queue / Routing Key
- Celery Broker / Worker / Backend
- `task.delay()` 与 `@app.task`
- 幂等消费、acks_late
- 重复消费与消息丢失
- Windows `--pool=solo`
- Beat 定时任务（认知）

---

## 17. 学完标准

- [ ] 说清为什么下单后要异步通知
- [ ] 能用 Docker 启动 RabbitMQ 并打开管理台
- [ ] 能配置 `celery_app.py` 并编写 `@task`
- [ ] 能启动 Worker 并看到任务 succeeded 日志
- [ ] 知道幂等消费与 `acks_late` 的作用
- [ ] 能区分 Celery 与 asyncio 的适用场景
- [ ] 对照 Java 08 能讲清 Exchange-Queue 模型

---

## 18. FAQ

**Q1：Celery 和 asyncio 都用 async，选哪个？**  
请求内短 IO（调外部 HTTP 几百 ms）用 asyncio；下单后发短信、写日志等**可延迟、可重试**的任务用 Celery。见 §10 对照表。

**Q2：Broker 用 RabbitMQ 还是 Redis？**  
本路线与 Java 对齐用 **RabbitMQ**；小项目 Redis 也可作 Broker，但路由与可靠性特性不如 RabbitMQ 完整。

**Q3：`delay()` 和 `apply_async()` 区别？**  
`delay(*args)` 是 `apply_async(args=args)` 的语法糖；需要 `countdown`、`queue` 等高级参数时用 `apply_async`。

**Q4：任务投递了但 Worker 没反应？**  
检查：① Worker 是否在跑 ② `include` 是否加载任务模块 ③ Broker URL 是否一致 ④ 管理台队列是否有积压。

**Q5：Windows 上 Celery 报错 PermissionError？**  
默认 prefork 多进程在 Windows 有问题，加 **`--pool=solo`** 或 `--pool=threads`。

**Q6：FastAPI 里能直接 `await send_order_sms()` 吗？**  
Celery 任务是**同步函数**在 Worker 进程跑；FastAPI 里只应 `delay()` 投递，不要 await 任务函数本身。

**Q7：消息会丢吗？**  
可能：Broker 未持久化、ACK 过早、Worker 崩溃。对策见 §6：`acks_late`、队列 durable、业务幂等。

**Q8：重复消费怎么办？**  
MQ 至少一次投递 + Worker 重试会导致重复；用 **Redis SETNX / DB 唯一键** 做业务幂等（§6.3）。

**Q9：结果 Backend 必须配吗？**  
发通知类任务通常**不需要**查返回值，Backend 可省略；需要 `AsyncResult` 查状态时用 Redis Backend。

**Q10：Beat 和 Worker 什么关系？**  
Beat 是**闹钟**（按 cron 触发任务）；Worker 是**干活的**。两者都要独立进程，生产可分开部署。

**Q11：和 Java Spring `@RabbitListener` 怎么对照？**  
Java 手写 Listener；Python 用 `@celery_app.task` + Worker 自动消费，概念上都是 Consumer。

**Q12：生产环境 Worker 开几个？**  
看任务 CPU/IO 与队列积压；Linux 用 `--concurrency=N`，配合监控 Flower 或 RabbitMQ 管理台。

---

## 19. 闭卷自测

1. **概念**：消息队列三大价值是什么？各举一个 demo-api 例子。
2. **概念**：Exchange、Queue、Routing Key 各扮演什么角色？
3. **概念**：Celery 的 Broker、Worker、Backend 分别是什么？
4. **概念**：`task_acks_late=True` 解决什么问题？
5. **概念**：为什么消费逻辑必须幂等？
6. **概念**：Celery 与 asyncio 各适合什么场景？
7. **动手**：写出启动 RabbitMQ Docker 的一条命令（含管理端口）。
8. **动手**：写出 FastAPI 里投递 `send_order_sms` 任务的一行代码。
9. **综合**：用户反馈「下单成功但没收到短信」，你按什么顺序排查？
10. **综合**：对照 Java 08，说明 Python Celery 与 Spring AMQP 的相同点与不同点。

### 自测参考答案

1. **异步**（下单立刻返回）、**解耦**（换短信商不改接口）、**削峰**（通知排队慢慢发）。
2. Exchange 路由；Queue 存消息；RoutingKey 决定进哪个 Queue。
3. Broker=RabbitMQ 存任务；Worker 执行 `@task`；Backend 存结果（可选 Redis）。
4. 任务执行成功后再 ACK，避免 Worker 中途崩溃导致消息丢失却已被确认。
5. 至少一次投递 + 重试会产生重复消息；幂等保证重复执行结果相同。
6. asyncio：请求内短 IO；Celery：可延迟、可重试、CPU 密集、跨进程后台任务。
7. `docker run -d --name study-rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management`
8. `send_order_sms.delay(order_no, "13800000000")`
9. ① Worker 是否运行 ② Worker 日志有无 error ③ RabbitMQ 队列是否积压 ④ 幂等 key 是否误 SKIP ⑤ 外部 SMS API（模拟）是否失败。
10. 相同：Producer-Exchange-Queue-Consumer、异步解耦、幂等、ACK。不同：Java 显式 `@RabbitListener`；Python Celery 封装 Broker 细节，`@task` + `delay()` 即可。

---

## 20. 费曼检验

请用 **3 分钟**向没学过后端的朋友解释：**「用户下单后，为什么还需要 Celery 和 RabbitMQ？」**

**对照提纲**：

1. **主流程**：下单 = 写 MySQL，用户希望**立刻**看到「下单成功」。
2. **附属流程**：发短信、写通知很慢，且可能失败——不应让用户干等或拖垮主事务。
3. **留言板模型**：FastAPI 把任务贴到 RabbitMQ；Celery Worker 后台慢慢做。
4. **好处**：接口快、模块解耦、高峰通知可排队（削峰）。
5. **注意**：数据库才是权威；MQ 可能重复，所以要**幂等**。

---

## 下一章预告

本地 FastAPI + MySQL + Redis + Celery 都跑通了，但还在「开发模式」——下一章（[09 Linux、Docker、Nginx 部署基础](./09-LinuxDockerNginx部署基础.md)）把 demo-api **部署上线**：

- **uvicorn** 生产启动、gunicorn + uvicorn worker
- **docker-compose** 一键起 mysql + redis + rabbitmq + api + worker
- **Nginx** 反向代理 `/api`，与 [Vue 08](../../前端学习/Vue/08-Axios网络请求与前后端联调.md) 前端静态资源同域部署

08 章解决「异步任务」，09 章解决「怎么让别人访问你的服务」。对照 [Java 09 部署](../Java/09-LinuxDockerNginx部署基础.md) 可对比 jar 部署与 Python 容器部署。

---

*下一章：09 Linux、Docker、Nginx 部署基础*

*本章已按 EXPANSION-STANDARD 扩充（§0 零基础导读 + 术语三件套 + 逐行读 + 手把手步骤表 + FAQ 12 + 闭卷 + 费曼）。*

**EXPANSION-STANDARD 自检**：☑ §0.1～0.5 ☑ 术语三件套 §2～§3 ☑ 逐行读 §4.4 ☑ 步骤表 §5 ☑ FAQ≥12 §18 ☑ 闭卷 10 题 §19 ☑ 费曼 §20
