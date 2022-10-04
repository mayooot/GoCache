package lru

import (
	"container/list"
)

// Cache 对外暴露的缓存对象。LRU缓存，在并发访问下线程不安全。
type Cache struct {
	// 允许使用的最大内存。默认值0代表不设置内存大小。
	maxBytes int64
	// 当前已经使用的内存。
	useBytes int64
	// Go语言标准库实现的双向链表。
	ll *list.List
	// 字典的定义，key是字符串，value是双向链表中对应节点的指针。
	// Element结构体定义 next, prev *Element	list *List	Value any
	cache map[string]*list.Element
	// 某条记录被移除时的回调函数，可以为nil。
	OnEvicted func(key string, value Value)
}

// 键值对entry是双向链表中节点的数据类型。在链表中保存每个值对应的key的好处在于淘汰队首节点时，需要用key从字典中删除对应的映射。
// 因为链表中存入的只有值，那么删除/更新值的时候，无法得知值对应的键，也就无法在字典中删除/更新对应的记录。
type entry struct {
	key   string
	value Value
}

// Value 为了通用性，值是实现了Value接口的任意类型，该接口只包含一个方法Len() int，用于返回值所占用的内存大小。
type Value interface {
	Len() int
}

// New 实例化函数。需要传递最大内存容量和回调函数。
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get 查找功能，分为两步，从字典中找出对应的双向链表的节点，将该节点移动到队尾。
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// 如果键对应的链表节点存在，则将对应节点移动到队尾，并返回查找到的值。
		// MoveToFront将链表中的节点ele移动到队尾（双向链表作为队列，队首和队尾是相对的，在这里约定front为队尾）。
		c.ll.MoveToFront(ele)
		// 双向链表存储的value为any类型（interface{}）type any = interface{}，这里需要强制转换成entry类型。
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 删除功能，实际上缓存淘汰。移除最近最少访问的节点（队首）。
func (c *Cache) RemoveOldest() {
	// Back函数，返回的是队头。因为我们规定了front为队尾，back为队头。
	ele := c.ll.Back()
	if ele != nil {
		// 从双向链表中删除节点。
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// 从字典中删除该条记录。
		delete(c.cache, kv.key)
		// 更新当所用的内存。
		c.useBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// 回调函数不为nil，调用回调函数。
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 新增/更新功能。
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; !ok {
		// 字典中不存在这条记录，新增操作。新增一个节点到队尾。
		ele := c.ll.PushFront(&entry{key, value})
		// 添加记录到字典中。
		c.cache[key] = ele
		// 更新已使用的内存。
		c.useBytes += int64(len(key)) + int64(value.Len())
	} else {
		// 字典中存在这条记录，则更新对应节点的值，并将节点移动到队尾。
		c.ll.MoveToFront(ele)
		// c.cache是一个map结构，key是字符串，value是双向链表中对应节点的指针。
		// Element中的Value为any类型（interface{}），所以需要强转为entry类型。
		kv := ele.Value.(*entry)
		// 更新已经使用内存，oldVal代表旧值所占内存，newVal代表新值所占内存。
		// oldVal < newVal，newVal - oldVal为正值。已使用内存增加。
		// oldVal > newVal，newVal - oldVal为负值。已使用内存减小。
		c.useBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	}
	// 如果添加/更新元素后，超过了最大的内存容量，循环淘汰旧值。
	for c.maxBytes != 0 && c.maxBytes < c.useBytes {
		c.RemoveOldest()
	}
}

// Len 返回容器的大小。
func (c *Cache) Len() int {
	return c.ll.Len()
}
