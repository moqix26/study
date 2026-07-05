# 高性能 C++：对齐、零拷贝与 SIMD 入门

> **文件编码**：UTF-8。  
> **前置**：[C++ 02 指针](../C++/02-指针引用与内存管理.md)、[04 CUDA 内存](04-CUDA核函数线程层次与内存模型.md)、[05 cuBLAS](05-矩阵运算cuBLAS与GEMM优化入门.md)。  
> **定位**：推理引擎 Host 侧瓶颈——权重 mmap、Pinned Memory、对齐分配、AVX 向量加；衔接 [C++ 18](../C++/18-高性能C++与内存对齐.md)。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**高性能 C++** = 少拷贝、对齐访问、SIMD 批处理——让 CPU 喂饱 GPU、快速加载百 GB 权重。

### 0.2 你需要提前知道什么

- C++ 指针、RAII（[C++ 07](../C++/07-异常处理与RAII.md)）
- CUDA H2D 拷贝（03 章）

### 0.3 本章知识地图（☐→☑）

- [ ] 解释 `alignas(64)` 与 false sharing
- [ ] 使用 `cudaHostAlloc` 分配 Pinned Memory
- [ ] 写一版 AVX2 float 向量加并验证结果
- [ ] 说明 mmap 加载权重流程（预告 12 章）
- [ ] 完成 §12 闭卷自测 ≥8/10

### 0.4 建议学习时长

- **4～6 天**（与 C++ 18 并行）

---

## 1. 这份文档学什么

- 内存对齐与 cache line
- False sharing 与 padding
- Pinned / Mapped memory 与异步 H2D
- mmap 只读权重映射
- SIMD（AVX2）向量运算入门
- `std::span` 与非 owning 视图（C++20）

---

## 2. 对齐与 Cache Line

现代 CPU cache line 常 **64 字节**。

```cpp
struct alignas(64) Counter {
    std::atomic<int64_t> value;
};
// 避免多线程 false sharing：每个 counter 独占 cache line
```

**False sharing**：两线程写同一 cache line 不同变量 → 缓存行乒乓。

与 [C++ 08 多线程](../C++/08-多线程与并发编程.md) 结合阅读。

---

## 3. 对齐分配

```cpp
#include <cstdlib>

void* aligned_alloc(size_t alignment, size_t size) {
#if defined(_MSC_VER)
    return _aligned_malloc(size, alignment);
#else
    void* p = nullptr;
    if (posix_memalign(&p, alignment, size) != 0) return nullptr;
    return p;
#endif
}

// C++17
// std::aligned_alloc(64, size);  // size 须为 alignment 倍数
```

GEMM/ SIMD 常要求 **32/64 字节对齐**。

---

## 4. Pinned Memory（Page-Locked）

```cpp
float* h_pinned = nullptr;
cudaHostAlloc(&h_pinned, bytes, cudaHostAllocDefault);
// 或 cudaHostAllocMapped 用于 zero-copy（慎用，一般不如显式 H2D）

cudaMemcpyAsync(d_dev, h_pinned, bytes, cudaMemcpyHostToDevice, stream);
```

**效果**：DMA 直接访问；**异步** `cudaMemcpyAsync` 需 Pinned 源/宿。

**预期**：Pinned H2D 可达 Pageable 的 **2～3×** 带宽（视平台）。

---

## 5. mmap 权重加载（概念）

```cpp
#include <sys/mman.h>
#include <fcntl.h>
#include <unistd.h>

// 简化：只读映射 safetensors/gguf 文件
int fd = open("model.bin", O_RDONLY);
off_t sz = lseek(fd, 0, SEEK_END);
void* ptr = mmap(nullptr, sz, PROT_READ, MAP_PRIVATE, fd, 0);
// 解析 header → tensor offset → 可选 cudaMemcpy 到 GPU
// munmap(ptr, sz); close(fd);
```

**零拷贝**：进程虚拟地址直接映射文件页；按需 page fault 加载。12 章展开格式解析。

Windows 对应：`CreateFileMapping` + `MapViewOfFile`。

---

## 6. SIMD：AVX2 向量加

```cpp
#include <immintrin.h>
#include <cstddef>

void add_avx2(const float* a, const float* b, float* c, size_t n) {
    size_t i = 0;
    for (; i + 8 <= n; i += 8) {
        __m256 va = _mm256_loadu_ps(a + i);
        __m256 vb = _mm256_loadu_ps(b + i);
        __m256 vc = _mm256_add_ps(va, vb);
        _mm256_storeu_ps(c + i, vc);
    }
    for (; i < n; ++i) c[i] = a[i] + b[i];
}
```

编译：

```bash
g++ -O3 -mavx2 -std=c++17 -o simd_add simd_add.cpp
./simd_add
# 预期：与 scalar 版 max error = 0
```

**Infra**：CPU 预处理、小算子、embedding lookup 可向量化；大 GEMM 交给 cuBLAS。

---

## 7. std::span 与非 owning 缓冲

```cpp
#include <span>
#include <vector>

void process(std::span<const float> weights) {
    // 可读 weights.size()，不拷贝
}

std::vector<float> w(1024);
process(w);
// mmap 后：process({static_cast<const float*>(ptr)+off, count});
```

避免推理引擎中 **重复分配** 与所有权混乱。见 [C++ 05 现代特性](../C++/05-现代C++新特性.md)。

---

## 8. 零拷贝语义层次

| 层次 | 含义 |
|------|------|
| mmap | 文件→虚拟内存，无 read 缓冲拷贝 |
| Pinned + Async | 减少 staging 缓冲 |
| GPU Direct / RDMA | 网卡→GPU，分布式推理 |
| Unified Memory | 自动迁移，延迟难控 |

Serving 路径优先：**mmap + 按需 cudaMemcpyAsync**。

---

## 9. 数据布局与 SIMD

Row-major 连续 float 最适合 `_mm256_loadu_ps`；结构体数组（AoS）向量化难，常改 **SoA**（08 章 KV block 布局）。

---

## 10. 练习建议

1. 对比 `new float[n]` vs `cudaHostAlloc` H2D 带宽（`cudaEvent` 计时）
2. 实现 scalar vs AVX2 add，n=1<<24，比 wall time
3. 读 llama.cpp `ggml.c` 中 `ggml_tensor` 对齐字段
4. 用 `alignas(64)` 结构体数组测 false sharing（两线程各写不同元素）

---

## 11. 学完标准

- [ ] 解释 64 字节对齐与 cache line 关系
- [ ] 写出 Pinned Memory 分配与 async 拷贝三行
- [ ] 口述 mmap 加载权重优缺点
- [ ] 编译运行 AVX2 示例
- [ ] 说明何时 CPU SIMD 值得做、何时交给 GPU

---

## 12. FAQ

**Q1：所有 Host 缓冲都要 Pinned 吗？**  
否；仅高频 H2D/D2H 路径；Pinned 占物理内存。

**Q2：Mapped memory 零拷贝为何少用？**  
GPU 访问 UVM 可能慢；可控性不如显式拷贝。

**Q3：MSVC 没有 posix_memalign？**  
用 `_aligned_malloc` / `_aligned_free`。

**Q4：SIMD 与 CUDA 关系？**  
Host 预处理 vs Device 计算；llama.cpp CPU 路径大量 SIMD。

**Q5：权重为何能 mmap？**  
推理只读；MAP_PRIVATE 写时复制。

**Q6：safetensors 优势？**  
Header JSON 指 offset，mmap 友好；无 pickle 风险。

**Q7：std::vector 对齐吗？**  
不保证 64 对齐；大缓冲用 aligned allocator。

**Q8：移动语义如何帮助零拷贝？**  
`std::vector` move 转移所有权；见 [C++ 05](../C++/05-现代C++新特性.md)。

**Q9：推理引擎 JSON 配置算零拷贝吗？**  
不算；指数据面少拷贝。

**Q10：ARM NEON 要学吗？**  
Apple/ARM CPU 推理岗需要；NVIDIA 主线 AVX2/AVX512。

---

## 13. 闭卷自测

1. Cache line 典型大小？
2. False sharing 定义？
3. Pinned memory 主要加速哪类操作？
4. mmap PROT_READ 含义？
5. AVX2 一次处理几个 float？
6. `std::span` 是否拥有内存？
7. SoA vs AoS 对 SIMD 影响？
8. cudaMemcpyAsync 为何要 Pinned？
9. 对齐到 64 字节常见原因？
10. llama.cpp 常用哪种权重格式？

<details>
<summary>参考答案</summary>

1. 64 字节（x86 常见）。
2. 多核写同一 cache line 不同变量导致无效化开销。
3. Host↔Device DMA 传输。
4. 映射为只读虚拟页。
5. 8 个（256 bit / 32 bit）。
6. 不拥有，仅视图。
7. SoA 连续同字段利于 SIMD；AoS 常需 gather。
8. Pageable 内存 DMA 需先 pin 到 staging。
9. 避免跨 cache line、满足 SIMD load 对齐优化。
10. GGUF 等；亦支持 safetensors 类。

</details>

---

## 14. 下一章预告

06 章解决了 **数据如何高效到达 GPU**——**推理引擎如何把 GEMM、Attention、KV 组织成一条流水线？** 07 章进入大模型推理引擎架构总览。

---

*下一章：[07 大模型推理引擎架构概览](07-大模型推理引擎架构概览.md)*
