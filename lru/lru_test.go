package lru

import (
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	// 传递的maxBytes为0，所以默认Cache最大容量无限。
	lru := New(int64(0), nil)
	lru.Add("k1", String("Hello World"))

	if v, ok := lru.Get("k1"); !ok || string(v.(String)) != "Hello World" {
		// 如果Cache中获取不到k1或者获取的value不正确。
		t.Fatalf("cache hit {key: k1, value: Hello World} failed")
	}

	if _, ok := lru.Get("k2"); ok {
		// 如果Cache获取到了不存在的k2。
		t.Fatalf("cache miss k2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"

	// Cache的容量仅能存下k1，k2键值对。
	capacity := len(k1 + k2 + v1 + v2)
	lru := New(int64(capacity), nil)

	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	// 加入k3键值对时，会淘汰最先加入的k1键值对。
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("k1"); ok || lru.Len() != 2 {
		// 如果Cache还能获取到被淘汰的k1，或者Cache的大小不为2。代表着淘汰功能执行错误。
		t.Fatalf("RemoveOldest k1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	// 注册回调函数，将被淘汰的键值对的key添加到splice中。
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}

	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	// 期望结果。
	expect := []string{"key1", "k2"}
	// length := len("key1")
	// t.Logf("string key1 length is %d\n", length)
	if !reflect.DeepEqual(expect, keys) {
		// 如果淘汰键值对组成的splice不等于预期。
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
