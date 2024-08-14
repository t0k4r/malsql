package malsql

import (
	"fmt"
	"malsql/mal"
	"slices"
	"sync"

	"github.com/t0k4r/x/slicesx"
)

func New(start int, end int, fast bool) {
	uc := NewUniq[int]()
	go func() {
		for i := start; i < end; i++ {
			uc.Send(i)
		}
		uc.Wait()
		uc.Close()
	}()
	for url, ok := uc.Recv(); ok; url, ok = uc.Recv() {
		anime, err := mal.FetchAnime(mal.UrlFromId(url))
		if err != nil {
			panic(err)
		}
		if anime != nil {
			fmt.Printf("%v %v\n", url, anime.Title)
			for _, urls := range anime.Related {
				uc.Send(slicesx.Map(urls, func(t mal.Url) int { return t.Id() })...)
			}
		} else {
			fmt.Println(nil)
		}
	}
}

type Uniq[T comparable] struct {
	in   chan T
	out  chan T
	done []T
	wg   *sync.WaitGroup
}

func NewUniq[T comparable]() Uniq[T] {
	u := Uniq[T]{
		in:   make(chan T),
		out:  make(chan T),
		done: []T{},
		wg:   &sync.WaitGroup{},
	}
	go u.run()
	return u
}
func NewUniqSized[T comparable](size int) Uniq[T] {
	u := Uniq[T]{
		in:   make(chan T, size),
		out:  make(chan T, size),
		done: []T{},
		wg:   &sync.WaitGroup{},
	}
	go u.run()
	return u
}

func (u *Uniq[T]) run() {
	for data := range u.in {
		if !slices.Contains(u.done, data) {
			u.done = append(u.done, data)
			u.out <- data
		}
	}
	close(u.out)
}

func (u *Uniq[T]) Wait() {
	u.wg.Wait()
}

func (u *Uniq[T]) Close() {
	close(u.in)
}

func (u *Uniq[T]) Send(items ...T) {
	u.wg.Add(1)
	for _, item := range items {
		u.in <- item
	}
	u.wg.Done()
}
func (u *Uniq[T]) Recv() (T, bool) {
	data, ok := <-u.out
	return data, ok
}
