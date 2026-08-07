# 03 · 模型与数据库

## Link 字段

见 `internal/model/link.go`：

| 字段 | 含义 |
|------|------|
| `Code` | 短码，**uniqueIndex** |
| `LongURL` | 原始长链 |
| `ClickCount` | 跳转次数（异步 +1） |
| `CreatedAt` | 创建时间 |

`AutoMigrate` 在启动时执行（`repo.AutoMigrate`）。学习阶段够用；生产更常用版本化 SQL 迁移（以后再说）。

## 常用操作

| 方法 | SQL 直觉 |
|------|----------|
| `Create` | INSERT |
| `Where("code = ?", code).First` | SELECT 一条 |
| `UpdateColumn("click_count", gorm.Expr("click_count + 1"))` | 原子 +1 |

## 为何点击用异步

跳转接口要快：先 302，计数在 goroutine 里做。计数失败只打日志，**不影响用户跳转**（最终一致，面试可以说清楚取舍）。

## 库名

默认写入 Docker 库 `study` 的表 `links`。可用客户端确认：

```powershell
docker exec study-mysql mysql -uroot -proot123 -e "DESCRIBE study.links; SELECT code,long_url,click_count FROM study.links LIMIT 5;"
```
