# Java 基础语法与面向对象

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0、FAQ、闭卷自测、费曼检验 -->

## 0. 读前导读（零基础也能跟上）

> **读者假设**：你会复制粘贴、会用电脑；**不一定**写过代码。本章从「写第一行 Java」到「用面向对象组织代码」。

### 0.1 用一句话弄懂本章

**一句话**：Java 是一门**面向对象**语言——你先写**类（蓝图）**，再 `new` 出**对象（房子）**，用**封装、继承、多态**把数据和操作组织起来，为后面 Spring Boot 写接口打底。

**生活类比——OOP = 建筑行业**：

| 概念 | 代码 | 生活类比 |
|------|------|----------|
| **类（Class）** | `class User { ... }` | **建筑蓝图**：规定房子有几室、门在哪，蓝图本身不住人 |
| **对象（Object）** | `User u = new User()` | **按蓝图盖好的房子**：张三家住 101，李四家住 102，各住各的 |
| **属性（Field）** | `String name; int age` | 房子的**户型数据**：面积、楼层 |
| **方法（Method）** | `void sayHello()` | 房子里的**功能**：开灯、开门 |
| **封装（Encapsulation）** | `private` + getter/setter | **防盗门 + 物业登记**：外人不能随意改你家门牌，要按规矩来 |
| **继承（Inheritance）** | `class Dog extends Animal` | **复式户型继承基础户型**：狗「是一种」动物，复用吃、睡的行为 |
| **多态（Polymorphism）** | `Animal a = new Dog()` | **同一张「宠物登记表」**：登记的是动物，实际可能是狗或猫，叫声不同 |

**为什么重要**：Spring 里每个 `@Service`、`@Controller` 都是类；JSON 转对象、数据库行映射对象，全是 OOP。不会类与对象，后面每一章都费劲。

**本章用到的地方**：§10 面向对象、§11～§13 继承多态、§3.1 IDEA 手把手。

---

### 0.2 你需要提前知道什么（真不会就先跳到哪一章）

| 你现在的水平 | 建议动作 |
|--------------|----------|
| 完全零基础 | 从 §3 第一个程序开始；00 路线图只看「工具安装」 |
| 学过 Python/JS 语法 | 重点看 §10～§14 OOP 与 Java 差异（类型、封装） |
| 已会 OOP 想直接写接口 | 速读 §0～§10，自测过关后跳 [04 Spring Boot](./04-SpringBoot核心开发.md) |
| 不懂 IDEA 怎么建项目 | **直接跟 §3.1 手把手**，每步对照「预期看到什么」 |

**最低门槛**：能安装 JDK 17、打开 IDEA；知道「变量 = 装数据的盒子」。

---

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

- [ ] 在 IDEA 创建 Java 项目并运行 `Hello Java`
- [ ] 声明 `int`/`double`/`boolean`/`String` 变量，理解基本类型 vs 引用类型
- [ ] 使用 `if/for/while/switch` 写分支与循环
- [ ] 定义方法：参数、返回值、`return`
- [ ] 用「蓝图 vs 房子」向他人解释类与对象
- [ ] 写带 `private` 字段 + getter/setter 的封装类
- [ ] 用 `extends` 做继承，用 `@Override` 做重写
- [ ] 说出多态：`父类引用指向子类对象` 的好处
- [ ] 知道 `==` 与 `equals` 对 String 的区别
- [ ] 会 `try-catch` 处理异常，避免 `NullPointerException`
- [ ] 完成分级练习「计算器 / 学生选课 / 银行账户」至少各一题
- [ ] 闭卷自测 10 题正确 ≥ 8 题

---

### 0.4 建议学习时长与节奏

| 阶段 | 建议时间 | 做什么 |
|------|----------|--------|
| §0 + §3～§9 语法 | 4～6 小时 | 变量、流程、方法、数组、String |
| §10～§14 OOP 核心 | 4～5 小时 | 类对象、封装继承多态、抽象类接口 |
| §15～§38 进阶与练习 | 3～4 小时 | 异常、static、enum、分级练习 |
| FAQ + 闭卷自测 + 费曼 | 1 小时 | 检验是否真懂 |

**节奏建议**：每学完一大节就写 10 行代码验证；OOP 部分务必动手写 `User`/`Student` 类，不要只读。

---

### 0.5 学完本章你能做什么（可验证的具体动作）

1. **创建** IDEA Java 项目，Run 出 `Hello Java`，`exit code 0`。
2. **编写** 含构造方法、`getName`/`setName` 的 `User` 类，在 `main` 里创建两个不同对象。
3. **实现** 成绩判断：输入 85 输出「良好」（`if-else`）。
4. **解释** 为什么 `String a = "abc"; String b = new String("abc");` 时 `a == b` 为 false 而 `equals` 为 true。
5. **用继承** 写 `Animal` + `Dog`/`Cat`，多态调用 `sound()` 输出不同叫声。
6. **排查** `; expected`、`NullPointerException` 两类常见报错。

---

### 0.6 核心术语首次出现速查

| 术语 | 一句话 | 生活类比 |
|------|--------|----------|
| **JDK** | 写 Java、运行 Java 的工具箱 | 厨房全套厨具 |
| **JVM** | 运行 `.class` 的虚拟机 | 电器用的电源适配器 |
| **类 Class** | 对象的模板 | 楼盘户型蓝图 |
| **对象 Object** | 类的具体实例 | 某一栋具体的房 |
| **封装** | 隐藏数据，通过方法访问 | 银行账户密码不能公开改 |
| **继承** | 子类获得父类属性和方法 | 子承父业 |
| **多态** | 同一引用不同实际类型不同行为 | 同一「叫一声」，猫喵狗汪 |
| **接口 interface** | 只规定能做什么，不管怎么做 | 插座标准：两孔三孔规范 |
| **异常 Exception** | 程序运行中的错误事件 | 做饭糊锅了，要处理不能装没事 |

---

### 0.7 手把手总览：第一次跑通 Java

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | 安装 JDK 17，IDEA 里 Project SDK 选 17 | File → Project Structure 显示 17 | 见 FAQ Q2；装 JDK 后重启 IDEA |
| 2 | New → Project → Java，建 `src/Main.java` | 项目树有 `src/Main.java` | 确认没选错 Spring Initializr |
| 3 | 粘贴 Hello Java 代码，点绿色三角 Run | Run 窗口：`Hello Java`，`exit code 0` | 见 §38.1 `; expected` |
| 4 | 故意删掉分号再 Run | 红色波浪线 + `; expected` | 补回分号 |
| 5 | 新建 `User.java`，写类 + `main` 里 `new User()` | 能编译运行，控制台有输出 | 检查类名与文件名一致 |
| 6 | 给字段加 `private`，写 getter/setter | 外部通过 `setName` 赋值成功 | 见 §10.2 封装示例 |
| 7 | 完成分级练习「计算器」 | `add(10,3)` 输出 13.0 | 对照 §38 参考答案 |

---

## 本章与上一章的关系

00 路线图告诉你「学什么、按什么顺序、用什么工具」——这一章就是正式出发的第一步。

后端开发最终都要写 Java 代码，Spring Boot、MyBatis 本质上都是 Java 框架。把这一章打牢，后面写 Controller、Service 才不会被语法绊住。本章目标很具体：能在 IDEA 里创建项目、写类和方法、理解封装/继承/多态，并完成几个小练习。

---

## 1. 这份文档学什么

这一份不是路线图，而是你可以直接拿来学的内容。

学完这一份，你应该能做到：

- 看懂并写出基础 Java 代码
- 理解类、对象、封装、继承、多态
- 具备继续学习 Spring Boot 的语言基础

## 2. Java 是什么

Java 是一门面向对象编程语言，特点是：

- 语法相对规范
- 生态成熟
- 企业项目很多
- 后端岗位需求大

Java 后端开发里最常见的场景是：

- 写接口
- 处理业务逻辑
- 操作数据库
- 调用缓存和中间件

所以你先不要把 Java 理解成“为了考试学语法”，而要理解成“后端项目开发的主语言”。

## 3. 第一个 Java 程序

```java
public class Main {
    public static void main(String[] args) {
        System.out.println("Hello Java");
    }
}
```

### 代码解释

- `public class Main` 表示定义一个类，类名叫 `Main`
- `main` 是程序入口
- `System.out.println` 用来输出内容

你要先习惯 Java 的几个特点：

- 一切代码都写在类里
- 语句结尾一般有分号
- 大小写敏感

---

<!-- 修改说明: 新增 IDEA 手把手创建第一个 Java 项目 -->

## 3.1 手把手：IDEA 创建并运行第一个 Java 项目

### 第一步：创建项目

1. 打开 IDEA → **File → New → Project**
2. 左侧选 **Java**（不是 Spring Initializr，那是 04 章用的）
3. **Project SDK** 选 JDK 17 或 21；没有就点 **Download JDK**
4. 不要勾选「Create project from template」→ 点 **Create**
5. 弹出窗口时选 **Don't generate**，我们手动建类

### 第二步：项目目录结构

```text
HelloJava/
├── HelloJava.iml          ← IDEA 项目文件，不用管
└── src/
    └── Main.java          ← 你要创建的源文件
```

### 第三步：创建 Main.java

1. 右键项目名 → **New → Directory**，输入 `src`
2. 右键 `src` → **New → Java Class**，Name 填 `Main`
3. 粘贴代码：

```java
public class Main {
    public static void main(String[] args) {
        System.out.println("Hello Java");
    }
}
```

### 第四步：运行

1. 在 `Main.java` 编辑器里，行号旁会出现绿色三角
2. 点击三角 → **Run 'Main.main()'**
3. 底部 Run 窗口应显示：

```text
# 预期输出：
Hello Java

Process finished with exit code 0
```

### 第五步：故意制造编译错误（练排查）

把分号删掉再 Run：

```java
System.out.println("Hello Java")   // 故意去掉 ;
```

```text
# 预期报错：
; expected
```

把分号加回去即可。这类编译错误在 01 章非常常见，习惯看红色波浪线和底部错误提示。

---

## 4. 变量和数据类型

### 4.1 什么是变量

变量就是程序运行时用来存储数据的容器。

```java
int age = 18;
double price = 99.9;
String name = "Tom";
boolean pass = true;
```

### 4.2 基本数据类型

你先重点掌握这几个：

- `int`：整数
- `long`：更大的整数
- `double`：小数
- `boolean`：布尔值
- `char`：单个字符

### 4.3 引用数据类型

常见的引用类型有：

- `String`
- 数组
- 类对象

### 4.4 基本类型和引用类型的区别

先记住最核心的一点：

- 基本类型变量里直接存值
- 引用类型变量里存的是对象的引用

这句话你现在不必抠得特别深，但以后学对象和 JVM 时会用到。

## 5. 运算符

### 5.1 算术运算符

```java
int a = 10;
int b = 3;
System.out.println(a + b);
System.out.println(a - b);
System.out.println(a * b);
System.out.println(a / b);
System.out.println(a % b);
```

注意：

- `10 / 3` 在 `int` 场景下结果是 `3`
- `%` 表示取余，常用于判断奇偶和分组

### 5.2 比较运算符

```java
System.out.println(a > b);
System.out.println(a == b);
System.out.println(a != b);
```

### 5.3 逻辑运算符

```java
boolean c1 = true;
boolean c2 = false;
System.out.println(c1 && c2);
System.out.println(c1 || c2);
System.out.println(!c1);
```

## 6. 流程控制

### 6.1 if else

```java
int score = 85;
if (score >= 90) {
    System.out.println("优秀");
} else if (score >= 60) {
    System.out.println("及格");
} else {
    System.out.println("不及格");
}
```

适用场景：

- 按条件分支处理业务
- 判断参数是否合法
- 判断订单状态

### 6.2 switch

```java
int status = 1;
switch (status) {
    case 0:
        System.out.println("待支付");
        break;
    case 1:
        System.out.println("已支付");
        break;
    default:
        System.out.println("未知状态");
}
```

适合枚举型状态判断。

### 6.3 for 循环

```java
for (int i = 0; i < 5; i++) {
    System.out.println(i);
}
```

### 6.4 while 循环

```java
int i = 0;
while (i < 5) {
    System.out.println(i);
    i++;
}
```

## 7. 方法

方法就是把一段可以复用的逻辑封装起来。

```java
public class Main {
    public static int add(int a, int b) {
        return a + b;
    }

    public static void main(String[] args) {
        int result = add(3, 5);
        System.out.println(result);
    }
}
```

### 7.1 方法的组成

- 修饰符
- 返回值类型
- 方法名
- 参数列表
- 方法体

### 7.2 什么时候应该抽方法

当你发现一段逻辑：

- 会重复使用
- 代码太长
- 可读性太差

就应该考虑抽成方法。

## 8. 数组

数组是同一种类型数据的集合。

```java
int[] nums = {1, 2, 3, 4};
System.out.println(nums[0]);
System.out.println(nums.length);
```

### 8.1 常见遍历方式

```java
for (int i = 0; i < nums.length; i++) {
    System.out.println(nums[i]);
}
```

```java
for (int num : nums) {
    System.out.println(num);
}
```

### 8.2 数组特点

- 长度固定
- 元素类型统一
- 查询快

## 9. String 字符串

后端开发里，字符串非常常见。

```java
String name = "java";
System.out.println(name.length());
System.out.println(name.toUpperCase());
System.out.println(name.contains("av"));
```

### 9.1 字符串比较

这是初学者高频坑点。

```java
String a = "abc";
String b = new String("abc");
System.out.println(a == b);       // false
System.out.println(a.equals(b));  // true
```

记住：

- `==` 比较的是引用是否相同
- `equals` 比较的是内容是否相同

### 9.2 字符串拼接

```java
String s = "hello" + "world";
```

如果频繁拼接，优先考虑 `StringBuilder`，这一点在后面的集合与常用类文档里继续讲。

## 10. 面向对象基础

### 10.1 类和对象

类是模板，对象是实例。

```java
class User {
    String name;
    int age;

    void sayHello() {
        System.out.println("你好，我是" + name);
    }
}

public class Main {
    public static void main(String[] args) {
        User user = new User();
        user.name = "张三";
        user.age = 18;
        user.sayHello();
    }
}
```

### 10.2 封装

封装就是把数据和操作数据的方法放在一起，同时控制外部访问。

```java
class User {
    private String name;

    public void setName(String name) {
        if (name == null || name.isEmpty()) {
            throw new IllegalArgumentException("name 不能为空");
        }
        this.name = name;
    }

    public String getName() {
        return name;
    }
}
```

封装的意义：

- 保护数据
- 限制非法赋值
- 提升可维护性

### 10.3 构造方法

构造方法用于创建对象时初始化数据。

```java
class User {
    String name;
    int age;

    public User(String name, int age) {
        this.name = name;
        this.age = age;
    }
}
```

### 10.4 this 关键字

`this` 表示当前对象。

在构造器或成员方法中，常用来区分成员变量和参数：

```java
this.name = name;
```

## 11. 继承

继承表示子类可以复用父类的属性和行为。

```java
class Animal {
    public void eat() {
        System.out.println("吃东西");
    }
}

class Dog extends Animal {
    public void bark() {
        System.out.println("汪汪叫");
    }
}
```

### 11.1 继承的作用

- 代码复用
- 表达“是一个”的关系

### 11.2 super

`super` 用来访问父类内容。

```java
class Animal {
    String name;

    public Animal(String name) {
        this.name = name;
    }
}

class Dog extends Animal {
    public Dog(String name) {
        super(name);
    }
}
```

## 12. 多态

多态表示同一个父类引用，可以指向不同子类对象。

```java
class Animal {
    public void sound() {
        System.out.println("动物叫");
    }
}

class Dog extends Animal {
    @Override
    public void sound() {
        System.out.println("汪");
    }
}

class Cat extends Animal {
    @Override
    public void sound() {
        System.out.println("喵");
    }
}
```

```java
Animal a1 = new Dog();
Animal a2 = new Cat();
a1.sound();
a2.sound();
```

多态的价值：

- 调用更灵活
- 扩展更方便

这在框架开发和接口设计中非常重要。

## 13. 重载和重写

### 13.1 重载

同一个类中，方法名相同，参数列表不同。

```java
public int add(int a, int b) {
    return a + b;
}

public int add(int a, int b, int c) {
    return a + b + c;
}
```

### 13.2 重写

子类重写父类方法，方法名和参数一致。

```java
@Override
public void sound() {
    System.out.println("汪");
}
```

## 14. 抽象类和接口

### 14.1 抽象类

抽象类不能直接创建对象，适合提取共性。

```java
abstract class Payment {
    public abstract void pay();
}
```

### 14.2 接口

接口更强调规范。

```java
interface LoginService {
    void login(String username, String password);
}
```

### 14.3 怎么理解这两个东西

你现阶段先这样记：

- 抽象类：抽共性
- 接口：定规范

## 15. 异常处理

### 15.1 try-catch

```java
try {
    int x = 1 / 0;
} catch (Exception e) {
    System.out.println("出现异常：" + e.getMessage());
}
```

### 15.2 finally

`finally` 里的代码通常无论是否异常都会执行，常用于资源释放。

### 15.3 throw 和 throws

- `throw`：主动抛异常
- `throws`：声明方法可能抛出异常

## 16. 初学者常见错误

### 16.1 用 `==` 比较字符串

应该优先用 `equals`。

### 16.2 忘记初始化对象

```java
User user = null;
user.getName(); // 空指针异常
```

### 16.3 方法写得太长

如果一个方法几十行甚至上百行，很难维护。

### 16.4 所有字段都设成 public

这会破坏封装。

## 17. 这一章学完后你该练什么

建议你自己动手写：

1. 学生管理类
2. 用户注册信息类
3. 商品类和订单类
4. 用继承写支付方式类
5. 用接口写登录服务规范

## 18. 学完的标准

当你能做到下面这些，就说明这一份已经基本过关：

- 能看懂并写基础 Java 语法
- 能定义类和对象
- 知道封装、继承、多态的含义
- 能写方法、构造器、基础异常处理
- 能看懂后面 Spring Boot 教程中的普通 Java 代码

## 19. 访问修饰符

Java 中常见访问修饰符有四种：

- `public`
- `protected`
- 默认不写
- `private`

### 作用范围

- `public`：任何地方都能访问
- `protected`：同包下可访问，不同包中的子类也可访问
- 默认不写：同包可访问
- `private`：只能在当前类内部访问

你做项目时最常见的是：

- 类对外暴露的方法通常用 `public`
- 成员变量通常用 `private`

## 20. static 和 final

### 20.1 static

`static` 表示属于类，而不是属于某个对象。

```java
class User {
    public static String type = "NORMAL";
}
```

常见场景：

- 工具类方法
- 常量
- 统计类变量

### 20.2 final

`final` 的常见含义：

- 修饰变量：值不能再改
- 修饰方法：不能被子类重写
- 修饰类：不能被继承

```java
final class Constants {
    public static final String SUCCESS = "success";
}
```

## 21. package 和 import

### 21.1 package

用于声明类属于哪个包。

```java
package com.example.demo;
```

### 21.2 import

用于导入其他包下的类。

```java
import java.util.List;
```

为什么要分包：

- 让项目结构更清晰
- 避免类名冲突
- 方便维护

## 22. 枚举 enum

很多业务里都有固定状态，比如订单状态、支付状态、用户角色。

这类场景非常适合枚举。

```java
public enum OrderStatus {
    CREATED,
    PAID,
    CANCELED
}
```

为什么不用纯字符串：

- 更规范
- 可读性更好
- 更不容易写错

## 23. Object 类的基础认知

所有 Java 类默认都继承自 `Object`。

你至少要知道这些方法：

- `toString`
- `equals`
- `hashCode`

### 23.1 toString

用于把对象转成字符串表示。

### 23.2 equals

用于比较对象内容是否相同。

### 23.3 hashCode

和哈希容器有关，比如 `HashMap`、`HashSet`。

## 24. 重写 equals 和 hashCode

如果两个对象你认为“逻辑上相等”，通常也应该重写 `equals` 和 `hashCode`。

```java
import java.util.Objects;

class User {
    private Long id;
    private String username;

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        User user = (User) o;
        return Objects.equals(id, user.id) &&
                Objects.equals(username, user.username);
    }

    @Override
    public int hashCode() {
        return Objects.hash(id, username);
    }
}
```

## 25. 可变参数

```java
public static int sum(int... nums) {
    int total = 0;
    for (int num : nums) {
        total += num;
    }
    return total;
}
```

适合参数个数不固定的场景。

## 26. 代码规范基础

你从一开始就应该养成一些好习惯：

- 类名用大驼峰
- 方法名和变量名用小驼峰
- 常量名全大写加下划线
- 一个方法不要过长
- 一个类不要承担太多职责

## 27. Java 学习中的重要思维

### 27.1 先会写，再优化

初学时先保证代码能写对，再慢慢追求优雅。

### 27.2 先理解业务，再套语法

语法不是目的，解决问题才是目的。

### 27.3 多用对象建模

后端开发不是写脚本，而是组织对象、方法、流程和数据。

## 28. Java 参数传递到底是什么

这是非常高频的易混点。

Java 里只有一种传递方式：

- 值传递

### 基本类型传递

传的是值本身。

```java
public static void add(int x) {
    x++;
}

public static void main(String[] args) {
    int a = 1;
    add(a);
    System.out.println(a); // 还是 1
}
```

### 引用类型传递

传的是引用的副本，不是对象本身被“按引用传递”。

```java
class User {
    String name;
}

public static void change(User user) {
    user.name = "李四";
}
```

这里对象内容可能被改掉，是因为两个引用副本都指向同一个对象。

## 29. main 方法和命令行参数

Java 程序入口通常是：

```java
public static void main(String[] args)
```

其中 `args` 表示命令行参数。

```java
public static void main(String[] args) {
    for (String arg : args) {
        System.out.println(arg);
    }
}
```

你现在可能用得不多，但后面做脚本工具和启动参数理解时会碰到。

## 30. 注释和文档注释

### 单行注释

```java
// 这是单行注释
```

### 多行注释

```java
/*
  这是多行注释
*/
```

### 文档注释

```java
/**
 * 用户服务
 */
```

写注释的原则：

- 解释“为什么”
- 不要解释显而易见的“做了什么”

## 31. break、continue、return 的区别

### break

跳出当前循环。

### continue

跳过本次循环剩余部分，进入下一次循环。

### return

直接结束整个方法。

这三个关键字初学时很容易混。

## 32. 方法递归基础

递归就是方法调用自己。

```java
public static int sum(int n) {
    if (n == 1) {
        return 1;
    }
    return n + sum(n - 1);
}
```

递归要特别注意：

- 终止条件
- 层数过深可能导致栈溢出

## 33. null 的基础认知

`null` 表示没有指向任何对象。

```java
String name = null;
```

常见风险：

- 空指针异常 `NullPointerException`

常见防御思路：

- 使用前判空
- 明确返回值语义

## 34. IDE 和调试习惯

学习 Java 不只是写代码，还要学会调试。

你至少要会：

- 打断点
- 单步执行
- 观察变量值
- 看调用栈

调试能力会大幅提升你的学习效率。

## 35. 编译错误和运行错误

### 编译错误

代码还没运行就报错，比如：

- 分号漏写
- 类型不匹配
- 变量没定义

### 运行错误

程序能启动，但运行时出错，比如：

- 除零
- 空指针
- 数组越界

你要逐渐学会区分这两类问题。

## 36. 初学 Java 的建议练习题方向

建议你做一些非常基础但有价值的练习：

1. 成绩判断
2. 数组最大值最小值
3. 登录用户名密码判断
4. 面向对象版学生管理
5. 面向对象版订单类设计

这类练习虽然不高级，但很重要。

---

## 37. 学完标准

- 能独立写含 class、构造方法、getter/setter 的 Java 类
- 理解封装、继承、多态，能举出生活化例子
- 会使用 `if/for/while`、数组、基础异常 `try-catch`
- 理解 `==` 与 `equals` 区别，`String` 不可变
- 会在 IDEA 里运行、断点调试程序
- 完成至少：成绩判断、学生类、简单订单类 三个练习

---

## 38. 分级练习

**基础**：计算器类（加减乘除方法）  
**进阶**：`Student` + `Course` 多对多关系，用 `ArrayList` 管理  
**挑战**：银行账户类：存款、取款、转账，余额不足抛自定义异常

<!-- 修改说明: 新增分级练习参考答案 -->

### 参考答案

#### 基础：计算器类

```java
public class Calculator {

    public double add(double a, double b) {
        return a + b;
    }

    public double subtract(double a, double b) {
        return a - b;
    }

    public double multiply(double a, double b) {
        return a * b;
    }

    public double divide(double a, double b) {
        if (b == 0) {
            throw new IllegalArgumentException("除数不能为0");
        }
        return a / b;
    }

    public static void main(String[] args) {
        Calculator calc = new Calculator();
        System.out.println(calc.add(10, 3));       // 13.0
        System.out.println(calc.subtract(10, 3));  // 7.0
        System.out.println(calc.multiply(10, 3));  // 30.0
        System.out.println(calc.divide(10, 3));    // 3.333...
    }
}
```

#### 进阶：Student + Course 多对多

```java
import java.util.ArrayList;
import java.util.List;

public class Course {
    private String name;
    private int credit;

    public Course(String name, int credit) {
        this.name = name;
        this.credit = credit;
    }

    public String getName() { return name; }
    public int getCredit() { return credit; }
}

public class Student {
    private String name;
    private final List<Course> courses = new ArrayList<>();

    public Student(String name) {
        this.name = name;
    }

    public void enroll(Course course) {
        courses.add(course);
    }

    public void showCourses() {
        System.out.println(name + " 选修的课程：");
        for (Course c : courses) {
            System.out.println("  - " + c.getName() + "（" + c.getCredit() + " 学分）");
        }
    }

    public static void main(String[] args) {
        Student s = new Student("张三");
        s.enroll(new Course("Java 程序设计", 4));
        s.enroll(new Course("数据结构", 3));
        s.showCourses();
    }
}
```

#### 挑战：银行账户 + 自定义异常

```java
public class InsufficientBalanceException extends Exception {
    public InsufficientBalanceException(String message) {
        super(message);
    }
}

public class BankAccount {
    private String owner;
    private double balance;

    public BankAccount(String owner, double initialBalance) {
        this.owner = owner;
        this.balance = initialBalance;
    }

    public void deposit(double amount) {
        if (amount <= 0) throw new IllegalArgumentException("存款金额必须大于0");
        balance += amount;
    }

    public void withdraw(double amount) throws InsufficientBalanceException {
        if (amount <= 0) throw new IllegalArgumentException("取款金额必须大于0");
        if (balance < amount) throw new InsufficientBalanceException("余额不足，当前余额：" + balance);
        balance -= amount;
    }

    public void transfer(BankAccount target, double amount) throws InsufficientBalanceException {
        this.withdraw(amount);
        target.deposit(amount);
    }

    public double getBalance() { return balance; }

    public static void main(String[] args) {
        BankAccount a = new BankAccount("张三", 1000);
        BankAccount b = new BankAccount("李四", 500);
        try {
            a.transfer(b, 300);
            System.out.println("张三余额：" + a.getBalance());  // 700.0
            System.out.println("李四余额：" + b.getBalance());  // 800.0
            a.transfer(b, 1000);  // 应抛 InsufficientBalanceException
        } catch (InsufficientBalanceException e) {
            System.out.println("转账失败：" + e.getMessage());
        }
    }
}
```

#### 补充：成绩判断（§36 推荐练习）

```java
public class GradeChecker {

    public static String check(int score) {
        if (score < 0 || score > 100) {
            return "分数无效";
        }
        if (score >= 90) return "优秀";
        if (score >= 80) return "良好";
        if (score >= 60) return "及格";
        return "不及格";
    }

    public static void main(String[] args) {
        System.out.println(check(95));  // 优秀
        System.out.println(check(72));  // 及格
        System.out.println(check(45));  // 不及格
    }
}
```

#### 补充：订单类设计（§36 推荐练习）

```java
import java.time.LocalDateTime;

public class Order {
    private Long id;
    private String productName;
    private int quantity;
    private double unitPrice;
    private String status;
    private LocalDateTime createTime;

    public Order(Long id, String productName, int quantity, double unitPrice) {
        this.id = id;
        this.productName = productName;
        this.quantity = quantity;
        this.unitPrice = unitPrice;
        this.status = "CREATED";
        this.createTime = LocalDateTime.now();
    }

    public double getTotalAmount() {
        return quantity * unitPrice;
    }

    public void pay() {
        if ("PAID".equals(status)) {
            throw new IllegalStateException("订单已支付");
        }
        this.status = "PAID";
    }

    public Long getId() { return id; }
    public String getStatus() { return status; }

    public static void main(String[] args) {
        Order order = new Order(1L, "Java 书籍", 2, 99.0);
        System.out.println("总额：" + order.getTotalAmount());  // 198.0
        order.pay();
        System.out.println("状态：" + order.getStatus());       // PAID
    }
}
```

---

<!-- 修改说明: 新增常见报错与排查 -->

## 38.1 常见报错与排查

| 报错信息（关键词） | 可能原因 | 解决方案 |
|-------------------|---------|---------|
| `'java' 不是内部或外部命令` | JDK 未安装或未配环境变量 | 安装 JDK 17+；配置 `JAVA_HOME` 和 `Path` |
| `; expected` | 语句末尾漏分号 | 看 IDE 红色波浪线指向的行，补 `;` |
| `cannot find symbol` | 变量/类名拼错，或未 import | 检查大小写；确认类在同一个包或已 import |
| `class X is public, should be declared in a file named X.java` | public 类名和文件名不一致 | 文件名必须和 public 类名完全一致 |
| `Error: Could not find or load main class` | 运行了错误的类，或没编译 | 确认 Run 的是含 `main` 方法的类；Rebuild Project |
| `NullPointerException` | 对 null 对象调用了方法 | 用断点看哪个变量是 null；使用前加 null 判断 |

---

## 39. FAQ

**Q1：`public static void main` 每个字什么意思？**  
`public` 入口可被 JVM 调用；`static` 不创建对象即可执行；`void` 无返回值；`main` 固定方法名。

**Q2：JDK 装了但 IDEA 找不到 / 命令行 `java` 不可用？**  
确认 `JAVA_HOME` 指向 JDK 目录，`Path` 含 `%JAVA_HOME%\bin`。IDEA：File → Project Structure → SDK 选手动添加的 JDK 路径。

**Q3：为什么企业用 Java 8/17/21？**  
LTS 长期支持版；新项目建议 17+。本路线按 JDK 17 讲解。

**Q4：先学 Kotlin 可以吗？**  
本路线以 Java 为主，面试与存量项目仍以 Java 为绝大多数。

**Q5：类和对象到底怎么记？**  
类 = 蓝图，对象 = 按蓝图造出来的实物。`class User` 是图纸，`new User()` 是盖了一栋房。

**Q6：`==` 和 `equals` 对 String 为什么不一样？**  
`==` 比是不是同一个对象（门牌号）；`equals` 比内容是不是一样（户型面积是否相同）。比较字符串内容永远用 `equals`。

**Q7：什么时候用 `int`，什么时候用 `Integer`？**  
基本类型 `int` 不能为 null；包装类 `Integer` 可为 null，集合里只能放对象。日常局部变量用 `int` 即可。

**Q8：封装是不是多此一举？直接 `public` 字段不行吗？**  
不行。`public` 字段谁都能改，无法校验（如年龄 -1、姓名为空）。封装让你在 setter 里统一拦截非法值。

**Q9：继承和多态有什么用？**  
继承复用代码；多态让调用方只依赖父类/接口，换实现不用改调用代码——Spring 里大量接口注入就是多态。

**Q10：`abstract class` 和 `interface` 选哪个？**  
初学：有公共代码用抽象类抽共性；只定规范、多种实现用接口。Payment 场景常见 `interface PayService`，多种支付实现类。

**Q11：异常是该 catch 还是 throws？**  
能处理（如转友好提示）就 catch；调用方更合适处理就 throws。Controller 层往往 catch 或交给全局异常处理。

**Q12：方法写多长合适？**  
经验：一个方法干一件事，超过 30～40 行考虑拆分。后端 Service 方法过长是 code review 常见意见。

---

## 39.1 闭卷自测

> 先遮住答案，逐题口述或默写。

### 概念题（6 道）

1. 用「蓝图与房子」说明类与对象的关系；`new` 关键字做什么？
2. 基本类型 `int` 和引用类型 `String` 在变量里各存什么？
3. 封装的三层含义：private 字段、public 方法、setter 里校验——各解决什么问题？
4. 重载和重写（Override）分别是什么？各举一个方法名相同的例子。
5. 为什么说 Java 只有值传递？引用类型传递时为什么对象内容可能被改掉？
6. `try-catch-finally` 中 `finally` 一般用在什么场景？

### 动手题（2 道）

7. 写 `GradeChecker.check(int score)`：`>=90` 优秀，`>=60` 及格，否则不及格；非法分数返回「分数无效」。
8. 写 `BankAccount`：`deposit`、`withdraw`（余额不足抛异常）、`getBalance`；`main` 里演示转账。

### 综合题（2 道）

9. 下面代码输出什么？说明原因：`String a = "hi"; String b = "hi"; String c = new String("hi");` 比较 `a==b`、`a==c`、`a.equals(c)`。
10. 设计「支付方式」：`interface PayService { void pay(double amount); }`，写 `AliPayService` 和 `WechatPayService`，在 `main` 里用父接口引用调用——说明哪里体现了多态。

### 自测参考答案

1. 类是模板，对象是实例；`new` 在堆上创建对象并返回引用。
2. `int` 存数值本身；`String` 变量存的是对象的引用（地址）。
3. private 防外部乱改；public 方法控制访问入口；setter 校验保证数据合法。
4. 重载：同类同名不同参数列表，如 `add(int,int)` 与 `add(int,int,int)`；重写：子类改父类同名同参方法，如 `Dog.sound()`。
5. 传的是引用的副本，副本指向同一对象，所以方法里改对象字段会影响外面。
6. 释放资源（关流、关连接），无论是否异常常要执行。
7. 见 §38 成绩判断参考答案。
8. 见 §38 银行账户参考答案。
9. `a==b` true（字符串常量池）；`a==c` false（不同对象）；`a.equals(c)` true（内容相同）。
10. `PayService p = new AliPayService(); p.pay(100);` 换 Wechat 实现不改调用类型即多态。

---

## 39.2 费曼检验

**任务**：请在不看资料的情况下，用 **3 分钟** 向没学过编程的朋友解释「Java 面向对象是什么」。

**对照提纲**（说完后自检是否覆盖）：

1. **类与对象**：蓝图 vs 真房子；程序里用类描述「用户、订单」，用对象表示「张三这个用户」。
2. **封装、继承、多态**：门锁着（封装）；狗继承动物会吃（继承）；同一张动物登记表，实际猫狗叫法不同（多态）。
3. **和写网页的关系**：后面 Spring Boot 写接口，底层全是 Java 类；这一章是后面所有章的语法地基。

若朋友能说出「类是模板、对象是实例、封装保护数据」，本章核心已掌握。

---

## 39.3 本章与后续章节衔接速查

| 本章学会 | 04 章怎么用 | 05 章怎么用 |
|----------|-------------|-------------|
| `class` / `new` | `@Service` `@Controller` 都是类 | `Entity` 映射表行 |
| 封装 getter/setter | DTO 校验字段 | Entity 不直接返前端 |
| `interface` | 后面多态注入 | `Mapper` 是接口 |
| `try-catch` | 全局异常可替代散落 catch | SQL 异常转业务码 |
| `enum` 状态 | 订单状态枚举 | 与 `tinyint` 对应 |

### 39.3.1 第一个 User 类逐行读（OOP 入门）

```java
public class User {
    private String name;
    private int age;

    public User(String name, int age) {
        this.name = name;
        this.age = age;
    }

    public String getName() { return name; }
    public void setName(String name) {
        if (name == null || name.isEmpty()) {
            throw new IllegalArgumentException("name 不能为空");
        }
        this.name = name;
    }
}
```

| 行号/片段 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `private String name` | 外部不能直接 `user.name = ...` | 改 `public` 破坏封装，非法值拦不住 |
| `public User(...)` | 构造器：创建对象时强制初始化 | 删掉则只能用无参构造，字段可能空 |
| `this.name = name` | 区分成员变量与参数 | 写成 `name = name` 可能赋错 |
| `getName()` | 读数据的唯一出口 | 没有 getter 外部读不到 private |
| `setName` 里校验 | 封装的核心价值 | 去掉校验则空名也能入库 |

### 39.3.2 多态示例逐行读

```java
Animal a1 = new Dog();
Animal a2 = new Cat();
a1.sound();  // 汪
a2.sound();  // 喵
```

| 行号 | 含义 | 改错会怎样 |
|------|------|----------|
| `Animal a1 = new Dog()` | 父类引用指向子类对象 | 若 `Animal a1 = new Animal()` 就没有多态演示 |
| `a1.sound()` | 运行时看实际类型调用 | 编译期类型是 Animal，运行期是 Dog |

**动手验收清单**（扩充版学完可打勾）：

- [ ] 不看文档用「蓝图/房子」解释类与对象
- [ ] 10 分钟内写出带封装、构造器的 `User` 类并运行
- [ ] 向他人解释 `==` vs `equals` 对 String 的区别
- [ ] 完成银行账户分级练习并处理自定义异常
- [ ] 闭卷自测 ≥ 8/10

---

## 39.4 常见学习弯路与纠正

| 弯路 | 表现 | 纠正 |
|------|------|------|
| 只背语法不写代码 | 看会了写不出 `User` 类 | 每节至少写 10 行，§3.1 必须跟做 |
| 跳过 OOP 直接学 Spring | Controller 看不懂 | §10～§14 做学生/订单类练习 |
| 滥用 `public` 字段 | 到处 `user.name = ""` | 坚持 private + setter 校验 |
| 用 `==` 比字符串 | 判断登录名相等翻车 | 内容比较一律 `equals` |
| 怕异常从不 try-catch | 一运行就崩 | 先 `getMessage()` 打印，再学全局处理 |
| 方法堆成上百行 | 一个 `main` 包打天下 | 按 §7 抽方法，按职责拆类 |
| 忽略 IDEA 红色波浪线 | 运行才暴雷 | 保存后先看底部 Problems |
| 不做分级练习 | 自测题不会 | §38 三道至少做两道 |

---

<!-- 修改说明: 新增下一章预告 -->

## 下一章预告

这一章你掌握了 Java 语法和 OOP——能写类、方法、继承、异常处理。但真实后端代码里，你不会自己造轮子去管理数据，而是用 Java 标准库里的 **String、ArrayList、HashMap** 等。

下一章（02 Java 常用类、集合与泛型）就是日常开发用得最多的 API：什么时候用 ArrayList、什么时候用 HashMap、为什么金额要用 BigDecimal、泛型 `<T>` 是什么意思。把这些练熟，后面 Spring Boot 里接 JSON、查数据库返回 List，你才不会陌生。

---

*下一章：02 Java 常用类、集合与泛型*
