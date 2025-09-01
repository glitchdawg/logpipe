package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	readCh := make(chan string, 10)

	wg.Add(1)
	go FileReader(ctx, &wg, "test.log", readCh)
	go func() {
		for i := range readCh {
			fmt.Println(i)
		}
	}()
	time.Sleep(2 * time.Second)
	cancel()
	wg.Wait()
}
