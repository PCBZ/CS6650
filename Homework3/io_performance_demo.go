package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func unbufferedWrite() time.Duration {
	file, _ := os.Create("unbuffered_output.txt")
	defer file.Close()

	start := time.Now()

	for range 100000 {
		line := "This is some sample data for performance testing.\n"
		_, _ = file.Write([]byte(line))
	}

	duration := time.Since(start)
	fmt.Println("Unbuffered write took:", duration)
	return duration
}

func bufferedWrite() time.Duration {
	file, _ := os.Create("buffered_output.txt")
	defer file.Close()

	buffer := bufio.NewWriter(file)
	defer buffer.Flush()

	start := time.Now()

	for range 100000 {
		line := "This is some sample data for performance testing.\n"
		_, _ = buffer.WriteString(line)
	}

	buffer.Flush()

	duration := time.Since(start)
	fmt.Println("Buffered write took:", duration)
	return duration
}

func main() {
	unbufferedWrite()
	bufferedWrite()
}
