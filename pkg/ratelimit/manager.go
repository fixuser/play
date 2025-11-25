package ratelimit

import (
	"context"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goccy/go-json"

	"github.com/play/play/pkg/meta"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// RateLimiterManager 从 Redis 配置动态管理限流器
type RateLimiterManager struct {
	rdb       redis.Cmdable
	configKey string
	interval  time.Duration
	configs   atomic.Value           // 存储 []*RateLimitConfig，按顺序检查
	limiters  map[string]RateLimiter // key: 限流器键（config_name:method:keyValue）
	limiterMu sync.RWMutex           // 保护 limiters
	ctx       context.Context
	cancel    context.CancelFunc
	useRedis  bool
}

// ManagerOption RateLimiterManager 的函数式选项
type ManagerOption func(*RateLimiterManager)

// WithInterval 设置配置重载间隔
func WithInterval(interval time.Duration) ManagerOption {
	return func(m *RateLimiterManager) {
		m.interval = interval
	}
}

// WithConfigKey 设置配置在 Redis 中的键名
func WithConfigKey(key string) ManagerOption {
	return func(m *RateLimiterManager) {
		m.configKey = key
	}
}

// WithRedis 启用基于 Redis 的限流
func WithRedis() ManagerOption {
	return func(m *RateLimiterManager) {
		m.useRedis = true
	}
}

// NewManager 创建一个新的 RateLimiterManager
func NewManager(rdb redis.Cmdable, opts ...ManagerOption) *RateLimiterManager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &RateLimiterManager{
		rdb:       rdb,
		configKey: defaultConfigKey,
		interval:  defaultInterval,
		limiters:  make(map[string]RateLimiter),
		ctx:       ctx,
		cancel:    cancel,
		useRedis:  false,
	}

	// 初始化 configs 为空切片
	m.configs.Store([]*RateLimitConfig{})

	for _, opt := range opts {
		opt(m)
	}

	// 加载初始配置
	if err := m.loadConfig(); err != nil {
		log.Error().Err(err).Msg("加载初始限流配置失败")
	}

	// 启动后台配置重载循环
	go m.reloadConfigLoop()

	return m
}

// loadConfig 从 Redis 加载配置
func (m *RateLimiterManager) loadConfig() error {
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	// 从 Redis hash 获取所有限流配置
	configs, err := m.rdb.HGetAll(ctx, m.configKey).Result()
	if err != nil {
		return err
	}

	newConfigs := make([]*RateLimitConfig, 0, len(configs))

	for name, configJSON := range configs {
		var cfg RateLimitConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			log.Error().Err(err).Str("name", name).Msg("解析限流配置失败")
			continue
		}

		// 编译路径正则表达式
		pathRegexp, err := regexp.Compile(cfg.Path)
		if err != nil {
			log.Error().Err(err).Str("name", name).Str("path", cfg.Path).Msg("路径正则表达式无效")
			continue
		}
		cfg.pathRegexp = pathRegexp

		// 构建方法映射以便快速查找
		cfg.methodsMap = make(map[string]bool)
		for _, method := range cfg.Methods {
			cfg.methodsMap[method] = true
		}

		// 存储配置名称，用于后续构建 Redis 键
		cfg.name = name

		newConfigs = append(newConfigs, &cfg)
		log.Debug().Str("name", name).Str("path", cfg.Path).Int("count", cfg.Count).Dur("duration", cfg.Duration).Msg("已加载限流配置")
	}

	m.configs.Store(newConfigs)

	log.Info().Int("count", len(newConfigs)).Msg("限流配置加载完成")
	return nil
}

// reloadConfigLoop 定期从 Redis 重新加载配置
func (m *RateLimiterManager) reloadConfigLoop() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if err := m.loadConfig(); err != nil {
				log.Error().Err(err).Msg("重新加载限流配置失败")
			}
		}
	}
}

// getKeyValue 根据键类型从上下文中提取键值
func getKeyValue(ctx context.Context, keyType string) string {
	switch keyType {
	case "device-id":
		return meta.Get[string](ctx, meta.MetaDeviceId)
	case "user-ip":
		return meta.Get[string](ctx, meta.MetaUserIp)
	case "token":
		return meta.Get[string](ctx, meta.HeaderToken)
	default:
		return ""
	}
}

// Limit 检查请求是否应该被限流
// 返回 true 表示应该限流
func (m *RateLimiterManager) Limit(ctx context.Context) bool {
	// 从上下文获取路径和方法
	path := meta.Get[string](ctx, meta.HeaderRequestPath)
	method := meta.Get[string](ctx, meta.HeaderRequestMethod)

	if path == "" || method == "" {
		return false
	}

	// 无锁加载配置（原子操作）
	configs := m.configs.Load().([]*RateLimitConfig)

	// 检查每个配置以找到匹配的规则
	for _, cfg := range configs {
		// 检查方法是否匹配
		if len(cfg.methodsMap) > 0 && !cfg.methodsMap[method] {
			continue
		}

		// 检查路径是否匹配
		if !cfg.pathRegexp.MatchString(path) {
			continue
		}

		// 如果指定了键类型，则获取键值
		var keyValue string
		if cfg.Key != "" {
			keyValue = getKeyValue(ctx, cfg.Key)
			if keyValue == "" {
				log.Debug().Str("path", path).Str("method", method).Str("key", cfg.Key).Msg("上下文中未找到限流键")
			}
		}

		// 构建限流器键
		limiterKey := cfg.name + ":" + method
		if keyValue != "" {
			limiterKey += ":" + keyValue
		}
		limiter := m.getLimiter(limiterKey, cfg)

		// 应用限流
		if limiter.Limit(ctx) {
			log.Warn().Str("path", path).Str("method", method).Str("key", cfg.Key).Str("key_value", keyValue).Str("limiterKey", limiterKey).Msg("超出限流限制")
			return true
		}

	}

	return false
}

// getLimiter 获取或创建指定键的限流器
func (m *RateLimiterManager) getLimiter(limiterKey string, cfg *RateLimitConfig) RateLimiter {
	// 首先使用读锁检查
	m.limiterMu.RLock()
	if limiter, ok := m.limiters[limiterKey]; ok {
		m.limiterMu.RUnlock()
		return limiter
	}
	m.limiterMu.RUnlock()

	// 需要创建，获取写锁
	m.limiterMu.Lock()
	defer m.limiterMu.Unlock()

	// 获取写锁后再次检查（双重检查）
	if limiter, ok := m.limiters[limiterKey]; ok {
		return limiter
	}

	// 创建新的限流器
	var limiter RateLimiter
	if m.useRedis {
		redisKey := "ratelimit:" + limiterKey
		limiter = NewRedis(m.rdb, redisKey, cfg.Count, cfg.Duration)
	} else {
		limiter = NewMemory(cfg.Count, cfg.Duration)
	}

	m.limiters[limiterKey] = limiter
	return limiter
}
