package buncache

import (
	"context"
	"math/rand/v2"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

type Cache[K comparable, V any] struct {
	Load   func()
	Get    func(k K) *V
	Gets   func(ks ...K) []*V
	Delete func(ks ...K)
	Flush  func()
}

type option[K comparable, V any] func(*CacheManager[K, V])

type CacheManager[K comparable, V any] struct {
	db           *bun.DB
	isLoad       atomic.Bool
	fieldKey     string
	sqlKey       string
	size         int
	interval     time.Duration
	loadInterval time.Duration
	modelName    string
	caches       *expirable.LRU[K, *V]
}

// WithInterval 设置缓存更新间隔
func WithInterval[K comparable, V any](d time.Duration) option[K, V] {
	return func(cm *CacheManager[K, V]) {
		cm.interval = d
	}
}

// WithFieldKey 设置实体的主键字段名，默认为 "Id"
func WithFieldKey[K comparable, V any](key string) option[K, V] {
	return func(cm *CacheManager[K, V]) {
		cm.fieldKey = key
	}
}

// WithSize 设置缓存大小，默认为 20000000
func WithSize[K comparable, V any](size int) option[K, V] {
	return func(cm *CacheManager[K, V]) {
		cm.size = size
	}
}

// WithLoadInterval 设置定时加载间隔，如果设置则会定时调用 Load 函数
func WithLoadInterval[K comparable, V any](d time.Duration) option[K, V] {
	return func(cm *CacheManager[K, V]) {
		cm.loadInterval = d
	}
}

func New[K comparable, V any](db *bun.DB, opts ...option[K, V]) *CacheManager[K, V] {
	table := db.Table(reflect.TypeOf(new(V)))
	if table == nil {
		panic("buncache: model is not registered, please register it before using cache")
	}

	key := "Id"
	if len(table.PKs) == 1 {
		key = table.PKs[0].StructField.Name
	}

	cm := &CacheManager[K, V]{
		db:        db,
		fieldKey:  key,
		size:      20000000,
		modelName: table.Name,
		interval:  time.Minute*2 + time.Duration(rand.Int64N(60))*time.Second,
	}
	for _, opt := range opts {
		opt(cm)
	}

	found := false
	for _, field := range table.Fields {
		if field.StructField.Name == cm.fieldKey {
			found = true
			cm.sqlKey = field.Name
			break
		}
	}
	if !found {
		panic("buncache: specified key field(" + cm.fieldKey + ") does not exist in the model")
	}

	cm.caches = expirable.NewLRU[K, *V](cm.size, nil, cm.interval)
	return cm
}

// buildLoad 构建加载所有数据到缓存的函数
// 使用行扫描方式，避免一次性加载所有数据到内存
func (cm *CacheManager[K, V]) buildLoad() func() {
	return func() {
		ctx := context.Background()

		rows, err := cm.db.NewSelect().Model((*V)(nil)).Rows(ctx)
		if err != nil {
			log.Warn().Str("model", cm.modelName).Err(err).Msg("buncache: failed to query rows")
			return
		}
		defer rows.Close()

		// 记录历史数据量
		oldLen := cm.caches.Len()

		// 清除历史数据
		cm.caches.Purge()

		// 逐行扫描并加入缓存
		loadedCount := 0
		for rows.Next() {
			var v V
			if err := cm.db.ScanRow(ctx, rows, &v); err != nil {
				log.Warn().Str("model", cm.modelName).Err(err).Msg("buncache: failed to scan row")
				// 扫描单行失败，跳过继续处理下一行
				continue
			}

			// 获取主键值
			keyVal := reflect.ValueOf(v).FieldByName(cm.fieldKey).Interface().(K)
			vPtr := new(V)
			*vPtr = v
			cm.caches.Add(keyVal, vPtr)
			loadedCount++
		}

		// 检查迭代过程中的错误
		if err := rows.Err(); err != nil {
			log.Warn().Str("model", cm.modelName).Err(err).Msg("buncache: error occurred during rows iteration")
			return
		}

		// 记录加载信息
		log.Info().
			Str("model", cm.modelName).
			Int("old_count", oldLen).
			Int("loaded_count", loadedCount).
			Msg("buncache: load completed")

		// 标记已加载
		cm.isLoad.Store(true)
	}
}

// loop 定时加载数据到缓存
func (cm *CacheManager[K, V]) loop(loadFunc func()) {
	if cm.loadInterval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(cm.loadInterval)
		defer ticker.Stop()

		loadFunc()
		for range ticker.C {
			loadFunc()
		}
	}()
}

// Build 构建缓存管理器
func (cm *CacheManager[K, V]) Build() (cache *Cache[K, V]) {
	cache = &Cache[K, V]{}
	cache.Load = cm.buildLoad()
	cache.Get = cm.buildGet()
	cache.Gets = cm.buildGets()
	cache.Delete = cm.buildDelete()
	cache.Flush = cm.buildFlush()

	// 启动定时加载
	cm.loop(cache.Load)

	return
}

// buildGet 构建单个实体缓存获取函数
func (cm *CacheManager[K, V]) buildGet() func(k K) *V {
	return func(k K) (value *V) {
		ctx := context.Background()
		v, ok := cm.caches.Get(k)
		if ok {
			// 如果缓存中存在，直接返回（包括 nil 值）
			return v
		}

		var v2 V
		err := cm.db.NewSelect().Model(&v2).Where(cm.sqlKey+" = ?", k).Scan(ctx)
		if err != nil {
			// 查询出错或未找到记录，缓存 nil 防止缓存穿透
			cm.caches.Add(k, nil)
			return nil
		}

		// 查询成功，缓存实际值
		cm.caches.Add(k, &v2)
		value = &v2
		return
	}
}

// buildGets 构建批量实体缓存获取函数
// 返回 []*V 切片，保持与输入 ks 相同的长度和顺序
// 不存在的记录返回 nil
// 当 isLoad 为 true 且 ks 为空时，返回所有缓存数据
func (cm *CacheManager[K, V]) buildGets() func(ks ...K) []*V {
	return func(ks ...K) []*V {
		// 如果 isLoad 为 true 且 ks 为空，返回所有缓存数据
		if len(ks) == 0 {
			if cm.isLoad.Load() {
				return cm.caches.Values()
			}
			return nil
		}

		ctx := context.Background()
		result := make([]*V, len(ks))
		var missingKeys []K
		missingIndices := make(map[K][]int) // 记录每个 key 在结果中的位置

		// 第一步：从缓存获取，记录缺失的 key
		for i, k := range ks {
			v, ok := cm.caches.Get(k)
			if ok {
				// 如果缓存中存在（包括 nil），直接使用
				result[i] = v
			} else {
				// 记录缺失的 key 及其在结果中的索引
				if _, exists := missingIndices[k]; !exists {
					missingKeys = append(missingKeys, k)
				}
				missingIndices[k] = append(missingIndices[k], i)
			}
		}

		// 第二步：如果没有缺失的 key，直接返回
		if len(missingKeys) == 0 {
			return result
		}

		// 第三步：批量查询数据库
		var vs []V
		cm.db.NewSelect().Model(&vs).Where(cm.sqlKey+" IN (?)", bun.In(missingKeys)).Scan(ctx)

		// 第四步：处理查询结果
		foundKeys := make(map[K]*V)
		for _, v := range vs {
			val := v
			keyVal := reflect.ValueOf(val).FieldByName(cm.fieldKey).Interface().(K)
			ptrVal := &val
			cm.caches.Add(keyVal, ptrVal)
			foundKeys[keyVal] = ptrVal
		}

		// 第五步：填充结果并缓存未找到的 key
		for _, k := range missingKeys {
			var valuePtr *V
			if ptr, found := foundKeys[k]; found {
				valuePtr = ptr
			} else {
				// 未找到，缓存 nil 防止缓存穿透
				cm.caches.Add(k, nil)
				valuePtr = nil
			}
			// 将值填充到所有对应的位置
			for _, idx := range missingIndices[k] {
				result[idx] = valuePtr
			}
		}

		return result
	}
}

// buildDelete 构建批量实体缓存删除函数
func (cm *CacheManager[K, V]) buildDelete() func(ks ...K) {
	return func(ks ...K) {
		for _, k := range ks {
			cm.caches.Remove(k)
		}
	}
}

// buildFlush 构建清空缓存函数
func (cm *CacheManager[K, V]) buildFlush() func() {
	return func() {
		cm.caches.Purge()
	}
}
