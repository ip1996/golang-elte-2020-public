package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// TODO: do not print results in case of any error
	results := make(map[string]int)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, path := range Files() {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			if hash, err := Hash(path); err == nil { // HL
				mu.Lock()
				defer mu.Unlock()
				results[path] = hash
			} else {
				fmt.Printf("ERROR %s\n", err) // HL
			}
		}(path)
	}
	wg.Wait()
	fmt.Println(results) // HL
	// END OMIT
}

func Hash(path string) (int, error) {
	time.Sleep(100 * time.Millisecond)
	if len(path) == 12 {
		return 0, fmt.Errorf("cannot calculate hash for %q", path) // HL
	}
	return len(path), nil
}

func Files() []string {
	return []string{"ex1/ex1.go", "ex2/cksum.go", "ex3/ex3.go"}
}
