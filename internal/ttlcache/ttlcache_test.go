package ttlcache

import (
	"testing"
	"time"
)

func TestTTLSet(t *testing.T) {
	localIP := "127.0.0.1"
	ttl := 100
	m := New(ttl)
	m.Initalize()
	m.Put(localIP)
	if ok := m.Exist(localIP); !ok {
		t.Errorf("%s doesnt exist before ttl ends", localIP)
	}
	time.Sleep(time.Duration(m.ttl*2) * time.Millisecond)
	if ok := m.Exist(localIP); ok {
		t.Errorf("%s still exists after ttl", localIP)
	}
}
