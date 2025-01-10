package vacuum

import (
	"sync"

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
	vc        *Controller
	done      chan struct{}
	closeOnce sync.Once
}

// newStoreQueue initializes the task queue with a single worker.
func newStoreQueue(vc *Controller) *StoreQueue {
	const maxTasks = 100
	sq := &StoreQueue{
		taskQueue: make(chan StoreRequest, maxTasks),
		done:      make(chan struct{}),
		vc:        vc,
	}

	go sq.worker()
	return sq
}

// worker processes tasks from the queue.
func (sq *StoreQueue) worker() {
	for {
		select {
		case task, ok := <-sq.taskQueue:
			if !ok {
				return
			}
			err := sq.vc.storePackageInternal(task.logE, task.pkg)
			if err != nil {
				logerr.WithError(task.logE, err).Error("store package asynchronously")
			}
			sq.wg.Done()
		case <-sq.done:
			// Process remaining tasks
			for len(sq.taskQueue) > 0 {
				task := <-sq.taskQueue
				err := sq.vc.storePackageInternal(task.logE, task.pkg)
				if err != nil {
					logerr.WithError(task.logE, err).Error("store package asynchronously during shutdown")
				}
				sq.wg.Done()
			}
			return
		}
	}
}

// enqueue adds a task to the queue.
func (sq *StoreQueue) enqueue(logE *logrus.Entry, pkg *Package) {
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
