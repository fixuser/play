/*
简单、线程安全的 Go 限流器。
受 Antti Huima 在 http://stackoverflow.com/a/668327 上的算法启发

示例：

	// 创建一个新的限流器，每秒允许最多 10 次调用
	rl := ratelimit.NewMemory(10, time.Second)

	for i:=0; i<20; i++ {
	  if rl.Limit() {
	    fmt.Println("DOH! Over limit!")
	  } else {
	    fmt.Println("OK")
	  }
	}
*/
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// MemoryRateLimit 实例是线程安全的
type MemoryRateLimit struct {
	mu sync.Mutex

	rate, allowance, max, unit uint64
	lastCheck                  int64
}

// NewMemory 创建一个新的基于内存的限流器实例
func NewMemory(rate int, per time.Duration) *MemoryRateLimit {
	nano := uint64(per)
	if nano < 1 {
		nano = uint64(time.Second)
	}
	if rate < 1 {
		rate = 1
	}

	return &MemoryRateLimit{
		rate:      uint64(rate),        // 存储限流频率
		allowance: uint64(rate) * nano, // 开始时将配额设置为最大值
		max:       uint64(rate) * nano, // 记录最大配额
		unit:      nano,                // 记录单位大小

		lastCheck: nowNano(),
	}
}

// UpdateRate 允许更新限流频率
func (rl *MemoryRateLimit) UpdateRate(rate int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.rate = uint64(rate)
	rl.max = uint64(rate) * rl.unit
}

// Limit 超过限流时返回 true
func (rl *MemoryRateLimit) Limit(_ context.Context) bool {
	now := nowNano()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 计算自上次调用以来经过的纳秒数
	passed := now - rl.lastCheck
	rl.lastCheck = now

	// 将经过的时间添加到我们的配额
	rl.allowance += uint64(passed) * uint64(rl.rate)
	current := rl.allowance

	// 确保我们的配额不超过最大值
	if current > rl.max {
		rl.allowance += ^((current - rl.max) - 1)
		current = rl.max
	}

	// 如果我们的配额小于一个单位，则限流！
	if current < rl.unit {
		return true
	}

	// 未限流，减去一个单位
	rl.allowance += ^(rl.unit - 1)
	return false
}

// Undo 撤销上一次 Limit() 调用，返还消耗的配额
func (rl *MemoryRateLimit) Undo() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.allowance += rl.unit

	// 确保我们的配额不超过最大值
	if current := rl.allowance; current > rl.max {
		rl.allowance += ^((current - rl.max) - 1)
	}
}

func nowNano() int64 {
	return time.Now().UnixNano()
}
