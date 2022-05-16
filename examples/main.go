package main

import (
	"log"
	"time"

	bossworker "github.com/alexalreadytaken/go-boss-worker"
)

type Response struct {
	val int
	err error
}

func main() {
	events := make(chan int)
	responses := make(chan Response)

	go produceValues(events)
	go consumeValues(responses)
	boss := bossworker.Boss[int, Response]{
		MaxWorkersCount: 3,
		WorkerLifetime:  time.Second * 4,
		Input:           events,
		Output:          responses,
		Worker:          worker,
	}
	boss.Run()
}

func worker(req int) Response {
	log.Println("work")
	time.Sleep(time.Millisecond * 200)
	return Response{
		val: req * 2,
		err: nil,
	}
}

func produceValues(actions chan int) {
	for i := 0; i < 10; i++ {
		log.Println("produce")
		time.Sleep(time.Millisecond * 100)
		actions <- 20
	}
}

func consumeValues(responses chan Response) {
	for resp := range responses {
		log.Println("response=", resp)
	}
}
