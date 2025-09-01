package main

import (
	"context"
	"encoding/json"
	"sync"
)

func Parser(ctx context.Context, wg *sync.WaitGroup, in <- chan string, out chan <- UserLog, mu *sync.Mutex, report *Report){
	defer wg.Done()
	for{
		select{
		case <-ctx.Done():
			return
		case l,ok:=<-in:
			if !ok{
				return
			}
			var log UserLog
			if err:=json.Unmarshal([]byte(l),&log);err!=nil{
				mu.Lock()
				report.Malformed++
				mu.Unlock()
				continue
			}
			select {
			case <-ctx.Done():
				return
			case out <- log:
				mu.Lock()
				report.Processed++
				mu.Unlock()
			}
		}
	}
}