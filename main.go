package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	//CLI FLAG
	
	cc := flag.Int("concurrency", 10, "Number of parser workers")
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--concurrency=N] file1 file2 ...\n", os.Args[0])
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//HANDLE SIGINT
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nSIGINT detected, shutting down gracefully...")
		cancel()
	}()
	time.Sleep(3*time.Second)//TO TEST GRACEFUL SHUTDOWN
	
	//HANDLE PERIODIC PROGRESS
	pTicker := time.NewTicker(5 * time.Second)
	defer pTicker.Stop()

	progressDone := make(chan struct{})
	var progressCounter atomic.Int64

	go func() {
		for {
			select {
			case <-pTicker.C:
				fmt.Printf("Progress: %d lines processed\n", progressCounter.Load())
			case <-progressDone:
				return
			}
		}
	}()

	startTime := time.Now()

	report := Pipeline(ctx, files, *cc, &progressCounter)

	close(progressDone)

	duration := time.Since(startTime)

	//PRINTING REPORT
	fmt.Printf("Time taken: %v\n", duration)
	fmt.Printf("Total processed: %d lines\n", report.Processed)
	fmt.Printf("Total malformed: %d lines\n", report.Malformed)
	
	fmt.Println("\nLogs per Level")
	for level, count := range report.Levels {
		fmt.Printf("%s: %d\n", level, count)
	}

	fmt.Println("\nTop 10 Users by Log Count:")
	printTop10Users(report.Users)
}

func printTop10Users(users map[string]int) {
	type userCount struct {
		User  string
		Count int
	}
	
	var userCounts []userCount
	for user, count := range users {
		userCounts = append(userCounts, userCount{User: user, Count: count})
	}

	sort.Slice(userCounts, func(i, j int) bool {
		return userCounts[i].Count > userCounts[j].Count
	})

	limit := 10
	if len(userCounts) < limit {
		limit = len(userCounts)
	}
	
	for i := 0; i < limit; i++ {
		fmt.Printf("%d. %s: %d logs\n", i+1, userCounts[i].User, userCounts[i].Count)
	}
}
func Pipeline(ctx context.Context, files []string, cc int, progressCounter *atomic.Int64) Report {
	var wg sync.WaitGroup
	var mu sync.Mutex

	readCh := make(chan string, 1000)
	parsedCh := make(chan UserLog, 1000)
	report := Report{
		Levels: make(map[string]int),
		Users:  make(map[string]int),
	}

	for _, f := range files {
		wg.Add(1)
		go FileReader(ctx, &wg, f, readCh)
	}

	go func() {
		wg.Wait()
		close(readCh)
	}()

	var parseWg sync.WaitGroup
	for i := 0; i < cc; i++ {
		parseWg.Add(1)
		go Parser(ctx, &parseWg, readCh, parsedCh, &mu, &report, progressCounter)
	}

	go func() {
		parseWg.Wait()
		close(parsedCh)
	}()

	var aggWg sync.WaitGroup
	aggWg.Add(1)
	go Aggregate(ctx, &aggWg, parsedCh, &mu, &report)
	aggWg.Wait()

	return report
}
