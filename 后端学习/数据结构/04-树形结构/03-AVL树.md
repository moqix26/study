# AVL 树

## 学习目标

- 理解 AVL 树同时维护的 BST 有序不变量与高度平衡不变量。
- 掌握高度更新、平衡因子以及 LL、RR、LR、RL 四类失衡的判断。
- 能手写插入与旋转，并解释为什么旋转不破坏中序顺序。
- 理解 AVL 高度为何严格为 `O(log n)`，以及它与红黑树的工程取舍。

## 1. 原理讲解：定义与不变量

普通 BST 的查找成本取决于树高，但插入顺序可能把它变成长链。AVL 树在 BST 节点上额外维护高度，并在更新后立即调整局部结构，从而把任意时刻的高度严格限制在 `O(log n)`。它解决的不是“如何比较键”，而是“如何在不破坏有序性的前提下阻止树长期偏向一侧”。

每个 AVL 节点同时满足：

1. **BST 不变量**：左子树所有键更小，右子树所有键更大。
2. **高度不变量**：`height = 1 + max(height(left), height(right))`。
3. **平衡不变量**：平衡因子 `BF = height(left) - height(right)` 只能是 `-1、0、1`。

本文约定空树高度为 0、叶子高度为 1。也可以约定空树为 -1、叶子为 0，只要计算始终一致。

平衡因子采用“左高减右高”：

- `BF = 1`：左子树比右子树高一层，仍合法；
- `BF = 0`：两侧等高；
- `BF = -1`：右子树比左子树高一层，仍合法；
- `|BF| >= 2`：当前节点失衡，必须旋转。

高度字段是一个缓存值。它必须等于由孩子真实高度重新计算出的结果，不能只在插入叶子时修改一次。若某个节点缓存高度错误，祖先的平衡因子也会连续错误，可能选择错误旋转方向。因此验证 AVL 时要同时检查键顺序、真实高度、缓存高度和每个节点的平衡因子。

插入首先按普通 BST 加入叶子。只有从新节点到根的祖先高度可能变化，所以递归返回时自底向上更新高度并修复第一个或后续失衡节点。

一次递归插入可以拆成两个阶段：

1. **向下搜索**：完全按照 BST 比较规则找到空位置，新节点作为叶子加入。
2. **向上回溯**：每返回一层，先用孩子高度更新当前节点，再计算平衡因子；若失衡，根据新键落在较高孩子的哪一侧选择旋转，并把旋转后的新子树根返回给父节点。

回溯顺序必须是自底向上，因为父节点的高度依赖已经修复后的孩子高度。递归函数返回值不是固定的原节点：一旦旋转，子树根会发生变化，调用者必须接住这个新根。

## 2. 旋转为什么正确

旋转是常数次指针重连。它只改变局部父子关系，不增加、删除或交换键。只要旋转前后的中序顺序一致，BST 有序不变量就保持不变。

### 2.1 右旋的结构与区间

以对节点 `y` 右旋为例：

```text
        y                 x
       / \               / \
      x   C    右旋      A   y
     / \                   / \
    A   B                 B   C
```

这里 `A、B、C` 代表整棵子树，而不是单个节点。旋转前由 BST 关系可知：

`A 中所有键 < x < B 中所有键 < y < C 中所有键`

旋转后，`A` 仍是 `x` 的左子树，`B` 从 `x` 的右子树变成 `y` 的左子树，`C` 保持为 `y` 的右子树。所有键的中序次序仍是 `A, x, B, y, C`，所以有序性没有变化。结构上的效果是把过深的 `x` 抬高一层，把旧根 `y` 下沉一层。

指针更新的逻辑顺序是：保存 `x = y.left` 与临时子树 `B = x.right`；令 `x.right = y`；令 `y.left = B`；最后返回 `x` 作为新子树根。若节点带父指针，还要同步更新 `x、y、B` 以及原祖父的连接。

高度更新必须先更新下沉的旧根 `y`，再更新上升的新根 `x`，因为 `x` 的新高度依赖 `y` 的新高度。左旋完全对称。

具体地说，右旋完成指针重连后：

1. `y` 的孩子已经变成 `B、C`，先计算 `height(y)`；
2. `x` 的孩子已经变成 `A、y`，再使用刚更新的 `height(y)` 计算 `height(x)`；
3. 返回 `x`。

若反过来先更新 `x`，它会读取旋转前旧的 `height(y)`，缓存高度可能多一层，错误会继续传播到祖先。

### 2.2 左旋的对称结构

```text
    x                       y
   / \                     / \
  A   y       左旋        x   C
     / \                 / \
    B   C               A   B
```

中序顺序同样始终是 `A, x, B, y, C`。左旋后先更新下沉的 `x`，再更新上升的 `y`。

## 3. 四类插入失衡

设当前失衡节点为 `z`：

| 类型 | 新键落点 | 判断 | 修复 |
| --- | --- | --- | --- |
| LL | `z` 左孩子的左子树 | `BF(z)>1` 且 `key < z.left.key` | 对 `z` 右旋 |
| RR | `z` 右孩子的右子树 | `BF(z)<-1` 且 `key > z.right.key` | 对 `z` 左旋 |
| LR | `z` 左孩子的右子树 | `BF(z)>1` 且 `key > z.left.key` | 先左旋左孩子，再右旋 `z` |
| RL | `z` 右孩子的左子树 | `BF(z)<-1` 且 `key < z.right.key` | 先右旋右孩子，再左旋 `z` |

双旋的本质是先把“折线”拉直成 LL 或 RR，再做一次单旋。插入修复后，该局部子树的高度恢复到插入前的高度，因此其更高祖先通常不再因这次插入而失衡；递归模板仍可统一地继续更新。

类型名称描述的是从失衡节点 `z` 到新节点方向的前两步。例如 LL 表示先走左孩子，再走左孩子；LR 表示先左后右。判断时应定位“第一个失衡祖先”和“更高的孩子方向”，而不是只看整棵树最终偏向哪边。

### 3.1 LL：一次右旋

依次插入 `30, 20, 10`：

```text
插入 10 后：             对 30 右旋后：
      30                       20
     /                        /  \
    20                       10   30
   /
  10
```

回溯到 20 时，`BF(20)=1`，仍合法；回溯到 30 时，左高为 2、右高为 0，`BF(30)=2`。新键 10 位于 30 的左孩子 20 的左侧，因此是 LL。右旋把 20 抬高，三个键的中序顺序仍是 `10,20,30`，两侧高度都变成 1。

### 3.2 RR：一次左旋

依次插入 `10, 20, 30`：

```text
插入 30 后：             对 10 左旋后：
  10                           20
    \                         /  \
     20                      10   30
       \
        30
```

失衡节点 10 的平衡因子为 -2，新键位于右孩子的右侧，是 RR。一次左旋恢复平衡。

### 3.3 LR：先把折线拉直，再右旋

依次插入 `30, 10, 20`：

```text
初始折线：             先对 10 左旋：          再对 30 右旋：
      30                     30                       20
     /                      /                        /  \
    10                     20                      10   30
      \                   /
       20                10
```

30 左侧过高，但新键 20 在左孩子 10 的右侧。若直接对 30 右旋，10 会升为根而 20 仍位于其右侧，局部仍可能偏斜。第一步左旋 10，把中间键 20 抬到左子树根，折线变成标准 LL；第二步再右旋 30。最终三个键中值 20 成为根，天然把较小的 10 和较大的 30 分到两侧。

在代码中，第一步必须写回 `z.left = leftRotate(z.left)`，因为左子树根已从 10 变成 20；随后再返回 `rightRotate(z)`。

### 3.4 RL：先右旋孩子，再左旋祖先

依次插入 `10, 30, 20`：

```text
初始折线：             先对 30 右旋：          再对 10 左旋：
  10                       10                        20
    \                        \                      /  \
     30                       20                  10   30
    /                           \
   20                            30
```

它与 LR 完全镜像：先令 `z.right = rightRotate(z.right)`，把折线整理为 RR，再对 `z` 左旋。

### 3.5 插入回溯的完整判断顺序

对当前节点 `z`，可靠的处理顺序是：

1. 递归插入左或右子树，并接收返回的新子树根；
2. `height(z) = 1 + max(height(left), height(right))`；
3. 计算 `BF(z)`；
4. 若 `BF > 1`，说明左侧过高，再根据新键与 `z.left.key` 判断 LL 或 LR；
5. 若 `BF < -1`，说明右侧过高，再判断 RR 或 RL；
6. 未失衡则原样返回 `z`，失衡则返回旋转后的新根。

插入一个叶子只会让路径上每棵子树的高度最多增加 1。修复最下面的失衡节点后，该局部子树通常恢复到插入前高度，因此更高祖先不会继续因本次插入产生新失衡。通用递归模板仍继续向上更新，这样实现更统一，也能确保缓存高度正确。

## 4. 删除修复思想

AVL 删除先执行 BST 删除，再沿路径向根更新高度。与插入不同，删除后一次旋转可能让子树高度继续降低，所以多个祖先都可能需要修复。判断旋转方向时不能依赖“被删键落在哪侧”，而应查看孩子的平衡因子：

- `BF(z) > 1` 且 `BF(z.left) >= 0`：右旋；否则先左旋左孩子再右旋。
- `BF(z) < -1` 且 `BF(z.right) <= 0`：左旋；否则先右旋右孩子再左旋。

这里的 `>= 0` 和 `<= 0` 包含孩子平衡因子为 0 的删除特有情况。

删除与插入的重要差异是：删除可能让一棵子树高度减少 1。即使在最下面的失衡处完成旋转，修复后的子树高度仍可能比删除前更低，于是祖父、曾祖父也可能继续失衡。因此删除必须一直回溯到根，逐层更新和修复。

例如某节点 `z` 左侧过高，即 `BF(z)=2`：

- 若 `BF(z.left)=1`，左孩子本身偏左，是普通 LL，右旋；
- 若 `BF(z.left)=0`，左右等高，删除场景中仍应右旋，且旋转后局部高度的变化与插入场景可能不同；
- 若 `BF(z.left)=-1`，左孩子偏右，是 LR，先左旋孩子再右旋。

删除实现还继承 BST 删除的三种结构情况。若用后继替换双孩子节点，要从右子树继续删除后继，并在整个返回路径上重新计算高度，不能只修复最初找到的节点。

一个能说明“孩子平衡因子为 0 也要单旋”的具体轨迹是：初始 AVL 树的根为 4，左子树根为 2（孩子为 1、3），右孩子为 5。删除 5 后，`BF(4)=2`，而 `BF(2)=0`；此时仍必须对 4 右旋。旋转后局部根变为 2，左孩子是 1，右孩子是 4，且 4 的左孩子是 3；各节点平衡因子重新回到 `-1..1`。若把删除判断误写成 `BF(z.left) > 0` 才右旋，这个例子就会漏修。若该子树上方还有祖先，调用者还要接住新根 2，并继续向上重算高度。

## 5. 高度与复杂度推导

设高度为 `h` 的 AVL 树最少节点数为 `N(h)`。为了节点最少，它的两棵子树高度应分别为 `h-1` 和 `h-2`，所以：

`N(h) = 1 + N(h-1) + N(h-2)`。

这个递推与斐波那契数同阶，`N(h)` 至少按黄金比例的 `h` 次方增长，因此反推得到 `h = O(log n)`。这不是“平均平衡”，而是严格的最坏情况上界。

按本文“空树高度 0、叶子高度 1”的约定，初值可取 `N(0)=0`、`N(1)=1`：

- `N(2)=1+N(1)+N(0)=2`；
- `N(3)=1+N(2)+N(1)=4`；
- `N(4)=1+N(3)+N(2)=7`；
- `N(5)=1+N(4)+N(3)=12`。

最少节点数随高度快速增长，反过来说明给定节点数时高度不可能线性增长。更精确的上界常写为约 `1.44*log2(n)`，但学习核心是理解递推来源：要让高度为 `h` 且节点尽量少，两棵子树必须取允许的最悬殊高度 `h-1` 与 `h-2`。

| 操作 | 时间复杂度 | 额外空间（递归版） |
| --- | --- | --- |
| 查找 | `O(log n)` | `O(log n)`，迭代可为 `O(1)` |
| 插入 | `O(log n)` | `O(log n)` |
| 删除 | `O(log n)` | `O(log n)` |
| 单次旋转 | `O(1)` | `O(1)` |
| 遍历 | `O(n)` | `O(log n)` |

一次插入或删除只沿一条根到叶路径搜索，并在回溯路径上做常数次高度计算和旋转；路径长度是 `O(log n)`，所以总时间仍是 `O(log n)`。单次旋转只改固定数量的指针和高度字段，不随树规模增长，故为 `O(1)`。

## 6. 适用场景、限制与对比

AVL 的高度约束比红黑树更严格，通常查找路径更短，适合查找显著多于更新的内存有序索引。代价是插入、尤其删除时可能更频繁地旋转和更新高度。

- 需要最坏 `O(log n)` 且读多写少：AVL 很合适。
- 更新频繁、希望标准库级折中：常用红黑树。
- 只需等值查询、不需要顺序：哈希表通常更直接。
- 面向磁盘页或数据库索引：优先 B/B+ 树，降低 I/O 层数。

AVL 的代价是每个节点需要存储高度或平衡因子，更新时还要维护这些增强字段。若节点还带有 `size、sum` 等子树聚合信息，旋转后也必须按“先下沉节点、后上升节点”的依赖顺序同步更新。

以下情况不应仅因为“AVL 很快”就使用它：只做等值查询时哈希表通常更简单；只需反复取最小值时堆更合适；需要磁盘友好索引时二叉分支太低；多线程写入场景还要考虑锁粒度，AVL 较频繁的结构修改会增加同步复杂度。

实现边界包括：重复键策略必须与 BST 层一致；旋转函数必须返回新根；空孩子高度读取必须安全；整数键接近极值时，验证函数宜传可选上下界而非用 `key±1`；递归版即使树高为对数，也应考虑极大数据量下的栈限制。

下面三种语言示例的范围是插入、查找、遍历与不变量验证，未实现删除。Go 版 `check` 使用 `int64` 最小值和最大值作为开区间哨兵；在 64 位平台上，若业务真的允许 `int` 键恰好等于这两个极值，应把验证函数改为“可选上下界 + 是否存在”的形式，否则极值键会被开区间判断误拒绝。

## 7. C++17 完整示例

```cpp
#include <algorithm>
#include <cmath>
#include <iostream>
#include <limits>
#include <vector>
using namespace std;

struct Node {
    int key;
    int height;
    Node* left;
    Node* right;
    explicit Node(int value) : key(value), height(1), left(nullptr), right(nullptr) {}
};

int height(Node* node) { return node ? node->height : 0; }

void updateHeight(Node* node) {
    node->height = 1 + max(height(node->left), height(node->right));
}

int balanceFactor(Node* node) {
    return node ? height(node->left) - height(node->right) : 0;
}

Node* rotateRight(Node* y) {
    Node* x = y->left;
    Node* middle = x->right;
    x->right = y;
    y->left = middle;
    updateHeight(y);
    updateHeight(x);
    return x;
}

Node* rotateLeft(Node* x) {
    Node* y = x->right;
    Node* middle = y->left;
    y->left = x;
    x->right = middle;
    updateHeight(x);
    updateHeight(y);
    return y;
}

Node* insert(Node* root, int key) {
    if (!root) return new Node(key);
    if (key < root->key) root->left = insert(root->left, key);
    else if (key > root->key) root->right = insert(root->right, key);
    else return root;

    updateHeight(root);
    int bf = balanceFactor(root);
    if (bf > 1 && key < root->left->key) return rotateRight(root);       // LL
    if (bf < -1 && key > root->right->key) return rotateLeft(root);     // RR
    if (bf > 1 && key > root->left->key) {                              // LR
        root->left = rotateLeft(root->left);
        return rotateRight(root);
    }
    if (bf < -1 && key < root->right->key) {                            // RL
        root->right = rotateRight(root->right);
        return rotateLeft(root);
    }
    return root;
}

bool contains(Node* root, int key) {
    while (root) {
        if (key == root->key) return true;
        root = key < root->key ? root->left : root->right;
    }
    return false;
}

void inorder(Node* root, vector<int>& out) {
    if (!root) return;
    inorder(root->left, out);
    out.push_back(root->key);
    inorder(root->right, out);
}

void preorder(Node* root, vector<int>& out) {
    if (!root) return;
    out.push_back(root->key);
    preorder(root->left, out);
    preorder(root->right, out);
}

int check(Node* root, bool& ok, long long low, long long high) {
    if (!root) return 0;
    if (root->key <= low || root->key >= high) ok = false;
    int leftHeight = check(root->left, ok, low, root->key);
    int rightHeight = check(root->right, ok, root->key, high);
    int expected = 1 + max(leftHeight, rightHeight);
    if (abs(leftHeight - rightHeight) > 1 || root->height != expected) ok = false;
    return expected;
}

void print(const vector<int>& values) {
    for (int value : values) cout << value << ' ';
    cout << '\n';
}

void destroy(Node* root) {
    if (!root) return;
    destroy(root->left);
    destroy(root->right);
    delete root;
}

int main() {
    Node* root = nullptr;
    for (int key : {10, 20, 30, 40, 50, 25}) root = insert(root, key);
    vector<int> in, pre;
    inorder(root, in);
    preorder(root, pre);
    cout << "inorder: "; print(in);
    cout << "preorder: "; print(pre);
    cout << boolalpha << "contains 25: " << contains(root, 25) << '\n';
    bool valid = true;
    check(root, valid, numeric_limits<long long>::min(), numeric_limits<long long>::max());
    cout << "valid AVL: " << valid << '\n';
    destroy(root);
    return 0;
}
```

## 8. Java 17 完整示例

```java
import java.util.ArrayList;
import java.util.List;

public class Main {
    static class Node {
        int key;
        int height = 1;
        Node left;
        Node right;

        Node(int key) {
            this.key = key;
        }
    }

    static int height(Node node) {
        return node == null ? 0 : node.height;
    }

    static void updateHeight(Node node) {
        node.height = 1 + Math.max(height(node.left), height(node.right));
    }

    static int balanceFactor(Node node) {
        return node == null ? 0 : height(node.left) - height(node.right);
    }

    static Node rotateRight(Node y) {
        Node x = y.left;
        Node middle = x.right;
        x.right = y;
        y.left = middle;
        updateHeight(y);
        updateHeight(x);
        return x;
    }

    static Node rotateLeft(Node x) {
        Node y = x.right;
        Node middle = y.left;
        y.left = x;
        x.right = middle;
        updateHeight(x);
        updateHeight(y);
        return y;
    }

    static Node insert(Node root, int key) {
        if (root == null) return new Node(key);
        if (key < root.key) root.left = insert(root.left, key);
        else if (key > root.key) root.right = insert(root.right, key);
        else return root;

        updateHeight(root);
        int bf = balanceFactor(root);
        if (bf > 1 && key < root.left.key) return rotateRight(root);
        if (bf < -1 && key > root.right.key) return rotateLeft(root);
        if (bf > 1 && key > root.left.key) {
            root.left = rotateLeft(root.left);
            return rotateRight(root);
        }
        if (bf < -1 && key < root.right.key) {
            root.right = rotateRight(root.right);
            return rotateLeft(root);
        }
        return root;
    }

    static boolean contains(Node root, int key) {
        while (root != null) {
            if (key == root.key) return true;
            root = key < root.key ? root.left : root.right;
        }
        return false;
    }

    static void inorder(Node root, List<Integer> out) {
        if (root == null) return;
        inorder(root.left, out);
        out.add(root.key);
        inorder(root.right, out);
    }

    static void preorder(Node root, List<Integer> out) {
        if (root == null) return;
        out.add(root.key);
        preorder(root.left, out);
        preorder(root.right, out);
    }

    static int check(Node root, boolean[] ok, long low, long high) {
        if (root == null) return 0;
        if (root.key <= low || root.key >= high) ok[0] = false;
        int leftHeight = check(root.left, ok, low, root.key);
        int rightHeight = check(root.right, ok, root.key, high);
        int expected = 1 + Math.max(leftHeight, rightHeight);
        if (Math.abs(leftHeight - rightHeight) > 1 || root.height != expected) ok[0] = false;
        return expected;
    }

    public static void main(String[] args) {
        Node root = null;
        for (int key : new int[]{10, 20, 30, 40, 50, 25}) root = insert(root, key);
        List<Integer> in = new ArrayList<>();
        List<Integer> pre = new ArrayList<>();
        inorder(root, in);
        preorder(root, pre);
        System.out.println("inorder: " + in);
        System.out.println("preorder: " + pre);
        System.out.println("contains 25: " + contains(root, 25));
        boolean[] valid = {true};
        check(root, valid, Long.MIN_VALUE, Long.MAX_VALUE);
        System.out.println("valid AVL: " + valid[0]);
    }
}
```

## 9. Go 完整示例

```go
package main

import "fmt"

type Node struct {
	Key         int
	Height      int
	Left, Right *Node
}

func height(node *Node) int {
	if node == nil {
		return 0
	}
	return node.Height
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func updateHeight(node *Node) {
	node.Height = 1 + max(height(node.Left), height(node.Right))
}

func balanceFactor(node *Node) int {
	if node == nil {
		return 0
	}
	return height(node.Left) - height(node.Right)
}

func rotateRight(y *Node) *Node {
	x := y.Left
	middle := x.Right
	x.Right = y
	y.Left = middle
	updateHeight(y)
	updateHeight(x)
	return x
}

func rotateLeft(x *Node) *Node {
	y := x.Right
	middle := y.Left
	y.Left = x
	x.Right = middle
	updateHeight(x)
	updateHeight(y)
	return y
}

func insert(root *Node, key int) *Node {
	if root == nil {
		return &Node{Key: key, Height: 1}
	}
	if key < root.Key {
		root.Left = insert(root.Left, key)
	} else if key > root.Key {
		root.Right = insert(root.Right, key)
	} else {
		return root
	}

	updateHeight(root)
	bf := balanceFactor(root)
	if bf > 1 && key < root.Left.Key {
		return rotateRight(root)
	}
	if bf < -1 && key > root.Right.Key {
		return rotateLeft(root)
	}
	if bf > 1 && key > root.Left.Key {
		root.Left = rotateLeft(root.Left)
		return rotateRight(root)
	}
	if bf < -1 && key < root.Right.Key {
		root.Right = rotateRight(root.Right)
		return rotateLeft(root)
	}
	return root
}

func contains(root *Node, key int) bool {
	for root != nil {
		if key == root.Key {
			return true
		}
		if key < root.Key {
			root = root.Left
		} else {
			root = root.Right
		}
	}
	return false
}

func inorder(root *Node, out *[]int) {
	if root == nil {
		return
	}
	inorder(root.Left, out)
	*out = append(*out, root.Key)
	inorder(root.Right, out)
}

func preorder(root *Node, out *[]int) {
	if root == nil {
		return
	}
	*out = append(*out, root.Key)
	preorder(root.Left, out)
	preorder(root.Right, out)
}

func check(root *Node, low, high int64) (int, bool) {
	if root == nil {
		return 0, true
	}
	if int64(root.Key) <= low || int64(root.Key) >= high {
		return 0, false
	}
	lh, leftOK := check(root.Left, low, int64(root.Key))
	rh, rightOK := check(root.Right, int64(root.Key), high)
	expected := 1 + max(lh, rh)
	diff := lh - rh
	if diff < 0 {
		diff = -diff
	}
	return expected, leftOK && rightOK && diff <= 1 && root.Height == expected
}

func main() {
	var root *Node
	for _, key := range []int{10, 20, 30, 40, 50, 25} {
		root = insert(root, key)
	}
	in, pre := []int{}, []int{}
	inorder(root, &in)
	preorder(root, &pre)
	fmt.Println("inorder:", in)
	fmt.Println("preorder:", pre)
	fmt.Println("contains 25:", contains(root, 25))
	_, valid := check(root, -1<<63, 1<<63-1)
	fmt.Println("valid AVL:", valid)
}
```

## 10. 示例输出

```text
inorder: [10, 20, 25, 30, 40, 50]
preorder: [30, 20, 10, 25, 40, 50]
contains 25: true
valid AVL: true
```

C++ 输出同样的键顺序，但列表不带方括号。

## 11. 易错点与要点总结

- 平衡因子符号必须统一；本文是“左高减右高”。
- 新节点高度应与空树高度约定匹配，本文叶子高度为 1。
- 旋转后先更新下沉节点，再更新上升节点。
- LR/RL 是两次旋转，不能只旋当前根。
- 递归插入必须接收并返回“旋转后的新子树根”。
- 判断四种插入失衡时，重复键策略必须明确；本文直接忽略重复键。
- 删除可能一路向上多次修复，不能照搬“插入至多修复最下面一个失衡点”的直觉。
- 验证 AVL 时要同时验证 BST 顺序、真实高度、缓存高度和平衡因子，单看根节点不够。

## 12. 扩展练习

1. 在三种语言中实现 AVL 删除，并用随机操作持续验证全部不变量。
2. 给每个节点增加 `size`，支持第 `k` 小和排名查询。
3. 分别插入升序、降序和随机序列，记录旋转次数和最终高度。
4. 把递归插入改成维护父指针的迭代版本。
5. 对比相同输入下普通 BST、AVL、红黑树的高度与更新次数。
