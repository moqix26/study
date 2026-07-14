# 查找算法：顺序、二分、插值与斐波那契查找

## 学习目标

- 掌握四种查找算法的前提、边界处理和返回约定。
- 理解二分查找的区间收缩、插值查找的位置估计和斐波那契分割。
- 能根据数据是否有序、分布特点和访问成本选择算法。

## 查找问题与返回约定

查找算法接收一组元素和目标值，回答目标是否存在；若存在，通常返回一个下标，否则返回特殊值。本文三种语言示例统一返回下标，未找到返回 `-1`。

开始实现前必须明确两个语义：

1. 数组中可能有重复值时，是返回任意一个匹配、第一次出现、最后一次出现，还是全部位置？
2. 数据是否已经有序，是否支持 O(1) 随机访问？

本文顺序查找从左向右，因此自然返回第一个匹配；二分、插值和斐波那契查找只保证返回某个匹配位置。若要求重复值的左右边界，需要专门的边界查找算法，不能依赖“碰巧找到哪一个”。

还要把预处理成本算入整体方案：对一批无序数据只查一次，先花 O(n log n) 排序再做 O(log n) 查找通常不如直接 O(n) 扫描；若同一数据要查询很多次，排序、建哈希表或建搜索树才可能摊薄成本。

## 顺序查找

### 问题模型与不变量

顺序查找不利用元素之间的大小关系，只从下标 0 开始逐个比较。它适用于任何可遍历序列，包括无序数组和链表。

扫描到下标 `i` 之前的不变量是：区间 `[0, i)` 已经检查完毕，里面没有目标；下标 `[i, n)` 尚未检查。若 `a[i] == target`，当前下标就是从左侧遇到的第一个匹配。

### 逐步过程与推演

在数组 `[3, 8, 13, 21, 34, 55, 89]` 中查找 34：

| 检查下标 | 元素值 | 结果 |
| ---: | ---: | --- |
| 0 | 3 | 不相等，继续 |
| 1 | 8 | 不相等，继续 |
| 2 | 13 | 不相等，继续 |
| 3 | 21 | 不相等，继续 |
| 4 | 34 | 命中，返回 4 |

若目标是 100，算法会检查全部 7 个元素后返回 -1。

### 为什么正确

算法按顺序检查每个可能位置。找到时，该位置确实保存目标；若扫描结束仍未找到，则所有元素都已被比较并确认不等于目标，因此目标不存在。因为从左向右在首次命中时立即返回，重复值情况下返回的是第一个位置。

### 复杂度与边界

- 目标恰好在首位时只比较一次，最好 O(1)。
- 目标在末位或不存在时比较 n 次，最坏 O(n)。
- 若目标位置近似均匀分布，成功查询平均检查约 `(n+1)/2` 个元素，仍为 O(n)。
- 只保存循环下标，额外空间 O(1)。

顺序查找对链表尤其自然，因为它不要求随机访问。数据量很小、只查一次或无序且无法建立索引时，它往往是最直接可靠的选择。

## 二分查找

### 前提与区间不变量

二分查找要求数组按与比较规则一致的顺序排列，并且能 O(1) 访问中间下标。本文采用闭区间 `[left, right]`。

每轮开始时的不变量是：如果目标存在，那么它一定在当前闭区间 `[left, right]` 中；区间之外的位置已被安全排除。循环条件必须是 `left <= right`，因为 `left == right` 时仍有一个候选元素需要检查。

### 逐步过程

1. 计算 `middle = left + (right-left)/2`。
2. 若 `a[middle] == target`，返回中点。
3. 若 `a[middle] < target`，利用有序性可知 `[left, middle]` 都不可能是目标，令 `left = middle + 1`。
4. 若 `a[middle] > target`，排除 `[middle, right]`，令 `right = middle - 1`。
5. 当 `left > right` 时，候选区间为空，返回不存在。

更新边界时必须越过已经比较过的 `middle`。若写成 `left = middle` 或 `right = middle`，在只剩两个元素时可能不再缩小区间而死循环。

### 手工推演

仍在 `[3, 8, 13, 21, 34, 55, 89]` 中查找 34：

| `left` | `right` | `middle` | `a[middle]` | 决策 |
| ---: | ---: | ---: | ---: | --- |
| 0 | 6 | 3 | 21 | 21 < 34，令 `left = 4` |
| 4 | 6 | 5 | 55 | 55 > 34，令 `right = 4` |
| 4 | 4 | 4 | 34 | 命中，返回 4 |

若查找 35，第三轮会检查下标 4 的 34 并令 `left = 5`，此时 `left > right`，候选区间为空。

### 为什么正确

数组有序。当中点值小于目标时，中点及其左侧所有值都不大于中点，因此不可能等于更大的目标，可以整体排除；中点值大于目标时同理排除右侧。每轮既不排除可能的目标，又让区间严格缩小。最终命中则返回正确位置；区间为空则所有位置都已被排除，目标不存在。

### 复杂度来源

每轮候选区间最多约减半。经过 t 轮后规模不超过 `n / 2^t`；当它小于 1 时结束，所以 `t` 约为 `log2 n`，时间 O(log n)。迭代实现只用三个下标，空间 O(1)；递归实现还需 O(log n) 调用栈。

### 重复值与边界版本

普通二分遇到相等值便返回，因此重复数组中结果不固定。寻找第一个不小于目标的位置时，即使命中也要继续向左收缩；寻找第一个大于目标的位置时，要根据 `<=` 或 `>` 更新边界。这两个边界可进一步组合出等值区间和目标出现次数。

二分还可用于“答案空间”：只要某个判定随候选答案单调变化，就可以查找第一个可行或最后一个不可行值，不要求真的有一个数组。

## 插值查找

### 位置估计的直觉

二分查找无论目标值靠近哪端都固定检查中点。插值查找假设数值在区间内近似均匀分布，根据目标在首尾值之间所占的比例估计下标：

`position = low + (target - a[low]) * (high - low) / (a[high] - a[low])`

如果一本按编号均匀排列的目录从 1 到 1000，而目标编号接近 900，人会先翻到靠后位置而不是正中间；插值查找就是这种“按值猜位置”的做法。

### 区间不变量与步骤

数组必须升序且键为可做差值运算的数值。每轮保持：若目标存在，它位于 `[low, high]`，并且目标值处于 `[a[low], a[high]]`。

1. 先检查 `low <= high`，并确认目标没有落在当前首尾值范围之外。
2. 若 `a[low] == a[high]`，当前区间所有值相等；直接判断该值是否为目标，避免除零。
3. 在更宽整数类型中计算估计位置，必要时校验位置仍在区间内。
4. 命中则返回；估计值小于目标就令 `low = position + 1`，大于目标则令 `high = position - 1`。

### 手工推演

在 `[3, 8, 13, 21, 34, 55, 89]` 中查找 34。这个数组增长并不均匀，所以不会一次命中：

| `low..high` | 位置估计 | 探测值 | 决策 |
| --- | --- | ---: | --- |
| `0..6` | `0 + (34-3)*6/(89-3) = 2` | 13 | 13 < 34，`low = 3` |
| `3..6` | `3 + (34-21)*3/(89-21) = 3` | 21 | 21 < 34，`low = 4` |
| `4..6` | `4 + (34-34)*2/(89-34) = 4` | 34 | 命中，返回 4 |

若数组为 `[10,20,30,40,50,60,70,80,90]`，查找 70 时第一次估计就会得到下标 6，这体现了均匀分布下的优势。

### 为什么正确

位置公式只决定“本轮检查哪里”，正确性仍来自有序区间的排除规则。若探测值小于目标，探测位置及左侧都不可能等于目标；大于目标则排除右侧。边界每轮越过已检查位置并严格收缩。目标落在首尾值范围之外时，有序性直接证明当前区间不可能含目标。

### 复杂度、退化与数值风险

均匀独立分布的理想模型下，估计位置非常接近目标，平均可达到 O(log log n)。这不是普遍保证：在指数增长、强烈偏斜或有巨大离群值的数据中，估计可能连续贴近同一端，每轮只排除少量元素，最坏 O(n)。因此普通工程数组查找通常更偏爱行为稳定的二分查找。

公式中的 `(target-a[low]) * (high-low)` 可能在乘法阶段溢出，即使最终商能放入普通整数。应先提升到更宽类型或使用安全运算；浮点估计还要处理舍入和越界。本文整数版本也必须先处理分母为零。

## 斐波那契查找

### 分割思想

斐波那契查找与二分查找同样利用有序性，但不按 1:1 分割，而是借助斐波那契数列 `0, 1, 1, 2, 3, 5, 8, ...` 把候选区间分成接近黄金比例的两部分。

先找到最小的斐波那契数 `F(k) >= n`。维护三个相邻数 `F(k)`、`F(k-1)`、`F(k-2)` 和一个 `offset`：`offset` 表示已经确认小于目标、被排除的最右下标。下一探测位置近似为：

`index = min(offset + F(k-2), n - 1)`

使用 `min` 是因为 `F(k)` 可能大于真实数组长度，理论分割位置可能落到末尾之外。

### 候选区间不变量与更新

每轮开始时，目标若存在，一定位于 `offset` 右侧尚未排除的候选区间，该区间长度由当前 `F(k)` 覆盖。

- 若 `a[index] < target`，排除从旧 `offset+1` 到 `index` 的左块，令 `offset = index`；剩余右块规模对应 `F(k-1)`，三个斐波那契数向下移动一阶。
- 若 `a[index] > target`，保留左块，规模对应 `F(k-2)`，三个数向下移动两阶；`offset` 不变。
- 若相等，直接返回。

当主要循环结束时，可能还剩一个未检查候选位置 `offset+1`，需要做最后一次边界与值检查。

### 手工推演

在 `[3, 8, 13, 21, 34, 55, 89]` 中查找 34，`n = 7`。最小的不小于 7 的斐波那契数是 8，此时相邻状态为 `F(k)=8`、`F(k-1)=5`、`F(k-2)=3`，初始 `offset=-1`：

| 当前状态 | 探测下标 | 探测值 | 更新 |
| --- | ---: | ---: | --- |
| `8,5,3`，`offset=-1` | `min(-1+3,6)=2` | 13 | 13 < 34，排除下标 0～2，`offset=2`，状态变 `5,3,2` |
| `5,3,2`，`offset=2` | `min(2+2,6)=4` | 34 | 命中，返回 4 |

与二分相比，它第一次选择下标 2 而非 3，体现了不同的分割比例。

### 为什么正确

每次探测后都依赖有序性排除不可能的一侧。斐波那契恒等式 `F(k)=F(k-1)+F(k-2)` 保证探测点两侧的候选规模可继续由更小的相邻斐波那契数描述；更新状态与保留的一侧相对应。候选规模严格下降，最终要么命中，要么缩小到至多一个待检查位置。最后检查失败即可证明目标不存在。

### 复杂度与现实边界

斐波那契数约按黄金比例 `phi` 指数增长，所以覆盖长度 n 只需 O(log n) 阶；每轮降一到两阶，查询时间 O(log n)，额外空间 O(1)。生成斐波那契数也需 O(log n) 次加法。

它历史上的吸引力之一是分割位置主要使用加减法；在现代处理器上，整数除法通常不是二分查找整体性能的决定因素，而且二分实现更简单，因此斐波那契查找较少作为默认选择。生成数列时还要防止斐波那契数溢出；数组长度不是斐波那契数时必须限制探测下标，并正确处理最后一个剩余位置。

## 复杂度

| 算法 | 前提 | 最好时间 | 平均时间 | 最坏时间 | 额外空间 |
| --- | --- | --- | --- | --- | --- |
| 顺序查找 | 无 | O(1) | O(n) | O(n) | O(1) |
| 二分查找 | 有序、可随机访问 | O(1) | O(log n) | O(log n) | O(1) |
| 插值查找 | 有序数值且分布较均匀 | O(1) | 理想 O(log log n) | O(n) | O(1) |
| 斐波那契查找 | 有序、可随机访问 | O(1) | O(log n) | O(log n) | O(1) |

### 复杂度比较应如何理解

- 顺序查找的 O(n) 不要求任何预处理，也适用于链表。
- 二分和斐波那契查找的 O(log n) 建立在“数据已排序且可 O(1) 访问任意下标”之上；用于链表时，定位中间位置本身不是 O(1)，优势会消失。
- 插值查找的 O(log log n) 是分布良好时的期望，不是最坏保证；其最坏 O(n) 比二分差。
- 四种迭代实现都只保存常数个下标或数值，额外空间 O(1)。

## 选型与不适用场景

| 数据条件 | 合适方法 | 原因 |
| --- | --- | --- |
| 无序、小规模、只查一次 | 顺序查找 | 无预处理成本 |
| 有序数组、需要稳定性能 | 二分查找 | 最坏 O(log n)，边界变种成熟 |
| 有序数值且近似均匀分布 | 可考虑插值查找 | 估计位置可能更接近目标 |
| 研究非等分策略或特定访问代价模型 | 斐波那契查找 | 使用斐波那契比例分割 |
| 高频无序按键查询 | 通常改用哈希索引 | 平均查找 O(1)，不必每次扫描 |
| 需要范围查询、动态有序集合 | 搜索树或排序数组边界查找 | 能利用顺序关系 |

若数据会频繁插入删除，维护有序数组本身可能需要 O(n) 搬移；不要只比较一次查询的成本，而忽略整个数据生命周期。插值与斐波那契查找也不适合只能顺序读取的流式输入，因为它们依赖随机访问探测位置。

## 三种语言的实现边界

- **C++**：标准库 `lower_bound`、`upper_bound` 和 `binary_search` 使用半开区间，要求范围按同一比较器有序；手写代码混用 `size_t` 与 `int` 时要防止空数组下标和无符号下溢。
- **Java**：`Arrays.binarySearch` 未找到时返回负的插入点编码，不只是固定 `-1`；比较对象时应使用一致的 `Comparator`。中间值和插值乘积仍应使用 `long` 防溢出。
- **Go**：`sort.Search` 接收一个从 false 单调变为 true 的判定函数，天然适合寻找边界；较新版本的 `slices.BinarySearch` 可直接搜索有序切片。手写插值公式时要注意 `int` 在 32 位与 64 位平台宽度不同。

## C++17 完整示例

```cpp
#include <algorithm>
#include <iostream>
#include <vector>

int sequentialSearch(const std::vector<int>& values, int target) {
    for (std::size_t i = 0; i < values.size(); ++i) {
        if (values[i] == target) return static_cast<int>(i);
    }
    return -1;
}

int binarySearch(const std::vector<int>& values, int target) {
    int left = 0, right = static_cast<int>(values.size()) - 1;
    while (left <= right) {
        int middle = left + (right - left) / 2;
        if (values[middle] == target) return middle;
        if (values[middle] < target) left = middle + 1;
        else right = middle - 1;
    }
    return -1;
}

int interpolationSearch(const std::vector<int>& values, int target) {
    if (values.empty()) return -1;
    int low = 0, high = static_cast<int>(values.size()) - 1;
    while (low <= high && target >= values[low] && target <= values[high]) {
        if (values[low] == values[high]) return values[low] == target ? low : -1;
        long long numerator = (static_cast<long long>(target) - values[low]) *
                              (high - low);
        long long denominator = static_cast<long long>(values[high]) - values[low];
        int position = low + static_cast<int>(numerator / denominator);
        if (values[position] == target) return position;
        if (values[position] < target) low = position + 1;
        else high = position - 1;
    }
    return -1;
}

int fibonacciSearch(const std::vector<int>& values, int target) {
    int n = static_cast<int>(values.size());
    long long fibMinus2 = 0;
    long long fibMinus1 = 1;
    long long fib = fibMinus1 + fibMinus2;
    while (fib < n) {
        fibMinus2 = fibMinus1;
        fibMinus1 = fib;
        fib = fibMinus1 + fibMinus2;
    }

    int offset = -1;
    while (fib > 1) {
        int index = static_cast<int>(std::min(
            static_cast<long long>(offset) + fibMinus2,
            static_cast<long long>(n) - 1));
        if (values[index] < target) {
            fib = fibMinus1;
            fibMinus1 = fibMinus2;
            fibMinus2 = fib - fibMinus1;
            offset = index;
        } else if (values[index] > target) {
            fib = fibMinus2;
            fibMinus1 = fibMinus1 - fibMinus2;
            fibMinus2 = fib - fibMinus1;
        } else {
            return index;
        }
    }
    if (fibMinus1 && offset + 1 < n && values[offset + 1] == target) return offset + 1;
    return -1;
}

int main() {
    const std::vector<int> values{3, 8, 13, 21, 34, 55, 89};
    int target = 34;
    std::cout << "顺序查找: " << sequentialSearch(values, target) << "\n";
    std::cout << "二分查找: " << binarySearch(values, target) << "\n";
    std::cout << "插值查找: " << interpolationSearch(values, target) << "\n";
    std::cout << "斐波那契查找: " << fibonacciSearch(values, target) << "\n";
    return 0;
}
```

## Java 17 完整示例

```java
public class Main {
    private static int sequentialSearch(int[] values, int target) {
        for (int i = 0; i < values.length; i++) {
            if (values[i] == target) return i;
        }
        return -1;
    }

    private static int binarySearch(int[] values, int target) {
        int left = 0, right = values.length - 1;
        while (left <= right) {
            int middle = left + (right - left) / 2;
            if (values[middle] == target) return middle;
            if (values[middle] < target) left = middle + 1;
            else right = middle - 1;
        }
        return -1;
    }

    private static int interpolationSearch(int[] values, int target) {
        if (values.length == 0) return -1;
        int low = 0, high = values.length - 1;
        while (low <= high && target >= values[low] && target <= values[high]) {
            if (values[low] == values[high]) return values[low] == target ? low : -1;
            long numerator = ((long) target - values[low]) * (high - low);
            long denominator = (long) values[high] - values[low];
            int position = low + (int) (numerator / denominator);
            if (values[position] == target) return position;
            if (values[position] < target) low = position + 1;
            else high = position - 1;
        }
        return -1;
    }

    private static int fibonacciSearch(int[] values, int target) {
        long fibMinus2 = 0;
        long fibMinus1 = 1;
        long fib = fibMinus1 + fibMinus2;
        while (fib < values.length) {
            fibMinus2 = fibMinus1;
            fibMinus1 = fib;
            fib = fibMinus1 + fibMinus2;
        }

        int offset = -1;
        while (fib > 1) {
            int index = (int) Math.min((long) offset + fibMinus2, values.length - 1L);
            if (values[index] < target) {
                fib = fibMinus1;
                fibMinus1 = fibMinus2;
                fibMinus2 = fib - fibMinus1;
                offset = index;
            } else if (values[index] > target) {
                fib = fibMinus2;
                fibMinus1 = fibMinus1 - fibMinus2;
                fibMinus2 = fib - fibMinus1;
            } else {
                return index;
            }
        }
        if (fibMinus1 == 1 && offset + 1 < values.length && values[offset + 1] == target) {
            return offset + 1;
        }
        return -1;
    }

    public static void main(String[] args) {
        int[] values = {3, 8, 13, 21, 34, 55, 89};
        int target = 34;
        System.out.println("顺序查找: " + sequentialSearch(values, target));
        System.out.println("二分查找: " + binarySearch(values, target));
        System.out.println("插值查找: " + interpolationSearch(values, target));
        System.out.println("斐波那契查找: " + fibonacciSearch(values, target));
    }
}
```

## Go 完整示例

```go
package main

import "fmt"

func sequentialSearch(values []int, target int) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}

func binarySearch(values []int, target int) int {
	left, right := 0, len(values)-1
	for left <= right {
		middle := left + (right-left)/2
		if values[middle] == target {
			return middle
		}
		if values[middle] < target {
			left = middle + 1
		} else {
			right = middle - 1
		}
	}
	return -1
}

func interpolationSearch(values []int, target int) int {
	if len(values) == 0 {
		return -1
	}
	low, high := 0, len(values)-1
	for low <= high && target >= values[low] && target <= values[high] {
		if values[low] == values[high] {
			if values[low] == target {
				return low
			}
			return -1
		}
		indexSpan := uint64(high - low)
		targetOffset := uint64(target) - uint64(values[low])
		valueRange := uint64(values[high]) - uint64(values[low])
		position := low + (high-low)/2
		if targetOffset <= ^uint64(0)/indexSpan {
			position = low + int(targetOffset*indexSpan/valueRange)
		}
		if values[position] == target {
			return position
		}
		if values[position] < target {
			low = position + 1
		} else {
			high = position - 1
		}
	}
	return -1
}

func fibonacciSearch(values []int, target int) int {
	fibMinus2, fibMinus1 := uint64(0), uint64(1)
	fib := fibMinus1 + fibMinus2
	for fib < uint64(len(values)) {
		fibMinus2 = fibMinus1
		fibMinus1 = fib
		fib = fibMinus1 + fibMinus2
	}

	offset := -1
	for fib > 1 {
		index := len(values) - 1
		if fibMinus2 <= uint64(len(values)-1-offset) {
			index = offset + int(fibMinus2)
		}
		if values[index] < target {
			fib = fibMinus1
			fibMinus1 = fibMinus2
			fibMinus2 = fib - fibMinus1
			offset = index
		} else if values[index] > target {
			fib = fibMinus2
			fibMinus1 = fibMinus1 - fibMinus2
			fibMinus2 = fib - fibMinus1
		} else {
			return index
		}
	}
	if fibMinus1 == 1 && offset+1 < len(values) && values[offset+1] == target {
		return offset + 1
	}
	return -1
}

func main() {
	values := []int{3, 8, 13, 21, 34, 55, 89}
	target := 34
	fmt.Println("顺序查找:", sequentialSearch(values, target))
	fmt.Println("二分查找:", binarySearch(values, target))
	fmt.Println("插值查找:", interpolationSearch(values, target))
	fmt.Println("斐波那契查找:", fibonacciSearch(values, target))
}
```

## 示例输出

```text
顺序查找: 4
二分查找: 4
插值查找: 4
斐波那契查找: 4
```

## 易错点

- 二分、插值和斐波那契查找都要求数据有序；排序成本也应计入整体方案评估。
- 闭区间二分的循环条件是 `left <= right`，更新时要越过中点，否则可能死循环。
- 数组含重复值时，上述算法只保证返回某个匹配位置；寻找第一个/最后一个位置要继续收缩边界。
- 插值公式要防止分母为零、中间乘法溢出和估计位置越界；示例会提升差值类型，Go 在乘积过大时退回安全的中点。
- 斐波那契查找在数组长度不是斐波那契数时，会把探测下标限制到末尾，最后还要检查剩余元素；斐波那契计数本身也应使用足够宽的类型。
- 浮点数据存在精度问题，通常不应直接用 `==` 做精确匹配。

## 练习建议

1. 分别实现 `lowerBound`（第一个不小于目标）与 `upperBound`（第一个大于目标）。
2. 用二分查找求旋转有序数组中的目标位置。
3. 在答案空间上二分：求平方根、最小可行速度或最大可行容量。
4. 生成均匀分布与指数分布数据，对比二分和插值查找的比较次数。
5. 为四种算法补充空数组、不存在、首元素、尾元素和重复值测试。
