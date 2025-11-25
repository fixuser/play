package redlock

import (
	"time"
)

// LockOptions 包含获取锁时的各种选项
type LockOptions struct {
	TTL        time.Duration // 锁的过期时间
	MaxRetries int           // 获取锁的最大重试次数
	RetryDelay time.Duration // 每次重试之间的延迟时间
}

// Option 定义了一个函数类型，用于设置 LockOptions
type Option func(*LockOptions)

// WithTtl 设置锁的过期时间
func WithTtl(ttl time.Duration) Option {
	return func(o *LockOptions) {
		o.TTL = ttl
	}
}

// WithMaxRetries 设置获取锁的最大重试次数
func WithMaxRetries(retries int) Option {
	return func(o *LockOptions) {
		o.MaxRetries = retries
	}
}

// WithRetryDelay 设置每次重试之间的延迟时间
func WithRetryDelay(delay time.Duration) Option {
	return func(o *LockOptions) {
		o.RetryDelay = delay
	}
}

// defaultLockOptions 返回默认的锁选项
func defaultLockOptions() *LockOptions {
	return &LockOptions{
		TTL:        3 * time.Second, // 默认TTL 3秒
		MaxRetries: 3,               // 默认重试3次
		RetryDelay: 100 * time.Millisecond, // 默认重试间隔100毫秒
	}
}