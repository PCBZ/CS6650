//go:build ignore

package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

func mapWrite() {
	m := make(map[int]int)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := goroutineID*1000 + j
				m[key] = key
			}
		}(i)
	}
	wg.Wait()
	fmt.Println("map length:", len(m))
}

func mapSafeWrite(mu sync.Locker) (time.Duration, int) {
	m := make(map[int]int)
	var wg sync.WaitGroup

	start := time.Now()

	for i := range 50 {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := goroutineID*1000 + j
				mu.Lock()
				m[key] = key
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	duration := time.Since(start)
	fmt.Println("map write with mutex lock took:", duration)
	fmt.Println("map length:", len(m))

	return duration, len(m)
}

func mapSyncMapWrite() (time.Duration, int) {
	var m sync.Map
	var wg sync.WaitGroup

	start := time.Now()

	for i := range 50 {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := goroutineID*1000 + j
				m.Store(key, j)
			}
		}(i)
	}
	wg.Wait()

	duration := time.Since(start)
	fmt.Println("sync map write took:", duration)

	count := 0
	m.Range(func(key, value any) bool {
		count++
		return true
	})
	fmt.Println("map length:", count)

	return duration, count
}

func saveDurationToCSV(durationMs float64, mapLength int, filename string) {
	// Open CSV file in append mode
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening CSV file: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Check if file is empty (new file) to add header
	fileInfo, _ := file.Stat()
	if fileInfo.Size() == 0 {
		// Write header
		writer.Write([]string{"duration_ms", "map_length", "timestamp"})
	}

	// Write data
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	record := []string{
		strconv.FormatFloat(durationMs, 'f', 3, 64),
		strconv.Itoa(mapLength),
		timestamp,
	}

	writer.Write(record)
}

func runMultipleTimes(numRuns int) {
	fmt.Printf("Running mapSafeWrite %d times and saving to CSV...\n", numRuns)

	for range numRuns {
		duration, length := mapSafeWrite(&sync.Mutex{})
		durationMs := float64(duration.Nanoseconds()) / 1000000.0
		saveDurationToCSV(durationMs, length, "mutex_durations.csv")
	}

	for range numRuns {
		duration, length := mapSafeWrite(&sync.RWMutex{})
		durationMs := float64(duration.Nanoseconds()) / 1000000.0
		saveDurationToCSV(durationMs, length, "rwmutex_durations.csv")
	}

	for range numRuns {
		duration, length := mapSyncMapWrite()
		durationMs := float64(duration.Nanoseconds()) / 1000000.0
		saveDurationToCSV(durationMs, length, "syncmap_durations.csv")
	}
}

func main() {
	// mapWrite()
	runMultipleTimes(100)
}
