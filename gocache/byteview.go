package gocache

// ByteView 只读数据结构，用来表示缓存值（value）。
type ByteView struct {
	// b会存储真正的缓存值。选择byte类型是为了能够支持任意的数据类型的存储，比如字符串、图片等。
	b []byte
}

// Len 返回缓存值的长度。
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回一个数据拷贝，防止缓存值被外部程序修改。
// 参考go_learning/ch37/problem_test.go
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 将数据转换为字符串并返回，必要时进行复制。
func (v ByteView) String() string {
	return string(v.b)
}

// 生成数据的副本
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
