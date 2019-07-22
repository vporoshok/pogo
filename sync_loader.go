package pogo

import "sync"

type wait chan struct{}

type syncLoader struct {
	m sync.Map
}

func (sl *syncLoader) Load(key interface{}, fn func() interface{}) interface{} {
	if value, ok := sl.m.Load(key); ok {
		return sl.await(key, value)
	}
	var ch = make(wait)

	if value, ok := sl.m.LoadOrStore(key, ch); ok {
		close(ch)
		return sl.await(key, value)
	}
	value := fn()
	sl.m.Store(key, value)
	close(ch)

	return value
}

func (sl *syncLoader) await(key, value interface{}) interface{} {
	if w, ok := value.(wait); ok {
		<-w
		value, _ = sl.m.Load(key)
	}
	return value
}
