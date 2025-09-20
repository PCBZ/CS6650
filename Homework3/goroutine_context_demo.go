//go:build ignore

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func contextSwitch(maxProcess int) time.Duration {
	runtime.GOMAXPROCS(maxProcess)

	ping := make(chan struct{})
	pong := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)

	start := time.Now()

	// ping goroutine
	go func() {
		defer wg.Done()
		for range 100000 {
			ping <- struct{}{}
			<-pong
		}
	}()

	// pong goroutine
	go func() {
		defer wg.Done()
		for range 100000 {
			<-ping
			pong <- struct{}{}
		}
	}()

	wg.Wait()

	duration := time.Since(start)
	fmt.Printf("Context switch with GOMAXPROCS=%d took: %v\n", maxProcess, duration)
	return duration
}

func calculateSwitchTime(duration time.Duration) time.Duration {
	totalSwitches := 100000 * 2
	averageSwitchTime := duration / time.Duration(totalSwitches)

	return averageSwitchTime
}

func main() {
	fmt.Println("=== Test 1: Single OS Thread ===")
	singleThreadDuration := contextSwitch(1)
	singleThreadAvg := calculateSwitchTime(singleThreadDuration)
	fmt.Printf("Average context switch time (1 thread): %v\n\n", singleThreadAvg)

	time.Sleep(2 * time.Second)

	fmt.Println("=== Test 2: Multiple OS Threads ===")
	threads := runtime.NumCPU()
	multiThreadDuration := contextSwitch(threads)
	multiThreadAvg := calculateSwitchTime(multiThreadDuration)
	fmt.Printf("Average context switch time (%d threads): %v\n", threads, multiThreadAvg)
}
