# 复杂度分析与算法基础

## 1. 学习目标

学完后应能：

- 用输入规模描述运行时间与额外空间的增长趋势；
- 区分最好、平均、最坏和均摊复杂度；
- 分析循环、嵌套循环、递归与常见分治递推；
- 理解为什么复杂度相同的算法，实际性能仍可能不同；
- 根据数据规模和操作比例选择结构。

## 2. 为什么不用“运行了多少毫秒”定义算法好坏

毫秒数受机器、编译器、语言运行时、缓存命中率和测试数据影响。同一个算法换一台机器可能快两倍，但它随输入规模增长的趋势不会因此改变。复杂度分析关心的是：**当输入规模变大时，算法所需的基本操作数和额外空间怎样增长。**

分析前先明确两个量：

- **输入规模 `n`**：需要根据问题定义。数组问题通常是元素数量，字符串问题通常是字符数，图问题常同时使用顶点数 `V` 和边数 `E`，背包问题还会使用容量 `C`。
- **基本操作**：比较、赋值、数组访问、哈希、指针移动等。在渐进分析中，通常把一次基本操作看作常数成本。

例如，在长度为 `n` 的无序数组中查找一个不存在的值，必须检查全部 `n` 个元素；而在有序数组中做二分查找，每次比较都能排除约一半候选区间。

假设数组长度从 `n` 变成 `2n`：

- `O(1)`：操作量近似不变；
- `O(log n)`：只增加常数级步骤；
- `O(n)`：约变为 2 倍；
- `O(n log n)`：略多于 2 倍；
- `O(n²)`：约变为 4 倍；
- `O(2^n)`：通常很快失控。

用具体数量看会更直观。下表只比较数量级，不代表真实运行时间：

| `n` | `log₂n` | `n` | `n log₂n` | `n²` | `2^n` |
|---:|---:|---:|---:|---:|---:|
| 10 | 约 3.3 | 10 | 约 33 | 100 | 1024 |
| 1000 | 约 10 | 1000 | 约 10000 | 100 万 | 已不可直接枚举 |
| 100 万 | 约 20 | 100 万 | 约 2000 万 | `10^12` | 无法枚举 |

因此复杂度的价值不是给出精确秒数，而是提前判断：数据规模扩大后，这个方法是否仍然有可能运行完。

## 3. O、Ω、Θ

| 记号 | 含义 | 常用理解 |
|---|---|---|
| `O(g(n))` | 渐进上界 | 足够大的 `n` 下，增长速度不会超过 `g(n)` 的常数倍 |
| `Ω(g(n))` | 渐进下界 | 足够大的 `n` 下，增长速度至少达到 `g(n)` 的常数倍 |
| `Θ(g(n))` | 紧确界 | 上下界同阶，增长速度可以准确归入该量级 |

工程和面试通常用大 O 描述最坏复杂度，但不要把“大 O”机械等同于“最坏情况”。大 O 是上界记号；只是实践中经常用它表达最坏上界。

例如 `3n² + 10n + 7` 同时属于 `O(n³)` 和 `O(n²)`，因为两者都是上界；但更精确的写法是 `Θ(n²)`。实际交流中说“复杂度是 `O(n²)`”，通常是在选择尽可能紧的常用上界。

渐进分析忽略两类信息：

1. **常数因子**：`100n` 和 `n` 都是 `O(n)`；
2. **低阶项**：`n² + 1000n` 最终由 `n²` 主导。

忽略它们的前提是讨论足够大的 `n`。如果数据永远只有几十个元素，常数、实现复杂度和可读性可能比渐进阶更重要。

## 4. 常见增长阶

| 复杂度 | 典型操作 | 增长原因 |
|---|---|---|
| `O(1)` | 数组按下标访问、栈顶操作 | 操作次数不随 `n` 增长 |
| `O(log n)` | 二分查找、平衡树查询 | 每一步按固定比例缩小问题 |
| `O(n)` | 遍历数组、链表查找 | 每个元素最多处理常数次 |
| `O(n log n)` | 归并、堆排、平均快排 | `n` 次 `O(log n)` 操作，或 `log n` 层且每层总工作为 `O(n)` |
| `O(n²)` | 双重枚举、基础排序 | 大量元素对之间发生比较 |
| `O(2^n)` | 遍历所有子集状态（不含结果写出） | 每个元素都有选/不选两种分支 |
| `O(n!)` | 遍历所有排列状态（不含结果写出） | 第一位 `n` 种、第二位 `n-1` 种，依次相乘 |

若要复制并保存每个最长为 `n` 的结果，写出成本还可能分别达到 `Θ(n·2^n)` 和 `Θ(n·n!)`。

分析时忽略常数与低阶项：

```text
3n² + 10n + 7 = Θ(n²)
```

这不代表常数永远不重要。数据规模较小时，缓存局部性、分配次数和语言运行时都可能决定实际速度。

## 5. 循环分析

### 5.1 单循环

循环执行 `n` 次，每次只做常数次操作，总计 `O(n)`。关键不是代码写了几行，而是这些行随 `n` 重复多少次。

```text
for i = 0 .. n-1:
    读取 a[i]
    与 target 比较
```

如果目标在第一个位置，最好情况是 `O(1)`；如果目标不存在，最坏情况是 `O(n)`。当题目未说明输入分布时，通常报告最坏复杂度。

### 5.2 顺序执行的代码相加

如果先完整遍历一次数组，再完整遍历一次，总代价是：

```text
O(n) + O(n) = O(2n) = O(n)
```

顺序执行的步骤相加，最后保留增长最快的项。不要看到两个循环就直接写成 `O(n²)`；只有循环发生嵌套或工作量相乘时才可能平方。

### 5.3 矩形嵌套循环

外层执行 `n` 次，每次都让内层完整执行 `n` 次，总操作数为 `n × n = n²`。

```text
for i = 0 .. n-1:
    for j = 0 .. n-1:
        处理 (i, j)
```

如果两层规模不同，例如遍历矩阵的 `rows × cols` 个元素，应写成 `O(rows × cols)`，不要强行都记成 `n`。

### 5.4 三角形循环与求和

若内层只运行到当前外层下标，次数不是简单的 `n × n`，而是：

```text
1 + 2 + ... + (n - 1) = n(n - 1) / 2 = Θ(n²)
```

虽然只遍历了“半个矩形”，常数 `1/2` 会被忽略，数量级仍是 `Θ(n²)`。排序中的插入移动、枚举所有无序元素对都常出现这种求和。

### 5.5 每次按比例缩小

```text
n → n/2 → n/4 → ... → 1
```

设执行 `k` 步后规模降到 1，则 `n / 2^k = 1`，所以 `k = log₂n`。因此二分查找、不断除以 2 的循环都是 `O(log n)`。

对数底数在大 O 中通常省略，因为：

```text
log₂n = log₁₀n / log₁₀2
```

不同底数只相差常数倍。

### 5.6 外层线性、内层对数

如果对每个元素都执行一次二分查找，总代价是：

```text
n × log n = O(n log n)
```

这类结构也出现在“`log n` 层递归，每层处理全部 `n` 个元素”的归并排序中。

### 5.7 分支与提前退出

`if/else` 只会执行其中一个分支，因此通常取两个分支中更大的复杂度，而不是相加。提前退出会改善最好情况，但最坏情况仍需考虑退出条件始终不成立的输入。

## 6. 递归与递推

递归复杂度要同时考虑：

1. 产生多少个子问题；
2. 每层除递归外做多少工作；
3. 递归深度带来的栈空间。

分析递归时，不要只数函数内部的几行代码。一次调用可能继续产生多个调用，总代价由整棵递归树决定。

常见递推：

| 递推 | 结果 | 例子 |
|---|---|---|
| `T(n)=T(n-1)+O(1)` | `O(n)` | 链表递归遍历 |
| `T(n)=T(n/2)+O(1)` | `O(log n)` | 二分查找 |
| `T(n)=2T(n/2)+O(n)` | `O(n log n)` | 归并排序 |
| `T(n)=2T(n-1)+O(1)` | `O(2^n)` | 未剪枝的二叉决策树 |

### 6.1 二分查找：单分支递归

二分每次只进入一个规模为一半的子问题：

```text
T(n) = T(n/2) + O(1)
```

递归链为 `n → n/2 → n/4 → ... → 1`，深度 `O(log n)`；每层工作 `O(1)`，总时间 `O(log n)`，递归栈空间也是 `O(log n)`。若改成循环，时间不变，额外空间可降为 `O(1)`。

### 6.2 归并排序：双分支但每层总量不变

归并把数组拆成两个一半大小的子问题，并在线性时间内合并：

```text
T(n) = 2T(n/2) + O(n)
```

以 `n=8` 为例：

| 层 | 子问题数量 | 每个规模 | 本层合并总工作 |
|---:|---:|---:|---:|
| 0 | 1 | 8 | 8 |
| 1 | 2 | 4 | 8 |
| 2 | 4 | 2 | 8 |
| 3 | 8 | 1 | 到达基本情况 |

共有 `log₂n` 个需要合并的层级，每层总工作 `O(n)`，所以总时间 `O(n log n)`。

### 6.3 回溯：分支数决定指数增长

如果每个位置都有“选或不选”两个分支，深度为 `n`，叶子最多 `2^n` 个。剪枝可以减少实际访问节点，但除非能证明新的上界，最坏复杂度仍通常按指数级描述。

### 6.4 递归空间不是所有节点之和

调用栈只保存当前尚未返回的一条路径。当每个栈帧只占 `O(1)` 额外空间时，栈空间由**最大递归深度**决定；一般情况则要累加活动路径上各栈帧的占用。归并排序有 `O(n)` 辅助数组和 `O(log n)` 调用栈，主导额外空间是 `O(n)`；遍历退化链式树时，递归深度可能达到 `O(n)`。

## 7. 空间复杂度

这里采用工程和面试中常见的**辅助空间**口径：计算算法执行期间输入之外额外占用的空间。若题目问总空间，还要加上输入本身的存储。分析时要分别检查：

- 原地交换几个变量：`O(1)`；
- 新建长度为 `n` 的辅助数组：`O(n)`；
- 平衡树递归遍历：调用栈 `O(log n)`；
- 极端退化树递归：调用栈 `O(n)`。

还要注意以下细节：

- **返回结果是否计入**：通常单独说明。生成 `n` 个答案至少需要 `O(n)` 输出空间，即使算法内部只用 `O(1)`。
- **切片或子串是否复制**：不同语言实现不同。某些操作只是共享底层数据，某些会重新分配并复制。
- **原地不等于零空间**：原地通常强调不另建随 `n` 线性增长的数据副本；严格口径下常要求 `O(1)` 辅助存储。若实现依赖递归，还要单独报告 `O(log n)` 或 `O(n)` 调用栈。
- **容器预留容量**：容量为 `n` 的哈希表、队列或访问数组都是 `O(n)` 空间。

时间与空间经常可以互换。例如哈希表用 `O(n)` 空间把大量查找从 `O(n)` 降到平均 `O(1)`；记忆化搜索保存状态，以空间换取避免重复递归的时间。

## 8. 最好、平均、最坏与均摊

### 8.1 最好与最坏情况

顺序查找时，目标恰好在第一个位置是最好 `O(1)`，目标不存在是最坏 `O(n)`。快速排序在划分均匀时为 `O(n log n)`，若每次都产生大小为 `0` 和 `n-1` 的子区间，最坏会退化为 `O(n²)`。

复杂度必须和输入条件一起陈述，不能只背一个数字。

### 8.2 平均复杂度

平均复杂度需要假设输入的概率分布。例如哈希表平均 `O(1)` 依赖哈希分布较均匀、负载因子受控；如果大量键落入同一个桶，链地址法的单次查找会退化为 `O(n)`。

“平均”不是随便取最好和最坏的平均值，而是对所有可能输入按概率加权。

### 8.3 均摊复杂度

均摊分析关注一串连续操作的总成本，不需要假设随机输入。以容量按两倍增长的动态数组尾部追加为例：

- 大多数追加直接写入：`O(1)`；
- 容量不足时扩容并复制：单次 `O(n)`；
- 连续追加很多次，总复制成本呈几何级数，因此**均摊 `O(1)`**。

为简化求和，假设初始容量为 1，连续追加 `m` 次，且 `m` 是 2 的幂；此时最终容量恰为 `m`，历史扩容复制量不超过：

```text
1 + 2 + 4 + ... + m/2 < m
```

再加上 `m` 次正常写入，总工作仍是 `O(m)`，均摊到每次追加就是 `O(1)`。当 `m` 不是 2 的幂时，复制总量仍受追加次数的常数倍约束，结论不变。

### 8.4 复杂度相同，实际速度仍可能不同

数组遍历和链表遍历都是 `O(n)`，但数组通常更快，因为元素连续，CPU 缓存命中率高；链表节点分散，还需要指针跳转。两个排序都为 `O(n log n)`，比较次数、内存分配和分支预测也可能不同。

正确做法是分两层判断：

1. 用复杂度排除无法承受的数据增长；
2. 在复杂度可接受的候选中，用真实数据做基准测试。

## 9. 三语言完整示例：线性查找与二分查找的操作次数

下面程序在相同有序数组中查找目标，输出结果和比较次数。它不是严谨基准测试，而是帮助观察 `O(n)` 与 `O(log n)` 的增长差异。

### 9.1 C++17

```cpp
#include <iostream>
#include <utility>
#include <vector>

using namespace std;

pair<int, int> linearSearch(const vector<int>& nums, int target) {
    int comparisons = 0;
    for (int i = 0; i < static_cast<int>(nums.size()); ++i) {
        ++comparisons;
        if (nums[i] == target) {
            return {i, comparisons};
        }
    }
    return {-1, comparisons};
}

pair<int, int> binarySearch(const vector<int>& nums, int target) {
    int left = 0;
    int right = static_cast<int>(nums.size()) - 1;
    int comparisons = 0;
    while (left <= right) {
        int mid = left + (right - left) / 2;
        ++comparisons;
        if (nums[mid] == target) {
            return {mid, comparisons};
        }
        if (nums[mid] < target) {
            left = mid + 1;
        } else {
            right = mid - 1;
        }
    }
    return {-1, comparisons};
}

int main() {
    vector<int> nums;
    for (int i = 1; i <= 63; ++i) {
        nums.push_back(i * 2);
    }

    int target = 126;
    auto linear = linearSearch(nums, target);
    auto binary = binarySearch(nums, target);
    cout << "linear: index=" << linear.first
         << ", comparisons=" << linear.second << '\n';
    cout << "binary: index=" << binary.first
         << ", comparisons=" << binary.second << '\n';
    return 0;
}
```

### 9.2 Java 17

```java
import java.util.ArrayList;
import java.util.List;

public class Main {
    static int[] linearSearch(List<Integer> nums, int target) {
        int comparisons = 0;
        for (int i = 0; i < nums.size(); i++) {
            comparisons++;
            if (nums.get(i) == target) {
                return new int[]{i, comparisons};
            }
        }
        return new int[]{-1, comparisons};
    }

    static int[] binarySearch(List<Integer> nums, int target) {
        int left = 0;
        int right = nums.size() - 1;
        int comparisons = 0;
        while (left <= right) {
            int mid = left + (right - left) / 2;
            comparisons++;
            if (nums.get(mid) == target) {
                return new int[]{mid, comparisons};
            }
            if (nums.get(mid) < target) {
                left = mid + 1;
            } else {
                right = mid - 1;
            }
        }
        return new int[]{-1, comparisons};
    }

    public static void main(String[] args) {
        List<Integer> nums = new ArrayList<>();
        for (int i = 1; i <= 63; i++) {
            nums.add(i * 2);
        }

        int target = 126;
        int[] linear = linearSearch(nums, target);
        int[] binary = binarySearch(nums, target);
        System.out.printf("linear: index=%d, comparisons=%d%n", linear[0], linear[1]);
        System.out.printf("binary: index=%d, comparisons=%d%n", binary[0], binary[1]);
    }
}
```

### 9.3 Go

```go
package main

import "fmt"

func linearSearch(nums []int, target int) (int, int) {
	comparisons := 0
	for i, value := range nums {
		comparisons++
		if value == target {
			return i, comparisons
		}
	}
	return -1, comparisons
}

func binarySearch(nums []int, target int) (int, int) {
	left, right := 0, len(nums)-1
	comparisons := 0
	for left <= right {
		mid := left + (right-left)/2
		comparisons++
		if nums[mid] == target {
			return mid, comparisons
		}
		if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return -1, comparisons
}

func main() {
	nums := make([]int, 63)
	for i := range nums {
		nums[i] = (i + 1) * 2
	}

	target := 126
	linearIndex, linearCount := linearSearch(nums, target)
	binaryIndex, binaryCount := binarySearch(nums, target)
	fmt.Printf("linear: index=%d, comparisons=%d\n", linearIndex, linearCount)
	fmt.Printf("binary: index=%d, comparisons=%d\n", binaryIndex, binaryCount)
}
```

### 9.4 示例输出

```text
linear: index=62, comparisons=63
binary: index=62, comparisons=6
```

## 10. 数据结构选型与复杂度

| 需求 | 常见选择 | 关键复杂度 |
|---|---|---|
| 按下标访问 | 数组 | `O(1)` |
| 频繁头尾插入 | 链表 / 双端队列 | 链表维护首尾指针时 `O(1)`；双端队列通常均摊 `O(1)` |
| 按键快速查询 | 哈希表 | 平均 `O(1)` |
| 动态保持有序 | 平衡搜索树 | `O(log n)` |
| 反复取最值 | 堆 | 取顶 `O(1)`，插入/删除顶 `O(log n)` |
| 前缀定位 / 判断 | Trie | `O(L)`，`L` 为前缀长度；返回全部匹配还需加遍历与输出成本 |
| 连通性合并查询 | 并查集 | 路径压缩并按秩或大小合并时，均摊 `O(α(n))`，工程上近似常数 |

## 11. 常见易错点

- 只看一层循环就判断复杂度，忽略内层次数依赖；
- 把连续的两个 `O(n)` 错写为 `O(n²)`，实际相加仍是 `O(n)`；
- 漏算递归调用栈；
- 把哈希表平均 `O(1)` 说成无条件最坏 `O(1)`；
- 只报时间复杂度，不说明输入前提和额外空间；
- 用复杂度替代基准测试：复杂度决定增长趋势，实测决定具体工程表现。

## 12. 练习

1. 分析三角形双循环的精确执行次数并化简。
2. 分析递归二分的时间与栈空间。
3. 解释动态数组追加为什么是均摊 `O(1)`。
4. 比较“排序一次后查找 `m` 次”和“每次线性查找”的总代价。
5. 给出一个 `O(n log n)` 算法实际慢于 `O(n²)` 算法的小规模场景，并说明原因。
