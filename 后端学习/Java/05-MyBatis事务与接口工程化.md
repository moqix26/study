# MyBatis、事务与接口工程化

<!-- 修改说明: 新增本章与上一章的关系 -->

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
