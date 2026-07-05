"""
Perplexity 评估脚本。

对应章节：14 预训练原理、19 模型评估与 Benchmark。
PPL = exp(平均 token 负对数似然)。
"""

from __future__ import annotations

import argparse
import math
from pathlib import Path

import torch
from datasets import load_dataset
from transformers import AutoModelForCausalLM, AutoTokenizer


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description="计算因果 LM 的 Perplexity")
    p.add_argument("--model_id", default="distilgpt2", help="HF 模型 ID 或本地路径")
    p.add_argument("--text_path", default="", help="可选本地 .txt；留空用 wikitext-2 test")
    p.add_argument("--max_length", type=int, default=512)
    p.add_argument("--device", default="auto")
    return p.parse_args()


def get_device(arg: str) -> torch.device:
    if arg == "auto":
        return torch.device("cuda" if torch.cuda.is_available() else "cpu")
    return torch.device(arg)


def iter_texts(text_path: str):
    """迭代评估文本段落。"""
    if text_path:
        for para in Path(text_path).read_text(encoding="utf-8").split("\n\n"):
            if para.strip():
                yield para.strip()
        return
    for t in load_dataset("wikitext", "wikitext-2-raw-v1", split="test")["text"]:
        if t.strip():
            yield t


@torch.no_grad()
def compute_ppl(model, tokenizer, texts, device, max_length: int) -> float:
    """累计 NLL 并返回 perplexity。"""
    model.eval()
    total_nll, total_tokens = 0.0, 0
    for text in texts:
        ids = tokenizer(text, return_tensors="pt", truncation=True, max_length=max_length)["input_ids"].to(device)
        if ids.numel() < 2:
            continue
        out = model(ids, labels=ids)
        n = ids.numel()
        total_nll += out.loss.item() * n
        total_tokens += n
    if total_tokens == 0:
        raise ValueError("无有效 token")
    return math.exp(total_nll / total_tokens)


def main() -> None:
    args = parse_args()
    device = get_device(args.device)
    print(f"模型: {args.model_id} | 设备: {device}")

    tok = AutoTokenizer.from_pretrained(args.model_id)
    model = AutoModelForCausalLM.from_pretrained(args.model_id).to(device)
    texts = list(iter_texts(args.text_path))
    ppl = compute_ppl(model, tok, texts, device, args.max_length)

    src = args.text_path or "wikitext-2-raw-v1/test"
    print(f"数据源: {src} | 段落数: {len(texts)} | PPL: {ppl:.4f}")


if __name__ == "__main__":
    main()
