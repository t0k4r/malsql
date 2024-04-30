package get

import (
	"runtime"
	"slices"
	"sync"

	"golang.org/x/sync/errgroup"
)

type UniqChan struct {
	wg   sync.WaitGroup
	done []int
	push chan int
}

func NewUniqChan() *UniqChan {
	return &UniqChan{
		wg:   sync.WaitGroup{},
		done: []int{},
		push: make(chan int),
	}
}

func (u *UniqChan) Skip(v ...int) *UniqChan {
	u.done = append(u.done, v...)
	return u
}

func (u *UniqChan) Generate(lo, hi int) *UniqChan {
	u.wg.Add(1)
	go func() {
		for i := lo; i < hi; i++ {
			u.push <- i
		}
		u.wg.Done()
	}()
	return u
}

func (u *UniqChan) Append(v ...int) *UniqChan {
	u.wg.Add(1)
	go func() {
		for _, i := range v {
			u.push <- i
		}
		u.wg.Done()
	}()
	return u
}
func (u *UniqChan) closeOnWait() {
	go func() {
		u.wg.Wait()
		close(u.push)
	}()
}

func (u *UniqChan) Each(do func(*UniqChan, int) error) error {
	u.closeOnWait()
	for i := range u.push {
		if slices.Contains(u.done, i) {
			continue
		}
		u.done = append(u.done, i)
		err := do(u, i)
		if err != nil {
			return err
		}

	}
	return nil
}
func (u *UniqChan) ParalellEach(do func(*UniqChan, int) error) error {
	u.closeOnWait()
	eg := new(errgroup.Group)
	eg.SetLimit(runtime.NumCPU())
	for i := range u.push {
		if slices.Contains(u.done, i) {
			continue
		}
		u.done = append(u.done, i)
		eg.Go(func(i int) func() error {
			u.wg.Add(1)
			return func() error {
				defer u.wg.Done()
				return do(u, i)
			}
		}(i))
	}
	return eg.Wait()
}
