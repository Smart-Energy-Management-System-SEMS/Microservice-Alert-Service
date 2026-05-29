package kafka

import (
	"context"
	"log"
	"strings"

	"github.com/segmentio/kafka-go"

	"microservice-alert-service/alert/application/eventhandlers"
	"microservice-alert-service/alert/infrastructure/configuration"
)

type ConsumptionConsumer struct {
	reader  *kafka.Reader
	handler *eventhandlers.ConsumptionEventHandler
	logger  *log.Logger
	enabled bool
}

func NewConsumptionConsumer(cfg configuration.Config, handler *eventhandlers.ConsumptionEventHandler, logger *log.Logger) *ConsumptionConsumer {
	brokers := normalizeKafkaHosts(cfg.KafkaBrokers)
	logger.Printf("consumer brokers effective: %v", brokers)
	if len(brokers) == 0 || cfg.KafkaConsumptionTopic == "" {
		return &ConsumptionConsumer{enabled: false}
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   cfg.KafkaConsumptionTopic,
		GroupID: cfg.KafkaConsumerGroup,
	})

	return &ConsumptionConsumer{reader: reader, handler: handler, logger: logger, enabled: true}
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

func (c *ConsumptionConsumer) Enabled() bool {
	return c.enabled
}

func (c *ConsumptionConsumer) Start(ctx context.Context) {
	defer c.reader.Close()
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Printf("kafka read error: %v", err)
			continue
		}

		if err := c.handler.HandleMessage(ctx, msg.Value); err != nil {
			c.logger.Printf("kafka handler error: %v", err)
		}
	}
}
