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
	actions := make(chan int)
	responses := make(chan Response)

	go produceValues(actions)
	go consumeValues(responses)
	boss := bossworker.Boss[int, Response]{
		WorkersCount:   3,
		WorkerLifetime: time.Second * 2,
		Actions:        actions,
		Responses:      responses,
		Worker:         worker,
	}
	boss.Run()
}

func worker(req int) Response {
	log.Println("work")
	time.Sleep(time.Second * 3)
	return Response{
		val: req ^ 2,
		err: nil,
	}
}

func produceValues(actions chan int) {
	for i := 0; i < 100; i++ {
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
