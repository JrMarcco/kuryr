package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var _ Producer[any] = (*GeneralProducer[any])(nil)

type GeneralProducer[T any] struct {
	topic    string
	producer *kafka.Producer
}

func (p *GeneralProducer[T]) Produce(ctx context.Context, event T) error {
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	// 创建专用投递 channel
	deliveryChan := make(chan kafka.Event, 1)

	// 发送消息，处理队列满的情况
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 没超时就继续执行。
		}

		err = p.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &p.topic,
				Partition: kafka.PartitionAny,
			},
			Value: data,
		}, deliveryChan)

		if err != nil {
			var kafkaErr kafka.Error
			ok := errors.As(err, &kafkaErr)
			if ok && kafkaErr.Code() == kafka.ErrQueueFull {
				// 队列满，等待 1000ms 后重试
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second):
					continue
				}
			}
			return fmt.Errorf("failed to produce message [ %s ] to topic [ %s ]: %w", data, p.topic, err)
		}

		for {
			flushed := p.producer.Flush(int(time.Second))
			if flushed == 0 {
				break
			}
		}

		// 成功提交到队列。
		break
	}

	// 等待投递报告。
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-deliveryChan:
		switch eventType := e.(type) {
		case *kafka.Message:
			if eventType.TopicPartition.Error != nil {
				return fmt.Errorf("[failed to produce message [ %s ] to topic [ %s ]: %w", data, p.topic, eventType.TopicPartition.Error)
			}
		case kafka.Error:
			return errors.New(eventType.Error())
		default:
			return errors.New(e.String())
		}
	}

	return nil
}

func (p *GeneralProducer[T]) Close() {
	// 等待所有消息发送完成。
	for {
		flushed := p.producer.Flush(int(time.Second))
		if flushed == 0 {
			break
		}
	}
	// 关闭生产者
	p.producer.Close()
}

func NewGeneralProducer[T any](topic string, producer *kafka.Producer) *GeneralProducer[T] {
	return &GeneralProducer[T]{
		topic:    topic,
		producer: producer,
	}
}
