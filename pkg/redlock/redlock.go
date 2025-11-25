package redlock

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// unlockScript 是一个Lua脚本，用于原子性地检查值并删除键
// 只有当键的值与传入的value匹配时，才删除键
const unlockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end
`

// RedisLocker 是一个Redis分布式锁的管理器
type RedisLocker struct {
	client         redis.Cmdable // Redis客户端接口
	defaultOptions *LockOptions  // 默认的锁选项
}

// Locker 定义了分布式锁的接口
type Locker interface {
	TryLock(ctx context.Context) (bool, error) // 尝试获取锁，不重试
	Lock(ctx context.Context) error            // 获取锁，带重试机制
	Unlock(ctx context.Context) (bool, error)  // 释放锁
	Value() string                             // 返回锁的值
	Key() string                               // 返回锁的键
}

// lock 实现了Locker接口，表示一个具体的锁实例
type lock struct {
	key     string        // 锁的键
	value   string        // 锁的值，用于区分不同实例持有的锁
	client  redis.Cmdable // Redis客户端
	options *LockOptions  // 当前锁实例的选项
}

// NewRedLock 创建一个新的RedisLocker实例
// client: Redis客户端，需要传入 redis.NewClient 或 redis.NewClusterClient 的返回值
// opts: 可选的全局默认锁选项
func NewRedLock(client redis.Cmdable, opts ...Option) *RedisLocker {
	if client == nil {
		log.Fatal().Msg("redis client cannot be nil")
	}

	defaultOpts := defaultLockOptions()
	for _, opt := range opts {
		opt(defaultOpts)
	}

	return &RedisLocker{
		client:         client,
		defaultOptions: defaultOpts,
	}
}

// Locker 返回一个具体的Locker实例，用于操作特定key的锁
// key: 锁的键名
// opts: 可选的、覆盖全局默认的锁选项
func (rl *RedisLocker) Locker(key string, opts ...Option) Locker {
	if key == "" {
		log.Error().Err(ErrInvalidArguments).Msg("lock key cannot be empty")
		return nil // 或者 panic，取决于错误处理策略
	}

	// 复制一份默认选项，然后应用传入的特定选项
	lockOpts := &LockOptions{
		TTL:        rl.defaultOptions.TTL,
		MaxRetries: rl.defaultOptions.MaxRetries,
		RetryDelay: rl.defaultOptions.RetryDelay,
	}
	for _, opt := range opts {
		opt(lockOpts)
	}

	return &lock{
		key:     key,
		value:   uuid.NewString(),
		client:  rl.client,
		options: lockOpts,
	}
}

// TryLock 尝试获取锁，不进行重试。如果成功获取锁，则返回true；否则返回false。
func (l *lock) TryLock(ctx context.Context) (bool, error) {
	// 使用SET NX PX 命令获取锁
	// NX: Only set the key if it does not already exist.
	// PX: Set the specified expire time, in milliseconds.
	cmd := l.client.SetNX(ctx, l.key, l.value, l.options.TTL)
	acquired, err := cmd.Result()

	if err != nil && err != redis.Nil { // redis.Nil 表示key不存在，不是错误
		log.Ctx(ctx).Error().Err(err).Str("key", l.key).Msg("failed to setnx for lock")
		return false, err
	}

	if acquired {
		log.Ctx(ctx).Debug().Str("key", l.key).Str("value", l.value).Dur("ttl", l.options.TTL).Msg("lock acquired successfully")
		return true, nil
	}

	log.Ctx(ctx).Debug().Str("key", l.key).Msg("lock already held by another instance or expired")
	return false, nil
}

// Lock 尝试获取锁，如果未能立即获取，则进行重试，直到成功或达到最大重试次数/上下文取消。
func (l *lock) Lock(ctx context.Context) error {
	log.Ctx(ctx).Debug().Str("key", l.key).Int("max_retries", l.options.MaxRetries).Dur("retry_delay", l.options.RetryDelay).Msg("attempting to acquire lock")

	for i := 0; i <= l.options.MaxRetries; i++ {
		acquired, err := l.TryLock(ctx)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("key", l.key).Int("attempt", i+1).Msg("error during lock attempt")
			return err // 如果是Redis本身的错误，直接返回
		}

		if acquired {
			return nil // 成功获取锁
		}

		// 如果不是最后一次尝试，则等待一段时间进行重试
		if i < l.options.MaxRetries {
			log.Ctx(ctx).Debug().Str("key", l.key).Int("attempt", i+1).Dur("delay", l.options.RetryDelay).Msg("lock not acquired, retrying after delay")
			select {
			case <-ctx.Done(): // 检查上下文是否已取消
				log.Ctx(ctx).Warn().Str("key", l.key).Err(ctx.Err()).Msg("context cancelled while waiting for lock")
				return ctx.Err()
			case <-time.After(l.options.RetryDelay): // 等待重试延迟
				// 继续下一次循环
			}
		}
	}

	log.Ctx(ctx).Warn().Str("key", l.key).Msg("failed to acquire lock after all retries")
	return ErrFailedToAcquireLock // 超过重试次数，未能获取锁
}

// Unlock 释放锁。只有当锁的value与当前实例的value匹配时才会被释放。
// 返回true表示成功释放锁，false表示锁不属于当前实例或已过期，error表示Redis操作错误。
func (l *lock) Unlock(ctx context.Context) (bool, error) {
	// 使用Lua脚本原子性地检查并删除键
	// KEYS[1] 是键名， ARGV[1] 是要匹配的值
	cmd := l.client.Eval(ctx, unlockScript, []string{l.key}, l.value)
	result, err := cmd.Result()

	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("key", l.key).Str("value", l.value).Msg("failed to execute unlock script")
		return false, err
	}

	if i, ok := result.(int64); ok && i == 1 {
		log.Ctx(ctx).Debug().Str("key", l.key).Str("value", l.value).Msg("lock released successfully")
		return true, nil
	}

	log.Ctx(ctx).Warn().Str("key", l.key).Str("value", l.value).Msg("lock not released, value mismatched or key already expired/deleted")
	return false, ErrLockNotHeld // 值不匹配或键已不存在
}

// Value 返回当前锁实例的随机值
func (l *lock) Value() string {
	return l.value
}

// Key 返回当前锁实例的键名
func (l *lock) Key() string {
	return l.key
}
