package kafka

import (
    "context"
    "log"

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
    if len(cfg.KafkaBrokers) == 0 || cfg.KafkaConsumptionTopic == "" {
        return &ConsumptionConsumer{enabled: false}
    }

    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: cfg.KafkaBrokers,
        Topic:   cfg.KafkaConsumptionTopic,
        GroupID: cfg.KafkaConsumerGroup,
    })

    return &ConsumptionConsumer{reader: reader, handler: handler, logger: logger, enabled: true}
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
