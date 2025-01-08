package vacuum

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	bolt "go.etcd.io/bbolt"
)

type Controller struct {
	stdout     io.Writer
	dbMutex    sync.RWMutex
	db         atomic.Pointer[bolt.DB]
	Param      *config.Param
	fs         afero.Fs
	storeQueue *StoreQueue
}

// New initializes a Controller with the given context, parameters, and dependencies.
func New(param *config.Param, fs afero.Fs) *Controller {
	vc := &Controller{
		stdout: os.Stdout,
		Param:  param,
		fs:     fs,
	}
	vc.storeQueue = newStoreQueue(vc)
	return vc
}

// getDB retrieves the database instance, initializing it if necessary.
func (vc *Controller) getDB() (*bolt.DB, error) {
	if db := vc.db.Load(); db != nil {
		return db, nil
	}

	vc.dbMutex.Lock()
	defer vc.dbMutex.Unlock()

	if db := vc.db.Load(); db != nil {
		return db, nil
	}

	const dbFileMode = 0o600
	db, err := bolt.Open(vc.Param.RootDir+"/"+dbFile, dbFileMode, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database %v: %w", dbFile, err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketNamePkgs))
		if err != nil {
			return fmt.Errorf("failed to create bucket '%v' in '%v': %w", bucketNamePkgs, dbFile, err)
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	vc.db.Store(db)
	return db, nil
}

// withDBRetry retries a database operation with exponential backoff.
func (vc *Controller) withDBRetry(logE *logrus.Entry, fn func(*bolt.Tx) error, dbAccessType DBAccessType) error {
	const (
		retries            = 2
		initialBackoff     = 100 * time.Millisecond
		exponentialBackoff = 2
	)
	backoff := initialBackoff
	for i := range retries {
		err := vc.withDB(logE, fn, dbAccessType)
		if err == nil {
			return nil
		}

		logE.WithFields(logrus.Fields{
			"attempt": i + 1,
			"error":   err,
		}).Warn("Retrying database operation")

		time.Sleep(backoff)
		backoff *= exponentialBackoff
	}

	return fmt.Errorf("database operation failed after %d retries", retries)
}

// withDB executes a function within a database transaction.
func (vc *Controller) withDB(logE *logrus.Entry, fn func(*bolt.Tx) error, dbAccessType DBAccessType) error {
	db, err := vc.getDB()
	if err != nil {
		return err
	}
	if db == nil {
		return nil
	}
	defer func() {
		if err := vc.closeDB(); err != nil {
			logE.WithError(err).Error("Failed to close database")
		}
	}()

	if dbAccessType == Update {
		err = db.Update(fn)
		if err != nil {
			return fmt.Errorf("failed to view database: %w", err)
		}
		return nil
	}
	err = db.View(fn)
	if err != nil {
		return fmt.Errorf("failed to view database: %w", err)
	}
	return nil
}

// Keep_DBOpen opens the database instance. This is used for testing purposes.
func (vc *Controller) TestKeepDBOpen() error {
	const dbFileMode = 0o600
	_, _ = bolt.Open(vc.Param.RootDir+"/"+dbFile, dbFileMode, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	return nil
}

// closeDB closes the database instance.
func (vc *Controller) closeDB() error {
	vc.dbMutex.Lock()
	defer vc.dbMutex.Unlock()

	if vc.db.Load() != nil {
		if err := vc.db.Load().Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		vc.db.Store(nil)
	}

	return nil
}

// Close closes the dependencies of the Controller.
func (vc *Controller) close(logE *logrus.Entry) error {
	logE.Debug("Closing vacuum controller")
	if vc.storeQueue != nil {
		vc.storeQueue.close()
	}

	vc.dbMutex.Lock()
	defer vc.dbMutex.Unlock()

	if vc.db.Load() != nil {
		if err := vc.db.Load().Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		vc.db.Store(nil)
	}
	return nil
}
