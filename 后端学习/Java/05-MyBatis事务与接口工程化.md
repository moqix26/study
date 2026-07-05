# MyBatis、事务与接口工程化

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0、FAQ、闭卷自测、费曼检验 -->

## 0. 读前导读（零基础也能跟上）

> **读者假设**：你已学完 [04 Spring Boot](./04-SpringBoot核心开发.md)，能写 Controller + Service，但数据还在内存 `ArrayList` 里。

### 0.1 用一句话弄懂本章

**一句话**：**MyBatis** 是 Java 与 MySQL 之间的**翻译官**——你写 SQL，它把查询结果翻译成 Java 对象；再加上**事务**和 **DTO/VO 分层**，接口就能像真实项目一样持久化、安全、可维护。

**生活类比——MyBatis = 专业翻译官**：

| 角色 | 代码层 | 生活类比 |
|------|--------|----------|
| **你（业务方）** | Service / Controller | 中国老板，说中文需求 |
| **MyBatis** | Mapper 接口 + XML | **翻译官**：把中文需求译成英文（SQL）发给老外，再把回复译成中文（Java 对象） |
| **MySQL** | 数据库 | 外国档案室，只认 SQL |
| **Entity** | `User.java` 与表字段对应 | 档案室里的**标准表格一行** |
| **DTO** | 前端 POST 进来的参数 | 客人填的**申请表**（可能含密码，不能原样对外展示） |
| **VO** | 返回给前端的视图 | 对外展示的**脱敏简历**（不含密码） |
| **@Transactional** | 事务注解 | **银行转账**：扣款和入账必须同时成功或同时取消 |

**为什么重要**：国内 Java 岗 MyBatis 使用率极高；`#{}` vs `${}` 是安全面试必考；不会事务，下单扣库存必出线上事故。

**本章用到的地方**：§7.1 手把手接入、§10 事务、§50 `#{}` vs `${}`。

---

### 0.2 你需要提前知道什么（真不会就先跳到哪一章）

| 你现在的水平 | 建议动作 |
|--------------|----------|
| 不会 Spring Boot 三层 | 先学 [04 Spring Boot](./04-SpringBoot核心开发.md) §2.1 |
| 没装 MySQL | 本章 §7.1 可先跟代码；数据库环境见 [06 MySQL](./06-MySQL基础索引与事务.md) §4.1 Docker |
| 不会写 SQL | 边学边抄本章建表语句；06 章系统补 SQL |
| 已跑通 04 demo | **从 §7.1 手把手开始**，把 List 换成 Mapper |

**最低门槛**：知道 HTTP POST/GET；理解 Controller 调 Service；能改 `application.yml`。

---

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

- [ ] 在 pom 加入 `mybatis-spring-boot-starter` 和 MySQL 驱动
- [ ] 配置 `spring.datasource` 和 `mybatis.mapper-locations`
- [ ] 写 `UserMapper` 接口 + `UserMapper.xml` 完成 CRUD
- [ ] 用 `#{}` 传参，说出为什么能防 SQL 注入
- [ ] 写 `<where>` + `<if>` 动态 SQL 条件查询
- [ ] 区分 Entity / DTO / VO 各在哪一层用
- [ ] 用 `@Transactional` 实现「下单 + 扣库存」同成功或同回滚
- [ ] 说出事务失效的 3 种常见原因（非 public、自调用、吞异常）
- [ ] 实现分页接口：`list` + `total` + `pageNum` + `pageSize`
- [ ] 排查 `BindingException`、`Invalid bound statement` 等常见报错
- [ ] 闭卷自测 10 题正确 ≥ 8 题

---

### 0.4 建议学习时长与节奏

| 阶段 | 建议时间 | 做什么 |
|------|----------|--------|
| §0 + §2～§7 概念 | 1.5 小时 | MyBatis 定位、`#{}`、Mapper 结构 |
| §7.1 手把手跟做 | 3～4 小时 | 04 demo 接 MySQL，**每步 curl 验证** |
| §9～§15 分层与事务 | 2 小时 | DTO/VO、@Transactional、幂等概念 |
| §48～§51 动态 SQL 与批量 | 2 小时 | if/foreach、传播行为 |
| 分级练习 + 自测 | 2 小时 | 订单事务 demo |

**节奏建议**：先跑通 CRUD 再学动态 SQL；事务 demo 务必故意制造异常看是否回滚。

---

### 0.5 学完本章你能做什么（可验证的具体动作）

1. **重启** Spring Boot 后，POST 新增用户，GET 仍能查到——数据已持久化。
2. **curl** 分页接口，返回 `list`、`total`、`pageNum`、`pageSize`。
3. **解释** 为什么用户输入拼进 `${}` 可能删表，而 `#{}` 不会。
4. **实现** `createOrder`：库存不足时抛异常，订单表和库存表都不该只成功一半。
5. **画出** 请求链路：Controller → Service → Mapper → XML → MySQL。

---

### 0.6 手把手总览：04 demo 接 MyBatis（步骤索引）

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | pom 加 MyBatis + mysql-connector-j | Maven Reload 无红错 | 见 §47.1 依赖下载失败 |
| 2 | `application.yml` 配 datasource | 启动无 `Communications link failure` | 确认 MySQL 已启动、库已建 |
| 3 | 执行建表 SQL | `SHOW TABLES` 有 `user` | 见 §47.1 Unknown database |
| 4 | 建 Entity + Mapper 接口 + XML | 启动无 `Invalid bound statement` | 查 namespace 与方法名 |
| 5 | Service 注入 Mapper 替代 List | 编译通过 | `@Mapper` 或 `@MapperScan` |
| 6 | POST 新增用户 | curl 返回 `id`；库里 `SELECT` 有记录 | 看 insert 是否 `useGeneratedKeys` |
| 7 | 重启后再 GET | 数据仍在 | 持久化验收通过 |
| 8 | 写 `@Transactional` 订单 demo | 库存不足时两表都不变 | 加 `rollbackFor = Exception.class` |

详细代码见 §7.1 第二～十步。

---

## 本章与上一章的关系

04 章你用 Spring Boot 写好了 REST 接口，但数据还在内存 `ArrayList` 里——重启就丢，也没法做复杂查询。这一章要解决的正是这个问题：**让接口真正连上 MySQL 数据库**。

MyBatis 是国内 Java 后端用得最多的持久层框架，SQL 你自己写、自己控，适合业务查询越来越复杂的场景。学完这章你会做 CRUD、分页、动态 SQL、事务控制，还会搞懂 `#{}` 和 `${}` 的安全差异——后者是面试必考，也是线上事故的常见根源。

---

## 1. 这份文档解决什么问题

当你已经会写 Spring Boot 接口后，接下来最关键的就是：

- 怎么连数据库
- 怎么写 SQL
- 怎么把事务和业务逻辑串起来
- 怎么把项目写得更像真实业务系统

## 2. MyBatis 是什么

MyBatis 是一个持久层框架，简单理解就是：

- 让 Java 代码更方便地执行 SQL
- 把查询结果映射成对象

为什么国内 Java 后端岗位经常用它：

- SQL 可控
- 上手快
- 适合复杂业务查询

## 3. 一个最基础的 Mapper

```java
public interface UserMapper {
    User selectById(Long id);
}
```

如果用注解写法：

```java
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;

@Mapper
public interface UserMapper {

    @Select("select id, username, age from user where id = #{id}")
    User selectById(Long id);
}
```

## 4. 实体类

```java
public class User {
    private Long id;
    private String username;
    private Integer age;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public Integer getAge() {
        return age;
    }

    public void setAge(Integer age) {
        this.age = age;
    }
}
```

## 5. 新增、修改、删除

```java
@Mapper
public interface UserMapper {

    @Insert("insert into user(username, age) values(#{username}, #{age})")
    int insert(User user);

    @Update("update user set username = #{username}, age = #{age} where id = #{id}")
    int update(User user);

    @Delete("delete from user where id = #{id}")
    int deleteById(Long id);
}
```

返回值通常是影响行数。

## 6. 为什么很多项目更喜欢 XML 写 SQL

因为复杂 SQL 用 XML 可读性更好。

比如条件查询：

```xml
<select id="selectByCondition" resultType="com.example.User">
    select id, username, age
    from user
    <where>
        <if test="username != null and username != ''">
            and username = #{username}
        </if>
        <if test="age != null">
            and age = #{age}
        </if>
    </where>
</select>
```

这就是动态 SQL 的典型场景。

## 7. `#` 和 `$` 的区别

这是面试高频。

### `#{}` 

- 使用预编译参数
- 更安全
- 推荐默认使用

### `${}`

- 直接字符串拼接
- 容易有 SQL 注入风险

一般原则：

- 普通参数优先用 `#{}` 
- `${}` 只在确有必要时谨慎使用

<!-- 修改说明: 补充 #{} 与 ${} 的深入解释与真实案例 -->

### 为什么 `#{}` 能防 SQL 注入，而 `${}` 不能？

**结论**：`#{}` 走预编译，`${}` 是字符串拼接——后者把用户输入直接拼进 SQL 语句结构里，攻击者可以改写 SQL 语义。

**底层原理**：

MyBatis 处理 `#{}` 时，会把 SQL 发给数据库驱动做 **PreparedStatement 预编译**。比如：

```sql
select * from user where username = ?
```

不管用户传 `'admin'` 还是 `'admin' OR '1'='1'`，数据库都只把整段字符串当作 **username 的值** 来匹配，不会改变 SQL 的逻辑结构。

而 `${}` 是在 MyBatis 拼 SQL **之前**就把字符串插进去：

```sql
select * from user where username = 'admin' OR '1'='1'
```

`'1'='1'` 永远为真，WHERE 条件被绕过，可能查出全部用户——这就是 SQL 注入。

**真实案例（模拟）**：

某后台管理系统用 `${}` 拼接排序字段：

```xml
<select id="list">
  SELECT * FROM user ORDER BY ${sortField} ${sortOrder}
</select>
```

攻击者传 `sortField=user; DROP TABLE user; --`，如果数据库权限配置不当，可能导致删表。即使不删表，也可能通过 `sortField=(CASE WHEN (SELECT password FROM admin LIMIT 1) LIKE 'a%' THEN id ELSE name END)` 做盲注，逐字猜出管理员密码。

**正确做法**：排序字段用白名单校验，只允许 `id`、`create_time` 等固定值；普通参数一律 `#{}`。

---

## 7.1 手把手：04 章 demo 接入 MySQL + MyBatis

下面在 04 章 `demo` 项目基础上继续改，把内存 List 换成真正的数据库操作。

### 第一步：加依赖（pom.xml 追加）

```xml
<dependency>
    <groupId>org.mybatis.spring.boot</groupId>
    <artifactId>mybatis-spring-boot-starter</artifactId>
    <version>3.0.3</version>
</dependency>
<dependency>
    <groupId>com.mysql</groupId>
    <artifactId>mysql-connector-j</artifactId>
    <scope>runtime</scope>
</dependency>
```

### 第二步：配置数据源

`application.yml` 追加（MySQL 需先启动，06 篇教 Docker 方式）：

```yaml
spring:
  datasource:
    url: jdbc:mysql://localhost:3306/study_db?useUnicode=true&characterEncoding=utf8&serverTimezone=Asia/Shanghai
    username: root
    password: 123456
    driver-class-name: com.mysql.cj.jdbc.Driver

mybatis:
  mapper-locations: classpath:mapper/*.xml
  configuration:
    map-underscore-to-camel-case: true
```

### 第三步：建表

在 MySQL 里执行：

```sql
CREATE DATABASE IF NOT EXISTS study_db DEFAULT CHARSET utf8mb4;
USE study_db;

CREATE TABLE user (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    username    VARCHAR(50)  NOT NULL,
    age         INT          NOT NULL DEFAULT 0,
    create_time DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

```bash
mysql -u root -p -e "source init.sql"
# 预期输出：无报错，或 Query OK
```

### 第四步：项目结构（在 04 章基础上新增）

```text
demo/
├── src/main/java/com/example/demo/
│   ├── entity/User.java              ← 新增
│   ├── mapper/UserMapper.java        ← 新增
│   └── service/UserService.java      ← 改造
└── src/main/resources/
    └── mapper/UserMapper.xml         ← 新增
```

### 第五步：Entity

```java
package com.example.demo.entity;

import java.time.LocalDateTime;

public class User {
    private Long id;
    private String username;
    private Integer age;
    private LocalDateTime createTime;

    public Long getId() { return id; }
    public void setId(Long id) { this.id = id; }
    public String getUsername() { return username; }
    public void setUsername(String username) { this.username = username; }
    public Integer getAge() { return age; }
    public void setAge(Integer age) { this.age = age; }
    public LocalDateTime getCreateTime() { return createTime; }
    public void setCreateTime(LocalDateTime createTime) { this.createTime = createTime; }
}
```

### 第六步：Mapper 接口

```java
package com.example.demo.mapper;

import com.example.demo.entity.User;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface UserMapper {
    User selectById(@Param("id") Long id);
    List<User> selectPage(@Param("offset") int offset, @Param("size") int size);
    int countAll();
    int insert(User user);
    int updateById(User user);
    int deleteById(@Param("id") Long id);
}
```

### 第七步：UserMapper.xml

`resources/mapper/UserMapper.xml`：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="com.example.demo.mapper.UserMapper">

    <select id="selectById" resultType="com.example.demo.entity.User">
        SELECT id, username, age, create_time
        FROM user WHERE id = #{id}
    </select>

    <select id="selectPage" resultType="com.example.demo.entity.User">
        SELECT id, username, age, create_time
        FROM user
        ORDER BY id DESC
        LIMIT #{offset}, #{size}
    </select>

    <select id="countAll" resultType="int">
        SELECT COUNT(*) FROM user
    </select>

    <insert id="insert" useGeneratedKeys="true" keyProperty="id">
        INSERT INTO user(username, age) VALUES(#{username}, #{age})
    </insert>

    <update id="updateById">
        UPDATE user SET username = #{username}, age = #{age} WHERE id = #{id}
    </update>

    <delete id="deleteById">
        DELETE FROM user WHERE id = #{id}
    </delete>
</mapper>
```

### 第八步：改造 UserService

```java
package com.example.demo.service;

import com.example.demo.dto.UserDTO;
import com.example.demo.entity.User;
import com.example.demo.mapper.UserMapper;
import com.example.demo.vo.UserVO;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.stream.Collectors;

@Service
public class UserService {

    private final UserMapper userMapper;

    public UserService(UserMapper userMapper) {
        this.userMapper = userMapper;
    }

    public UserVO create(UserDTO dto) {
        User user = new User();
        user.setUsername(dto.getName());
        user.setAge(dto.getAge());
        userMapper.insert(user);
        return toVO(user);
    }

    public UserVO findById(Long id) {
        User user = userMapper.selectById(id);
        return user == null ? null : toVO(user);
    }

    public List<UserVO> findPage(int pageNum, int pageSize) {
        int offset = (pageNum - 1) * pageSize;
        return userMapper.selectPage(offset, pageSize).stream()
                .map(this::toVO)
                .collect(Collectors.toList());
    }

    public int countAll() {
        return userMapper.countAll();
    }

    public boolean deleteById(Long id) {
        return userMapper.deleteById(id) > 0;
    }

    private UserVO toVO(User user) {
        return new UserVO(user.getId(), user.getUsername(), user.getAge());
    }
}
```

### 第九步：Controller 加分页

```java
@GetMapping
public Result<Map<String, Object>> list(
        @RequestParam(defaultValue = "1") Integer pageNum,
        @RequestParam(defaultValue = "10") Integer pageSize) {
    List<UserVO> list = userService.findPage(pageNum, pageSize);
    int total = userService.countAll();
    Map<String, Object> page = Map.of(
            "list", list,
            "total", total,
            "pageNum", pageNum,
            "pageSize", pageSize
    );
    return Result.success(page);
}
```

### 第十步：运行验证

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"李四\",\"age\":22}"
# 预期输出：
# {"code":0,"message":"success","data":{"id":1,"name":"李四","age":22}}

curl "http://localhost:8080/api/users?pageNum=1&pageSize=10"
# 预期输出：
# {"code":0,"message":"success","data":{"list":[...],"total":1,"pageNum":1,"pageSize":10}}
```

重启项目后再查，数据还在——这就是持久化的意义。

---

## 8. Service 层如何调用 Mapper

```java
import org.springframework.stereotype.Service;

@Service
public class UserService {

    private final UserMapper userMapper;

    public UserService(UserMapper userMapper) {
        this.userMapper = userMapper;
    }

    public User getById(Long id) {
        return userMapper.selectById(id);
    }
}
```

## 9. DTO、Entity、VO 的配合

### 9.1 请求进来

前端传来 `DTO`

### 9.2 入库或查库

数据库映射 `Entity`

### 9.3 返回前端

返回 `VO`

这样做的好处：

- 分层清晰
- 不把数据库字段直接暴露出去
- 更利于长期维护

## 10. 事务

### 10.1 事务是什么

事务是一组操作，要么都成功，要么都失败。

比如下单流程可能要做：

1. 写订单
2. 扣库存
3. 写支付状态

如果中途某一步失败，就不应该只成功一半。

### 10.2 `@Transactional`

```java
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class OrderService {

    private final OrderMapper orderMapper;
    private final StockMapper stockMapper;

    public OrderService(OrderMapper orderMapper, StockMapper stockMapper) {
        this.orderMapper = orderMapper;
        this.stockMapper = stockMapper;
    }

    @Transactional
    public void createOrder() {
        orderMapper.insertOrder();
        stockMapper.reduceStock();
    }
}
```

### 10.3 事务常见失效场景

这是面试高频。

常见失效原因：

- 方法不是 `public`
- 同类内部直接调用
- 异常被吞掉
- 没有真正抛出运行时异常

## 11. 接口幂等

真实项目里很常见。

什么是幂等：

- 同一个请求重复提交，多次执行结果应该一致或可控

场景：

- 防重复下单
- 防重复支付
- 防按钮连点

常见实现思路：

- token 防重复提交
- 唯一索引
- Redis 锁或去重标记

## 12. 分页查询

分页是业务接口的高频需求。

```java
public class PageRequest {
    private Integer pageNum;
    private Integer pageSize;
}
```

SQL 层通常会用：

```sql
select id, username, age
from user
limit 0, 10;
```

## 13. 统一错误码

项目里最好不要直接到处返回随机字符串。

可以统一定义错误码，比如：

- `0`：成功
- `1001`：参数错误
- `1002`：用户不存在
- `1003`：库存不足

这样前后端协作更清晰。

## 14. 登录鉴权的工程化思路

一个基础登录系统通常会有：

1. 注册接口
2. 登录接口
3. token 生成
4. 拦截器校验 token
5. 获取当前登录用户

你做项目时最好至少落地一版简单实现。

## 15. 统一时间字段和逻辑删除

真实项目常见公共字段：

- `create_time`
- `update_time`
- `is_deleted`

为什么要有：

- 审计数据
- 排查问题
- 逻辑删除后可恢复或追踪

## 16. 项目工程化建议

### 16.1 包结构统一

不要写得乱七八糟。

### 16.2 命名统一

比如：

- `UserController`
- `UserService`
- `UserMapper`
- `UserDTO`
- `UserVO`

### 16.3 公共代码抽离

如：

- 返回结果类
- 全局异常类
- 工具类
- 常量类

### 16.4 不要把 SQL 写得过于分散

复杂 SQL 最好有清晰归属。

## 17. 这一章该怎么练

建议你自己实现：

1. 用户新增、删除、修改、查询
2. 用户分页查询
3. 条件查询
4. 一个下单事务 demo
5. 一个登录接口

## 18. 学完标准

如果你能做到下面这些，就说明你这部分已经很扎实：

- 会用 MyBatis 做基本 CRUD
- 知道动态 SQL 怎么写
- 知道 `#{}` 和 `${}` 的区别
- 知道事务该怎么用
- 知道项目为什么要分 DTO、Entity、VO

## 19. resultType 和 resultMap

### resultType

适合字段和对象属性比较简单、能直接对应的场景。

### resultMap

适合更复杂的映射，比如：

- 字段名和属性名不完全一致
- 需要更细粒度控制

## 20. 一对多查询的基础认知

例如：

- 一个用户有多个订单

这类场景你可以先理解为：

- 可以联表查
- 也可以分两步查

初学阶段更重要的是先把简单查询写稳。

## 21. 事务传播行为基础认知

Spring 事务里有一个常见知识点：

- 传播行为

你现在先记最常见的：

- `REQUIRED`

含义可以简单理解为：

- 有事务就加入，没有就新建

## 22. 事务隔离级别和数据库的关系

事务不只是 Spring 的事，也和 MySQL 本身的隔离级别相关。

所以你学习事务时要把两层都串起来：

- Spring 如何声明事务
- 数据库如何实现事务隔离

## 23. 接口设计的基本原则

设计接口时建议注意：

- 路径命名清晰
- 请求方法语义清晰
- 入参结构明确
- 返回结构统一
- 错误码统一

例如：

- `GET /users/{id}`
- `POST /users`
- `PUT /users/{id}`
- `DELETE /users/{id}`

## 24. 幂等实现的更多思路

除了前面提到的 token 和 Redis，还可以从这些角度理解：

- 唯一业务号
- 数据库唯一索引
- 状态机控制

比如支付回调，往往就会结合：

- 订单号唯一
- 支付状态幂等校验

## 25. 代码分层时常见错误

### controller 写太多业务

会导致难维护。

### service 里 SQL 到处拼

职责不清晰。

### VO、DTO、Entity 混着用

后面很容易失控。

## 26. 工程化中常见公共组件

后端项目里你可以逐步抽这些公共能力：

- 统一响应类
- 统一异常类
- 统一错误码
- 用户上下文工具
- 日期工具
- 分页对象

## 27. 接口安全基础认知

虽然现在你还不一定做得很深，但应该知道这些点：

- 密码不能明文存储
- 登录 token 要校验
- 敏感接口要鉴权
- 关键参数要校验
- 防重复提交

## 28. MyBatis 这一章的补充知识点

后面你还可以继续学习：

- 插件机制
- 分页插件
- 批量操作
- 乐观锁字段
- 逻辑删除方案

## 29. MyBatis 一级缓存和二级缓存基础认知

这是面试里经常会提到的知识点。

### 一级缓存

- 默认开启
- 作用域通常在同一个 SqlSession 内

### 二级缓存

- 跨 SqlSession
- 需要额外配置

不过你要知道，真实项目里很多团队对缓存使用很谨慎，因为：

- 容易造成一致性理解复杂

## 30. 批量插入和批量更新基础认知

当数据量较大时，一条一条执行 SQL 会慢。

所以会有：

- 批量插入
- 批量更新

你现在至少要知道：

- 这是性能优化的常见方向

## 31. 乐观锁基础认知

数据库并发更新时，除了悲观锁，还常见乐观锁。

常见做法：

- 增加 `version` 字段

更新时带上旧版本号：

```sql
update product
set stock = stock - 1, version = version + 1
where id = 1 and version = 3;
```

如果影响行数为 0，说明版本冲突。

## 32. 逻辑删除

很多业务不会真的把数据物理删除，而是：

- 增加 `is_deleted`

好处：

- 数据可追溯
- 误删后更容易恢复

代价：

- 查询时要注意过滤

## 33. 登录鉴权的工程化细节

一个更像样的登录系统通常要考虑：

- 密码加密
- token 生成
- token 校验
- 退出登录
- 权限拦截

你现在做项目时，至少要做到：

- 密码不明文存储
- 有基础 token 机制

## 34. JWT 基础认知

JWT 经常用于登录态。

你先知道几个核心点：

- 登录成功后生成 token
- 前端后续请求携带 token
- 后端校验 token

常见问题方向：

- token 过期
- token 篡改
- 退出登录如何处理

## 35. 分页对象设计

项目里很适合统一一个分页请求对象和分页返回对象。

### 分页请求

- 页码
- 页大小

### 分页返回

- 数据列表
- 总数
- 当前页
- 总页数

这样接口风格会统一很多。

## 36. 接口文档和联调习惯

做接口不只是“写完能跑”，还要能协作。

建议养成这些习惯：

- 入参字段写清楚
- 返回字段写清楚
- 错误码写清楚
- 用接口文档工具联调

## 37. API 设计常见坏味道

### 一个接口职责太多

会导致维护困难。

### 路径命名混乱

不利于团队协作。

### 参数随意堆在 query 里

复杂场景应合理使用 JSON 请求体。

### 返回值毫无统一结构

会让前端和排查都很难受。

## 38. Service 层怎么写更像真实项目

建议遵循：

- 一个方法对应一个明确业务动作
- 一个方法内流程尽量清楚
- 复杂逻辑适当拆小方法
- 参数对象和返回对象清晰

## 39. 订单创建流程的工程化拆解

一个更像样的 `createOrder` 方法往往会包含：

1. 参数校验
2. 查询商品
3. 校验库存
4. 计算金额
5. 创建订单
6. 扣库存
7. 写事务日志或发送异步消息

这个过程里会用到：

- MyBatis
- 事务
- 统一异常
- 日志

## 40. 这一章的高频知识点总清单

建议整理这些点：

- MyBatis 基本 CRUD
- XML 和注解
- 动态 SQL
- `#{}` 和 `${}`
- `resultType` 和 `resultMap`
- 一级缓存、二级缓存
- 事务
- 传播行为
- 逻辑删除
- 乐观锁
- DTO、VO、Entity
- 统一响应
- 接口幂等

---

## 41. MyBatis 完整 CRUD 示例

### Entity

```java
@Data
public class User {
    private Long id;
    private String username;
    private String password;
    private LocalDateTime createTime;
}
```

### Mapper 接口

```java
@Mapper
public interface UserMapper {
    User selectById(@Param("id") Long id);
    List<User> selectPage(@Param("offset") int offset, @Param("size") int size);
    int insert(User user);
    int updateById(User user);
    int deleteById(@Param("id") Long id);
}
```

### UserMapper.xml

```xml
<mapper namespace="com.example.mapper.UserMapper">
  <select id="selectById" resultType="com.example.entity.User">
    SELECT id, username, password, create_time AS createTime
    FROM user WHERE id = #{id}
  </select>

  <insert id="insert" useGeneratedKeys="true" keyProperty="id">
    INSERT INTO user(username, password, create_time)
    VALUES(#{username}, #{password}, #{createTime})
  </insert>

  <update id="updateById">
    UPDATE user SET username = #{username} WHERE id = #{id}
  </update>
</mapper>
```

**安全**：参数一律用 `#{}`，`${}` 仅用于动态表名/排序列（且要白名单校验）。

---

## 42. 动态 SQL 示例

```xml
<select id="selectByCondition" resultType="com.example.entity.User">
  SELECT * FROM user
  <where>
    <if test="username != null and username != ''">
      AND username LIKE CONCAT('%', #{username}, '%')
    </if>
    <if test="status != null">
      AND status = #{status}
    </if>
  </where>
  ORDER BY create_time DESC
</select>
```

---

## 43. `@Transactional` 实战要点

```java
@Service
public class OrderService {
    @Transactional(rollbackFor = Exception.class)
    public Long createOrder(CreateOrderDTO dto) {
        // 1. 查商品 2. 校验库存 3. 插订单 4. 扣库存
        // 任一步抛异常 → 全部回滚
    }
}
```

| 传播行为 | 含义（初学记 2 个） |
|----------|---------------------|
| REQUIRED（默认） | 有事务就加入，没有就新建 |
| REQUIRES_NEW | 挂起当前，新建独立事务 |

**失效场景**：同类内部自调用、方法非 public、异常被吞掉。

---

## 44. 乐观锁示例

```sql
UPDATE product SET stock = stock - #{num}, version = version + 1
WHERE id = #{id} AND stock >= #{num} AND version = #{version}
```

更新行数为 0 → 并发冲突，可重试或提示用户。

---

## 45. 统一异常与 Result（工程化）

```java
@RestControllerAdvice
public class GlobalExceptionHandler {
    @ExceptionHandler(BusinessException.class)
    public Result<?> handleBusiness(BusinessException e) {
        return Result.fail(e.getCode(), e.getMessage());
    }
    @ExceptionHandler(Exception.class)
    public Result<?> handleOther(Exception e) {
        log.error("系统异常", e);
        return Result.fail(500, "系统繁忙");
    }
}
```

---

## 46. 学完标准

- 能配置 MyBatis，写接口 + XML 完成 CRUD
- 会使用 `#{}`、动态 SQL、`resultMap`
- 理解一级/二级缓存概念，知道何时关闭二级缓存
- 会用 `@Transactional` 控制订单类业务事务
- 能区分 Entity / DTO / VO，接口返回统一结构

---

## 47. 分级练习

**基础**：用户表 CRUD + 分页查询  
**进阶**：订单创建（订单表 + 库存扣减）同一事务  
**挑战**：接口幂等：订单号 `SETNX` + 唯一索引双保险

<!-- 修改说明: 新增分级练习参考答案 -->

### 参考答案

#### 基础：用户表 CRUD + 分页

7.1 节的完整代码就是标准答案。验证清单：

- [ ] POST 新增后数据库有记录
- [ ] GET `/api/users/{id}` 能查到
- [ ] GET `/api/users?pageNum=1&pageSize=10` 返回 list + total
- [ ] DELETE 后查不到

#### 进阶：订单创建事务

**建表**：

```sql
CREATE TABLE product (
    id    BIGINT PRIMARY KEY AUTO_INCREMENT,
    name  VARCHAR(100) NOT NULL,
    stock INT NOT NULL DEFAULT 0
);

CREATE TABLE orders (
    id         BIGINT PRIMARY KEY AUTO_INCREMENT,
    product_id BIGINT NOT NULL,
    quantity   INT    NOT NULL,
    amount     DECIMAL(10,2) NOT NULL,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**OrderService（核心）**：

```java
@Service
public class OrderService {

    private final OrderMapper orderMapper;
    private final ProductMapper productMapper;

    public OrderService(OrderMapper orderMapper, ProductMapper productMapper) {
        this.orderMapper = orderMapper;
        this.productMapper = productMapper;
    }

    @Transactional(rollbackFor = Exception.class)
    public Long createOrder(CreateOrderDTO dto) {
        Product product = productMapper.selectById(dto.getProductId());
        if (product == null) {
            throw new RuntimeException("商品不存在");
        }
        if (product.getStock() < dto.getQuantity()) {
            throw new RuntimeException("库存不足");
        }

        Order order = new Order();
        order.setProductId(dto.getProductId());
        order.setQuantity(dto.getQuantity());
        order.setAmount(product.getPrice().multiply(BigDecimal.valueOf(dto.getQuantity())));
        orderMapper.insert(order);

        int rows = productMapper.reduceStock(dto.getProductId(), dto.getQuantity());
        if (rows == 0) {
            throw new RuntimeException("扣库存失败");
        }
        return order.getId();
    }
}
```

**ProductMapper.xml 扣库存 SQL**：

```xml
<update id="reduceStock">
    UPDATE product
    SET stock = stock - #{quantity}
    WHERE id = #{productId} AND stock >= #{quantity}
</update>
```

**验证**：库存 5，下单 3 成功；再下单 3 应抛异常且第一次订单不受影响（事务回滚）。

#### 挑战：接口幂等（思路）

1. 前端生成唯一 `orderNo`，提交时带上来
2. 数据库 `orders` 表对 `order_no` 建 **唯一索引**
3. 下单前 Redis `SETNX order:lock:{orderNo} 1 EX 60`
4. SETNX 失败 → 直接返回"请勿重复提交"
5. 插入订单若触发唯一索引冲突 → 查已有订单返回，不重复扣库存

---

<!-- 修改说明: 新增常见报错与排查 -->

## 47.1 常见报错与排查

| 报错信息（关键词） | 可能原因 | 解决方案 |
|-------------------|---------|---------|
| `Communications link failure` | MySQL 没启动或地址/port 错 | 确认 MySQL 运行中；检查 `application.yml` 的 url、用户名密码 |
| `Unknown database 'study_db'` | 库没建 | 执行 `CREATE DATABASE study_db` |
| `BindingException: Parameter 'id' not found` | XML 里 `#{}` 参数名和 `@Param` 不一致 | 多参数接口加 `@Param("id")`；XML 用 `#{id}` |
| `Invalid bound statement (not found)` | XML namespace 或 id 和接口方法对不上 | 检查 `namespace="com.example.demo.mapper.UserMapper"` 和方法名 |
| `Table 'study_db.user' doesn't exist` | 表没建 | 执行建表 SQL |
| `@Transactional` 不回滚 | 异常被 catch 吞了，或抛的是 checked Exception | 加 `rollbackFor = Exception.class`；不要吞异常 |

---

## 48. 动态 SQL（MyBatis XML 进阶）

### 48.1 if + where

```xml
<select id="findByConditions" resultType="User">
    SELECT id, username, age FROM user
    <where>
        <if test="username != null and username != ''">
            AND username LIKE CONCAT('%', #{username}, '%')
        </if>
        <if test="age != null">
            AND age = #{age}
        </if>
    </where>
    ORDER BY id DESC
</select>
```

### 48.2 foreach（批量操作）

```xml
<!-- 批量插入 -->
<insert id="batchInsert">
    INSERT INTO user (username, age, create_time) VALUES
    <foreach collection="list" item="user" separator=",">
        (#{user.username}, #{user.age}, NOW())
    </foreach>
</insert>

<!-- 批量删除 -->
<delete id="batchDelete">
    DELETE FROM user WHERE id IN
    <foreach collection="ids" item="id" open="(" separator="," close=")">
        #{id}
    </foreach>
</delete>
```

### 48.3 choose / when / otherwise（类似 switch）

```xml
<select id="findByDynamic" resultType="User">
    SELECT id, username, age FROM user
    <where>
        <choose>
            <when test="keyword != null and keyword != ''">
                username LIKE CONCAT('%', #{keyword}, '%')
            </when>
            <when test="age != null">
                age >= #{age}
            </when>
            <otherwise>
                status = 1
            </otherwise>
        </choose>
    </where>
</select>
```

### 48.4 set（动态更新）

```xml
<update id="updateSelective">
    UPDATE user
    <set>
        <if test="username != null">username = #{username},</if>
        <if test="age != null">age = #{age},</if>
        <if test="status != null">status = #{status},</if>
    </set>
    WHERE id = #{id}
</update>
```

`<set>` 自动处理末尾多余的逗号。

---

## 49. 事务传播行为

`@Transactional` 除了 `rollbackFor`，还有一个重要属性 `propagation`：

| 传播行为 | 含义 | 场景 |
|----------|------|------|
| `REQUIRED`（默认） | 有事务就加入，没有就新建 | 绝大多数业务 |
| `REQUIRES_NEW` | 总是新建事务，挂起当前事务 | 记录日志（不受外层失败影响） |
| `NESTED` | 嵌套事务，内层失败不影响外层 | 批量处理（单条失败回退但不影响整批） |
| `SUPPORTS` | 有就用，没有也无所谓 | 只读查询 |
| `MANDATORY` | 必须在事务里，没有就报错 | 强制事务保护 |
| `NOT_SUPPORTED` | 以非事务方式执行 | 不需要事务的重操作 |
| `NEVER` | 不能在事务里，有就报错 | — |

```java
// 日志记录：不管下单成功与否，日志都要写入
@Service
public class OrderLogService {

    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void log(String orderNo, String action) {
        orderLogMapper.insert(new OrderLog(orderNo, action));
    }
}
```

---

## 50. `#{}` vs `${}` 彻底搞懂

| | `#{}`（预编译占位符） | `${}`（字符串替换） |
|----|----|----|
| 防 SQL 注入 | ✅ 防（参数不会拼入 SQL 结构） | ❌ 不防 |
| 处理方式 | 用 `?` 占位，参数单独传 | 直接拼接字符串 |
| 何时用 | **所有参数值**（99% 的情况） | 表名、列名、ORDER BY 字段 |

```xml
<!-- ✅ 正确：用户名字段用 #{} -->
<select id="findByName">
    SELECT id, username FROM user WHERE username = #{username}
</select>

<!-- ✅ 正确：动态排序字段必须用 ${}，但需白名单校验 -->
<select id="findByOrder">
    SELECT id, username FROM user ORDER BY ${orderColumn} ${orderDir}
</select>
```

```java
// 用 ${} 时必须白名单校验！
private static final Set<String> ALLOWED_COLUMNS = Set.of("id", "username", "age", "create_time");
private static final Set<String> ALLOWED_DIR = Set.of("ASC", "DESC");

public List<User> findByOrder(String orderColumn, String orderDir) {
    if (orderColumn == null || !ALLOWED_COLUMNS.contains(orderColumn)) {
        throw new IllegalArgumentException("无效的排序列: " + orderColumn);
    }
    if (orderDir == null || !ALLOWED_DIR.contains(orderDir.toUpperCase())) {
        throw new IllegalArgumentException("无效的排序方向: " + orderDir);
    }
    return mapper.findByOrder(orderColumn, orderDir.toUpperCase());
}
```

---

## 51. 批量操作优化

### 51.1 批量插入（一次 SQL 插多行）

```java
@Mapper
public interface UserMapper {
    int batchInsert(@Param("list") List<User> users);
}
```

```xml
<insert id="batchInsert">
    INSERT INTO user (username, age) VALUES
    <foreach collection="list" item="item" separator=",">
        (#{item.username}, #{item.age})
    </foreach>
</insert>
```

### 51.2 批量更新（每条 SQL 不同值）

```java
@Mapper
public interface OrderMapper {
    int batchUpdateStatus(@Param("orders") List<OrderStatusDTO> orders);
}
```

```xml
<update id="batchUpdateStatus">
    <foreach collection="orders" item="order" separator=";">
        UPDATE `order`
        SET status = #{order.status}
        WHERE id = #{order.id}
    </foreach>
</update>
```

⚠️ 需要在 JDBC URL 加参数允许分号多条 SQL：

```yaml
spring:
  datasource:
    url: jdbc:mysql://localhost:3306/study_db?allowMultiQueries=true
```

### 51.3 BATCH 模式（ExecutorType.BATCH）

大批量操作时，启用 MyBatis 的 BATCH 模式可大幅提升性能：

```java
@Service
public class UserBatchService {

    private final SqlSessionTemplate sqlSessionTemplate;

    public void batchInsert(List<User> users) {
        SqlSession session = sqlSessionTemplate.getSqlSessionFactory()
                .openSession(ExecutorType.BATCH, false);
        try {
            UserMapper mapper = session.getMapper(UserMapper.class);
            for (User user : users) {
                mapper.insert(user);  // 只是攒着，还没执行
            }
            session.commit();  // 一批提交
        } catch (Exception e) {
            session.rollback();
            throw e;
        } finally {
            session.close();
        }
    }
}
```

---

## 52. 代码生成器（MyBatis-Plus CodeGenerator）

手写 Entity、Mapper、XML 太枯燥？社区常用 MyBatis-Plus 的生成器：

```java
public class CodeGenerator {
    public static void main(String[] args) {
        FastAutoGenerator.create(
                "jdbc:mysql://localhost:3306/study_db", "root", "123456")
            .globalConfig(builder -> builder
                .author("你的名字")
                .outputDir(System.getProperty("user.dir") + "/src/main/java"))
            .packageConfig(builder -> builder
                .parent("com.example.demo"))
            .strategyConfig(builder -> builder
                .addTablePrefix("t_")
                .addInclude("user", "product", "order"))
            .execute();
    }
}
```

**注意**：初学阶段手写 SQL 和理解映射关系更重要，生成器是熟练后的提效工具。

---

## 53. 学完标准（扩充版）

- [ ] MyBatis + MySQL 实现完整 CRUD（Controller → Service → Mapper → XML）
- [ ] 会用 `<where>`、`<if>`、`<foreach>`、`<set>` 写动态 SQL
- [ ] 理解 `#{}`（预编译防注入）与 `${}`（字符串拼接）的区别
- [ ] 会使用 `@Transactional(rollbackFor = Exception.class)` 控制事务
- [ ] 了解事务传播行为（REQUIRED / REQUIRES_NEW）及适用场景
- [ ] 知道批量操作的 BATCH 模式和分页插件使用
- [ ] 能独立完成"下单扣库存"类事务业务
- [ ] 能区分 `rollbackFor = Throwable.class` vs 默认回滚规则

---

## 53.1 常见困惑 FAQ

### Q1：MyBatis 和 JPA/Hibernate 怎么选？

**A**：MyBatis **SQL 自己写、可控**，适合复杂查询和国内多数项目；JPA 自动生成 SQL，简单 CRUD 快但复杂 SQL 难调。本路线以 MyBatis 为主。

### Q2：Mapper 用注解还是 XML？

**A**：简单单表 CRUD 注解够用；**多条件、动态 SQL、联表**优先 XML，可读性和维护性更好。

### Q3：`#{}` 和 `${}` 一句话区别？

**A**：`#{}` 预编译占位 `?`，防注入；`${}` 字符串拼接，仅用于**白名单校验过**的表名/排序列。

### Q4：为什么 `@Transactional` 加了却不回滚？

**A**：常见：方法非 public、同类 `this.xxx()` 自调用绕过代理、catch 后没再抛出、抛的是 checked 异常且未配置 `rollbackFor`。

### Q5：Entity 能直接返回给前端吗？

**A**：不要。Entity 含数据库字段（密码、删除标记）；用 **VO** 脱敏后返回，DTO 接收入参。

### Q6：`@Param` 什么时候必须写？

**A**：Mapper 方法有**多个参数**时，XML 里 `#{id}` 必须和 `@Param("id")` 对应，否则 `BindingException`。

### Q7：insert 后怎么拿到自增 id？

**A**：XML 加 `useGeneratedKeys="true" keyProperty="id"`，插入后 Entity 的 `id` 字段会被回填。

### Q8：一级缓存和二级缓存要用吗？

**A**：知道概念即可。一级缓存同 SqlSession 默认开；二级缓存真实项目常关或慎用，一致性复杂。

### Q9：分页用 `LIMIT` 还是 PageHelper？

**A**：初学手写 `LIMIT offset, size` + `count(*)` 理解原理；熟练后可用 PageHelper 插件提效。

### Q10：事务传播 `REQUIRED` 和 `REQUIRES_NEW`？

**A**：`REQUIRED` 默认，有事务加入没有新建；`REQUIRES_NEW` 挂起外层新建独立事务，适合**记日志**不受外层回滚影响。

### Q11：`map-underscore-to-camel-case` 干什么？

**A**：数据库 `create_time` 自动映射 Java `createTime`，少写 `AS createTime`。

### Q12：接口幂等为什么重要？

**A**：用户连点支付、网络重试会导致重复下单；用唯一订单号 + 唯一索引 + Redis SETNX 保证多次请求结果一致。

---

## 53.2 闭卷自测

> 先遮住答案，逐题口述或默写。

### 概念题（6 道）

1. 用「翻译官」类比说明 MyBatis 在 Java 和 MySQL 之间做什么？
2. `#{}` 防 SQL 注入的底层原理是什么？`${}` 为什么危险？
3. Entity、DTO、VO 分别在请求的哪一段出现？各解决什么问题？
4. 事务 ACID 四个字母各表示什么？下单+扣库存为什么需要事务？
5. `@Transactional` 三种常见失效场景？
6. 动态 SQL `<where>` + `<if>` 解决什么问题？

### 动手题（2 道）

7. 写 `UserMapper` 方法：`selectByUsername(String username)`，XML 用 `#{}` 查询。
8. 写 `OrderService.createOrder` 伪代码：`insert order` + `reduceStock`，加 `@Transactional(rollbackFor = Exception.class)`。

### 综合题（2 道）

9. 攻击者传 `sortField=id; DROP TABLE user` 给用 `${sortField}` 的接口，会发生什么？正确防御步骤？
10. 画出 POST `/api/users` 从 Controller 到 MySQL 再返回 VO 的完整链路，标出 DTO 和 Entity 转换位置。

### 自测参考答案

1. Java 调 Mapper 像说中文；MyBatis 译成 SQL 给 MySQL，结果集译回 Java 对象。
2. `#{}` 用 PreparedStatement `?` 占位，输入只是参数值；`${}` 直接拼进 SQL 结构，可改写语句语义。
3. DTO：入参；Entity：库表映射；VO：出参。避免暴露密码、统一 API 契约。
4. 原子性、一致性、隔离性、持久性；任一步失败应全部回滚，不能只有订单无库存。
5. 非 public、同类自调用、吞异常（或 checked 异常未 rollbackFor）。
6. 条件可选时自动拼 WHERE，避免 `WHERE AND` 语法错或写死多个 SQL。
7. `@Select` 或 XML：`WHERE username = #{username}`。
8. 查商品→校验库存→insert→update stock；任一步 `throw` 则回滚。
9. 可能执行多语句删表（视权限）；排序列白名单 + 普通参数一律 `#{}`。
10. Controller 收 DTO→Service 转 Entity→Mapper/XML→MySQL→Entity→Service 转 VO→Result。

---

## 53.3 费曼检验

**任务**：请在不看资料的情况下，用 **3 分钟** 向朋友解释「MyBatis 和事务是干什么的」。

**对照提纲**：

1. **持久化**：04 章数据在内存，重启没了；MyBatis 连 MySQL，翻译 SQL 和对象。
2. **分层**：DTO 进、VO 出，Entity 对表；翻译官不直接把档案室钥匙给前端。
3. **事务**：转账两步必须绑在一起，@Transactional 保证「要么全成要么全撤」。

若朋友能说出「MyBatis 写 SQL 映射对象、事务防止只成功一半」，本章核心已掌握。

---

## 53.4 本章与后续章节衔接速查

| 本章学会 | 06 章怎么用 | 07 章怎么用 |
|----------|-------------|-------------|
| 写 SQL / 索引字段 | 理解为什么 `WHERE` 列要有索引 | 缓存前先查库 |
| `DECIMAL` 金额 | 表设计与 Java `BigDecimal` 一致 | 商品价格缓存注意精度 |
| `@Transactional` | 隔离级别在 MySQL 实现 | 缓存与库一致性 |
| 分页 `LIMIT` | 深分页性能问题 | 热点列表可缓存 |
| `#{}` 安全 | 防注入是双方责任 | — |

### 53.4.1 UserMapper.xml `selectById` 逐行读

```xml
<select id="selectById" resultType="com.example.demo.entity.User">
    SELECT id, username, age, create_time
    FROM user WHERE id = #{id}
</select>
```

| 行号/字段 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `id="selectById"` | 与 Mapper 接口方法名一致 | 对不上报 `Invalid bound statement` |
| `resultType=...User` | 每行结果映射为 User 对象 | 包名错则映射失败或字段全 null |
| `create_time` | 库字段下划线 | 无 `map-underscore-to-camel-case` 时需 `AS createTime` |
| `#{id}` | 预编译参数，防注入 | 改成 `${id}` 有注入风险 |
| `namespace`（文件头） | 必须等于 Mapper 接口全限定名 | 错则所有 SQL 绑不上 |

### 53.4.2 `@Transactional` 订单方法逐行读

```java
@Transactional(rollbackFor = Exception.class)
public Long createOrder(CreateOrderDTO dto) {
    Product product = productMapper.selectById(dto.getProductId());
    if (product.getStock() < dto.getQuantity()) {
        throw new RuntimeException("库存不足");
    }
    orderMapper.insert(order);
    productMapper.reduceStock(dto.getProductId(), dto.getQuantity());
    return order.getId();
}
```

| 步骤 | 含义 | 改错会怎样 |
|------|------|------------|
| `rollbackFor = Exception.class` | checked 异常也回滚 | 默认只回滚 RuntimeException，业务异常可能不回滚 |
| 先查库存再 insert | 业务校验顺序 | 颠倒可能插入无效订单 |
| `reduceStock` 放最后 | 与 insert 同事务 | 中间抛异常则 insert 也撤销 |
| 同类 `this.createOrder()` 调用 | 绕过代理 | 事务不生效，需注入自身或拆类 |

**动手验收清单**：

- [ ] 04 demo 重启后用户数据仍在
- [ ] 能口述 `#{}` vs `${}` 并写白名单排序
- [ ] 库存不足时订单表无脏数据（事务回滚）
- [ ] 闭卷自测 ≥ 8/10

---

## 53.5 常见学习弯路与纠正

| 弯路 | 表现 | 纠正 |
|------|------|------|
| Entity 直接返前端 | 密码字段泄露 | 严格 DTO 进、VO 出 |
| 全用 `${}` 图省事 | SQL 注入风险 | 参数值一律 `#{}` |
| 不写 `@Param` | 多参数 BindingException | XML 名与 `@Param` 对齐 |
| `@Transactional` 自调用 | 事务不回滚 | 拆 Service 或注入自身代理 |
| catch 后不抛出 | 吞异常导致半成功 | 记录后 rethrow 或转业务异常 |
| XML namespace 抄错 | 启动正常查询 404 | 等于 Mapper 接口全限定名 |
| 忽略 `useGeneratedKeys` | insert 后 id 为 null | XML 配 `keyProperty="id"` |
| 不做事务 demo | 以为注解自动万能 | §47 进阶练习故意失败验回滚 |

---

<!-- 修改说明: 新增下一章预告 -->

## 下一章预告

这一章你会写 SQL、用 MyBatis 连库、控制事务了——但有个问题：MySQL 本身是怎么工作的？为什么有的查询快、有的慢到超时？为什么 `UPDATE` 有时锁整表？

下一章（06 MySQL 基础、索引与事务）从数据库底层补全这些知识：

- 表设计规范、三范式、字段类型怎么选（金额为什么用 `DECIMAL`）
- **索引和 B+ 树**：为什么加了索引查询能快 100 倍
- **事务隔离级别**：脏读、不可重复读、幻读到底是什么
- 用 Docker 一键启动 MySQL，不用再折腾本地安装

MyBatis 是"怎么写 SQL"，MySQL 是"SQL 在数据库里怎么执行"——两章合在一起，你对数据层就有完整认知了。

---

*下一章：06 MySQL 基础、索引与事务*
