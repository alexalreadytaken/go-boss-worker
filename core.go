package bossworker

import (
	"context"
	"log"
	"time"
)

type Worker[Event any, Response any] func(context.Context, Event) Response

func NewBoss[Event any, Response any](
	maxWorkersCount uint,
	workerLifetime time.Duration,
	eventsChannelBuffer int,
	worker Worker[Event, Response]) (chan Event, chan Response) {
	events, responses := makeChannels[Event, Response](eventsChannelBuffer)
	go run(maxWorkersCount, workerLifetime, worker, events, responses)
	return events, responses
}

func makeChannels[Event any, Response any](
	eventsChannelBuffer int) (chan Event, chan Response) {
	var events chan Event
	if 0 >= eventsChannelBuffer {
		events = make(chan Event)
	} else {
		events = make(chan Event, eventsChannelBuffer)
	}
	return events, make(chan Response)
}

func run[Event any, Response any](
	maxWorkersCount uint,
	workerLifetime time.Duration,
	worker Worker[Event, Response],
	input chan Event,
	output chan Response) {
	activeWorkers := make(chan struct{}, maxWorkersCount)
	for {
		select {
		case activeWorkers <- struct{}{}:
			select {
			case newEvent := <-input:
				go execute(activeWorkers, workerLifetime, output, worker, newEvent)
			default: //without deadlock
				time.Sleep(time.Millisecond * 50)
				<-activeWorkers
				continue
			}
		}
	}
}

func execute[Event any, Response any](
	activeWorkers chan struct{},
	workerLifetime time.Duration,
	output chan Response,
	worker Worker[Event, Response],
	event Event) {
	subResponse := make(chan Response)
	ctx, cancel := context.WithCancel(context.Background()) //with timeout?
	defer cancel()
	go func() {
		subResponse <- worker(ctx, event)
	}()
	select {
	case newResp := <-subResponse:
		output <- newResp
		log.Println("worker done")
		<-activeWorkers
		return
	case <-time.After(workerLifetime):
		log.Println("worker too long")
		cancel()
		<-activeWorkers
		return
	}
}
