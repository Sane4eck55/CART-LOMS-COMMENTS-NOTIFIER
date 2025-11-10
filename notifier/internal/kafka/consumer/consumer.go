// Package consumer ...
package consumer

import (
	"fmt"

	"github.com/IBM/sarama"
)

// Consumer ...
type Consumer struct {
	Partition int32
}

// Setup ...
func (c *Consumer) Setup(sess sarama.ConsumerGroupSession) error {
	fmt.Printf("✅ [Consumer] Участник ID : %s группы запущен, ассайн получен", sess.MemberID())
	return nil
}

// Cleanup ...
func (c *Consumer) Cleanup(sess sarama.ConsumerGroupSession) error {
	fmt.Printf("👋 [Consumer] Участник ID : %s группы завершает работу ", sess.MemberID())
	return nil
}

// ConsumeClaim ...
func (c *Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	fmt.Printf("-> [Consumer] Чтение партиции %d\n", claim.Partition())

	for msg := range claim.Messages() {
		if msg.Partition == c.Partition {
			fmt.Printf("💬 [Consumer] %s: раздел=%d офсет=%d ключ=%s значение=%s\n",
				msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
			// помечаем сообщение как прочитанное
			sess.MarkMessage(msg, "")
		}
	}

	sess.Commit()
	return nil
}
