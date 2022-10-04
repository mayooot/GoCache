package gocache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBashPath = "/_gocache/"

// HTTPPool 以http://example.com/_gocache/开头的请求，用于节点间的访问。
type HTTPPool struct {
	self     string // 记录自己的地址，包括主机名/ip和端口。
	bashPath string // 节点间通讯地址的前缀。默认是/_gocache/。
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

// ServeHTTP 负责处理所有的http请求。
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

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应体为二进制流。
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}
