package main

import (
	"context"
	"sync"
)

func Aggregate(ctx context.Context, wg *sync.WaitGroup,  in <-chan UserLog, mu *sync.Mutex, report *Report) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case log, ok := <-in:
			if !ok {
				return
			}
			mu.Lock()
			report.Levels[log.Level]++
			report.Users[log.User]++
			mu.Unlock()
		}

	}
}
