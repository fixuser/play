package ratelimit

import (
	"context"
	"regexp"
	"time"
)

const (
	defaultConfigKey = "ratelimit:config" // 默认配置在 Redis 中的键名
	defaultInterval  = time.Minute        // 默认配置重载间隔
)

// RateLimiter 限流器接口
type RateLimiter interface {
	Limit(ctx context.Context) bool
}

// RateLimitConfig 特定路径的限流配置
type RateLimitConfig struct {
	Path     string        `json:"path"`     // 路径模式（支持正则表达式）
	Methods  []string      `json:"methods"`  // HTTP 方法列表 ["GET", "POST" 等]
	Count    int           `json:"count"`    // 允许的请求数量
	Duration time.Duration `json:"duration"` // 时间窗口（如 "1s", "1m"）
	Key      string        `json:"key"`      // 限流键类型: "device-id", "user-ip", "token"，为空表示仅按 path+method 限流

	// 内部字段
	pathRegexp *regexp.Regexp
	methodsMap map[string]bool
	name       string // 配置名称，用于构建 Redis 键
}
