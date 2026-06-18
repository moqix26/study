# SQLAlchemy 事务与接口工程化

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

## 下一章预告

05 章会写 SQL、会连库——但查询慢、索引失效、金额字段选错类型，根源在 **MySQL 本身**。下一章（06 MySQL 基础、索引与事务）从数据库底层补全：表设计、EXPLAIN、B+ 树、隔离级别——SQL 在数据库里怎么跑。

---

*下一章：06 MySQL 基础、索引与事务*
