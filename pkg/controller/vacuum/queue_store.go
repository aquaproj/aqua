package vacuum

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

// StoreRequest represents a task to be processed.
type StoreRequest struct {
	logE *logrus.Entry
	pkg  *Package
}

// StoreQueue manages a queue for handling tasks sequentially.
type StoreQueue struct {
	taskQueue chan StoreRequest
	wg        sync.WaitGroup
	vc        Storage
	done      chan struct{}
	closeOnce sync.Once
}

type Storage interface {
	Store(ctx context.Context, logE *logrus.Entry, pkg *Package, lastUsedTime time.Time) error
}

// newStoreQueue initializes the task queue with a single worker.
func newStoreQueue(ctx context.Context, vc Storage) *StoreQueue {
	const maxTasks = 100
	sq := &StoreQueue{
		taskQueue: make(chan StoreRequest, maxTasks),
		done:      make(chan struct{}),
		vc:        vc,
	}

	go sq.worker(ctx)
	return sq
}

// worker processes tasks from the queue.
func (sq *StoreQueue) worker(ctx context.Context) {
	for {
		select {
		case task, ok := <-sq.taskQueue:
			if !ok {
				return
			}
			if err := sq.vc.Store(ctx, task.logE, task.pkg, time.Now()); err != nil {
				logerr.WithError(task.logE, err).Error("store package asynchronously")
			}
			sq.wg.Done()
		case <-sq.done:
			// Process remaining tasks
			for len(sq.taskQueue) > 0 {
				task := <-sq.taskQueue
				if err := sq.vc.Store(ctx, task.logE, task.pkg, time.Now()); err != nil {
					logerr.WithError(task.logE, err).Error("store package asynchronously during shutdown")
				}
				sq.wg.Done()
			}
			return
		}
	}
}

// Enqueue adds a task to the queue.
func (sq *StoreQueue) Enqueue(logE *logrus.Entry, pkg *Package) {
	select {
	case <-sq.done:
		return
	default:
		sq.wg.Add(1)
		sq.taskQueue <- StoreRequest{
			logE: logE,
			pkg:  pkg,
		}
	}
}

// close waits for all tasks to complete and stops the worker.
func (sq *StoreQueue) close() {
	sq.closeOnce.Do(func() {
		close(sq.done)
		sq.wg.Wait()
		close(sq.taskQueue)
	})
}
