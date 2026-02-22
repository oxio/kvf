package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/gofrs/flock"
)

var baseLockDir string

type Lock struct {
	lockFilePath string
	lock         *flock.Flock
}

func New(filePath string) (*Lock, error) {
	lockFilePath, err := obtainLockFilePath(filePath)
	if err != nil {
		return nil, err
	}

	fileLock := flock.New(lockFilePath)

	res, err := retryWithBackoff(
		func() (bool, error) {
			return true, fileLock.Lock()
		},
		10,
		time.Millisecond,
		2.0,
	)

	if err != nil {
		return nil, err
	}

	if !res {
		return nil, fmt.Errorf("failed to obtain lock for file")
	}

	return &Lock{
		lock: fileLock,
	}, nil
}

func (l *Lock) Release() error {
	return l.lock.Unlock()
}

func obtainLockFilePath(filePath string) (string, error) {
	hash := xxhash.New64()
	_, err := hash.WriteString(filePath)
	if err != nil {
		return "", err
	}
	sum := strconv.FormatUint(hash.Sum64(), 10)
	lockFilePath := filepath.Join(baseLockDir, sum+".lock")

	return lockFilePath, nil
}

func retryWithBackoff(
	fn func() (bool, error),
	maxRetries int,
	initialInterval time.Duration,
	factor float64,
) (bool, error) {
	interval := initialInterval

	for i := 0; i < maxRetries; i++ {
		success, err := fn()

		if err != nil {
			return false, err
		}

		if success {
			return true, nil
		}

		if i < maxRetries-1 {
			time.Sleep(interval)
			interval = time.Duration(float64(interval) * factor)
		}
	}

	return false, nil
}

func init() {
	if runtime.GOOS == "windows" {
		baseLockDir = filepath.Join(os.TempDir(), "kvf-locks")
	} else {
		// Use user's cache directory or temp directory for locks
		// to avoid permission issues with /var/lock
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			baseLockDir = filepath.Join(os.TempDir(), "kvf-locks")
		} else {
			baseLockDir = filepath.Join(cacheDir, "kvf-locks")
		}
	}

	err := os.MkdirAll(baseLockDir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
}
