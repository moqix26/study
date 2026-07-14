# WSL2、虚拟机与云服务器差异

学习 Linux 时最容易出现的误区，是把三个环境当成完全相同的服务器。它们共享大量命令，但网络、启动系统、磁盘、权限和故障边界并不相同。

## 1. 推荐分工

| 环境 | 最适合做什么 | 不适合拿来证明什么 |
|------|--------------|--------------------|
| PowerShell | Windows 文件、API 测试、SSH 客户端 | Linux 权限和 systemd |
| WSL2 | 每日命令、Git、Go 编译、Shell、文本工具 | 完整模拟云防火墙和公网部署 |
| Ubuntu VM | systemd、独立网卡、磁盘、用户、快照和破坏性实验 | 真实公网安全组与域名 TLS |
| 云服务器 | SSH、域名、TLS、安全组、真实上线和备份 | 随意执行危险实验 |

对当前 Go 学习路线，推荐：

```text
现在：Windows + PowerShell 写代码
需要日常 Linux 命令时：WSL2
进入短链部署周：Ubuntu VM 完整演练一次
准备简历 demo：再用云服务器上线
```

不需要先连续学习数周 Linux 才能写 Gin。

## 2. WSL2

优点：

- 与 Windows 文件和编辑器集成方便。
- 启动快，适合 `grep`、`sed`、Shell、Go 工具链。
- 现代 WSL2 可以启用 systemd，但行为仍受 WSL 生命周期和 Windows 网络影响。

注意：

- Linux 项目应放在 WSL 自己的 ext4 文件系统，例如 `~/projects`；大量小文件放 `/mnt/c`、`/mnt/f` 可能更慢，并有权限语义差异。
- Windows 和 WSL 各有 localhost 转发与防火墙规则，版本更新后网络模式可能变化。
- WSL 关闭后，长期服务是否继续运行不能直接类比云服务器。
- Docker Desktop 的 WSL 后端与“在云服务器直接安装 Docker Engine”不是同一运维边界。

常用检查：

```powershell
wsl --status
wsl --list --verbose
wsl --shutdown
```

```bash
cat /etc/wsl.conf
systemctl is-system-running
ip -brief address
```

## 3. Ubuntu 虚拟机

虚拟机拥有独立 Guest OS、虚拟磁盘和虚拟网卡，更适合练：

- systemd 开机启动。
- 用户、组、sudo、文件权限。
- NAT、桥接、Host-only 网卡。
- 磁盘扩容、inode、快照恢复。
- 防火墙、SSH、Nginx、Docker Engine。

### NAT 的正确理解

VMware NAT 通常允许：

- Guest 主动访问互联网。
- 宿主机通过 VMnet8 私网地址访问 Guest。

端口转发主要用于“宿主机之外的设备通过宿主机端口进入 NAT Guest”等场景，不应简单写成“宿主机访问 NAT VM 必须端口转发”。实际行为还受 VMware 配置、Windows 防火墙和 Guest 防火墙影响，应以 `ip address`、`Test-NetConnection` 和 `ss` 实测。

### 桥接

Guest 获得与局域网同网段地址，像另一台真实机器。它方便局域网访问，但公共 Wi-Fi、校园网和企业网络可能限制额外 MAC 或客户端隔离，因此桥接不一定总能使用。

### Host-only

Guest 与宿主机在私有网段互通，默认不直接访问互联网。适合安全的隔离实验，也可与第二块 NAT 网卡组合：一块出网，一块专供宿主机访问。

## 4. 云服务器

云服务器比本地 VM 多出几层：

```text
公网 DNS
  → 云安全组 / 云防火墙
  → 云主机公网或弹性 IP
  → Ubuntu ufw / nftables
  → Nginx
  → Go 服务
```

因此“本机 curl 正常但公网不通”时，必须检查安全组，而不仅是 ufw。

云服务器还需要面对：

- 公网扫描和暴力登录。
- 动态公网 IP、域名解析与证书续期。
- 云盘快照不等于数据库一致性备份。
- 费用、流量和误删资源。
- 控制台/VNC/救援模式，这是 SSH 配错后的最后入口。

## 5. 路径与换行

Windows：

```text
F:\study\project
CRLF 常见
```

Linux：

```text
/home/honor/project
LF 常见
```

Shell 脚本出现以下错误时，通常是 CRLF：

```text
/usr/bin/env: 'bash\r': No such file or directory
```

修复：

```bash
sed -i 's/\r$//' deploy.sh
```

Git 项目应明确 `.gitattributes`，例如：

```gitattributes
*.sh text eol=lf
*.service text eol=lf
*.conf text eol=lf
*.ps1 text eol=crlf
```

## 6. 端口绑定矩阵

| 运行位置 | 应用监听 | 外部访问方式 |
|----------|----------|--------------|
| Windows 本地 Go | `127.0.0.1:8080` | PowerShell 直连 |
| WSL2 开发 | `127.0.0.1:8080` 或按跨环境需求调整 | Windows localhost 转发或 WSL IP |
| VM 内直接实验 | VM 私网 IP 或 `0.0.0.0:8080`，防火墙限宿主机 | 宿主机访问 VM IP |
| 云主机同机 Nginx | `127.0.0.1:8080` | 公网只访问 Nginx 443 |
| Docker 容器内 Go | `0.0.0.0:8080` | 宿主机映射 `127.0.0.1:8080:8080` 或内部 Nginx 网络 |

不要把“容器内必须监听 0.0.0.0”误写成“所有生产 Go 服务都必须监听 0.0.0.0”。

## 7. 什么时候算真正掌握

至少完成以下闭环：

1. 在 WSL2 用 Shell 和 Go 工具完成日常开发操作。
2. 在 VM 中创建专用用户，通过 systemd 运行 Go 服务。
3. 从 Windows SSH 到 VM，通过 Nginx 访问 API。
4. 故意制造端口占用、权限错误、配置错误，并用日志定位。
5. 在云服务器配置安全组、域名、TLS 和备份恢复。

这五步比“把所有 Linux 命令背一遍”更能证明工程能力。
