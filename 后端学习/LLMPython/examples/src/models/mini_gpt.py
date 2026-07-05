"""
字符级 Mini-GPT 实现。

对应章节：11 Transformer 从零实现、14 预训练与语言模型原理。
包含 GPTBlock、MiniGPT、train_step 与 generate。
"""

from __future__ import annotations

import math

import torch
import torch.nn as nn
import torch.nn.functional as F


def scaled_dot_product_attention(q, k, v, causal: bool = True) -> torch.Tensor:
    """缩放点积注意力；causal 时屏蔽未来 token。"""
    scale = 1.0 / math.sqrt(q.size(-1))
    scores = (q @ k.transpose(-2, -1)) * scale
    if causal:
        t = q.size(-2)
        mask = torch.triu(torch.ones(t, t, device=q.device, dtype=torch.bool), 1)
        scores = scores.masked_fill(mask, float("-inf"))
    return F.softmax(scores, dim=-1) @ v


class CausalSelfAttention(nn.Module):
    """因果多头自注意力。"""

    def __init__(self, n_embd: int, n_head: int, dropout: float = 0.1):
        super().__init__()
        assert n_embd % n_head == 0
        self.n_head, self.head_dim = n_head, n_embd // n_head
        self.qkv = nn.Linear(n_embd, 3 * n_embd, bias=False)
        self.proj = nn.Linear(n_embd, n_embd, bias=False)
        self.dropout = nn.Dropout(dropout)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        b, t, c = x.shape
        qkv = self.qkv(x).reshape(b, t, 3, self.n_head, self.head_dim)
        q, k, v = qkv.permute(2, 0, 3, 1, 4)
        out = scaled_dot_product_attention(q, k, v).transpose(1, 2).reshape(b, t, c)
        return self.dropout(self.proj(out))


class MLP(nn.Module):
    """FFN：Linear → GELU → Linear。"""

    def __init__(self, n_embd: int, dropout: float = 0.1):
        super().__init__()
        self.net = nn.Sequential(
            nn.Linear(n_embd, 4 * n_embd), nn.GELU(),
            nn.Linear(4 * n_embd, n_embd), nn.Dropout(dropout),
        )

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        return self.net(x)


class GPTBlock(nn.Module):
    """Pre-LN Transformer Block。"""

    def __init__(self, n_embd: int, n_head: int, dropout: float = 0.1):
        super().__init__()
        self.ln1, self.ln2 = nn.LayerNorm(n_embd), nn.LayerNorm(n_embd)
        self.attn = CausalSelfAttention(n_embd, n_head, dropout)
        self.mlp = MLP(n_embd, dropout)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        x = x + self.attn(self.ln1(x))
        return x + self.mlp(self.ln2(x))


class MiniGPT(nn.Module):
    """字符级 Mini-GPT；lm_head 与 tok_emb 共享权重。"""

    def __init__(
        self, vocab_size: int, n_embd: int = 128, n_head: int = 4,
        n_layer: int = 4, block_size: int = 128, dropout: float = 0.1,
    ):
        super().__init__()
        self.block_size = block_size
        self.tok_emb = nn.Embedding(vocab_size, n_embd)
        self.pos_emb = nn.Embedding(block_size, n_embd)
        self.drop = nn.Dropout(dropout)
        self.blocks = nn.ModuleList([GPTBlock(n_embd, n_head, dropout) for _ in range(n_layer)])
        self.ln_f = nn.LayerNorm(n_embd)
        self.lm_head = nn.Linear(n_embd, vocab_size, bias=False)
        self.lm_head.weight = self.tok_emb.weight

    def forward(self, idx: torch.Tensor, targets: torch.Tensor | None = None):
        b, t = idx.shape
        pos = torch.arange(t, device=idx.device)
        x = self.drop(self.tok_emb(idx) + self.pos_emb(pos))
        for blk in self.blocks:
            x = blk(x)
        logits = self.lm_head(self.ln_f(x))
        loss = None
        if targets is not None:
            loss = F.cross_entropy(logits.view(-1, logits.size(-1)), targets.view(-1))
        return logits, loss

    @torch.no_grad()
    def generate(self, idx: torch.Tensor, max_new_tokens: int, temperature: float = 1.0) -> torch.Tensor:
        """自回归采样；temperature 越小越确定。"""
        for _ in range(max_new_tokens):
            logits, _ = self(idx[:, -self.block_size:])
            logits = logits[:, -1, :] / max(temperature, 1e-8)
            idx = torch.cat([idx, torch.multinomial(F.softmax(logits, -1), 1)], 1)
        return idx


def train_step(
    model: MiniGPT, x: torch.Tensor, y: torch.Tensor,
    optimizer: torch.optim.Optimizer, max_grad_norm: float = 1.0,
) -> float:
    """单步训练：前向 → 反向 → 梯度裁剪 → 更新。"""
    model.train()
    _, loss = model(x, y)
    optimizer.zero_grad(set_to_none=True)
    loss.backward()
    torch.nn.utils.clip_grad_norm_(model.parameters(), max_grad_norm)
    optimizer.step()
    return float(loss.item())


def make_char_batch(text: str, block_size: int, batch_size: int, device: torch.device):
    """构造字符级 CLM batch，返回 (xs, ys, stoi, itos)。"""
    chars = sorted(set(text))
    stoi = {c: i for i, c in enumerate(chars)}
    itos = {i: c for c, i in stoi.items()}
    data = torch.tensor([stoi[c] for c in text], dtype=torch.long)
    ix = torch.randint(0, len(data) - block_size, (batch_size,))
    xs = torch.stack([data[i : i + block_size] for i in ix]).to(device)
    ys = torch.stack([data[i + 1 : i + block_size + 1] for i in ix]).to(device)
    return xs, ys, stoi, itos


if __name__ == "__main__":
    demo = ("To be, or not to be, that is the question:\n" * 5)
    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    block_size, batch_size = 32, 16

    model = MiniGPT(len(set(demo)), n_embd=64, n_head=4, n_layer=2, block_size=block_size).to(device)
    opt = torch.optim.AdamW(model.parameters(), lr=3e-3)
    print(f"参数量: {sum(p.numel() for p in model.parameters()):,} | 设备: {device}")

    for step in range(200):
        x, y, stoi, itos = make_char_batch(demo, block_size, batch_size, device)
        loss = train_step(model, x, y, opt)
        if step % 50 == 0:
            print(f"step {step:3d} | loss {loss:.4f}")

    seed = torch.tensor([[stoi[demo[0]]]], device=device)
    out = model.generate(seed, 80, temperature=0.8)[0].tolist()
    print("\n生成样例:\n", "".join(itos[i] for i in out))
