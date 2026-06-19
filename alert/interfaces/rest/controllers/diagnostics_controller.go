package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"microservice-alert-service/alert/infrastructure/configuration"
)

type DiagnosticsController struct {
	cfg       configuration.Config
	db        *gorm.DB
	publisher kafkaPublisher
}

func NewDiagnosticsController(cfg configuration.Config, db *gorm.DB, publisher kafkaPublisher) *DiagnosticsController {
	return &DiagnosticsController{
		cfg:       cfg,
		db:        db,
		publisher: publisher,
	}
}

func (c *DiagnosticsController) ValidateAll(ctx *gin.Context) {
	requiredEnv := map[string]bool{
		"SERVICE_NAME":             strings.TrimSpace(c.cfg.ServiceName) != "",
		"CONFIG_SERVICE_URL":       strings.TrimSpace(c.cfg.ConfigServiceURL) != "",
		"DATABASE_URL":             strings.TrimSpace(c.cfg.DatabaseURL) != "",
		"KAFKA_BROKERS":            len(c.cfg.KafkaBrokers) > 0,
		"KAFKA_CONSUMPTION_TOPICS": len(c.cfg.KafkaConsumptionTopics) > 0,
		"KAFKA_ALERTS_TOPIC":       strings.TrimSpace(c.cfg.KafkaAlertCreatedTopic) != "",
		"MAIL_USERNAME":            strings.TrimSpace(c.cfg.MailUsername) != "",
		"MAIL_PASSWORD":            strings.TrimSpace(c.cfg.MailPassword) != "",
		"TWILIO_ACCOUNT_SID":       strings.TrimSpace(c.cfg.TwilioAccountSID) != "",
		"TWILIO_API_KEY":           strings.TrimSpace(c.cfg.TwilioAPIKey) != "",
		"TWILIO_API_SECRET":        strings.TrimSpace(c.cfg.TwilioAPISecret) != "",
	}

	dbStatus := gin.H{
		"ok": false,
	}
	if c.db != nil {
		if sqlDB, err := c.db.DB(); err != nil {
			dbStatus["error"] = err.Error()
		} else {
			pingCtx, cancel := context.WithTimeout(ctx.Request.Context(), 3*time.Second)
			defer cancel()
			if err := sqlDB.PingContext(pingCtx); err != nil {
				dbStatus["error"] = err.Error()
			} else {
				dbStatus["ok"] = true
			}
		}
	}

	consumerTopics := make(map[string]bool, len(c.cfg.KafkaConsumptionTopics))
	for _, topic := range c.cfg.KafkaConsumptionTopics {
		consumerTopics[strings.TrimSpace(topic)] = true
	}

	kafkaStatus := gin.H{
		"enabled":                c.cfg.KafkaEnabled,
		"publisherEnabled":       c.publisher != nil && c.publisher.Enabled(),
		"brokers":                c.cfg.KafkaBrokers,
		"consumerGroup":          c.cfg.KafkaConsumerGroup,
		"consumptionTopics":      c.cfg.KafkaConsumptionTopics,
		"alertsTopic":            c.cfg.KafkaAlertCreatedTopic,
		"alertsTopicValid":       strings.EqualFold(strings.TrimSpace(c.cfg.KafkaAlertCreatedTopic), "alerts.events"),
		"energyTopicPresent":     consumerTopics["energy.events"],
		"analyticsTopicPresent":  consumerTopics["analytics.events"],
		"expectedEventHubs":      []string{"energy.events", "analytics.events", "alerts.events"},
		"azureEventHubsReminder": "Asegura que existan los Event Hubs energy.events y analytics.events, porque este servicio los consume.",
	}

	missingEnv := make([]string, 0)
	for key, ok := range requiredEnv {
		if !ok {
			missingEnv = append(missingEnv, key)
		}
	}

	allGood := dbStatus["ok"] == true &&
		kafkaStatus["alertsTopicValid"] == true &&
		kafkaStatus["energyTopicPresent"] == true &&
		kafkaStatus["analyticsTopicPresent"] == true &&
		len(missingEnv) == 0

	statusCode := http.StatusOK
	if !allGood {
		statusCode = http.StatusServiceUnavailable
	}

	ctx.JSON(statusCode, gin.H{
		"status":          map[bool]string{true: "ok", false: "warning"}[allGood],
		"service":         c.cfg.ServiceName,
		"environment":     c.cfg.Environment,
		"database":        dbStatus,
		"kafka":           kafkaStatus,
		"requiredEnv":     requiredEnv,
		"missingEnv":      missingEnv,
		"configService":   c.cfg.ConfigServiceURL,
		"swaggerEndpoint": "/swagger/index.html",
	})
}
