package service

import "errors"

type AsyncWorker struct {
	workerTaskCh chan func()
}

var ErrIllegalArgumentMaxWorkOnFly = errors.New("illegal argument: number of max work on fly")

func (w AsyncWorker) ExecuteTask(task func()) {
	go func() {
		w.workerTaskCh <- task
	}()
}

func NewAsyncWorker(maxWorkOnFly int) (*AsyncWorker, error) {
	if maxWorkOnFly < 2 {
		return &AsyncWorker{}, ErrIllegalArgumentMaxWorkOnFly
	}

	workerTaskCh := make(chan func(), maxWorkOnFly)
	go func() {
		for functionTask := range workerTaskCh {
			functionTask := functionTask
			go func() {
				functionTask()
			}()
		}
	}()

	return &AsyncWorker{workerTaskCh: workerTaskCh}, nil

}
