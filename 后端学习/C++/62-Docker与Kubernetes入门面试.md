# Docker 与 Kubernetes 入门面试

> **文件编码**：UTF-8。Namespace/Cgroups、镜像、Pod/Service/Deployment、kubelet、滚动更新、C++ 服务部署
> **交叉阅读**：[53 OS](53-操作系统面试八股与口述模板.md) · [09 CMake](09-CMake与项目工程化.md) · [19 gRPC](19-gRPC与Protobuf工程化.md) · [59 etcd](59-分布式理论CAP-Raft与共识算法面试.md)

---

## 本章与前后章的关系

| 上一章 | 本章 | 下一章 |
|--------|------|--------|
| [61 线上排障](61-线上故障排查与性能诊断实战.md) | **本章** | [63 JWT 幂等](63-JWT认证与接口幂等性实战.md) |

**学习链扩展（51～63）**：

| [51 MySQL](51-MySQL原理与索引事务面试专章.md) | [52 Redis](52-Redis数据结构与缓存面试专章.md) | [53 OS](53-操作系统面试八股与口述模板.md) |
| [54 计网](54-计算机网络TCP与HTTP面试深度专章.md) | [55 笔试](55-大厂C++笔试选择题与代码输出陷阱题集.md) | [56 系统设计](56-系统设计案例库RPC-KV与限流秒杀.md) |
| [57 Kafka](57-消息队列Kafka与中间件面试专题.md) | [58 模拟面试](58-模拟面试完整流程与压测数据模板.md) | [59 分布式](59-分布式理论CAP-Raft与共识算法面试.md) |
| [60 抓包](60-抓包与网络排障Wireshark实战.md) | [61 排障](61-线上故障排查与性能诊断实战.md) | [62 K8s](62-Docker与Kubernetes入门面试.md) |
| [63 JWT 幂等](63-JWT认证与接口幂等性实战.md) | | |

```mermaid
flowchart LR
  A[51 MySQL] --> B[52 Redis]
  B --> C[53 OS]
  C --> D[54 计网]
  D --> E[55 笔试]
  E --> F[56 系统设计]
  F --> G[57 Kafka]
  G --> H[58 模拟]
  H --> I[59 分布式]
  I --> J[60 抓包]
  J --> K[61 排障]
  K --> L[62 K8s]
  L --> M[63 JWT]
```

---

## §0 读前导读

### §0.1 用一句话弄懂本章

K8s 面试 = **容器底层（Namespace/Cgroups）→ 镜像与 Dockerfile → 工作负载（Pod/Deploy/Service）→ 调度与发布 → C++ 服务工程化**；与 [59 etcd](59-分布式理论CAP-Raft与共识算法面试.md) 控制面、[61 排障](61-线上故障排查与性能诊断实战.md) 容器场景衔接。

### §0.2 你需要提前知道什么

| 状态 | 动作 |
|------|------|
| 只会用不会讲 | 每节 Q&A 限时 2min 口述 |
| C++ 后端岗 | 必串 [08 多线程](08-多线程与并发编程.md) [10 网络](10-网络编程与简易HTTP服务.md) [23 IO](23-IO多路复用与高性能Server.md) |
| 前置章节 | [61 排障](61-线上故障排查与性能诊断实战.md) cgroup/OOM |
| 后续章节 | [63 JWT](63-JWT认证与接口幂等性实战.md) 网关部署 |

### §0.3 本章知识地图（☐→☑）

- ☐ 模块 1 能闭卷口述
- ☐ 模块 2 能闭卷口述
- ☐ 模块 3 能闭卷口述
- ☐ 模块 4 能闭卷口述
- ☐ 模块 5 能闭卷口述
- ☐ 模块 6 能闭卷口述
- ☐ 模块 7 能闭卷口述
- ☐ 模块 8 能闭卷口述

### §0.4 建议节奏

| 阶段 | 时长 | 内容 |
|------|------|------|
| 首轮通读 | 3h | §1～§N Q&A |
| 二轮口述 | 2h | 口述稿录音 |
| 交叉刷 | 2h | 51～58 + C++ 工程文档 |
| 闭卷自测 | 1h | ≥8/10 |

---

## §1 Namespace 与 Cgroups

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | Docker 与虚拟机区别？ | 容器共享内核、隔离 namespace/cgroup；VM 独立内核更重。 | 隔离强度 |
| Q2 | Linux Namespace 类型？ | pid/net/mnt/ipc/uts/user/cgroup；Docker 组合隔离。 | [53 OS](53-操作系统面试八股与口述模板.md) |
| Q3 | Cgroups 作用？ | 限制 CPU/内存/IO；docker stats 来源。 | OOM 行为 |
| Q4 | Union FS 镜像层？ | 只读层叠加+写时复制 cow；docker image 分层缓存。 | 构建优化 |
| Q5 | 容器进程在宿主机可见吗？ | 可见不同 PID namespace；ps 能看到。 | 安全边界 |
| Q6 | Capabilities？ | 降权 root；CAP_NET_BIND_SERVICE 绑 80。 | seccomp/AppArmor |
| Q7 | overlay2 驱动？ | 主流存储驱动；inode 耗尽问题。 | 清理 dangling |
| Q8 | 容器网络 veth？ | 一对虚拟网卡连 bridge/CNI。 | [54 网络](54-计算机网络TCP与HTTP面试深度专章.md) |
| Q9 | pid1 问题？ | 僵尸进程需 init 回收；tini/dumb-init。 | C++ 多进程 |
| Q10 | 与 K8s 关系？ | K8s 调度多个容器；Docker 可只作 runtime（containerd）。 | dockershim 移除 |

## §2 镜像与 Dockerfile

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | Dockerfile 最佳实践？ | 多阶段构建、非 root、最小基础镜像、层缓存顺序。 | alpine vs distroless |
| Q2 | 多阶段 C++ 构建？ | builder 装 gcc+cmake → runtime 只拷二进制+so。 | [09 CMake](09-CMake与项目工程化.md) |
| Q3 | COPY vs ADD？ | COPY 简单；ADD 可解压不推荐。 | 可重复构建 |
| Q4 | 镜像 tag 策略？ | semver+git sha；勿只用 latest 生产。 | immutable |
| Q5 | 私有仓库？ | Harbor/ECR；imagePullSecrets。 | 漏洞扫描 |
| Q6 | distroless 好处？ | 无 shell 减攻击面；调试靠 kubectl debug。 | [61 排障](61-线上故障排查与性能诊断实战.md) |
| Q7 | 静态链接 C++？ | 减少 runtime 依赖；体积 trade-off。 | glibc vs musl |
| Q8 | HEALTHCHECK？ | Docker 自身健康；K8s 用 probe。 | 启动慢服务 |
| Q9 | 构建缓存失效？ | 频繁改 COPY 前置层；依赖清单单独层。 | CI 加速 |
| Q10 | SBOM 安全？ | 扫描 CVE；基础镜像升级流程。 | 供应链 |

## §3 Pod / Service / Deployment

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | Pod 是什么？ | 最小调度单元；一或多容器共享 net/ipc/volume。 | Sidecar 模式 |
| Q2 | Deployment 作用？ | 声明式 ReplicaSet；滚动更新/回滚。 | 与 StatefulSet |
| Q3 | Service 类型？ | ClusterIP/NodePort/LoadBalancer；kube-proxy iptables/ipvs。 | Headless |
| Q4 | ConfigMap/Secret？ | 配置与敏感数据挂载；Secret base64 非加密。 | 外部 Secret 管理 |
| Q5 | Ingress？ | 七层路由 TLS；与 Service 配合。 | Nginx/Traefik |
| Q6 | Namespace 隔离？ | 逻辑多租户；RBAC 边界。 | 非硬隔离 |
| Q7 | Label/Selector？ | Deployment 关联 Pod；Service Endpoints。 | 推荐规范 |
| Q8 | PV/PVC？ | 持久卷；Stateful 数据库。 | C++ 无状态优先 |
| Q9 | DaemonSet？ | 每节点一份；日志/agent。 | node exporter |
| Q10 | Job/CronJob？ | 批处理；定时任务。 | 幂等 [63 章](63-JWT认证与接口幂等性实战.md) |

## §4 kubelet 与滚动更新

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | kubelet 职责？ | Pod 生命周期、挂载卷、探针、上报节点状态。 | CRI 接口 |
| Q2 | 三类探针？ | liveness 重启；readiness 摘流量；startup 慢启动。 | 误杀案例 |
| Q3 | 滚动更新策略？ | maxSurge/maxUnavailable；逐步替换。 | 零 downtime |
| Q4 | 回滚？ | kubectl rollout undo；Revision 历史。 | 变更记录 |
| Q5 | 资源 request/limit？ | 调度按 request；超限 throttle/OOMKill。 | [61 CPU](61-线上故障排查与性能诊断实战.md) |
| Q6 | HPA？ | CPU/自定义指标扩缩；需 metrics-server。 | 冷却时间 |
| Q7 | 调度器 predicate/priority？ | 资源/亲和/污点容忍。 | 亲和 spread |
| Q8 | etcd 存储什么？ | K8s 全对象；[59 章](59-分布式理论CAP-Raft与共识算法面试.md) Raft。 | 备份 |
| Q9 | CNI 插件？ | Calico/Flannel 分配 pod IP。 | NetworkPolicy |
| Q10 | QoS Class？ | Guaranteed/Burstable/BestEffort；OOM 顺序。 | 关键服务 Guaranteed |

## §5 C++ 服务部署实战

| 编号 | 面试问题 | 标准答法（口述版） | 追问/坑点 |
|------|----------|-------------------|-----------|
| Q1 | C++ 服务容器化步骤？ | CMake Release 构建→多阶段镜像→探针→ConfigMap→Deployment+Service。 | [35 KV](35-项目实战高性能KV-Store.md) |
| Q2 | graceful shutdown？ | SIGTERM 处理；preStop sleep；drain 连接。 | [23 服务器](23-IO多路复用与高性能Server.md) |
| Q3 | gRPC 上 K8s？ | ClusterIP + headless for client LB；或 mesh。 | [19 gRPC](19-gRPC与Protobuf工程化.md) |
| Q4 | 环境变量配置？ | 12-factor；flags vs env。 | 热更新限制 |
| Q5 | 日志 stdout？ | 集群 ELK/Loki 采集；勿写容器层文件。 | [32 可观测](32-fmt-spdlog与可观测性工程.md) |
| Q6 | CPU 绑核？ | limits 导致 throttle；适当 request=limit。 | 性能测试 |
| Q7 | 内存 limit 设置？ | 留 ASan 以外 headroom；C++ RSS 含 cache。 | OOM 调优 |
| Q8 | Init 容器？ | 等依赖 ready；迁移 job。 | 启动顺序 |
| Q9 | Helm Chart？ | 模板化多环境 values。 | GitOps |
| Q10 | CI/CD？ | build→scan→push→helm upgrade；蓝绿/金丝雀。 | [58 压测](58-模拟面试完整流程与压测数据模板.md) |

## §6 STAR 案例

### §6.1 STAR 案例 1

**S（情境）**：C++ 交易网关发布后出现 502，滚动更新中。

**T（任务）**：完成无感知发布。

**A（行动）**：readiness 等 Warmup；preStop 30s；maxUnavailable 0；gRPC graceful。

**R（结果）**：错误率 0；发布窗口 10min。

**连环追问**：
- liveness 误杀？
- 镜像拉取慢？
- 探针路径？
## §7 Dockerfile 与 K8s YAML 示例

```dockerfile
# 多阶段 C++ 服务示例（见 09 CMake）
FROM gcc:13 AS builder
WORKDIR /src
COPY . .
RUN cmake -B build -DCMAKE_BUILD_TYPE=Release && cmake --build build -j

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends libstdc++6 ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /src/build/my_service /usr/local/bin/
USER 65534:65534
EXPOSE 8080
HEALTHCHECK CMD curl -f http://127.0.0.1:8080/health || exit 1
ENTRYPOINT ["/usr/local/bin/my_service"]
```

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cpp-gateway
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: cpp-gateway
  template:
    metadata:
      labels:
        app: cpp-gateway
    spec:
      containers:
      - name: gateway
        image: registry.example.com/cpp-gateway:v1.2.0
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "2"
            memory: "2Gi"
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 15
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 15"]
```

## §9 口述模板（2 分钟版）

### §9.1 容器 30 秒

Namespace 隔离视图，Cgroups 限资源；镜像分层 cow；K8s 编排调度。

### §9.2 滚动更新 45 秒

Deployment 调 maxSurge/Unavailable；readiness 合格才接流量；失败 rollout undo。

### §9.3 C++ 部署 45 秒

多阶段构建减体积；SIGTERM graceful；request/limit 压测标定；日志 stdout。
## §10 闭卷自测清单

- [ ] 能画 Pod 与 Service 关系
- [ ] 能画 Pod 与 Service 关系
- [ ] 能画 Pod 与 Service 关系
- [ ] 能解释 probe 差异
- [ ] 能写多阶段 Dockerfile
- [ ] 能关联 [59 etcd](59-分布式理论CAP-Raft与共识算法面试.md)
- [ ] 能口述滚动更新参数
**下一章**：[63-JWT认证与接口幂等性实战.md](63-JWT认证与接口幂等性实战.md)
### 附录 D.1 K8s 面试追问 1

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.2 K8s 面试追问 2

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.3 K8s 面试追问 3

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.4 K8s 面试追问 4

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.5 K8s 面试追问 5

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.6 K8s 面试追问 6

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.7 K8s 面试追问 7

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.8 K8s 面试追问 8

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.9 K8s 面试追问 9

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.10 K8s 面试追问 10

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.11 K8s 面试追问 11

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.12 K8s 面试追问 12

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.13 K8s 面试追问 13

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.14 K8s 面试追问 14

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.15 K8s 面试追问 15

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.16 K8s 面试追问 16

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.17 K8s 面试追问 17

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.18 K8s 面试追问 18

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.19 K8s 面试追问 19

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.20 K8s 面试追问 20

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.21 K8s 面试追问 21

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.22 K8s 面试追问 22

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.23 K8s 面试追问 23

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.24 K8s 面试追问 24

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.25 K8s 面试追问 25

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.26 K8s 面试追问 26

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.27 K8s 面试追问 27

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.28 K8s 面试追问 28

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.29 K8s 面试追问 29

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.30 K8s 面试追问 30

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.31 K8s 面试追问 31

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.32 K8s 面试追问 32

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.33 K8s 面试追问 33

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.34 K8s 面试追问 34

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.35 K8s 面试追问 35

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.36 K8s 面试追问 36

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.37 K8s 面试追问 37

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.38 K8s 面试追问 38

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.39 K8s 面试追问 39

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.40 K8s 面试追问 40

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。

### 附录 D.41 K8s 面试追问 41

**Q**：Deployment 与 StatefulSet 何时选？

**A**：无状态 C++ API 用 Deployment；需稳定网络标识与持久盘用 StatefulSet（如 etcd、Kafka broker）。见 [57 Kafka](57-消息队列Kafka与中间件面试专题.md)。


## 附录扩展 Q&A（自测用）

### 自测 1

**Q1**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 1**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 2

**Q2**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 2**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 3

**Q3**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 3**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 4

**Q4**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 4**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 5

**Q5**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 5**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 6

**Q6**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 6**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 7

**Q7**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 7**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 8

**Q8**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 8**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 9

**Q9**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 9**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 10

**Q10**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 10**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 11

**Q11**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 11**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 12

**Q12**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 12**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 13

**Q13**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 13**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 14

**Q14**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 14**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 15

**Q15**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 15**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 16

**Q16**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 16**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 17

**Q17**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 17**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 18

**Q18**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 18**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 19

**Q19**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 19**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 20

**Q20**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 20**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 21

**Q21**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 21**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 22

**Q22**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 22**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 23

**Q23**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 23**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 24

**Q24**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 24**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 25

**Q25**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 25**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 26

**Q26**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 26**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 27

**Q27**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 27**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 28

**Q28**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 28**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 29

**Q29**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 29**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 30

**Q30**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 30**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 31

**Q31**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 31**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 32

**Q32**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 32**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 33

**Q33**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 33**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 34

**Q34**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 34**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 35

**Q35**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 35**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 36

**Q36**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 36**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 37

**Q37**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 37**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 38

**Q38**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 38**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 39

**Q39**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 39**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 40

**Q40**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 40**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 41

**Q41**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 41**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 42

**Q42**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 42**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 43

**Q43**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 43**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 44

**Q44**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 44**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 45

**Q45**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 45**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 46

**Q46**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 46**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 47

**Q47**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 47**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 48

**Q48**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 48**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 49

**Q49**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 49**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 50

**Q50**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 50**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 51

**Q51**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 51**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 52

**Q52**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 52**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 53

**Q53**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 53**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 54

**Q54**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 54**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 55

**Q55**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 55**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 56

**Q56**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 56**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 57

**Q57**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 57**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

### 自测 58

**Q58**：请结合本章与 [51～58 章](51-MySQL原理与索引事务面试专章.md) 口述一个 2 分钟答案。

**参考要点 58**：定义 → 原理 → 工程 trade-off → C++ 落地 → 与相邻章节交叉。

