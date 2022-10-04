package GoCache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	// 在这个测试用例中，我们借助GetterFunc的类型转换，将一个匿名回调函数转成了接口f Getter。
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		// []byte(key)是将key转换为字节数组。
		return []byte(key), nil
	})

	expect := []byte("key")
	// 调用该接口的方法 f.Get(key string)，实际上在调用匿名回调函数。
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

// 模拟本地的数据库缓存。
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	// 记录表：统计某个key调用回调函数的次数。
	// 如果代码正确，一个key在Cache中找不到数据时，只会调用一次回调函数从本地获取数据，然后populate到Cache中。
	loadCounts := make(map[string]int, len(db))

	group := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			// 回调函数，从本地数据库db中查找。
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				// 成功从本地数据库db查找到数据。
				if _, ok := loadCounts[key]; !ok {
					// 如果记录表loadCounts中没有记录过这个key。
					loadCounts[key] = 0
				}
				// 该key的回调次数+1。
				loadCounts[key] += 1
				// 返回字节数组。
				return []byte(v), nil
			}
			// 使用回调函数从本地数据库找不到。
			return nil, fmt.Errorf("%s not exist", key)
		}))

	// 遍历本地数据库db。
	for k, v := range db {
		// 从分布式缓存Cache中获取value，如果未命中缓存，则调用回调函数从本地获取数据，并将数据添加到Cache中。
		if view, err := group.Get(k); err != nil || view.String() != v {
			// 如果出现异常或者获取到的值value和本地数据库存储的值value不同。
			t.Fatal("failed to get value of Tom")
		}

		// 再次获取，如果一个key不存在Cache中，通过上一步的调用，查询本地缓存后，就应该存在于Cache中，且loadCounts记录了该key。[key][1]
		if _, err := group.Get(k); err != nil || loadCounts[k] > 1 {
			// 如果出现异常或者一个key调用回调函数大于1次。
			t.Fatalf("cache %s miss", k)
		}
		// cache hit
	}

	if view, err := group.Get("unknown"); err == nil {
		// 获取一个不存在的值，如果获取成功了。
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
