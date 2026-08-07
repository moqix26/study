# 05 · 创建接口 POST /api/links

## 请求

```json
{"url":"https://www.bilibili.com"}
```

## 处理步骤

1. `ShouldBindJSON`  
2. `urlx.Normalize`：必须有 scheme+host，且 http/https  
3. `service.Create` → 短码 + INSERT  
4. 201 返回：

```json
{
  "code": "aB3xY9",
  "short_url": "http://localhost:8080/aB3xY9",
  "long_url": "https://www.bilibili.com"
}
```

## PowerShell

```powershell
Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST -ContentType "application/json" -Body '{"url":"https://www.bilibili.com"}'
```

## 错误

| 情况 | 状态码 |
|------|--------|
| JSON 坏了 | 400 |
| URL 不合法 | 400 |
| DB 挂了 | 500 |

代码：`handler.CreateLink` + `service.Create`。
