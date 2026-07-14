# Trie 字典树

## 学习目标

- 理解 Trie 用“共享前缀路径”存储字符串集合的结构不变量。
- 掌握插入、完整单词查找、前缀判断和按前缀枚举。
- 能分析 Trie 的时间、空间复杂度，并与哈希表、BST、压缩前缀树比较。
- 理解删除、字符集选择和 Unicode 处理中的常见陷阱。

## 1. 原理讲解：核心结构与不变量

Trie（前缀树/字典树）把字符串集合按字符前缀组织成树。它与 BST 的关键区别是：BST 的一次分支由“整个键的大小比较”决定；Trie 的一层只消费一个字符，查找过程与已有单词数量和树的平衡程度无关，主要取决于待查字符串长度。

根节点代表空前缀。边带有字符标签；从根沿边走到某节点，把沿途字符按顺序拼接起来，就得到该节点代表的唯一前缀。有些实现把字符存到孩子节点中，语义等价，但应明确“真正区分分支的是从父到子的字符”。

节点通常包含：

- 指向下一字符节点的孩子集合；
- `isWord` 标记，表示当前路径本身是否是一条完整单词；
- 可选的计数、权重、Top K 建议、原始记录编号等增强信息。

核心不变量：

1. 根代表空前缀，不对应普通字符。
2. 从同一节点出发，同一个字符至多有一条边。
3. 一个已插入单词的每个前缀都对应一条可达路径。
4. 路径存在不等于单词存在；只有末端节点 `isWord=true` 才表示完整单词。

例如插入 `app` 与 `apple` 后，两者共享 `a -> p -> p`。节点 `app` 必须标记为单词结尾，而 `appl` 只是前缀。

以单词集合 `{app, apple, apply, bat, bath}` 为例，可画成：

```text
(root)
├─ a
│  └─ p
│     └─ p  [word: app]
│        └─ l
│           ├─ e  [word: apple]
│           └─ y  [word: apply]
└─ b
   └─ a
      └─ t  [word: bat]
         └─ h  [word: bath]
```

`app` 节点既是完整单词结尾，又有孩子，因此“叶子节点”等价于“单词结尾”是错误的。`ap` 路径存在却没有结尾标记，所以它只是前缀。`bath` 的存在也不能取消 `bat` 自己的结尾标记。

Trie 的结构不变量可以从“前缀唯一性”理解：同一父节点下相同字符只能有一个孩子，因此某个前缀最多对应一个节点；所有拥有该前缀的单词都共享到该节点为止的路径，之后再按下一个字符分叉。

## 2. 核心操作

### 2.1 插入

从根依次读取字符。对应孩子不存在就创建，存在就复用。处理完最后一个字符后设置 `isWord=true`。重复插入在集合语义下是幂等的；若需要频次，可维护 `wordCount`。

按顺序插入示例单词时：

1. 插入 `apple`：根下没有 `a`，依次创建 `a-p-p-l-e` 五个节点，把 `e` 节点标为单词。
2. 插入 `app`：`a-p-p` 三段路径都已存在，不创建节点，只把第二个 `p` 节点标为单词。原有 `apple` 路径继续保留。
3. 插入 `apply`：复用 `a-p-p-l`，在 `l` 下新建 `y` 分支并标记。
4. 插入 `bat`：根下新建独立的 `b-a-t` 路径。
5. 插入 `bath`：复用 `b-a-t`，从 `t` 新建 `h`；`t` 仍保持 `bat` 的结尾标记。

插入后，每个已处理前缀都存在；最后节点被正确标记为单词。因此完整单词一定可沿路径重新找到。操作不会修改其他分支，也不会取消已有终点标记，所以不会破坏其他单词。

若允许重复单词并需要统计频次，单个 `isWord` 不够，可用 `wordCount` 表示某个完整单词出现次数。还可在沿途节点维护 `prefixCount`，表示有多少个单词实例经过当前前缀。

### 2.2 完整查找与前缀查找

二者都先沿字符路径下降：

- `search(word)` 要求路径存在且终点 `isWord=true`。
- `startsWith(prefix)` 只要求整条路径存在，不关心终点是否为单词。

这是最常见的逻辑差异。把两者都写成“路径存在就返回 true”会把 `ap` 错判为已插入单词。

查找 `app` 时，依次从根走过 `a、p、p`，终点 `isWord=true`，所以成功。查找 `ap` 时路径也完整存在，但终点没有单词标记，因此 `search("ap")` 为假，而 `startsWith("ap")` 为真。

查找 `apt` 时，在走过 `a、p` 后找不到字符 `t` 的孩子，可以立即失败，不必检查其他分支。查找 `apple` 则走到 `e` 终点并检查标记。由此可见，完整查找和前缀查找共享同一个 `walk` 过程，差别仅在到达终点后的判断。

正确性来自路径唯一性：每一步当前前缀和下一个字符共同确定唯一孩子。若某一步孩子不存在，就不存在任何已插入单词能拥有这条完整前缀；若所有边都存在，则终点正是该字符串对应的节点，再由结尾标记区分完整词和纯前缀。

### 2.3 前缀枚举

先走到前缀末端，再从该节点 DFS。每到 `isWord=true` 的节点就收集当前字符串。若孩子按字符数组下标从小到大遍历，结果天然是字典序；若孩子使用哈希表，迭代顺序通常不稳定，需要排序键或最终结果。

枚举前缀 `app` 时，先用 `O(3)` 走到第二个 `p` 节点：

1. 当前节点本身是单词结尾，先收集 `app`。
2. 继续走字符 `l`，到达前缀 `appl`，它不是完整单词。
3. 按字符顺序访问 `e`，收集 `apple`；回溯并撤销字符 `e`。
4. 再访问 `y`，收集 `apply`。

因此结果为 `app, apple, apply`。回溯时必须恢复字符串缓冲区到进入孩子前的长度；否则处理完 `e` 分支后再去 `y` 分支，路径可能错误拼成 `appley`。

输出 `k` 个单词至少要写出这些字符串，因此枚举复杂度必须包含输出成本。若自动补全只要前 10 个结果，可以在 DFS 收集够 10 个后停止；若还按热度排序，常在节点上缓存 Top K 候选，避免每次扫描整棵前缀子树。

### 2.4 删除

删除分两步：先把单词末端 `isWord` 取消，再自底向上清理无用节点。只有同时满足以下条件的节点才可剪掉：

- 它不再是任何单词的结尾；
- 它没有孩子。

删除 `app` 时不能直接删掉整条路径，因为 `apple` 仍依赖同一前缀。若维护前缀计数，可在路径上递减计数，并在计数归零时剪枝。

删除前必须先确认完整单词存在。若路径中途不存在，或终点 `isWord=false`，说明待删字符串未存储，不能递减计数或误删共享节点。

以集合 `{app, apple, apply}` 为例：

- 删除 `app`：只把第二个 `p` 的 `isWord` 改为假。该节点仍有 `l` 孩子，整条路径都不能剪，`apple、apply` 保留。
- 删除 `apple`：取消 `e` 节点标记。`e` 没有孩子，可删除 `e`；其父 `l` 仍有 `y` 孩子，所以到此停止。
- 接着删除 `apply`：`y` 可删除；此时 `l` 不再是单词且没有孩子，也可删除。若 `app` 此前已经删除，则向上的 `p` 链还可继续检查；若 `app` 仍存在，就必须在标记节点停止。

有两种常用删除实现：

1. **递归剪枝**：递归函数返回“当前节点是否可被父节点删除”，条件是当前节点非单词且没有孩子。它不需要计数，但每层可能检查全部孩子槽。
2. **计数剪枝**：插入时沿路径增加 `prefixCount`，删除时递减；某个孩子计数降为 0 时，其后整棵分支都不再被任何单词使用，可直接断开。若支持重复词，还要配合 `wordCount`，只有频次确实大于 0 才允许删除一次。

若节点关联业务记录，逻辑删除可以只减少结尾计数而暂不释放节点；高并发读场景下直接释放共享节点还涉及生命周期和同步问题。

## 3. 字符集与存储选择

本文示例只接受小写英文字母，用长度 26 的孩子数组：访问快、常数稳定，但稀疏节点会浪费空间。

| 孩子表示 | 单步访问 | 空间特点 | 适合情况 |
| --- | --- | --- | --- |
| 固定数组 `[26]` | `O(1)` | 每节点固定 26 个指针 | 小写字母、节点较稠密 |
| 哈希表 | 平均 `O(1)` | 只存实际孩子，有哈希开销 | 大字符集、分支稀疏 |
| 有序映射 | `O(log σ)` | 支持稳定字典序 | 需有序枚举 |
| 压缩边字符串 | 依实现而定 | 合并单孩子链 | 路径很长、稀疏前缀 |

UTF-8 中一个用户可见字符可能占多个字节。Go 若按 `byte` 建 Trie 实际存的是 UTF-8 字节路径；Java 的 `char` 是 UTF-16 代码单元；C++ `char` 通常也是字节。处理任意 Unicode 时应先明确单位：字节、Unicode 码点还是用户感知字符，并使用适合的映射键。

固定数组访问的真实成本稳定：字符减去 `'a'` 得到 `0..25` 下标。但每个节点都预留 26 个引用，即使绝大多数为空。若有一百万个稀疏节点，空槽成本可能远大于字符数据本身。哈希孩子只存实际分支，更节省稀疏空间，但每次访问要计算哈希并有桶结构开销。

有序映射适合必须稳定按字符顺序枚举的场景。压缩 Trie/Radix Tree 会把单孩子链 `a->p->p->l` 合并为一条字符串边 `appl`，显著减少节点数，但查找时要比较一段边标签，插入还可能在边中间拆分。

Unicode 还要区分规范化形式。例如视觉相同的字符可能由单个码点或“基础字符 + 组合符号”表示；如果业务希望它们视为同一个词，应在进入 Trie 前统一大小写、正规化和分词规则。数据结构本身不会自动完成语言规范化。

## 4. 复杂度

设待处理字符串长度为 `L`，前缀长度为 `P`，字符集大小为 `σ`：

| 操作 | 固定数组孩子 | 哈希孩子 |
| --- | --- | --- |
| 插入 | `O(L)` | 平均 `O(L)` |
| 完整查找 | `O(L)` | 平均 `O(L)` |
| 前缀判断 | `O(P)` | 平均 `O(P)` |
| 前缀枚举 | `O(P + 输出相关子树规模)` | 同阶，排序时另计 |

空间不是简单的“单词数乘长度”，而是所有**不同前缀节点数**乘每节点开销。最坏无共享时为 `O(总字符数)` 个节点；大量共享前缀时节点数显著减少，但固定孩子数组的指针槽仍可能占用大量内存。

插入和查找之所以是 `O(L)`，是因为每个字符恰好决定一次孩子访问。固定数组为严格常数访问；哈希孩子通常写平均 `O(L)`，最坏行为还取决于哈希表实现和冲突情况。有序映射孩子每个字符需要 `O(log σ)`，总计 `O(L log σ)`。

枚举不能只写 `O(P)`：定位前缀后必须遍历相关子树并构造输出。若访问子树中 `v` 个节点、输出总字符数为 `S`，可以写成 `O(P+v+S)`；是否排序还会增加额外成本。

删除定位单词需要 `O(L)`。固定 26 槽且递归逐层检查“是否有孩子”时仍是 `O(26L)`，因字符集固定可简写 `O(L)`；若字符集很大，维护孩子数量字段可避免每层扫描全部槽。

## 5. 适用场景、限制与结构对比

- 搜索框自动补全、命令补全、路由前缀匹配。
- 词典的完整词与前缀判断。
- 拼写检查、敏感词过滤的基础结构。
- 01-Trie：按整数二进制位存储，用于最大异或查询。

与哈希表比较：哈希表擅长完整键等值查询，通常更节省节点对象开销；Trie 的优势是前缀查询成本只与前缀长度有关，并能共享前缀。与 BST 比较：BST 支持更一般的顺序和范围关系，Trie 对字符串前缀更直接。

进一步结构：压缩 Trie/Radix Tree 合并单孩子链；Aho-Corasick 自动机在 Trie 上增加失败指针，可一次扫描匹配多个模式；后缀树/后缀数组解决任意子串问题，不要与前缀树混淆。

Trie 最适合查询语义本身就是“从字符串开头逐字符匹配”的场景。若只需要判断完整键是否存在，哈希表常有更低的对象和指针开销；若需要任意区间、按完整键排序和通用比较器，平衡 BST 更灵活；若要匹配文本中任意位置出现的多个模式，应在 Trie 上构建 Aho-Corasick，而不是对每个文本起点重新搜索。

Trie 也不自动适合模糊搜索。编辑距离、通配符会让查询在多个孩子间分叉，最坏可能访问大量节点，需要结合动态规划、剪枝或专用索引。前缀自动补全若数据非常大，还要考虑结果排序、热度更新、持久化与内存压缩，而不只是路径是否存在。

实现边界包括：输入字符必须符合孩子索引规则；空字符串是否允许要预先定义；集合语义和多重集合语义要区分；删除前必须验证词频；哈希孩子枚举顺序不稳定；递归枚举深度等于最长词长，极长输入可能造成栈问题；并发插入、删除与枚举共享节点时需要同步或不可变快照。

下面三种语言示例实现插入、完整查找、前缀判断和前缀枚举，没有实现 2.4 节的删除；删除保留为末尾练习，实现时应覆盖“删短词但保留长词”和“删最后分支并连续剪枝”两类测试。

## 6. C++17 完整示例

```cpp
#include <array>
#include <iostream>
#include <stdexcept>
#include <string>
#include <vector>
using namespace std;

class Trie {
private:
    struct Node {
        bool isWord = false;
        array<Node*, 26> children{};
    };

    Node* root = new Node();

    static int indexOf(char ch) {
        if (ch < 'a' || ch > 'z') throw invalid_argument("only lowercase a-z is supported");
        return ch - 'a';
    }

    const Node* walk(const string& text) const {
        const Node* current = root;
        for (char ch : text) {
            int index = indexOf(ch);
            if (!current->children[index]) return nullptr;
            current = current->children[index];
        }
        return current;
    }

    void collect(const Node* node, string& current, vector<string>& out) const {
        if (node->isWord) out.push_back(current);
        for (int i = 0; i < 26; ++i) {
            if (!node->children[i]) continue;
            current.push_back(static_cast<char>('a' + i));
            collect(node->children[i], current, out);
            current.pop_back();
        }
    }

    void destroy(Node* node) {
        if (!node) return;
        for (Node* child : node->children) destroy(child);
        delete node;
    }

public:
    ~Trie() { destroy(root); }

    void insert(const string& word) {
        Node* current = root;
        for (char ch : word) {
            int index = indexOf(ch);
            if (!current->children[index]) current->children[index] = new Node();
            current = current->children[index];
        }
        current->isWord = true;
    }

    bool search(const string& word) const {
        const Node* node = walk(word);
        return node && node->isWord;
    }

    bool startsWith(const string& prefix) const {
        return walk(prefix) != nullptr;
    }

    vector<string> wordsWithPrefix(const string& prefix) const {
        vector<string> out;
        const Node* node = walk(prefix);
        if (!node) return out;
        string current = prefix;
        collect(node, current, out);
        return out;
    }
};

int main() {
    Trie trie;
    for (const string& word : {"apple", "app", "apply", "bat", "bath"}) trie.insert(word);
    cout << boolalpha;
    cout << "search app: " << trie.search("app") << '\n';
    cout << "search ap: " << trie.search("ap") << '\n';
    cout << "startsWith ap: " << trie.startsWith("ap") << '\n';
    cout << "prefix app:";
    for (const string& word : trie.wordsWithPrefix("app")) cout << ' ' << word;
    cout << '\n';
    return 0;
}
```

## 7. Java 17 完整示例

```java
import java.util.ArrayList;
import java.util.List;

public class Main {
    static class Trie {
        static class Node {
            boolean isWord;
            Node[] children = new Node[26];
        }

        private final Node root = new Node();

        private int indexOf(char ch) {
            if (ch < 'a' || ch > 'z') {
                throw new IllegalArgumentException("only lowercase a-z is supported");
            }
            return ch - 'a';
        }

        void insert(String word) {
            Node current = root;
            for (int i = 0; i < word.length(); i++) {
                int index = indexOf(word.charAt(i));
                if (current.children[index] == null) current.children[index] = new Node();
                current = current.children[index];
            }
            current.isWord = true;
        }

        private Node walk(String text) {
            Node current = root;
            for (int i = 0; i < text.length(); i++) {
                int index = indexOf(text.charAt(i));
                if (current.children[index] == null) return null;
                current = current.children[index];
            }
            return current;
        }

        boolean search(String word) {
            Node node = walk(word);
            return node != null && node.isWord;
        }

        boolean startsWith(String prefix) {
            return walk(prefix) != null;
        }

        private void collect(Node node, StringBuilder current, List<String> out) {
            if (node.isWord) out.add(current.toString());
            for (int i = 0; i < 26; i++) {
                if (node.children[i] == null) continue;
                current.append((char) ('a' + i));
                collect(node.children[i], current, out);
                current.deleteCharAt(current.length() - 1);
            }
        }

        List<String> wordsWithPrefix(String prefix) {
            List<String> out = new ArrayList<>();
            Node node = walk(prefix);
            if (node == null) return out;
            collect(node, new StringBuilder(prefix), out);
            return out;
        }
    }

    public static void main(String[] args) {
        Trie trie = new Trie();
        for (String word : new String[]{"apple", "app", "apply", "bat", "bath"}) trie.insert(word);
        System.out.println("search app: " + trie.search("app"));
        System.out.println("search ap: " + trie.search("ap"));
        System.out.println("startsWith ap: " + trie.startsWith("ap"));
        System.out.println("prefix app: " + trie.wordsWithPrefix("app"));
    }
}
```

## 8. Go 完整示例

```go
package main

import "fmt"

type Node struct {
	IsWord   bool
	Children [26]*Node
}

type Trie struct {
	root *Node
}

func newTrie() *Trie {
	return &Trie{root: &Node{}}
}

func indexOf(ch byte) int {
	if ch < 'a' || ch > 'z' {
		panic("only lowercase a-z is supported")
	}
	return int(ch - 'a')
}

func (t *Trie) insert(word string) {
	current := t.root
	for i := 0; i < len(word); i++ {
		index := indexOf(word[i])
		if current.Children[index] == nil {
			current.Children[index] = &Node{}
		}
		current = current.Children[index]
	}
	current.IsWord = true
}

func (t *Trie) walk(text string) *Node {
	current := t.root
	for i := 0; i < len(text); i++ {
		index := indexOf(text[i])
		if current.Children[index] == nil {
			return nil
		}
		current = current.Children[index]
	}
	return current
}

func (t *Trie) search(word string) bool {
	node := t.walk(word)
	return node != nil && node.IsWord
}

func (t *Trie) startsWith(prefix string) bool {
	return t.walk(prefix) != nil
}

func collect(node *Node, current []byte, out *[]string) {
	if node.IsWord {
		*out = append(*out, string(current))
	}
	for i, child := range node.Children {
		if child == nil {
			continue
		}
		current = append(current, byte('a'+i))
		collect(child, current, out)
		current = current[:len(current)-1]
	}
}

func (t *Trie) wordsWithPrefix(prefix string) []string {
	node := t.walk(prefix)
	if node == nil {
		return []string{}
	}
	out := []string{}
	collect(node, []byte(prefix), &out)
	return out
}

func main() {
	trie := newTrie()
	for _, word := range []string{"apple", "app", "apply", "bat", "bath"} {
		trie.insert(word)
	}
	fmt.Println("search app:", trie.search("app"))
	fmt.Println("search ap:", trie.search("ap"))
	fmt.Println("startsWith ap:", trie.startsWith("ap"))
	fmt.Println("prefix app:", trie.wordsWithPrefix("app"))
}
```

## 9. 示例输出

```text
search app: true
search ap: false
startsWith ap: true
prefix app: [app, apple, apply]
```

C++ 的前缀结果没有方括号，但单词与顺序一致。

## 10. 易错点与要点总结

- 路径存在只代表前缀存在，完整单词还必须检查 `isWord`。
- 插入结束后不要忘记设置单词结尾标记。
- 删除短词时不能删除仍被长词共享的节点。
- 固定 26 个孩子只适用于已约定的小写英文输入；必须验证字符范围。
- 递归枚举时回溯要移除刚追加的字符，否则兄弟分支字符串会串在一起。
- 使用哈希孩子时枚举顺序不一定是字典序。
- 空字符串是否允许要预先定义；当前实现允许插入空字符串并把根标记为单词。
- Trie 的时间优势来自按字符直接下降，但节点和指针开销可能远大于哈希表。
- 处理中文或 emoji 时不要默认一个字节或一个 UTF-16 `char` 就是一个完整字符。

## 11. 扩展练习

1. 实现安全删除，并覆盖“删除前缀词但保留长词”的测试。
2. 每个节点维护 `prefixCount` 和 `wordCount`，支持重复单词计数。
3. 为自动补全维护每个前缀下频率最高的 3 个候选词。
4. 使用 `map/rune`、`Map<Integer,Node>` 等形式实现 Unicode 码点 Trie。
5. 实现 01-Trie，在固定 32 位整数集合中查询与给定数异或最大的元素。
