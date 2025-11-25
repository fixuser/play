package worker

import (
	"errors"
	"sync/atomic"
)

var (
	ErrPoolClosed = errors.New("pool is closed")
)

// WorkerPool is a pool of workers that can be used to limit the number of
type WorkerPool struct {
	limit   int
	tickers chan int
	num     atomic.Int32
}

// NewWorkerPool creates a new worker pool with the given limit
func NewWorkerPool(limit int) *WorkerPool {
	if limit <= 0 {
		limit = 10
	}

	wp := &WorkerPool{
		limit:   limit,
		tickers: make(chan int, limit),
	}

	for i := 0; i < limit; i++ {
		wp.tickers <- 1
	}

	return wp
}

// Do	add a job to the pool
func (wp *WorkerPool) Do(job func()) (ticket int, err error) {
	ticket, ok := <-wp.tickers
	if !ok {
		return -1, ErrPoolClosed
	}

	wp.num.Add(1)

	go func() {
		if job != nil {
			job()
		}
		wp.tickers <- ticket
		wp.num.Add(-1)
	}()

	return ticket, nil
}

// Wait waits for all workers to finish
func (wp *WorkerPool) Wait() {
	for i := 0; i < wp.limit; i++ {
		<-wp.tickers
	}
	close(wp.tickers)
}

// Num returns the number in progress of workers in the pool
func (wp *WorkerPool) Num() int {
	return int(wp.num.Load())
}
