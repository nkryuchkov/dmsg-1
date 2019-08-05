package dmsg

import (
	"math/rand"
	"sync"
	"time"
)

const (
	firstEphemeralPort = 49152
	lastEphemeralPort  = 65535
)

type PortManager struct {
	mu        sync.RWMutex
	rand      *rand.Rand
	listeners map[uint16]Listener
}

func newPortManager() *PortManager {
	return &PortManager{
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
		listeners: make(map[uint16]Listener),
	}
}

func (pm *PortManager) Listener(port uint16) (Listener, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	l, ok := pm.listeners[port]
	return l, ok
}

func (pm *PortManager) AddListener(l Listener, port uint16) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.listeners[port] = l
}

func (pm *PortManager) NextEmptyEphemeralPort() uint16 {
	for {
		port := pm.randomEphemeralPort()
		if _, ok := pm.Listener(port); !ok {
			return port
		}
	}
}

func (pm *PortManager) randomEphemeralPort() uint16 {
	return uint16(firstEphemeralPort + pm.rand.Intn(lastEphemeralPort-firstEphemeralPort))
}
