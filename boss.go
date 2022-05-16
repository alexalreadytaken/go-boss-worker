package bossworker

import (
	"log"
	"sync"
	"time"
)

type Worker[Event any, Response any] func(Event) Response

type Boss[Event any, Response any] struct {
	MaxWorkersCount uint
	WorkerLifetime  time.Duration
	Input           chan Event
	Output          chan Response
	Worker          Worker[Event, Response]
	eventsQueue     []Event
	eventsMutex     sync.Mutex
}

func (boss *Boss[Event, Response]) Run() {
	go boss.listenEvents()
	boss.handleEvents()
}

func (boss *Boss[Event, Response]) listenEvents() {
	for {
		select {
		case newAction := <-boss.Input:
			boss.enqueue(newAction)
		}
	}
}

func (boss *Boss[Event, Response]) handleEvents() {
	activeWorkers := make(chan struct{}, boss.MaxWorkersCount)
	for {
		select {
		case activeWorkers <- struct{}{}:
			if !boss.haveActions() {
				<-activeWorkers
				time.Sleep(time.Second) //maybe sleep in default: ?
				continue
			}
			go boss.execute(boss.dequeue(), activeWorkers)
		}
	}
}

func (boss *Boss[Event, Response]) haveActions() bool {
	boss.eventsMutex.Lock()
	defer boss.eventsMutex.Unlock()
	return len(boss.eventsQueue) != 0
}

func (boss *Boss[Event, Response]) execute(event Event, activeWorkers chan struct{}) {
	subResponse := make(chan Response)
	go func() {
		subResponse <- boss.Worker(event)
	}()
	select {
	case newResp := <-subResponse:
		boss.Output <- newResp
		log.Println("worker done")
		<-activeWorkers
		return
	case <-time.After(boss.WorkerLifetime):
		log.Println("worker too long")
		<-activeWorkers
		return
	}
}

func (boss *Boss[Event, Response]) enqueue(event Event) {
	boss.eventsMutex.Lock()
	boss.eventsQueue = append(boss.eventsQueue, event)
	boss.eventsMutex.Unlock()
}

func (boss *Boss[Event, Response]) dequeue() Event {
	boss.eventsMutex.Lock()
	defer boss.eventsMutex.Unlock()
	elem := boss.eventsQueue[0]
	boss.eventsQueue = boss.eventsQueue[1:]
	return elem
}
