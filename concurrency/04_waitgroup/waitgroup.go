package waitgroup

import (
	"primitives/internal/futex"
	"sync/atomic"
)

type WaitGroup struct {
	count uint32
}

func (wg *WaitGroup) Add(delta int) {

	v := atomic.AddUint32(&wg.count, uint32(delta))
	if int32(v) < 0 {
		panic("счетчик ушел в минус")
	}
	if v == 0 {
		futex.WakeAll(&wg.count)
	}

}

func (wg *WaitGroup) Done() {
	wg.Add(-1)

}

func (wg *WaitGroup) Wait() {
	for {
		c := atomic.LoadUint32(&wg.count)
		if c == 0 {
			return
		}
		futex.Wait(&wg.count, c)
	}
}
