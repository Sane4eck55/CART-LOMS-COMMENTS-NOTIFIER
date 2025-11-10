// Package producer ...
package producer

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
)

// Producer ...
type Producer struct {
	producer  sarama.SyncProducer
	topicName string
}

// NewProducer ...
func NewProducer(producer sarama.SyncProducer, topicName string) *Producer {
	return &Producer{
		producer:  producer,
		topicName: topicName,
	}
}

// SendMsg ...
func (p *Producer) SendMsg(msg *model.OrderEvent) (int32, int64, error) {
	value, err := json.Marshal(msg)
	if err != nil {
		return -1, -1, err
	}

	event := &sarama.ProducerMessage{
		Topic: p.topicName,
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", msg.OrderID)),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(event)
	if err != nil {
		return partition, offset, err
	}

	return partition, offset, nil
}
