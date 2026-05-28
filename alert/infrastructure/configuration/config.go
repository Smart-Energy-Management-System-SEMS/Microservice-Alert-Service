package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName           string
	ConfigServiceURL      string
	AutoMigrate           bool
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
		ServiceName:           getEnvOrDefault("SERVICE_NAME", "alert-service"),
		ConfigServiceURL:      strings.TrimSpace(os.Getenv("CONFIG_SERVICE_URL")),
		AutoMigrate:           getBoolEnvOrDefault("AUTO_MIGRATE", true),
		ServerPort:            getEnvOrDefault("SERVER_PORT", "8085"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		KafkaBrokers:          splitEnv("KAFKA_BROKERS", ""),
		KafkaConsumerGroup:    getFirstEnv([]string{"KAFKA_CONSUMER_GROUP", "KAFKA_GROUP_ID"}, ""),
		KafkaConsumptionTopic: getFirstEnv([]string{"KAFKA_CONSUMPTION_TOPIC", "KAFKA_TOPIC_DEVICE_READING_CREATED"}, ""),
		TwilioAccountSID:      os.Getenv("TWILIO_ACCOUNT_SID"),
		TwilioAPIKey:          os.Getenv("TWILIO_API_KEY"),
		TwilioAPISecret:       os.Getenv("TWILIO_API_SECRET"),
		TwilioPhoneNumber:     os.Getenv("TWILIO_PHONE_NUMBER"),
		MailHost:              os.Getenv("MAIL_HOST"),
		MailUsername:          os.Getenv("MAIL_USERNAME"),
		MailPassword:          os.Getenv("MAIL_PASSWORD"),
		MailFrom:              os.Getenv("MAIL_FROM"),
	}

	if err := cfg.loadFromConfigService(); err != nil {
		return cfg, err
	}

	if cfg.KafkaConsumerGroup == "" {
		cfg.KafkaConsumerGroup = "alert-service-group"
	}
	if cfg.KafkaConsumptionTopic == "" {
		cfg.KafkaConsumptionTopic = "energy.consumption.recorded"
	}
	if len(cfg.KafkaBrokers) == 0 {
		cfg.KafkaBrokers = splitCSV("localhost:9092")
	}
	if cfg.MailHost == "" {
		cfg.MailHost = "smtp.gmail.com"
	}

	mailPortValue := getEnvOrDefault("MAIL_PORT", "")
	if mailPortValue == "" {
		mailPortValue = "587"
	}
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

func (c *Config) loadFromConfigService() error {
	if c.ConfigServiceURL == "" {
		return nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := strings.TrimRight(c.ConfigServiceURL, "/")

	serviceData, err := fetchConfigMap(client, fmt.Sprintf("%s/api/v1/config/%s", baseURL, c.ServiceName))
	if err != nil {
		return err
	}
	if len(serviceData) == 0 {
		servicesData, err := fetchConfigMap(client, fmt.Sprintf("%s/api/v1/config/services", baseURL))
		if err != nil {
			return err
		}
		if nested, ok := servicesData[c.ServiceName].(map[string]any); ok {
			serviceData = nested
		}
	}
	kafkaData, err := fetchConfigMap(client, fmt.Sprintf("%s/api/v1/config/kafka", baseURL))
	if err != nil {
		return err
	}

	if c.ServerPort == "" {
		c.ServerPort = getString(serviceData, "serverPort", "server_port", "port")
	}
	if c.KafkaConsumerGroup == "" {
		c.KafkaConsumerGroup = getString(serviceData, "kafkaConsumerGroup", "kafka_consumer_group", "consumerGroup", "groupId")
	}
	if c.KafkaConsumptionTopic == "" {
		c.KafkaConsumptionTopic = getString(serviceData, "kafkaConsumptionTopic", "kafka_consumption_topic", "consumptionTopic", "topic")
	}
	if len(c.KafkaBrokers) == 0 {
		c.KafkaBrokers = firstBrokers(serviceData, kafkaData)
	}
	if c.MailHost == "" {
		c.MailHost = getString(serviceData, "mailHost", "mail_host", "smtpHost")
	}
	if c.MailFrom == "" {
		c.MailFrom = getString(serviceData, "mailFrom", "mail_from")
	}

	return nil
}

func fetchConfigMap(client *http.Client, url string) (map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return map[string]any{}, nil
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("config service returned status %d for %s", resp.StatusCode, url)
	}

	body := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return flattenEnvelope(body), nil
}

func flattenEnvelope(body map[string]any) map[string]any {
	// Supports {data:{...}} and direct object payloads.
	if data, ok := body["data"]; ok {
		if mapped, ok := data.(map[string]any); ok {
			return mapped
		}
	}
	return body
}

func getString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := values[key]; ok {
			switch t := v.(type) {
			case string:
				trimmed := strings.TrimSpace(t)
				if trimmed != "" {
					return trimmed
				}
			case float64:
				if t > 0 {
					return strconv.Itoa(int(t))
				}
			}
		}
	}
	return ""
}

func firstBrokers(maps ...map[string]any) []string {
	for _, m := range maps {
		if len(m) == 0 {
			continue
		}
		for _, key := range []string{"bootstrapServers", "bootstrap_servers", "brokers", "kafkaBrokers", "kafka_brokers"} {
			if value, ok := m[key]; ok {
				switch t := value.(type) {
				case string:
					brokers := splitCSV(t)
					if len(brokers) > 0 {
						return brokers
					}
				case []any:
					collected := make([]string, 0, len(t))
					for _, entry := range t {
						if broker, ok := entry.(string); ok && strings.TrimSpace(broker) != "" {
							collected = append(collected, strings.TrimSpace(broker))
						}
					}
					if len(collected) > 0 {
						return collected
					}
				}
			}
		}
	}
	return []string{}
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

	return splitCSV(value)
}

func splitCSV(value string) []string {
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

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}
