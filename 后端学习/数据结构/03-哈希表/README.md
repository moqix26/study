# 哈希表

## 学习目标

- 理解哈希函数、桶、冲突和负载因子的关系。
- 掌握链地址法与开放寻址法的插入、查询、删除逻辑。
- 能根据数据规模、删除频率和内存特点选择冲突解决方案。

## 问题模型与直觉

哈希表解决的是“已知键，怎样尽量不经过逐个比较就直接找到对应值”的问题。数组已经能通过整数下标 O(1) 访问；哈希表的核心想法是把任意键先转换成一个整数哈希值，再映射成数组桶下标，从而把按键查询转化为按下标访问。

整个映射分两层：

1. **哈希函数**把键转换成一个较大范围的整数哈希值，例如 `hash(key)`。
2. **桶定位**把哈希值压缩到容量 `m` 的数组范围，常见形式是 `index = hash(key) mod m`。

如果键是字符串，哈希函数必须综合所有字符；如果键是对象，必须选择真正参与“键相等”判断的字段。理想哈希函数应满足：同一个键每次结果稳定、计算足够快、不同键尽量均匀分散。它不要求不同键一定产生不同值，因为当键空间远大于桶数组时，冲突在数学上不可避免。

例如容量为 5，整数键直接按 5 取模，则键 1 和 6 都映射到桶 1。哈希表必须在桶内继续保存并区分二者，不能因为发生冲突就覆盖旧键。

## 键相等与哈希一致性

哈希表判断键通常分两步：先比较哈希或进入同一桶，再用相等规则确认键。自定义键必须遵守：

> 若 `a` 与 `b` 按键语义相等，则 `hash(a)` 必须等于 `hash(b)`。

反过来不成立：哈希值相同的键可以不相等，这正是冲突。若相等对象产生不同哈希值，查询会进入另一个桶，即使表中实际存在该键也找不到。

可变对象作为键还会带来风险：插入后若参与哈希或相等判断的字段发生改变，它的新哈希下标可能与存放位置不同，导致键“失踪”。因此键通常应不可变，或至少在驻留哈希表期间不改变相关字段。

## 负载因子与扩容

设元素数为 `n`、桶容量为 `m`，负载因子定义为：

`alpha = n / m`

- 链地址法中，`alpha` 可理解为平均每个桶中的元素数。它可以超过 1，但值越大，桶内扫描通常越长。
- 开放寻址法中，每个桶最多放一个元素，所以必须有 `n <= m`。当 `alpha` 接近 1 时，空槽越来越难找到，查询和插入的探测次数会急剧增加。

工程实现通常在负载因子超过阈值时扩容，例如把容量扩大约两倍。扩容不能只延长原数组，因为桶下标依赖容量：`hash mod 5` 与 `hash mod 11` 的结果可能不同。正确做法是创建新桶数组，把每个有效键按新容量重新计算下标并插入，这称为**重新哈希**。

一次重新哈希是 O(n)，但若容量按常数倍增长，连续插入中的扩容次数是对数级，总搬移量为 O(n)，因此插入通常仍可视为均摊 O(1)。

## 链地址法

### 存储不变量

桶数组中的每个位置保存一个容器，常见是链表、动态数组；冲突严重时也可树化。任意键值对必须位于由其哈希值计算出的桶中。同一键最多保留一份记录，重复 `put` 应更新值而不是追加重复键。

### 插入或更新

1. 计算目标桶下标。
2. 扫描该桶，使用键相等规则查找已有记录。
3. 若找到相同键，更新其值，元素总数不变。
4. 若未找到，在桶中追加新键值对，元素总数加一。
5. 若负载因子超过阈值，触发扩容并重新哈希。

### 查询

1. 由查询键计算桶下标。
2. 只扫描该桶，不需要查看其他桶。
3. 找到相等键就返回对应值；扫描完仍未找到则判定不存在。

### 删除

定位桶后，在桶内找到相等键并移除节点即可。删除不会破坏同桶其他键的查找路径，所以无需特殊占位标记。

### 手工推演

容量为 5，依次插入 `(1, one)` 与 `(6, six)`。假设两者哈希后仍可视为整数本身：

| 键 | 桶下标 | 桶 1 的内容 |
| ---: | ---: | --- |
| 1 | `1 mod 5 = 1` | `(1, one)` |
| 6 | `6 mod 5 = 1` | `(1, one) -> (6, six)` |

查询键 6 时只进入桶 1，先比较键 1，不相等；再比较键 6，命中并返回 `six`。删除键 1 后，桶 1 仍含 `(6, six)`，查询 6 不受影响。

## 开放寻址法

开放寻址把所有键值对直接放在同一个槽位数组中。冲突后不创建链表，而是沿确定的探测序列寻找其他槽位。本文示例使用线性探测：

`position(i) = (start + i) mod capacity, i = 0, 1, 2, ...`

其中 `start` 是键的初始桶下标。除了线性探测，还常见二次探测和双重哈希；它们的目标都是减少连续聚集。

### 三态槽位不变量

每个槽位必须区分三种状态：

- `EMPTY`：从未使用。查询走到这里可以停止，因为按同样探测序列插入的键不可能越过一个从未使用的槽位。
- `OCCUPIED`：当前保存有效键值对。
- `DELETED`：曾经使用、后来删除的墓碑。查询不能在这里停止，因为目标可能位于探测链后方；插入则可以记录并复用它。

仅用“是否有值”两态无法正确实现删除。把删除槽直接恢复为 `EMPTY` 会截断后续冲突键的查询路径。

### 插入或更新

1. 从初始桶开始沿探测序列检查。
2. 遇到相同键的 `OCCUPIED` 槽，更新值并结束。
3. 遇到第一个 `DELETED` 槽时先记下位置，但继续探测，以防后方已存在相同键。
4. 遇到 `EMPTY` 时，优先把新键放入此前记录的第一个墓碑；没有墓碑才使用当前空槽。
5. 探测整张表仍没有空槽时，若记录过墓碑可复用；否则表已满，需要扩容或报告失败。

若见到墓碑就立即插入而不继续找相同键，可能使一个键在表中出现两份，破坏更新语义。

### 查询

从键的初始桶沿相同序列探测：

- 遇到同键的占用槽，查询成功。
- 遇到墓碑，继续。
- 遇到从未使用的空槽，立即判定不存在。
- 最多探测 `capacity` 次，避免表满或状态异常时无限循环。

### 删除

找到目标后把状态改为 `DELETED`，并按需要清理值，但不能改为 `EMPTY`。墓碑过多会使成功和失败查询都走更长路径，因此扩容策略常同时考虑有效元素数与“有效元素 + 墓碑”的占用量，必要时在相同容量上重新哈希以清除墓碑。

### 手工推演：墓碑为何必要

容量为 7，键 2、9、16 的初始桶都为 2。线性探测后：

| 槽位 | 0 | 1 | 2 | 3 | 4 | 5 | 6 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 插入后 | 空 | 空 | 2 | 9 | 16 | 空 | 空 |

查询 16 的路径是 `2 -> 3 -> 4`。现在删除键 9：

| 槽位 | 2 | 3 | 4 |
| --- | --- | --- | --- |
| 正确状态 | `OCCUPIED(2)` | `DELETED` | `OCCUPIED(16)` |

再次查询 16 时，经过槽 3 的墓碑继续到槽 4，能够命中。如果错误地把槽 3 改为 `EMPTY`，查询会在槽 3 提前停止并误判 16 不存在。

## 为什么算法正确

- 链地址法中，插入、查询和删除都使用同一哈希规则定位桶；所有可能相等的键必然进入同一桶，因此扫描完整个桶足以决定是否存在。
- 开放寻址中，同一键每次生成相同探测序列。插入把键放在序列中第一个可用位置之前不可能存在的“断点”后；查询沿同一序列前进，只有遇到 `EMPTY` 才能证明后方不可能曾插入目标。
- 墓碑保留“这里曾经被越过”的信息，所以删除不会截断其他键的探测链。
- 重新哈希逐个按新容量插入所有有效键，因而恢复了“键位于其新探测规则可达位置”的不变量。

## 复杂度

| 方案/操作 | 平均时间 | 最坏时间 | 额外空间 |
| --- | --- | --- | --- |
| 链地址法查询/插入/删除 | O(1) | O(n) | O(n + m) |
| 开放寻址查询/插入/删除 | O(1) | O(m) | O(m) |
| 扩容并重新哈希 | O(n + m') | O(n + m') | O(m') |

其中 `n` 是有效元素数，`m` 是当前桶容量，`m'` 是扩容后的新容量。工程实现通常维持 `m = Theta(n)`，并令 `m'` 是 `m` 的常数倍；在这个前提下，表中的 O(m) 与 O(n + m') 都可简写为 O(n)。这里保留容量符号，是为了说明墓碑很多或容量明显大于有效元素数时，开放寻址仍可能检查接近整张长度为 `m` 的表。哈希表的 O(1) 是哈希分布合理且负载受控时的平均结论，不是无条件保证。

### 复杂度如何推导

- 链地址法在均匀散列假设下，每桶平均约 `alpha = n/m` 个元素，因此操作期望成本为 O(1 + alpha)。当容量与元素数同阶、负载受控时简写为 O(1)。若所有键落入同一桶，需扫描 n 个元素，最坏 O(n)。
- 开放寻址的成本等于探测槽位数。负载较低且分布合理时，期望探测次数是常数；接近满表、键大量聚集或墓碑过多时，成功与失败查询都可能扫描全部 `m` 个槽位，所以严格最坏时间是 O(m)。只有在扩缩容策略保证 `m = Theta(n)` 时，才可进一步写成 O(n)。
- 扩容先创建并初始化 `m'` 个新槽位，再访问全部 `n` 个有效元素并按新容量重新插入，因此时间为 O(n + m')。迁移期间新旧桶数组会同时存在；若把旧表视为已有输入，新增辅助空间是 O(m')，总峰值容量则是 O(m + m')。常见的常数倍扩容使这些式子都可化简为 O(n)。

## 适用场景

- 字典、集合、缓存、计数器、去重与索引。
- 只关心按键快速访问，不要求键按顺序遍历。
- 链地址法适合删除频繁、元素大小不固定或负载可能偏高的场景。
- 开放寻址法没有链节点开销、缓存局部性好，适合容量可控且追求紧凑存储的场景。

## 不适用场景与安全边界

- 需要按键有序遍历、范围查询、前驱后继时，平衡搜索树通常更合适。
- 需要严格最坏 O(log n) 或 O(1) 响应上界时，普通哈希表可能不合适：链地址法最坏要扫描 n 个元素，开放寻址最坏要扫描 m 个槽位；在常见的 `m = Theta(n)` 策略下二者都表现为 O(n)。
- 哈希函数若可被外部输入刻意构造大量冲突，可能形成哈希洪泛攻击；服务端应使用运行时随机种子、可靠哈希或其他防护。
- 容量不能为 0；开放寻址还必须处理满表，探测次数必须有上限。
- 不能依赖普通哈希表遍历顺序，扩容或运行时随机化都可能改变顺序。

## 三种语言的实现差异

- **C++**：标准容器 `std::unordered_map` 允许自定义哈希器与相等比较器，两者必须保持一致。手写开放寻址时可用枚举状态或可区分的控制字节；对象析构、移动和重新哈希时的引用失效需要特别关注。
- **Java**：`HashMap` 依赖键的 `hashCode()` 与 `equals()`；重写一个时通常必须同时重写另一个。它允许一个 `null` 键，但并发读写不能直接使用普通 `HashMap`。
- **Go**：内建 `map` 提供语言级哈希表，键类型必须可比较；切片、映射和函数不能直接作为键。读取不存在键会得到值类型零值，应使用 `value, ok := m[key]` 区分“不存在”和“存在但值恰为零值”。遍历顺序不保证稳定。

## C++17 完整示例

示例同时实现链地址法和采用线性探测的开放寻址法。为让前面的手工推演在不同编译器和语言中都完全一致，三个版本都把“整数键对容量安全取模”作为教学用哈希函数；真实项目应根据键类型选择分布质量更好的哈希函数，不能把简单取模机械套用到所有输入。

```cpp
#include <iostream>
#include <list>
#include <optional>
#include <stdexcept>
#include <string>
#include <vector>

std::size_t bucketIndex(int key, std::size_t capacity) {
    long long remainder = static_cast<long long>(key) % static_cast<long long>(capacity);
    if (remainder < 0) remainder += static_cast<long long>(capacity);
    return static_cast<std::size_t>(remainder);
}

class ChainedHashMap {
    using Entry = std::pair<int, std::string>;
    std::vector<std::list<Entry>> buckets;

    std::size_t index(int key) const {
        return bucketIndex(key, buckets.size());
    }

public:
    explicit ChainedHashMap(std::size_t capacity) : buckets(capacity) {
        if (capacity == 0) throw std::invalid_argument("capacity must be positive");
    }

    void put(int key, const std::string& value) {
        auto& bucket = buckets[index(key)];
        for (auto& entry : bucket) {
            if (entry.first == key) {
                entry.second = value;
                return;
            }
        }
        bucket.emplace_back(key, value);
    }

    std::optional<std::string> get(int key) const {
        const auto& bucket = buckets[index(key)];
        for (const auto& entry : bucket) {
            if (entry.first == key) return entry.second;
        }
        return std::nullopt;
    }

    bool remove(int key) {
        auto& bucket = buckets[index(key)];
        for (auto it = bucket.begin(); it != bucket.end(); ++it) {
            if (it->first == key) {
                bucket.erase(it);
                return true;
            }
        }
        return false;
    }
};

class LinearProbingHashMap {
    enum class State { EMPTY, OCCUPIED, DELETED };
    struct Slot {
        int key = 0;
        std::string value;
        State state = State::EMPTY;
    };
    std::vector<Slot> table;

    std::size_t start(int key) const {
        return bucketIndex(key, table.size());
    }

public:
    explicit LinearProbingHashMap(std::size_t capacity) : table(capacity) {
        if (capacity == 0) throw std::invalid_argument("capacity must be positive");
    }

    bool put(int key, const std::string& value) {
        std::size_t firstDeleted = table.size();
        for (std::size_t i = 0; i < table.size(); ++i) {
            std::size_t pos = (start(key) + i) % table.size();
            if (table[pos].state == State::OCCUPIED && table[pos].key == key) {
                table[pos].value = value;
                return true;
            }
            if (table[pos].state == State::DELETED && firstDeleted == table.size()) {
                firstDeleted = pos;
            }
            if (table[pos].state == State::EMPTY) {
                if (firstDeleted != table.size()) pos = firstDeleted;
                table[pos] = {key, value, State::OCCUPIED};
                return true;
            }
        }
        if (firstDeleted != table.size()) {
            table[firstDeleted] = {key, value, State::OCCUPIED};
            return true;
        }
        return false;
    }

    std::optional<std::string> get(int key) const {
        for (std::size_t i = 0; i < table.size(); ++i) {
            std::size_t pos = (start(key) + i) % table.size();
            if (table[pos].state == State::EMPTY) return std::nullopt;
            if (table[pos].state == State::OCCUPIED && table[pos].key == key) {
                return table[pos].value;
            }
        }
        return std::nullopt;
    }

    bool remove(int key) {
        for (std::size_t i = 0; i < table.size(); ++i) {
            std::size_t pos = (start(key) + i) % table.size();
            if (table[pos].state == State::EMPTY) return false;
            if (table[pos].state == State::OCCUPIED && table[pos].key == key) {
                table[pos].state = State::DELETED;
                table[pos].value.clear();
                return true;
            }
        }
        return false;
    }
};

std::string show(const std::optional<std::string>& value) {
    return value ? *value : "<不存在>";
}

int main() {
    ChainedHashMap chained(5);
    chained.put(1, "one");
    chained.put(6, "six"); // 1 和 6 在容量 5 时发生冲突
    std::cout << "链地址 1: " << show(chained.get(1)) << "\n";
    std::cout << "链地址 6: " << show(chained.get(6)) << "\n";
    chained.remove(1);
    std::cout << "删除 1 后: " << show(chained.get(1)) << "\n";

    LinearProbingHashMap probing(7);
    probing.put(2, "two");
    probing.put(9, "nine"); // 2 和 9 的初始桶相同
    probing.put(16, "sixteen");
    std::cout << "开放寻址 9: " << show(probing.get(9)) << "\n";
    probing.remove(9);
    std::cout << "删除 9 后仍可查 16: " << show(probing.get(16)) << "\n";
    return 0;
}
```

## Java 17 完整示例

```java
import java.util.LinkedList;
import java.util.ListIterator;

public class Main {
    private record Entry(int key, String value) {}

    private static class ChainedHashMap {
        private final LinkedList<Entry>[] buckets;

        @SuppressWarnings("unchecked")
        ChainedHashMap(int capacity) {
            if (capacity <= 0) throw new IllegalArgumentException("capacity must be positive");
            buckets = (LinkedList<Entry>[]) new LinkedList<?>[capacity];
            for (int i = 0; i < capacity; i++) buckets[i] = new LinkedList<>();
        }

        private int index(int key) {
            return Math.floorMod(Integer.hashCode(key), buckets.length);
        }

        void put(int key, String value) {
            LinkedList<Entry> bucket = buckets[index(key)];
            for (ListIterator<Entry> iterator = bucket.listIterator(); iterator.hasNext();) {
                if (iterator.next().key() == key) {
                    iterator.set(new Entry(key, value));
                    return;
                }
            }
            bucket.add(new Entry(key, value));
        }

        String get(int key) {
            for (Entry entry : buckets[index(key)]) {
                if (entry.key() == key) return entry.value();
            }
            return null;
        }

        boolean remove(int key) {
            return buckets[index(key)].removeIf(entry -> entry.key() == key);
        }
    }

    private static class LinearProbingHashMap {
        private enum State { EMPTY, OCCUPIED, DELETED }
        private static class Slot {
            int key;
            String value;
            State state = State.EMPTY;
        }

        private final Slot[] table;

        LinearProbingHashMap(int capacity) {
            if (capacity <= 0) throw new IllegalArgumentException("capacity must be positive");
            table = new Slot[capacity];
            for (int i = 0; i < capacity; i++) table[i] = new Slot();
        }

        private int start(int key) {
            return Math.floorMod(Integer.hashCode(key), table.length);
        }

        boolean put(int key, String value) {
            int firstDeleted = -1;
            for (int i = 0; i < table.length; i++) {
                int pos = (start(key) + i) % table.length;
                Slot slot = table[pos];
                if (slot.state == State.OCCUPIED && slot.key == key) {
                    slot.value = value;
                    return true;
                }
                if (slot.state == State.DELETED && firstDeleted == -1) firstDeleted = pos;
                if (slot.state == State.EMPTY) {
                    pos = firstDeleted == -1 ? pos : firstDeleted;
                    occupy(table[pos], key, value);
                    return true;
                }
            }
            if (firstDeleted != -1) {
                occupy(table[firstDeleted], key, value);
                return true;
            }
            return false;
        }

        private void occupy(Slot slot, int key, String value) {
            slot.key = key;
            slot.value = value;
            slot.state = State.OCCUPIED;
        }

        String get(int key) {
            for (int i = 0; i < table.length; i++) {
                Slot slot = table[(start(key) + i) % table.length];
                if (slot.state == State.EMPTY) return null;
                if (slot.state == State.OCCUPIED && slot.key == key) return slot.value;
            }
            return null;
        }

        boolean remove(int key) {
            for (int i = 0; i < table.length; i++) {
                Slot slot = table[(start(key) + i) % table.length];
                if (slot.state == State.EMPTY) return false;
                if (slot.state == State.OCCUPIED && slot.key == key) {
                    slot.state = State.DELETED;
                    slot.value = null;
                    return true;
                }
            }
            return false;
        }
    }

    private static String show(String value) {
        return value == null ? "<不存在>" : value;
    }

    public static void main(String[] args) {
        ChainedHashMap chained = new ChainedHashMap(5);
        chained.put(1, "one");
        chained.put(6, "six");
        System.out.println("链地址 1: " + show(chained.get(1)));
        System.out.println("链地址 6: " + show(chained.get(6)));
        chained.remove(1);
        System.out.println("删除 1 后: " + show(chained.get(1)));

        LinearProbingHashMap probing = new LinearProbingHashMap(7);
        probing.put(2, "two");
        probing.put(9, "nine");
        probing.put(16, "sixteen");
        System.out.println("开放寻址 9: " + show(probing.get(9)));
        probing.remove(9);
        System.out.println("删除 9 后仍可查 16: " + show(probing.get(16)));
    }
}
```

## Go 完整示例

```go
package main

import "fmt"

type entry struct {
	key   int
	value string
}

type chainedHashMap struct {
	buckets [][]entry
}

func newChainedHashMap(capacity int) *chainedHashMap {
	if capacity <= 0 {
		panic("capacity must be positive")
	}
	return &chainedHashMap{buckets: make([][]entry, capacity)}
}

func (m *chainedHashMap) index(key int) int {
	index := key % len(m.buckets)
	if index < 0 {
		index += len(m.buckets)
	}
	return index
}

func (m *chainedHashMap) put(key int, value string) {
	index := m.index(key)
	for i := range m.buckets[index] {
		if m.buckets[index][i].key == key {
			m.buckets[index][i].value = value
			return
		}
	}
	m.buckets[index] = append(m.buckets[index], entry{key, value})
}

func (m *chainedHashMap) get(key int) (string, bool) {
	for _, item := range m.buckets[m.index(key)] {
		if item.key == key {
			return item.value, true
		}
	}
	return "", false
}

func (m *chainedHashMap) remove(key int) bool {
	index := m.index(key)
	for i, item := range m.buckets[index] {
		if item.key == key {
			m.buckets[index] = append(m.buckets[index][:i], m.buckets[index][i+1:]...)
			return true
		}
	}
	return false
}

type state uint8

const (
	empty state = iota
	occupied
	deleted
)

type slot struct {
	key   int
	value string
	state state
}

type probingHashMap struct {
	table []slot
}

func newProbingHashMap(capacity int) *probingHashMap {
	if capacity <= 0 {
		panic("capacity must be positive")
	}
	return &probingHashMap{table: make([]slot, capacity)}
}

func (m *probingHashMap) start(key int) int {
	index := key % len(m.table)
	if index < 0 {
		index += len(m.table)
	}
	return index
}

func (m *probingHashMap) put(key int, value string) bool {
	firstDeleted := -1
	for i := 0; i < len(m.table); i++ {
		pos := (m.start(key) + i) % len(m.table)
		if m.table[pos].state == occupied && m.table[pos].key == key {
			m.table[pos].value = value
			return true
		}
		if m.table[pos].state == deleted && firstDeleted == -1 {
			firstDeleted = pos
		}
		if m.table[pos].state == empty {
			if firstDeleted != -1 {
				pos = firstDeleted
			}
			m.table[pos] = slot{key, value, occupied}
			return true
		}
	}
	if firstDeleted != -1 {
		m.table[firstDeleted] = slot{key, value, occupied}
		return true
	}
	return false
}

func (m *probingHashMap) get(key int) (string, bool) {
	for i := 0; i < len(m.table); i++ {
		item := m.table[(m.start(key)+i)%len(m.table)]
		if item.state == empty {
			return "", false
		}
		if item.state == occupied && item.key == key {
			return item.value, true
		}
	}
	return "", false
}

func (m *probingHashMap) remove(key int) bool {
	for i := 0; i < len(m.table); i++ {
		pos := (m.start(key) + i) % len(m.table)
		if m.table[pos].state == empty {
			return false
		}
		if m.table[pos].state == occupied && m.table[pos].key == key {
			m.table[pos].state = deleted
			m.table[pos].value = ""
			return true
		}
	}
	return false
}

func show(value string, ok bool) string {
	if !ok {
		return "<不存在>"
	}
	return value
}

func main() {
	chained := newChainedHashMap(5)
	chained.put(1, "one")
	chained.put(6, "six")
	value, ok := chained.get(1)
	fmt.Println("链地址 1:", show(value, ok))
	value, ok = chained.get(6)
	fmt.Println("链地址 6:", show(value, ok))
	chained.remove(1)
	value, ok = chained.get(1)
	fmt.Println("删除 1 后:", show(value, ok))

	probing := newProbingHashMap(7)
	probing.put(2, "two")
	probing.put(9, "nine")
	probing.put(16, "sixteen")
	value, ok = probing.get(9)
	fmt.Println("开放寻址 9:", show(value, ok))
	probing.remove(9)
	value, ok = probing.get(16)
	fmt.Println("删除 9 后仍可查 16:", show(value, ok))
}
```

## 示例输出

```text
链地址 1: one
链地址 6: six
删除 1 后: <不存在>
开放寻址 9: nine
删除 9 后仍可查 16: sixteen
```

## 易错点

- 直接对负数键取模可能得到负下标，应使用安全取模或无符号哈希值。
- 开放寻址删除后必须保留“墓碑”，否则查询会在该处提前停止，漏掉探测链后方的键。
- 墓碑过多也会降低性能，需要定期重新哈希。
- 更新已有键时不能增加元素计数；扩容阈值应依据实际占用和墓碑策略设计。
- 自定义对象作为键时，哈希值与相等判断必须一致：相等对象必须具有相同哈希值。
- 不要依赖普通哈希表的遍历顺序。

## 练习建议

1. 为两种实现增加 `size`、负载因子阈值与自动扩容。
2. 把线性探测改为二次探测或双重哈希，比较聚集现象。
3. 实现字符串键的多项式哈希，并用大量随机字符串统计桶分布。
4. 用哈希表解决“两数之和”“最长连续序列”和 LRU 缓存中的快速定位问题。
