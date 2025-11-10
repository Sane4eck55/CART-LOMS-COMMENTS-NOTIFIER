// Package main ...
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	config "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/notifier/configs"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/notifier/internal/kafka/consumer"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Println("started server notifier")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("LoadConfig : %v", err)
	}

	consumerGroup, err := initKafkaConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroupID)
	if err != nil {
		log.Fatalf("initKafkaConsumerGroup : %v", err)
	}
	defer func() {
		if err := consumerGroup.Close(); err != nil {
			log.Printf("consumerGroup.Close : %v", err)
		}
	}()

	consumer := &consumer.Consumer{}

	go func() {
		// слушаем ошибки группы в отдельной горутине
		for err := range consumerGroup.Errors() {
			log.Printf("ConsumerGroup error: %v", err)
		}
	}()

	go func() {
		for {
			err := consumerGroup.Consume(ctx, []string{cfg.Kafka.TopicName}, consumer)
			if err != nil {
				log.Printf("Consume: %v", err)
			}

			if ctx.Err() != nil {
				break
			}
		}
	}()

	<-sigterm
	log.Println("Termination signal received, shutting down...")
	cancel()
}

// initKafkaConsumerGroup ...
func initKafkaConsumerGroup(brokers, groupID string) (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()

	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false

	consumerGroup, err := sarama.NewConsumerGroup([]string{brokers}, groupID, config)
	if err != nil {
		log.Fatalf("NewConsumerGroup: %v", err)
	}

	return consumerGroup, nil
}

/*
// getPartitionID ...
func getPartitionID() (int32, error) {
	partitionID := os.Getenv("PARTITION_ID")
	if partitionID == "" {
		return -2, nil
	}
	partID, err := strconv.Atoi(partitionID)
	if err != nil {
		return -1, err
	}
	//nolint:gosec
	return int32(partID), nil
}
*/
