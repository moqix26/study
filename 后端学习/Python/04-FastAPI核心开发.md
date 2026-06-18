# FastAPI 核心开发

## 本章与上一章的关系

03 章你学了 asyncio——知道 `async/await` 和非阻塞 IO。但真实后端工作不是写脚本就完事了：你要**对外提供 HTTP 接口**，让 [Vue 08](../../前端学习/Vue/08-Axios网络请求与前后端联调.md) / [React 08](../../前端学习/React/08-Axios网络请求与前后端联调.md) 的前端来调用。

这一章是转折点：**从「写 Python 程序」进入「写 Web 后端服务」**。FastAPI 基于 Starlette + Pydantic，自动生成 OpenAPI 文档、原生支持 async。学完这章，你能启动 demo-api、写 REST 接口、做参数校验和统一返回——这是后面接 SQLAlchemy、MySQL、Redis 的前置条件。

### FastAPI 请求处理流程

```mermaid
sequenceDiagram
    participant Browser as 浏览器/前端
    participant Uvicorn as Uvicorn
    participant Router as APIRouter
    participant Service as Service
    participant DB as 数据库(05章接入)

    Browser->>Uvicorn: HTTP GET /api/users/1
    Uvicorn->>Router: 路由匹配
    Router->>Service: 调用业务逻辑
    Service->>DB: 查数据(05章)
    DB-->>Service: 返回行
    Service-->>Router: UserSchema
    Router-->>Browser: JSON Result
```

---

## 1. FastAPI 是什么

FastAPI 是现代 Python Web 框架，特点：

- **高性能**：基于 Starlette（async）和 Uvicorn
- **类型驱动**：Pydantic 自动校验请求/响应
- **自动文档**：Swagger UI `/docs`、ReDoc `/redoc`
- **与前端友好**：JSON API、OpenAPI 契约

对比 [Java Spring Boot](../Java/04-SpringBoot核心开发.md)：

| 概念 | Spring Boot | FastAPI |
|------|-------------|---------|
| 路由 | `@GetMapping` | `@app.get` / `APIRouter` |
| 入参校验 | `@Valid` + Bean Validation | Pydantic `BaseModel` |
| 依赖注入 | `@Autowired` | `Depends()` |
| 统一返回 | `Result<T>` 类 | 自定义 `Result` Schema |
| 文档 | Swagger 插件 | 内置 |

---

## 2. 项目分层结构

```text
demo-api/
├── app/
│   ├── main.py              ← FastAPI 入口
│   ├── core/
│   │   ├── config.py        ← 配置
│   │   └── exceptions.py    ← 全局异常
│   ├── routers/
│   │   └── user.py          ← 路由（≈ Controller）
│   ├── services/
│   │   └── user_service.py  ← 业务逻辑
│   ├── schemas/
│   │   ├── user.py          ← Pydantic 入参/出参（≈ DTO/VO）
│   │   └── common.py        ← Result 等
│   └── models/              ← SQLAlchemy 模型（05 章）
├── requirements.txt
└── .env
```

| 层 | 职责 |
|----|------|
| routers | 接 HTTP、调 Service、返回 Schema |
| services | 业务逻辑，不直接碰 Request 对象 |
| schemas | 请求体/响应体的类型与校验 |
| models | 数据库表映射（05 章） |

---

## 2.1 手把手：从零创建 demo-api（内存版 CRUD）

### 第一步：创建项目

```powershell
mkdir demo-api
cd demo-api
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install "fastapi[standard]" uvicorn pydantic-settings python-multipart
pip freeze > requirements.txt
```

```text
# pip 预期输出片段：
# Successfully installed fastapi-x.x.x uvicorn-x.x.x pydantic-x.x.x ...
```

### 第二步：requirements.txt 参考

```text
fastapi[standard]>=0.115.0
uvicorn[standard]>=0.32.0
pydantic>=2.0
pydantic-settings>=2.0
python-multipart>=0.0.9
```

### 第三步：app/schemas/common.py

```python
from typing import Generic, TypeVar, Optional

T = TypeVar("T")

class Result(Generic[T]):
    code: int = 0
    message: str = "ok"
    data: Optional[T] = None

    @classmethod
    def ok(cls, data: T = None, message: str = "ok") -> "Result[T]":
        return cls(code=0, message=message, data=data)

    @classmethod
    def fail(cls, code: int, message: str) -> "Result":
        return cls(code=code, message=message, data=None)
```

Pydantic v2 中 `Result` 作为响应模型需继承 `BaseModel`：

```python
from pydantic import BaseModel
from typing import Generic, TypeVar, Optional

T = TypeVar("T")

class Result(BaseModel, Generic[T]):
    code: int = 0
    message: str = "ok"
    data: Optional[T] = None

    @classmethod
    def ok(cls, data: T = None, message: str = "ok"):
        return cls(code=0, message=message, data=data)

    @classmethod
    def fail(cls, code: int, message: str):
        return cls(code=code, message=message, data=None)
```

### 第四步：app/schemas/user.py

```python
from pydantic import BaseModel, Field, field_validator

class UserCreate(BaseModel):
    username: str = Field(..., min_length=2, max_length=32)
    age: int = Field(..., ge=0, le=150)

    @field_validator("username")
    @classmethod
    def username_strip(cls, v: str) -> str:
        return v.strip()


class UserUpdate(BaseModel):
    username: str | None = Field(None, min_length=2, max_length=32)
    age: int | None = Field(None, ge=0, le=150)


class UserVO(BaseModel):
    id: int
    username: str
    age: int
```

### 第五步：app/services/user_service.py（内存存储）

```python
from app.schemas.user import UserCreate, UserUpdate, UserVO

_users: dict[int, dict] = {}
_next_id = 1


def list_users() -> list[UserVO]:
    return [UserVO(**u) for u in _users.values()]


def get_user(user_id: int) -> UserVO | None:
    raw = _users.get(user_id)
    return UserVO(**raw) if raw else None


def create_user(dto: UserCreate) -> UserVO:
    global _next_id
    user = {"id": _next_id, "username": dto.username, "age": dto.age}
    _users[_next_id] = user
    _next_id += 1
    return UserVO(**user)


def update_user(user_id: int, dto: UserUpdate) -> UserVO | None:
    raw = _users.get(user_id)
    if not raw:
        return None
    data = dto.model_dump(exclude_unset=True)
    raw.update(data)
    return UserVO(**raw)


def delete_user(user_id: int) -> bool:
    return _users.pop(user_id, None) is not None
```

### 第六步：app/routers/user.py

```python
from fastapi import APIRouter, HTTPException
from app.schemas.common import Result
from app.schemas.user import UserCreate, UserUpdate, UserVO
from app.services import user_service

router = APIRouter(prefix="/api/users", tags=["users"])


@router.get("", response_model=Result[list[UserVO]])
def list_users():
    return Result.ok(user_service.list_users())


@router.get("/{user_id}", response_model=Result[UserVO])
def get_user(user_id: int):
    user = user_service.get_user(user_id)
    if not user:
        raise HTTPException(status_code=404, detail="用户不存在")
    return Result.ok(user)


@router.post("", response_model=Result[UserVO], status_code=201)
def create_user(dto: UserCreate):
    return Result.ok(user_service.create_user(dto))


@router.put("/{user_id}", response_model=Result[UserVO])
def update_user(user_id: int, dto: UserUpdate):
    user = user_service.update_user(user_id, dto)
    if not user:
        raise HTTPException(status_code=404, detail="用户不存在")
    return Result.ok(user)


@router.delete("/{user_id}", response_model=Result[None])
def delete_user(user_id: int):
    if not user_service.delete_user(user_id):
        raise HTTPException(status_code=404, detail="用户不存在")
    return Result.ok(message="删除成功")
```

### 第七步：app/main.py

```python
from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from app.routers import user
from app.schemas.common import Result

app = FastAPI(title="demo-api", version="0.1.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:5173", "http://127.0.0.1:5173"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(user.router)


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    return JSONResponse(
        status_code=422,
        content=Result.fail(code=422, message=str(exc.errors())).model_dump(),
    )


@app.get("/health")
def health():
    return {"status": "ok"}
```

### 第八步：启动与验证

```powershell
# 在 demo-api 目录下
uvicorn app.main:app --reload --port 8080
```

```text
# 预期输出：
# INFO:     Uvicorn running on http://127.0.0.1:8080
# INFO:     Application startup complete.
```

浏览器打开：

- http://127.0.0.1:8080/docs — Swagger UI
- http://127.0.0.1:8080/health — `{"status":"ok"}`

**创建用户**（Swagger 或 curl）：

```powershell
curl -X POST "http://127.0.0.1:8080/api/users" `
  -H "Content-Type: application/json" `
  -d '{"username":"tom","age":18}'
```

```json
// 预期响应：
{"code":0,"message":"ok","data":{"id":1,"username":"tom","age":18}}
```

**列表**：

```powershell
curl "http://127.0.0.1:8080/api/users"
```

---

## 3. 路由与 HTTP 方法

| 方法 | 用途 | FastAPI |
|------|------|---------|
| GET | 查询 | `@router.get` |
| POST | 创建 | `@router.post` |
| PUT | 全量更新 | `@router.put` |
| PATCH | 部分更新 | `@router.patch` |
| DELETE | 删除 | `@router.delete` |

### 路径参数与查询参数

```python
@router.get("/search")
def search(keyword: str, page: int = 1, size: int = 10):
    # GET /api/users/search?keyword=tom&page=1&size=10
    return {"keyword": keyword, "page": page, "size": size}


@router.get("/{user_id}/orders/{order_id}")
def get_order(user_id: int, order_id: int):
    return {"user_id": user_id, "order_id": order_id}
```

### 请求体

```python
@router.post("")
def create(dto: UserCreate):  # JSON body 自动解析 + 校验
    ...
```

---

## 4. Pydantic 校验深入

### 4.1 为什么用 Pydantic

前端传来 JSON，字段可能缺失、类型错误、超长。Pydantic 在进业务逻辑前**自动校验**，失败返回 422 + 详细错误——比手写 `if` 可靠得多。

### 4.2 常用 Field 与 validator

```python
from pydantic import BaseModel, Field, EmailStr, field_validator

class RegisterDTO(BaseModel):
    username: str = Field(..., min_length=3, max_length=20)
    password: str = Field(..., min_length=6)
    email: EmailStr   # 需 pip install pydantic[email]

    @field_validator("username")
    @classmethod
    def no_space(cls, v: str) -> str:
        if " " in v:
            raise ValueError("用户名不能含空格")
        return v
```

### 4.3 model_dump 与 exclude_unset

```python
dto = UserUpdate(username="new_name")
dto.model_dump(exclude_unset=True)  # 只含传了的字段 {'username': 'new_name'}
```

更新接口必用，避免把未传字段覆盖成 `None`。

---

## 5. 依赖注入 Depends

```python
from fastapi import Depends, Header, HTTPException

def get_current_user_id(authorization: str = Header(default="")) -> int:
    if not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="未登录")
    token = authorization.removeprefix("Bearer ")
    # 04 章简化：假 token 就是 user_id
    try:
        return int(token)
    except ValueError:
        raise HTTPException(status_code=401, detail="Token 无效")


@router.get("/me")
def me(user_id: int = Depends(get_current_user_id)):
    return Result.ok({"user_id": user_id})
```

测试：`Authorization: Bearer 1`

依赖可复用：鉴权、数据库 Session（05 章）、分页参数。

---

## 6. 统一异常处理

```python
from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse

class BusinessError(Exception):
    def __init__(self, code: int, message: str):
        self.code = code
        self.message = message


@app.exception_handler(BusinessError)
async def business_error_handler(request, exc: BusinessError):
    return JSONResponse(
        status_code=400,
        content=Result.fail(code=exc.code, message=exc.message).model_dump(),
    )


# Service 里
def pay(order_id: int):
    if order_id <= 0:
        raise BusinessError(40001, "订单不存在")
```

与 Java `@ControllerAdvice` 同理。

---

## 7. CORS 与前后端联调

开发时前端 Vite 跑在 `5173`，后端在 `8080`，浏览器会拦跨域。`CORSMiddleware` 已在 main.py 配置。

生产推荐：**Nginx 同域反代** `/api` → 后端（09 章），前端 `VITE_API_BASE=/api`。

对齐 [Vue 08](../../前端学习/Vue/08-Axios网络请求与前后端联调.md) 的 axios 封装：

```typescript
// 前端期望的统一响应
interface Result<T> {
  code: number
  message: string
  data: T
}
```

后端 `Result` 字段名必须一致。

---

## 8. JWT 登录入门

```powershell
pip install python-jose[cryptography] passlib[bcrypt]
```

```python
# app/core/security.py
from datetime import datetime, timedelta, timezone
from jose import jwt

SECRET_KEY = "dev-only-change-in-production"
ALGORITHM = "HS256"

def create_access_token(subject: str, expires_minutes: int = 60) -> str:
    expire = datetime.now(timezone.utc) + timedelta(minutes=expires_minutes)
    return jwt.encode({"sub": subject, "exp": expire}, SECRET_KEY, algorithm=ALGORITHM)
```

```python
# 登录路由简化版
@router.post("/login")
def login(username: str, password: str):
    # 04 章：假校验
    if username != "admin" or password != "123456":
        raise HTTPException(status_code=401, detail="用户名或密码错误")
    token = create_access_token(subject=username)
    return Result.ok({"accessToken": token, "tokenType": "Bearer"})
```

完整 JWT + 密码哈希见 [10 章](./10-后端项目实战与面试准备.md) 挑战练习。

---

## 9. 配置管理 pydantic-settings

```python
# app/core/config.py
from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    app_name: str = "demo-api"
    debug: bool = True
    database_url: str = "mysql+pymysql://root:123456@localhost:3306/study_db"

    model_config = SettingsConfigDict(env_file=".env")


settings = Settings()
```

`.env` 文件：

```text
DATABASE_URL=mysql+pymysql://root:123456@localhost:3306/study_db
DEBUG=true
```

---

## 10. 异步路由

```python
import asyncio

@router.get("/async-demo")
async def async_demo():
    await asyncio.sleep(0.01)
    return Result.ok({"msg": "async route"})


@router.get("/sync-demo")
def sync_demo():
    return Result.ok({"msg": "sync route in threadpool"})
```

IO 密集（调 DB、Redis、HTTP）优先 `async def` + async 驱动（05～07 章）。

---

## 11. 项目结构演进

```mermaid
flowchart LR
    A[04 内存 list/dict] --> B[05 SQLAlchemy + MySQL]
    B --> C[07 Redis 缓存]
    C --> D[08 Celery 异步]
    D --> E[09 Docker 部署]
```

---

## 12. 常见报错与排查

| 报错 | 原因 | 解决 |
|------|------|------|
| `ModuleNotFoundError: No module named 'app'` | 启动目录不对 | 在 demo-api 根目录运行 uvicorn |
| `422 Unprocessable Entity` | 请求体不符合 Schema | 看响应 body 里 errors 详情 |
| `404 Not Found` | 路由路径或方法不对 | 对照 `/docs` 里的路径 |
| CORS 浏览器报错 | 未配 CORSMiddleware 或 origin 不对 | 检查 allow_origins 含前端地址 |
| `Address already in use` | 8080 被占用 | 换 `--port 8081` 或杀占用进程 |
| `ValidationError: Field required` | 缺少必填字段 | 补全 JSON 字段 |
| uvicorn 改了代码不生效 | 未开 `--reload` | 加 `--reload` 或手动重启 |
| `ImportError: email-validator` | 用了 EmailStr | `pip install pydantic[email]` |
| Swagger 打不开 | 防火墙/端口 | 确认 `127.0.0.1:8080/docs` |
| 前端收到 HTML 不是 JSON | 路径错打到 Vite | 检查 axios baseURL |

---

## 13. RESTful 设计规范

| 资源 | GET | POST | PUT | DELETE |
|------|-----|------|-----|--------|
| /api/users | 列表 | 创建 | - | - |
| /api/users/{id} | 详情 | - | 更新 | 删除 |

- 名词复数、小写、连字符：`/api/order-items`
- 状态码：200 成功、201 创建、400 业务错、401 未登录、404 不存在、422 参数校验失败、500 服务器错
- 与 [Java 04 §RESTful](../Java/04-SpringBoot核心开发.md) 保持一致，方便前端 Switch 后端

---

## 14. 练习建议

### 基础

1. 给 demo-api 加 `GET /api/users/count` 返回用户总数
2. 用户名重复时返回 400 BusinessError

### 进阶

3. 加 `Product` 资源 CRUD（内存版），字段：id、name、price（Decimal 字符串）
4. 实现 `Depends` 分页：`page`、`size`，返回 `PageResult`

### 挑战

5. JWT 登录 + 受保护路由 `/api/users/me`
6. 对接 [Vue 08](../../前端学习/Vue/08-Axios网络请求与前后端联调.md) 替换 mock 数据

---

## 15. 参考答案（分页 Depends）

```python
from fastapi import Query

class PageParams:
    def __init__(self, page: int = Query(1, ge=1), size: int = Query(10, ge=1, le=100)):
        self.page = page
        self.size = size
        self.offset = (page - 1) * size


@router.get("/page")
def list_page(p: PageParams = Depends()):
    all_users = user_service.list_users()
    total = len(all_users)
    items = all_users[p.offset : p.offset + p.size]
    return Result.ok({"items": items, "total": total, "page": p.page, "size": p.size})
```

---

## 16. 学完标准

- [ ] 独立创建 demo-api 并用 uvicorn 启动
- [ ] 能写 Router + Service + Schema 分层
- [ ] 会用 Pydantic 校验请求体
- [ ] 配置 CORS，Swagger 能调通 CRUD
- [ ] 理解 Depends、统一 Result、全局异常
- [ ] 知道 JWT 登录基本流程

---

## 下一章预告

04 章接口数据在内存 dict 里——重启就丢。下一章（05 SQLAlchemy 事务与接口工程化）把内存换成 **MySQL**：ORM 模型、Session 管理、CRUD、事务、分页——demo-api 第一次真正持久化。

---

*下一章：05 SQLAlchemy 事务与接口工程化*
