# Celery 与消息队列实战

> **文件编码**：UTF-8。任务 payload、日志输出建议 UTF-8。

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

---

## 2. RabbitMQ 核心概念

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

### 4.6 FastAPI 下单接口

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

## 下一章预告

本地 FastAPI + MySQL + Redis + Celery 都跑通了，但还在「开发模式」——下一章（[09 Linux、Docker、Nginx 部署基础](./09-LinuxDockerNginx部署基础.md)）把 demo-api **部署上线**：

- **uvicorn** 生产启动、gunicorn + uvicorn worker
- **docker-compose** 一键起 mysql + redis + rabbitmq + api + worker
- **Nginx** 反向代理 `/api`，与 [Vue 08](../../前端学习/Vue/08-Axios网络请求与前后端联调.md) 前端静态资源同域部署

08 章解决「异步任务」，09 章解决「怎么让别人访问你的服务」。对照 [Java 09 部署](../Java/09-LinuxDockerNginx部署基础.md) 可对比 jar 部署与 Python 容器部署。

---

*下一章：09 Linux、Docker、Nginx 部署基础*
