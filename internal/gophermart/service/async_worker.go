package service

import (
	"errors"
	"sync"
	"time"

	"github.com/apolsh/yapr-gophermart/internal/logger"
)

var workerLogger = logger.LoggerOfComponent("async_worker")

type AsyncWorker struct {
	workerTaskCh chan func()
	wg           sync.WaitGroup
}

var ErrIllegalArgumentMaxWorkOnFly = errors.New("illegal argument: number of max work on fly")

func (w *AsyncWorker) ExecuteTask(task func()) {
	go func() {
		w.workerTaskCh <- task
	}()
}

func NewAsyncWorker(maxWorkOnFly int) (*AsyncWorker, error) {
	if maxWorkOnFly < 2 {
		return &AsyncWorker{}, ErrIllegalArgumentMaxWorkOnFly
	}

	workerTaskCh := make(chan func(), maxWorkOnFly)
	worker := &AsyncWorker{workerTaskCh: workerTaskCh}
	go func() {
		for functionTask := range workerTaskCh {
			worker.wg.Add(1)
			functionTask := functionTask
			go func() {
				functionTask()
				worker.wg.Done()
			}()
		}
	}()

	return &AsyncWorker{workerTaskCh: workerTaskCh}, nil

}

func (w *AsyncWorker) Close() {
	waitWithTimeout(&w.wg, 5*time.Second)
}

func waitWithTimeout(wg *sync.WaitGroup, t time.Duration) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
	case <-time.After(t):
		workerLogger.Info("async worker graceful shutdown timeout exceeded")
	}
}
