// Binary watch tracks changes in a directory structure.
package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

//main is checking for modifications on the fileset in every second
func main() {
	prev := HashAll()
	for ts := range time.Tick(time.Second) {
		curr := HashAll()
		added, edited, deleted := CompareFileSets(prev, curr)
		prev = curr
		if len(added)+len(edited)+len(deleted) > 0 {
			fmt.Printf("files have changed at %v\n", ts)
			fmt.Printf("\tadded: %q\n", added)
			fmt.Printf("\tedited: %q\n", edited)
			fmt.Printf("\tdeleted: %q\n", deleted)
		}
	}
}

//Hashed is a struct for storing the file path and hash value or error
type Hashed struct {
	Path string
	Hash []byte
	Err  error // Hash is invalid, in case of an error
}

//HashedEqual compares two hashed values
func HashedEqual(before, after *Hashed) bool {
	if before == nil || after == nil {
		return before == nil && after == nil
	}
	if before.Path != after.Path {
		return false
	}
	if be, ae := before.Err != nil, after.Err != nil; be || ae {
		return be == ae
	}
	if len(before.Hash) != len(after.Hash) {
		return false
	}
	for i := 0; i < len(before.Hash); i++ {
		if before.Hash[i] != after.Hash[i] {
			return false
		}
	}
	return true
}

//FileSet is a mapping between pathes and hashed values
type FileSet map[string]*Hashed

// CompareFileSets compares checksums of files to detect differences.
func CompareFileSets(before, after FileSet) (added, edited, deleted []string) {
	for bp, bh := range before {
		switch ah, has := after[bp]; {
		case !has:
			deleted = append(deleted, bp)
		case !HashedEqual(bh, ah):
			edited = append(edited, bp)
		}
	}
	for ap := range after {
		if _, has := before[ap]; !has {
			added = append(added, ap)
		}
	}
	return added, edited, deleted
}

//HashAll fills up a shared map during goroutines with the hashed values of the filesystem
func HashAll() FileSet {
	// TODO: parallelize the checksum calculation
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	results := make(FileSet)
	for _, path := range Files() {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			h := Hash(p)
			mu.Lock()
			defer mu.Unlock()
			results[p] = h
		}(path)
	}
	wg.Wait()
	// END OMIT
	return results
}

// Hash calculates a checksum of a file.
// It returns an error, if the file was not readable.
func Hash(path string) *Hashed {
	f, err := os.Open(path)
	if err != nil {
		return &Hashed{Path: path, Err: err}
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return &Hashed{Path: path, Err: err}
	}
	return &Hashed{Path: path, Hash: h.Sum(nil)}
}

// Files returns the list of file paths that are expanded from walking the tree
// of every command line arguments.
func Files() []string {
	var files []string
	flag.Parse()
	for _, path := range flag.Args() {
		// Walk will return no error, because all WalkFunc always returns nil.
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("ERROR: unable to access %q\n", path)
				return nil
			}
			if info.Mode()&os.ModeType != 0 {
				return nil // Not a regular file.
			}
			files = append(files, path)
			return nil
		})
	}
	return files
}
