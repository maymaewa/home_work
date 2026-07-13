package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return nil
	}

	if m <= 0 {
		m = len(tasks) + 1
	}

	taskCh := make(chan Task)

	var (
		wg          sync.WaitGroup
		errorsCount atomic.Int32
	)

	worker := func() {
		defer wg.Done()

		for task := range taskCh {
			if task() != nil {
				errorsCount.Add(1)
			}
		}
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker()
	}

	for _, task := range tasks {
		if int(errorsCount.Load()) >= m {
			break
		}

		taskCh <- task
	}

	close(taskCh)
	wg.Wait()

	if int(errorsCount.Load()) >= m {
		return ErrErrorsLimitExceeded
	}

	return nil
}
