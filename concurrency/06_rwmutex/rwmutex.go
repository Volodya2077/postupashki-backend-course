package rwmutex

import (
	"primitives/internal/futex"
	"sync/atomic"
)

const (
	writer  = 1 << 31
	waiting = 1 << 30
	readers = waiting - 1
)

type RWMutex struct {
	state uint32
}

func (rw *RWMutex) RLock() {
	for {
		a := atomic.LoadUint32(&rw.state)
		if (a&writer | a&waiting) != 0 {
			futex.Wait(&rw.state, a)
			continue

		}
		if atomic.CompareAndSwapUint32(&rw.state, a, a+1) {
			return
		}
	}
}

func (rw *RWMutex) RUnlock() {
	for {
		a := atomic.LoadUint32(&rw.state)

		if a&readers == 0 || a&writer != 0 {
			panic("чуть паники")
		}
		if atomic.CompareAndSwapUint32(&rw.state, a, a-1) {
			if (a-1)&readers == 0 {
				futex.WakeAll(&rw.state)
			}
			return
		}
	}
}

func (rw *RWMutex) Lock() {

	for {
		a := atomic.LoadUint32(&rw.state)
		if a&(writer|readers) == 0 {
			if atomic.CompareAndSwapUint32(&rw.state, a, writer) {
				return
			}
			continue
		}

		if a&waiting == 0 {
			if !atomic.CompareAndSwapUint32(&rw.state, a, a|waiting) {
				continue
			}
			a |= waiting
		}

		futex.Wait(&rw.state, a)

	}
}

func (rw *RWMutex) Unlock() {
	for {
		a := atomic.LoadUint32(&rw.state)
		if a&writer == 0 {
			panic("паника")
		}
		if atomic.CompareAndSwapUint32(&rw.state, a, a&^writer) {
			break
		}
	}
	futex.WakeAll(&rw.state)
}
