# 06 SSH 密钥与远程运维

SSH 不只是“远程打开一个终端”。它同时提供：

- 服务器身份验证。
- 用户身份验证。
- 加密与完整性保护。
- 远程命令执行。
- 文件传输。
- 端口转发。

本章以“Windows 宿主机管理 Ubuntu 24.04 上的 Go 短链服务”为主线。最终目标是：

- 能验证自己连到的是正确服务器。
- 使用带口令的 Ed25519 用户密钥登录。
- 确认密钥登录可用后再关闭密码登录。
- 不开放 MySQL/Redis 公网端口，通过 SSH 隧道临时维护。
- 能用日志和详细连接信息排查失败。

---

## 1. SSH 连接中有三件不同的密码学工作

初学时最容易把“host key、用户 key、密钥交换”混成一件事。

### 1.1 服务器 host key：证明服务器是谁

服务器安装 OpenSSH 时会生成 host key，例如：

~~~text
/etc/ssh/ssh_host_ed25519_key
/etc/ssh/ssh_host_ed25519_key.pub
~~~

客户端第一次连接时看到的指纹来自服务器 host public key。客户端应通过云控制台、虚拟机控制台、管理员或其他可信渠道核对。

如果不核对，只点击 yes，你只是记住了“当前回答者”，并没有真正证明它是目标服务器。

### 1.2 密钥交换：协商本次连接的会话密钥

SSH 使用密钥交换算法为每次连接协商临时会话密钥。后续数据使用高效的对称加密和完整性保护。

服务器会用 host key 对交换过程签名，防止中间人冒充。

密钥交换算法和用户登录密钥不是同一概念。看到 no matching key exchange method，不能通过重建用户 authorized_keys 来解决。

### 1.3 用户 key：证明客户端用户是谁

用户在客户端生成：

~~~text
private key  → 只留在客户端
public key   → 放到服务器 authorized_keys
~~~

登录时，客户端用 private key 完成签名证明，private key 本身不会被发送到服务器。

服务器泄露 authorized_keys 不等于客户端 private key 泄露；但攻击者会知道允许哪些公钥，仍应保护服务器文件。

### 1.4 private key 的口令

ssh-keygen 询问的 passphrase 用于加密客户端磁盘上的 private key。它不是服务器账户密码，也不会发给服务器。

即使 private key 文件被复制，攻击者还要破解口令。个人长期管理密钥应设置足够强的口令，并配合 ssh-agent 减少重复输入。

---

## 2. 安装和验证 OpenSSH Server

在 Ubuntu 24.04：

~~~bash
sudo apt update
sudo apt install -y openssh-server

systemctl status ssh --no-pager
sudo ss -lntp 'sport = :22'
~~~

Ubuntu 的 systemd 服务通常叫 ssh.service，不是 sshd.service。

如果服务未启用：

~~~bash
sudo systemctl enable --now ssh
journalctl -u ssh -n 100 --no-pager
~~~

### 2.1 防火墙顺序

启用 UFW 前先允许 SSH：

~~~bash
sudo ufw allow OpenSSH
sudo ufw enable
sudo ufw status verbose
~~~

远程服务器必须同时确认云安全组允许管理来源访问 TCP 22。最好只允许你的固定出口 IP、VPN 网段或堡垒机，而不是对整个公网开放。

### 2.2 确认虚拟机地址

~~~bash
ip -brief address
ip route
~~~

VMware NAT 常见情况下，Windows 宿主机可以直接访问虚拟机私网 IP，不需要端口转发。若失败，按网络章节检查监听、防火墙和虚拟网卡，而不是先改 SSH 配置。

Windows PowerShell：

~~~powershell
Test-NetConnection 192.168.80.128 -Port 22
ssh ubuntu_user@192.168.80.128
~~~

---

## 3. 先核对服务器 host key

### 3.1 在服务器可信控制台查看指纹

~~~bash
sudo ssh-keygen \
  -lf /etc/ssh/ssh_host_ed25519_key.pub \
  -E sha256
~~~

典型输出：

~~~text
256 SHA256:AbCdEf... root@server (ED25519)
~~~

把这串 SHA256 指纹与客户端首次连接提示逐字核对。

也可列出所有 host key：

~~~bash
for key in /etc/ssh/ssh_host_*_key.pub; do
  sudo ssh-keygen -lf "$key" -E sha256
done
~~~

### 3.2 第一次连接

~~~powershell
ssh ubuntu_user@192.168.80.128
~~~

客户端会提示服务器尚未知。只有指纹与可信控制台一致时才输入 yes。接受后，host key 会写入客户端 known_hosts。

Windows OpenSSH 常用路径：

~~~text
C:\Users\你的用户名\.ssh\known_hosts
~~~

Linux、WSL 和 Git Bash 通常使用各自环境中的 ~/.ssh/known_hosts。它们不一定共享同一文件。

### 3.3 ssh-keyscan 的边界

~~~bash
ssh-keyscan -t ed25519 192.168.80.128
~~~

ssh-keyscan 只能从网络取回“对方声称的公钥”，它不能证明对方身份。如果同一网络上存在中间人，你可能把攻击者的 key 写入 known_hosts。

它适合在你已经通过其他可信渠道获得指纹后批量收集 key，不适合代替核验。

### 3.4 host key changed

可能原因：

- 服务器重装。
- IP 被分配给另一台机器。
- host key 被管理员轮换。
- DNS 指向错误。
- 中间人攻击。

先通过可信渠道核对新指纹。确认是合法变更后再移除旧记录：

~~~powershell
ssh-keygen -F 192.168.80.128
ssh-keygen -R 192.168.80.128
~~~

不要把删除 known_hosts 或 StrictHostKeyChecking=no 当成固定解决方案。

---

## 4. 生成用户登录密钥

### 4.1 Windows PowerShell

先检查是否已有同名文件，避免覆盖：

~~~powershell
Get-ChildItem $env:USERPROFILE\.ssh -Force
~~~

生成项目专用 Ed25519 密钥：

~~~powershell
ssh-keygen -t ed25519 -a 64 -f $env:USERPROFILE\.ssh\id_ed25519_shortlink -C "honor-shortlink-admin"
~~~

- -t ed25519：现代、短小、速度快。
- -a 64：提高 private key 口令派生的计算成本；它不代表 SSH 网络握手轮数。
- -f：使用项目专用文件，避免覆盖默认 key。
- -C：注释，用于识别用途，不参与认证。

设置强口令。生成结果：

~~~text
id_ed25519_shortlink      private key，不能上传或分享
id_ed25519_shortlink.pub  public key，可以安装到服务器
~~~

查看公钥：

~~~powershell
Get-Content $env:USERPROFILE\.ssh\id_ed25519_shortlink.pub
~~~

### 4.2 Linux 或 WSL

~~~bash
ssh-keygen \
  -t ed25519 \
  -a 64 \
  -f ~/.ssh/id_ed25519_shortlink \
  -C 'honor-shortlink-admin'
~~~

注意：Windows 原生 OpenSSH、WSL 和 Git Bash 是不同客户端环境。密钥在哪个文件系统、哪个 ssh-agent 中，应当明确，不要为了方便到处复制 private key。

### 4.3 什么时候考虑硬件密钥

支持 FIDO2 的硬件设备可生成 ed25519-sk 或 ecdsa-sk 密钥，让签名需要硬件参与。它更适合高价值生产管理账户。

是否支持取决于客户端 OpenSSH、硬件和服务器策略。初学阶段先正确使用带口令的 Ed25519，再理解硬件密钥。

---

## 5. 安装 public key

### 5.1 ssh-copy-id

Linux、WSL 或安装了相应工具的环境：

~~~bash
ssh-copy-id \
  -i ~/.ssh/id_ed25519_shortlink.pub \
  ubuntu_user@192.168.80.128
~~~

它会用当前可用登录方式进入服务器，并把 public key 加入 authorized_keys。

### 5.2 Windows PowerShell 方式

~~~powershell
Get-Content $env:USERPROFILE\.ssh\id_ed25519_shortlink.pub | ssh ubuntu_user@192.168.80.128 'umask 077; mkdir -p ~/.ssh; touch ~/.ssh/authorized_keys; read -r key; grep -qxF "$key" ~/.ssh/authorized_keys || printf "%s\n" "$key" >> ~/.ssh/authorized_keys'
~~~

这个远程命令：

- 用 umask 077 限制新文件权限。
- 创建 ~/.ssh 和 authorized_keys。
- 仅在完全相同的 key 不存在时追加，避免重复。

public key 是一行文本。不要把 private key 通过管道、scp 或剪贴板放到服务器。

### 5.3 服务器端权限

在服务器：

~~~bash
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys
sudo chown -R "$USER:$(id -gn)" "$HOME/.ssh"
namei -l ~/.ssh/authorized_keys
~~~

如果 home、.ssh 或 authorized_keys 可由其他用户写入，sshd 可能拒绝使用它。

管理员排查另一个用户时：

~~~bash
sudo -u ubuntu_user \
  test -r /home/ubuntu_user/.ssh/authorized_keys

sudo namei -l \
  /home/ubuntu_user/.ssh/authorized_keys
~~~

### 5.4 测试指定 key

Windows PowerShell：

~~~powershell
ssh -i $env:USERPROFILE\.ssh\id_ed25519_shortlink -o IdentitiesOnly=yes ubuntu_user@192.168.80.128
~~~

只有新密钥登录成功后，才考虑关闭密码登录。

---

## 6. 客户端 SSH config

频繁输入 IP、用户和 key 路径容易出错。编辑客户端：

~~~text
~/.ssh/config
~~~

Windows 原生 OpenSSH 对应：

~~~text
C:\Users\你的用户名\.ssh\config
~~~

示例：

~~~sshconfig
Host shortlink-prod
    HostName 192.168.80.128
    User ubuntu_user
    Port 22
    IdentityFile ~/.ssh/id_ed25519_shortlink
    IdentitiesOnly yes
    ServerAliveInterval 30
    ServerAliveCountMax 3
    ForwardAgent no
~~~

之后：

~~~powershell
ssh shortlink-prod
scp .\dist\shortlink shortlink-prod:/tmp/
~~~

查看最终客户端配置：

~~~powershell
ssh -G shortlink-prod
~~~

重点检查 hostname、user、port、identityfile、proxyjump。

### 6.1 IdentitiesOnly 的意义

ssh-agent 中密钥很多时，客户端可能依次尝试，服务器在正确 key 出现前就达到 MaxAuthTries，报 Too many authentication failures。

IdentitiesOnly yes 让该 Host 只尝试明确配置的 key。

### 6.2 不在全局关闭 host key 检查

不要配置：

~~~sshconfig
Host *
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
~~~

这等于主动放弃服务器身份验证。自动化应预先以可信方式分发 host key，而不是关闭检查。

---

## 7. 安全关闭密码登录

这是高风险步骤，顺序比配置内容更重要。

### 7.1 开两个会话

1. 保留当前已登录的 SSH 会话。
2. 新开第二个终端，用指定 key 成功登录。
3. 保留云控制台或虚拟机控制台作为恢复通道。

没有完成这三项，不要关闭密码登录。

### 7.2 Ubuntu drop-in 与“先出现的值”

OpenSSH 的很多配置使用“读取到的第一个值”。Ubuntu 主配置通常较早 Include：

~~~text
/etc/ssh/sshd_config.d/*.conf
~~~

云镜像还可能有 50-cloud-init.conf。为了让本地策略更早生效，可创建：

~~~bash
sudoedit /etc/ssh/sshd_config.d/00-shortlink-hardening.conf
~~~

内容示例：

~~~text
PermitRootLogin no
PubkeyAuthentication yes
PasswordAuthentication no
KbdInteractiveAuthentication no
PermitEmptyPasswords no
UsePAM yes

LoginGraceTime 30
MaxAuthTries 4

X11Forwarding no
AllowAgentForwarding no
AllowTcpForwarding local
GatewayPorts no

AllowUsers ubuntu_user deploy
~~~

必须把 ubuntu_user 替换成你实际已经验证密钥登录的管理员用户名。不要原样照抄，否则 AllowUsers 可能把真实用户挡在外面。

解释：

- PermitRootLogin no：禁止 root 直接 SSH 登录，管理员通过 sudo 提权。
- PasswordAuthentication no：关闭 SSH 密码认证。
- KbdInteractiveAuthentication no：关闭常被 PAM 用于口令交互的方式。
- UsePAM yes：仍保留账户、会话等 PAM 处理，不等于重新开启密码。
- AllowAgentForwarding no：服务器不能借用客户端 agent。
- AllowTcpForwarding local：允许本地转发，便于临时维护 MySQL/Redis；不允许 remote forwarding。
- GatewayPorts no：转发监听默认不扩展到外部地址。
- AllowUsers：白名单，最容易造成锁定，必须谨慎。

端口改成非 22 可以减少扫描日志噪声，但不是核心安全控制。密钥认证、来源限制、及时更新和日志监控更重要。

### 7.3 语法和有效配置

~~~bash
sudo sshd -t
~~~

无输出通常表示语法通过。

查看全局有效值：

~~~bash
sudo sshd -T |
  grep -E \
  '^(permitrootlogin|pubkeyauthentication|passwordauthentication|kbdinteractiveauthentication|usepam|allowagentforwarding|allowtcpforwarding|gatewayports|maxauthtries|logingracetime) '
~~~

针对某个连接条件：

~~~bash
sudo sshd -T \
  -C user=ubuntu_user,host=server.example,addr=192.168.80.1
~~~

这是发现 drop-in 顺序、Match 规则和云配置覆盖问题的关键命令。不要只看某一行文件就认定它生效。

### 7.4 reload 并用新会话测试

~~~bash
sudo systemctl reload ssh
systemctl status ssh --no-pager
~~~

保持旧会话不退出，在新终端：

~~~powershell
ssh shortlink-prod
~~~

然后明确测试密码方式不会回退：

~~~powershell
ssh -o PubkeyAuthentication=no -o PreferredAuthentications=password,keyboard-interactive ubuntu_user@192.168.80.128
~~~

预期认证失败。

### 7.5 如果新会话失败

在仍保留的旧会话：

~~~bash
sudo mv \
  /etc/ssh/sshd_config.d/00-shortlink-hardening.conf \
  /etc/ssh/sshd_config.d/00-shortlink-hardening.conf.disabled

sudo sshd -t
sudo systemctl reload ssh
journalctl -u ssh -n 100 --no-pager
~~~

恢复后再逐项分析，不要关闭最后一个可用会话。

---

## 8. 管理用户与运行用户分离

短链服务账户：

~~~text
shortlink
~~~

它只负责运行 Go 进程，通常不需要 SSH 登录。

管理员或部署账户：

~~~text
ubuntu_user
deploy
~~~

它们可以通过受控 sudo 执行特定管理操作。

禁用服务账户交互登录：

~~~bash
getent passwd shortlink
sudo usermod --shell /usr/sbin/nologin shortlink
~~~

不要把运行服务的账户顺手加入 sudo 组。应用被攻破时，账户权限越小，影响越可控。

创建部署账户：

~~~bash
sudo adduser deploy
sudo install -d \
  -o deploy \
  -g deploy \
  -m 0700 \
  /home/deploy/.ssh
~~~

安装 deploy public key 后，再根据发布脚本设计最小 sudoers。不要直接：

~~~text
deploy ALL=(ALL) NOPASSWD: ALL
~~~

脚本、父目录和 sudoers 都必须由 root 控制，deploy 不可写。

---

## 9. 文件传输

### 9.1 scp

现代 OpenSSH 的 scp 默认通常使用 SFTP 协议，但命令界面仍是 scp：

~~~powershell
ssh shortlink-prod 'install -d -m 0700 /tmp/shortlink-upload'
scp .\dist\shortlink .\dist\shortlink.sha256 shortlink-prod:/tmp/shortlink-upload/
~~~

上传后在服务器验证：

~~~bash
cd /tmp/shortlink-upload
sha256sum -c shortlink.sha256
~~~

不要直接覆盖 /opt/shortlink/current/shortlink。应上传到暂存目录，校验后交给发布脚本创建新 release。

### 9.2 sftp

~~~powershell
sftp shortlink-prod
~~~

常用交互命令：

~~~text
pwd
lpwd
ls
lls
put local-file
get remote-file
mkdir upload
exit
~~~

sftp 更适合受限文件传输账户和交互传文件。

### 9.3 rsync

Linux/WSL：

~~~bash
rsync -av --progress \
  ./dist/ \
  shortlink-prod:/tmp/shortlink-upload/
~~~

尾部斜杠语义很重要：

~~~text
source/  → 复制 source 目录里的内容
source   → 复制 source 目录本身
~~~

不要对生产目标随手使用 --delete。它会删除目标中源端不存在的文件，必须先 dry-run：

~~~bash
rsync -avn --delete source/ host:target/
~~~

即使 dry-run 正确，也不应让 rsync 直接同步不可变 release 的 active 目录。

### 9.4 远程命令的引号边界

~~~bash
ssh shortlink-prod 'date -u; systemctl is-active shortlink'
~~~

单引号让本地 shell 不展开内容，由远程 shell解释。

如果把不可信输入拼到远程命令字符串中，可能造成命令注入。自动化优先：

- 使用固定远程脚本。
- 只传受校验的简单参数。
- 对参数做白名单。
- 避免 eval 和多层字符串拼接。

---

## 10. SSH 本地隧道：不公开数据库也能维护

### 10.1 MySQL 本地转发

服务器 MySQL 只监听 127.0.0.1:3306。从 Windows 建立：

~~~powershell
ssh -N -T -o ExitOnForwardFailure=yes -L 127.0.0.1:13306:127.0.0.1:3306 shortlink-prod
~~~

含义：

~~~text
Windows 127.0.0.1:13306
          ↓ 加密 SSH 连接
服务器从自己的 127.0.0.1:3306 访问 MySQL
~~~

此时本地数据库客户端连接：

~~~text
host=127.0.0.1
port=13306
~~~

MySQL 3306 仍没有向公网开放。

- -N：不执行远程命令。
- -T：不分配伪终端。
- ExitOnForwardFailure=yes：端口绑定失败时直接退出。
- 第一个 127.0.0.1 限制隧道只供 Windows 本机使用。

### 10.2 Redis 本地转发

另开本地端口：

~~~powershell
ssh -N -T -o ExitOnForwardFailure=yes -L 127.0.0.1:16379:127.0.0.1:6379 shortlink-prod
~~~

本地 Redis 客户端连接 127.0.0.1:16379，仍需 Redis ACL 用户名和密码。SSH 隧道不替代数据库层认证。

### 10.3 同时转发

~~~powershell
ssh -N -T -o ExitOnForwardFailure=yes -L 127.0.0.1:13306:127.0.0.1:3306 -L 127.0.0.1:16379:127.0.0.1:6379 shortlink-prod
~~~

保持该终端运行。结束时 Ctrl+C。

### 10.4 隧道边界

- SSH 用户必须被允许 local forwarding。
- 服务端 AllowTcpForwarding local 支持本地转发。
- 更高安全要求可通过 Match User 和 PermitOpen 限制可转发的目标。
- 不要把本地监听写成 0.0.0.0，除非明确要让其他机器使用并已设计访问控制。
- 隧道断开时客户端连接会中断，不适合作为未监控的长期生产网络架构。

Remote forwarding 会让服务器一侧开放入口，风险模型不同。本项目日常维护不需要，不要因为命令相似而混用 -R。

---

## 11. ssh-agent 与代理转发

### 11.1 agent 的作用

ssh-agent 把已解锁 private key 保存在当前登录会话的受保护进程中，客户端请求 agent 完成签名。

Windows 可查看：

~~~powershell
Get-Service ssh-agent
~~~

是否允许启用服务取决于本机策略。启用后添加：

~~~powershell
ssh-add $env:USERPROFILE\.ssh\id_ed25519_shortlink
ssh-add -l
~~~

Linux：

~~~bash
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519_shortlink
ssh-add -l
~~~

不要把 eval 用法泛化到不可信文本；这里执行的是本机 ssh-agent 输出的固定环境设置。

### 11.2 为什么默认关闭 agent forwarding

ForwardAgent yes 会让远程服务器通过转发 socket 使用你的本地 agent。private key 不会被复制过去，但如果远程服务器被攻破，攻击者可能在会话期间借用 agent 对其他服务器认证。

优先使用：

- 从本机直接连接目标。
- ProxyJump 经过堡垒机。
- 为每个环境准备专用 key。
- CI 使用受控部署凭据。

而不是把 agent 到处转发。

### 11.3 ProxyJump

~~~sshconfig
Host bastion
    HostName bastion.example.com
    User admin
    IdentityFile ~/.ssh/id_ed25519_bastion
    IdentitiesOnly yes

Host shortlink-private
    HostName 10.0.1.20
    User deploy
    IdentityFile ~/.ssh/id_ed25519_shortlink
    IdentitiesOnly yes
    ProxyJump bastion
    ForwardAgent no
~~~

客户端通过堡垒机转发字节流，但仍直接验证目标服务器 host key，并使用目标专用用户 key。

---

## 12. 自动化和 key 生命周期

### 12.1 一把 key 不要通吃所有环境

建议至少区分：

- 个人开发虚拟机。
- 测试环境。
- 生产管理。
- CI 部署。

这样某一把 key 泄露时，撤销范围可控。

### 12.2 authorized_keys 可以附加限制

一行 public key 前可以加选项：

~~~text
from="203.0.113.10/32",no-agent-forwarding,no-port-forwarding,no-X11-forwarding,no-pty ssh-ed25519 AAAA... ci-deploy
~~~

它限制来源并关闭不需要的能力。

边界：

- CI 出口 IP 可能变化。
- no-port-forwarding 会阻止隧道。
- no-pty 不会自动限制所有远程命令。
- 仅靠注释没有安全效果。
- forced command 需要安全处理 SSH_ORIGINAL_COMMAND，否则仍可能注入。

高要求部署应使用专门上传目录、固定发布入口、最小 sudoers 和制品签名，而不是只在 authorized_keys 上堆选项。

### 12.3 撤销 key

从对应用户 authorized_keys 中删除准确的一行，然后重新测试：

~~~bash
cp -a ~/.ssh/authorized_keys ~/.ssh/authorized_keys.backup
editor ~/.ssh/authorized_keys
~~~

不要用模糊 grep -v 直接覆盖，可能误删其他人的 key。生产应记录：

- key 指纹。
- 所有人。
- 用途。
- 创建时间。
- 到期/复核时间。
- 撤销原因。

查看用户 public key 指纹：

~~~bash
ssh-keygen -lf ~/.ssh/authorized_keys -E sha256
~~~

### 12.4 private key 泄露

立即：

1. 从所有服务器撤销对应 public key。
2. 生成新 key，不复用旧 private key。
3. 检查认证日志和来源。
4. 轮换可能通过该身份访问的其他凭据。
5. 查明泄露路径并修复。

给 private key 改口令不能撤销攻击者已经复制的旧文件。

---

## 13. 排障

### 13.1 客户端详细日志

~~~powershell
ssh -vvv shortlink-prod
~~~

关注：

- 实际连接的 hostname 和 port。
- known_hosts 匹配哪条记录。
- 提供了哪些 public key。
- 服务器接受了哪种认证方法。
- 失败发生在网络、密钥交换、host key 还是用户认证。

-vvv 会暴露用户名、主机、路径和部分环境信息。分享日志前脱敏，但它不会正常打印 private key 内容。

### 13.2 服务端日志

~~~bash
sudo journalctl -u ssh -n 100 --no-pager
sudo journalctl -u ssh --since '10 minutes ago' --no-pager
~~~

同时验证：

~~~bash
sudo sshd -t
sudo sshd -T
sudo ss -lntp 'sport = :22'
~~~

### 13.3 Connection refused

- IP 可达，但 22 没监听。
- ssh 服务未启动。
- SSH 实际改到另一个端口。
- 防火墙主动拒绝。

在服务器控制台：

~~~bash
systemctl status ssh --no-pager
sudo ss -lntp
~~~

### 13.4 Connection timed out

- IP、路由或 NAT 错误。
- UFW、云安全组或上游防火墙静默丢包。
- 目标机器关机。

先用 Test-NetConnection 和服务器控制台确认，不要重建 key。

### 13.5 Permission denied (publickey)

检查：

~~~bash
getent passwd ubuntu_user
sudo namei -l /home/ubuntu_user/.ssh/authorized_keys
sudo -u ubuntu_user \
  ssh-keygen -lf /home/ubuntu_user/.ssh/authorized_keys
sudo sshd -T \
  -C user=ubuntu_user,host=server,addr=192.168.80.1
~~~

常见原因：

- 客户端用了另一把 key。
- public key 装到了另一个用户。
- 文件权限或所有者错误。
- AllowUsers 不包含该用户。
- authorized_keys 行被折断。
- 客户端 agent 尝试太多 key。
- 用户账户被锁定或 shell 不允许登录。

客户端显式测试：

~~~powershell
ssh -vvv -o IdentitiesOnly=yes -i $env:USERPROFILE\.ssh\id_ed25519_shortlink ubuntu_user@192.168.80.128
~~~

### 13.6 Too many authentication failures

这不一定是密码错。agent 可能在正确 key 之前提交太多 key。

~~~powershell
ssh -o IdentitiesOnly=yes -i $env:USERPROFILE\.ssh\id_ed25519_shortlink ubuntu_user@192.168.80.128
~~~

然后把相同设置写入该 Host 的 config。

### 13.7 no matching host key type 或 key exchange method

查看客户端支持列表：

~~~powershell
ssh -Q HostKeyAlgorithms
ssh -Q kex
ssh -Q cipher
~~~

查看协商过程：

~~~powershell
ssh -vvv legacy-host
~~~

不要全局重新启用这些过时算法：

~~~text
ssh-dss
ssh-rsa 旧 SHA-1 签名方式
diffie-hellman-group1-sha1
~~~

注意：RSA key 本身与 ssh-rsa 这一个 SHA-1 签名算法不是完全同义。现代 OpenSSH 可以用 RSA key 配合 rsa-sha2-256/512。看到报错应读清算法类别。

最佳方案是升级旧服务器。若业务被迫临时兼容，应只对单个 Host 加最小例外、记录风险和删除期限，不能写进 Host *。

### 13.8 登录成功但一会儿断开

区分：

- 网络/NAT 空闲超时。
- 服务端 ClientAlive 策略。
- 客户端睡眠或网络切换。
- 命令自身退出。

客户端 config：

~~~sshconfig
ServerAliveInterval 30
ServerAliveCountMax 3
~~~

这是客户端通过加密通道做存活探测，不等于让坏网络永不掉线。

---

## 14. 实验：完成一次安全远程运维闭环

### 阶段一：验证身份

1. 在 Ubuntu 控制台查看 Ed25519 host key 指纹。
2. Windows 第一次连接时核对指纹。
3. 用 ssh-keygen -F 确认 known_hosts 已记录。

预期：你能解释“为什么这个指纹证明的是服务器，不是登录用户”。

### 阶段二：切换用户密钥

1. 生成 id_ed25519_shortlink。
2. 安装 public key。
3. 用 -i 和 IdentitiesOnly=yes 新开连接。
4. 保留旧连接。

预期：新连接只询问 private key 口令，不询问服务器用户密码。

### 阶段三：加固 sshd

1. 创建 00-shortlink-hardening.conf。
2. sshd -t。
3. sshd -T 检查有效值。
4. reload ssh。
5. 新窗口验证 key 登录。
6. 显式验证密码登录失败。

预期：整个过程中始终有一个恢复会话。

### 阶段四：传输 Go 产物

~~~powershell
ssh shortlink-prod 'install -d -m 0700 /tmp/shortlink-upload'
scp .\dist\shortlink .\dist\shortlink.sha256 shortlink-prod:/tmp/shortlink-upload/
~~~

服务器：

~~~bash
cd /tmp/shortlink-upload
sha256sum -c shortlink.sha256
file shortlink
~~~

预期：校验 OK，file 显示与服务器架构匹配的 Linux 可执行文件。

### 阶段五：建立 MySQL 隧道

~~~powershell
ssh -N -T -o ExitOnForwardFailure=yes -L 127.0.0.1:13306:127.0.0.1:3306 shortlink-prod
~~~

另开 PowerShell：

~~~powershell
Test-NetConnection 127.0.0.1 -Port 13306
~~~

服务器验证 MySQL 仍只监听回环：

~~~bash
sudo ss -lntp 'sport = :3306'
sudo ufw status verbose
~~~

预期：本地隧道可用，公网和虚拟机外部仍不能直连 3306。

---

## 15. 生产边界

### 15.1 SSH 不是公开服务的万能保护壳

SSH 隧道适合：

- 临时数据库管理。
- 运维排障。
- 跨堡垒机访问私网。

不适合替代：

- 正式服务发现。
- 长期高可用网络。
- 数据库连接池的稳定网络设计。
- VPN 或私网架构。

### 15.2 暴力扫描与 Fail2ban

如果安全组已把 22 限制到可信来源，收益通常高于单纯安装 Fail2ban。

公网开放 SSH 时，Fail2ban 可以根据日志临时封禁重复失败来源，但它：

- 不是密钥认证的替代。
- 可能误封共享出口。
- 需要理解日志、后端防火墙和恢复方式。

先完成来源限制、关闭密码、禁止 root、补丁更新，再考虑它。

### 15.3 更新与恢复

OpenSSH 更新前：

- 保留控制台。
- 检查配置语法。
- 确认服务更新后监听。
- 不要同时改端口、算法、认证和防火墙。

备份配置不能代替恢复通道。真正被锁在外面时，云控制台、虚拟机控制台或带外管理才是恢复手段。

---

## 16. 本章验收

知识验收：

1. host key、用户 key 和密钥交换分别做什么？
2. 第一次连接为什么不能只靠 ssh-keyscan 建立信任？
3. 用户 private key 是否会发送到服务器？
4. 为什么必须先验证 key 登录，再关闭密码？
5. Ubuntu drop-in 文件顺序为什么会影响配置？
6. sshd -t 和 sshd -T 的区别是什么？
7. 为什么 ProxyJump 通常优于 agent forwarding？
8. SSH 本地隧道为什么不需要公开 MySQL 3306？
9. no matching key exchange method 为什么不是 authorized_keys 问题？

动手验收：

- 能从可信控制台核对服务器 host key 指纹。
- 能生成带口令的项目专用 Ed25519 key。
- 能解释并修复 authorized_keys 权限。
- 能通过客户端 config 使用短别名登录。
- 能在不关闭旧会话的情况下安全禁用密码登录。
- 能用 ssh -vvv 和 journalctl -u ssh 从两端排障。
- 能上传 Go 二进制并验证 SHA-256。
- 能建立仅监听本机的 MySQL 或 Redis SSH 隧道。

做到这些，你就不是“会输 ssh 命令”，而是能够安全地建立、验证、使用和恢复一条远程运维通道。
