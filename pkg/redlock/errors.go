package redlock

import "errors"

var (
	// ErrFailedToAcquireLock 表示在多次重试后未能获取到锁
	ErrFailedToAcquireLock = errors.New("failed to acquire lock after multiple retries")
	// ErrLockNotHeld 表示当前实例并未持有该锁，或者锁已过期
	ErrLockNotHeld = errors.New("lock not held by this instance or already expired")
	// ErrInvalidArguments 表示提供了无效的参数
	ErrInvalidArguments = errors.New("invalid arguments provided")
)