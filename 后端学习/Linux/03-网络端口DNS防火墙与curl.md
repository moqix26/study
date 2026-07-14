# 03 网络、端口、DNS、防火墙与 curl

这一章不追求背完网络协议，而是建立一条后端开发最常用的排障链：

> 域名解析到哪个 IP → 数据包走哪张网卡和路由 → 目标端口是否有人监听 → 监听地址是否允许当前来源访问 → 防火墙是否放行 → HTTP 服务返回了什么。

学完后，你应该能独立判断“访问不了 Gin 接口”到底是代码、监听地址、DNS、虚拟机网络、防火墙，还是上游代理的问题。

---

## 1. 先建立一张请求路径图

浏览器访问 https://api.example.com/healthz 时，大致会经过：

1. DNS 把 api.example.com 解析成 IP。
2. 操作系统根据路由表选择出口网卡和下一跳。
3. 客户端选择一个临时源端口，连接服务器的 443 端口。
4. 云安全组、主机防火墙和中间网络设备决定是否放行。
5. Nginx 在 443 端口接收连接并完成 TLS。
6. Nginx 把请求转发到 127.0.0.1:8080。
7. Go/Gin 进程在 127.0.0.1:8080 接收请求并返回结果。

这条链上任何一环出错，用户看到的都可能只是“连接失败”。

后端排障不要从猜测开始，按层验证：

~~~text
进程是否活着
  ↓
端口是否监听、监听在哪个地址
  ↓
服务器本机 curl 是否成功
  ↓
同一局域网或宿主机是否成功
  ↓
防火墙、安全组、NAT 是否放行
  ↓
DNS、Nginx、TLS 是否正确
~~~

---

## 2. IP、网卡与路由

### 2.1 查看网卡和地址

Ubuntu 24.04 推荐使用 ip 命令：

~~~bash
ip -brief address
ip address show
ip link show
~~~

典型输出：

~~~text
lo               UNKNOWN        127.0.0.1/8 ::1/128
ens33            UP             192.168.80.128/24 fe80::20c:29ff:...
~~~

需要分清：

- lo 是回环网卡，只在本机内部使用。
- 127.0.0.1 是 IPv4 回环地址。
- ::1 是 IPv6 回环地址。
- 192.168.x.x、10.x.x.x、172.16.x.x 到 172.31.x.x 通常是私网地址。
- UP 说明链路已启用，不代表 DNS、路由和互联网一定正常。

查看某个地址属于哪张网卡：

~~~bash
ip address show dev ens33
~~~

### 2.2 查看路由

~~~bash
ip route
ip -6 route
~~~

典型 IPv4 输出：

~~~text
default via 192.168.80.2 dev ens33 proto dhcp
192.168.80.0/24 dev ens33 proto kernel scope link src 192.168.80.128
~~~

含义：

- 访问 192.168.80.0/24 时直接从 ens33 发出。
- 其他未匹配的 IPv4 流量交给默认网关 192.168.80.2。

想知道访问某个目标具体走哪条路由：

~~~bash
ip route get 1.1.1.1
ip route get 192.168.80.1
~~~

这比只看默认路由更直接。多网卡、VPN、Docker 和云服务器上经常存在更具体的路由。

### 2.3 ping 能说明什么

~~~bash
ping -c 4 192.168.80.2
ping -c 4 1.1.1.1
ping -c 4 example.com
~~~

三次测试分别偏向验证：

- 到网关的二层/三层连通性。
- 不依赖 DNS 的互联网连通性。
- DNS 加互联网连通性。

但 ping 失败不能直接推出“服务器宕机”。很多服务器或安全策略会丢弃 ICMP，却仍允许 TCP 80/443。验证 Web 服务应继续使用 curl 或 nc。

### 2.4 用 tracepath 观察路径

~~~bash
sudo apt update
sudo apt install -y iputils-tracepath
tracepath example.com
~~~

中间某一跳显示无响应不一定是故障，路由器可能只是不回复探测包。重点看最终目标是否可达以及路径是否在预期网络中。

---

## 3. VMware 网络模式：NAT 不是“宿主机必须端口转发”

这是很容易学错的一点。

### 3.1 NAT 模式

VMware 通常会在宿主机创建一张虚拟网卡，并为虚拟机分配一个私网 IP。常见情况下：

- 虚拟机可以通过 NAT 访问互联网。
- 宿主机可以直接访问虚拟机的私网 IP。
- 同一 NAT 私网中的虚拟机通常也能互相访问。
- 局域网中的其他真实机器默认不一定能直接访问虚拟机。

因此，从 Windows 宿主机访问 NAT 虚拟机里的 Gin：

~~~powershell
curl.exe http://192.168.80.128:8080/healthz
~~~

通常不需要 VMware 端口转发。前提是：

- Gin 监听虚拟机可达地址，而不只是 127.0.0.1。
- Ubuntu 防火墙允许 8080。
- VMware 虚拟网络没有被手动隔离。

端口转发主要用于把“宿主机的某个端口”映射到虚拟机，或让 NAT 网络之外的来源经宿主机进入虚拟机。不同 VMware 产品和网络配置可能不同，应以实际网段、路由和连通测试为准。

### 3.2 桥接模式

虚拟机像局域网中的一台独立设备，通常从真实路由器获取 IP。优点是其他局域网设备容易访问；缺点是公共 Wi-Fi、校园网、企业网可能限制额外设备或二层通信。

### 3.3 Host-only 模式

虚拟机通常只能与宿主机及同一 Host-only 网络中的虚拟机通信，默认没有互联网。适合封闭实验。

### 3.4 遇到虚拟机访问失败时

在 Ubuntu 中：

~~~bash
ip -brief address
ip route
ss -lntp
sudo ufw status verbose
~~~

在 Windows PowerShell 中：

~~~powershell
Test-NetConnection 192.168.80.128 -Port 8080
curl.exe -v http://192.168.80.128:8080/healthz
~~~

先验证宿主机到虚拟机 IP 的连接，不要一开始就配置端口转发。

---

## 4. 端口、监听地址与连接状态

### 4.1 端口属于传输层，不属于某个程序语言

一台机器可同时运行多个网络服务，因为它们监听不同的“协议 + 地址 + 端口”组合：

~~~text
tcp 127.0.0.1:8080  → Gin
tcp 0.0.0.0:80      → Nginx
tcp 0.0.0.0:443     → Nginx
tcp 127.0.0.1:3306  → MySQL
tcp 127.0.0.1:6379  → Redis
~~~

TCP 和 UDP 的端口空间彼此独立。一个进程占用 TCP 8080，并不等于 UDP 8080 也被占用。

### 4.2 查看监听端口

~~~bash
sudo ss -lntp
sudo ss -lnup
sudo ss -lntp 'sport = :8080'
~~~

参数含义：

- l：只看监听 socket。
- n：不把 IP 和端口反查成名称。
- t：TCP。
- u：UDP。
- p：显示进程信息，通常需要 sudo 才完整。

典型输出：

~~~text
LISTEN 0 4096 127.0.0.1:8080 0.0.0.0:* users:(("shortlink",pid=2314,fd=7))
~~~

这说明进程存在、TCP 8080 已监听，并且只接受本机 IPv4 访问。

### 4.3 127.0.0.1、0.0.0.0 和具体私网 IP

假设服务器地址是 192.168.80.128：

| 监听地址 | 本机访问 | 宿主机/局域网访问 | 典型用途 |
|---|---:|---:|---|
| 127.0.0.1:8080 | 可以 | 不可以 | Nginx 与 Go 同机反代 |
| 192.168.80.128:8080 | 可以 | 可以 | 只在指定网卡提供服务 |
| 0.0.0.0:8080 | 可以 | 视防火墙而定 | 接受所有 IPv4 网卡的入站连接 |
| [::1]:8080 | 可以 | 不可以 | IPv6 本机回环 |
| [::]:8080 | 视系统设置而定 | 视防火墙而定 | 所有 IPv6 地址，部分系统也可能接 IPv4 |

0.0.0.0 是“监听所有本机 IPv4 地址”，不是一个应该拿来 curl 的真实目标地址。

生产环境如果 Nginx 和 Go 在同一台机器，优先让 Go 监听 127.0.0.1:8080，只公开 Nginx 的 80/443。这样即使防火墙误配，业务进程也不会直接暴露。

开发时需要从 Windows 宿主机直连虚拟机中的 Gin，可以临时监听 0.0.0.0:8080，并用 UFW 限制来源。不要把这一开发配置机械照搬到生产。

### 4.4 Gin 的监听地址

Gin 的 Run 参数最终仍是 TCP 地址：

~~~go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 同机 Nginx 反代时推荐：
	if err := r.Run("127.0.0.1:8080"); err != nil {
		panic(err)
	}
}
~~~

如果实验阶段要让宿主机访问，可暂时改为：

~~~go
if err := r.Run("0.0.0.0:8080"); err != nil {
	panic(err)
}
~~~

更成熟的项目会从配置读取地址，并为 HTTP Server 设置超时，而不是把地址写死。

### 4.5 已建立连接与 TIME_WAIT

~~~bash
ss -nt
ss -ant state established
ss -ant state time-wait
ss -s
~~~

TIME_WAIT 是 TCP 正常关闭过程的一部分，通常出现在主动关闭连接的一方。它由内核维护，用来避免旧报文污染后续同四元组连接，并确保必要时能重传最后的 ACK。

需要纠正两个误区：

- TIME_WAIT 不等于“某个应用进程仍占着一个文件描述符”。正常 close 后，应用的文件描述符已经释放，内核仍保留一段连接状态。
- 看到一些 TIME_WAIT 不代表故障。高并发短连接下数量较多很常见。

真正需要关注的是：

- 临时端口范围是否耗尽。
- 客户端是否没有复用 HTTP 连接。
- 是否出现异常短连接风暴。
- NAT 或负载均衡设备的连接跟踪表是否耗尽。

查看临时端口范围：

~~~bash
sysctl net.ipv4.ip_local_port_range
~~~

不要因为看到 TIME_WAIT 就随意复制网络参数优化命令。先确认请求模型、连接复用和实际资源瓶颈。

---

## 5. DNS：名称如何变成地址

### 5.1 应用实际使用的解析结果

~~~bash
getent ahosts example.com
getent hosts example.com
~~~

getent 走系统名称服务配置，通常比只用某个 DNS 工具更接近应用看到的结果，因为它还会考虑 /etc/hosts 和 /etc/nsswitch.conf。

### 5.2 systemd-resolved

Ubuntu 24.04 桌面版和常见服务器环境通常使用 systemd-resolved：

~~~bash
resolvectl status
resolvectl query example.com
ls -l /etc/resolv.conf
~~~

/etc/resolv.conf 常常是指向 systemd-resolved 生成文件的符号链接。不要看到 127.0.0.53 就误以为 DNS 服务器在公网；它是本机 stub resolver，再把查询转发给网卡对应的 DNS。

清理本机解析缓存：

~~~bash
sudo resolvectl flush-caches
~~~

### 5.3 dig 做定向查询

~~~bash
sudo apt update
sudo apt install -y dnsutils
dig example.com
dig +short example.com A
dig +short example.com AAAA
dig @1.1.1.1 example.com
~~~

定向询问 1.1.1.1 成功，而系统默认解析失败，说明问题更可能在本机或当前网络下发的 DNS 配置。

注意：某些网络会拦截外部 DNS，企业环境也可能要求使用内部 DNS。不要把公共 DNS 当成所有环境的固定答案。

### 5.4 /etc/hosts 的作用和边界

~~~bash
cat /etc/hosts
getent hosts short.test
~~~

实验时可添加：

~~~text
192.168.80.128 short.test
~~~

/etc/hosts 只影响本机，不会自动同步给宿主机、手机或其他服务器。Windows 宿主机有自己的 hosts 文件。

### 5.5 curl 绕过 DNS 做定位

假设域名应指向 203.0.113.10：

~~~bash
curl --resolve api.example.com:443:203.0.113.10 \
  https://api.example.com/healthz
~~~

--resolve 会同时保留正确的 Host 和 TLS SNI，只临时替换本次请求的解析结果。它比直接访问 https://203.0.113.10 更适合排查 HTTPS 虚拟主机。

---

## 6. curl：后端开发的探针

### 6.1 最常用组合

~~~bash
curl -i http://127.0.0.1:8080/healthz
curl -v http://127.0.0.1:8080/healthz
curl -sS -f http://127.0.0.1:8080/healthz
~~~

- -i：把响应头一起打印。
- -v：显示解析、连接、请求和响应过程，调试时使用。
- -sS：安静输出，但错误仍显示。
- -f：HTTP 400 及以上返回非零退出码，适合脚本健康检查。

脚本中应设置超时，防止无限等待：

~~~bash
curl -sS -f \
  --connect-timeout 2 \
  --max-time 5 \
  http://127.0.0.1:8080/healthz
~~~

### 6.2 调用 JSON API

创建短链的示例：

~~~bash
curl -i \
  -X POST \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://go.dev/doc/"}' \
  http://127.0.0.1:8080/api/v1/short-links
~~~

如果接口使用 Bearer Token：

~~~bash
TOKEN='replace-with-a-temporary-token'

curl -sS -f \
  -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:8080/api/v1/me
~~~

不要把真实长期密钥直接写进文档、Git 仓库或共享终端截图。命令行参数和环境变量在某些系统上也可能被同机用户观察，生产自动化应使用受控的 secret 文件或密钥服务。

### 6.3 只输出状态码和耗时

~~~bash
curl -sS -o /dev/null \
  -w 'status=%{http_code} connect=%{time_connect}s total=%{time_total}s\n' \
  http://127.0.0.1:8080/healthz
~~~

常用指标：

- time_namelookup：DNS 用时。
- time_connect：TCP 建连用时。
- time_appconnect：TLS 握手完成用时。
- time_starttransfer：收到首字节前的总用时。
- time_total：完整请求用时。

### 6.4 常见错误如何解读

Connection refused：

- IP 可达，但该地址端口没有监听。
- 进程已退出。
- 服务只监听了另一个地址族或地址。
- 防火墙明确返回拒绝。

Connection timed out：

- 路由不通。
- 防火墙或安全组静默丢包。
- 目标地址错误。
- 中间网络设备没有响应。

Could not resolve host：

- DNS 或命令引号有问题。
- 域名拼错。

HTTP 502：

- 你通常已经访问到反向代理。
- 代理连不上 Go 上游，或上游返回无效响应。

HTTP 404：

- 网络和 HTTP 服务大概率已经通了。
- 路由、Host、路径或方法不匹配。

HTTP 401/403：

- 服务已经响应。
- 继续检查身份认证和授权，不要再反复改防火墙。

---

## 7. UFW 与云安全组

### 7.1 三层边界要分开

常见生产环境同时存在：

1. 云安全组或云防火墙。
2. Ubuntu 主机防火墙。
3. 应用自己的监听地址和认证授权。

其中一层允许，不代表最终一定可达；其中一层拒绝，就可能不可达。

### 7.2 安全启用 UFW

远程服务器上，先允许 SSH，再启用防火墙：

~~~bash
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow OpenSSH
sudo ufw enable
sudo ufw status verbose
~~~

如果 SSH 使用非默认端口，应先精确放行实际端口。启用前保留云控制台或其他恢复通道，避免把自己锁在服务器外。

公开 Web 服务：

~~~bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
~~~

如果只是从固定管理 IP 做临时实验：

~~~bash
sudo ufw allow from 192.168.80.1 to any port 8080 proto tcp
~~~

删除规则前先查看编号：

~~~bash
sudo ufw status numbered
sudo ufw delete 3
~~~

编号会在删除后变化，一次删除一条并重新查看。

### 7.3 不应公开的端口

短链项目同机部署时，通常只公开：

- 22：SSH，最好再由云安全组限制管理来源。
- 80/443：Nginx。

通常不公开：

- 8080：Gin，仅监听 127.0.0.1。
- 3306：MySQL，仅监听 127.0.0.1。
- 6379：Redis，仅监听 127.0.0.1。

防火墙不是数据库弱密码或无认证的补救措施。监听地址、账户最小权限、认证和防火墙需要同时正确。

### 7.4 UFW 与 Docker 的边界

Docker 会管理自己的 nftables/iptables 规则。发布容器端口后，流量路径可能不完全符合你对 UFW 的直觉。部署容器时应：

- 不需要宿主机访问的端口不要 publish。
- 仅供本机反代时绑定到 127.0.0.1，例如 127.0.0.1:8080:8080。
- 用 ss、docker ps 和实际外部连接共同验证。

Docker 细节放到容器章节展开。

### 7.5 nftables 是底层能力

Ubuntu 的 UFW 通常以 nftables 为后端。初学阶段用 UFW 管理规则即可：

~~~bash
sudo nft list ruleset
~~~

可以查看底层规则，但不要同时手工维护一套互相冲突的 nftables 规则和 UFW 规则。

---

## 8. 实验：让宿主机访问虚拟机中的 Go 服务

### 8.1 准备最小服务

新建一个临时目录：

~~~bash
mkdir -p ~/labs/net-health
cd ~/labs/net-health
go mod init net-health
~~~

main.go：

~~~go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"time":   time.Now().UTC(),
		})
	})

	server := &http.Server{
		Addr:              "0.0.0.0:8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
~~~

启动：

~~~bash
go run .
~~~

### 8.2 在虚拟机内验证

另开终端：

~~~bash
ss -lntp 'sport = :8080'
curl -i http://127.0.0.1:8080/healthz
~~~

预期：

- ss 显示 0.0.0.0:8080 正在 LISTEN。
- curl 返回 HTTP 200 和 JSON。

### 8.3 放行指定来源

先确认 Windows 宿主机在 VMware 虚拟网络中的 IP，再精确放行。假设是 192.168.80.1：

~~~bash
sudo ufw allow from 192.168.80.1 to any port 8080 proto tcp
sudo ufw status numbered
~~~

在 Windows PowerShell：

~~~powershell
Test-NetConnection 192.168.80.128 -Port 8080
curl.exe -i http://192.168.80.128:8080/healthz
~~~

预期 TCP 测试成功，curl 返回 200。

### 8.4 观察监听地址造成的差异

把 Go 服务地址改成 127.0.0.1:8080，重启服务：

~~~bash
go run .
~~~

虚拟机内的 curl 仍成功，Windows 宿主机访问应失败。此时即使 UFW 允许 8080，也不会让一个仅监听回环地址的服务变成外部可达。

实验结束后删除临时防火墙规则：

~~~bash
sudo ufw status numbered
sudo ufw delete 规则编号
~~~

---

## 9. 一套可复用的排障流程

假设用户报告 https://s.example.com/abc 无法访问。

### 第一步：进程和服务

~~~bash
systemctl status shortlink --no-pager
journalctl -u shortlink -n 100 --no-pager
~~~

### 第二步：监听

~~~bash
sudo ss -lntp 'sport = :8080'
~~~

确认是 127.0.0.1:8080、0.0.0.0:8080，还是根本没有监听。

### 第三步：绕过 Nginx

~~~bash
curl -sS -f -v \
  --connect-timeout 2 \
  --max-time 5 \
  http://127.0.0.1:8080/healthz
~~~

如果这里失败，先处理 Go 服务，不要改 DNS。

### 第四步：验证 Nginx 本机入口

~~~bash
curl -v -H 'Host: s.example.com' http://127.0.0.1/healthz
~~~

如果 Go 成功而 Nginx 返回 502，检查 upstream 地址、端口和 Nginx 日志。

### 第五步：验证 DNS

~~~bash
getent ahosts s.example.com
resolvectl query s.example.com
~~~

确认 A/AAAA 记录是否指向预期服务器。若发布了 AAAA 记录，也必须保证 IPv6 路由、防火墙和监听正确；否则部分客户端会优先尝试错误的 IPv6。

### 第六步：验证边界

~~~bash
sudo ufw status verbose
sudo ss -lntp
~~~

同时检查云安全组。云安全组无法从 Ubuntu 命令中完整看到。

### 第七步：从外部测试

~~~bash
curl -v --connect-timeout 3 --max-time 10 https://s.example.com/healthz
~~~

把报错准确归类为解析、建连、TLS、HTTP 状态或响应内容问题。

---

## 10. 常见错误与边界

### 错误一：服务启动了，所以外部一定能访问

启动成功只证明进程没有立即退出。还要确认监听地址、端口、防火墙和路由。

### 错误二：把所有服务都监听到 0.0.0.0

0.0.0.0 扩大了可达面。只供同机组件使用的 Gin、MySQL、Redis应优先监听回环地址。

### 错误三：为解决问题直接关闭防火墙

可以通过临时、精确规则做诊断，不应长期关闭所有防护。改动前记录现状，改动后立即验证。

### 错误四：ping 不通就认定 HTTP 不通

ICMP 和 TCP/HTTPS 可以有不同策略。用与真实服务相同的协议验证。

### 错误五：把 502 当成用户请求没到服务器

502 通常说明请求已到代理，只是代理访问上游失败。

### 错误六：忽略 IPv6

域名同时存在 A 和 AAAA 时，IPv6 配置错误会造成“有的人能访问、有的人很慢或失败”。用 dig 分别检查，并用 curl -4、curl -6 对比：

~~~bash
curl -4 -v https://s.example.com/healthz
curl -6 -v https://s.example.com/healthz
~~~

### 错误七：生产数据库端口向公网开放

即使设置了密码，也不应把 3306/6379 作为普通公网服务暴露。优先使用同机回环、私有网络、VPN、堡垒机或 SSH 临时隧道。

---

## 11. 本章验收

不查资料，尝试回答：

1. 127.0.0.1:8080 和 0.0.0.0:8080 的可达范围有什么区别？
2. 为什么宿主机访问 VMware NAT 虚拟机通常不需要端口转发？
3. curl 返回 404 时，为什么通常不应该继续排查防火墙？
4. TIME_WAIT 是否意味着应用还占着对应文件描述符？
5. getent、resolvectl、dig 各适合验证什么？
6. 同机 Nginx + Gin + MySQL + Redis，哪些端口应公开？
7. 为什么云安全组放行后，主机仍可能无法访问？

动手验收：

- 能用 ss 找到监听 8080 的进程。
- 能让服务分别监听 127.0.0.1 和 0.0.0.0，并解释外部访问差异。
- 能用 curl 输出状态码和总耗时。
- 能使用 --resolve 绕过 DNS 验证指定 HTTPS 服务器。
- 能按“进程 → 监听 → 本机 curl → 防火墙/安全组 → DNS/代理”的顺序定位问题。

做到这些，就已经具备 Go 后端实习中非常实用的网络排障能力。
