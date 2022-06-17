package tarantool

import (
	"github.com/tarantool/go-tarantool"
)

type hash struct {
	nodes int
}

func (h *hash) Hash(key int) int {
	return key % h.nodes
}

type Hasher interface {
	Hash(int) int
}

type Cluster struct {
	conn   []*tarantool.Connection
	hasher Hasher
}

func New(conn []*tarantool.Connection) *Cluster {
	hasher := &hash{nodes: len(conn)}
	return &Cluster{conn: conn, hasher: hasher}
}

func (c *Cluster) Node(key int) *tarantool.Connection {
	return c.conn[c.hasher.Hash(key)]
}

func (c *Cluster) Nodes() []*tarantool.Connection {
	return c.conn
}
