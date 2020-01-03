package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/hyeongyuan/go-crawler/src/imbc"
)

func main() {
	cpuNumber := 3
	runtime.GOMAXPROCS(cpuNumber)

	startTime := time.Now()

	// ytn.Crawler(startTime)
	imbc.Crawler(startTime)

	elapsedTime := time.Since(startTime)

	fmt.Printf("Run time: %s\n", elapsedTime)
}
