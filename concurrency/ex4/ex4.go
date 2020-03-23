package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// TODO: wait for all results
	results := make(map[string]int)
	mu := sync.Mutex{}
	for _, path := range Files() {
		go func(path string) {
			mu.Lock()
			defer mu.Unlock()
			results[path] = Hash(path)
		}(path)
	}
	// END OMIT

	fmt.Println(results)
}

func Hash(path string) int {
	time.Sleep(100 * time.Millisecond)
	return len(path) // *not* collision free hash
}

func Files() []string {
	return []string{"ex1/ex1.go", "ex2/cksum.go", "ex3/ex3.go"}
}
