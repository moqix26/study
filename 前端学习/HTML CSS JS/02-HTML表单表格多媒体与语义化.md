# HTML 表单、表格、多媒体与语义化

## 1. 这一份文档学什么

上一份主要讲 HTML 的基础结构和常见标签。

这一份继续往更实用的方向走，重点是：

- 表单
- 表格
- 多媒体
- 语义化标签
- 无障碍基础认知

这些内容在真实网页开发里非常常见。

## 2. 表单是什么

表单是网页中用来收集用户输入的区域。

比如这些场景：

- 登录
- 注册
- 搜索框
- 留言
- 提交订单信息

HTML 中表单的核心标签是：

- `form`

### 为什么表单对前端特别重要

因为用户和页面交互时，最常见的几种方式就是：

- 输入文字
- 选择选项
- 点击提交

而这些几乎都和表单有关。

你以后写的很多页面，本质上都离不开表单：

- 登录页
- 注册页
- 搜索框
- 意见反馈页
- 下单填写收货信息页

## 3. 最基础的表单

```html
<form>
  <input type="text" />
  <button type="submit">提交</button>
</form>
```

虽然很简单，但它已经具备了表单的最基础结构。

### 但这个例子为什么还不够“真实”

因为真正能提交有效数据的表单，往往至少还需要：

- `name`
- 提示文字
- 更明确的输入类型
- 校验

也就是说，表单不是“有个输入框和按钮就行了”。

## 4. `form` 标签常见属性

### 4.1 `action`

表示表单提交到哪里。

```html
<form action="/login"></form>
```

### 4.2 `method`

表示提交方式，最常见的是：

- `get`
- `post`

```html
<form action="/login" method="post"></form>
```

### 4.3 `target`

控制提交结果在哪里打开。

### 4.4 `autocomplete`

控制浏览器是否自动填充。

### 4.5 `enctype` 基础认知

这个属性在上传文件时很重要。

你现在先知道：

- 普通表单和文件上传表单的提交编码方式可能不同

后面学文件上传时会更容易理解。

## 5. `get` 和 `post` 的基础区别

### `get`

- 参数通常拼在 URL 后面
- 常用于查询类请求

### `post`

- 参数通常放在请求体中
- 常用于提交数据

你现在不用把它理解得太底层，但要知道：

- 搜索一般可以是 `get`
- 注册、登录、提交信息更常用 `post`

### 一个更容易记忆的方式

你可以先这样理解：

#### `get`

更像“我要拿点东西”

#### `post`

更像“我要交点东西过去”

这虽然不是严格技术定义，但对初学记忆很友好。

## 6. `input` 标签

这是表单里最常见的元素。

```html
<input type="text" />
```

### 为什么说 input 是表单的核心

因为绝大多数用户输入，最终都会通过各种 `input` 形态收集：

- 文本
- 密码
- 手机号
- 邮箱
- 单选
- 多选
- 上传文件

## 7. 常见 input 类型

### 7.1 文本输入框

```html
<input type="text" />
```

适合：

- 用户名
- 昵称
- 标题

### 文本输入框最常见的配套属性

- `placeholder`
- `name`
- `maxlength`
- `required`

### 7.2 密码框

```html
<input type="password" />
```

特点：

- 输入内容默认隐藏显示

### 这里要注意什么

隐藏显示不等于安全。

它只是：

- 页面视觉上不明文显示

真正安全还和：

- 网络传输
- 后端处理

有关。

### 7.3 数字输入

```html
<input type="number" />
```

### 使用注意

它适合数字输入，但不代表你可以完全不做校验。

因为：

- 用户输入行为
- 浏览器兼容差异

都可能带来问题。

### 7.4 邮箱输入

```html
<input type="email" />
```

### 7.5 电话输入

```html
<input type="tel" />
```

### 7.6 日期输入

```html
<input type="date" />
```

### 7.7 文件上传

```html
<input type="file" />
```

### 真实项目里文件上传还会考虑什么

- 限制文件类型
- 限制文件大小
- 是否允许多选

这些后面你会在 JavaScript 和后端接口里继续碰到。

### 7.8 单选框

```html
<input type="radio" name="gender" value="male" /> 男
<input type="radio" name="gender" value="female" /> 女
```

注意：

- 同一组单选按钮通常要有相同的 `name`

### 为什么一定要同名

因为浏览器需要知道：

- 这些按钮属于同一组

只有同一组，才会实现“多选一”。

### 7.9 复选框

```html
<input type="checkbox" value="java" /> Java
<input type="checkbox" value="css" /> CSS
```

### 和单选的区别怎么记

#### radio

- 一组选一个

#### checkbox

- 一组选多个

### 7.10 提交按钮

```html
<input type="submit" value="提交" />
```

### 7.11 重置按钮

```html
<input type="reset" value="重置" />
```

## 8. input 常见属性

### `name`

表单提交时非常重要。

```html
<input type="text" name="username" />
```

### 为什么它重要

因为提交给服务器时，后端往往识别的是：

- 字段名
- 字段值

而 `name` 就是字段名的重要来源之一。

如果你只写了输入框，却没写 `name`，后端常常就拿不到你想要的数据字段。

### `value`

输入框默认值。

```html
<input type="text" value="默认内容" />
```

### 注意区分 value 和 placeholder

这是初学者很容易混的点。

#### `value`

- 真正的值
- 会被提交

#### `placeholder`

- 只是提示文字
- 用户没输入时显示

### `placeholder`

占位提示文字。

```html
<input type="text" placeholder="请输入用户名" />
```

### `required`

必填。

```html
<input type="text" required />
```

### `disabled`

禁用。

### `readonly`

只读。

### `maxlength`

最大输入长度。

### `disabled` 和 `readonly` 的区别

#### `disabled`

- 不能输入
- 通常也不能参与正常交互

#### `readonly`

- 内容不能改
- 但仍然可聚焦、可查看

它们的语义不同，不要混着用。

## 9. `label` 标签

这是表单里非常重要，但很多初学者会忽略的标签。

```html
<label for="username">用户名</label>
<input id="username" type="text" />
```

作用：

- 提高可用性
- 提高无障碍体验
- 点击文字也能聚焦输入框

### 为什么 label 对初学者也必须重视

因为它不是锦上添花，而是一个真正能提升体验的标签。

比如在移动端：

- 点小输入框不方便
- 但点文字更容易

有了 `label`，体验会更自然。

## 10. `textarea`

用于多行文本输入。

```html
<textarea rows="5" cols="30"></textarea>
```

常见场景：

- 留言
- 备注
- 简介

### `textarea` 和普通文本框的区别

你可以这样记：

- 单行输入：`input`
- 多行输入：`textarea`

## 11. `select`、`option`

下拉选择框。

```html
<select name="city">
  <option value="beijing">北京</option>
  <option value="shanghai">上海</option>
  <option value="guangzhou">广州</option>
</select>
```

### 常见属性

- `selected`
- `disabled`

### 下拉框常见场景

- 选择城市
- 选择年级
- 选择性别
- 选择分类

它非常适合“选项固定且不太多”的场景。

## 12. `button`

```html
<button type="button">普通按钮</button>
<button type="submit">提交</button>
<button type="reset">重置</button>
```

一般来说，实际开发中很多人更喜欢用：

- `button`

因为：

- 可读性更好
- 可包裹更多内容

### `button` 和 `input type="submit"` 怎么选

初学阶段你可以先记：

- 简单可用都行
- 真实项目里通常 `button` 更灵活

## 13. 表单分组：`fieldset` 和 `legend`

```html
<fieldset>
  <legend>注册信息</legend>
  <input type="text" placeholder="用户名" />
</fieldset>
```

作用：

- 把表单内容做逻辑分组

### 什么时候你会想用分组

例如一个注册页可能有：

- 账号信息
- 基本资料
- 联系方式

这时分组会让结构更清楚。

## 14. 基础登录表单示例

```html
<form action="/login" method="post">
  <div>
    <label for="username">用户名</label>
    <input id="username" name="username" type="text" placeholder="请输入用户名" required />
  </div>

  <div>
    <label for="password">密码</label>
    <input id="password" name="password" type="password" placeholder="请输入密码" required />
  </div>

  <button type="submit">登录</button>
</form>
```

### 这个示例里有哪些值得你注意的小点

1. 每个输入框都有 `label`
2. 每个输入框都有 `name`
3. 密码框用了 `type="password"`
4. 加了 `placeholder`
5. 加了 `required`

这就比“只有两个裸 input”的表单更像真实页面。

## 15. 表格 `table`

表格适合展示规则数据。

例如：

- 学生成绩表
- 商品价格表
- 订单列表

### 为什么表格适合展示“规则数据”

因为表格天然强调的是：

- 行和列
- 每一列代表一类字段
- 每一行代表一条记录

这和数据库表的展示很像。

```html
<table>
  <tr>
    <th>姓名</th>
    <th>年龄</th>
  </tr>
  <tr>
    <td>张三</td>
    <td>18</td>
  </tr>
</table>
```

## 16. 表格常见标签

### `table`

整个表格。

### `tr`

表格中的一行。

### `th`

表头单元格。

### `td`

普通单元格。

### 一个很容易理解的映射方式

你可以把表格想成：

- `table`：整张表
- `tr`：一行
- `th`：标题单元格
- `td`：数据单元格

## 17. 更规范的表格结构

```html
<table>
  <caption>学生成绩表</caption>
  <thead>
    <tr>
      <th>姓名</th>
      <th>语文</th>
      <th>数学</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>张三</td>
      <td>90</td>
      <td>95</td>
    </tr>
  </tbody>
</table>
```

### 为什么这样更好

- 结构更清晰
- 语义更明确

### `caption` 有什么用

它是表格标题，不是必须写，但在说明型表格中很有帮助。

例如：

- 成绩表
- 课程表
- 商品价格对照表

## 18. 合并单元格

### 横向合并 `colspan`

```html
<td colspan="2">合并两列</td>
```

### 纵向合并 `rowspan`

```html
<td rowspan="2">合并两行</td>
```

## 19. 表格使用注意

表格适合：

- 展示规则数据

不适合：

- 整个页面布局

### 为什么不推荐用 table 布局整页

因为：

- 可维护性差
- 响应式适配不灵活
- 语义不合适

现代布局应该交给 CSS。

现在网页布局一般不用 table，而用：

- Flex
- Grid

## 20. 多媒体标签

### 20.1 音频 `audio`

```html
<audio controls>
  <source src="music.mp3" type="audio/mpeg" />
</audio>
```

### 20.2 视频 `video`

```html
<video controls width="400">
  <source src="movie.mp4" type="video/mp4" />
</video>
```

### 常见属性

- `controls`
- `autoplay`
- `loop`
- `muted`
- `poster`

### 多媒体标签为什么重要

因为网页不只是文字和图片，很多页面还会有：

- 教学视频
- 产品演示视频
- 音频播放

所以你至少要认识这些标签。

## 21. 语义化标签

HTML5 提供了很多更有语义的标签。

例如：

- `header`
- `nav`
- `main`
- `section`
- `article`
- `aside`
- `footer`

### 为什么 HTML5 会增加这些标签

因为以前很多页面大量使用：

- `div`

浏览器看得懂，但人和工具不容易快速理解结构角色。

有了这些语义化标签后：

- 页面更好读
- 更容易维护

## 22. 为什么语义化很重要

语义化不是“看起来高级”，而是有真实价值：

- 结构更清晰
- 更利于维护
- 更利于 SEO
- 更利于无障碍访问

### 再补一个更现实的理解

如果你以后接手别人的页面代码：

- 全是 `div` 会很难读
- 有 `header`、`nav`、`main`、`footer` 会更容易一眼看懂结构

## 23. 语义化布局示例

```html
<header>网站头部</header>
<nav>导航栏</nav>
<main>
  <section>
    <article>文章内容</article>
  </section>
  <aside>侧边栏</aside>
</main>
<footer>页脚</footer>
```

## 24. `div` 和语义化标签的区别

不是说不能用 `div`，而是：

- `div` 更通用
- 语义化标签能表达结构角色

正确思路是：

- 有明确语义时优先用语义化标签
- 通用容器再用 `div`

### 一句话记忆

- `div` 不是错
- 但不应该在所有地方都代替语义化标签

## 25. 无障碍基础认知

前端不是只给视觉正常用户看的。

还要考虑：

- 屏幕阅读器用户
- 键盘操作用户

你现在至少要知道这些基础点：

- 图片写 `alt`
- 表单配 `label`
- 页面结构语义清晰

### 为什么无障碍不是“可有可无”

因为网页不是只给我们这种正常看屏幕、正常用鼠标的人看的。

有些用户可能：

- 依赖屏幕阅读器
- 依赖键盘操作
- 看不清图片内容

所以良好的结构和说明文本真的有价值。

## 26. data-* 自定义属性

```html
<div data-id="1001" data-role="card"></div>
```

适合：

- 存放页面相关自定义数据

### 为什么不用乱造属性名

HTML 提供了 `data-*` 这种规范方式，就是为了安全、清晰地放自定义数据。

后面 JavaScript 读取元素数据时会很方便。

后面 JavaScript 操作 DOM 时很常见。

## 27. iframe 基础认知

```html
<iframe src="https://example.com"></iframe>
```

作用：

- 在当前页面嵌入另一个页面

但现在使用时要谨慎，因为会涉及：

- 安全
- 性能
- 交互复杂度

### 初学阶段怎么对待 iframe

你先认识它即可，不需要一开始就依赖它做主要功能。

## 28. 初学者常见错误

### 28.1 表单项没有 `name`

提交时后端可能收不到值。

### 28.2 标签有视觉效果却没语义

长期会让页面结构变乱。

### 28.3 用 table 做整页布局

现在不是推荐做法。

### 28.4 不写 `label`

会降低可用性。

### 28.5 把所有输入都写成 `type="text"`

这样会损失很多标签语义和浏览器自带能力。

例如：

- 密码就该用 `password`
- 邮箱就该用 `email`
- 文件就该用 `file`

## 29. 一个更完整的注册表示例

```html
<form action="/register" method="post">
  <fieldset>
    <legend>账号信息</legend>

    <div>
      <label for="reg-username">用户名</label>
      <input id="reg-username" name="username" type="text" placeholder="请输入用户名" required maxlength="20" />
    </div>

    <div>
      <label for="reg-password">密码</label>
      <input id="reg-password" name="password" type="password" placeholder="请输入密码" required />
    </div>
  </fieldset>

  <fieldset>
    <legend>个人信息</legend>

    <div>
      <span>性别</span>
      <label><input type="radio" name="gender" value="male" /> 男</label>
      <label><input type="radio" name="gender" value="female" /> 女</label>
    </div>

    <div>
      <label>
        <input type="checkbox" name="agree" required />
        我已阅读并同意协议
      </label>
    </div>
  </fieldset>

  <button type="submit">注册</button>
</form>
```

### 你应该从这个例子中看懂什么

1. 表单不是单个输入框，而是一整套结构
2. 不同输入类型有不同语义
3. 分组能让表单更清楚
4. `label`、`name`、`required` 都很重要

## 30. 练习建议

建议你自己做：

1. 登录页表单
2. 注册页表单
3. 联系我们页面
4. 学生成绩表
5. 含视频和音频的演示页
6. 语义化结构的博客首页骨架

### 更好的练习方式

建议你不要只做“静态抄写”，而是这样练：

1. 先照着写登录表单
2. 再自己独立写注册表单
3. 再自己增加新字段
4. 再尝试做一个更完整的资料填写页

---

## 31. 完整实战：无障碍友好的登录表单

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>登录</title>
</head>
<body>
  <main>
    <h1>用户登录</h1>
    <form action="/api/login" method="post" autocomplete="on">
      <p>
        <label for="email">邮箱</label>
        <input
          id="email"
          name="email"
          type="email"
          required
          placeholder="you@example.com"
          autocomplete="email"
        />
      </p>
      <p>
        <label for="password">密码</label>
        <input
          id="password"
          name="password"
          type="password"
          required
          minlength="6"
          autocomplete="current-password"
        />
      </p>
      <p>
        <label>
          <input type="checkbox" name="remember" value="1" />
          记住我
        </label>
      </p>
      <button type="submit">登录</button>
      <button type="reset">清空</button>
    </form>
  </main>
</body>
</html>
```

### 这个示例你要看懂什么

1. **`label for` 和 `input id` 成对**：点击文字也能聚焦输入框（无障碍）
2. **`name` 必须有**：后端才能收到字段名
3. **`type="email"`**：移动端弹出合适键盘；浏览器可做格式校验
4. **`required` / `minlength`**：HTML5 原生校验
5. **`autocomplete`**：帮助浏览器正确自动填充
6. **语义标签 `main`、`h1`**：结构清晰，利于 SEO 和读屏器

---

## 32. 语义化页面骨架模板

```html
<body>
  <header>
    <nav aria-label="主导航">...</nav>
  </header>
  <main>
    <article>
      <header><h1>文章标题</h1></header>
      <section>...</section>
    </article>
    <aside>侧边栏</aside>
  </main>
  <footer>版权信息</footer>
</body>
```

| 标签 | 含义 | 不要用 div 替代的场景 |
|------|------|------------------------|
| `header` | 页头或区块头 | 站点顶部、文章标题区 |
| `nav` | 导航链接组 | 主导航、面包屑 |
| `main` | 页面唯一主内容 | 正文区域（每页一个） |
| `article` | 独立完整内容 | 博客帖、新闻 |
| `section` | 主题分组 | 章节、功能区块 |
| `aside` | 侧边补充 | 相关链接、广告 |
| `footer` | 页脚或区块脚 | 版权、联系方式 |

---

## 33. 表单属性速查表

| 属性 | 作用 | 示例 |
|------|------|------|
| `name` | 提交时的字段名 | `name="username"` |
| `value` | 默认值/提交值 | checkbox、radio |
| `placeholder` | 占位提示（非标签） | 不能替代 label |
| `required` | 必填 | 提交前校验 |
| `disabled` | 禁用，不提交 | 灰色不可点 |
| `readonly` | 只读，会提交 | 展示用 |
| `maxlength` | 最大字符数 | 配合字数统计 |
| `pattern` | 正则校验 | 手机号格式 |
| `autofocus` | 自动聚焦 | 首页搜索框 |

---

## 34. 分级练习

**基础**：注册表单（用户名、密码、确认密码、协议勾选）  
**进阶**：带 `fieldset` + `legend` 的分组问卷页  
**挑战**：语义化写一个博客文章页（header/nav/main/article/footer）

---

## 35. FAQ

**Q：`button` 默认 type 是什么？**  
`submit`。非提交按钮务必写 `type="button"`，否则会触发表单提交。

**Q：表格还能用来做布局吗？**  
历史上有，现在**禁止**用 table 做页面布局，只用 CSS 布局。

**Q：`alt` 和 `title` 区别？**  
`alt` 是图片替代文字（必写）；`title` 是悬停提示（可选）。

---

---

## 37. input 类型完整速查表

| type 值 | 外观 | 用途 | 移动端键盘 |
|----------|------|------|-----------|
| `text` | 单行输入框 | 用户名、搜索 | 普通键盘 |
| `password` | 密码框（点号） | 密码输入 | 普通键盘 |
| `email` | 邮箱输入 | 邮箱地址 | 邮箱键盘（含 @） |
| `tel` | 电话输入 | 手机号 | 电话数字键盘 |
| `number` | 数字输入 | 数量、价格 | 数字键盘 |
| `url` | URL 输入 | 网址 | URL 键盘 |
| `search` | 搜索框 | 搜索关键词 | 搜索键盘（含回车=搜索） |
| `date` | 日期选择器 | 日期 | 日期选择器 |
| `time` | 时间选择器 | 时间 | 时间选择器 |
| `datetime-local` | 日期+时间 | 日期时间 | 日期时间选择器 |
| `month` | 月份选择器 | 月份 | 月份选择器 |
| `week` | 周选择器 | 周次 | 周选择器 |
| `color` | 颜色选择器 | 颜色 | — |
| `range` | 滑块 | 音量、价格区间 | — |
| `file` | 文件选择 | 上传文件 | — |
| `checkbox` | 复选框 | 多选 | — |
| `radio` | 单选框 | 单选 | — |
| `hidden` | 不可见 | 传递不可见数据 | — |
| `submit` | 提交按钮 | 表单提交 | — |
| `reset` | 重置按钮 | 清空表单 | — |
| `image` | 图片提交按钮 | 图片提交 | — |

### 什么时候用哪个 type

```html
<!-- ✅ 根据不同场景选对 type -->
<input type="email" placeholder="邮箱" />       <!-- 移动端弹出含@的键盘，浏览器可校验格式 -->
<input type="tel" placeholder="手机号" />        <!-- 移动端弹出数字键盘 -->
<input type="number" min="1" max="99" />        <!-- 数字键盘 + 范围限制 -->
<input type="url" placeholder="个人网站" />      <!-- URL 键盘 -->
<input type="search" placeholder="搜索..." />    <!-- 搜索键盘，回车键显示"搜索" -->

<!-- ⚠️ 不推荐：所有都用 type="text"（浪费浏览器能力）-->
```

---

## 38. HTML5 原生表单校验详解

HTML5 提供了一套无需 JS 的校验能力。用于快速基本校验，JS 做补充。

### 38.1 必填 `required`

```html
<input type="text" required />
<!-- 提交时如果为空，浏览器会阻止并显示提示 -->
```

### 38.2 长度限制 `minlength` / `maxlength`

```html
<input type="text" minlength="2" maxlength="20" />
<!-- 用户名 2~20 个字符 -->
```

### 38.3 数值范围 `min` / `max` / `step`

```html
<input type="number" min="1" max="99" step="1" />
<input type="range" min="0" max="100" step="5" value="50" />
```

### 38.4 正则表达式 `pattern`

```html
<!-- 手机号：1 开头 + 10 位数字 -->
<input type="tel" pattern="1[3-9]\d{9}" title="请输入有效手机号" />

<!-- 6 位数字验证码 -->
<input type="text" pattern="\d{6}" title="请输入 6 位数字验证码" />
```

### 38.5 完整示例：带原生校验的注册表单

```html
<form action="/register" method="post">
  <p>
    <label for="reg-user">用户名</label>
    <input id="reg-user" name="username" type="text"
           required minlength="2" maxlength="20"
           placeholder="2~20 个字符" />
  </p>
  <p>
    <label for="reg-email">邮箱</label>
    <input id="reg-email" name="email" type="email" required />
  </p>
  <p>
    <label for="reg-phone">手机号</label>
    <input id="reg-phone" name="phone" type="tel"
           pattern="1[3-9]\d{9}"
           title="请输入 11 位有效手机号" />
  </p>
  <p>
    <label for="reg-age">年龄</label>
    <input id="reg-age" name="age" type="number" min="1" max="120" />
  </p>
  <button type="submit">注册</button>
</form>
```

### 38.6 自定义校验信息 `setCustomValidity`

```js
// 原生校验的提示文字不好看？可以用 JS 自定义
const input = document.getElementById("reg-phone");
input.addEventListener("input", () => {
  if (!input.checkValidity()) {
    input.setCustomValidity("请输入 11 位有效手机号");
  } else {
    input.setCustomValidity(""); // 清除自定义错误
  }
});
```

---

## 39. datalist + input 联想输入

实现带下拉建议的输入框，无需 JS 库：

```html
<label for="city">城市</label>
<input id="city" name="city" list="city-list" placeholder="输入或选择城市" />
<datalist id="city-list">
  <option value="北京" />
  <option value="上海" />
  <option value="广州" />
  <option value="深圳" />
  <option value="杭州" />
  <option value="成都" />
</datalist>
```

- `input` 的 `list` 属性指向 `datalist` 的 `id`
- 用户既可以自己输入，也可以从下拉列表中选择
- 适合：城市、标签、常见选项（选项不多时）

---

## 40. 常见表单设计模式

### 40.1 两栏表单（label 左 + input 右）

```html
<form style="max-width: 500px;">
  <div style="display: flex; align-items: center; margin-bottom: 12px;">
    <label for="name" style="width: 100px; text-align: right; margin-right: 12px;">姓名</label>
    <input id="name" type="text" style="flex: 1;" />
  </div>
  <div style="display: flex; align-items: center; margin-bottom: 12px;">
    <label for="mail" style="width: 100px; text-align: right; margin-right: 12px;">邮箱</label>
    <input id="mail" type="email" style="flex: 1;" />
  </div>
</form>
```

### 40.2 行内表单（搜索栏）

```html
<form style="display: flex; gap: 8px;">
  <input type="search" placeholder="搜索..." style="flex: 1;" />
  <button type="submit">搜索</button>
</form>
```

### 40.3 多步骤表单（分步指示）

```html
<!-- 步骤指示器 -->
<ol style="display: flex; gap: 16px; list-style: none; margin-bottom: 24px;">
  <li style="color: #6366f1; font-weight: bold;">① 账号信息</li>
  <li style="color: #94a3b8;">② 个人资料</li>
  <li style="color: #94a3b8;">③ 确认提交</li>
</ol>
<!-- 每步一个 fieldset，JS 控制显示 -->
```

---

## 41. 表格复杂实战：课程表（合并单元格）

```html
<table border="1" style="border-collapse: collapse; width: 100%; text-align: center;">
  <caption>📅 2026 年春季学期课程表</caption>
  <thead>
    <tr>
      <th>时间</th>
      <th>周一</th>
      <th>周二</th>
      <th>周三</th>
      <th>周四</th>
      <th>周五</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>08:00-09:30</td>
      <td>数学</td>
      <td>英语</td>
      <td rowspan="2">计算机<br/>实验课</td>
      <td>物理</td>
      <td>数学</td>
    </tr>
    <tr>
      <td>10:00-11:30</td>
      <td>英语</td>
      <td>物理</td>
      <!-- 这里被 rowspan 占了 -->
      <td>体育</td>
      <td>英语</td>
    </tr>
    <tr>
      <td>14:00-15:30</td>
      <td colspan="5">选修课（自选）</td>
    </tr>
  </tbody>
</table>
```

---

## 42. 无障碍表单完整 Checklist

- [ ] 每个 `input` 都有对应的 `<label>`（通过 `for` + `id` 关联）
- [ ] 图片有 `alt` 文本
- [ ] 必填字段有明确标识（`required` + 视觉提示如红色 `*`）
- [ ] 错误信息有语义关联（用 `aria-describedby` 连接 input 和 error span）
- [ ] 整个页面有语义化结构（`main`、`h1` 等）
- [ ] 所有交互元素可用键盘操作（Tab 键可聚焦，Enter/Space 可激活）
- [ ] 焦点顺序合理（tabindex 不乱设）
- [ ] 自定义交互元素有相应 ARIA 属性

```html
<!-- 无障碍示例：aria-describedby 关联错误信息 -->
<label for="email">邮箱</label>
<input id="email" type="email" required aria-describedby="email-hint email-error" />
<span id="email-hint">请输入常用邮箱</span>
<span id="email-error" role="alert" style="color: red;"></span>
```

---

## 43. iframe 安全属性

```html
<!-- 基础 -->
<iframe src="https://example.com" width="600" height="400"></iframe>

<!-- 安全增强 -->
<iframe
  src="https://example.com"
  sandbox="allow-scripts allow-same-origin"
  loading="lazy"
  referrerpolicy="no-referrer"
  title="嵌入内容描述"
></iframe>
```

| 属性 | 说明 |
|------|------|
| `sandbox` | 限制 iframe 能力：`allow-scripts`（允许JS）、`allow-same-origin`（允许同源）、`allow-forms`（允许表单）等 |
| `loading="lazy"` | 延迟加载（首屏之外的不立即请求） |
| `referrerpolicy` | 控制是否发送 Referer：`no-referrer`/`origin`/`strict-origin` |
| `title` | **必须写**，描述 iframe 内容（无障碍要求） |

---

## 44. 学完标准（深度版）

- [ ] 会写带 label、校验、多种 input 类型的完整表单
- [ ] 知道所有 input type 并会根据场景选择（不只用 `type="text"`）
- [ ] 能用 HTML5 原生校验（required / pattern / min / max / minlength）
- [ ] 理解 `name`、`method`、`action` 与提交的关系
- [ ] 知道 button 默认 type=submit，非提交按钮必须写 `type="button"`
- [ ] 会写合并单元格的表格（colspan / rowspan），知道表格只用于数据展示
- [ ] 会使用 `video`/`audio` 基础标签
- [ ] 能用语义化标签搭页面骨架，具备基本无障碍意识
- [ ] 会使用 `datalist` 做联想输入
- [ ] 能独立完成登录/注册页完整 HTML 结构
