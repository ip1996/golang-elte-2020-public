package main

import (
	"fmt"
	"time"
)

func main() {
	// TODO: parallelize the calculation
	results := make(map[string]int)
	for _, path := range Files() {
		results[path] = Hash(path)
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
