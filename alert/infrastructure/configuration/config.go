package configuration

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort            string
	DatabaseURL           string
	KafkaBrokers          []string
	KafkaConsumerGroup    string
	KafkaConsumptionTopic string
	TwilioAccountSID      string
	TwilioAPIKey          string
	TwilioAPISecret       string
	TwilioPhoneNumber     string
	MailHost              string
	MailPort              int
	MailUsername          string
	MailPassword          string
	MailFrom              string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		ServerPort:            getEnvOrDefault("SERVER_PORT", "8085"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		KafkaBrokers:          splitEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaConsumerGroup:    getFirstEnv([]string{"KAFKA_CONSUMER_GROUP", "KAFKA_GROUP_ID"}, "alert-service-group"),
		KafkaConsumptionTopic: getFirstEnv([]string{"KAFKA_CONSUMPTION_TOPIC", "KAFKA_TOPIC_DEVICE_READING_CREATED"}, "energy.consumption.recorded"),
		TwilioAccountSID:      os.Getenv("TWILIO_ACCOUNT_SID"),
		TwilioAPIKey:          os.Getenv("TWILIO_API_KEY"),
		TwilioAPISecret:       os.Getenv("TWILIO_API_SECRET"),
		TwilioPhoneNumber:     os.Getenv("TWILIO_PHONE_NUMBER"),
		MailHost:              getEnvOrDefault("MAIL_HOST", "smtp.gmail.com"),
		MailUsername:          os.Getenv("MAIL_USERNAME"),
		MailPassword:          os.Getenv("MAIL_PASSWORD"),
		MailFrom:              os.Getenv("MAIL_FROM"),
	}

	mailPortValue := getEnvOrDefault("MAIL_PORT", "587")
	port, err := strconv.Atoi(mailPortValue)
	if err != nil {
		return cfg, err
	}
	cfg.MailPort = port

	if cfg.DatabaseURL == "" {
		return cfg, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func getFirstEnv(keys []string, defaultValue string) string {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}

	return defaultValue
}

func splitEnv(key string, defaultValue string) []string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
