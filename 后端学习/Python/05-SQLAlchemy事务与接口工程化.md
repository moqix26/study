# SQLAlchemy 事务与接口工程化

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0、FAQ≥12、闭卷自测、费曼检验 -->

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

**一句话**：把 04 章内存里的用户/商品 **搬进 MySQL**——用 **SQLAlchemy（≈ Java MyBatis）** 做 ORM、Session 和事务，让 demo-api 重启不丢数据。

**生活类比**：

| 概念 | 类比 | 对照 Java 05 |
|------|------|--------------|
| **ORM 模型** | 数据库表的「复印件类」 | Entity |
| **Session** | 一次办事的「工作台」 | SqlSession |
| **commit** | 盖章生效 | `@Transactional` 提交 |
| **rollback** | 办砸了全部撤销 | 事务回滚 |

---

### 0.2 你需要提前知道什么

| 水平 | 建议 |
|------|------|
| 学完 04 demo-api | 跟做本章，替换内存存储 |
| 学过 Java 05 MyBatis | 对照 Mapper vs ORM 风格差异 |
| 不懂 SQL | 可并行预览 06 章建表语句 |

---

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

- [ ] demo-api 用户 CRUD 持久化到 MySQL
- [ ] 理解 `Depends(get_db)` 与 Session 生命周期
- [ ] 会用 `select`、`db.get`、`commit/rollback`
- [ ] 完成「下单扣库存」事务练习
- [ ] 对照 MyBatis 说出 SQLAlchemy 的 Session/Model 对应关系
- [ ] 闭卷自测 ≥ 8/10

---

### 0.4 建议学习时长与节奏

| 阶段 | 时间 | 内容 |
|------|------|------|
| 连库 + 模型 | 1～2 天 | engine、Base、User 模型 |
| CRUD 接入 | 2 天 | Service 层改查库 |
| 事务 | 1～2 天 | 下单扣库存 |
| 自测 | 0.5 天 | 闭卷 + 费曼 |

---

### 0.5 学完本章你能做什么

1. 重启 demo-api 后用户数据仍在 MySQL 里。
2. 在 Swagger 完成用户 CRUD，echo 日志能看到 INSERT/SELECT。
3. 实现 `POST /orders` **事务内**扣库存 + 写订单。
4. 解释 Session 为何不能全局单例（线程/请求安全）。

---

## 本章与上一章的关系

04 章你用 FastAPI 写好了 REST 接口，但数据还在内存 `dict` 里——重启就丢，也没法做复杂查询。这一章要解决的正是这个问题：**让接口真正连上 MySQL 数据库**。

SQLAlchemy 是 Python 生态最主流的 ORM，支持 Core 与 ORM 两种风格。本资料以 **ORM 2.0 风格**（`Mapped`、`mapped_column`）为主。学完这章你会做 CRUD、分页、事务控制，并把 demo-api 从内存版升级为数据库版——与 [Java 05 MyBatis](../Java/05-MyBatis事务与接口工程化.md) 能力对齐。

---

## 1. 这份文档解决什么问题

- 怎么连 MySQL
- 怎么定义表模型、写 CRUD
- 怎么在 FastAPI 里管理 Session 生命周期
- 怎么用事务保证「下单扣库存」一致性
- 怎么把项目写得更像真实业务系统

---

## 2. SQLAlchemy 是什么

ORM（Object-Relational Mapping）：把数据库表映射成 Python 类，把行映射成对象。

| 对比 | MyBatis (Java) | SQLAlchemy (Python) |
|------|----------------|---------------------|
| SQL | 手写 XML/注解 | ORM 生成 + 可写原生 SQL |
| 模型 | Entity 类 | Declarative Base Model |
| 会话 | SqlSession | Session |

**术语（ORM）**：把数据库表映射成 Python 类，把一行映射成一个对象。  
**生活类比**：Excel 表 = MySQL 表，每一行 = 一个 User 对象，ORM 帮你「按行读写对象」而不是手写 INSERT 字符串。  
**为什么重要**：05 章起 demo-api 重启不丢数据；能力对齐 [Java 05 MyBatis](../Java/05-MyBatis事务与接口工程化.md)。  
**本章用到的地方**：§3 模型、§7 get_db、§15 事务。

### 2.1 database_url 逐行读

```python
DATABASE_URL = (
    "mysql+pymysql://root:123456@127.0.0.1:3306/study_db"
    "?charset=utf8mb4"
)
engine = create_engine(DATABASE_URL, echo=True, pool_pre_ping=True)
SessionLocal = sessionmaker(bind=engine, autoflush=False, autocommit=False)
```

| 字段/行 | 含义 | 改错会怎样 |
|---------|------|------------|
| `mysql+pymysql` | 驱动 dialect | 缺驱动 → ModuleNotFoundError |
| `root:123456` | 账号密码 | Access denied |
| `127.0.0.1:3306` | 主机端口 | Can't connect |
| `study_db` | 库名 | Unknown database |
| `charset=utf8mb4` | 中文 emoji | 乱码 |
| `echo=True` | 打印 SQL | 生产应 False |
| `pool_pre_ping` | 取连接前 ping | 避免 stale 连接 |

---

## 3. 环境与依赖

### 3.1 Docker 启动 MySQL（若 06 章未做）

```powershell
docker run -d --name study-mysql -p 3306:3306 `
  -e MYSQL_ROOT_PASSWORD=123456 `
  -e MYSQL_DATABASE=study_db mysql:8.0
docker ps
# 预期：study-mysql Up 0.0.0.0:3306->3306/tcp
```

### 3.2 安装依赖

```powershell
pip install sqlalchemy pymysql cryptography alembic
```

`cryptography` 供 MySQL 8 认证；`alembic` 做迁移（可选，进阶）。

---

## 4. 数据库连接与 Base

```python
# app/core/database.py
from sqlalchemy import create_engine
from sqlalchemy.orm import DeclarativeBase, sessionmaker
from app.core.config import settings

engine = create_engine(
    settings.database_url,
    pool_pre_ping=True,      # 连接池检测断线重连
    pool_size=10,
    max_overflow=20,
    echo=settings.debug,     # debug 时打印 SQL
)

SessionLocal = sessionmaker(bind=engine, autocommit=False, autoflush=False)


class Base(DeclarativeBase):
    pass


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
```

```python
# app/core/config.py 追加
database_url: str = "mysql+pymysql://root:123456@127.0.0.1:3306/study_db?charset=utf8mb4"
```

---

## 5. 定义 Model

```python
# app/models/user.py
from datetime import datetime
from sqlalchemy import String, Integer, DateTime, func
from sqlalchemy.orm import Mapped, mapped_column
from app.core.database import Base


class User(Base):
    __tablename__ = "user"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    username: Mapped[str] = mapped_column(String(64), unique=True, nullable=False)
    age: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    create_time: Mapped[datetime] = mapped_column(
        DateTime, server_default=func.now(), nullable=False
    )
```

### 5.1 建表

**方式 A：开发期快速建表**

```python
# app/main.py 启动时（仅开发）
from app.core.database import engine, Base
from app.models import user  # 确保模型被 import

Base.metadata.create_all(bind=engine)
```

**方式 B：执行 SQL（推荐与 06 章一致）**

```sql
CREATE TABLE user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(64) NOT NULL UNIQUE,
    age INT NOT NULL DEFAULT 0,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

## 6. CRUD 与 Service 改造

```python
# app/services/user_service.py
from sqlalchemy.orm import Session
from sqlalchemy import select
from app.models.user import User
from app.schemas.user import UserCreate, UserUpdate, UserVO


def list_users(db: Session) -> list[UserVO]:
    rows = db.scalars(select(User).order_by(User.id)).all()
    return [UserVO.model_validate(r) for r in rows]


def get_user(db: Session, user_id: int) -> UserVO | None:
    row = db.get(User, user_id)
    return UserVO.model_validate(row) if row else None


def create_user(db: Session, dto: UserCreate) -> UserVO:
    row = User(username=dto.username, age=dto.age)
    db.add(row)
    db.commit()
    db.refresh(row)
    return UserVO.model_validate(row)


def update_user(db: Session, user_id: int, dto: UserUpdate) -> UserVO | None:
    row = db.get(User, user_id)
    if not row:
        return None
    for k, v in dto.model_dump(exclude_unset=True).items():
        setattr(row, k, v)
    db.commit()
    db.refresh(row)
    return UserVO.model_validate(row)


def delete_user(db: Session, user_id: int) -> bool:
    row = db.get(User, user_id)
    if not row:
        return False
    db.delete(row)
    db.commit()
    return True
```

### 6.1 Router 注入 Session

```python
from fastapi import Depends
from sqlalchemy.orm import Session
from app.core.database import get_db

@router.get("")
def list_users(db: Session = Depends(get_db)):
    return Result.ok(user_service.list_users(db))
```

**每个请求**一个 Session，请求结束 `finally` 关闭——避免连接泄漏。

---

## 7. 查询进阶

### 7.1 条件与分页

```python
from sqlalchemy import select, func

def list_users_page(db: Session, page: int, size: int, keyword: str | None = None):
    stmt = select(User)
    if keyword:
        stmt = stmt.where(User.username.like(f"%{keyword}%"))
    total = db.scalar(select(func.count()).select_from(stmt.subquery()))
    rows = db.scalars(stmt.offset((page - 1) * size).limit(size)).all()
    return {"items": [UserVO.model_validate(r) for r in rows], "total": total, "page": page, "size": size}
```

### 7.2 原生 SQL（复杂报表）

```python
from sqlalchemy import text

def stats_by_age(db: Session):
    result = db.execute(text("SELECT age, COUNT(*) AS cnt FROM user GROUP BY age"))
    return [dict(row._mapping) for row in result]
```

复杂 SQL 可手写，与 MyBatis XML 思路相同。

---

## 8. 事务

### 8.1 为什么需要事务

下单要：扣库存 + 写订单 + 写流水。任一步失败应**全部回滚**。

### 8.2 单 Session 事务

```python
def create_order(db: Session, user_id: int, product_id: int, qty: int):
    try:
        product = db.get(Product, product_id)
        if not product or product.stock < qty:
            raise ValueError("库存不足")
        product.stock -= qty
        order = Order(user_id=user_id, product_id=product_id, qty=qty, status="CREATED")
        db.add(order)
        db.commit()
        db.refresh(order)
        return order
    except Exception:
        db.rollback()
        raise
```

### 8.3 深入：ACID

- **Atomicity**：commit 全成功，rollback 全撤销
- **Consistency**：业务约束不被破坏（库存不为负）
- **Isolation**：并发事务隔离级别见 [06 章](./06-MySQL基础索引与事务.md)
- **Durability**：commit 后落盘

### 8.4 装饰器封装（可选）

```python
from functools import wraps

def transactional(fn):
    @wraps(fn)
    def wrapper(db: Session, *args, **kwargs):
        try:
            result = fn(db, *args, **kwargs)
            db.commit()
            return result
        except Exception:
            db.rollback()
            raise
    return wrapper
```

Service 内若已 `commit`，注意避免双重提交。

---

## 9. 关系映射（一对多）

```python
# app/models/order.py
from sqlalchemy import ForeignKey
from sqlalchemy.orm import Mapped, mapped_column, relationship

class Order(Base):
    __tablename__ = "order"
    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    user_id: Mapped[int] = mapped_column(ForeignKey("user.id"))
    amount: Mapped[float] = mapped_column(nullable=False)

class User(Base):
    __tablename__ = "user"
    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    username: Mapped[str] = mapped_column(String(64))
    orders: Mapped[list["Order"]] = relationship(back_populates="user")
```

查询用户及订单：

```python
from sqlalchemy.orm import selectinload

stmt = select(User).options(selectinload(User.orders)).where(User.id == 1)
user = db.scalars(stmt).first()
```

---

## 10. 工程化实践

### 10.1 Schema 与 Model 分离

| 层 | 用途 |
|----|------|
| Model | 数据库表，含 SQLAlchemy 列 |
| Schema (Pydantic) | API 入参/出参，不含 ORM 细节 |

```python
class UserVO(BaseModel):
    model_config = ConfigDict(from_attributes=True)  # 允许从 ORM 对象构造

    id: int
    username: str
    age: int
```

Pydantic v2 用 `from_attributes=True`（v1 叫 `orm_mode`）。

### 10.2 目录结构（05 章完成后）

```text
demo-api/
├── app/
│   ├── models/
│   │   ├── user.py
│   │   └── product.py
│   ├── services/
│   ├── routers/
│   └── schemas/
├── sql/
│   └── schema.sql
└── alembic/          # 可选迁移
```

### 10.3 日志与 SQL 调试

```python
# echo=True 或 logging
import logging
logging.getLogger("sqlalchemy.engine").setLevel(logging.INFO)
```

---

## 11. 手把手：demo-api 接 MySQL 完整流程

### 步骤清单

1. Docker 起 MySQL，`study_db` 已创建
2. `pip install sqlalchemy pymysql`
3. 配置 `database_url`
4. 建 `User` Model + `schema.sql`
5. 改 `user_service` 接 Session
6. Router `Depends(get_db)`
7. `uvicorn app.main:app --reload`
8. Swagger 创建用户 → MySQL 客户端 `SELECT * FROM user;` 验证

```powershell
docker exec -it study-mysql mysql -uroot -p123456 study_db -e "SELECT * FROM user;"
# 预期：能看到刚插入的行
```

---

## 12. 请求链路（含数据库）

```mermaid
sequenceDiagram
    participant R as Router
    participant S as Service
    participant DB as Session
    participant MySQL as MySQL

    R->>S: create_user(dto)
    S->>DB: db.add(User)
    S->>DB: db.commit()
    DB->>MySQL: INSERT
    MySQL-->>DB: OK
    DB-->>S: User 实例
    S-->>R: UserVO
```

---

## 13. 常见报错与排查

| 报错 | 原因 | 解决 |
|------|------|------|
| `Can't connect to MySQL server` | MySQL 未启动或地址错 | docker ps；检查 host/port |
| `Access denied for user` | 账号密码错 | 核对 database_url |
| `Unknown database` | 库不存在 | 创建 study_db |
| `IntegrityError: Duplicate entry` | 唯一键冲突 | 业务层先查或捕获异常 |
| `DetachedInstanceError` | Session 关闭后访问懒加载 | commit 前 eager load 或 refresh |
| `Table doesn't exist` | 未建表 | create_all 或执行 schema.sql |
| `InvalidRequestError: Session is closed` | Session 已关仍使用 | 在 Depends 生命周期内操作 |
| ` pymysql.err.DataError` | 字段类型/长度不匹配 | 对照表结构 |
| 中文乱码 | charset 不对 | URL 加 `?charset=utf8mb4` |
| 连接池耗尽 | Session 未关闭 | 确保 get_db 有 finally close |

---

## 14. 练习建议

### 基础

1. 完成 User CRUD 接 MySQL
2. 用户名唯一冲突返回友好错误

### 进阶

3. 增加 `Product` 表（name, price DECIMAL, stock）
4. 实现 `POST /api/orders`：扣库存 + 写订单，事务包裹

### 挑战

5. 分页 + 关键字搜索
6. 用 Alembic 做第一次 migration

---

## 15. 参考答案（下单事务）

```python
# models 简化
class Product(Base):
    __tablename__ = "product"
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(128))
    stock: Mapped[int] = mapped_column(default=0)
    price: Mapped[float] = mapped_column()

class Order(Base):
    __tablename__ = "order"
    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    user_id: Mapped[int] = mapped_column()
    product_id: Mapped[int] = mapped_column()
    qty: Mapped[int] = mapped_column()
    amount: Mapped[float] = mapped_column()
    status: Mapped[str] = mapped_column(String(32), default="CREATED")


def place_order(db: Session, user_id: int, product_id: int, qty: int):
    product = db.get(Product, product_id)
    if not product:
        raise ValueError("商品不存在")
    if product.stock < qty:
        raise ValueError("库存不足")
    product.stock -= qty
    amount = float(product.price) * qty
    order = Order(user_id=user_id, product_id=product_id, qty=qty, amount=amount)
    db.add(order)
    db.commit()
    db.refresh(order)
    return order
```

---

## 16. 学完标准

- [ ] demo-api 用户 CRUD 数据持久化到 MySQL
- [ ] 理解 Session 生命周期与 `Depends(get_db)`
- [ ] 会用 `select`、`db.get`、`commit/rollback`
- [ ] 完成至少一个「多表 + 事务」练习
- [ ] 能读懂 echo 打印的 SQL

---

## 17. SQLAlchemy vs MyBatis 对照表

| 主题 | Java 05 MyBatis | Python 05 SQLAlchemy |
|------|-----------------|----------------------|
| 映射 | XML / 注解 SQL | Declarative Model |
| 查一条 | `selectById` | `session.get(User, id)` |
| 条件查询 | `<where>` | `select(User).where(...)` |
| 事务 | `@Transactional` | `session.commit()` / `rollback()` |
| 连接池 | HikariCP | SQLAlchemy pool |
| 分页 | PageHelper | `limit/offset` 或 04 章 Depends |

**术语（Session）**：ORM 与数据库之间的「工作单元」，跟踪对象变更。  
**为什么重要**：一个 HTTP 请求通常对应一个 Session，请求结束必须 close。

---

## 18. FAQ

**Q1：SQLAlchemy 和 Django ORM？**  
本路线用 SQLAlchemy + FastAPI；Django 自带 ORM，不混用。

**Q2：同步 Session 还是 async？**  
本资料 05 章以同步入门清晰；06 章后可选 async engine，概念相同。

**Q3：create_all 和生产迁移？**  
练手用 `create_all`；生产用 **Alembic**（练习挑战题）。

**Q4：DetachedInstanceError？**  
Session 关闭后访问懒加载属性；commit 前取齐数据或 refresh。

**Q5：IntegrityError 怎么处理？**  
捕获后 rollback，返回 409「用户名已存在」。

**Q6：和 Java @Transactional 一样吗？**  
思想一致：多步 DB 操作要么全成功要么全回滚；Python 显式 commit。

**Q7：为什么 get_db 用 yield？**  
FastAPI Depends 在请求结束后执行 finally，确保 Session 关闭。

**Q8：N+1 查询问题？**  
循环里 lazy load 会多次 SELECT；用 joinedload 或一次 join。

**Q9：float 存金额？**  
**不要**；MySQL 用 DECIMAL，Python 用 Decimal（02 章）。

**Q10：05 和 06 章顺序？**  
05 先能连库 CRUD；06 深入索引/SQL 优化。

**Q11：内存 dict 何时完全删除？**  
05 章练完 CRUD 后 Service 只走 DB。

**Q12：echo=True 干什么？**  
打印 SQL，学习用；生产关闭。

---

## 19. 闭卷自测

1. **概念**：ORM 解决什么问题？
2. **概念**：Session 为何不能全局单例共享？
3. **概念**：commit 和 rollback 各何时调用？
4. **概念**：SQLAlchemy 与 MyBatis 在 SQL 控制上的主要差异？
5. **概念**：`Depends(get_db)` 如何保证 Session 不泄漏？
6. **概念**：下单扣库存为什么必须包在同一事务？
7. **动手**：写出 `select(User).where(User.username == "admin")` 的用途。
8. **动手**：IntegrityError 后为什么要 rollback？
9. **综合**：04 内存版改 05 数据库版，Service 层改什么、Router 改什么？
10. **综合**：对照 Java 05，Mapper 接口相当于 SQLAlchemy 的哪一层？

### 自测参考答案

1. 用类/对象操作表，减少手写 SQL 拼接。
2. 多请求并发会互相污染同一 Session 状态。
3. 成功持久化 commit；异常或业务失败 rollback。
4. MyBatis SQL 显式；SQLAlchemy ORM 自动生成，也可写原生 SQL。
5. yield 后 finally `session.close()`，每请求独立 Session。
6. 防止扣了库存订单却写失败，或反之。
7. 按用户名查 User 记录。
8. 否则 Session 处于非法状态，后续操作报错。
9. Service 改查 DB；Router 通常只改 Depends 注入，路径可不变。
10. 类似 Repository/Mapper；ORM 可用 select/Query 表达。

---

## 20. 费曼检验

3 分钟解释：**「05 章为什么要把内存 dict 换成 MySQL + SQLAlchemy？」**

**对照提纲**：

1. **重启不丢数据**——内存 dict 是临时的，MySQL 持久化。
2. **SQLAlchemy** 像 Java 的 MyBatis，用类代表表，用 Session 办事。
3. **事务**保证下单扣库存要么一起成功要么一起撤销。
4. **下一章 06** 学怎么让 SQL 跑得更快（索引）。

---

## 下一章预告

05 章会写 SQL、会连库——但查询慢、索引失效、金额字段选错类型，根源在 **MySQL 本身**。下一章（06 MySQL 基础、索引与事务）从数据库底层补全：表设计、EXPLAIN、B+ 树、隔离级别——SQL 在数据库里怎么跑。

---

*下一章：06 MySQL 基础、索引与事务*
