# 代码提取与批量验证

## 1. 作用

`verify_examples.ps1` 会递归扫描数据结构学习库中的 Markdown，提取所有标记为 `cpp`、`java`、`go` 的代码块，并在临时目录中逐个编译。

由于每个代码块都被约定为独立程序，因此验证器能及时发现：

- 缺少头文件、导包或 `package main`；
- Java 公共类名与文件名不一致；
- 语法错误和类型错误；
- Go 格式或编译问题；
- 可选运行阶段的非零退出码。

## 2. 环境要求

命令行中可直接使用：

```text
g++
javac
java
go
```

当前代码按 C++17、Java 17 和当前稳定 Go 语法编写。

## 3. 使用方法

在 PowerShell 中运行：

```powershell
cd F:\study\后端学习\数据结构\99-工具与验证

# 只编译
.\verify_examples.ps1

# 编译后运行所有示例
.\verify_examples.ps1 -Run
```

脚本只在系统临时目录创建中间文件，不会修改 Markdown 中的代码。

## 4. 文档作者约定

- C++ 围栏使用 `cpp`，程序必须包含 `main`；
- Java 围栏使用 `java`，入口类统一为 `public class Main`；
- Go 围栏使用 `go`，程序必须是 `package main`；
- 片段和伪代码使用 `text`，不要冒充可独立编译示例。
