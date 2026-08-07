# 04 · 短码生成

## 算法

`internal/pkg/shortcode`：从 62 个字符（0-9a-zA-Z）里用 `crypto/rand` 抽 `CodeLength`（默认 6）次。

- 空间约 \(62^6\)，练习足够  
- 用密码学随机，避免可预测短码被扫

## 碰撞怎么办

数据库 `uniqueIndex` 兜底。`service.Create` 循环最多 `MaxRetries` 次：

1. 生成 code  
2. INSERT  
3. 若唯一冲突 → 再生成  
4. 其它错误 → 直接失败  

**口述**：应用层重试 + DB 唯一约束，双保险。

## 不要做的

- 不要用自增 id 当短码直接暴露（可枚举）  
- V1 不必上雪花 ID；讲清随机+唯一即可  
