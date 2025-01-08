package vacuum

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// storeRequest represents a task to be processed.
type StoreRequest struct {
	logE *logrus.Entry
	pkg  []*ConfigPackage
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
				task.logE.WithField("vacuum", dbFile).WithError(err).Error("Failed to store package asynchronously")
			}
			sq.wg.Done()
		case <-sq.done:
			// Process remaining tasks
			for len(sq.taskQueue) > 0 {
				task := <-sq.taskQueue
				err := sq.vc.storePackageInternal(task.logE, task.pkg)
				if err != nil {
					task.logE.WithError(err).Error("Failed to store package asynchronously during shutdown")
				}
				sq.wg.Done()
			}
			return
		}
	}
}

// Enqueue adds a task to the queue.
func (sq *StoreQueue) enqueue(logE *logrus.Entry, pkg []*ConfigPackage) {
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

// Close waits for all tasks to complete and stops the worker.
func (sq *StoreQueue) close() {
	sq.closeOnce.Do(func() {
		close(sq.done)
		sq.wg.Wait()
		close(sq.taskQueue)
	})
}
