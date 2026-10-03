package mutex

import (
	"primitives/internal/futex"
	"sync/atomic"
)

const (
	free = iota
	held
	contended
	spinIterations = 100
)

type Mutex struct {
	state uint32
}

func (m *Mutex) Lock() {
	// это типо быстрый
	if m.TryLock() {
		return
	}
	// это типо медлнный
	for i := 0; i < spinIterations; i++ {
		if atomic.LoadUint32(&m.state) == free && m.TryLock() {
			return
		}
	}
	for {
		if atomic.SwapUint32(&m.state, contended) == free {
			return
		}

		futex.Wait(&m.state, contended)
	}
}

func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapUint32(&m.state, free, held)
}

func (m *Mutex) Unlock() {
	old := atomic.SwapUint32(&m.state, free)
	if old == free {
		panic("паника открыли уже открытый")
	}
	if old == contended {
		futex.Wake(&m.state)
	}
}
