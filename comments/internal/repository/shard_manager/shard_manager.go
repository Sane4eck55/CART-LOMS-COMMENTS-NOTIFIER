// Package shardmanager ...
package shardmanager

import (
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spaolacci/murmur3"
)

// shard ...
type node struct {
	shardID int
	hash    uint32
	shard   *pgxpool.Pool
}

// Manager ...
type Manager struct {
	nodes    []node
	shards   []*pgxpool.Pool
	numNodes int
}

// NewShardManager ...
func NewShardManager(numNodes int) *Manager {
	return &Manager{
		numNodes: numNodes,
	}
}

// AddShard ...
func (m *Manager) AddShard(shard *pgxpool.Pool, shardID int) {
	m.shards = append(m.shards, shard)

	for i := 0; i < m.numNodes; i++ {
		hash := hashVirtualNode(shardID, i)
		m.nodes = append(m.nodes, node{hash: hash, shard: shard, shardID: shardID})
	}
	sort.Slice(m.nodes, func(i, j int) bool {
		return m.nodes[i].hash < m.nodes[j].hash
	})
}

// GetShard ...
func (m *Manager) GetShard(key string) (*pgxpool.Pool, int, error) {
	if len(m.nodes) < 1 {
		return nil, -1, errors.New("not found shards")
	}

	h := hashKey(key)

	idx := sort.Search(len(m.nodes), func(i int) bool {
		return m.nodes[i].hash >= h
	})

	if idx == len(m.nodes) {
		idx = 0
	}

	return m.nodes[idx].shard, m.nodes[idx].shardID, nil
}

// GetShardByCommentID ...
func (m *Manager) GetShardByCommentID(id int64) *pgxpool.Pool {
	idx := int(id % 10)
	return m.shards[idx-1]
}

// Close ...
func (m *Manager) Close() {
	for _, pool := range m.shards {
		pool.Close()
	}
}

// GetShardPool ...
func (m *Manager) GetShardPool() []*pgxpool.Pool {
	return m.shards
}

// hashVirtualNode ...
func hashVirtualNode(shardID, i int) uint32 {
	virtualNodeKey := fmt.Sprintf("%d-%d", shardID, i)
	hasher := murmur3.New32()
	defer hasher.Reset()
	_, _ = hasher.Write([]byte(virtualNodeKey))
	return hasher.Sum32()
}

// hashKey ...
func hashKey(key string) uint32 {
	hasher := murmur3.New32()
	defer hasher.Reset()
	_, _ = hasher.Write([]byte(key))
	return hasher.Sum32()
}
