package gocache

import (
	"GoCache/gocache/consistenthash"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBashPath = "/_gocache/"
	defaultReplicas = 50
)

// HTTPPool 以http://example.com/_gocache/开头的请求，用于节点间的访问。
type HTTPPool struct {
	self        string                 // 记录自己的地址，包括主机名/ip和端口。
	bashPath    string                 // 节点间通讯地址的前缀。默认是/_gocache/
	mu          sync.Mutex             // 互斥锁。
	peers       *consistenthash.Map    // 一致性哈希算法的Map，用来根据具体的key来选择节点。
	httpGetters map[string]*httpGetter // 映射远程节点和对应的httpGetter。每一个远程节点对应一个httpGetter。因为httpGetter与远程节点的地址baseURL有关。
}

// NewHTTPPool = HTTPPool的实例化方法。
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		bashPath: defaultBashPath,
	}
}

// Log 打印日志信息，format：请求的方法如GET、POST等，v：任意类型的多个参数信息。
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 负责处理所有的http请求。实现了ServeHTTP(ResponseWriter, *Request)方法有，因此是实现了Handler的实例。
// 所以可以作为func ListenAndServe(addr string, handler Handler)中的handler参数。
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.bashPath) {
		// 如果请求的URL不是以basePath（"/_gocache/"）开头。
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// Path[len(p.basePath):] 代表请求的URL截去basePath的部分。SplitN代表将剩下的部分，按照"/"分割成两部分。
	parts := strings.SplitN(r.URL.Path[len(p.bashPath):], "/", 2)
	if len(parts) != 2 {
		// 如果请求的URL为 "/_gocache/"，那么len(parts)为0；如果为"/_gocache/first"，那么那么len(parts)为1。
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	// 获取到特点名称的group。
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group"+groupName, http.StatusNotFound)
		return
	}

	// 在group中获取value。
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应体为二进制流。
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

// Set 实例化一致性哈希算法，并且传入新的节点。并未每个节点创建一个HTTP客户端httpGetter。
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// 初始化一致性哈希算法结构体。
	p.peers = consistenthash.New(defaultReplicas, nil)
	// 加入多个真实节点。
	p.peers.Add(peers...)
	// 初始化httpGetters
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.bashPath}
	}
}

// PickPeer 包装了一致性算法的Get()方法，根据具体的key，选择节点，返回节点对应的HTTP客户端。
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

// 实现PeerGetter接口。
type httpGetter struct {
	// 表示将要访问的远程节点的地址，例如http://example.com/_gocache/
	baseURL string
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		// QueryEscape函数对参数进行转码使之可以安全的用在URL查询里。
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	log.Println("here: ",u)

	// 使用http.Get()方式获取返回值，并转换为[]byte类型。
	// http.Get函数返回值是 *Response和error。
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	// Response中的Body是ReaderCloser类型（Reader and Closer）。
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	// func ReadAll(r Reader) ([]byte, error)
	// ReadAll()函数接收一个Reader，返回[]byte。
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)
