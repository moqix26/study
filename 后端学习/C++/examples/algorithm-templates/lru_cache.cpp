// LeetCode 146 风格 LRU — 单文件自测
// g++ -std=c++17 -Wall -O2 lru_cache.cpp -o lru_test && ./lru_test

#include <iostream>
#include <list>
#include <unordered_map>

class LRUCache {
public:
    explicit LRUCache(int capacity) : cap_(capacity) {}

    int get(int key) {
        auto it = map_.find(key);
        if (it == map_.end()) return -1;
        touch(it);
        return it->second->second;
    }

    void put(int key, int value) {
        auto it = map_.find(key);
        if (it != map_.end()) {
            it->second->second = value;
            touch(it);
            return;
        }
        if (static_cast<int>(list_.size()) >= cap_) {
            int old_key = list_.back().first;
            list_.pop_back();
            map_.erase(old_key);
        }
        list_.emplace_front(key, value);
        map_[key] = list_.begin();
    }

private:
    using Node = std::pair<int, int>;
    using ListIt = std::list<Node>::iterator;

    void touch(std::unordered_map<int, ListIt>::iterator it) {
        list_.splice(list_.begin(), list_, it->second);
    }

    int cap_;
    std::list<Node> list_;
    std::unordered_map<int, ListIt> map_;
};

int main() {
    LRUCache cache(2);
    cache.put(1, 1);
    cache.put(2, 2);
    std::cout << cache.get(1) << "\n";  // 1
    cache.put(3, 3);                    // evict key 2
    std::cout << cache.get(2) << "\n";  // -1
    cache.put(4, 4);                    // evict key 1
    std::cout << cache.get(1) << "\n";  // -1
    std::cout << cache.get(3) << "\n";  // 3
    std::cout << cache.get(4) << "\n";  // 4
    return 0;
}
