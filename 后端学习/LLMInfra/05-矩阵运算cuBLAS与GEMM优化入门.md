# 矩阵运算、cuBLAS 与 GEMM 优化入门

> **文件编码**：UTF-8。  
> **前置**：[04 CUDA 内存模型](04-CUDA核函数线程层次与内存模型.md)、[01 线性代数](01-线性代数与数值计算基础.md)。  
> **定位**：LLM 算力 ~90% 在 GEMM；掌握 cuBLAS API、Row/Col-major 陷阱、Roofline 与 Tensor Core 入门。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**GEMM**（General Matrix Multiply）= `C = α·op(A)·op(B) + β·C`；**cuBLAS** = NVIDIA 高度优化的 GEMM 库，推理引擎默认走它或 CUTLASS。

### 0.2 你需要提前知道什么

- 04 章 shared memory tile 概念
- 01 章矩阵维度与 Row-major

### 0.3 本章知识地图（☐→☑）

- [ ] 调用 `cublasSgemm` 完成 C=A·B
- [ ] 解释 cuBLAS 列主序与 C++ 行主序转换
- [ ] 估算 GEMM FLOPs 与算术强度
- [ ] 画 Roofline 上 GEMM 位置
- [ ] 完成 §12 闭卷自测 ≥8/10

### 0.4 建议学习时长

- **5～8 天**

---

## 1. 这份文档学什么

- GEMM 定义与 LLM 中的出现位置
- cuBLAS 基本 API：`cublasCreate`、`cublasSgemm`
- Leading dimension `lda`
- FP16/BF16 Tensor Core：`cublasGemmEx`
- Naive vs cuBLAS vs CUTLASS 性能层次
- Roofline 模型入门

---

## 2. GEMM 在 LLM 中的角色

| 算子 | GEMM 形式 |
|------|-----------|
| Q/K/V 投影 | `[B·S, H] × [H, H]` |
| Output proj | `[B·S, H] × [H, H]` |
| FFN up/gate/down | `[B·S, H] × [H, I]` 等 |
| LM Head | `[B·S, H] × [H, V]` |

**Batch 维** 常 fold 进 M 维；**Continuous Batching** 改变有效 M（16 章）。

---

## 3. cuBLAS 列主序陷阱

cuBLAS 假设 **Fortran/Col-major**：`A_col[i,j]` 在内存中列连续。

C Row-major `A[i,j]` 等价于 Col-major 的 **A^T**。

求 Row-major 下 `C = A·B`（A: M×K, B: K×N）：

```cpp
// 等价 trick：C^T = B^T · A^T
// cublas: C = op(B) * op(A)
cublasSgemm(handle,
    CUBLAS_OP_N, CUBLAS_OP_N,
    N, M, K,
    &alpha,
    d_B, N,      // ldb = N (col-major B^T 视角)
    d_A, K,      // lda = K
    &beta,
    d_C, N);     // ldc = N
```

**面试必背**：传指针时不转置数据，用 `OP` 与维度交换表达数学含义。

---

## 4. 完整 cuBLAS 示例

```cpp
#include <cublas_v2.h>
#include <cuda_runtime.h>
#include <cstdio>

int main() {
    const int M = 512, N = 512, K = 512;
    float *d_A, *d_B, *d_C;
    cudaMalloc(&d_A, M * K * sizeof(float));
    cudaMalloc(&d_B, K * N * sizeof(float));
    cudaMalloc(&d_C, M * N * sizeof(float));

    cublasHandle_t handle;
    cublasCreate(&handle);

    float alpha = 1.f, beta = 0.f;
    cublasSgemm(handle,
        CUBLAS_OP_N, CUBLAS_OP_N,
        N, M, K,
        &alpha, d_B, N, d_A, K,
        &beta, d_C, N);

    cudaDeviceSynchronize();
    printf("cublas gemm done M=%d N=%d K=%d\n", M, N, K);

    cublasDestroy(handle);
    cudaFree(d_A); cudaFree(d_B); cudaFree(d_C);
    return 0;
}
```

### 4.1 编译运行

```bash
nvcc -O2 -o gemm_cublas gemm_cublas.cu -lcublas
./gemm_cublas
```

**预期输出**：

```text
cublas gemm done M=512 N=512 K=512
```

用 `nvidia-smi` 可见短暂 GPU 活动。

---

## 5. Leading Dimension

Col-major M×K 矩阵 **lda ≥ M**（允许子矩阵 padding）。

Row-major 转 cuBLAS 时 **ldc** 是结果矩阵 leading dim（列主序下 ≥ 行数）。

---

## 6. Tensor Core 与 `cublasGemmEx`

Ampere+ 上 FP16/BF16 输入、FP32 累加：

```cpp
cublasGemmEx(handle,
    CUBLAS_OP_N, CUBLAS_OP_N,
    N, M, K,
    &alpha,
    d_B, CUDA_R_16F, N,
    d_A, CUDA_R_16F, K,
    &beta,
    d_C, CUDA_R_32F, N,
    CUDA_R_32F,
    CUBLAS_GEMM_DEFAULT_TENSOR_OP);
```

**Infra**：推理权重 FP16、激活 FP16、累加 FP32 是常态。

---

## 7. Roofline 模型

```mermaid
xychart-beta
    title "Roofline 示意"
    x-axis "Arithmetic Intensity (FLOP/Byte)" 0 --> 100
    y-axis "GFLOPS" 0 --> 500
    line "Compute Roof" 10 --> 500
    line "Memory Roof" 0 --> 50
```

- **算术强度** AI = FLOPs / Bytes moved
- GEMM 大矩阵 AI 高 → **compute bound**；小矩阵 / decode → **memory bound**
- 优化方向：增大 M（batch）、融合、量化（09 章）

FLOPs：`2·M·K·N`（乘加各算 1 FLOP）。

---

## 8. Naive vs cuBLAS 性能实验

```bash
# 伪代码流程
# 1. cudaEvent 计时 naive tile gemm (04 章)
# 2. 计时 cublasSgemm 同形状
# 3. 打印 GFLOPS = 2MNK / time
```

**预期**（RTX 3080 量级，512³ FP32）：cuBLAS 通常为 naive 的 **10～50×**。

---

## 9. CUTLASS 简介

NVIDIA 模板 GEMM 库，vLLM/TensorRT-LLM 自定义 shape 常用。

- cuBLAS：**黑盒**、极快、通用
- CUTLASS：**可定制** tile、epilogue fusion
- 自写 naive：**学习** 用

阅读入口：`github.com/NVIDIA/cutlass` 中 `examples/00_basic_gemm`。

---

## 10. Batch GEMM

`cublasSgemmStridedBatched`：多组相同 shape GEMM 一次 launch。

Multi-Head Attention 中 batched `QK^T` 可映射为 **strided batch**（head 维）。

---

## 11. 练习建议

1. 实现 Row-major CPU GEMM，与 cuBLAS 结果对比 max error
2. 测 M=N=K=64 vs 4096 的 GFLOPS 差异，写 Roofline 解释
3. 阅读 PyTorch `torch.matmul` 文档中 cuBLAS 后端说明
4. 改 `cublasGemmEx` 为 FP16 输入，对比 FP32 吞吐

---

## 12. 学完标准

- [ ] 独立写出一次 `cublasSgemm` 调用
- [ ] 口述 Row/Col-major 转换口诀
- [ ] 计算 7B 模型单层 FFN GEMM FLOPs 数量级
- [ ] 解释 decode 阶段 GEMM 为何变小仍可能 memory bound
- [ ] 说出 Tensor Core 对 dtype 的要求

---

## 13. FAQ

**Q1：为什么 LLM 不说「卷积」？**  
Transformer 以 GEMM/Attention 为主；视觉 backbone 才有卷积。

**Q2：BLAS 和 cuBLAS？**  
BLAS 是标准接口；cuBLAS 是 NVIDIA GPU 实现。

**Q3：OpenBLAS 在 GPU 上？**  
CPU 库；GPU 用 cuBLAS/CUTLASS/rocBLAS。

**Q4：α、β 作用？**  
`C = α·AB + β·C`；β=0 纯乘；β=1 累加（融合 bias 等）。

**Q5：Strided batch 与 Batched 区别？**  
Strided：固定间隔；Batched：指针数组——shape 相同时 strided 更简单。

**Q6：LM Head 超大 vocab 怎么 GEMM？**  
Sampled softmax、分块、权重 offloading（14 章）。

**Q7：INT8 GEMM？**  
09 章；`cublasLtMatmul` 支持量化。

**Q8：如何查 cuBLAS 版本？**  
`cublasGetProperty(MAJOR_VERSION, ...)` 或链接库版本。

**Q9：M=1 的 GEMV 用什么？**  
`cublasSgemv`；decode token 常见。

**Q10：自写 kernel 何时有必要？**  
融合 epilogue、特殊 sparse、小算子——大 GEMM 别重复造轮子。

---

## 14. 闭卷自测

1. GEMM (M×K)·(K×N) 的 FLOPs？
2. cuBLAS 默认哪种 storage order？
3. Row-major C=A·B 常如何映射到 cublasSgemm？
4. lda 含义？
5. Tensor Core 典型输入 dtype？
6. Roofline 上横轴是什么？
7. 为何大矩阵 GEMM 更接近峰值 GFLOPS？
8. `cublasGemmStridedBatched` 用途？
9. CUTLASS 相对 cuBLAS 优势？
10. Decode 时 M 维典型值？

<details>
<summary>参考答案</summary>

1. 2MK N。
2. Column-major (Fortran order)。
3. 利用 C^T=B^T A^T，交换 A/B 角色与 M/N/K。
4. Leading dimension，列主序矩阵第一维 stride。
5. FP16/BF16/TF32 等（架构相关）。
6. Arithmetic Intensity（FLOP/Byte）。
7. 复用数据多，AI 高，compute bound。
8. 多组同 shape GEMM 一次提交（如 multi-head）。
9. 可融合 epilogue、自定义 tile。
10. batch_size（常为 1～几十），远小于 prefill。

</details>

---

## 15. 下一章预告

05 章在 GPU 上算得快——**Host 端如何把数据零拷贝、对齐地交给 GPU？** 06 章讲 C++ 对齐、Pinned Memory、SIMD 与 [C++ 18 高性能习惯](../C++/18-高性能C++与内存对齐.md) 衔接。

---

*下一章：[06 高性能 C++ 对齐零拷贝与 SIMD 入门](06-高性能C++对齐零拷贝与SIMD入门.md)*
