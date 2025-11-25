package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimit 基于 Redis 的限流器，使用滑动窗口算法
type RedisRateLimit struct {
	rdb      redis.Cmdable
	key      string
	rate     int
	duration time.Duration
}

// NewRedis 创建一个新的基于 Redis 的限流器实例
func NewRedis(rdb redis.Cmdable, key string, rate int, duration time.Duration) *RedisRateLimit {
	if rate < 1 {
		rate = 1
	}
	if duration < 1 {
		duration = time.Second
	}
	return &RedisRateLimit{
		rdb:      rdb,
		key:      key,
		rate:     rate,
		duration: duration,
	}
}

// Limit 使用 Redis 滑动窗口检查是否超过限流，超过返回 true
func (rl *RedisRateLimit) Limit(ctx context.Context) bool {
	now := time.Now()
	windowStart := now.Add(-rl.duration)

	key := rl.key
	pipe := rl.rdb.Pipeline()

	// 移除当前窗口外的旧条目
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// 计算窗口内的当前条目数
	zcount := pipe.ZCount(ctx, key, "-inf", "+inf")

	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false // 出错时允许请求通过
	}

	count := zcount.Val()

	// 检查是否超过限制
	if count >= int64(rl.rate) {
		return true
	}

	// 添加当前请求
	pipe = rl.rdb.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})
	pipe.Expire(ctx, key, rl.duration+time.Second)
	_, _ = pipe.Exec(ctx)

	return false
}

// UpdateRate 更新限流频率
func (rl *RedisRateLimit) UpdateRate(rate int) {
	if rate < 1 {
		rate = 1
	}
	rl.rate = rate
}
