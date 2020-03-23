package main

import (
	"fmt"
	"time"
)

func translate(word string) string {
	time.Sleep(100 * time.Millisecond)
	return word
}

func main() {
	text := []string{"hello", "world"}
	start := time.Now()

	// TODO: parallelize the translation of all words in 'text'
	for _, word := range text {
		fmt.Println(translate(word))
	}
	// END OMIT

	if time.Since(start) > time.Duration(len(text))*80*time.Millisecond {
		fmt.Println("Too late...")
	}
}
