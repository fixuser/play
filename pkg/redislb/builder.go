package redislb

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
)

type BuilderOptApplyFn func(*builder)

func WithLogger(logger grpclog.LoggerV2) BuilderOptApplyFn {
	return func(builder *builder) {
		builder.logger = logger
	}
}

func WithIPv4Whitelist(subnets []string) BuilderOptApplyFn {
	return func(builder *builder) {
		builder.whitelistSubnets = subnets
	}
}

var _ resolver.Builder = (*builder)(nil)

type builder struct {
	rds              redis.Cmdable
	logger           grpclog.LoggerV2
	whitelistSubnets []string
}

func NewBuilder(rds redis.Cmdable, opts ...BuilderOptApplyFn) *builder {
	b := &builder{
		rds: rds,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

func (b *builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	if !strings.EqualFold(schema, target.URL.Scheme) {
		return nil, fmt.Errorf("unexpected schema: %s", target.URL.Scheme)
	}

	whitelistNets := make([]*net.IPNet, 0, len(b.whitelistSubnets))
	for _, subnet := range b.whitelistSubnets {
		_, net, err := net.ParseCIDR(subnet)
		if err == nil {
			whitelistNets = append(whitelistNets, net)
		}
	}

	// 创建公共的 Resolver
	commonResolver := NewResolver(b.rds, target.URL.Host, whitelistNets, "grpc")

	// 创建 schemaResolver，内嵌公共 Resolver
	res := &schemaResolver{
		Resolver:    commonResolver,
		watchTicker: time.NewTicker(time.Second * 2),
		client:      cc,
		rnChan:      make(chan struct{}),
		closeChan:   make(chan struct{}),
	}

	if b.logger == nil {
		res.logger = grpclog.Component(b.Scheme())
	} else {
		res.logger = b.logger
	}

	if res.serviceName != "" {
		go res.watcher()
		res.resolveNow(false)
	}
	return res, nil
}

func (b *builder) Scheme() string {
	return schema
}
