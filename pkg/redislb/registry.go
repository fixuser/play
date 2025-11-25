package redislb

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/play/play/pkg/compile"
	"github.com/redis/go-redis/v9"
)

type Registry struct {
	rdb               redis.Cmdable
	ttl               time.Duration
	KeepAliveDuration time.Duration
	isClosed          atomic.Bool
	closeChan         chan struct{}
}

type RegistryOptionApplyFn func(*Registry)

func WithTtl(ttl time.Duration) RegistryOptionApplyFn {
	return func(r *Registry) {
		r.ttl = ttl
	}
}

func WithKeepAliveDuration(d time.Duration) RegistryOptionApplyFn {
	return func(r *Registry) {
		r.KeepAliveDuration = d
	}
}

func NewRegistry(rdb redis.Cmdable, opts ...RegistryOptionApplyFn) *Registry {
	r := &Registry{
		ttl:               10 * time.Second,
		KeepAliveDuration: 3 * time.Second,
		closeChan:         make(chan struct{}),
	}

	for _, opt := range opts {
		opt(r)
	}
	r.rdb = rdb
	return r
}

func (r *Registry) Register() error {
	if r.ttl <= r.KeepAliveDuration {
		return errors.New("registry TTL should be greater than registry keepalive duration")
	}

	go func() {
		ticker := time.NewTicker(r.KeepAliveDuration)
		ctx := context.Background()

		for {
			select {
			case <-ticker.C:
				r.rdb.SetEx(ctx, compile.Node.Key(), compile.Node.Value(), r.ttl)
			case <-r.closeChan:
				ticker.Stop()
				r.rdb.Del(ctx, compile.Node.Key())
				return
			}
		}
	}()
	return nil
}

func (r *Registry) Close() {
	if r.isClosed.CompareAndSwap(false, true) {
		close(r.closeChan)
	}
}
