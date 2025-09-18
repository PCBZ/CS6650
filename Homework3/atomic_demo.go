package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func atomic_increment() {
	var ops atomic.Uint64

	var wg sync.WaitGroup

	for range 50 {
		wg.Go(func() {
			for range 1000 {

				ops.Add(1)
			}
		})
	}

	wg.Wait()

	fmt.Println("ops:", ops.Load())
}

func normal_increment() {
	var ops uint64
	var wg sync.WaitGroup

	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 1000 {
				ops++
			}
		}()
	}

	wg.Wait()
	fmt.Println("ops:", ops)
}

func main() {

	atomic_increment()
	normal_increment()
}
