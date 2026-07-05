"""
LoRA 微调脚本（peft + HuggingFace Trainer）。

对应章节：15 微调 SFT 与 LoRA、29 TRL/SFTTrainer 实战。
支持 --model_id 与 --data_path（json/jsonl，instruction+output 或 messages）。
"""

from __future__ import annotations

import argparse
import json
from pathlib import Path

import torch
from datasets import Dataset
from peft import LoraConfig, TaskType, get_peft_model
from transformers import (
    AutoModelForCausalLM, AutoTokenizer,
    DataCollatorForLanguageModeling, Trainer, TrainingArguments,
)


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description="LoRA 微调小型因果语言模型")
    p.add_argument("--model_id", default="distilgpt2", help="HF 模型 ID 或本地路径")
    p.add_argument("--data_path", required=True, help="训练数据 .json / .jsonl")
    p.add_argument("--output_dir", default="./lora-output")
    p.add_argument("--max_length", type=int, default=512)
    p.add_argument("--epochs", type=int, default=1)
    p.add_argument("--batch_size", type=int, default=2)
    p.add_argument("--lr", type=float, default=2e-4)
    p.add_argument("--lora_r", type=int, default=8)
    p.add_argument("--lora_alpha", type=int, default=16)
    p.add_argument("--use_wandb", action="store_true")
    return p.parse_args()


def load_records(path_str: str) -> list[dict]:
    """加载 json 数组或 jsonl。"""
    path = Path(path_str)
    if not path.exists():
        raise FileNotFoundError(path_str)
    if path.suffix == ".jsonl":
        return [json.loads(l) for l in path.read_text(encoding="utf-8").splitlines() if l.strip()]
    data = json.loads(path.read_text(encoding="utf-8"))
    return data if isinstance(data, list) else [data]


def to_text(rec: dict) -> str:
    """messages 或 instruction/output → 训练文本。"""
    if "messages" in rec:
        return "\n".join(f"{m.get('role','user')}: {m.get('content','')}" for m in rec["messages"]) + "\n"
    ins = rec.get("instruction", rec.get("prompt", ""))
    out = rec.get("output", rec.get("response", ""))
    return f"### Instruction:\n{ins}\n\n### Response:\n{out}\n"


def build_dataset(tokenizer, records: list[dict], max_length: int) -> Dataset:
    texts = [to_text(r) for r in records]
    ds = Dataset.from_dict({"text": texts})
    return ds.map(
        lambda b: tokenizer(b["text"], truncation=True, max_length=max_length, padding="max_length"),
        batched=True, remove_columns=["text"],
    )


def main() -> None:
    args = parse_args()
    if args.use_wandb:
        import wandb
        wandb.init(project="llm-python-sft", config=vars(args))

    records = load_records(args.data_path)
    print(f"样本数: {len(records)}")

    tok = AutoTokenizer.from_pretrained(args.model_id)
    if tok.pad_token is None:
        tok.pad_token = tok.eos_token

    dtype = torch.bfloat16 if torch.cuda.is_available() else torch.float32
    model = AutoModelForCausalLM.from_pretrained(
        args.model_id, torch_dtype=dtype,
        device_map="auto" if torch.cuda.is_available() else None,
    )
    targets = ["c_attn"] if "gpt2" in args.model_id.lower() else ["q_proj", "v_proj"]
    model = get_peft_model(model, LoraConfig(
        task_type=TaskType.CAUSAL_LM, r=args.lora_r, lora_alpha=args.lora_alpha,
        lora_dropout=0.05, target_modules=targets, bias="none",
    ))
    model.print_trainable_parameters()

    trainer = Trainer(
        model=model,
        args=TrainingArguments(
            output_dir=args.output_dir, num_train_epochs=args.epochs,
            per_device_train_batch_size=args.batch_size, learning_rate=args.lr,
            logging_steps=10, save_strategy="epoch", bf16=torch.cuda.is_available(),
            report_to="wandb" if args.use_wandb else "none", remove_unused_columns=False,
        ),
        train_dataset=build_dataset(tok, records, args.max_length),
        data_collator=DataCollatorForLanguageModeling(tok, mlm=False),
    )
    trainer.train()
    trainer.save_model(args.output_dir)
    tok.save_pretrained(args.output_dir)
    print(f"已保存: {args.output_dir}")


if __name__ == "__main__":
    main()
