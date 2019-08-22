package broadcast

import (
	"sync"
	"testing"
	"time"
)

type fakeWebSocket struct {
	message []byte
	lock    *sync.Mutex
}

func (fws *fakeWebSocket) WriteMessage(mt int, m []byte) error {
	fws.lock.Lock()
	defer fws.lock.Unlock()
	fws.message = m
	return nil
}

func (fws *fakeWebSocket) Close() error {
	return nil
}

func TestWebsocketBroadcaster_Broadcast(t *testing.T) {
	m := []byte("test")
	ch := make(chan []byte)
	b := New(ch)

	b.Broadcast()
	f := fakeWebSocket{lock: &sync.Mutex{}}
	b.Add(&f)
	ch <- m
	time.Sleep(100 * time.Millisecond)
	f.lock.Lock()
	defer f.lock.Unlock()
	for i := range f.message {
		if f.message[i] != m[i] {
			t.Errorf("%q doesnt match %q on channel", string(f.message), string(m))
			return
		}
	}
}

func TestWebsocketBroadcaster_Stop(t *testing.T) {
	ch := make(chan []byte)
	b := New(ch)

	b.Broadcast()
	f := fakeWebSocket{lock: &sync.Mutex{}}
	b.Add(&f)
	s := b.Stop()
	time.Sleep(100 * time.Millisecond)
	if s != b.qchan {
		t.Errorf("recieved unexpected message back from qchan: %+v", s)
	}
}
