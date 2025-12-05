// Meta 用于在 context.Context 中安全地存储和传递元数据
// 线程安全，支持并发读写
package meta

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
	"google.golang.org/grpc/metadata"
)

// metadataKey 用于 context.WithValue 的唯一 key，防止冲突
// 仅在本包内部使用
type metadataKey struct{}

// Meta 结构体用于存储和管理元数据
type Meta struct {
	mu   sync.RWMutex    // 保护 data 的并发安全
	ctx  context.Context // 关联的 context
	data map[string]any  // 存储元数据的 map
}

// FromContext 从 context.Context 提取 gRPC metadata 并构造 Meta
// 如果 context 中没有 metadata，则返回空 Meta
func FromContext(ctx context.Context) (meta *Meta) {
	meta = &Meta{
		ctx:  ctx,
		data: make(map[string]any),
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Warn().Msg("no grpc metadata found in context")
		return
	}

	for k, v := range md {
		if len(v) > 0 {
			meta.mu.Lock()
			meta.data[k] = v[0]
			meta.mu.Unlock()
		} else {
			log.Trace().Str("key", k).Msg("metadata value is empty")
		}
	}
	return meta
}

// Set 设置指定 key 的元数据，支持链式调用
func (m *Meta) Set(key string, value any) *Meta {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.data == nil {
		log.Trace().Msg("meta.data is nil, initializing new map")
		m.data = make(map[string]any)
	}
	m.data[key] = value
	return m
}

// Get 获取指定 key 的元数据，如果不存在则返回 nil
func (m *Meta) Get(key string) (value any) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.data == nil {
		log.Trace().Msg("meta.data is nil in Get")
		return nil
	}
	value = m.data[key]
	return
}

// GetString 获取指定 key 的字符串类型元数据
func (m *Meta) GetString(key string) (value string) {
	val := m.Get(key)
	return cast.ToString(val)
}

// GetInt64 获取指定 key 的 int64 类型元数据
func (m *Meta) GetInt64(key string) (value int64) {
	val := m.Get(key)
	return cast.ToInt64(val)
}

// Context 返回带有当前元数据的 context.Context
func (m *Meta) Context() (ctx context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ctx == nil {
		log.Trace().Msg("meta.ctx is nil in Context")
		m.ctx = context.Background()
	}
	return context.WithValue(m.ctx, metadataKey{}, m)
}

// Get 从 context.Context 中获取指定 key 的元数据
// 如果不存在则返回零值
func Get[T any](ctx context.Context, key string) (value T) {
	m, ok := ctx.Value(metadataKey{}).(*Meta)
	if !ok {
		log.Warn().Msg("context does not contain Meta")
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.data == nil {
		log.Trace().Msg("meta.data is nil in Get")
		return
	}
	v, ok := m.data[key].(T)
	if !ok {
		log.Warn().Str("key", key).Msg("type assertion failed")
		return
	}
	value = v
	return
}
