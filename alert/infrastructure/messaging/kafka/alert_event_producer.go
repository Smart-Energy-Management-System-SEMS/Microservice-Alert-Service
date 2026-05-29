package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/segmentio/kafka-go"

	"microservice-alert-service/alert/infrastructure/configuration"
)

type AlertEventProducer struct {
	writer  *kafka.Writer
	logger  *log.Logger
	topic   string
	enabled bool
}

func NewAlertEventProducer(cfg configuration.Config, logger *log.Logger) *AlertEventProducer {
	topic := cfg.KafkaAlertCreatedTopic
	brokers := normalizeKafkaHosts(cfg.KafkaBrokers)
	logger.Printf("producer brokers effective: %v", brokers)
	if len(brokers) == 0 || topic == "" {
		return &AlertEventProducer{enabled: false}
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		RequiredAcks: kafka.RequireOne,
	}

	return &AlertEventProducer{
		writer:  writer,
		logger:  logger,
		topic:   topic,
		enabled: true,
	}
}

func (p *AlertEventProducer) Enabled() bool {
	return p.enabled
}

func (p *AlertEventProducer) Topic() string {
	return p.topic
}

func (p *AlertEventProducer) PublishJSON(ctx context.Context, key string, payload any) error {
	if !p.enabled {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: body,
	}); err != nil {
		return err
	}

	p.logger.Printf("kafka event published topic=%s key=%s", p.topic, key)
	return nil
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
		out = append(out, t)
	}
	return out
}
