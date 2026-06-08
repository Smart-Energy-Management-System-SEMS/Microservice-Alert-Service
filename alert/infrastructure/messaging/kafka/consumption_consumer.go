package kafka

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/segmentio/kafka-go"

	"microservice-alert-service/alert/infrastructure/configuration"
)

type TopicMessageHandler interface {
	HandleMessage(ctx context.Context, topic string, payload []byte) error
}

type ConsumptionConsumer struct {
	readers []readerBinding
	logger  *log.Logger
	enabled bool
}

type readerBinding struct {
	topic  string
	reader *kafka.Reader
}

func NewConsumptionConsumer(cfg configuration.Config, handler TopicMessageHandler, logger *log.Logger) *ConsumptionConsumer {
	brokers := normalizeKafkaHosts(cfg.KafkaBrokers)
	topics := normalizeTopics(cfg.KafkaConsumptionTopics, cfg.KafkaConsumptionTopic)
	logger.Printf("consumer brokers effective: %v", brokers)
	logger.Printf("consumer topics effective: %v", topics)
	if len(brokers) == 0 || len(topics) == 0 || handler == nil {
		return &ConsumptionConsumer{enabled: false}
	}

	readers := make([]readerBinding, 0, len(topics))
	for _, topic := range topics {
		readers = append(readers, readerBinding{
			topic: topic,
			reader: kafka.NewReader(kafka.ReaderConfig{
				Brokers: brokers,
				Topic:   topic,
				GroupID: cfg.KafkaConsumerGroup,
			}),
		})
	}

	return &ConsumptionConsumer{
		readers: readers,
		logger:  logger,
		enabled: true,
	}
}

func normalizeKafkaHosts(brokers []string) []string {
	out := make([]string, 0, len(brokers))
	for _, b := range brokers {
		t := strings.TrimSpace(b)
		if t == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(t), "kafka:") {
			out = append(out, "localhost:"+strings.TrimPrefix(t, "kafka:"))
			continue
		}
		if strings.EqualFold(t, "localhost:29092") || strings.EqualFold(t, "127.0.0.1:29092") {
			out = append(out, "localhost:9092")
			continue
		}
		out = append(out, t)
	}
	return out
}

func normalizeTopics(topics []string, legacyTopic string) []string {
	if len(topics) == 0 && strings.TrimSpace(legacyTopic) != "" {
		topics = []string{legacyTopic}
	}

	seen := make(map[string]struct{}, len(topics))
	out := make([]string, 0, len(topics))
	for _, topic := range topics {
		trimmed := strings.TrimSpace(topic)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}

	return out
}

func (c *ConsumptionConsumer) Enabled() bool {
	return c.enabled
}

func (c *ConsumptionConsumer) Start(ctx context.Context, handler TopicMessageHandler) {
	var wg sync.WaitGroup
	for _, binding := range c.readers {
		wg.Add(1)
		go func(binding readerBinding) {
			defer wg.Done()
			defer binding.reader.Close()
			for {
				msg, err := binding.reader.ReadMessage(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					c.logger.Printf("kafka read error topic=%s: %v", binding.topic, err)
					continue
				}

				if err := handler.HandleMessage(ctx, binding.topic, msg.Value); err != nil {
					c.logger.Printf("kafka handler error topic=%s: %v", binding.topic, err)
				}
			}
		}(binding)
	}

	<-ctx.Done()
	wg.Wait()
}
