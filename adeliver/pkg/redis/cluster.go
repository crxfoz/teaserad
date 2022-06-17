package redis

import "github.com/go-redis/redis/v8"

type Hasher interface {
	Hash(int) int
}

type Cluster struct {
	conn   []*redis.Client
	hasher Hasher
}

func New(conn []*redis.Client) *Cluster {
	hasher := &hash{nodes: len(conn)}
	return &Cluster{conn: conn, hasher: hasher}
}

func (c *Cluster) Node(key int) *redis.Client {
	return c.conn[c.hasher.Hash(key)]
}

func (c *Cluster) Nodes() []*redis.Client {
	return c.conn
}
