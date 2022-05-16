package bossworker

import (
	"log"
	"time"
)

type Worker[Action any, Response any] func(Action) Response

type Boss[Action any, Response any] struct {
	WorkersCount   uint
	WorkerLifetime time.Duration
	Actions        chan Action
	Responses      chan Response
	Worker         Worker[Action, Response]
	actionsQueue   []Action
}

func (boss *Boss[Action, Response]) Run() {
	activeWorkers := make(chan struct{}, boss.WorkersCount)
	for {
		select {
		case newAction := <-boss.Actions:
			boss.enqueue(newAction)
		case activeWorkers <- struct{}{}:
			// how
			if !boss.haveActions() {
				<-activeWorkers
				continue
			}
			go boss.execute(boss.dequeue(), activeWorkers)
		}
	}
}

func (boss *Boss[Action, Response]) haveActions() bool {
	return len(boss.actionsQueue) != 0
}

func (boss *Boss[Action, Response]) execute(action Action, activeWorkers chan struct{}) {
	subResponse := make(chan Response)
	go func() {
		subResponse <- boss.Worker(action)
	}()
	select {
	case newResp := <-subResponse:
		boss.Responses <- newResp
		log.Println("worker done")
		<-activeWorkers
		return
	case <-time.After(boss.WorkerLifetime):
		log.Println("worker too long")
		<-activeWorkers
		return
	}
}

func (boss *Boss[Action, Response]) enqueue(action Action) {
	boss.actionsQueue = append(boss.actionsQueue, action)
}

func (boss *Boss[Action, Response]) dequeue() Action {
	elem := boss.actionsQueue[0]
	boss.actionsQueue = boss.actionsQueue[1:]
	return elem
}
