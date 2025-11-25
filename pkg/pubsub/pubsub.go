package pubsub

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// 全局错误码
var (
	ErrQueueFull          = errors.New("queue is full")
	ErrInvalidFunction    = errors.New("invalid function provided")
	ErrArgMismatch        = errors.New("function arguments mismatch")
	ErrSubscriptionClosed = errors.New("subscription is closed")
	ErrPubSubClosed       = errors.New("pubsub is closed")
)

const (
	redisKeyPrefix      = "pubsub:topic:"
	blpopTimeout        = 1 * time.Second // BLPOP 的阻塞超时时间
	defaultQueueSize    = 1000            // 默认队列大小限制
	defaultDataChanSize = 100             // 默认内部数据通道大小
)

// Option 是用于 PubSub 或 Subscription 的配置选项函数
type Option func(any)

// PubSub 结构体
type PubSub struct {
	redisClient   redis.Cmdable
	queueSize     int
	subscriptions map[string][]*Subscription // key: topic
	mu            sync.RWMutex
	closed        chan struct{}
	wg            sync.WaitGroup // 用于等待所有 Subscription 关闭
	useRecovery   bool           // 是否开启panic recovery（全局）
}

// Subscription 结构体
type Subscription struct {
	pubSub       *PubSub
	topic        string
	redisKey     string
	handler      reflect.Value      // 订阅的函数
	handlerType  reflect.Type       // 订阅函数的类型
	concurrency  int                // 并发worker数量
	useRecovery  bool               // 是否开启panic recovery
	dataChan     chan []byte        // 内部数据通道，BLPOP将数据放入此通道
	stopChan     chan struct{}      // 通知goroutine停止
	wg           sync.WaitGroup     // 用于等待worker goroutine结束
	ctx          context.Context    // 订阅的上下文
	cancel       context.CancelFunc // 用于取消订阅的上下文
	processingMu sync.Mutex         // 确保 Stop 时不会有新的消息被处理
}

// --- PubSub 配置选项 ---

// WithQueueSize 设置Publish时检查的Redis List最大长度
func WithQueueSize(qs int) Option {
	return func(o any) {
		if ps, ok := o.(*PubSub); ok {
			if qs > 0 {
				ps.queueSize = qs
			}
		}
	}
}

// --- Subscription 配置选项 ---

// WithRecovery 使订阅的函数在发生panic时被recovery
func WithRecovery() Option {
	return func(o any) {
		switch v := o.(type) {
		case *Subscription:
			v.useRecovery = true
		case *PubSub:
			v.useRecovery = true
		}
	}
}

// WithConcurrency 设置消费函数的并发数量
// 如果 c <= 0, 则默认为 1
func WithConcurrency(c int) Option {
	return func(o any) {
		if s, ok := o.(*Subscription); ok {
			if c <= 0 {
				s.concurrency = 1
			} else {
				s.concurrency = c
			}
		}
	}
}

// New 创建一个新的 PubSub 实例
func New(redisClient redis.Cmdable, opts ...Option) *PubSub {
	ps := &PubSub{
		redisClient:   redisClient,
		queueSize:     defaultQueueSize, // 默认值
		subscriptions: make(map[string][]*Subscription),
		closed:        make(chan struct{}),
		useRecovery:   false, // 默认不recovery
	}
	for _, opt := range opts {
		opt(ps)
	}
	log.Trace().Int("queue_size", ps.queueSize).Bool("recovery", ps.useRecovery).Msg("new pubsub initialized")
	return ps
}

func formatTopicKey(topic string) string {
	return redisKeyPrefix + topic
}

// Publish 发布消息到指定的topic
// 自动检测参数类型：
// - 如果所有 args 都是 []any 类型，则作为批量发布（每个 []any 是一条消息）
// - 否则将 args 包装成 []any{args} 作为单条消息批量发布
func (p *PubSub) Publish(ctx context.Context, topic string, args ...any) error {
	select {
	case <-p.closed:
		log.Error().Str("topic", topic).Msg("cannot publish on closed pubsub")
		return ErrPubSubClosed
	default:
	}

	if len(args) == 0 {
		log.Trace().Str("topic", topic).Msg("no messages to publish")
		return nil
	}

	// 检测是否所有参数都是 []any 类型
	var argsList [][]any
	isBatch := true
	for _, arg := range args {
		if slice, ok := arg.([]any); ok {
			argsList = append(argsList, slice)
		} else {
			isBatch = false
			break
		}
	}

	// 如果不是批量模式，将 args 包装成 []any{args}
	if !isBatch {
		argsList = [][]any{args}
	}

	return p.publishBatch(ctx, topic, argsList)
}

// publishBatch 批量发布多条消息到指定的topic
// 每个 args 元素会被序列化为 JSON 数组字符串存储到 Redis List
func (p *PubSub) publishBatch(ctx context.Context, topic string, argsList [][]any) error {
	redisKey := formatTopicKey(topic)

	// 检查队列长度
	if p.queueSize > 0 {
		length, err := p.redisClient.LLen(ctx, redisKey).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			log.Error().Err(err).Str("topic", topic).Msg("failed to get list length for batch queue size check")
			return fmt.Errorf("redis LLen failed: %w", err)
		}

		// 检查批量发布后是否会超出队列大小限制
		if length+int64(len(argsList)) > int64(p.queueSize) {
			log.Warn().Str("topic", topic).Int64("current_length", length).Int("batch_size", len(argsList)).Int("queue_size_limit", p.queueSize).Msg("batch publish would exceed queue size limit")
			return ErrQueueFull
		}
	}

	// 序列化所有消息
	payloads := make([]any, 0, len(argsList))
	for i, args := range argsList {
		payload, err := json.Marshal(args)
		if err != nil {
			log.Error().Err(err).Str("topic", topic).Int("batch_index", i).Interface("args", args).Msg("failed to marshal batch publish arguments")
			return fmt.Errorf("json marshal failed for batch index %d: %w", i, err)
		}
		payloads = append(payloads, payload)
	}

	// 使用 RPush 批量推送所有消息
	err := p.redisClient.RPush(ctx, redisKey, payloads...).Err()
	if err != nil {
		log.Error().Err(err).Str("topic", topic).Int("batch_size", len(argsList)).Msg("failed to publish batch messages to redis")
		return fmt.Errorf("redis RPush batch failed: %w", err)
	}

	log.Trace().Str("topic", topic).Int("batch_size", len(argsList)).Msg("batch messages published successfully")
	return nil
}

// Subscribe 订阅一个topic
// fn 是处理消息的函数，其参数类型和数量必须与Publish时的args对应
// opts 是Subscription的配置选项
func (p *PubSub) Subscribe(ctx context.Context, topic string, fn any, opts ...Option) (*Subscription, error) {
	select {
	case <-p.closed:
		log.Error().Str("topic", topic).Msg("cannot subscribe on closed pubsub")
		return nil, ErrPubSubClosed
	default:
	}

	fnVal := reflect.ValueOf(fn)
	if fnVal.Kind() != reflect.Func {
		log.Error().Str("topic", topic).Msg("provided handler is not a function")
		return nil, ErrInvalidFunction
	}

	subCtx, subCancel := context.WithCancel(ctx) // 创建一个独立的上下文，方便 Subscription.Stop()

	s := Subscription{
		pubSub:      p,
		topic:       topic,
		redisKey:    formatTopicKey(topic),
		handler:     fnVal,
		handlerType: fnVal.Type(),
		concurrency: 1, // 默认并发为1
		dataChan:    make(chan []byte, defaultDataChanSize),
		stopChan:    make(chan struct{}),
		ctx:         subCtx,
		cancel:      subCancel,
		useRecovery: p.useRecovery, // 默认继承PubSub的useRecovery
	}

	for _, opt := range opts {
		opt(&s)
	}

	p.mu.Lock()
	p.subscriptions[topic] = append(p.subscriptions[topic], &s)
	p.mu.Unlock()

	p.wg.Add(1) // PubSub 等待此 Subscription

	log.Trace().Str("topic", topic).Int("concurrency", s.concurrency).Bool("recovery", s.useRecovery).Msg("new subscription created")
	return &s, nil
}

// Close 关闭PubSub服务，停止所有订阅，并等待它们完成
func (p *PubSub) Close() error {
	p.mu.Lock()
	select {
	case <-p.closed:
		p.mu.Unlock()
		log.Warn().Msg("pubsub already closed")
		return nil // 已经关闭
	default:
		close(p.closed)
		log.Info().Msg("pubsub closing...")
	}

	// 复制一份订阅列表，因为Subscription.Stop()会修改p.subscriptions
	var allSubs []*Subscription
	for _, subs := range p.subscriptions {
		allSubs = append(allSubs, subs...)
	}
	p.subscriptions = make(map[string][]*Subscription) // 清空，防止重复关闭
	p.mu.Unlock()

	for _, sub := range allSubs {
		if err := sub.Stop(); err != nil {
			log.Error().Err(err).Str("topic", sub.topic).Msg("error stopping subscription during pubsub close")
		}
	}

	p.wg.Wait() // 等待所有 Subscription 的 Loop 真正退出
	log.Info().Msg("pubsub closed gracefully")
	return nil
}

// --- Subscription 方法 ---

// Loop 启动订阅的处理循环
// 它会启动一个goroutine用于BLPOP，以及N个worker goroutine用于处理消息
func (s *Subscription) Loop() {
	log.Trace().Str("topic", s.topic).Msg("subscription loop starting")

	// 启动BLPOP goroutine
	s.wg.Add(1) // 为了BLPOP goroutine
	go s.blpopLoop()

	// 启动worker goroutines
	for i := 0; i < s.concurrency; i++ {
		s.wg.Add(1) // 为了每个worker goroutine
		go s.worker(i)
	}
	log.Info().Str("topic", s.topic).Int("workers", s.concurrency).Msg("subscription loop and workers started")
}

func (s *Subscription) blpopLoop() {
	defer s.wg.Done()
	defer log.Trace().Str("topic", s.topic).Msg("blpop loop stopped")

	log.Trace().Str("topic", s.topic).Msg("blpop loop started")
	for {
		select {
		case <-s.stopChan: // 外部要求停止
			return
		case <-s.ctx.Done(): // 订阅上下文取消
			return
		case <-s.pubSub.closed: // PubSub 关闭
			return
		default:
			// BLPOP 从 Redis list 中阻塞弹出元素
			// 返回值是一个 []string，格式为 [keyName, value]
			results, err := s.pubSub.redisClient.BLPop(s.ctx, blpopTimeout, s.redisKey).Result()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					continue
				}
				if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context canceled") {
					// 超时或上下文取消，继续循环检查 stopChan
					log.Trace().Str("topic", s.topic).Msg("blpop timed out or context canceled, retrying")
					continue
				}
				// 其他 Redis 错误
				log.Error().Err(err).Str("topic", s.topic).Msg("blpop failed")
				// 考虑错误恢复策略，例如指数退避重试或停止订阅
				// 为简单起见，这里在严重错误时短暂sleep后重试
				time.Sleep(1 * time.Second)
				continue
			}

			if len(results) == 2 {
				payload := []byte(results[1]) // results[0] is the key name
				log.Trace().Str("topic", s.topic).Bytes("payload", payload).Msg("message received from blpop")
				select {
				case s.dataChan <- payload:
					// 成功发送到处理通道
				case <-s.stopChan:
					log.Warn().Str("topic", s.topic).Msg("blpop loop stopping, discarding message")
					return
				case <-s.ctx.Done():
					log.Warn().Str("topic", s.topic).Msg("blpop loop context done, discarding message")
					return
				case <-s.pubSub.closed:
					log.Warn().Str("topic", s.topic).Msg("pubsub closed, blpop loop stopping, discarding message")
					return
				}
			} else {
				log.Warn().Str("topic", s.topic).Int("results_len", len(results)).Msg("blpop returned unexpected result length")
			}
		}
	}
}

func (s *Subscription) worker(workerId int) {
	defer s.wg.Done()
	defer log.Trace().Str("topic", s.topic).Int("worker_id", workerId).Msg("worker stopped")

	log.Trace().Str("topic", s.topic).Int("worker_id", workerId).Msg("worker started")
	for {
		select {
		case <-s.stopChan:
			return
		case <-s.ctx.Done():
			return
		case <-s.pubSub.closed:
			return
		case payload, ok := <-s.dataChan:
			if !ok { // dataChan 被关闭 (虽然在这个设计中不会主动关闭dataChan，但以防万一)
				log.Warn().Str("topic", s.topic).Int("worker_id", workerId).Msg("data channel closed")
				return
			}
			s.processMessage(workerId, payload)
		}
	}
}

func (s *Subscription) processMessage(workerId int, payload []byte) {
	s.processingMu.Lock() // 确保在Stop时，不会有新的处理逻辑开始
	defer s.processingMu.Unlock()

	// 检查是否已经停止，防止在锁定后stopChan关闭但依然处理
	select {
	case <-s.stopChan:
		log.Warn().Str("topic", s.topic).Int("worker_id", workerId).Msg("processing aborted, subscription stopping")
		return
	default:
	}

	log.Trace().Str("topic", s.topic).Int("worker_id", workerId).Bytes("payload", payload).Msg("worker processing message")

	if s.useRecovery {
		defer func() {
			if r := recover(); r != nil {
				log.Error().Str("topic", s.topic).Int("worker_id", workerId).Interface("panic", r).Msg("recovered panic in subscription handler")
				// 可以加入堆栈打印: string(debug.Stack())
			}
		}()
	}

	// 反序列化参数
	var rawArgs []json.RawMessage
	if err := json.Unmarshal(payload, &rawArgs); err != nil {
		log.Error().Err(err).Str("topic", s.topic).Int("worker_id", workerId).Bytes("payload", payload).Msg("failed to unmarshal raw arguments from payload")
		return
	}

	numIn := s.handlerType.NumIn()
	if len(rawArgs) != numIn {
		log.Error().Str("topic", s.topic).Int("worker_id", workerId).Int("expected_args", numIn).Int("actual_args", len(rawArgs)).Msg("argument count mismatch")
		// 可以考虑将消息放入死信队列
		return
	}

	// 准备调用函数的参数
	callArgs := make([]reflect.Value, numIn)
	for i := 0; i < numIn; i++ {
		argType := s.handlerType.In(i)
		// 创建该类型的零值指针，用于Unmarshal
		valPtr := reflect.New(argType)
		if err := json.Unmarshal(rawArgs[i], valPtr.Interface()); err != nil {
			log.Error().Err(err).Str("topic", s.topic).Int("worker_id", workerId).Int("arg_index", i).Str("target_type", argType.String()).Bytes("raw_arg", rawArgs[i]).Msg("failed to unmarshal argument for handler")
			// 可以考虑将消息放入死信队列
			return
		}
		callArgs[i] = valPtr.Elem() // 获取指针指向的实际值
	}

	// 调用函数
	s.handler.Call(callArgs)
	log.Trace().Str("topic", s.topic).Int("worker_id", workerId).Msg("handler called successfully")
}

// Stop 停止订阅，关闭worker pool和BLPOP goroutine
func (s *Subscription) Stop() error {
	s.pubSub.mu.Lock() // 操作PubSub的subscriptions列表前加锁
	// 从PubSub的跟踪中移除自己
	if subs, ok := s.pubSub.subscriptions[s.topic]; ok {
		newSubs := make([]*Subscription, 0, len(subs)-1)
		for _, sub := range subs {
			if sub != s { // 指针比较
				newSubs = append(newSubs, sub)
			}
		}
		if len(newSubs) == 0 {
			delete(s.pubSub.subscriptions, s.topic)
		} else {
			s.pubSub.subscriptions[s.topic] = newSubs
		}
	}
	s.pubSub.mu.Unlock()

	s.processingMu.Lock() // 等待当前可能正在处理的消息完成或确保没有新的开始
	defer s.processingMu.Unlock()

	select {
	case <-s.stopChan:
		log.Warn().Str("topic", s.topic).Msg("subscription already stopping or stopped")
		// 已经关闭或正在关闭，PubSub的wg计数器可能已经减过
		return ErrSubscriptionClosed
	default:
		close(s.stopChan) // 发送停止信号
		s.cancel()        // 取消订阅的上下文，会影响BLPOP
		log.Info().Str("topic", s.topic).Msg("subscription stopping...")
	}

	// 等待所有goroutine (blpopLoop, workers) 退出
	// 这里可以设置一个超时，防止wg.Wait()无限等待
	waitTimeout := time.After(10 * time.Second) // 例如10秒超时
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Str("topic", s.topic).Msg("subscription stopped gracefully")
	case <-waitTimeout:
		log.Error().Str("topic", s.topic).Msg("subscription stop timed out waiting for goroutines")
		// 即使超时，也应该减少 PubSub 的 WaitGroup 计数器
	}

	s.pubSub.wg.Done() // 通知PubSub，此Subscription已关闭

	// 清理dataChan中可能残留的数据（理论上blpopLoop停止后不会再写入）
	// 但为了安全，可以尝试清空，以防有worker在stopChan信号后仍尝试读取
	// 这个操作不是必须的，因为goroutine都退出了，channel会被GC
	// for len(s.dataChan) > 0 {
	//	<-s.dataChan
	// }
	// close(s.dataChan) // dataChan 由 blpopLoop 写入，由 workers 读取，不应该在这里关闭

	return nil
}
