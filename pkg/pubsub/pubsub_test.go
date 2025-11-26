package pubsub

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TestPubSubThroughput 测试 pubsub 的吞吐量
// 使用 goroutine 发布消息，两个 atomic 计数器分别统计发布和消费数量
// 每秒打印一次计数，并计算增量
func TestPubSubThroughput(t *testing.T) {
	// 设置日志级别，避免 trace 日志干扰
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	log.Logger = log.Logger.Level(zerolog.ErrorLevel)

	// 连接 Redis（根据实际情况修改配置）
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	// 测试 Redis 连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// 清理测试数据
	testTopic := "test_throughput_topic"
	rdb.Del(ctx, formatTopicKey(testTopic))
	defer rdb.Del(ctx, formatTopicKey(testTopic))

	// 创建 PubSub 实例
	ps := New(rdb, WithQueueSize(0), WithRecovery())
	defer ps.Close()

	// 创建 atomic 计数器
	var publishCount atomic.Int64
	var consumeCount atomic.Int64

	// 订阅消息处理函数
	messageHandler := func(msg string) {
		consumeCount.Add(1)
		// 模拟消息处理时间
		// time.Sleep(time.Microsecond)
	}

	// 订阅 topic，设置并发度
	sub, err := ps.Subscribe(ctx, testTopic, messageHandler, WithConcurrency(10))
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	sub.Loop()

	// 启动统计 goroutine
	stopStats := make(chan struct{})
	defer close(stopStats)

	go func() {
		var lastPublish int64
		var lastConsume int64
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				currentPublish := publishCount.Load()
				currentConsume := consumeCount.Load()

				publishDelta := currentPublish - lastPublish
				consumeDelta := currentConsume - lastConsume

				fmt.Printf("[Stats] Publish: %d (+%d/s) | Consume: %d (+%d/s) | Lag: %d\n",
					currentPublish, publishDelta,
					currentConsume, consumeDelta,
					currentPublish-currentConsume)

				lastPublish = currentPublish
				lastConsume = currentConsume

			case <-stopStats:
				return
			}
		}
	}()

	// 启动多个 publisher goroutine
	publisherCount := 5
	messagesPerPublisher := 2000
	publishDone := make(chan struct{})

	for i := 0; i < publisherCount; i++ {
		go func(publisherID int) {
			for j := 0; j < messagesPerPublisher; j++ {
				message := fmt.Sprintf("publisher-%d-msg-%d", publisherID, j)
				err := ps.Publish(ctx, testTopic, message)
				if err != nil {
					t.Logf("Publish error: %v", err)
					continue
				}
				publishCount.Add(1)

				// 可选：添加小延迟控制发布速率
				// time.Sleep(time.Microsecond * 100)
			}
		}(i)
	}

	// 等待一段时间让发布和消费完成
	go func() {
		time.Sleep(15 * time.Second)
		close(publishDone)
	}()

	<-publishDone

	// 最终统计
	finalPublish := publishCount.Load()
	finalConsume := consumeCount.Load()

	fmt.Printf("\n[Final Stats]\n")
	fmt.Printf("Total Published: %d\n", finalPublish)
	fmt.Printf("Total Consumed: %d\n", finalConsume)
	fmt.Printf("Lag: %d\n", finalPublish-finalConsume)

	// 验证消费数量接近发布数量（允许一定误差，因为可能还在处理中）
	expectedTotal := int64(publisherCount * messagesPerPublisher)
	if finalPublish < expectedTotal {
		t.Logf("Warning: Published %d messages, expected %d", finalPublish, expectedTotal)
	}

	// 等待一段时间让剩余消息被消费
	time.Sleep(2 * time.Second)
	finalConsume = consumeCount.Load()
	fmt.Printf("After wait - Total Consumed: %d\n", finalConsume)
}

// TestPubSubHighLoad 高负载测试
func TestPubSubHighLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high load test in short mode")
	}

	// 设置日志级别，避免 trace 日志干扰
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	log.Logger = log.Logger.Level(zerolog.ErrorLevel)

	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	testTopic := "test_high_load_topic"
	rdb.Del(ctx, formatTopicKey(testTopic))
	defer rdb.Del(ctx, formatTopicKey(testTopic))

	ps := New(rdb, WithQueueSize(0), WithRecovery())
	defer ps.Close()

	var publishCount atomic.Int64
	var consumeCount atomic.Int64

	// 订阅处理
	sub, err := ps.Subscribe(ctx, testTopic, func(data map[string]interface{}) {
		consumeCount.Add(1)
	}, WithConcurrency(10), WithBatchSize(10))
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	sub.Loop()
	defer sub.Stop()

	// 统计 goroutine
	stopStats := make(chan struct{})
	defer close(stopStats)

	go func() {
		var lastPublish int64
		var lastConsume int64
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		startTime := time.Now()
		for {
			select {
			case <-ticker.C:
				currentPublish := publishCount.Load()
				currentConsume := consumeCount.Load()

				publishDelta := currentPublish - lastPublish
				consumeDelta := currentConsume - lastConsume

				elapsed := time.Since(startTime).Seconds()
				avgPublishRate := float64(currentPublish) / elapsed
				avgConsumeRate := float64(currentConsume) / elapsed

				fmt.Printf("[%ds] Pub: %d (+%d/s, avg: %.0f/s) | Con: %d (+%d/s, avg: %.0f/s) | Lag: %d\n",
					int(elapsed),
					currentPublish, publishDelta, avgPublishRate,
					currentConsume, consumeDelta, avgConsumeRate,
					currentPublish-currentConsume)

				lastPublish = currentPublish
				lastConsume = currentConsume

			case <-stopStats:
				return
			}
		}
	}()

	// 启动 10 个高速发布器
	publisherCount := 10
	messagesPerPublisher := 10000000000
	testDuration := 30 * time.Second

	testCtx, cancel := context.WithTimeout(ctx, testDuration)
	defer cancel()

	// 使用独立的 context 进行发布，避免 testCtx 超时影响 Redis 操作
	publishCtx := context.Background()

	for i := 0; i < publisherCount; i++ {
		go func(publisherID int) {
			msgID := 0
			for {
				select {
				case <-testCtx.Done():
					return
				default:
					if msgID >= messagesPerPublisher {
						return
					}

					data := map[string]interface{}{
						"publisher_id": publisherID,
						"msg_id":       msgID,
						"timestamp":    time.Now().Unix(),
					}

					// 使用 publishCtx 而不是 testCtx，避免 context deadline exceeded 错误
					err := ps.Publish(publishCtx, testTopic, data)
					if err != nil {
						continue
					}
					publishCount.Add(1)
					msgID++
				}
			}
		}(i)
	}

	// 等待测试完成
	<-testCtx.Done()

	// 最终统计
	time.Sleep(3 * time.Second)
	finalPublish := publishCount.Load()
	finalConsume := consumeCount.Load()

	fmt.Printf("\n[Final High Load Stats]\n")
	fmt.Printf("Total Published: %d\n", finalPublish)
	fmt.Printf("Total Consumed: %d\n", finalConsume)
	fmt.Printf("Lag: %d (%.2f%%)\n", finalPublish-finalConsume,
		float64(finalPublish-finalConsume)/float64(finalPublish)*100)
}
