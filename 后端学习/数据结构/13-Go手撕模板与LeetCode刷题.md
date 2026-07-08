# Go 手撕模板与 LeetCode 刷题

> **文件编码**：UTF-8。题解默认 **Go 1.22+**；配合 [go-backend-learning-plan.md](../../go-backend-learning-plan.md) 暑假 **80 题**目标。  
> **读者**：CCPC 省金 + ICPC 省银背景、主攻 Go 后端——竞赛思维保留，语言与面试表达切换到 Go。

---

## 本章与上一章的关系

[01～10 章](00-学习路线图与说明.md) 讲 **结构原理**（默认 Python）；[11 章](11-LeetCode刷题路线与题型汇总.md) 提供 **70 题 GPS**；本章提供 **Go 手撕模板 + Hot 100 映射 + 暑假 80 题节奏**，题号与 [Java 13](../Java/13-算法与数据结构基础.md) / [C++ 13](../C++/13-算法与数据结构C++实现.md) 对齐。

| 模块 | 定位 |
|------|------|
| 数据结构 11 | 70 题顺序、八周计划 |
| **本章 13** | Go 模板 + Hot 100 + 80 题 |
| [Go 04 并发](../Go/04-Go并发编程goroutine与channel.md) | goroutine/channel（W2 并行） |

---

## 0. 读前导读

### 0.1 一句话

用 **slice/map + container/heap** 替代 C++ STL，按模板刷够 **80 题**，面试 **25 分钟** 白板写 Go。

### 0.2 知识地图（☐→☑）

- [ ] 闭卷：两数之和、反转链表、有效括号、层序遍历、二分、TopK
- [ ] 暑假 Go 提交 **80 题**（11 章 70 + 扩展 10）
- [ ] §16 闭卷自测 ≥8/10

### 0.3 暑假 80 题周均（对齐 go-backend-learning-plan §8.3）

| 周次 | 题量 | 侧重 |
|------|------|------|
| W1 | 7 | 1 题/天，熟悉 Go 提交 |
| W2 | 10 | 双指针、哈希 |
| W3 | 10 | 链表、栈 |
| W4 | 10 | 树 DFS/BFS |
| W5 | 10 | 二分、滑动窗口 |
| W6 | 10 | 堆、并查集 |
| W7 | 11 | 回溯、图 |
| W8 | 12 | 扩展题 + 错题 |
| **合计** | **80** | W2～W8 约 2 题/天 |

**每日节奏**（与总计划块 3 一致）：模板复习 20 min → 1～2 题 60 min → 复盘 10 min。

---

## 1. ACM C++ → Go 迁移（竞赛背景）

| C++ | Go | 注意 |
|-----|-----|------|
| `vector<int>` | `[]int` | `make([]int, n)` 定长 |
| `unordered_map` | `map[K]V` | key 需 comparable |
| `priority_queue` | `container/heap` | 实现 `heap.Interface` |
| `set` / `lower_bound` | `sort` + 二分 / 双指针 | 无内置平衡树 |
| `stack` | `st := []int{}` append/pop | 空栈判 `len==0` |
| 链表 `ListNode*` | `*ListNode` | `nil` 非 nullptr |
| `<algorithm>` | **Go 1.18 前无泛型算法库** | 手写模板；1.18+ 可用 `slices.Sort` |

**面试三句话**：① 思路同竞赛，语言换 map/slice；② 堆用 `container/heap`；③ 先报边界再写循环。

**并行学习**：算法用本章；并发用 [Go 04](../Go/04-Go并发编程goroutine与channel.md)，**分时段**不混练。

---

## 2. 提交骨架（可本地 `go run`）

```go
package main

func twoSum(nums []int, target int) []int {
	seen := make(map[int]int)
	for i, v := range nums {
		if j, ok := seen[target-v]; ok {
			return []int{j, i}
		}
		seen[v] = i
	}
	return nil
}

func main() {
	println(twoSum([]int{2, 7, 11, 15}, 9)) // [0 1]
}
```

LeetCode 提交时删 `main`，保留题面要求的 `type` 与 `func`。

---

## 3. 双指针（§4）

**适用**：[02 章](02-数组与字符串.md) — 167、26、876、141。

```go
// 对撞指针 — 167 两数之和 II
func twoSumSorted(numbers []int, target int) []int {
	left, right := 0, len(numbers)-1
	for left < right {
		sum := numbers[left] + numbers[right]
		switch {
		case sum == target:
			return []int{left + 1, right + 1}
		case sum < target:
			left++
		default:
			right--
		}
	}
	return nil
}

// 快慢指针 — 26 删除有序重复项
func removeDuplicates(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	slow := 1
	for fast := 1; fast < len(nums); fast++ {
		if nums[fast] != nums[fast-1] {
			nums[slow] = nums[fast]
			slow++
		}
	}
	return slow
}
```

---

## 4. 滑动窗口（§5）

**适用**：[02 章](02-数组与字符串.md) — 3、76、438。

```go
// 可变窗口 — 3 无重复字符最长子串
func lengthOfLongestSubstring(s string) int {
	cnt := make(map[byte]int)
	left, ans := 0, 0
	for right := 0; right < len(s); right++ {
		c := s[right]
		cnt[c]++
		for cnt[c] > 1 {
			cnt[s[left]]--
			left++
		}
		if right-left+1 > ans {
			ans = right - left + 1
		}
	}
	return ans
}
```

固定窗口（438）：窗口长 `len(p)`，用 `[26]int` 比较频次，见 [11 章 #18](11-LeetCode刷题路线与题型汇总.md)。

---

## 5. 二分查找（§6）

**适用**：[09 章](09-排序与查找算法.md) — 704、34、33。

```go
// 标准二分 — 704
func search(nums []int, target int) int {
	left, right := 0, len(nums)-1
	for left <= right {
		mid := left + (right-left)/2
		switch {
		case nums[mid] == target:
			return mid
		case nums[mid] < target:
			left = mid + 1
		default:
			right = mid - 1
		}
	}
	return -1
}

// 左边界 — 34
func lowerBound(nums []int, target int) int {
	left, right := 0, len(nums)
	for left < right {
		mid := left + (right-left)/2
		if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid
		}
	}
	return left
}
```

---

## 6. 链表（§7）

**适用**：[03 章](03-链表.md) — 206、21、142。

```go
type ListNode struct {
	Val  int
	Next *ListNode
}

// 206 反转链表
func reverseList(head *ListNode) *ListNode {
	var prev *ListNode
	for cur := head; cur != nil; cur = cur.Next {
		nxt := cur.Next
		cur.Next = prev
		prev = cur
	}
	return prev
}

// 21 合并两个有序链表（dummy）
func mergeTwoLists(l1, l2 *ListNode) *ListNode {
	dummy := &ListNode{}
	tail := dummy
	for l1 != nil && l2 != nil {
		if l1.Val <= l2.Val {
			tail.Next, l1 = l1, l1.Next
		} else {
			tail.Next, l2 = l2, l2.Next
		}
		tail = tail.Next
	}
	if l1 != nil {
		tail.Next = l1
	} else {
		tail.Next = l2
	}
	return dummy.Next
}

// 142 环入口（Floyd）
func detectCycle(head *ListNode) *ListNode {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow, fast = slow.Next, fast.Next.Next
		if slow == fast {
			for ptr := head; ptr != slow; ptr, slow = ptr.Next, slow.Next {
			}
			return slow
		}
	}
	return nil
}
```

---

## 7. 二叉树 DFS / BFS（§8）

**适用**：[06 章](06-树与二叉树.md) — 104、102、226。

```go
type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func maxDepth(root *TreeNode) int {
	if root == nil {
		return 0
	}
	return max(maxDepth(root.Left), maxDepth(root.Right)) + 1
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 102 层序遍历
func levelOrder(root *TreeNode) [][]int {
	if root == nil {
		return nil
	}
	var ans [][]int
	q := []*TreeNode{root}
	for len(q) > 0 {
		size := len(q)
		level := make([]int, 0, size)
		for i := 0; i < size; i++ {
			node := q[0]
			q = q[1:]
			level = append(level, node.Val)
			if node.Left != nil {
				q = append(q, node.Left)
			}
			if node.Right != nil {
				q = append(q, node.Right)
			}
		}
		ans = append(ans, level)
	}
	return ans
}
```

---

## 8. 堆 container/heap（§9）

**适用**：[07 章](07-堆与优先队列.md) — 347、215、23。

```go
import "container/heap"

type pair struct{ val, cnt int }
type minHeap []pair

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool  { return h[i].cnt < h[j].cnt }
func (h minHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)         { *h = append(*h, x.(pair)) }
func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func topKFrequent(nums []int, k int) []int {
	freq := make(map[int]int)
	for _, v := range nums {
		freq[v]++
	}
	h := &minHeap{}
	heap.Init(h)
	for val, cnt := range freq {
		heap.Push(h, pair{val, cnt})
		if h.Len() > k {
			heap.Pop(h)
		}
	}
	ans := make([]int, h.Len())
	for i := len(ans) - 1; i >= 0; i-- {
		ans[i] = heap.Pop(h).(pair).val
	}
	return ans
}
```

23 合并 K 链：堆存各链头，`Pop` 后 `Push(node.Next)`。

---

## 9. 并查集（§10）

**适用**：[10 章](10-并查集Trie与高级结构.md) — 547、684。

```go
type UnionFind struct {
	parent, rank []int
}

func NewUnionFind(n int) *UnionFind {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &UnionFind{parent: p, rank: make([]int, n)}
}

func (uf *UnionFind) Find(x int) int {
	if uf.parent[x] != x {
		uf.parent[x] = uf.Find(uf.parent[x])
	}
	return uf.parent[x]
}

func (uf *UnionFind) Union(a, b int) bool {
	ra, rb := uf.Find(a), uf.Find(b)
	if ra == rb {
		return false
	}
	if uf.rank[ra] < uf.rank[rb] {
		ra, rb = rb, ra
	}
	uf.parent[rb] = ra
	if uf.rank[ra] == uf.rank[rb] {
		uf.rank[ra]++
	}
	return true
}
```

---

## 10. 回溯（§11）

**适用**：[08 章](08-图论基础.md) — 46、78、39、22。

```go
func permute(nums []int) [][]int {
	var ans [][]int
	var cur []int
	used := make([]bool, len(nums))
	var dfs func()
	dfs = func() {
		if len(cur) == len(nums) {
			tmp := make([]int, len(cur))
			copy(tmp, cur)
			ans = append(ans, tmp)
			return
		}
		for i := 0; i < len(nums); i++ {
			if used[i] {
				continue
			}
			used[i] = true
			cur = append(cur, nums[i])
			dfs()
			cur = cur[:len(cur)-1]
			used[i] = false
		}
	}
	dfs()
	return ans
}
```

子集（78）：从 `start` 递归选/不选；组合总和（39）：排序 + 同层剪枝。

---

## 11. Hot 100 × 模板速查

先按 [11 章 70 题](11-LeetCode刷题路线与题型汇总.md) 刷，再 Hot 100 补缺口至 **80**。

| 模板 | Hot 100 题号 | 优先（★面试） |
|------|--------------|---------------|
| map / 哈希 | 1, 49, 128, 136, 169, 347, 560 | 1, 128, 347 |
| §3 双指针 | 11, 15, 75, 283, 42, 160, 234, 287 | 15, 283, 160 |
| §4 滑动窗口 | 3, 76, 438, 239, 567 | 3, 76, 438 |
| §5 二分 | 33, 34, 35, 153, 240, 287, 875 | 33, 34 |
| §6 链表 | 19, 21, 23, 141, 142, 148, 160, 206, 234 | 206, 21, 141, 142 |
| slice 栈 / 单调栈 | 20, 32, 84, 85, 155, 394, 739 | 20, 739 |
| §7 树 DFS/BFS | 94, 98, 101, 102, 104, 105, 108, 114, 226, 230, 236, 297, 437, 538, 543, 617 | 102, 104, 226, 236 |
| §8 堆 | 215, 253, 295, 347, 621 | 215, 347 |
| §9 并查集 / 图 | 200, 207, 399, 547 | 200, 207 |
| §10 回溯 / DFS 网格 | 17, 22, 39, 46, 78, 79, 131 | 46, 78, 79 |
| DP / 贪心 | 53, 55, 62, 64, 70, 72, 121, 139, 152, 198, 279, 300, 322, 416, 494 | 53, 70, 121, 300 |
| 设计 / 其他 | 31, 48, 56, 146, 208, 448, 461, 581, 647, 763 | 146, 208 |

**扩展 10 题凑 80**（11 章 §4）：53, 70, 200, 207, 146, 208, 547, 875 + Hot 中任选 2（如 239, 560）。

---

## 12. 与 11 章配合 & 常见坑

1. **顺序**：跟 [11 章 §2 八周计划](11-LeetCode刷题路线与题型汇总.md#2-八周刷题计划)，语言换 Go。  
2. **仓库**：`F:\study\code\leetcode-go\p0001\main.go` 一题目录。  
3. **坑**：空 slice/nil 链；二分 `mid := left+(right-left)/2`；回溯入 ans 要 `copy`；heap 的 `Less` 小根堆。

| 现象 | 解决 |
|------|------|
| TLE | 禁 O(n²)；字符串用 `strings.Builder` |
| 深递归 | 树改 BFS / 显式栈 |
| map 误判 | `v, ok := m[k]` |

---

## 13. 面试话术 & 学完标准

1. 澄清 n、可否改原数组 → 暴力 → 优化 → 报复杂度。  
2. TopK 挂钩 Redis ZSet；LRU 挂钩 [Go 08](../Go/08-Redis与go-redis缓存实战.md)。  
3. 竞赛成绩写简历，**手撕统一 Go**（[Go 15](../Go/15-Go面试专题与知识点总表.md)）。

- [ ] Go 提交 **80 题**；Hot 100 **≥60**  
- [ ] 25 min 一道 Medium；闭卷 §3～§10 核心模板

---

## 14. FAQ

**Q：C++ 还要练吗？** 简历保留 CCPC/ICPC；面试手撕 Go。  
**Q：和 Python 13 重复？** 题号故意重叠，做一遍，代码放 `leetcode-go`。  
**Q：Hard？** 暑假 Medium 为主；23/124/84 有余力再碰。

---

## 15. 文档索引

| 链接 | 内容 |
|------|------|
| [11 刷题路线](11-LeetCode刷题路线与题型汇总.md) | 70 题 GPS |
| [go-backend-learning-plan](../../go-backend-learning-plan.md) | 暑假总计划 |
| [Go 04 并发](../Go/04-Go并发编程goroutine与channel.md) | goroutine/channel |
| [Go 10 短链](../Go/10-短链服务项目实战上.md) | 项目实战 |

---

## 16. 闭卷自测

1. Go 如何替代 C++ `priority_queue`？  
2. 滑动窗口 vs 对撞指针的题面特征？  
3. 暑假 W1 为何 1 题/天？  
4. 142 环入口第二步为何 head 与 slow 同速？  
5. 70 题与 Hot 100 如何凑满 80？  
6. `heap.Push` 为何在指针接收者上？  
7. slice 头部插入复杂度？  
8. 并查集路径压缩作用？  
9. 回溯 `path` 为何要 copy？  
10. LeetCode 算法与 Go 并发如何分配时间？

<details>
<summary>参考答案</summary>

1. `container/heap` + 实现 `heap.Interface`。  
2. 窗口：连续子串/子数组约束；对撞：有序或两端收敛。  
3. 同步 Go 语法，不挤压 Go 04 与项目。  
4. Floyd 证明，同速相遇即入口。  
5. 先 11 章 70，再扩展/Hot 补 10。  
6. 需 append 修改底层 slice。  
7. O(n)。  
8. Find 均摊近常数。  
9. slice 共享底层，撤销会污染 ans。  
10. 块 3 刷题，块 1/2 语法与项目；见 go-backend-learning-plan §3。

</details>

---

**Go 轨 = 01～10 原理 + 本章模板 + 11 题单 80 题 + 短链能讲。**
