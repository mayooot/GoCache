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
	// 缓存表：统计某个键调用回调函数的次数。
	loadCounts := make(map[string]int, len(db))

	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			// 回调函数，从本地数据库db中查找。
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				// 成功从本地数据库db查找到数据。
				if _, ok := loadCounts[key]; !ok {
					// 这个key没有调用过回调函数，记录这个key。
					loadCounts[key] = 0
				}
				// 否则，这个key调用回调函数次数+1
				loadCounts[key] += 1
				// 返回字节数组
				return []byte(v), nil
			}
			// 本地数据库找不到。
			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k, v := range db {
		// 遍历本地数据库db
		if view, err := gee.Get(k); err != nil || view.String() != v {
			// 使用GoGroup来获取，进行对比。
			t.Fatal("failed to get value of Tom")
		} // load from callback function
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}

}
