# Python 内置类型、模块与类型注解

## 本章与上一章的关系

01 章你学会了 Python 语法和 OOP——能写类和方法。但真实后端代码里，**80% 的时间在处理 list、dict、字符串、模块拆分和类型注解**，而不是反复写 `class` 语法。

这一章就是「Python 日常开发工具箱」：列表推导式、dict 统计、Decimal 金额、`datetime`、包结构、`typing` 与 Pydantic 前置知识。02 章练熟后，03 章学 asyncio 时你会理解「为什么 dict 不是线程安全的」，04 章 FastAPI 接 JSON 返回 `list[UserSchema]` 也不会陌生。

---

## 1. 为什么这部分很重要

后端项目里几乎每条接口路径都会用到：

- `list` / `dict` 组织数据
- 字符串解析与拼接
- `Decimal` 处理金额
- `datetime` 记录时间
- 模块拆分项目结构
- 类型注解让 IDE 和 FastAPI 帮你查错

---

## 2. list 深入

### 2.1 创建与常用方法

```python
nums = [1, 2, 3]
nums.append(4)           # 末尾追加
nums.insert(0, 0)        # 指定位置插入
nums.extend([5, 6])      # 合并另一个 list
nums.remove(3)           # 删除第一个值为 3 的元素
popped = nums.pop()      # 弹出末尾
print(nums.count(2))     # 计数
print(len(nums))
```

### 2.2 切片

```python
data = [0, 1, 2, 3, 4, 5]
print(data[1:4])    # [1, 2, 3]
print(data[:3])     # [0, 1, 2]
print(data[::2])    # [0, 2, 4]  步长 2
print(data[::-1])   # 反转
```

### 2.3 列表推导式

```python
squares = [x * x for x in range(1, 6)]
# [1, 4, 9, 16, 25]

evens = [x for x in range(20) if x % 2 == 0]

users = [{"id": 1, "name": "Tom"}, {"id": 2, "name": "Jerry"}]
names = [u["name"] for u in users]
```

后端常见：从 ORM 查询结果 `[User(...), ...]` 转成 `[UserSchema(...), ...]`。

### 2.4 深入：可变默认参数陷阱

```python
# 错误写法
def add_item(item, bucket=[]):
    bucket.append(item)
    return bucket

print(add_item(1))  # [1]
print(add_item(2))  # [1, 2]  ← 意外！共享同一个 list

# 正确写法
def add_item(item, bucket=None):
    if bucket is None:
        bucket = []
    bucket.append(item)
    return bucket
```

FastAPI 依赖注入里也要注意：不要用可变对象做默认参数。

---

## 3. dict 深入

### 3.1 创建与访问

```python
user = {"id": 1, "username": "tom", "age": 18}
user["email"] = "tom@example.com"
print(user.get("phone"))           # None
print(user.get("phone", "未设置"))  # 未设置

for key, value in user.items():
    print(key, value)
```

### 3.2 dict 推导式

```python
words = ["apple", "banana", "apple", "cherry"]
count = {}
for w in words:
    count[w] = count.get(w, 0) + 1
# {'apple': 2, 'banana': 1, 'cherry': 1}

# 推导式写法
from collections import Counter
count2 = dict(Counter(words))
```

### 3.3 defaultdict 与 Counter

```python
from collections import defaultdict, Counter

# defaultdict：键不存在时自动创建默认值
groups = defaultdict(list)
groups["admin"].append("tom")
groups["admin"].append("jerry")

# Counter：计数神器
text = "hello world"
c = Counter(text)
print(c.most_common(3))  # [('l', 3), ('o', 2), ('h', 1)]
```

### 3.4 合并 dict（3.9+）

```python
defaults = {"role": "user", "status": 1}
override = {"status": 0}
merged = defaults | override   # {'role': 'user', 'status': 0}
```

---

## 4. set 与 tuple

### 4.1 set：去重与集合运算

```python
tags = {"python", "backend", "python"}
print(tags)  # {'python', 'backend'}  自动去重

a = {1, 2, 3}
b = {3, 4, 5}
print(a & b)   # 交集 {3}
print(a | b)   # 并集 {1, 2, 3, 4, 5}
print(a - b)   # 差集 {1, 2}
```

### 4.2 tuple：不可变序列

```python
point = (10, 20)
x, y = point   # 解包

# 单元素 tuple 必须加逗号
t = (42,)
```

适合：函数返回多个值、dict 的 key（list 不能当 key）。

---

## 5. 字符串进阶

```python
s = "  hello@example.com  "
print(s.strip())
print(s.split("@"))           # ['  hello', 'example.com  ']
print(",".join(["a", "b", "c"]))  # a,b,c

# f-string 格式化
price = 99.5
print(f"价格：{price:.2f} 元")   # 价格：99.50 元

# removeprefix / removesuffix (3.9+)
url = "https://api.example.com/users"
print(url.removeprefix("https://"))
```

### 正则入门（后端校验）

```python
import re

phone = "13800138000"
if re.fullmatch(r"1[3-9]\d{9}", phone):
    print("手机号合法")

email = "tom@example.com"
if re.fullmatch(r"[\w.-]+@[\w.-]+\.\w+", email):
    print("邮箱格式 OK")
```

Pydantic 有更优雅的 `EmailStr` 校验（04 章），但理解正则仍有帮助。

---

## 6. Decimal 与金额

```python
from decimal import Decimal, ROUND_HALF_UP

price = Decimal("99.90")
count = Decimal("2")
total = price * count
print(total)  # 99.90 * 2 = 199.80  精确

# 四舍五入到 2 位
amount = Decimal("10.125")
print(amount.quantize(Decimal("0.01"), rounding=ROUND_HALF_UP))  # 10.13
```

**规则**：金额在 Python 用 `Decimal`，在 MySQL 用 `DECIMAL(10,2)`，**永远不要 `float` 做金额运算**。

---

## 7. datetime 与时间处理

```python
from datetime import datetime, date, timedelta, timezone

now = datetime.now()
today = date.today()
print(now.strftime("%Y-%m-%d %H:%M:%S"))

# 解析字符串
dt = datetime.strptime("2026-06-18 10:30:00", "%Y-%m-%d %H:%M:%S")

#  timedelta
tomorrow = today + timedelta(days=1)
week_ago = now - timedelta(days=7)

# UTC（后端存储推荐 UTC，展示再转本地时区）
utc_now = datetime.now(timezone.utc)
```

### 深入：为什么后端时间存 UTC？

服务器可能部署在不同时区，用户也在全球各地。统一存 UTC + 展示层转本地，避免「夏令时」「跨日统计错位」等问题。MySQL 用 `DATETIME` 或 `TIMESTAMP`，Java 用 `Instant`，Python 用带时区的 `datetime`。

---

## 8. 模块、包与项目结构

### 8.1 模块 import

```python
# utils/math_helper.py
def add(a: int, b: int) -> int:
    return a + b

# main.py
from utils.math_helper import add
# 或
import utils.math_helper as mh
```

### 8.2 包结构

```text
demo_pkg/
├── app/
│   ├── __init__.py      ← 标记 app 为包（可为空）
│   ├── main.py
│   ├── routers/
│   │   ├── __init__.py
│   │   └── user.py
│   └── services/
│       ├── __init__.py
│       └── user_service.py
├── tests/
└── requirements.txt
```

### 8.3 `if __name__ == "__main__"`

```python
# user_service.py
def get_user(user_id: int) -> dict:
    return {"id": user_id, "name": "Tom"}


if __name__ == "__main__":
    # 仅直接运行此文件时执行，被 import 时不执行
    print(get_user(1))
```

### 8.4 相对导入（包内）

```python
# app/routers/user.py
from ..services.user_service import get_user   # 上一级包
from . import common                            # 同包
```

---

## 9. 类型注解（typing）

Python 3.9+ 推荐用内置泛型写法；旧项目可能仍见 `typing.List`。

### 9.1 基础注解

```python
def greet(name: str) -> str:
    return f"Hello, {name}"


age: int = 18
scores: list[int] = [90, 85, 88]
user_map: dict[str, int] = {"tom": 1, "jerry": 2}
maybe_name: str | None = None   # 3.10+ 联合类型
```

### 9.2 TypedDict 与 dataclass

```python
from typing import TypedDict

class UserDict(TypedDict):
    id: int
    username: str
    age: int


from dataclasses import dataclass

@dataclass
class User:
    id: int
    username: str
    age: int

    def is_adult(self) -> bool:
        return self.age >= 18
```

`dataclass` 少写样板代码；FastAPI 更常用 **Pydantic BaseModel**（04 章）。

### 9.3 Optional 与 Union

```python
from typing import Optional

def find_user(user_id: int) -> Optional[dict]:
    if user_id <= 0:
        return None
    return {"id": user_id, "name": "Tom"}
```

### 9.4 泛型

```python
from typing import TypeVar, Generic

T = TypeVar("T")

class PageResult(Generic[T]):
    def __init__(self, items: list[T], total: int, page: int, size: int):
        self.items = items
        self.total = total
        self.page = page
        self.size = size
```

对应前端 TypeScript 的 `PageResult<T>` 与 Java 的 `Result<T>`。

### 9.5 Protocol（结构化类型）

```python
from typing import Protocol

class Payable(Protocol):
    def pay(self, amount: Decimal) -> bool: ...


def checkout(obj: Payable, amount: Decimal) -> bool:
    return obj.pay(amount)
```

类似 Java 的 interface，但 Python 是**鸭子类型**：只要对象有 `pay` 方法即可，不必显式继承。

---

## 10. 常用标准库速查

| 模块 | 用途 | 示例 |
|------|------|------|
| `json` | JSON 序列化 | `json.loads(s)` / `json.dumps(obj)` |
| `pathlib` | 路径操作 | `Path("data/a.txt").read_text()` |
| `os` / `sys` | 环境、退出码 | `os.getenv("DB_URL")` |
| `logging` | 日志 | `logging.info("user login")` |
| `enum` | 枚举 | `class Status(str, Enum): ACTIVE = "active"` |
| `functools` | 装饰器工具 | `@lru_cache` |
| `itertools` | 迭代工具 | `chain`, `groupby` |
| `copy` | 深拷贝 | `copy.deepcopy(d)` |

### json 示例

```python
import json

data = {"code": 0, "message": "ok", "data": {"id": 1}}
text = json.dumps(data, ensure_ascii=False)
print(text)
parsed = json.loads(text)
```

FastAPI 自动做 JSON 序列化，但脚本、测试、Celery 任务里常手动用 `json`。

### logging 示例

```python
import logging

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)

logger.info("服务启动")
logger.error("数据库连接失败", exc_info=True)
```

---

## 11. 枚举 Enum

```python
from enum import Enum

class OrderStatus(str, Enum):
    CREATED = "CREATED"
    PAID = "PAID"
    SHIPPED = "SHIPPED"
    CANCELLED = "CANCELLED"


def can_cancel(status: OrderStatus) -> bool:
    return status in (OrderStatus.CREATED, OrderStatus.PAID)


print(OrderStatus.PAID.value)  # PAID
```

继承 `str` 后，JSON 序列化直接输出字符串值，与 FastAPI 配合友好。

---

## 12. 装饰器入门

```python
import functools
import time

def timer(func):
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        start = time.perf_counter()
        result = func(*args, **kwargs)
        elapsed = time.perf_counter() - start
        print(f"{func.__name__} 耗时 {elapsed:.4f}s")
        return result
    return wrapper


@timer
def fetch_users():
    time.sleep(0.1)
    return [{"id": 1}]


fetch_users()
```

FastAPI 的路由 `@app.get("/users")`、依赖 `@Depends()` 底层都是装饰器思想。

---

## 13. 数据流：从数据库行到 API JSON

```mermaid
flowchart LR
    subgraph db [数据库]
        Row["MySQL 行"]
    end
    subgraph py [Python 层]
        ORM["SQLAlchemy Model"]
        Schema["Pydantic Schema"]
    end
    subgraph api [接口]
        JSON["JSON 响应"]
    end
    Row --> ORM
    ORM --> Schema
    Schema --> JSON
```

02 章你要建立中间两层的数据结构直觉：`dict` / `dataclass` / 未来的 `BaseModel`。

---

## 14. 手把手：拆分词频统计 mini 项目

### 第一步：创建目录

```powershell
mkdir word-count
cd word-count
python -m venv .venv
.\.venv\Scripts\Activate.ps1
```

### 第二步：项目结构

```text
word-count/
├── app/
│   ├── __init__.py
│   ├── counter.py
│   └── main.py
└── sample.txt
```

### 第三步：counter.py

```python
from collections import Counter
from pathlib import Path

def count_words(text: str) -> dict[str, int]:
    words = text.lower().split()
    return dict(Counter(words))


def count_file(path: str | Path) -> dict[str, int]:
    content = Path(path).read_text(encoding="utf-8")
    return count_words(content)
```

### 第四步：main.py

```python
from app.counter import count_file

if __name__ == "__main__":
    result = count_file("sample.txt")
    top5 = sorted(result.items(), key=lambda x: x[1], reverse=True)[:5]
    for word, cnt in top5:
        print(f"{word}: {cnt}")
```

### 第五步：sample.txt + 运行

```powershell
echo "python is great python backend python" > sample.txt
python -m app.main
# 预期输出类似：
# python: 3
# is: 1
# great: 1
# ...
```

---

## 15. 常见报错与排查

| 报错信息 | 可能原因 | 解决方案 |
|---------|---------|---------|
| `KeyError: 'xxx'` | dict 键不存在 | 用 `.get()` 或先 `if key in d` |
| `IndexError: list index out of range` | 下标越界 | 检查 `len(list)` |
| `TypeError: unhashable type: 'list'` | list 当 dict key 或 set 元素 | 改用 tuple 或 str |
| `ImportError: attempted relative import with no known parent package` | 直接运行含相对导入的文件 | 用 `python -m app.main` 从包根运行 |
| `ModuleNotFoundError: No module named 'app'` | 工作目录不对 | cd 到项目根；或设置 `PYTHONPATH` |
| `TypeError: 'NoneType' object is not subscriptable` | 函数返回 None 却当 dict 用 | 检查返回值 |
| `JSONDecodeError` | 字符串不是合法 JSON | 打印原始内容；校验接口响应 |
| `decimal.InvalidOperation` | Decimal 构造参数非法 | 用字符串构造 `Decimal("10.5")` |
| `AttributeError: 'datetime.datetime' object has no attribute 'xxx'` | 混淆 date 与 datetime | 查文档确认类型 |
| `Mutable default argument` 类 bug | 默认参数用 `[]` / `{}` | 默认 `None`，函数内创建 |

---

## 16. 练习建议

### 基础

1. 用 dict 统计字符串中每个字符出现次数
2. 用列表推导式生成 1～100 内能被 3 整除的数
3. 写 `format_money(amount: Decimal) -> str` 输出 `¥99.90`

### 进阶

4. 把 01 章 Student 练习拆成 `models/student.py` + `main.py` 两个模块
5. 用 `dataclass` 重写 User，加 `to_dict()` 方法
6. 实现简易 LRU：`get(key)` / `put(key, value)`，容量满时淘汰最久未用

### 挑战

7. 读 CSV 文件（可用标准库 `csv`），按类别聚合统计金额（用 Decimal）
8. 为 `PageResult[T]` 写类型注解 + 单元测试

---

## 17. 分级练习参考答案

### 基础：字符统计

```python
def char_count(text: str) -> dict[str, int]:
    result: dict[str, int] = {}
    for ch in text:
        if ch.isspace():
            continue
        result[ch] = result.get(ch, 0) + 1
    return result
```

### 进阶：LRU 缓存（OrderedDict）

```python
from collections import OrderedDict

class LRUCache:
    def __init__(self, capacity: int):
        self.capacity = capacity
        self.cache: OrderedDict = OrderedDict()

    def get(self, key: str):
        if key not in self.cache:
            return None
        self.cache.move_to_end(key)
        return self.cache[key]

    def put(self, key: str, value):
        if key in self.cache:
            self.cache.move_to_end(key)
        self.cache[key] = value
        if len(self.cache) > self.capacity:
            self.cache.popitem(last=False)
```

---

## 18. 学完标准

- [ ] 熟练使用 list/dict 推导式、Counter、defaultdict
- [ ] 金额用 Decimal，时间用 datetime
- [ ] 能拆分模块与包，理解相对导入
- [ ] 能写函数类型注解 `def f(x: int) -> str`
- [ ] 理解 dataclass、Enum、装饰器基础
- [ ] 独立完成词频统计 mini 项目

---

## 下一章预告

02 章掌握了数据结构和模块组织——代码还是**同步顺序执行**的。真实 Web 服务要同时处理成千上万个请求，FastAPI 的异步路由依赖 **asyncio**。

下一章（03 Python 并发编程与 asyncio）讲 GIL、线程 vs 协程、`async/await`、`asyncio.gather`，以及「什么时候该用 async、什么时候该用 Celery」——这是进入 FastAPI 前的最后一关语言课。

---

*下一章：03 Python 并发编程与 asyncio*
