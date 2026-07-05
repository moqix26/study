# MySQL 基础、索引与事务

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0、FAQ、闭卷自测、费曼检验 -->

## 0. 读前导读（零基础也能跟上）

> **读者假设**：你已会用 MyBatis 写 CRUD（05 章），但 SQL 慢、金额精度、事务隔离还似懂非懂。

### 0.1 用一句话弄懂本章

**一句话**：**MySQL** 是存在硬盘上的**超级 Excel**——表 = 工作表，行 = 一条记录，列 = 字段；**索引**像 Excel 的筛选/排序加速，**事务**保证多步操作「要么全成功要么全撤销」。

**生活类比——MySQL = Excel 升级版**：

| 概念 | SQL / MySQL | Excel 类比 |
|------|-------------|------------|
| **数据库** | `study_db` | 整个 **工作簿文件** |
| **表** | `user`、`order` | 工作簿里的 **Sheet** |
| **行** | 一条 `INSERT` 记录 | **一行数据**（张三、18 岁） |
| **列** | `username`、`age` | **列标题** |
| **主键** | `PRIMARY KEY id` | **行号/唯一编号**，不能重复 |
| **索引** | `CREATE INDEX` | **按某列排序的速查目录**，不用翻全表 |
| **事务** | `BEGIN` … `COMMIT` | **一批操作绑在一起**：要么全保存要么全撤销 |
| **JOIN** | `LEFT JOIN` | **VLOOKUP**：订单表关联用户表查姓名 |
| **DECIMAL** | `DECIMAL(10,2)` | 金额列设「数值格式」避免 0.1+0.2 怪数 |

**为什么重要**：05 章写 SQL，本章懂 SQL **怎么被执行**；慢查询、锁表、金额精度线上事故都源于这里。

**本章用到的地方**：§4.1 Docker、§11 索引、§15 事务、§47 EXPLAIN。

---

### 0.2 你需要提前知道什么（真不会就先跳到哪一章）

| 你现在的水平 | 建议动作 |
|--------------|----------|
| 不会 MyBatis CRUD | 先学 [05 MyBatis](./05-MyBatis事务与接口工程化.md) §7.1 |
| 没装 MySQL | 从 §4.1 Docker 一键启动 |
| 只会 `SELECT *` | 本章重点补索引、EXPLAIN、字段类型 |
| 想优化项目慢接口 | 精读 §11～§12、§47～§48、§55 |

**最低门槛**：会执行 `INSERT/SELECT/UPDATE`；知道表、行、列；05 章已建过 `user` 表。

---

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

- [ ] 用 Docker 启动 MySQL 8.0 并 `docker exec` 进入
- [ ] 建 user/product/order 三表，金额用 `DECIMAL(10,2)`
- [ ] 写 `WHERE`、`ORDER BY`、`LIMIT` 分页、`GROUP BY` 统计
- [ ] 写 `LEFT JOIN` 联表查询用户订单
- [ ] 用「Excel 速查目录」理解索引；建联合索引
- [ ] 说出最左前缀原则并举例
- [ ] 用 `EXPLAIN` 看 `type`、`key`、`rows`
- [ ] 口述 ACID、四种隔离级别、脏读/幻读
- [ ] 写防超卖：`UPDATE ... SET stock = stock - ? WHERE stock >= ?`
- [ ] 闭卷自测 10 题正确 ≥ 8 题

---

### 0.4 建议学习时长与节奏

| 阶段 | 建议时间 | 做什么 |
|------|----------|--------|
| §0 + §4.1 Docker | 1 小时 | 起环境、建库 |
| §5～§9 SQL 基础 | 3 小时 | CRUD、条件、分页、JOIN |
| §10～§12 表设计与索引 | 3 小时 | 三范式入门、B+ 树、联合索引 |
| §15～§17 事务与锁 | 2 小时 | ACID、隔离级别、脏读幻读 |
| §47～§55 EXPLAIN 实战 | 2 小时 | 慢 SQL 优化 |
| 分级练习 + 自测 | 2 小时 | 有索引 vs 无索引对比 rows |

---

### 0.5 学完本章你能做什么（可验证的具体动作）

1. **Docker** 启动 MySQL，`SHOW DATABASES` 见 `study_db`。
2. **建表** 含 `DECIMAL` 金额、`UNIQUE` 订单号、联合索引 `idx_user_status_time`。
3. **EXPLAIN** 用户订单列表 SQL，`type` 从 `ALL` 优化到 `ref`。
4. **解释** 为什么 `double` 存金额可能对账差一分钱。
5. **写 SQL** 扣库存且不允许超卖（`WHERE stock >= ?`）。

---

### 0.6 手把手总览：Docker 起 MySQL + 建三表

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | `docker run ... mysql:8.0`（§4.1） | `docker ps` 显示 Up | 3306 被占用则改 3307 |
| 2 | `docker exec -it study-mysql mysql -uroot -p` | 进入 `mysql>` | 密码与 `-e` 一致 |
| 3 | `USE study_db;` 执行 §46 建表 SQL | `SHOW TABLES` 三张表 | 保留字 `` `order` `` 加反引号 |
| 4 | 插入 §52 测试数据 | `COUNT(*)` 符合预期 | 查字段名是否拼错 |
| 5 | `EXPLAIN` 用户订单查询 | 有索引时 `key` 非 NULL | 无索引先记 `type=ALL` |
| 6 | 加联合索引后再 EXPLAIN | `rows` 明显下降 | 对照 §54 最左前缀 |

---

## 本章与上一章的关系

05 章你学会了用 MyBatis 写 SQL、连 MySQL——但有个尴尬的情况：SQL 写对了，查询却慢到超时；或者金额字段用了 `double`，用户余额莫名其妙少了 0.01 元。这些问题根源都在 **MySQL 本身**：表怎么设计、索引怎么建、事务怎么隔离。

这一章从数据库底层补全认知。你会用 Docker 快速起一个 MySQL 环境（不用再折腾本地安装），学会设计电商三表、看懂 EXPLAIN 执行计划、理解 B+ 树索引为什么能让查询快 100 倍。05 章是"怎么写 SQL"，这一章是"SQL 在数据库里怎么跑"。

---

## 1. MySQL 在后端里的位置

MySQL 是 Java 后端最重要的数据库之一。

你做的很多核心业务最终都要落到 MySQL：

- 用户数据
- 订单数据
- 商品数据
- 支付记录
- 库存信息

所以你学习 MySQL，不能只停留在“会写 `select * from user`”，而要逐步理解：

- 表怎么设计
- SQL 怎么写
- 索引怎么建
- 事务怎么保证一致性

## 2. 数据库、表、行、列

你先把这几个概念吃透：

- 数据库：一组相关数据的集合
- 表：存储某类数据的结构
- 行：一条记录
- 列：一个字段

例如用户表可能有：

- `id`
- `username`
- `phone`
- `status`
- `create_time`

## 3. 建表基础

```sql
create table user (
    id bigint primary key auto_increment,
    username varchar(64) not null,
    phone varchar(20),
    age int,
    status tinyint not null default 1,
    create_time datetime not null default current_timestamp,
    update_time datetime not null default current_timestamp on update current_timestamp
);
```

### 字段说明

- `bigint`：适合主键
- `varchar`：适合变长字符串
- `tinyint`：适合状态值
- `datetime`：适合记录时间

## 4. 常见数据类型怎么选

### 4.1 整数

- `tinyint`
- `int`
- `bigint`

主键常用 `bigint`，因为扩展余地大。

### 4.2 字符串

- `char`
- `varchar`
- `text`

一般业务字段优先考虑 `varchar`。

### 4.3 金额

用：

- `decimal(10,2)`

不要用：

- `float`
- `double`

<!-- 修改说明: 补充 BigDecimal/decimal 的深入解释与真实案例 -->

### 为什么金额用 DECIMAL，而不用 double/float？

**结论**：`float` 和 `double` 是二进制浮点数，很多十进制小数（如 0.1）无法精确表示，累加后会出现精度丢失。

**底层原理**：

计算机用二进制存储浮点数。十进制的 `0.1` 在二进制里是无限循环小数，存储时被截断。单次误差很小，但金融场景里"加加减减"累积起来，就会出现 `0.1 + 0.2 != 0.3` 的经典问题：

```java
System.out.println(0.1 + 0.2);  // 输出 0.30000000000000004
```

MySQL 的 `DECIMAL(10,2)` 是以 **字符串方式存储精确十进制数**，做加减乘除按十进制规则计算，不会有二进制截断误差。

**真实案例（模拟）**：

某支付系统早期用 `DOUBLE` 存用户余额。用户 A 连续充值 0.1 元 10 次，系统显示余额 0.9999999999999999 元；用户 B 提现 100.00 元，实际扣了 99.99999999999999 元，财务对账每天差几分钱，月底对不上账，排查两周才发现是字段类型问题。迁移到 `DECIMAL(10,2)` 后问题解决。

**Java 侧对应**：数据库用 `DECIMAL`，Java 代码用 `BigDecimal`，不要用 `double` 做金额运算。

---

## 4.1 手把手：Docker 启动 MySQL

05 章项目需要 MySQL，这里教你用 Docker 一键启动，比本地安装省心。

### 前提

电脑已安装 [Docker Desktop](https://www.docker.com/products/docker-desktop/)。

### 启动命令

```bash
docker run -d \
  --name study-mysql \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=123456 \
  -e MYSQL_DATABASE=study_db \
  mysql:8.0
```

Windows PowerShell 单行写法：

```powershell
docker run -d --name study-mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=123456 -e MYSQL_DATABASE=study_db mysql:8.0
```

```bash
# 预期输出（一行容器 ID）：
# a1b2c3d4e5f6...
```

### 验证 MySQL 是否启动成功

```bash
docker ps
# 预期输出：
# CONTAINER ID   IMAGE       STATUS          PORTS                    NAMES
# a1b2c3d4e5f6   mysql:8.0   Up 30 seconds   0.0.0.0:3306->3306/tcp   study-mysql
```

### 进入 MySQL 执行 SQL

```bash
docker exec -it study-mysql mysql -uroot -p123456
# 预期输出：
# mysql>

# 在 mysql> 提示符下：
mysql> SHOW DATABASES;
# 预期输出包含 study_db

mysql> USE study_db;
mysql> SHOW TABLES;
# 预期输出：Empty set（刚建库还没有表）
```

### 导入 05/06 章建表 SQL

把下面 SQL 保存为 `init.sql`，然后：

```bash
docker exec -i study-mysql mysql -uroot -p123456 study_db < init.sql
# 预期输出：无报错

docker exec -it study-mysql mysql -uroot -p123456 -e "USE study_db; SHOW TABLES;"
# 预期输出：
# +--------------------+
# | Tables_in_study_db |
# +--------------------+
# | user               |
# | product            |
# | order              |
# +--------------------+
```

### 常见 Docker 启动失败

```bash
docker run ...
# 失败输出示例：
# docker: Error response from daemon: driver failed programming external connectivity
# Bind for 0.0.0.0:3306 failed: port is already allocated.
```

说明 3306 端口被占用（可能本地已装 MySQL）。解决：停掉本地 MySQL 服务，或把 `-p 3306:3306` 改成 `-p 3307:3306`，同时改 `application.yml` 里的端口。

---

## 5. 基础 CRUD

### 5.1 插入

```sql
insert into user(username, phone, age)
values ('zhangsan', '13800000000', 18);
```

### 5.2 查询

```sql
select id, username, age
from user
where age >= 18;
```

### 5.3 更新

```sql
update user
set age = 20
where id = 1;
```

### 5.4 删除

```sql
delete from user
where id = 1;
```

## 6. 条件查询

```sql
select *
from user
where status = 1
  and age >= 18;
```

### 常见条件

- `=`
- `>`
- `<`
- `>=`
- `<=`
- `in`
- `between`
- `like`
- `is null`

## 7. 排序与分页

### 7.1 排序

```sql
select *
from user
order by create_time desc;
```

### 7.2 分页

```sql
select *
from user
limit 0, 10;
```

其中：

- `0` 是偏移量
- `10` 是取多少条

## 8. 聚合查询

### 8.1 count

```sql
select count(*)
from user;
```

### 8.2 分组

```sql
select status, count(*)
from user
group by status;
```

### 8.3 having

```sql
select status, count(*) as total
from user
group by status
having total > 5;
```

## 9. 多表查询

比如订单表和用户表联查：

```sql
select o.id, o.total_amount, u.username
from orders o
left join user u on o.user_id = u.id;
```

### join 的核心理解

- `inner join`：两边都匹配才返回
- `left join`：左边都返回，右边没有就补 `null`

## 10. 表设计思路

### 10.1 用户表

核心字段通常有：

- 用户 ID
- 用户名
- 手机号
- 密码
- 状态
- 创建时间

### 10.2 商品表

核心字段通常有：

- 商品 ID
- 标题
- 价格
- 库存
- 状态

### 10.3 订单表

核心字段通常有：

- 订单 ID
- 用户 ID
- 订单金额
- 订单状态
- 支付状态
- 创建时间

### 10.4 表设计原则

- 一张表描述一个核心实体
- 字段命名统一
- 状态字段明确
- 不要过早过度设计

## 11. 索引

### 11.1 为什么需要索引

索引是帮助数据库快速定位数据的数据结构。

没有索引时，数据库可能需要全表扫描。

### 11.2 常见索引类型

- 主键索引
- 唯一索引
- 普通索引
- 联合索引

### 11.3 建索引示例

```sql
create index idx_user_phone on user(phone);
```

联合索引：

```sql
create index idx_user_status_age on user(status, age);
```

## 12. B+ 树的基础理解

MySQL InnoDB 的常见索引底层是 B+ 树。

<!-- 修改说明: 新增 B+ 树 Mermaid 结构简图 -->

### B+ 树三层结构简图

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

你现在先理解这几个优点：

- 适合磁盘 IO：每个节点大小约 16KB，一次 IO 读一整页
- 层级低：百万级数据通常 3~4 层，最多 3~4 次磁盘 IO
- 支持范围查询：叶子节点通过链表相连，`WHERE age BETWEEN 18 AND 25` 顺着链表扫即可

对比 B 树：B+ 树非叶子节点只存索引键不存数据，单页能放更多键，树更矮，IO 更少。

---

如果你建了索引：

```sql
create index idx_user_status_age on user(status, age);
```

那么这些查询更容易用到索引：

```sql
where status = 1
where status = 1 and age = 18
```

但如果只按 `age` 查，可能就用不上这个联合索引。

这就是最左前缀原则的基础理解。

## 14. 覆盖索引和回表

### 覆盖索引

查询所需字段都在索引里，不需要再去主表取数据。

### 回表

先通过二级索引找到主键，再根据主键回主表取完整数据。

面试里经常会问这两个概念。

## 15. 事务

### 15.1 什么是事务

事务是一组操作，要么都成功，要么都失败。

例如下单：

1. 写订单
2. 扣库存
3. 扣余额

这三个动作要尽量保持一致。

### 15.2 ACID

- 原子性
- 一致性
- 隔离性
- 持久性

这是事务最基础的四个特性。

## 16. 隔离级别

常见四种：

- 读未提交
- 读已提交
- 可重复读
- 串行化

MySQL InnoDB 默认一般是：

- 可重复读

## 17. 并发读问题

### 17.1 脏读

读到了别人还没提交的数据。

### 17.2 不可重复读

同一事务里，两次读取同一行数据结果不同。

### 17.3 幻读

同一事务里，两次范围查询返回的记录数不同。

## 18. 锁

### 18.1 行锁

锁住某一行记录。

### 18.2 表锁

锁住整张表。

### 18.3 共享锁和排他锁

- 共享锁偏读
- 排他锁偏写

## 19. `explain`

这是分析 SQL 性能的重要工具。

```sql
explain select * from user where phone = '13800000000';
```

你至少要学会观察：

- 是否用了索引
- 扫描行数多不多
- 查询类型好不好

## 20. 慢 SQL 优化思路

可以按这个顺序排查：

1. SQL 写法是否合理
2. 是否缺索引
3. 索引是否失效
4. 是否查了不必要的字段
5. 是否联表过多

## 21. 初学者常见错误

### 21.1 到处 `select *`

这会增加不必要的数据传输。

### 21.2 没有 where 就 update/delete

非常危险。

### 21.3 给每个字段都建索引

索引不是越多越好，写入也有成本。

### 21.4 一个表堆太多不相关字段

会影响维护和扩展。

## 22. 这一章练习建议

你最好自己建表并练这些内容：

1. 用户表
2. 商品表
3. 订单表
4. 分页查询
5. 条件查询
6. 分组统计
7. 联表查询
8. 加索引并看 `explain`

## 23. 学完标准

如果你能做到下面这些，就说明这一章已经比较扎实：

- 能自己设计基础业务表
- 能写常见 SQL
- 知道索引为什么重要
- 知道事务和隔离级别的基础含义
- 能分析简单慢 SQL 的问题方向

## 24. 数据库设计范式基础认知

你会在面试里经常听到：

- 第一范式
- 第二范式
- 第三范式

你现在先把它理解成：

- 避免字段设计混乱
- 避免不必要的数据冗余
- 提高数据一致性

但同时要知道：

- 真正业务设计不是死背范式
- 有时会为了性能做适度反规范化

## 25. 唯一约束和普通索引

### 唯一约束

保证字段值不能重复。

适合：

- 手机号
- 用户名
- 订单号

### 普通索引

主要为了加速查询。

## 26. 索引失效常见场景

这是面试高频。

常见情况包括：

- 对索引列做函数操作
- 对索引列做运算
- 联合索引不满足最左前缀
- 模糊查询前面加 `%`
- 类型隐式转换

## 27. MVCC 的基础理解

MVCC 是多版本并发控制。

它的核心价值是：

- 提升并发读写性能
- 让部分读操作不用直接加重锁

你现在先不必深挖实现细节，但要知道它和事务隔离有关系。

## 28. 死锁基础认知

事务并发时可能发生死锁。

常见原因：

- 两个事务访问资源顺序不一致

避免思路：

- 固定访问顺序
- 缩短事务
- 尽快提交

## 29. 大分页问题

当你这样查：

```sql
select * from user limit 100000, 10;
```

性能可能会变差。

因为数据库要先跳过大量数据。

基础优化方向：

- 通过主键范围分页
- 记录上次最后一条 ID

## 30. 慢查询日志基础认知

MySQL 支持慢查询日志。

它的价值是：

- 帮你发现执行慢的 SQL
- 帮助定位性能瓶颈

## 31. 主从复制和读写分离基础认知

这属于进阶，但你最好先知道：

- 主库负责写
- 从库负责复制和部分读请求

用途：

- 提升读能力
- 增强可用性

## 32. 数据库这一章的进一步知识点

后面你还可以继续学习：

- Binlog
- Undo Log
- Redo Log
- Buffer Pool
- 分库分表
- 数据库中间件

## 33. SQL 执行顺序

很多人会写 SQL，但不清楚逻辑执行顺序。

一个典型查询语句，大致理解顺序是：

1. `from`
2. `where`
3. `group by`
4. `having`
5. `select`
6. `order by`
7. `limit`

这对你理解：

- 为什么某些别名不能直接在某些位置用
- 为什么聚合和过滤顺序不同

很有帮助。

## 34. where 和 having 的区别

### where

在分组前过滤原始数据。

### having

在分组后过滤聚合结果。

示例：

```sql
select status, count(*) as total
from user
where age >= 18
group by status
having total > 5;
```

## 35. 子查询基础认知

子查询就是在一个 SQL 中嵌套另一个 SQL。

例如：

```sql
select *
from user
where id in (
    select user_id from orders where total_amount > 100
);
```

子查询很常见，但要注意：

- 写得太复杂可能影响性能

## 36. `count(*)`、`count(1)`、`count(字段)`

你至少要知道：

- `count(*)`：统计总行数
- `count(字段)`：只统计该字段非空的行

面试里有时会问这些差异，但当前阶段你先把使用语义理解清楚更重要。

## 37. 聚簇索引和二级索引

在 InnoDB 里，这是高频概念。

### 聚簇索引

通常就是主键索引。

它的叶子节点存的是整行数据。

### 二级索引

叶子节点通常存的是主键值。

所以通过二级索引查完整行时，可能还要：

- 回表

## 38. 临时表和排序开销基础认知

当 SQL 写得不合理时，可能会出现：

- Using temporary
- Using filesort

你现在不一定要精通这些执行计划细节，但要知道：

- 排序和分组有额外成本

## 39. Redo Log、Undo Log、Binlog 基础认知

这是数据库底层高频概念。

### Redo Log

偏向保障持久性。

### Undo Log

偏向支持回滚和 MVCC。

### Binlog

偏向记录变更操作，常用于复制和恢复。

当前阶段你先知道三者职责不同即可。

## 40. Buffer Pool 基础认知

MySQL 不会每次都直接从磁盘处理数据页，它会用内存做缓存。

Buffer Pool 的价值可以简单理解为：

- 提高数据页访问效率

## 41. 行锁什么时候可能失效成更大范围锁

你要知道一个经验：

- 想用行锁，前提通常是命中了合适索引

如果索引没走好，锁范围可能变大，性能会受影响。

## 42. `select *` 为什么不推荐

除了“多查了没用字段”，还有这些问题：

- 增加网络传输
- 可能影响覆盖索引机会
- 可维护性差

## 43. 订单表索引设计示例

例如订单表常见查询：

- 按用户查订单列表
- 按订单号查详情
- 按状态和时间查订单

可能会考虑：

- 订单号唯一索引
- 用户 ID 普通索引
- 状态 + 创建时间联合索引

为什么索引设计必须结合查询场景：

- 索引不是抽象题，是业务题

## 44. 数据库字段设计常见坑

### 状态字段语义混乱

后面维护非常痛苦。

### 时间字段不统一

排查问题很难受。

### 金额用浮点数

容易有精度问题。

### 主键风格不统一

项目结构会变乱。

## 45. 这一章的高频知识点总清单

建议整理这些点：

- 基本 SQL
- join
- group by
- having
- 分页
- 表设计
- 索引类型
- 联合索引
- 最左前缀
- 覆盖索引
- 回表
- 索引失效
- 事务 ACID
- 隔离级别
- 脏读、不可重复读、幻读
- 行锁、表锁
- MVCC
- explain

---

## 46. 电商表设计示例

```sql
CREATE TABLE `user` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `username` VARCHAR(64) NOT NULL,
  `password` VARCHAR(128) NOT NULL COMMENT 'BCrypt',
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

**字段规范**：金额 `DECIMAL`；时间 `DATETIME`；状态用 `TINYINT` 并注释语义。

---

## 47. EXPLAIN 实战解读

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

**最左前缀**：索引 `(user_id, status, create_time)` 可支持 `user_id`、`user_id+status` 查询；单独 `status` 用不上该索引。

---

## 48. 索引失效常见场景

1. 对索引列使用函数：`WHERE YEAR(create_time) = 2024`
2. 隐式类型转换：`WHERE phone = 13800138000`（phone 是 varchar）
3. 左模糊：`LIKE '%abc'`
4. OR 一侧无索引
5. 联合索引跳过最左列

---

## 49. 事务隔离级别速记

| 级别 | 脏读 | 不可重复读 | 幻读 |
|------|------|------------|------|
| READ UNCOMMITTED | 可能 | 可能 | 可能 |
| READ COMMITTED | 否 | 可能 | 可能 |
| REPEATABLE READ（MySQL 默认） | 否 | 否 | 理论上可能，InnoDB 用 MVCC+间隙锁缓解 |
| SERIALIZABLE | 否 | 否 | 否 |

初学：**知道默认 RR，业务用 `@Transactional` 保证原子性即可**。

---

## 50. 慢 SQL 优化步骤

1. 开启慢查询日志，定位 SQL
2. `EXPLAIN` 看 type、key、rows
3. 补/改索引（结合 WHERE、ORDER BY）
4. 避免 `SELECT *`，只查需要的列（利于覆盖索引）
5. 分页深翻页改游标或延迟关联

---

## 51. 学完标准

- 能写多表 JOIN、分组统计、分页 SQL
- 能设计带合理索引的表，金额不用 float
- 会用 EXPLAIN 判断索引是否生效
- 能口述 ACID、隔离级别、脏读幻读
- 能写防超卖 `UPDATE ... WHERE stock >= ?`

---

## 52. 分级练习

**基础**：建 user/product/order 三表，插入测试数据  
**进阶**：写「用户订单列表」SQL 并 EXPLAIN 优化  
**挑战**：模拟没有索引的慢查询，加联合索引后对比 rows

<!-- 修改说明: 新增分级练习参考答案 -->

### 参考答案

#### 基础：三表设计与测试数据

46 节的建表 SQL 就是标准答案。插入测试数据：

```sql
INSERT INTO user (username, password) VALUES
('zhangsan', '$2a$10$xxx'), ('lisi', '$2a$10$yyy');

INSERT INTO product (name, price, stock) VALUES
('Java 编程思想', 99.00, 100),
('Spring Boot 实战', 79.00, 50);

INSERT INTO `order` (order_no, user_id, total_amount, status) VALUES
('ORD20250101001', 1, 99.00, 1),
('ORD20250101002', 1, 79.00, 0),
('ORD20250102001', 2, 99.00, 1);
```

验证：

```sql
SELECT COUNT(*) FROM user;    -- 预期：2
SELECT COUNT(*) FROM product; -- 预期：2
SELECT COUNT(*) FROM `order`; -- 预期：3
```

#### 进阶：用户订单列表 SQL + EXPLAIN 优化

**需求**：查用户 ID=1 的已支付订单，按时间倒序，分页 10 条。

**第一版（可能全表扫）**：

```sql
SELECT o.id, o.order_no, o.total_amount, o.status, o.create_time
FROM `order` o
WHERE o.user_id = 1 AND o.status = 1
ORDER BY o.create_time DESC
LIMIT 10;
```

**EXPLAIN**：

```sql
EXPLAIN SELECT o.id, o.order_no, o.total_amount, o.status, o.create_time
FROM `order` o
WHERE o.user_id = 1 AND o.status = 1
ORDER BY o.create_time DESC
LIMIT 10;
# 预期输出（有 idx_user_status_time 索引时）：
# type: ref
# key: idx_user_status_time
# rows: 较小数值
```

如果没有索引，`type` 可能是 `ALL`，`rows` 等于全表行数。加上 46 节的 `idx_user_status_time` 联合索引后，`rows` 应显著下降。

#### 挑战：慢查询对比

**无索引时**（先 DROP INDEX）：

```sql
ALTER TABLE `order` DROP INDEX idx_user_status_time;

EXPLAIN SELECT * FROM `order` WHERE user_id = 1;
# 预期：type=ALL, rows=全表行数
```

**加回索引后**：

```sql
CREATE INDEX idx_user_status_time ON `order`(user_id, status, create_time);

EXPLAIN SELECT * FROM `order` WHERE user_id = 1;
# 预期：type=ref, key=idx_user_status_time, rows=该用户的订单数
```

截图或记录两次 `rows` 对比，就是这次挑战的交付物。

---

<!-- 修改说明: 新增常见报错与排查 -->

## 52.1 常见报错与排查

| 报错信息（关键词） | 可能原因 | 解决方案 |
|-------------------|---------|---------|
| `Access denied for user` | 用户名或密码错 | 检查连接参数；Docker 容器用 `-e MYSQL_ROOT_PASSWORD` 设的密码 |
| `Unknown column 'xxx' in 'field list'` | 字段名拼错或表结构不一致 | `DESC table_name` 看实际字段 |
| `Duplicate entry 'xxx' for key 'uk_xxx'` | 违反唯一约束 | 检查是否重复插入；业务上先查再插 |
| `Lock wait timeout exceeded` | 事务持锁太久，别的会话在等 | 缩短事务；检查是否有未提交的大事务 |
| `You have an error in your SQL syntax` | SQL 语法错误 | 注意 MySQL 8 保留字：`order` 表名需反引号 `` `order` `` |
| `Data too long for column` | 插入字符串超长 | 加大 `VARCHAR` 长度或截断输入 |

---

## 54. 索引设计实战指南

### 54.1 联合索引的最左前缀原则

```sql
-- 联合索引 (a, b, c)
CREATE INDEX idx_abc ON t(a, b, c);

-- 以下查询能用上索引：
WHERE a = 1;                    -- ✅ 走 a
WHERE a = 1 AND b = 2;          -- ✅ 走 a,b
WHERE a = 1 AND b = 2 AND c = 3;-- ✅ 走全部
WHERE a = 1 AND c = 3;          -- ⚠️ 只走 a（b 断了就不能往下了）

-- 以下查询用不上索引：
WHERE b = 2;                    -- ❌ 没从 a 开始
WHERE c = 3;                    -- ❌ 没从 a 开始
WHERE b = 2 AND c = 3;          -- ❌ 没从 a 开始
```

### 54.2 索引设计检查清单

- [ ] WHERE 条件里的等值字段放在联合索引最前面
- [ ] ORDER BY 的字段可以接在 WHERE 字段后面，避免 filesort
- [ ] 区分度高的字段放前面（如 `user_id` > `status`）
- [ ] 每个表主键必备，一般用 `BIGINT AUTO_INCREMENT`
- [ ] 唯一约束用 `UNIQUE INDEX`（防重复 + 加速查找）
- [ ] 不要给每个字段单独建索引（索引占空间、拖慢写入）

---

## 55. EXPLAIN 结果解读速查

```sql
EXPLAIN SELECT * FROM `order` WHERE user_id = 1 ORDER BY create_time DESC LIMIT 10;
```

| 字段 | 含义 | 期望值 |
|------|------|--------|
| `type` | 访问类型（性能排序） | `const` > `eq_ref` > `ref` > `range` > `index` > `ALL` |
| `key` | 实际使用的索引 | 不为 NULL（NULL 表示全表扫） |
| `rows` | 预估扫描行数 | 越小越好 |
| `Extra` | 额外信息 | 避免 `Using filesort`、`Using temporary` |
| `possible_keys` | 候选索引 | 和 `key` 对比，差距大说明缺索引 |

**type 速记**：
- `ALL`：全表扫描（最差，必须优化）
- `index`：全索引扫描（也只是好一点）  
- `range`：索引范围扫描（`>` `<` `BETWEEN` `IN`）
- `ref`：非唯一索引查找（常用）
- `eq_ref`：唯一索引关联查找（JOIN 用）
- `const`：主键/唯一索引等值查询（最优）

---

## 56. 事务隔离级别详解

| 隔离级别 | 脏读 | 不可重复读 | 幻读 | 默认 |
|----------|:--:|:--:|:--:|------|
| READ UNCOMMITTED | ✅ | ✅ | ✅ | — |
| READ COMMITTED | ❌ | ✅ | ✅ | Oracle/PostgreSQL |
| REPEATABLE READ | ❌ | ❌ | ✅（InnoDB 通过间隙锁防住了） | **MySQL 默认** |
| SERIALIZABLE | ❌ | ❌ | ❌ | 最严格但性能最差 |

```sql
-- 查看当前隔离级别
SELECT @@transaction_isolation;

-- 设置隔离级别
SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;
```

---

## 57. MySQL 锁机制

| 锁类型 | 粒度 | 加锁方式 | 场景 |
|--------|------|----------|------|
| 表锁 | 整张表 | `LOCK TABLES` / MyISAM 默认 | 备份、DDL |
| 行锁 | 单行 | InnoDB 通过索引加锁 | `UPDATE ... WHERE id = 1` |
| 间隙锁 (Gap Lock) | 索引间隙 | RR 级别下自动 | 防幻读 |
| 临键锁 (Next-Key Lock) | 行 + 间隙 | RR 级别默认 | 行锁 + 间隙锁组合 |
| 意向锁 | 表级信号 | 自动 | 行锁前声明意向 |
| 乐观锁 | 逻辑锁 | `WHERE version = ?` | 用户手动实现 |
| 悲观锁 | 物理锁 | `SELECT ... FOR UPDATE` | 需要事务中强一致 |

```sql
-- 悲观锁示例：锁定某行直到事务结束
START TRANSACTION;
SELECT stock FROM product WHERE id = 1 FOR UPDATE;
UPDATE product SET stock = stock - 1 WHERE id = 1;
COMMIT;
```

---

## 58. 连接池配置（HikariCP）

Spring Boot 2+ 默认使用 HikariCP，核心参数：

```yaml
spring:
  datasource:
    hikari:
      maximum-pool-size: 20       # 最大连接数（默认 10）
      minimum-idle: 5             # 最小空闲连接
      idle-timeout: 300000        # 空闲超时（5分钟）
      connection-timeout: 30000   # 获取连接超时（30秒）
      max-lifetime: 1800000       # 连接最大生命周期（30分钟）
```

**问题排查**：应用日志出现 `Connection is not available` → 连接池耗尽，增大 `maximum-pool-size` 或检查是否有未释放的连接。

---

## 59. Binlog 与主从复制

Binlog 是 MySQL 的操作日志，用途：
1. **主从复制**：Master 写 Binlog → Slave 拉取并重放
2. **数据恢复**：备份 + Binlog 可恢复到任意时间点
3. **CDC（Change Data Capture）**：Canal/Debezium 监听 Binlog 同步到 Redis/ES/MQ

### 三种格式

| 格式 | 记录方式 | 特点 |
|------|----------|------|
| STATEMENT | SQL 语句 | 省空间，但 `NOW()` 等函数可能不一致 |
| ROW（推荐） | 每行变化 | 精准，空间稍大，主从一致性好 |
| MIXED | 混合 | 默认 STATEMENT，必要时切 ROW |

---

## 60. 学完标准（扩充版）

- [ ] 会建表、选对字段类型（金额 `DECIMAL` 不用 `FLOAT`）
- [ ] 理解联合索引的最左前缀和索引失效场景
- [ ] 会用 EXPLAIN 分析 SQL，看懂 `type`/`key`/`rows`
- [ ] 能口述 B+ 树索引原理、ACID、四种隔离级别
- [ ] 知道行锁/间隙锁/乐观锁/悲观锁的区别和应用场景
- [ ] 会用 `FOR UPDATE` 悲观锁、`version` 字段乐观锁
- [ ] 了解 HikariCP 核心参数和连接池耗尽排查
- [ ] 知道 Binlog 三种格式和主从复制基本概念
- [ ] 能完成"防超卖"场景的 SQL 实现（`UPDATE ... WHERE stock >= ?`）

---

## 60.1 常见困惑 FAQ

### Q1：数据库、表、行、列怎么记？

**A**：想成 Excel：库 = 文件，表 = Sheet，行 = 记录，列 = 字段。

### Q2：金额为什么必须用 DECIMAL？

**A**：`float/double` 二进制浮点有精度误差，累加后对账会差分；`DECIMAL` 精确十进制，Java 侧用 `BigDecimal`。

### Q3：索引是不是越多越好？

**A**：不是。索引占空间、拖慢写入；按**查询场景**建，用 EXPLAIN 验证。

### Q4：什么是联合索引最左前缀？

**A**：索引 `(a,b,c)` 可用于 `a`、`a+b`、`a+b+c` 条件；单独 `b` 或 `c` 开头用不上。

### Q5：覆盖索引和回表？

**A**：查询列全在索引里叫覆盖，不用回主表；否则二级索引查到主键再回表取完整行。

### Q6：MySQL 默认隔离级别？

**A**：InnoDB 默认 **REPEATABLE READ（可重复读）**；业务层仍用 `@Transactional` 保证原子性。

### Q7：脏读、不可重复读、幻读区别？

**A**：脏读：读到未提交数据；不可重复读：同事务两次读同一行不同；幻读：范围查询行数变多。

### Q8：`EXPLAIN` 的 `type=ALL` 什么意思？

**A**：全表扫描，最差；应通过索引优化到 `range`/`ref` 等。

### Q9：为什么 `UPDATE` 有时锁整表？

**A**：未命中索引时 InnoDB 可能锁升级扩大范围；`WHERE` 条件列要有索引。

### Q10：`LIMIT 100000, 10` 为什么慢？

**A**：要先跳过 10 万行；深分页用游标（上次最后 id）或延迟关联优化。

### Q11：char 和 varchar？

**A**：`char` 定长适合短且长度固定（如状态码）；`varchar` 变长，业务字符串常用。

### Q12：逻辑删除 `is_deleted` 好处？

**A**：数据可恢复、可审计；查询要加 `WHERE is_deleted = 0`，索引设计要考虑到。

---

## 60.2 闭卷自测

> 先遮住答案，逐题口述或默写。

### 概念题（6 道）

1. 用 Excel 类比说明库、表、行、列、主键。
2. 为什么金额用 `DECIMAL(10,2)` 不用 `double`？Java 用什么类型？
3. B+ 树索引为什么能让查询变快？（说「少翻页」即可）
4. 联合索引 `(user_id, status, create_time)` 能加速哪些 WHERE 组合？不能加速哪些？
5. 事务 ACID 各是什么？下单扣库存缺了事务会怎样？
6. `READ COMMITTED` 和 `REPEATABLE READ` 对幻读处理有何不同（了解即可）？

### 动手题（2 道）

7. 写建表 SQL：`product(id, name, price DECIMAL(10,2), stock INT)`，主键自增。
8. 写防超卖 SQL：扣 3 件库存，仅当 `stock >= 3` 时成功。

### 综合题（2 道）

9. `EXPLAIN` 显示 `type=ALL, rows=100000`，你会按什么顺序优化？
10. 订单表常查「某用户已支付订单按时间倒序分页」，索引怎么设计？为什么？

### 自测参考答案

1. 库=工作簿，表=Sheet，行=记录，列=字段，主键=唯一行号。
2. 浮点二进制误差；Java 用 `BigDecimal`。
3. 树高低，几次磁盘 IO 定位到范围，不用扫全表每一行。
4. 能：`user_id`；`user_id+status`；+时间排序。不能：单独 `status`。
5. 原子/一致/隔离/持久；可能订单写了库存没扣或反之。
6. RC 每次读最新提交；RR 同事务快照读，InnoDB 间隙锁缓解幻读。
7. `CREATE TABLE product (... price DECIMAL(10,2) NOT NULL, stock INT NOT NULL DEFAULT 0);`
8. `UPDATE product SET stock = stock - 3 WHERE id = ? AND stock >= 3;` 看 `affected rows`。
9. 查 SQL 写法→补索引→避免 `SELECT *`→避免函数破坏索引→减少 JOIN。
10. `(user_id, status, create_time)`；等值在前、排序字段在后，避免 filesort。

---

## 60.3 费曼检验

**任务**：请在不看资料的情况下，用 **3 分钟** 向朋友解释「MySQL 索引和事务」。

**对照提纲**：

1. **Excel 升级版**：表存业务数据；05 章 MyBatis 来读写它。
2. **索引**：像目录，按手机号查人不用翻全通讯录；乱建索引浪费纸（磁盘）。
3. **事务**：转账两步绑定；要么都成功要么都撤销，防止「钱扣了对方没收到」。

若朋友能说出「索引加快查询、事务保证一批操作一致」，本章核心已掌握。

---

## 60.4 本章与后续章节衔接速查

| 本章学会 | 05 章怎么用 | 07 章怎么用 |
|----------|-------------|-------------|
| 表设计与索引 | MyBatis SQL 要命中索引 | 缓存 key 常含 id |
| `EXPLAIN` | 优化 Mapper 里的慢 SQL | 判断慢在库还是缓存 |
| 事务隔离 | 配合 `@Transactional` 理解 | 缓存一致性场景 |
| `DECIMAL` | Entity 字段类型对应 | 价格缓存序列化 |
| Docker MySQL | 05 章 datasource 连接 | Redis 也常 Docker 起 |

### 60.4.1 电商 order 表建表逐行读

```sql
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

| 字段/子句 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `` `order` `` 反引号 | `order` 是保留字 | 不加反引号语法错误 |
| `DECIMAL(10,2)` | 精确金额 | 用 `DOUBLE` 可能对账差分 |
| `uk_order_no` | 订单号唯一，防重复下单 | 无唯一约束可插入重复单号 |
| `idx_user_status_time` | 用户订单列表 + 状态 + 时间排序 | 缺索引则 `EXPLAIN type=ALL` |
| `InnoDB` | 支持事务与行锁 | MyISAM 不适合订单事务场景 |
| `utf8mb4` | 完整 Unicode | `utf8` 三字节可能 emoji 异常 |

### 60.4.2 EXPLAIN 结果逐列读

| 列 | 好现象 | 差现象 | 优化方向 |
|----|--------|--------|----------|
| `type` | `ref`、`range`、`const` | `ALL` 全表扫 | 补索引、改 WHERE |
| `key` | 显示索引名 | `NULL` | 检查联合索引与最左前缀 |
| `rows` | 很小 | 接近表总行数 | 缩小扫描范围 |
| `Extra` | `Using index` 覆盖 | `Using filesort`、`Using temporary` | 调整索引包含 ORDER BY 列 |

**动手验收清单**：

- [ ] Docker 启动 MySQL 并导入三表
- [ ] 无索引 vs 有索引截图对比 `EXPLAIN rows`
- [ ] 写防超卖 UPDATE 并解释 `affected rows`
- [ ] 闭卷自测 ≥ 8/10

---

## 60.5 常见学习弯路与纠正

| 弯路 | 表现 | 纠正 |
|------|------|------|
| 金额用 `DOUBLE` | 余额 0.999... | `DECIMAL` + Java `BigDecimal` |
| 到处 `SELECT *` | 慢且难覆盖索引 | 只查需要的列 |
| 每个字段单独建索引 | 写入变慢、占空间 | 按查询建联合索引 |
| 忽略最左前缀 | 建了索引仍全表扫 | §54 对照 WHERE 条件 |
| 无 WHERE 的 UPDATE | 全表被改 | 强制带主键/条件 |
| 不跑 EXPLAIN | 慢 SQL 靠猜 | 改前后对比 `type`/`rows` |
| 表名 `order` 不加引号 | 语法错误 | 反引号 `` `order` `` |
| 只会 Docker 不会 SQL | 容器起了查不了 | §4.1 必须 `docker exec` 练 CRUD |

---

<!-- 修改说明: 新增下一章预告 -->

## 面试深挖补充：MySQL 底层原理与锁机制

前面 §1～§60.5 把 MySQL 的"用法和工程套路"铺开了，但面试官深挖的是几个**底层原理**：MVCC 到底怎么实现快照读？索引下推改了什么执行流程？间隙锁到底锁的是什么区间？change buffer 为什么只对非唯一索引有效？这一节把这些问题一次讲透。

> 这节是给 §12/§15/§16/§27/§37/§39/§40/§57 等"基础认知"小节补上"底层为什么"，建议对照着读。

### A. MVCC 的 ReadView 机制（高频深挖）

**一句话**：MVCC 通过给每行加**隐藏的事务ID字段** + **undo log 版本链** + **ReadView 可见性判断**，让读操作不加锁也能看到某个一致的快照；RC 和 RR 的区别就在于"什么时候生成 ReadView"。

**InnoDB 每行的三个隐藏字段**：
- `DB_TRX_ID`：最近一次修改这行的事务 ID。
- `DB_ROLL_PTR`：回滚指针，指向 undo log 里这行的上一个版本。
- `DB_ROW_ID`：没有主键时的隐藏主键（有主键就不存在）。

**undo log 版本链**：每次修改一行，旧值被写入 undo log，通过 `DB_ROLL_PTR` 串成一条链——链头是当前值，往后是越来越老的版本。这条链就是"多版本"的来源。

**ReadView 的四个字段**（一次快照读生成一份）：
- `m_ids`：生成 ReadView 时，当前所有**活跃（未提交）**事务的 ID 列表。
- `min_trx_id`：`m_ids` 里最小的。
- `max_trx_id`：下一个要分配的事务 ID（不是 m_ids 最大值，而是系统已分配的最大 +1）。
- `creator_trx_id`：创建这个 ReadView 的事务自己的 ID。

**可见性判断规则**（对某行当前版本的 `DB_TRX_ID` 做）：
1. `trx_id < min_trx_id`：这个版本在生成 ReadView 前就提交了 → **可见**。
2. `trx_id >= max_trx_id`：这个版本是 ReadView 生成后才启动的事务写的 → **不可见**。
3. `min_trx_id <= trx_id < max_trx_id`：
   - 如果 `trx_id` 在 `m_ids` 里：写这个版本的事务在生成 ReadView 时还活跃 → **不可见**。
   - 如果不在 `m_ids` 里：写这个版本的事务在生成 ReadView 前已提交 → **可见**。
4. `trx_id == creator_trx_id`：自己改的 → **可见**。
5. 不可见时，顺着 `DB_ROLL_PTR` 找 undo log 的更老版本，再套上面的规则，直到找到可见版本或链尾。

**RC 和 RR 的核心区别（ReadView 生成时机）**：
- **RC（Read Committed）**：**每次 SELECT 都生成新的 ReadView**。所以事务里两次读之间，别的事务提交了，新 ReadView 能看到新值——这就是"不可重复读"。
- **RR（Repeatable Read）**：**事务第一次 SELECT 时生成 ReadView，整个事务复用这一份**。所以无论别的事务怎么提交，本事务看到的都是第一次读时的快照——这就是"可重复读"。

**为什么 MVCC 能解决快照读的幻读**：RR 下 ReadView 固定，新插入的行 `DB_TRX_ID` 一定 ≥ max_trx_id（不可见），所以快照读看不到新插入的行。但**当前读**（`SELECT ... FOR UPDATE`/`UPDATE`/`DELETE`）不走 MVCC，仍能看到新行，所以 RR 下当前读仍有幻读风险，需要间隙锁来防（见 §C）。

**面试标准答法**：
> MVCC 靠每行隐藏的 trx_id 和 roll_ptr 串起 undo log 版本链，配合 ReadView（m_ids/min_trx_id/max_trx_id/creator_trx_id）做可见性判断：版本事务ID < min 可见、≥ max 不可见、在区间看是否在活跃列表。RC 每次 SELECT 生成新 ReadView 所以能读到别人新提交，RR 第一次 SELECT 生成后复用所以可重复读。快照读靠 MVCC 防幻读，当前读仍需间隙锁。

---

### B. 索引下推 ICP（Index Condition Pushdown）

**一句话**：没有 ICP 时，存储引擎按联合索引最左前缀找到记录就回表，由 server 层用 WHERE 其它条件过滤；有 ICP 时，WHERE 中能用上联合索引剩余列的条件被"下推"到引擎层，先在索引上过滤再回表，减少回表次数。5.6 引入。

**为什么需要 ICP**：回表是随机 IO，很贵。联合索引 `(a, b)`，如果 `WHERE a LIKE '张%' AND b = 18`：
- `a LIKE '张%'` 能用最左前缀（LIKE 前缀匹配走索引）。
- 但 `b = 18` 在 `a` 是范围匹配时，按 B+ 树规则用不到索引第二列做范围裁剪（只有 a 是等值时 b 才能继续走索引）。

**没有 ICP 的执行流程**：
1. 存储引擎按 `a LIKE '张%'` 在联合索引上找到所有张姓记录的主键。
2. **逐条回表**拿到完整行，返回 server 层。
3. server 层用 `b = 18` 过滤，丢掉不符合的。
- 问题：如果张姓有 10000 条但 b=18 的只有 10 条，回表了 10000 次只为留 10 条，浪费严重。

**有 ICP 的执行流程**：
1. 存储引擎按 `a LIKE '张%'` 在联合索引上找张姓记录。
2. **在索引上直接用 `b = 18` 过滤**（b 也在联合索引里，索引项包含 a 和 b）。
3. 只有 b=18 的记录才回表。
- 结果：只回表 10 次，大幅减少随机 IO。

**ICP 的生效条件**：
- 是联合索引（单列索引谈不上下推）。
- WHERE 条件里有一部分能用到联合索引的后续列（但受最左前缀/范围匹配限制，原本用不上做索引扫描的部分）。
- 不是聚簇索引（聚簇索引本身就是整行，没有"回表"概念）。
- 子查询的某些场景不支持。

**怎么判断用了 ICP**：`EXPLAIN` 的 `Extra` 列出现 `Using index condition` 就是用了 ICP（区别于 `Using where` 是 server 层过滤、`Using index` 是覆盖索引）。

**面试标准答法**：
> 索引下推 ICP 是 5.6 引入的优化。联合索引 (a,b) 查 a LIKE '张%' AND b=18 时，无 ICP 引擎按 a 范围扫到所有张姓记录回表，server 层再过滤 b；有 ICP 把 b=18 下推到引擎层在索引上先过滤再回表，大幅减少回表随机 IO。EXPLAIN Extra 显示 Using index condition。生效需联合索引且条件能用到索引后续列、非聚簇索引。

---

### C. 间隙锁与临键锁（锁机制深挖）

**一句话**：间隙锁锁的是索引记录之间的"间隙"，防止别的事务往间隙插入，解决 RR 下的幻读；临键锁（Next-Key Lock）= 行锁 + 间隙锁，锁住左开右闭区间，是 RR 下 InnoDB 的默认行锁。

**为什么要有间隙锁**：RR 级别下，一个事务内两次范围读要看到相同结果，不能有别的事务往这个范围里"插入新行"（插入新行就是幻读）。行锁只能锁住已有行，锁不住"还不存在的间隙"——所以引入间隙锁锁住间隙。

**三种锁的关系**：
- **记录锁（Record Lock）**：锁住索引上的某一条记录。
- **间隙锁（Gap Lock）**：锁住记录之间的间隙，**只防插入、不防读/改已有记录**。间隙锁之间不互斥（多个事务可以同时"持有"同一间隙锁），但间隙锁和插入意向锁冲突。
- **临键锁（Next-Key Lock）**：记录锁 + 它前面的间隙，锁住 `(前一条记录, 当前记录]` 这种左开右闭区间。是 InnoDB 在 RR 下的默认行锁。

**锁退化的两种情况（面试常考）**：
- **等值查询唯一索引，命中记录**：临键锁退化为**记录锁**（唯一索引命中，间隙里不可能再插入相同值，不用锁间隙）。
- **等值查询，未命中记录**：临键锁退化为**间隙锁**（锁住命中的间隙，防止插入这个值）。
- 等值查询非唯一索引命中：仍是临键锁，且还会向右继续加间隙锁直到下一个不匹配的值。
- 范围查询：按临键锁逐个加。

**举例（RR 级别，表有 id=5,10,15，id 是主键）**：
- 事务 A `SELECT * FROM t WHERE id = 10 FOR UPDATE`（等值命中唯一索引）→ 退化为记录锁，只锁 id=10。
- 事务 A `SELECT * FROM t WHERE id = 7 FOR UPDATE`（等值未命中）→ 退化为间隙锁，锁 (5,10) 间隙，事务 B 想 INSERT id=7/8/9 都阻塞。
- 事务 A `SELECT * FROM t WHERE id > 10 FOR UPDATE`（范围）→ 临键锁锁 (10,15]、(15,+∞)，事务 B INSERT id=11/12/16 都阻塞。

**只在 RR 级别有间隙锁**：RC 级别不加间隙锁（只有记录锁），所以 RC 下当前读会有幻读，但并发度更高。这是 RC 和 RR 并发性能差异的主要原因。

**面试标准答法**：
> 间隙锁锁索引记录之间的间隙防插入，解决 RR 幻读，只防插入不防读写已有行，多个间隙锁不互斥。临键锁=行锁+间隙锁锁左开右闭区间，是 RR 下 InnoDB 默认行锁。等值查唯一索引命中退化为记录锁，等值未命中退化为间隙锁，范围查询用临键锁。RC 不加间隙锁所以当前读有幻读但并发高。这是 RC/RR 性能差异主因。

---

### D. change buffer / 自适应哈希索引 / MRR（InnoDB 优化三件套）

**这三个是 InnoDB 提升性能的"隐藏机制"，面试加分点。**

**change buffer（5.5 前叫 insert buffer）**：
- **作用**：对**非唯一二级索引**的修改，如果目标索引页不在 Buffer Pool 中，不立即从磁盘读页，先把修改记在 change buffer，等这个页后续被读到内存时再 merge。
- **为什么只对非唯一索引**：唯一索引修改必须读页判断是否违反唯一性约束，不读页就没法判断，所以用不了 change buffer。非唯一索引不需要判断唯一性，可以延迟 merge。
- **收益**：减少随机读 IO，写多读少的非唯一二级索引场景提升明显。
- **代价**：merge 前索引和实际数据不一致（但通过 change buffer 能查到正确结果），崩溃恢复要 redo log 保护 change buffer。

**自适应哈希索引（Adaptive Hash Index, AHI）**：
- **作用**：InnoDB 自动监控热点查询，对经常被等值访问的索引页建内存中的哈希索引，把 B+ 树的 O(log n) 等值查找变成 O(1)。
- **全自动**：不需要建索引语句，InnoDB 自己决定建哪些。
- **什么时候要关掉**：高并发等值查询场景下，AHI 的哈希表用读写锁保护，竞争可能成为瓶颈，关掉 `innodb_adaptive_hash_index=off` 反而更快。这是少见的"优化项可能反噬"的例子。

**MRR（Multi-Range Read）**：
- **作用**：范围查询走二级索引拿到一批主键后，**先按主键排序再回表**，把回表的随机 IO 转成顺序 IO。
- **为什么快**：随机 IO 的寻道开销远大于顺序 IO，按主键排序后回表能命中相邻数据页，磁盘读更高效。
- **启用**：`SET optimizer_switch='mrr=on,mrr_cost_based=off'`（默认开但 cost-based 可能因成本估算而不用，强制开要关 cost-based）。
- **配套**：Batched Key Access（BKA）join 在 MRR 基础上批量回表。

**三个的共性**：都是 InnoDB 在 B+ 树基础上的性能补丁——change buffer 优化写、AHI 优化等值读、MRR 优化范围读回表。面试能讲全这三个，区分度高。

**面试标准答法**：
> change buffer 对非唯一二级索引的修改延迟 merge，减少随机读 IO，唯一索引必须读页判唯一性用不了。自适应哈希索引 InnoDB 自动给热点索引建内存哈希表把等值查变 O(1)，但高并发下哈希表读写锁竞争可能反噬可关闭。MRR 把范围查询的二级索引主键排序后回表，随机 IO 转顺序 IO，配 BKA join 批量回表。三者是 InnoDB 在 B+ 树上的写/等值读/范围读优化补丁。

---

### E. 这几个深挖点的关联

- **A MVCC + C 间隙锁**：MVCC 解决快照读的幻读（看不到新插入），间隙锁解决当前读的幻读（防新插入）。两者一起才让 RR 真正防幻读。
- **A MVCC + §16/§17 隔离级别**：RC/RR 的可重复读差异本质就是 ReadView 生成时机，理解了 MVCC 就理解了隔离级别为什么这样表现。
- **B ICP + §11/§12 索引/B+树**：ICP 优化的是"联合索引回表前过滤"，前提是理解 B+ 树索引项包含哪些列、最左前缀怎么用。
- **D change buffer + §37 唯一/二级索引**：change buffer 只对非唯一二级索引生效，正好呼应"唯一索引要读页判唯一性"。
- **D AHI + B+ 树**：AHI 是给热点 B+ 树页加哈希旁路，理解 B+ 树的 O(log n) 才知道 AHI 把它降到 O(1) 的价值。

---

### F. 面试自检（这节看完应能答）

- [ ] MVCC 的三个隐藏字段是什么？ReadView 的四个字段？可见性判断的几条规则？
- [ ] RC 和 RR 的 ReadView 生成时机有什么区别？这导致了什么现象？
- [ ] 索引下推 ICP 改了哪一步执行流程？EXPLAIN 怎么看用了 ICP？
- [ ] 间隙锁锁的是什么？等值查唯一索引命中、未命中分别退化成什么锁？
- [ ] 为什么 RC 没有间隙锁？这带来什么好处和坏处？
- [ ] change buffer 为什么只对非唯一索引有效？
- [ ] 自适应哈希索引什么情况下反而要关掉？MRR 把什么 IO 转成什么 IO？

---

## 下一章预告

MySQL 能持久化存储，但磁盘 IO 是瓶颈——商品详情页每次查库，高峰期数据库可能扛不住。下一章（07 Redis 核心原理与缓存实战）引入 **缓存层**：

- Redis 为什么比 MySQL 快那么多（内存 + 单线程 + IO 多路复用）
- **Cache Aside** 模式：读时先查缓存、写时先更库再删缓存
- 用 Redis 做商品详情缓存、ZSet 排行榜、SETNX 分布式锁

06 章解决"数据怎么存、怎么查快"，07 章解决"热点数据怎么扛高并发"——这是后端性能优化的第一道防线。

---

*下一章：07 Redis 核心原理与缓存实战*
