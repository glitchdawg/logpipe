package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"
)

func FileReader(ctx context.Context, wg *sync.WaitGroup, path string, out chan<- string) {
	defer wg.Done()
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Could not open file: %v", err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		select {
		case <-ctx.Done():
			fmt.Printf("Stopping reader for %s...", path)
			return
		case out <- line:
			//PASS TO OUT CHANNEL
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Printf("error reading file %s: %v\n", path, err)
	}

}
