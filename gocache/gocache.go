package gocache

import (
	"GoCache/gocache/singleflight"
	"fmt"
	"log"
	"sync"
)

// Group 一个Group可以认为是一个缓存的命名空间，每个Group拥有一个唯一的名称name。
// 比如可以创建三个Group，缓存学生的成绩命名为scores，缓存学生的信息命名为infos，缓存学生的课程命名为courses。
type Group struct {
	name      string
	getter    Getter // 缓存未命中时获取源数据的回调
	mainCache cache  // 并发缓存
	peers     PeerPicker
	// use singleFlight.Group to make sure that
	// each key is only fetched once
	loader    *singleflight.Group
}

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

var (
	mu     sync.RWMutex              // 读写锁
	groups = make(map[string]*Group) // 全部的缓存字典
)

// NewGroup 实例化Group并将group存储在全局变量groups中。
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		// 如果没有传入回调函数。
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			// lru中的最大缓存容量。
			cacheBytes: cacheBytes,
		},
		loader: &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup 用来返回特定名称的Group，使用了只读锁RLock()，因为不涉及到任何冲突变量的写操作。
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 用来返回key对应的value。
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		// 传入的key为空。
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		// 命中缓存。
		log.Println("[GoCache] hit")
		return v, nil
	}

	// 使用用户传递的回调函数，来加载缓存。
	return g.load(key)
}

// RegisterPeers 将实现了PeerPicker接口的HTTPPool注入到Group中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// 使用PickPeer()方法选择节点，若非本机节点，则调用getFromPeer()从远程节点获取。若是本机节点或失败，则回退到getLocally()。
func (g *Group) load(key string) (value ByteView, err error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		// 将从其他节点或从数据库中获取数据，封装进方法中。确保只会执行一次。
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GoCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// 将键值对数据添加到分布式缓存Cache中。
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// 调用开发者传递的回调函数，从本地获取数据。
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	// 拷贝一份数据，用于返回。
	value := ByteView{b: cloneBytes(bytes)}
	// 将数据添加到分布式缓存Cache中。
	g.populateCache(key, value)
	return value, nil
}

// 使用实现了PeerGetter接口的httpGetter访问远程节点，获取缓存值。
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
