package uniqchan

import (
	"slices"
	"sync"
)

type Uniqchan[T comparable] struct {
	Chan chan T
	done []T
}

func New[T comparable]() *Uniqchan[T] {
	return &Uniqchan[T]{
		Chan: make(chan T),
	}
}

func (u *Uniqchan[T]) Push(v ...T) {
	for _, v := range v {
		if slices.Contains(u.done, v) {
			continue
		}
		u.done = append(u.done, v)
		u.Chan <- v
	}
}

func (u *Uniqchan[T]) GoPush(wg *sync.WaitGroup, v ...T) {
	var vals []T
	for _, v := range v {
		if slices.Contains(u.done, v) {
			continue
		}
		u.done = append(u.done, v)
		vals = append(vals, v)
	}
	if wg != nil {
		wg.Add(1)
	}
	go func() {
		for _, v := range vals {
			u.Chan <- v
		}
		if wg != nil {
			wg.Done()
		}
	}()
}

func Generate(uq *Uniqchan[int], closeWg *sync.WaitGroup, lo int, hi int) {
	go func() {
		for i := lo; i < hi; i++ {
			uq.Push(i)
		}
		closeWg.Wait()
		close(uq.Chan)
	}()
}
