package lib

import (
	"context"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
)

type Transport struct {
	*http.Transport
	*net.Dialer
	connections map[string]*int32
	mu          sync.RWMutex
}

func (t *Transport) GetConnectionStats() map[string]int {
	t.mu.RLock()
	defer t.mu.RUnlock() // Keep the read lock for the entire function

	copiedMap := make(map[string]int)
	for key, value := range t.connections {
		copiedMap[key] = int(atomic.LoadInt32(value))
	}
	return copiedMap
}

func (t *Transport) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := t.Dialer.DialContext(ctx, network, address)

	if err == nil {
		t.mu.Lock()
		counter, ok := t.connections[address]
		if !ok {
			intCounter := int32(0)
			counter = &intCounter
			t.connections[address] = counter
		}
		t.mu.Unlock()
		atomic.AddInt32(counter, 1)
		return &Conn{Conn: conn, counter: counter}, err
	}

	return conn, err
}

type Conn struct {
	net.Conn
	counter *int32
	sync.Once
}

func (c *Conn) Close() (err error) {
	err = c.Conn.Close()
	c.Do(func() { atomic.AddInt32(c.counter, -1) })

	return
}

func GetConnectionStatsFromClient(client *http.Client) (map[string]int, bool) {
	transport, ok := client.Transport.(*Transport)
	if !ok {
		return nil, false
	}
	return transport.GetConnectionStats(), true
}
