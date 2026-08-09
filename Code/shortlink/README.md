# shortlink · 短链服务 V1

本地可运行的 Go 短链：**创建短码 + 302 跳转 + Redis 缓存 + 点击计数**。

## 目录

```text
shortlink/
├── cmd/server/          # 推荐入口
├── internal/
│   ├── app/             # 依赖组装 + 启动
│   ├── config/          # 环境变量配置
│   ├── model/           # GORM 模型
│   ├── repo/            # MySQL
│   ├── cache/           # Redis
│   ├── service/         # 业务
│   ├── handler/         # HTTP
│   ├── middleware/
│   └── pkg/             # shortcode / urlx
├── configs/             # 配置示例
├── markdown/            # 逐模块详解
├── study.md             # 完整学习路线（跟这个学）
└── README.md
```

## 依赖

- Docker MySQL：`study-mysql`，主机端口 **3307**，库 `study`，密码示例 `root123`（仅本地）
- Docker Redis：`study-redis`，端口 **6379**

```powershell
docker start study-mysql
docker start study-redis
```

## 运行

```powershell
cd F:\study\Code\shortlink
go mod tidy
go run ./cmd/server
```

看到 `mysql ok` / `redis ok` / `:8080 is on`。

## 验收

```powershell
Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST -ContentType "application/json" -Body '{"url":"https://www.bilibili.com"}'

# 把返回的 code 换成真实短码
curl.exe -i http://localhost:8080/<code>
curl.exe -i http://localhost:8080/<code>
curl.exe -i http://localhost:8080/api/links/<code>
```

期望：第一次跳转 `X-Cache: MISS`，第二次 `HIT`；`Location` 为长链。

## 学习入口

1. **主教材**：打开 [`study.md`](./study.md)，按 S0→S5 聊天式推进  
2. **精读加餐**：[`markdown/00-index.md`](./markdown/00-index.md)（每篇都是逐块代码精讲）
3. 不要通读 markdown 当第二套主线；study 点名再打开

## 说明

- 练习密码写在默认配置里，**不要用于生产**；可用环境变量覆盖，见 `configs/config.example.env`。
