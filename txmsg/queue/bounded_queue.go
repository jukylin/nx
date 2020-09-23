package queue

import (
	"github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/jukylin/esim/log"
	"study-go/txmsg/domain/entity"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"context"
)

// 实现 LocalQueue 接口
type BoundedQueue struct {
	queue *queue.BoundedQueue

	logger log.Logger

	// 队列容量
	capacity int

	// 消费者数量
	consumersNum int

	// 队列名称，区分多个队列
	name string

	cancel context.CancelFunc

	// 统计指标时间
	reportPeriod time.Duration
}

type BoundedQueueOption func(c *BoundedQueue)

func NewBoundedQueue(options ...BoundedQueueOption) LocalQueue {
	bq := &BoundedQueue{}

	for _, option := range options {
		option(bq)
	}

	if bq.capacity == 0 {
		bq.capacity = 500
	}

	if bq.consumersNum == 0 {
		bq.consumersNum = 50
	}

	if bq.reportPeriod == 0 {
		bq.reportPeriod = 3 * time.Second
	}

	bq.queue = queue.NewBoundedQueue(bq.capacity, bq.DroppedItem)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	bq.reportSize(ctx)
	bq.cancel = cancel

	return bq
}

func WithLogger(logger log.Logger) BoundedQueueOption {
	return func(tm *BoundedQueue) {
		tm.logger = logger
	}
}

func WithCapacity(capacity int) BoundedQueueOption {
	return func(tm *BoundedQueue) {
		tm.capacity = capacity
	}
}

func WithConsumersNum(consumersNum int) BoundedQueueOption {
	return func(tm *BoundedQueue) {
		tm.consumersNum = consumersNum
	}
}

func WithName(name string) BoundedQueueOption {
	return func(tm *BoundedQueue) {
		tm.name = name
	}
}

func WithReportPeriod(reportPeriod time.Duration) BoundedQueueOption {
	return func(tm *BoundedQueue) {
		tm.reportPeriod = reportPeriod
	}
}

func (bq *BoundedQueue) Produce(msg entity.Msg) bool {
	return bq.queue.Produce(msg)
}

func (bq *BoundedQueue) Consumer(consumer func(item interface{})) bool {
	bq.queue.StartConsumers(bq.consumersNum, consumer)
	return true
}

func (bq *BoundedQueue) DroppedItem(item interface{}) {
	bq.logger.Debugf("%+v 被丢弃", item)
	lab := prometheus.Labels{"name": bq.name}
	queueDroppedTotal.With(lab).Inc()
}

func (bq *BoundedQueue) reportSize(ctx context.Context) {
	ticker := time.NewTicker(bq.reportPeriod)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				lab := prometheus.Labels{"name": bq.name}
				size := bq.queue.Size()
				queueSize.With(lab).Set(float64(size))
			case <-ctx.Done():
				bq.logger.Infoc(ctx, "report定时器已停止")
				return
			}
		}
	}()

}

func (bq *BoundedQueue) Close() error {
	bq.queue.Stop()
	bq.cancel()
	return nil
}
