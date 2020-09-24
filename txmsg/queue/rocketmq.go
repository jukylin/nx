package queue

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/jukylin/esim/log"
)

type RocketMqOption func(c *RocketMq)

type RocketMq struct {
	logger log.Logger

	producer rocketmq.Producer

	producerOption []producer.Option
}

func NewRocketMq(options ...RocketMqOption) {
	rm := &RocketMq{}

	for _, option := range options {
		option(rm)
	}

	pd, err := rocketmq.NewProducer()
	if err != nil {
		rm.logger.Errorf(err.Error())
	}

	rm.producer = pd

	pd.Start()
}

func WithRocketMqLogger(logger log.Logger) RocketMqOption {
	return func(rm *RocketMq) {
		rm.logger = logger
	}
}

func WithProducerOption(options ...producer.Option) RocketMqOption {
	return func(rm *RocketMq) {
		rm.producerOption = options
	}
}

func (rm *RocketMq) Send(ctx context.Context, msg Message) error {
	message := primitive.NewMessage(msg.Topic, []byte(msg.Content))

	sendResult, err := rm.producer.SendSync(ctx, message)
	if err != nil {
		return err
	}

	if sendResult.Status == primitive.SendOK {
		return nil
	} else {
		return fmt.Errorf("发送失败 %s", sendResult)
	}
}

func (rm *RocketMq) Close() error {
	return rm.producer.Shutdown()
}
