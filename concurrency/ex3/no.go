package main

import (
	"fmt"
	"time"
)

func main() {
	// TODO: parallelize the calculation
	ch := make(chan struct{})
	results := make(map[string]int)
	files := Files()
	for _, path := range files {
		go func(path string) {
			results[path] = Hash(path) // HLrace
			ch <- struct{}{}
		}(path)
	}

	for range files {
		<-ch
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
