package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 定义函数类型Hash，采取依赖注入的方式，允许用于替换成自定义的哈希函数，默认为crc.ChecksumIEEE。
type Hash func(data []byte) uint32

// Map 一致性哈希的主数据结构。
type Map struct {
	hash     Hash
	replicas int            // 虚拟节点倍数。一个真实节点对应replicas个虚拟节点。
	keys     []int          // 排序后的哈希环。
	hashMap  map[int]string // 虚拟节点与真实节点的映射表，键是虚拟节点的哈希值，值是真实节点的名称。
}

// New 实例化Map并返回。允许自定义虚拟节点倍数和哈希函数。
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		// 如果传入的自定义哈希函数为空，则默认使用crc32.ChecksumIEEE算法。
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加0个或者多个真实节点的名称。
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 每添加一个真实节点，生成replicas个虚拟节点。
		for i := 0; i < m.replicas; i++ {
			// 使用m.hash()计算虚拟节点的哈希值。
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 哈希环中加入虚拟节点。
			m.keys = append(m.keys, hash)
			// 建立虚拟节点和真实节点的对应关系。
			m.hashMap[hash] = key
		}
	}
	// 哈希环排序。
	sort.Ints(m.keys)
}

// Get 获取该key所归属的真实节点。
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		// 如果哈希环为空。
		return ""
	}

	// 使用m.hash()计算虚拟节点的哈希值。
	hash := int(m.hash([]byte(key)))
	// sort.Search使用二分法搜索到[0,n)区间内最小的满足f(i) == true的值i。如果找不到返回n。
	// 顺时针找到第一个匹配的虚拟节点的下标idx。
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// idx可能为n，m.keys[n]会越界，所以使用idx % len(m.keys)来解决。
	// 如果idx == len(m.keys)，说明应该选择m.keys[0]。
	// 通过hashMap映射得到真实的节点。
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
