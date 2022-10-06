package singleflight

import "sync"

// call 代表正在进行中，或已经结束的请求。使用sync.WaitGroup锁避免重入。
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group singleFlight的主数据结构，管理不同的key的请求(call)。
type Group struct {
	mu sync.Mutex       // 保护Group的成员变量m不被并发读写而加上的锁。
	m  map[string]*call // 保存key和call的映射关系
}

// Do 作用是针对相同的key，无论Do被调用多少次，函数fn都只会被调用一次。等待fn调用结束，返回返回值或者错误。
// 并发协程之间不需要消息传递，非常适合sync.WaitGroup。
// wg.Add(1)锁加1。 	wg.wait()阻塞，直到锁被释放。	wg.Done()锁减1。
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock() // 加锁，防止并发读写Group.m。
	if g.m == nil {
		// 延迟初始化，提高内存使用效率。
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		// 有其他请求正在获取该key，等待。
		g.mu.Unlock()       // 释放锁，让其他请求进入Do方法。
		c.wg.Wait()         // 如果有请求正在进行中，则等待。
		return c.val, c.err // 请求结束，返回结果。
	} // 没有其他请求在进行。
	c := new(call) // 初始化call。
	c.wg.Add(1)    // 准备发起请求，发起请求前加锁，锁+1。
	g.m[key] = c   // 添加到g.m中，表明key已经有对应的请求在处理。对应上面的if判断。
	g.mu.Unlock()  // 释放锁，让其他请求进入Do方法。

	c.val, c.err = fn() // 调用fn，发起请求。
	c.wg.Done()         // 请求结束。锁-1。

	g.mu.Lock()
	delete(g.m, key) // 更新g.m。
	g.mu.Unlock()

	return c.val, c.err // 返回结果
}
