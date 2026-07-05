// 并查集 — LeetCode 200/207 基础
// g++ -std=c++17 -Wall -O2 union_find.cpp -o uf_test && ./uf_test

#include <iostream>
#include <numeric>
#include <vector>

class UnionFind {
public:
    explicit UnionFind(int n) : parent_(n), rank_(n, 0) {
        std::iota(parent_.begin(), parent_.end(), 0);
    }

    int find(int x) {
        if (parent_[x] != x) {
            parent_[x] = find(parent_[x]);
        }
        return parent_[x];
    }

    bool unite(int a, int b) {
        int ra = find(a), rb = find(b);
        if (ra == rb) return false;
        if (rank_[ra] < rank_[rb]) std::swap(ra, rb);
        parent_[rb] = ra;
        if (rank_[ra] == rank_[rb]) ++rank_[ra];
        return true;
    }

private:
    std::vector<int> parent_;
    std::vector<int> rank_;
};

int main() {
    UnionFind uf(5);
    uf.unite(0, 1);
    uf.unite(1, 2);
    std::cout << (uf.find(0) == uf.find(2) ? "connected" : "not") << "\n";
    std::cout << (uf.find(0) == uf.find(3) ? "connected" : "not") << "\n";
    return 0;
}
