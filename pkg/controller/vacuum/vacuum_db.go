package vacuum

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/timer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"go.etcd.io/bbolt"
	bolt "go.etcd.io/bbolt"
)

type DB struct {
	stdout     io.Writer
	dbMutex    sync.RWMutex
	db         atomic.Pointer[bolt.DB]
	Param      *config.Param
	fs         afero.Fs
	storeQueue *StoreQueue
}

// NewDB initializes a Controller with the given context, parameters, and dependencies.
func NewDB(ctx context.Context, param *config.Param, fs afero.Fs) *DB {
	db := &DB{
		stdout: os.Stdout,
		Param:  param,
		fs:     fs,
	}
	db.storeQueue = newStoreQueue(ctx, db)
	return db
}

// Get retrieves a package entry from the database by key. for testing purposes.
func (d *DB) Get(ctx context.Context, logE *logrus.Entry, key string) (*PackageEntry, error) {
	var pkgEntry *PackageEntry
	err := d.withDBRetry(ctx, logE, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketNamePkgs))
		if b == nil {
			return nil
		}
		value := b.Get([]byte(key))
		if value == nil {
			return nil
		}

		var err error
		pkgEntry, err = decodePackageEntry(value)
		return err
	}, View)
	return pkgEntry, err
}

// withDBRetry retries a database operation with exponential backoff.
func (d *DB) withDBRetry(ctx context.Context, logE *logrus.Entry, fn func(*bolt.Tx) error, dbAccessType DBAccessType) error {
	const (
		retries            = 2
		initialBackoff     = 100 * time.Millisecond
		exponentialBackoff = 2
	)
	backoff := initialBackoff
	for i := range retries {
		err := d.withDB(logE, fn, dbAccessType)
		if err == nil {
			return nil
		}

		logerr.WithError(logE, err).WithField("attempt", i+1).Warn("retrying database operation")

		if err := timer.Wait(ctx, backoff); err != nil {
			return fmt.Errorf("wait for retrying database operation: %w", err)
		}
		backoff *= exponentialBackoff
	}

	return fmt.Errorf("database operation failed after %d retries", retries)
}

// withDB executes a function within a database transaction.
func (d *DB) withDB(logE *logrus.Entry, fn func(*bolt.Tx) error, dbAccessType DBAccessType) error {
	db, err := d.getDB()
	if err != nil {
		return err
	}
	if db == nil {
		return nil
	}
	defer func() {
		if err := d.Close(); err != nil {
			logerr.WithError(logE, err).Error("close database")
		}
	}()

	if dbAccessType == Update {
		if err := db.Update(fn); err != nil {
			return fmt.Errorf("update database: %w", err)
		}
		return nil
	}
	if err := db.View(fn); err != nil {
		return fmt.Errorf("view database: %w", err)
	}
	return nil
}

// getDB retrieves the database instance, initializing it if necessary.
func (d *DB) getDB() (*bolt.DB, error) {
	if db := d.db.Load(); db != nil {
		return db, nil
	}

	d.dbMutex.Lock()
	defer d.dbMutex.Unlock()

	if db := d.db.Load(); db != nil {
		return db, nil
	}

	const dbFileMode = 0o600
	db, err := bolt.Open(filepath.Join(d.Param.RootDir, dbFile), dbFileMode, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("open database %v: %w", dbFile, err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketNamePkgs))
		if err != nil {
			return fmt.Errorf("create bucket '%v' in '%v': %w", bucketNamePkgs, dbFile, err)
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, fmt.Errorf("create bucket: %w", err)
	}

	d.db.Store(db)
	return db, nil
}

// Keep_DBOpen opens the database instance. This is used for testing purposes.
func (d *DB) TestKeepDBOpen() error {
	const dbFileMode = 0o600
	if _, err := bolt.Open(filepath.Join(d.Param.RootDir, dbFile), dbFileMode, &bolt.Options{
		Timeout: 1 * time.Second,
	}); err != nil {
		return fmt.Errorf("open database %v: %w", dbFile, err)
	}
	return nil
}

// Close closes the database instance.
func (d *DB) Close() error {
	d.dbMutex.Lock()
	defer d.dbMutex.Unlock()

	if d.db.Load() != nil {
		if err := d.db.Load().Close(); err != nil {
			return fmt.Errorf("close database: %w", err)
		}
		d.db.Store(nil)
	}

	return nil
}

// StorePackageInternal stores package entries in the database.
func (d *DB) StorePackageInternal(ctx context.Context, logE *logrus.Entry, pkg *Package, dateTime ...time.Time) error {
	lastUsedTime := time.Now()
	if len(dateTime) > 0 {
		lastUsedTime = dateTime[0]
	}
	return d.withDBRetry(ctx, logE, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketNamePkgs))
		if b == nil {
			return errors.New("bucket not found")
		}
		logE.WithFields(logrus.Fields{
			"package_name":    pkg.Name,
			"package_version": pkg.Version,
			"package_path":    pkg.PkgPath,
		}).Debug("storing package in vacuum database")

		pkgKey := pkg.PkgPath
		pkgEntry := &PackageEntry{
			LastUsageTime: lastUsedTime,
			Package:       pkg,
		}

		data, err := encodePackageEntry(pkgEntry)
		if err != nil {
			logerr.WithError(logE, err).WithFields(
				logrus.Fields{
					"package_name":    pkg.Name,
					"package_version": pkg.Version,
					"package_path":    pkg.PkgPath,
				}).Error("encode package")
			return fmt.Errorf("encode package %s: %w", pkg.Name, err)
		}

		if err := b.Put([]byte(pkgKey), data); err != nil {
			logerr.WithError(logE, err).WithField("package_path", pkgKey).Error("store package in vacuum database")
			return fmt.Errorf("store package %s: %w", pkg.Name, err)
		}
		return nil
	}, Update)
}
