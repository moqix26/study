# 08 · 验收清单

## 启动

```powershell
cd F:\study\Code\shortlink
docker start study-mysql
docker start study-redis
go run .
```

## 接口

```powershell
# 健康检查
Invoke-RestMethod http://localhost:8080/health

# 创建
$r = Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST -ContentType "application/json" -Body '{"url":"https://www.bilibili.com"}'
$r
$code = $r.code

# 跳转两次
curl.exe -i "http://localhost:8080/$code"
curl.exe -i "http://localhost:8080/$code"

# JSON + 缓存头
curl.exe -i "http://localhost:8080/api/links/$code"
```

## 期望

| 检查 | 期望 |
|------|------|
| 创建 | 201，有 code / short_url |
| 第一次 curl -i | 302，`X-Cache: MISS`，`Location` 正确 |
| 第二次 | 302，`X-Cache: HIT` |
| Redis | `GET link:{code}` 有长链 |
| 点击 | `links.click_count` 增加（可能略延迟） |

## Redis / MySQL 抽查

```powershell
docker exec study-redis redis-cli GET link:$code
docker exec study-mysql mysql -uroot -proot123 -e "SELECT code,click_count FROM study.links WHERE code='$code';"
```
