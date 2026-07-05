# 扩充进度（更新于 全库扩充完成）

## 图例
- ✅ 已按 EXPANSION-STANDARD 扩充（§0 导读、FAQ、闭卷自测、费曼检验等）

---

## ✅ 全部用户相关系列 — 完成

| 系列 | 进度 | 章节 |
|------|------|------|
| **HTML/CSS/JS** | **15/15** | 00～14 |
| **Vue** | **14/14** | 01～14 |
| **React** | **15/15** | 00～14 |
| **TypeScript** | **12/12** | 00～11 |
| **计网** | **8/8** | 00～07 |
| **Git** | **6/6** | 00～05 |
| **Java** | **17/17** | 01～17 |
| **Python** | **16/16** | 00～15 |
| **C++** | **16/16** | 00～15 |
| **Linux** | **16/16** | 00～15 |
| **AIAgent** | **25/25** | 00～24（13～16 工程深挖、17～22 大模型岗底层+前沿、23～24 项目与面试冲刺，2026-06 新增） |
| **Web安全** | **8/8** | 00～07 |
| **系统设计** | **11/11** | 00～10 |
| **数据结构** | **13/13** | 00～12 |
| **浏览器与性能** | **7/7** | 00～06 |

**合计：约 204 篇 md 学习资料 — 扩充覆盖率 ~100%**

---

## 扩充标准（每章典型结构）

1. **§0 读前导读**（一句话、前置、知识地图、节奏、可验证成果）
2. **生活类比** + 术语三件套
3. **手把手步骤表**（步骤 | 动作 | 预期 | 若不对）
4. **逐行读** 主代码/配置
5. **FAQ ≥10**
6. **闭卷自测 10 题 + 参考答案**
7. **费曼检验**

规范：`EXPANSION-STANDARD.md`

---

## 你的主线路（Agent + Java + 全栈）

```text
HTML 00→05 → JS 06→10 → Java 01→07 → AIAgent 01→10
→ Vue 01→08 联调 → Git/Linux/计网/Web安全
→ 数据结构 + Java 13 → 系统设计 + Java 14
```

**可选加深**：TypeScript 01～06（Vue 前）| C++（ACM 对照）| Python（第二后端）

---

## 源文件深度增强（面试深挖，2026-06-30）

在现有章节**源文件**末尾"下一章预告"前插入「面试深挖补充」大节，补足"面试深挖但原文只到基础认知"的底层原理。未新增文件，未改原有编号，与相关小节对照阅读。

| 章节 | 补充的深挖点 |
|------|--------------|
| `后端学习/Java/03-Java并发编程与JVM.md` | 锁升级（偏向→轻量→重量，JDK15废弃偏向锁）/ AQS（state+CLH队列+模板方法）/ 线程池执行时序（core→queue→max→reject）/ CAS的ABA与AtomicStampedReference / ConcurrentHashMap JDK7分段锁 vs JDK8桶级锁+红黑树 / G1(Region+GarbageFirst+RSet)与ZGC(染色指针+读屏障) / 打破双亲委派(SPI上下文类加载器/Tomcat/OSGi) |
| `后端学习/Java/07-Redis核心原理与缓存实战.md` | 五种数据结构底层（SDS/listpack替代ziplist/quicklist/skiplist+dict/intset/编码切换）/ 单线程模型+6.0多线程IO / 过期删除(惰性+定期)与内存淘汰(8种+近似LRU+LFU Morris) / 哨兵(主观/客观下线+Raft选leader)与Cluster(16384槽+CRC16+Gossip+MOVED) / RedLock算法与Kleppmann争议+Redisson看门狗续期 / AOF重写+混合持久化(RDB前段+AOF后段) |
| `后端学习/Java/06-MySQL基础索引与事务.md` | MVCC的ReadView机制(隐藏字段trx_id/roll_ptr+版本链+m_ids/min/max/creator可见性判断+RC/RR生成时机) / 索引下推ICP(5.6+，Using index condition) / 间隙锁/临键锁(RR专属，等值命中唯一索引退化为记录锁、未命中退化间隙锁) / change buffer(仅非唯一索引)+自适应哈希索引AHI+MRR(随机IO转顺序IO) |
| `后端学习/Linux/06-进程与服务管理.md` | 进程状态机(R/S/D/Z/T+修饰位，D状态kill-9无效原因) / 僵尸进程(父进程没wait，占PID)+孤儿进程(init收养无害) / OOM Killer(oom_score打分+oom_score_adj+Java常被杀+与JVM堆OOM区别) / SIGCHLD与wait根因 / systemd unit编写([Unit]/[Service]/[Install]+Type类型+Wants/Requires/After+daemon-reload+SuccessExitStatus=143) |
| `后端学习/Linux/07-网络命令与防火墙基础.md` | TCP状态全景(LISTEN→ESTABLISHED→TIME_WAIT/CLOSE_WAIT) / TIME_WAIT堆积(2MSL原因+耗端口+长连接/tcp_tw_reuse/禁用tcp_tw_recycle) / CLOSE_WAIT堆积(必是应用bug没close+排查) / ss vs netstat(netlink快于遍历/proc) / tcpdump抓包(三次握手S/S./.、四次挥手F./.) |

每节均含「一句话+原理+对照表+面试标准答法+与其它深挖点的关联+面试自检」，与 EXPANSION-STANDARD 风格一致。

---

## 维护

- 事实核查：`verified-facts/spring-ai.md`
- 接力：`HANDOFF-for-next-session.md`

每章 **闭卷自测 ≥7/10** 再进下一章。
