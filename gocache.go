package GoCache

import (
	"fmt"
	"log"
	"sync"
)

/*
定义一个函数类型F，并且实现接口A的方法，然后在这个方法中调用自己。
这是Go语言将其他函数（参数返回值定义和F一致）转换为接口A的常用技巧。
*/

// Getter 定义接口Getter和回调函数 Get(key string) ([]byte, error)，参数是key，返回值是[]byte。
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 定义函数类型GetterFunc，并实现Getter接口的Get方法。
// 函数类型实现某一个接口，称之为接口型函数，方便使用者在调用时既能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数。
// 关键字 type 类型名 GetterFunc 类型 func(key string) ([]byte, error)函数。
type GetterFunc func(key string) ([]byte, error)

// Get GetterFunc类型的成员方法。
// 传入参数是string类型的key，返回值是用GetterFunc中的方法来处理key并返回字节数组。
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 一个Group可以认为是一个缓存的命名空间，每个Group拥有一个唯一的名称name。
// 比如可以创建三个Group，缓存学生的成绩命名为scores，缓存学生的信息命名为infos，缓存学生的课程命名为courses。
type Group struct {
	name      string
	getter    Getter // 缓存未命中时获取源数据的回调
	mainCache cache  // 并发缓存
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 实例化Group并将group存储在全局变量groups中。
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
		},
	}
	groups[name] = g
	return g
}

// GetGroup 用来返回特定名称的Group，使用了只读锁RLock()，因为不涉及到任何冲突变量的写操作。
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.Unlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GoCache] hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
