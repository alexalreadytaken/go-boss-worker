package main

import (
	"context"
	"log"
	"time"

	bossworker "github.com/alexalreadytaken/go-boss-worker"
)

type Response struct {
	val int
	err error
}

const count = 7

func main() {
	events, responses := bossworker.NewBoss(3, time.Second, 100, worker)
	go produceValues(events)
	for i := 0; i < count; i++ {
		resp := <-responses
		log.Println("response=", resp)
	}
	log.Println("done")
}

func worker(ctx context.Context, req int) Response {
	log.Println("work")
	time.Sleep(time.Millisecond * 1200)
	select {
	case <-ctx.Done():
		return Response{err: ctx.Err()}
	default:
		return Response{val: req * 2}
	}
}

func produceValues(events chan int) {
	for i := 0; i < count; i++ {
		log.Println("produce")
		time.Sleep(time.Millisecond * 100)
		events <- 20
	}
}
