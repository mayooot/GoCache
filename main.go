package main

import (
	"GoCache/gocache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// 实例化一个Group。
func createGroup() *gocache.Group {
	return gocache.NewGroup("scores", 2<<10, gocache.GetterFunc(
		// 实例化的Group的名称为scores，大小为1024字节，回调函数为调用本地DB查询数据。
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 用来启动一个API服务（端口9999），与用户进行交互，用户感知。
// 这里的apiServer和一个缓存服务绑定在一起了，所以先检查这个绑定的缓存服务器，
// 如果key不是属于这个缓存服务器再从其他peer获取，因为这里只是向外暴露了一个api,
// 所以兔老大实现的版本无法先找到key再直接从那个节点找需要的值
func startAPIServer(apiAddr string, goc *gocache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// curl "http://localhost:9999/api?key=Tom"
			// 获取请求参数，key。
			key := r.URL.Query().Get("key")
			// 调用主服务Group获取缓存。
			view, err := goc.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// 响应数据。
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("frontend server is running at", apiAddr)
	// 启动服务。
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

// 启动缓存服务器：创建HTTPPool，添加节点信息，注册到gocache中，启动HTTP服务（共3个端口，8001/8002/8003），用户不感知。
func startCacheServer(addr string, addrs []string, goc *gocache.Group) {
	peers := gocache.NewHTTPPool(addr)
	peers.Set(addrs...)
	goc.RegisterPeers(peers)
	log.Println("gocache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func main() {
	var port int
	var api bool
	// func (f *FlagSet) IntVar(p *int, name string, value int, usage string)
	// IntVar用指定的名称、默认值、使用信息注册一个int类型flag，并将flag的值保存到p指向的变量。
	flag.IntVar(&port, "port", 8001, "GoCache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	// 1. 创建一个Group，保存学生分数scores。
	goc := createGroup()

	if api {
		// 多协程启动API服务，与用户交互。
		go startAPIServer(apiAddr, goc)
	}
	// 3. 启动多个缓存服务器。
	startCacheServer(addrMap[port], []string(addrs), goc)
}
