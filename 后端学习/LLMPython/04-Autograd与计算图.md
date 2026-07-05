# Autograd 与计算图

> **文件编码**：UTF-8。  
> **前置**：[03 PyTorch 张量操作](03-PyTorch入门与张量操作.md)、[LLMInfra 01 线代](../LLMInfra/01-线性代数与数值计算基础.md)。  
> **定位**：理解 PyTorch **自动微分**——`requires_grad`、`backward`、`detach`、`no_grad`，训练神经网络的核心机制。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**Autograd** = 记录 tensor 运算构建 **计算图**，对 loss 反向传播自动求 **所有 requires_grad 参数的梯度**。

### 0.2 你需要提前知道什么

| 背景 | 建议 |
|------|------|
| 偏导 / 链式法则 | 高数复习即可 |
| 03 章 tensor | 必须熟练 shape 与 device |
| 05 章 Module | 本章是训练循环的理论基础 |

### 0.3 本章知识地图（☐→☑）

- [ ] 设置 `requires_grad` 并查看 `.grad`
- [ ] 对标量 loss 调用 `backward()`
- [ ] 解释 `retain_graph`、`grad_fn`
- [ ] 正确使用 `detach`、`no_grad`、`inference_mode`
- [ ] 完成 §13 闭卷自测 ≥8/10

### 0.4 建议学习时长

- **3～4 天**

### 0.5 学完你能做什么

手推简单网络的反向传播并与 PyTorch 核对；debug「梯度为 None / 不更新」；区分训练与推理代码模式。

---

## 1. 为什么需要自动微分

训练 = 调整参数 θ 使 loss L(θ) 下降。需 ∂L/∂θ。深度网络手工求导不可行；Autograd **按运算节点链式法则** 自动累积。

```mermaid
flowchart RL
  x["x (input)"] --> w["w (param)"]
  w --> y["y = w * x"]
  y --> L["L = y.sum()"]
  L -->|backward| gw["w.grad"]
```

---

## 2. requires_grad 与 leaf

```python
import torch

w = torch.tensor([2.0, 3.0], requires_grad=True)
x = torch.tensor([1.0, 4.0])   # 默认 requires_grad=False
y = (w * x).sum()
print("y:", y)
print("grad_fn:", y.grad_fn)
print("w is leaf:", w.is_leaf)
print("y is leaf:", y.is_leaf)
```

**预期输出**：

```text
y: tensor(14., grad_fn=<SumBackward0>)
grad_fn: <SumBackward0 object at ...>
w is leaf: True
y is leaf: False
```

- **leaf**：用户创建且 `requires_grad=True` 的 tensor（通常是 `nn.Parameter`）
- 非 leaf 的 `.grad` 默认不保留（省内存）

---

## 3. backward 基础

```python
w = torch.tensor(2.0, requires_grad=True)
y = w ** 2 + 3 * w
y.backward()
print(w.grad)   # dy/dw = 2w + 3 = 7
w.grad.zero_()  # 下次 backward 前必须清零
```

**向量输出**需传入 `gradient` 权重（通常为全 1 或与 loss 同 shape）：

```python
v = torch.tensor([1.0, 2.0], requires_grad=True)
y = v ** 2
y.backward(torch.tensor([1.0, 1.0]))
print(v.grad)   # [2., 4.]  因 d(v_i^2)/dv_i = 2*v_i
```

训练时 **loss 是标量**，直接 `loss.backward()`。

---

## 4. 简单线性模型手算对照

```python
# y = w*x + b, 目标 minimize (y-3)^2, x=2 → y=3 时 loss=0
x = torch.tensor(2.0)
w = torch.tensor(1.0, requires_grad=True)
b = torch.tensor(0.0, requires_grad=True)

y = w * x + b
loss = (y - 3) ** 2
loss.backward()

print("w.grad:", w.grad)  # 2*(y-3)*x = 2*(2-3)*2 = -4
print("b.grad:", b.grad)  # 2*(y-3) = -2
```

---

## 5. 计算图与 grad_fn

```python
a = torch.tensor(1.0, requires_grad=True)
b = a + 2
c = b ** 2
print(c.grad_fn)          # PowBackward0
print(b.grad_fn)          # AddBackward0
c.backward()
print(a.grad)             # dc/da = 2*(a+2) = 6
```

多次 backward 同一图需 `retain_graph=True`（除非中间 `.grad` 已 zero）：

```python
loss.backward(retain_graph=True)
loss.backward()  # 第二次会累加 grad，除非 zero_grad
```

正常训练：**每个 step 一次 backward**，不需 retain。

---

## 6. detach：切断梯度

```python
x = torch.tensor(2.0, requires_grad=True)
y = x ** 2
z = y.detach()          # 新 tensor，与图无关
# z.requires_grad False
try:
    z.backward()
except RuntimeError:
    print("detach 后不能 backward 到 x")
```

**用途**：

- GAN 中冻结判别器输入
- 把 tensor 当常数参与 loss
- 转 numpy / 日志记录

---

## 7. no_grad 与 inference_mode

```python
model_out = torch.randn(4, 10, requires_grad=True)

with torch.no_grad():
    pred = model_out.argmax(dim=-1)
    acc = (pred == torch.randint(0, 10, (4,))).float().mean()
    print(acc)

# PyTorch 1.9+ 推理更快
with torch.inference_mode():
    y = model_out * 2
```

| API | 训练 | 推理 | 说明 |
|-----|------|------|------|
| 默认 | ✓ | | 建图 |
| `no_grad()` | | ✓ | 不建图，可写 tensor |
| `inference_mode()` | | ✓ | 更严格更快 |

05 章训练循环：`model.train()` vs `model.eval()` + `no_grad()`。

---

## 8. 梯度清零与 optimizer

Autograd 只 **累加** `.grad`；**optimizer.step() 不自动清零**。

```python
w = torch.tensor(1.0, requires_grad=True)
opt = torch.optim.SGD([w], lr=0.1)

for step in range(2):
    loss = w ** 2
    loss.backward()
    print("step", step, "grad", w.grad)
    opt.step()
    opt.zero_grad()
```

05 章标准顺序：`zero_grad → forward → loss → backward → step`。

---

## 9. 常见陷阱

### 9.1 原地操作破坏版本

```python
x = torch.tensor([1.0, 2.0], requires_grad=True)
y = x * 2
# x.add_(1)  # 可能报错或梯度错误 — 避免对 leaf in-place
```

### 9.2 非 leaf 要 grad 需 `retain_grad()`

```python
x = torch.tensor(1.0, requires_grad=True)
y = x * 2
y.retain_grad()
y.backward()
print(y.grad)
```

### 9.3 整数 tensor 不能求导

```python
idx = torch.tensor([0, 1], dtype=torch.long)
# idx.requires_grad = True  # 报错
```

---

## 10. 与 nn.Parameter

```python
import torch.nn as nn

layer = nn.Linear(3, 1)
for name, p in layer.named_parameters():
    print(name, p.shape, p.requires_grad)
```

`nn.Parameter` = 自动 `requires_grad=True` 的 leaf，optimizer 注册这些参数。

---

## 11. 高阶：梯度检查（了解）

```python
def grad_check(fn, x, eps=1e-3):
    x = x.detach().clone().requires_grad_(True)
    y = fn(x)
    y.backward()
    anal = x.grad.clone()
    num = torch.zeros_like(x)
    for i in range(x.numel()):
        xi = x.detach().clone().view(-1)
        xi[i] += eps
        fp = fn(xi.view_as(x)).item()
        xi[i] -= 2 * eps
        fm = fn(xi.view_as(x)).item()
        num.view(-1)[i] = (fp - fm) / (2 * eps)
    return torch.max((anal - num).abs()).item()
```

自定义 autograd Function 时用；日常训练很少手调。

---

## 12. 练习

1. 对 `L = (wx+b-y)^2` 手推 ∂L/∂w、∂L/∂b，与 PyTorch 对比。
2. 画 `a→b=a+1→c=b*2→L=c.sum()` 的计算图（Mermaid 或纸笔）。
3. 解释为何 eval 阶段要用 `torch.no_grad()`。
4. 演示忘记 `zero_grad` 时梯度累加现象。
5. 用 `detach` 实现：只更新 generator，不更新 discriminator 的 stub（概念即可）。

---

## 13. 学完标准

- [ ] 闭卷描述 backward 链式法则流程
- [ ] 区分 leaf / non-leaf、`grad_fn` 含义
- [ ] 正确使用 `zero_grad`、`detach`、`no_grad`
- [ ] 解释 loss 必须是标量（或提供 grad_weights）
- [ ] 知道 in-place 对 autograd 的风险

---

## 14. FAQ

**Q1：.grad 存在哪里？**  
与 parameter 同 device 的 tensor；初始 None，第一次 backward 后分配。

**Q2：为什么 loss 要 `.item()`？**  
取 Python 标量打日志，避免无意保留计算图占显存。

**Q3：backward 能传非标量吗？**  
可以，需 `grad_tensors` 与 y 同 shape；训练 loss 设计为标量最简。

**Q4：冻结层怎么实现？**  
`for p in layer.parameters(): p.requires_grad = False` 或 `no_grad` 包 forward。

**Q5：`torch.set_grad_enabled(False)`？**  
全局开关，等价于外层 `no_grad` 语境。

**Q6：二阶导？**  
`create_graph=True`；meta-learning 用，入门少见。

**Q7：梯度爆炸/消失与本章关系？**  
图太深导致梯度乘积极端；07 章 LR、08 章 AMP、架构（残差）缓解。

**Q8：`inference_mode` 和 `no_grad` 选哪个？**  
纯推理用 `inference_mode`；训练中嵌套验证用 `no_grad` 更灵活。

**Q9：CPU 上 autograd 行为与 GPU 一致吗？**  
一致；仅 device 不同。

**Q10：与 TensorFlow 静态图区别？**  
PyTorch **动态图**（define-by-run）；2.x 也有 `torch.compile` 融合（26 章预告）。

---

## 15. 闭卷自测

1. leaf tensor 定义？
2. `loss.backward()` 后谁一定有 `.grad`（若 requires_grad=True）？
3. 为何每 step 要 `optimizer.zero_grad()`？
4. `detach()` 后原 tensor 还能 backward 吗？
5. `y.grad_fn` 表示什么？
6. 整数 label tensor 为何不能 requires_grad？
7. eval 时为何省显存（与 autograd 关系）？
8. `retain_graph=True` 何时需要？
9. 非 leaf 默认为何不存 grad？
10. `nn.Parameter` 与普通 tensor 区别？

<details>
<summary>参考答案</summary>

1. 用户直接创建且 requires_grad=True 的 tensor（如 Parameter）。
2. 参与运算的 leaf 参数（如 w）；非 leaf 默认不存。
3. backward 梯度累加；不清零则 step 用到多步之和。
4. 能；detach 创建新 tensor 不参与原图。
5. 产生 y 的 backward 函数节点，用于链式传播。
6. 离散索引无连续导数定义。
7. no_grad 不建图，不存中间 activations 用于求导。
8. 同一图上多次 backward 且未释放图时。
9. 省内存；通常只需 param.grad。
10. Parameter 自动 requires_grad，注册到 Module，optimizer 更新对象。

</details>

---

## 16. 下一章预告

05 章组装 **nn.Module、loss、optimizer** 成完整训练循环——本章 autograd 的工程落地。

---

*上一章：[03 PyTorch 张量](03-PyTorch入门与张量操作.md)*  
*下一章：[05 nn.Module 与训练循环](05-nn.Module与训练循环.md)*
