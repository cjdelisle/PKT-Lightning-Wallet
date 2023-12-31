package event_test

import (
	"sync"
	"testing"

	"github.com/pkt-cash/pktd/btcutil/event"
	"github.com/pkt-cash/pktd/btcutil/lock"
	"github.com/stretchr/testify/assert"
)

type secretNumber struct {
	n int
}

func TestEventSimple(t *testing.T) {
	sn := secretNumber{3}
	ee := event.NewEmitter[secretNumber]("my emitter")
	var active sync.WaitGroup
	var done sync.WaitGroup
	complete := lock.AtomicInt32{}

	for i := 0; i < 10; i++ {
		active.Add(1)
		event.GoWg(&done, func(loop *event.Loop) {
			ee.On(loop, func(sn secretNumber) {
				assert.Equal(t, sn.n, 3)
				sn.n = 5
				loop.CurrentHandler().Cancel()
				complete.Add(1)
			})
			active.Done()
		})
	}

	active.Wait()
	ee.TryEmit(sn)
	done.Wait()
	if complete.Load() != 10 {
		t.Fail()
	}
}

func TestMulti(t *testing.T) {
	ee1 := event.NewEmitter[secretNumber]("my emitter")
	ee2 := event.NewEmitter[secretNumber]("my emitter2")
	var active sync.WaitGroup
	var done sync.WaitGroup
	complete := lock.AtomicInt32{}

	for i := 0; i < 10; i++ {
		active.Add(1)
		event.GoWg(&done, func(loop *event.Loop) {
			ecount1 := 0
			ecount2 := 0
			ee1.On(loop, func(sn secretNumber) {
				assert.Equal(t, sn.n, 3)
				sn.n = 5
				ecount1 += 1
				if ecount1 == 5 {
					loop.CurrentHandler().Cancel()
					complete.Add(1)
				}
			})
			ee2.On(loop, func(sn secretNumber) {
				assert.Equal(t, sn.n, 7)
				sn.n = 5
				ecount2 += 1
				if ecount2 == 5 {
					loop.CurrentHandler().Cancel()
					complete.Add(1)
				}
			})
			active.Done()
		})
	}

	active.Wait()
	for i := 0; i < 5; i++ {
		ee1.TryEmit(secretNumber{3})
		ee2.TryEmit(secretNumber{7})
	}
	done.Wait()
	if complete.Load() != 20 {
		t.Fail()
	}
}
