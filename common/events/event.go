package events

import (
	"TKMall/common/log"
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
)

type EventType string

const (
	OrderCreated   EventType = "order.created"
	OrderPaid      EventType = "order.paid"
	StockUpdated   EventType = "stock.updated"
	UserRegistered EventType = "user.registered"
)

type Event struct {
	Type      EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType EventType, handler func(context.Context, Event) error)
}

type KafkaEventBus struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
	handlers map[EventType][]func(context.Context, Event) error
}

func NewKafkaEventBus(brokers []string) (*KafkaEventBus, error) {
	// 初始化 Kafka 配置
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &KafkaEventBus{
		producer: producer,
		consumer: consumer,
		handlers: make(map[EventType][]func(context.Context, Event) error),
	}, nil
}

func (eb *KafkaEventBus) Publish(ctx context.Context, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: string(event.Type),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = eb.producer.SendMessage(msg)
	return err
}

func (eb *KafkaEventBus) Subscribe(eventType EventType, handler func(context.Context, Event) error) {
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)

	// 启动消费者
	go func() {
		consumer, err := eb.consumer.ConsumePartition(string(eventType), 0, sarama.OffsetNewest)
		if err != nil {
			log.Debugf("Failed to start consumer for %s: %v", eventType, err)
			return
		}

		for msg := range consumer.Messages() {
			var event Event
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Debugf("Failed to unmarshal event: %v", err)
				continue
			}

			// 调用所有注册的处理器
			for _, h := range eb.handlers[eventType] {
				if err := h(context.Background(), event); err != nil {
					log.Debugf("Failed to handle event: %v", err)
				}
			}
		}
	}()
}

// 用户注册事件的payload结构
type UserRegisteredPayload struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
