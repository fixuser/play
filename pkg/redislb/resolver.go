package redislb

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goccy/go-json"
	"github.com/play/play/pkg/compile"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
)

// Address 表示一个服务地址，不依赖 gRPC
type Address struct {
	Addr       string         // 地址，格式为 host:port
	Attributes map[string]any // 附加属性
	ServerName string         // 服务名称
}

// Resolver 是一个不依赖 gRPC 的公共解析器结构体
type Resolver struct {
	mu            sync.RWMutex
	rdb           redis.Cmdable
	serviceName   string
	whitelistNets []*net.IPNet
	scheme        string // 协议类型，如 "grpc", "http" 等
}

// NewResolver 创建一个新的 Resolver
func NewResolver(rdb redis.Cmdable, serviceName string, whitelistNets []*net.IPNet, scheme string) *Resolver {
	if scheme == "" {
		scheme = "grpc" // 默认为 grpc
	}
	return &Resolver{
		rdb:           rdb,
		serviceName:   serviceName,
		whitelistNets: whitelistNets,
		scheme:        scheme,
	}
}

// isWhitelist returns true if the given IP is in the whitelist.
func (r *Resolver) isWhitelist(ip net.IP) bool {
	if len(r.whitelistNets) == 0 {
		return true
	}

	for _, net := range r.whitelistNets {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}

// Resolve 从 Redis 解析服务地址列表
func (r *Resolver) Resolve(ctx context.Context) ([]Address, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := "services:" + r.serviceName + ":*"
	keys, err := r.rdb.Keys(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	if len(keys) == 0 {
		return []Address{}, nil
	}

	values, err := r.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get values: %w", err)
	}

	var addrs []Address
	for _, item := range values {
		if item == nil {
			continue
		}
		value := item.(string)
		var nodeInfo compile.NodeInfo
		err := json.Unmarshal([]byte(value), &nodeInfo)
		if err != nil {
			continue
		}

		for _, endpoint := range nodeInfo.Endpoints {
			if endpoint.Scheme != r.scheme {
				continue
			}
			ipAddr, _, _ := net.SplitHostPort(endpoint.Host)
			ip := net.ParseIP(ipAddr)
			if ip == nil {
				continue
			}
			// 检查IP是否在白名单中
			if !r.isWhitelist(ip) {
				continue
			}

			addrs = append(addrs, Address{
				Addr: endpoint.Host,
				Attributes: map[string]any{
					"id": nodeInfo.Id,
				},
				ServerName: r.serviceName,
			})
		}
	}

	return addrs, nil
}

var _ resolver.Resolver = (*schemaResolver)(nil)

// schemaResolver 是基于 gRPC 的解析器，内嵌 Resolver
type schemaResolver struct {
	*Resolver
	mu          sync.RWMutex
	watchTicker *time.Ticker
	logger      grpclog.LoggerV2
	client      resolver.ClientConn
	isClosed    atomic.Bool
	rnChan      chan struct{}
	closeChan   chan struct{}
}

func (r *schemaResolver) watcher() {
	for !r.isClosed.Load() {
		select {
		case <-r.watchTicker.C:
			r.resolveNow(true)
		case <-r.rnChan:
			r.resolveNow(true)
		case <-r.closeChan:
			return
		}
	}
}

func (r *schemaResolver) ResolveNow(opts resolver.ResolveNowOptions) {
	r.rnChan <- struct{}{}
}

func (r *schemaResolver) resolveNow(force bool) {
	if r.isClosed.Load() {
		return
	}

	ctx := context.Background()
	// 使用公共 Resolver 的 Resolve 方法
	addrs, err := r.Resolver.Resolve(ctx)
	if err != nil {
		r.logger.Warningf("failed to resolve: %v", err)
		r.client.ReportError(fmt.Errorf("failed to resolve: %v", err))
		return
	}

	r.updateState(addrs, force)
}

func (r *schemaResolver) updateState(addrs []Address, force bool) {
	// 转换为 gRPC resolver.Address
	grpcAddrs := make([]resolver.Address, 0, len(addrs))
	for _, addr := range addrs {
		grpcAddr := resolver.Address{
			Addr:       addr.Addr,
			ServerName: addr.ServerName,
		}
		// 转换属性
		if len(addr.Attributes) > 0 {
			for k, v := range addr.Attributes {
				grpcAddr.Attributes = attributes.New(k, v)
			}
		}
		grpcAddrs = append(grpcAddrs, grpcAddr)
	}

	if len(grpcAddrs) == 0 && !force {
		r.client.ReportError(fmt.Errorf("no valid service registeration for %s://%s", schema, r.serviceName))
		return
	}

	err := r.client.UpdateState(resolver.State{
		Addresses: grpcAddrs,
	})
	if err != nil {
		r.client.ReportError(fmt.Errorf("failed to update state: %v", err))
	}
}

func (r *schemaResolver) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isClosed.CompareAndSwap(false, true) {
		r.watchTicker.Stop()
		close(r.closeChan)
	}
}
