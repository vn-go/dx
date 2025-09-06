package internal

import (
	"sync"
)

type oneCallItem[TResult any] struct {
	Val TResult
	Err error
}
type initOnceCall[Tkey any, TResult any] struct {
	Once sync.Once
	Item *oneCallItem[TResult]
}

var cacheOnceCall sync.Map

func OnceCall[Tkey any, TResult any](key Tkey, fn func() (TResult, error)) (TResult, error) {
	actual, _ := cacheOnceCall.LoadOrStore(key, &initOnceCall[Tkey, TResult]{})
	onceCall := actual.(*initOnceCall[Tkey, TResult])
	onceCall.Once.Do(func() {
		onceCall.Item = &oneCallItem[TResult]{}
		onceCall.Item.Val, onceCall.Item.Err = fn()
	})
	return onceCall.Item.Val, onceCall.Item.Err
}
