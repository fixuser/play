package redislb

import (
	"sync"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/resolver"
)

const schema = "redislb"

var _registryOnce sync.Once

func RegisterSchema(rds redis.Cmdable, opts ...BuilderOptApplyFn) {
	_registryOnce.Do(func() {
		resolver.Register(NewBuilder(rds, opts...))
	})
}
