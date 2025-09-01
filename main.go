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
	var mu sync.Mutex
	readCh := make(chan string, 10)
	parseCh :=make(chan UserLog, 10)
	report := Report{
		Levels: make(map[string]int),
		Users:  make(map[string]int),
	}
	wg.Add(1)
	//TESTING WITH TESTLOG
	go FileReader(ctx, &wg, "test.log", readCh)
	wg.Add(1)
	go Parser(ctx, &wg, readCh, parseCh, &mu, &report)
	go func() {
		for log := range parseCh {
			fmt.Println("log:", log)
		}
	}()
	time.Sleep(2 * time.Second)
	cancel()
	wg.Wait()
	close(parseCh)
	fmt.Println("Processed:", report.Processed)
	fmt.Println("Malformed:", report.Malformed)
}
