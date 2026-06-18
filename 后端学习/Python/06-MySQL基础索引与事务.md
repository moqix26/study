# MySQL 基础、索引与事务

> **文件编码**：UTF-8。SQL 脚本、Python 源文件建议同样使用 UTF-8。

---

## 本章与上一章的关系

05 章你学会了用 **SQLAlchemy** 写 ORM、连 MySQL、在 FastAPI 里做 CRUD——但有个尴尬的情况：SQL 写对了，查询却慢到超时；或者金额字段用了 `float`，用户余额莫名其妙少了 0.01 元。这些问题根源都在 **MySQL 本身**：表怎么设计、索引怎么建、事务怎么隔离。

这一章从数据库底层补全认知。你会用 Docker 快速起一个 MySQL 环境（不用再折腾本地安装），学会设计电商三表、看懂 EXPLAIN 执行计划、理解 B+ 树索引为什么能让查询快 100 倍。05 章是「怎么写 SQL / ORM」，这一章是「SQL 在数据库里怎么跑」。

**与 Java 路线对照**：SQL 概念与 [Java 06 MySQL 基础索引与事务](../Java/06-MySQL基础索引与事务.md) 一致；本章侧重 **Python + SQLAlchemy + aiomysql/asyncmy** 的工程接入方式。

---

## 本章衔接

| 上一章（05） | 本章（06） | 下一章（07） |
|--------------|------------|--------------|
| SQLAlchemy Session、CRUD 接口 | 表设计、索引、B+ 树、事务隔离 | Redis 缓存扛热点读 |
| demo-api 能连 MySQL 写数据 | Docker 起 `study-mysql`，EXPLAIN 优化 | Cache Aside 模式 |
| `@transactional` 概念（Python 版 session.commit） | 防超卖 `UPDATE ... WHERE stock >= ?` | 穿透/击穿/雪崩 |

```mermaid
flowchart LR
    A[05 SQLAlchemy ORM] --> B[06 MySQL 索引与事务]
    B --> C[07 Redis 缓存]
    B --> D[持久化可靠但磁盘慢]
    C --> E[热点读走内存]
```

---

## 1. MySQL 在 Python 后端里的位置

MySQL 是 Python 后端（FastAPI / Django）最常用的关系型数据库之一。

你做的很多核心业务最终都要落到 MySQL：

- 用户数据
- 订单数据
- 商品数据
- 支付记录
- 库存信息

所以你学习 MySQL，不能只停留在「会写 `select * from user`」，而要逐步理解：

- 表怎么设计
- SQL 怎么写
- 索引怎么建
- 事务怎么保证一致性

**Python 侧对应关系**：

| MySQL 概念 | Python / SQLAlchemy |
|------------|---------------------|
| 建表 DDL | `Base.metadata.create_all()` 或 Alembic 迁移 |
| 查询 | `session.execute(select(User).where(...))` |
| 事务 | `with session.begin():` 或 `session.commit()` / `rollback()` |
| 连接池 | SQLAlchemy `create_engine(pool_size=...)` |
| 金额字段 | DB 用 `DECIMAL`；Python 用 `Decimal`，不用 `float` |

---

## 2. 数据库、表、行、列

你先把这几个概念吃透：

- **数据库**：一组相关数据的集合（如 `study_db`）
- **表**：存储某类数据的结构（如 `user`）
- **行**：一条记录
- **列**：一个字段

例如用户表可能有：

- `id`
- `username`
- `phone`
- `status`
- `create_time`

---

## 3. 建表基础

```sql
CREATE TABLE user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(64) NOT NULL,
    phone VARCHAR(20),
    age INT,
    status TINYINT NOT NULL DEFAULT 1,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 字段说明

- `bigint`：适合主键
- `varchar`：适合变长字符串
- `tinyint`：适合状态值
- `datetime`：适合记录时间
- `utf8mb4`：支持 emoji，比 `utf8` 更完整

---

## 4. 常见数据类型怎么选

### 4.1 整数

- `tinyint` / `int` / `bigint`
- 主键常用 `bigint`，扩展余地大

### 4.2 字符串

- `char` / `varchar` / `text`
- 一般业务字段优先考虑 `varchar`

### 4.3 金额

**用** `decimal(10,2)`，**不要用** `float` / `double`。

### 为什么金额用 DECIMAL，而不用 float？

**结论**：`float` 和 `double` 是二进制浮点数，很多十进制小数（如 0.1）无法精确表示，累加后会出现精度丢失。

**Python 侧验证**：

```python
>>> 0.1 + 0.2
0.30000000000000004
```

**正确做法**：

```python
from decimal import Decimal

price = Decimal("99.00")
total = price + Decimal("0.10")   # Decimal('99.10')
```

SQLAlchemy 模型中：

```python
from sqlalchemy import Numeric
from decimal import Decimal

class Product(Base):
    __tablename__ = "product"
    id: Mapped[int] = mapped_column(primary_key=True)
    price: Mapped[Decimal] = mapped_column(Numeric(10, 2), nullable=False)
```

**真实案例（模拟）**：某支付系统早期用 `DOUBLE` 存余额，连续充值 0.1 元 10 次后显示 0.9999999999999999 元；财务对账每天差几分钱。迁移到 `DECIMAL(10,2)` + Python `Decimal` 后问题解决。

---

## 4.1 手把手：Docker 启动 MySQL

05 章 / demo-api 项目需要 MySQL，这里教你用 Docker 一键启动。

### 前提

电脑已安装 [Docker Desktop](https://www.docker.com/products/docker-desktop/)。

### 启动命令

Windows PowerShell：

```powershell
docker run -d --name study-mysql -p 3306:3306 `
  -e MYSQL_ROOT_PASSWORD=123456 `
  -e MYSQL_DATABASE=study_db `
  mysql:8.0
```

```powershell
# 预期输出（一行容器 ID）：
# a1b2c3d4e5f6789012345678901234567890abcd

docker ps --filter name=study-mysql
# 预期：STATUS 为 Up，PORTS 含 0.0.0.0:3306->3306/tcp
```

### 进入 MySQL 客户端

```powershell
docker exec -it study-mysql mysql -uroot -p123456 study_db
```

```sql
-- 预期：mysql> 提示符
SHOW TABLES;
-- 预期：Empty set（新库尚无表）
```

### 初始化 demo-api 连接串

`.env` 或 `app/core/config.py`：

```python
DATABASE_URL = "mysql+asyncmy://root:123456@127.0.0.1:3306/study_db?charset=utf8mb4"
# 同步脚本可用：mysql+pymysql://root:123456@127.0.0.1:3306/study_db
```

### 端口被占用

```text
Bind for 0.0.0.0:3306 failed: port is already allocated.
```

解决：停本地 MySQL 服务，或改 `-p 3307:3306`，同时改连接串端口。

### 数据持久化（推荐）

```powershell
docker run -d --name study-mysql -p 3306:3306 `
  -e MYSQL_ROOT_PASSWORD=123456 `
  -e MYSQL_DATABASE=study_db `
  -v study-mysql-data:/var/lib/mysql `
  mysql:8.0
```

---

## 5. 基础 CRUD

### 5.1 插入

```sql
INSERT INTO user(username, phone, age)
VALUES ('zhangsan', '13800000000', 18);
```

SQLAlchemy 2.0 风格：

```python
from sqlalchemy import insert

await session.execute(
    insert(User).values(username="zhangsan", phone="13800000000", age=18)
)
await session.commit()
```

### 5.2 查询

```sql
SELECT id, username, age
FROM user
WHERE age >= 18;
```

```python
from sqlalchemy import select

result = await session.execute(select(User).where(User.age >= 18))
users = result.scalars().all()
```

### 5.3 更新

```sql
UPDATE user SET age = 20 WHERE id = 1;
```

### 5.4 删除

```sql
DELETE FROM user WHERE id = 1;
```

---

## 6. 条件查询、排序与分页

```sql
SELECT *
FROM user
WHERE status = 1 AND age >= 18
ORDER BY create_time DESC
LIMIT 0, 10;
```

`LIMIT offset, count` 中：

- `0` 是偏移量
- `10` 是取多少条

---

## 7. 聚合与多表查询

```sql
SELECT status, COUNT(*) AS total
FROM user
WHERE age >= 18
GROUP BY status
HAVING total > 5;
```

订单与用户联查：

```sql
SELECT o.id, o.total_amount, u.username
FROM `order` o
LEFT JOIN user u ON o.user_id = u.id;
```

### join 的核心理解

- `inner join`：两边都匹配才返回
- `left join`：左边都返回，右边没有就补 `null`

---

## 8. 表设计思路（电商三表）

### 8.1 用户表

核心字段：用户 ID、用户名、手机号、密码、状态、创建时间

### 8.2 商品表

核心字段：商品 ID、标题、价格（DECIMAL）、库存、状态

### 8.3 订单表

核心字段：订单 ID、用户 ID、订单金额、订单状态、创建时间

### 8.4 表设计原则

- 一张表描述一个核心实体
- 字段命名统一（蛇形 `user_id`）
- 状态字段明确并加注释
- 不要过早过度设计

---

## 9. 电商表设计完整示例

```sql
CREATE TABLE `user` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `username` VARCHAR(64) NOT NULL,
  `password` VARCHAR(128) NOT NULL COMMENT 'bcrypt hash',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `product` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `price` DECIMAL(10,2) NOT NULL,
  `stock` INT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1上架 0下架',
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `order` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `order_no` VARCHAR(32) NOT NULL,
  `user_id` BIGINT NOT NULL,
  `total_amount` DECIMAL(10,2) NOT NULL,
  `status` TINYINT NOT NULL COMMENT '0待付 1已付 2关闭',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_status_time` (`user_id`, `status`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**SQLAlchemy 模型对应**（节选）：

```python
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column
from sqlalchemy import String, Numeric, Integer, DateTime, func
from decimal import Decimal
from datetime import datetime

class Base(DeclarativeBase):
    pass

class Order(Base):
    __tablename__ = "order"
    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    order_no: Mapped[str] = mapped_column(String(32), unique=True)
    user_id: Mapped[int] = mapped_column(Integer, index=True)
    total_amount: Mapped[Decimal] = mapped_column(Numeric(10, 2))
    status: Mapped[int] = mapped_column(Integer)
    create_time: Mapped[datetime] = mapped_column(DateTime, server_default=func.now())
```

---

## 10. 索引

### 10.1 为什么需要索引

索引是帮助数据库快速定位数据的数据结构。没有索引时，数据库可能需要**全表扫描**。

### 10.2 常见索引类型

- 主键索引
- 唯一索引
- 普通索引
- 联合索引

```sql
CREATE INDEX idx_user_phone ON user(phone);
CREATE INDEX idx_user_status_age ON user(status, age);
```

---

## 11. B+ 树的基础理解

MySQL InnoDB 的常见索引底层是 **B+ 树**。

```mermaid
graph TD
    subgraph 根节点
        R["[10 | 20 | 30]"]
    end
    subgraph 中间节点
        I1["[5 | 8]"]
        I2["[15 | 18]"]
        I3["[25 | 28]"]
    end
    subgraph 叶子节点_存数据
        L1["1,2,3,4,5,6,7,8"]
        L2["10,11,12,13,14,15,16,17,18"]
        L3["20,21,22,23,24,25,26,27,28"]
        L4["30,31,32,33..."]
    end
    R --> I1
    R --> I2
    R --> I3
    I1 --> L1
    I2 --> L2
    I3 --> L3
    R --> L4
    L1 -.链表.-> L2
    L2 -.链表.-> L3
    L3 -.链表.-> L4
```

优点：

- 适合磁盘 IO：每个节点约 16KB，一次 IO 读一整页
- 层级低：百万级数据通常 3~4 层
- 支持范围查询：叶子节点链表相连

---

## 12. 最左前缀、覆盖索引与回表

联合索引 `(status, age)`：

```sql
-- 能用上索引
WHERE status = 1
WHERE status = 1 AND age = 18

-- 可能用不上（没从最左列 status 开始）
WHERE age = 18
```

**覆盖索引**：查询字段全在索引里，无需回主表。  
**回表**：二级索引找到主键后，再回聚簇索引取完整行。

---

## 13. 索引失效常见场景

1. 对索引列使用函数：`WHERE YEAR(create_time) = 2024`
2. 隐式类型转换：`WHERE phone = 13800138000`（phone 是 varchar）
3. 左模糊：`LIKE '%abc'`
4. OR 一侧无索引
5. 联合索引跳过最左列

---

## 14. EXPLAIN 实战解读

```sql
EXPLAIN SELECT * FROM `order`
WHERE user_id = 1 AND status = 0
ORDER BY create_time DESC
LIMIT 10;
```

| 列 | 关注点 |
|----|--------|
| type | `ALL` 全表扫（差）→ `range` → `ref` → `const`（好） |
| key | 实际用到的索引 |
| rows | 预估扫描行数，越小越好 |
| Extra | `Using filesort` / `Using temporary` 需优化 |

**type 速记**：`const` > `eq_ref` > `ref` > `range` > `index` > `ALL`

在 Python 里用原生 SQL 跑 EXPLAIN：

```python
from sqlalchemy import text

rows = await session.execute(text(
    "EXPLAIN SELECT * FROM `order` WHERE user_id = :uid LIMIT 10"
), {"uid": 1})
for row in rows:
    print(dict(row._mapping))
```

---

## 15. 事务

### 15.1 什么是事务

事务是一组操作，**要么都成功，要么都失败**。

例如下单：

1. 写订单
2. 扣库存
3. 扣余额

### 15.2 ACID

- **A**tomicity 原子性
- **C**onsistency 一致性
- **I**solation 隔离性
- **D**urability 持久性

### SQLAlchemy 事务写法

```python
async def create_order(session: AsyncSession, user_id: int, product_id: int, qty: int):
    async with session.begin():
        # 1. 扣库存（防超卖）
        result = await session.execute(
            text("""
                UPDATE product
                SET stock = stock - :qty
                WHERE id = :pid AND stock >= :qty
            """),
            {"pid": product_id, "qty": qty},
        )
        if result.rowcount == 0:
            raise ValueError("库存不足")

        # 2. 写订单
        await session.execute(
            insert(Order).values(
                order_no=f"ORD{user_id}{product_id}",
                user_id=user_id,
                total_amount=Decimal("99.00"),
                status=0,
            )
        )
    # begin 块结束自动 commit；异常则 rollback
```

FastAPI 依赖注入中常见模式：

```python
async def get_db():
    async with AsyncSessionLocal() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
```

---

## 16. 隔离级别与并发读问题

| 级别 | 脏读 | 不可重复读 | 幻读 |
|------|------|------------|------|
| READ UNCOMMITTED | 可能 | 可能 | 可能 |
| READ COMMITTED | 否 | 可能 | 可能 |
| REPEATABLE READ（MySQL 默认） | 否 | 否 | 理论上可能，InnoDB 用 MVCC+间隙锁缓解 |
| SERIALIZABLE | 否 | 否 | 否 |

```sql
SELECT @@transaction_isolation;
-- 预期：REPEATABLE-READ
```

### 并发读问题

- **脏读**：读到别人未提交的数据
- **不可重复读**：同一事务两次读同一行结果不同
- **幻读**：同一事务两次范围查询记录数不同

---

## 17. 锁与防超卖

### 悲观锁

```sql
START TRANSACTION;
SELECT stock FROM product WHERE id = 1 FOR UPDATE;
UPDATE product SET stock = stock - 1 WHERE id = 1;
COMMIT;
```

### 乐观锁（version 字段）

```sql
UPDATE product
SET stock = stock - 1, version = version + 1
WHERE id = 1 AND version = :old_version AND stock >= 1;
```

**推荐初学**：`UPDATE ... WHERE stock >= ?` 配合事务，简单有效。

---

## 18. 慢 SQL 优化步骤

1. 开启慢查询日志，定位 SQL
2. `EXPLAIN` 看 type、key、rows
3. 补/改索引（结合 WHERE、ORDER BY）
4. 避免 `SELECT *`
5. 深分页改游标：`WHERE id > :last_id LIMIT 10`

---

## 19. SQLAlchemy 连接池配置

```python
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker

engine = create_async_engine(
    DATABASE_URL,
    pool_size=10,          # 常驻连接数
    max_overflow=20,       # 超出 pool_size 的临时连接
    pool_pre_ping=True,    # 取连接前 ping，避免 stale 连接
    pool_recycle=1800,     # 30 分钟回收
    echo=False,
)
AsyncSessionLocal = async_sessionmaker(engine, expire_on_commit=False)
```

**排查**：日志出现 `QueuePool limit ... timeout` → 连接池耗尽，增大 `pool_size` 或检查是否有未关闭的 session。

---

## 20. 常见报错与排查

| 报错信息（关键词） | 可能原因 | 解决方案 |
|-------------------|---------|---------|
| `Access denied for user` | 用户名或密码错 | 核对 `.env`；Docker 用 `-e MYSQL_ROOT_PASSWORD` 设的密码 |
| `(2003) Can't connect to MySQL server` | MySQL 未启动或端口错 | `docker start study-mysql`；检查 3306/3307 |
| `Unknown column 'xxx'` | 字段名拼错或迁移未执行 | `DESC table_name`；跑 Alembic upgrade |
| `Duplicate entry for key 'uk_xxx'` | 违反唯一约束 | 业务先查再插；捕获 `IntegrityError` |
| `Lock wait timeout exceeded` | 事务持锁太久 | 缩短事务；检查未提交的大事务 |
| `You have an error in your SQL syntax` | 语法错误 | `` `order` `` 等保留字加反引号 |
| `Data too long for column` | 字符串超长 | 加大 VARCHAR 或校验输入 |
| `(1054) Unknown column in 'field list'` | ORM 模型与表结构不一致 | 对齐模型字段或重新迁移 |
| `QueuePool limit ... timeout` | 连接池耗尽 | 增大 pool_size；确保 session 关闭 |
| `Lost connection to MySQL server` | 连接超时/网络 | `pool_pre_ping=True`；检查 Docker 网络 |

Python 捕获唯一约束：

```python
from sqlalchemy.exc import IntegrityError

try:
    await session.commit()
except IntegrityError:
    await session.rollback()
    raise HTTPException(status_code=409, detail="用户名已存在")
```

---

## 21. 与前端联调的数据层注意点

当前端通过 [Vue 08 Axios 联调](../../前端学习/Vue/08-Axios网络请求与前后端联调.md) 调 demo-api 时，数据库层常见问题：

- 分页参数 `page`/`size` 应对应 SQL `LIMIT offset, count`，offset 过大需索引优化
- 接口返回金额用字符串或固定小数位 JSON，避免 Python `float` 序列化误差
- 时间字段统一 UTC 或带时区，前后端格式约定 `ISO 8601`

---

## 22. 分级练习

### 基础

建 user/product/order 三表，插入测试数据（见 §9 SQL）。

### 进阶

写「用户 ID=1 的已支付订单列表」SQL，用 EXPLAIN 对比有无 `idx_user_status_time`。

### 挑战

模拟无索引慢查询 vs 加联合索引后 `rows` 对比，截图记录。

---

## 23. 参考答案

### 基础：测试数据

```sql
INSERT INTO user (username, password) VALUES
('zhangsan', '$2b$12$xxx'), ('lisi', '$2b$12$yyy');

INSERT INTO product (name, price, stock) VALUES
('Python 编程', 89.00, 100),
('FastAPI 实战', 79.00, 50);

INSERT INTO `order` (order_no, user_id, total_amount, status) VALUES
('ORD20250618001', 1, 89.00, 1),
('ORD20250618002', 1, 79.00, 0),
('ORD20250618003', 2, 89.00, 1);
```

```sql
SELECT COUNT(*) FROM user;     -- 预期：2
SELECT COUNT(*) FROM product; -- 预期：2
SELECT COUNT(*) FROM `order`; -- 预期：3
```

### 进阶：用户订单列表

```sql
SELECT o.id, o.order_no, o.total_amount, o.status, o.create_time
FROM `order` o
WHERE o.user_id = 1 AND o.status = 1
ORDER BY o.create_time DESC
LIMIT 10;
```

有索引时 EXPLAIN 预期：`type=ref`, `key=idx_user_status_time`, `rows` 较小。

### 挑战：索引对比

```sql
ALTER TABLE `order` DROP INDEX idx_user_status_time;
EXPLAIN SELECT * FROM `order` WHERE user_id = 1;
-- 预期：type=ALL

CREATE INDEX idx_user_status_time ON `order`(user_id, status, create_time);
EXPLAIN SELECT * FROM `order` WHERE user_id = 1;
-- 预期：type=ref, key=idx_user_status_time
```

---

## 24. 学完标准

- [ ] 能自己设计 user/product/order 三表，金额用 `DECIMAL`
- [ ] 能写常见 SQL：JOIN、GROUP BY、分页
- [ ] 理解 B+ 树、最左前缀、覆盖索引、回表
- [ ] 会用 EXPLAIN 判断索引是否生效
- [ ] 能口述 ACID、四种隔离级别、脏读/幻读
- [ ] 会用 SQLAlchemy `session.begin()` 写事务
- [ ] 能写防超卖 `UPDATE ... WHERE stock >= ?`
- [ ] 会用 Docker 启动 `study-mysql` 并连接 demo-api

---

## 下一章预告

MySQL 能持久化存储，但磁盘 IO 是瓶颈——商品详情页每次查库，高峰期数据库可能扛不住。下一章（[07 Redis 核心原理与缓存实战](./07-Redis核心原理与缓存实战.md)）引入 **缓存层**：

- Redis 为什么比 MySQL 快（内存 + 单线程 + IO 多路复用）
- **Cache Aside** 模式：读时先查缓存、写时先更库再删缓存
- 用 **redis-py** 做商品详情缓存、ZSet 排行榜、SETNX 分布式锁

06 章解决「数据怎么存、怎么查快」，07 章解决「热点数据怎么扛高并发」——这是后端性能优化的第一道防线。对照 [Java 07 Redis](../Java/07-Redis核心原理与缓存实战.md) 可加深理解。

---

*下一章：07 Redis 核心原理与缓存实战*
