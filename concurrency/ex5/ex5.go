package main

import (
	"fmt"
	"time"
)

func main() {
	// TODO: close the channel after the last send
	type result struct {
		path string
		hash int
	}
	ch := make(chan *result)
	for _, path := range Files() {
		go func(path string) {
			hash := Hash(path)
			ch <- &result{path, hash}
		}(path)
	}
	time.Sleep(200 * time.Millisecond) // to wait for all results
	close(ch)                          // that "range ch" works

	results := make(map[string]int)
	for r := range ch {
		results[r.path] = r.hash
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
