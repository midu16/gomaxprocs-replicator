package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func main() {
	fmt.Println("=== Go 1.24 GOMAXPROCS Replicator ===")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("Host Logical CPUs (runtime.NumCPU): %d\n", runtime.NumCPU())
	fmt.Printf("Default GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Println("=====================================")

	// Simulate a CPU-intensive bursty workload
	fmt.Println("Starting CPU burst simulation...")
	start := time.Now()
	
	var wg sync.WaitGroup
	// Spawn goroutines equal to GOMAXPROCS to force thread creation
	workers := runtime.GOMAXPROCS(0)
	wg.Add(workers)
	
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			// Arbitrary tight loop to burn CPU cycles
			count := 0
			for j := 0; j < 50000000; j++ {
				count++
			}
		}(i)
	}
	
	wg.Wait()
	elapsed := time.Since(start)
	
	fmt.Printf("Burst completed in: %v\n", elapsed)
	fmt.Println("If running under a tight K8s CPU limit, this time will be artificially inflated due to severe CFS throttling.")
	
	// Keep the pod alive for inspection via OpenShift CLI or Podman
	fmt.Println("Container is now idling. Press Ctrl+C to exit.")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nShutting down.")
}