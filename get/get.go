package get

import (
	"slices"
	"sync"
)

type UniqChan2 struct {
	wg   sync.WaitGroup
	done []int
	gen  chan int
	pull chan int
	push chan []int
}

func NewUniqChan2() *UniqChan2 {
	return &UniqChan2{
		wg:   sync.WaitGroup{},
		done: []int{},
		gen:  make(chan int),
		pull: make(chan int),
		push: make(chan []int),
	}
}
func (u *UniqChan2) Generate(lo, hi int) *UniqChan2 {
	go func() {
		for i := lo; i < hi; i++ {
			u.gen <- i
		}
		close(u.gen)
	}()
	return u
}
func (u *UniqChan2) Run() (chan int, chan []int) {
	go func() {
		for {
			select {
			case ids, ok := <-u.push:
				if !ok {
					u.push = nil
				}
				for _, i := range ids {
					if slices.Contains(u.done, i) {
						continue
					}
					u.done = append(u.done, i)
					u.pull <- i
				}
			case i, ok := <-u.gen:
				if !ok {
					u.gen = nil
				}
				if slices.Contains(u.done, i) {
					continue
				}
				u.done = append(u.done, i)
				u.pull <- i
			}
			if u.gen == nil && u.push == nil {
				close(u.push)
			}
		}
	}()
	outPush := make(chan []int)
	go func() {
		wg := sync.WaitGroup{}
		for {
			ids, ok := <-outPush
			u.wg.Add(1)
			if !ok {
				wg.Wait()
				close(u.push)
			}
			wg.Add(1)
			go func() {
				u.push <- ids
				wg.Done()
			}()
		}
	}()
	return u.pull, outPush
}

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

// func (u *UniqChan) Skip(v ...int) *UniqChan {
// 	u.done = append(u.done, v...)
// 	return u
// }

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

// func (u *UniqChan) ParalellEach(do func(*UniqChan, int) error) error {
// 	u.closeOnWait()
// 	eg := new(errgroup.Group)
// 	eg.SetLimit(runtime.NumCPU())
// 	for i := range u.push {
// 		if slices.Contains(u.done, i) {
// 			continue
// 		}
// 		u.done = append(u.done, i)
// 		eg.Go(func(i int) func() error {
// 			u.wg.Add(1)
// 			return func() error {
// 				defer u.wg.Done()
// 				return do(u, i)
// 			}
// 		}(i))
// 	}
// 	return eg.Wait()
// }
