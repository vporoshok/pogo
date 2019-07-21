package pogo

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSyncLoader(t *testing.T) {
	var counter int64

	fn := func() interface{} {
		atomic.AddInt64(&counter, 1)
		time.Sleep(200 * time.Millisecond)

		return atomic.LoadInt64(&counter)
	}

	sl := new(syncLoader)
	wg := &sync.WaitGroup{}

	for i := 1; i < 10; i++ {
		wg.Add(1)
		go func() {
			v := sl.Load(1, fn)
			assert.EqualValues(t, 1, v)
			wg.Done()
		}()
	}
	wg.Wait()
	v := sl.Load(1, fn)
	assert.EqualValues(t, 1, v)
	assert.EqualValues(t, 1, counter)
}
