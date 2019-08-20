package ttlcache

import (
	"sync"
	"time"
)

type TTLSet struct {
	m    map[string]int64
	lock sync.Mutex
	ttl  int
}

func New(maxTTL int) *TTLSet {
	return &TTLSet{m: make(map[string]int64), ttl: maxTTL}
}

func (m *TTLSet) Initalize() {
	go func() {
		for now := range time.Tick(time.Duration(m.ttl/2) * time.Millisecond) {
			m.lock.Lock()
			for k, v := range m.m {
				if now.UnixNano()-v > int64(m.ttl) {
					delete(m.m, k)
				}
			}
			m.lock.Unlock()
		}
	}()
}

func (m *TTLSet) Put(k string) {
	m.lock.Lock()
	m.m[k] = time.Now().UnixNano()
	m.lock.Unlock()
}

func (m *TTLSet) Exist(k string) bool {
	m.lock.Lock()
	_, ok := m.m[k]
	m.lock.Unlock()
	return ok

}
