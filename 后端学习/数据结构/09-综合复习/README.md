# 综合复习、选型与面试检查表

## 1. 使用方式

本章不是第一次学习入口，而是完成各模块后用于：

- 快速恢复复杂度与适用前提；
- 检查是否只会背代码、不会解释；
- 做跨模块综合练习；
- 面试前按表查漏补缺。

## 2. 核心数据结构复杂度总表

| 结构 | 查询 | 插入 | 删除 | 额外说明 |
|---|---:|---:|---:|---|
| 数组 | 下标 `O(1)`；按值 `O(n)` | 中间 `O(n)` | 中间 `O(n)` | 连续内存、缓存友好 |
| 单链表 | `O(n)` | 已知位置后 `O(1)` | 已知前驱 `O(1)` | 不支持随机访问 |
| 栈/队列 | 顶/首 `O(1)` | `O(1)` | `O(1)` | LIFO / FIFO |
| 哈希表 | 平均 `O(1)` | 平均 `O(1)` | 平均 `O(1)` | 受冲突和负载因子影响 |
| BST | 平均 `O(log n)` | 平均 `O(log n)` | 平均 `O(log n)` | 退化时 `O(n)` |
| AVL/红黑树 | `O(log n)` | `O(log n)` | `O(log n)` | AVL 更严格，红黑树更新更平衡 |
| 堆 | 查任意值 `O(n)`；取顶 `O(1)` | `O(log n)` | 删除顶 `O(log n)` | 只保证父子偏序 |
| 并查集 | 近似 `O(1)` | 合并近似 `O(1)` | 不擅长删除 | `O(α(n))` 均摊 |
| Trie | `O(L)` | `O(L)` | `O(L)` | 空间与字符集、节点数有关 |

## 3. 排序与查找速查

| 算法 | 平均时间 | 最坏时间 | 空间 | 稳定性/前提 |
|---|---:|---:|---:|---|
| 冒泡 | `O(n²)` | `O(n²)` | `O(1)` | 稳定 |
| 选择 | `O(n²)` | `O(n²)` | `O(1)` | 不稳定 |
| 插入 | `O(n²)` | `O(n²)` | `O(1)` | 稳定，近乎有序时好 |
| 希尔 | 依增量而定 | 常写 `O(n²)` | `O(1)` | 不稳定 |
| 归并 | `O(n log n)` | `O(n log n)` | `O(n)` | 稳定 |
| 快速 | `O(n log n)` | `O(n²)` | 平均栈 `O(log n)` | 不稳定 |
| 堆排序 | `O(n log n)` | `O(n log n)` | `O(1)` | 不稳定 |
| 计数/桶/基数 | 依值域或位数 | 依实现 | 非原地为主 | 必须满足数据分布前提 |
| 二分查找 | `O(log n)` | `O(log n)` | `O(1)` | 数据有序或答案单调 |

## 4. 图算法速查

| 问题 | 算法 | 复杂度 | 关键前提 |
|---|---|---:|---|
| 无权最短步数 | BFS | `O(V+E)` | 边权相同或不计权 |
| 连通、路径存在、枚举 | DFS/BFS | `O(V+E)` | 正确维护 visited |
| DAG 依赖顺序 | Kahn/DFS 拓扑 | `O(V+E)` | 必须是 DAG 才有完整序列 |
| 无向图最小生成树 | Prim/Kruskal | `O(E log V)` / `O(E log E)` | 连通无向带权图 |
| 单源非负最短路 | Dijkstra | `O((V+E) log V)` | 不允许负权边 |
| 多源最短路 | Floyd | `O(V³)` | 顶点数较小，可处理负边但不能有负环 |

## 5. 选型决策

```text
需要按下标访问？                 → 数组
需要频繁从两端进出？             → 双端队列
需要 key → value 快速映射？       → 哈希表
需要始终有序并做范围查询？        → 平衡搜索树
只需反复取得最大/最小值？         → 堆
需要前缀匹配？                   → Trie
需要动态判断两个点是否连通？      → 并查集
需要表示任意关系网络？            → 图
答案空间具有单调性？              → 二分答案
连续区间满足可维护条件？          → 滑动窗口
枚举选择且可撤销？                → 回溯
存在重叠子问题和最优子结构？      → 动态规划
```

## 6. 三语言综合冒烟程序

程序组合三个知识点：二分查找、图 BFS、零钱兑换动态规划。它用于验证基础模板能否在一个程序中正确协作。

### 6.1 C++17

```cpp
#include <algorithm>
#include <iostream>
#include <limits>
#include <queue>
#include <vector>

using namespace std;

int binarySearch(const vector<int>& nums, int target) {
    int left = 0, right = static_cast<int>(nums.size()) - 1;
    while (left <= right) {
        int mid = left + (right - left) / 2;
        if (nums[mid] == target) return mid;
        if (nums[mid] < target) left = mid + 1;
        else right = mid - 1;
    }
    return -1;
}

vector<int> bfs(const vector<vector<int>>& graph, int start) {
    vector<int> order;
    vector<bool> visited(graph.size(), false);
    queue<int> q;
    q.push(start);
    visited[start] = true;
    while (!q.empty()) {
        int node = q.front();
        q.pop();
        order.push_back(node);
        for (int next : graph[node]) {
            if (!visited[next]) {
                visited[next] = true;
                q.push(next);
            }
        }
    }
    return order;
}

int coinChange(const vector<int>& coins, int amount) {
    const int inf = numeric_limits<int>::max() / 2;
    vector<int> dp(amount + 1, inf);
    dp[0] = 0;
    for (int value = 1; value <= amount; ++value) {
        for (int coin : coins) {
            if (coin <= value) dp[value] = min(dp[value], dp[value - coin] + 1);
        }
    }
    return dp[amount] == inf ? -1 : dp[amount];
}

int main() {
    cout << "binary index: " << binarySearch({1, 3, 5, 7, 9}, 7) << '\n';
    vector<vector<int>> graph{{1, 2}, {0, 3}, {0, 3}, {1, 2}};
    cout << "bfs:";
    for (int node : bfs(graph, 0)) cout << ' ' << node;
    cout << "\ncoin change: " << coinChange({1, 2, 5}, 11) << '\n';
    return 0;
}
```

### 6.2 Java 17

```java
import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Queue;

public class Main {
    static int binarySearch(int[] nums, int target) {
        int left = 0, right = nums.length - 1;
        while (left <= right) {
            int mid = left + (right - left) / 2;
            if (nums[mid] == target) return mid;
            if (nums[mid] < target) left = mid + 1;
            else right = mid - 1;
        }
        return -1;
    }

    static List<Integer> bfs(List<List<Integer>> graph, int start) {
        List<Integer> order = new ArrayList<>();
        boolean[] visited = new boolean[graph.size()];
        Queue<Integer> queue = new ArrayDeque<>();
        queue.offer(start);
        visited[start] = true;
        while (!queue.isEmpty()) {
            int node = queue.poll();
            order.add(node);
            for (int next : graph.get(node)) {
                if (!visited[next]) {
                    visited[next] = true;
                    queue.offer(next);
                }
            }
        }
        return order;
    }

    static int coinChange(int[] coins, int amount) {
        int[] dp = new int[amount + 1];
        Arrays.fill(dp, amount + 1);
        dp[0] = 0;
        for (int value = 1; value <= amount; value++) {
            for (int coin : coins) {
                if (coin <= value) dp[value] = Math.min(dp[value], dp[value - coin] + 1);
            }
        }
        return dp[amount] > amount ? -1 : dp[amount];
    }

    public static void main(String[] args) {
        System.out.println("binary index: " + binarySearch(new int[]{1, 3, 5, 7, 9}, 7));
        List<List<Integer>> graph = List.of(
                List.of(1, 2), List.of(0, 3), List.of(0, 3), List.of(1, 2));
        System.out.println("bfs: " + bfs(graph, 0));
        System.out.println("coin change: " + coinChange(new int[]{1, 2, 5}, 11));
    }
}
```

### 6.3 Go

```go
package main

import "fmt"

func binarySearch(nums []int, target int) int {
	left, right := 0, len(nums)-1
	for left <= right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			return mid
		}
		if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return -1
}

func bfs(graph [][]int, start int) []int {
	visited := make([]bool, len(graph))
	queue := []int{start}
	visited[start] = true
	order := make([]int, 0, len(graph))
	for head := 0; head < len(queue); head++ {
		node := queue[head]
		order = append(order, node)
		for _, next := range graph[node] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	return order
}

func coinChange(coins []int, amount int) int {
	dp := make([]int, amount+1)
	for i := 1; i <= amount; i++ {
		dp[i] = amount + 1
		for _, coin := range coins {
			if coin <= i && dp[i-coin]+1 < dp[i] {
				dp[i] = dp[i-coin] + 1
			}
		}
	}
	if dp[amount] > amount {
		return -1
	}
	return dp[amount]
}

func main() {
	fmt.Println("binary index:", binarySearch([]int{1, 3, 5, 7, 9}, 7))
	graph := [][]int{{1, 2}, {0, 3}, {0, 3}, {1, 2}}
	fmt.Println("bfs:", bfs(graph, 0))
	fmt.Println("coin change:", coinChange([]int{1, 2, 5}, 11))
}
```

### 6.4 示例输出

```text
binary index: 3
bfs: [0 1 2 3]
coin change: 3
```

不同语言的列表打印格式略有差异，但结果一致。

## 7. 理解验收清单

### 7.1 线性结构与哈希

- 能从内存布局解释数组随机访问快、链表定位慢但已知节点后插入方便。
- 能解释循环队列必须用空槽或 `size` 区分“头尾下标相同”时的空与满。
- 能说明哈希冲突无法彻底消除，以及扩容为何需要重新分配并重算桶位置。

### 7.2 树

- 能用 BST 的左右子树不变量证明中序有序，并说明有序插入为何会退化成链。
- 能根据失衡节点和新键方向识别 AVL 的 LL、RR、LR、RL，并选择正确旋转。
- 能从红黑树黑高一致与红节点不能相邻推导高度上界。
- 能说明堆只保证父子偏序，因此不能像 BST 一样沿唯一方向查任意值。

### 7.3 图

- 能根据图的稠密程度和查边需求选择邻接矩阵或邻接表。
- 能说明 BFS 在入队时标记可避免同一节点被多个前驱重复入队。
- 能根据拓扑输出数量判断有向环，并理解拓扑序可能不唯一。
- 能比较 Prim 的“扩展一个连通块”和 Kruskal 的“合并多个连通块”。
- 能用一个反例说明负权边会破坏 Dijkstra 的贪心确定性。

### 7.4 高级算法

- 能明确滑动窗口在每次扩张或收缩后必须保持的条件。
- 能把回溯代码对应到路径、候选选择、终止条件、递归与撤销操作。
- 能区分分治的相对独立子问题和动态规划的重叠子问题。
- 能从状态含义推导转移、初始化和遍历顺序，而不是先背公式。

## 8. 四周复习建议

| 周 | 重点 | 产出 |
|---:|---|---|
| 1 | 线性结构、哈希、排序、二分 | 手写 6 个基础模板 |
| 2 | 二叉树、BST、堆、Trie、并查集 | 完成遍历、TopK、连通性题 |
| 3 | 图遍历、拓扑、MST、最短路 | 对每种算法写适用前提 |
| 4 | 滑窗、回溯、分治、DP | 每类至少完成 3 道代表题 |

## 9. 最终验收

- [ ] 三种语言示例至少全部运行过一次；
- [ ] 主语言能不看答案写出二分、遍历、快排/归并、BFS/DFS；
- [ ] 能解释 AVL、红黑树、堆的有序性差异；
- [ ] 能根据图的权重和规模选择最短路算法；
- [ ] 能从暴力递归推导一个 DP；
- [ ] 能给每个算法说出一个不适用场景；
- [ ] 能在 30 分钟内完成一次综合冒烟实现。
