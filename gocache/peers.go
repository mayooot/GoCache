package gocache

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
type PeerPicker interface {
	// PickPeer 根据传入的key选择相应节点的PeerGetter。
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	// Get 从对应的group中查询缓存值。相当于HTTP客户端。
	Get(group string, key string) ([]byte, error)
}
