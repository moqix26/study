# 系统设计资料验证工具

## 运行

~~~powershell
Set-Location F:\study\后端学习\系统设计
.\99-工具与验证\verify_system_design.ps1
.\99-工具与验证\verify_system_design.ps1 -Run
~~~

默认检查：

- 所有 Markdown 能被严格 UTF-8 解码；
- 每份 Markdown 在代码围栏外恰好有一个 H1；
- 代码围栏成对闭合；
- 本地 Markdown 链接目标存在；
- 不残留旧的扩写模板标记；
- Go 实验能够编译。

增加 Run 后，会逐个运行 Go 实验并验证断言。

该脚本验证结构、链接和可执行示例，不替代对中间件版本与架构结论的人工审查。

