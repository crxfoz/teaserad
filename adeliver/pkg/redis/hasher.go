package redis

type hash struct {
	nodes int
}

func (h *hash) Hash(key int) int {
	return key % h.nodes
}
