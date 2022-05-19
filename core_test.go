package bossworker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type BossWorkerCoreTestSuite struct {
	suite.Suite
	workerTimeout      time.Duration
	workersCount       uint
	workersInputBuffer int
	bossInput          chan TestEvent
	bossOutput         chan int
}

type TestEvent struct {
	value       int
	timeToSleep time.Duration
}

func testWorker(ctx context.Context, event TestEvent) int {
	time.Sleep(event.timeToSleep)
	select {
	case <-ctx.Done():
		return 0
	default:
		return event.value * 10
	}
}

func TestBossWorkerCore(t *testing.T) {
	suite.Run(t, &BossWorkerCoreTestSuite{})
}

func (s *BossWorkerCoreTestSuite) SetupSuite() {
	//load ?
	var count uint = 2
	timeout := time.Second * 2
	buffer := 10
	input, output := NewBoss(uint(count), timeout, buffer, testWorker)
	s.workersCount = count
	s.workerTimeout = timeout
	s.workersInputBuffer = buffer
	s.bossInput = input
	s.bossOutput = output
}

//fixme?
func (s *BossWorkerCoreTestSuite) TestCore() {
	testWorkerDonePositive(s)
	testWorkerTimeout(s)
	testConcurrency(s)
	testBuffer(s)
}

func testWorkerDonePositive(s *BossWorkerCoreTestSuite) {
	val := 10
	s.bossInput <- TestEvent{val, time.Second}
	res := <-s.bossOutput
	s.Equal(100, res)
}

func testWorkerTimeout(s *BossWorkerCoreTestSuite) {
	val := 10
	s.bossInput <- TestEvent{val, time.Second * 3}
	select {
	case <-s.bossOutput:
		s.Fail("worker continue execution after timeout")
	case <-time.After(s.workerTimeout):
		return
	}
}

func testConcurrency(s *BossWorkerCoreTestSuite) {
	s.bossInput <- TestEvent{10, time.Second}
	s.bossInput <- TestEvent{10, time.Second}
	select {
	case <-time.After(s.workerTimeout):
		s.Fail("workers executing longer than must")
	default:
		for i := 0; i < int(s.workersCount); i++ {
			res := <-s.bossOutput
			s.Equal(100, res)
		}
	}
}

func testBuffer(s *BossWorkerCoreTestSuite) {
	for i := 0; i < s.workersInputBuffer; i++ {
		s.bossInput <- TestEvent{i, time.Second}
	}
	time.Sleep(time.Millisecond * 500)
	expectedBufferCap := s.workersInputBuffer - int(s.workersCount)
	s.Equal(expectedBufferCap, len(s.bossInput))
}
