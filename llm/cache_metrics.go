package llm

import (
	"sync/atomic"
	"time"
)

// Metrics 存储缓存的统计信息
type Metrics struct {
	hits       uint64
	misses     uint64
	adds       uint64
	evicts     uint64
	lastAccess time.Time
}

func (m *Metrics) recordHit()   { atomic.AddUint64(&m.hits, 1) }
func (m *Metrics) recordMiss()  { atomic.AddUint64(&m.misses, 1) }
func (m *Metrics) recordAdd()   { atomic.AddUint64(&m.adds, 1) }
func (m *Metrics) recordEvict() { atomic.AddUint64(&m.evicts, 1) }

// GetStats 返回当前统计信息
func (m *Metrics) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"hits":        atomic.LoadUint64(&m.hits),
		"misses":      atomic.LoadUint64(&m.misses),
		"adds":        atomic.LoadUint64(&m.adds),
		"evicts":      atomic.LoadUint64(&m.evicts),
		"last_access": m.lastAccess,
	}
}
