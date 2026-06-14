package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type KafkaController struct {
	publisher kafkaPublisher
}

type kafkaPublisher interface {
	Enabled() bool
	Topic() string
	PublishJSON(ctx context.Context, key string, payload any) error
}

func NewKafkaController(publisher kafkaPublisher) *KafkaController {
	return &KafkaController{publisher: publisher}
}

func (c *KafkaController) PublishTestEvent(ctx *gin.Context) {
	if !c.publisher.Enabled() {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "kafka producer disabled: missing brokers or topic",
		})
		return
	}

	payload := map[string]any{}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		respondBadRequest(ctx, err)
		return
	}

	if len(payload) == 0 {
		payload = map[string]any{
			"eventType":  "alert.created",
			"event":      "alert.created",
			"source":     "alert-service",
			"occurredAt": time.Now().UTC().Format(time.RFC3339),
			"data": map[string]any{
				"message": "test event from alert-service",
			},
		}
	}

	key := "alert-service-test"
	if err := c.publisher.PublishJSON(ctx.Request.Context(), key, payload); err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"status": "published",
		"topic":  c.publisher.Topic(),
		"key":    key,
		"event":  payload,
	})
}
