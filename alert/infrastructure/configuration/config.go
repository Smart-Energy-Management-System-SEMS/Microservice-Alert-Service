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
	ServiceName            string
	Environment            string
	ConfigServiceURL       string
	AutoMigrate            bool
	KafkaEnabled           bool
	ServerPort             string
	CORSAllowedOrigins     []string
	DatabaseURL            string
	KafkaBrokers           []string
	KafkaSecurityProtocol  string
	KafkaSASLMechanism     string
	KafkaUsername          string
	KafkaPassword          string
	KafkaClientID          string
	KafkaConsumerGroup     string
	KafkaConsumptionTopics []string
	KafkaConsumptionTopic  string
	KafkaAlertCreatedTopic string
	TwilioAccountSID       string
	TwilioAPIKey           string
	TwilioAPISecret        string
	TwilioPhoneNumber      string
	MailHost               string
	MailPort               int
	MailUsername           string
	MailPassword           string
	MailFrom               string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		ServiceName:            getEnvOrDefault("SERVICE_NAME", "alert-service"),
		Environment:            getFirstEnv([]string{"ENVIRONMENT", "APP_ENV"}, "local"),
		ConfigServiceURL:       strings.TrimSpace(os.Getenv("CONFIG_SERVICE_URL")),
		AutoMigrate:            getBoolEnvOrDefault("AUTO_MIGRATE", true),
		KafkaEnabled:           getBoolEnvOrDefault("KAFKA_ENABLED", true),
		ServerPort:             getFirstEnv([]string{"PORT", "SERVER_PORT"}, ""),
		CORSAllowedOrigins:     getCSVEnv([]string{"CORS_ALLOWED_ORIGINS", "ALLOWED_ORIGINS"}, ""),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		KafkaBrokers:           splitEnv("KAFKA_BROKERS", ""),
		KafkaSecurityProtocol:  strings.TrimSpace(os.Getenv("KAFKA_SECURITY_PROTOCOL")),
		KafkaSASLMechanism:     strings.TrimSpace(os.Getenv("KAFKA_SASL_MECHANISM")),
		KafkaUsername:          strings.TrimSpace(os.Getenv("KAFKA_USERNAME")),
		KafkaPassword:          os.Getenv("KAFKA_PASSWORD"),
		KafkaClientID:          strings.TrimSpace(os.Getenv("KAFKA_CLIENT_ID")),
		KafkaConsumerGroup:     getFirstEnv([]string{"KAFKA_CONSUMER_GROUP", "KAFKA_GROUP_ID"}, ""),
		KafkaConsumptionTopics: getTopicsFromEnv(),
		KafkaConsumptionTopic:  getFirstEnv([]string{"KAFKA_CONSUMPTION_TOPIC", "KAFKA_TOPIC_DEVICE_READING_CREATED"}, ""),
		KafkaAlertCreatedTopic: getFirstEnv([]string{"KAFKA_ALERTS_TOPIC", "KAFKA_TOPIC_ALERT_CREATED"}, ""),
		TwilioAccountSID:       os.Getenv("TWILIO_ACCOUNT_SID"),
		TwilioAPIKey:           os.Getenv("TWILIO_API_KEY"),
		TwilioAPISecret:        os.Getenv("TWILIO_API_SECRET"),
		TwilioPhoneNumber:      os.Getenv("TWILIO_PHONE_NUMBER"),
		MailHost:               os.Getenv("MAIL_HOST"),
		MailUsername:           os.Getenv("MAIL_USERNAME"),
		MailPassword:           os.Getenv("MAIL_PASSWORD"),
		MailFrom:               os.Getenv("MAIL_FROM"),
	}

	if err := cfg.loadFromConfigService(); err != nil {
		return cfg, err
	}

	if cfg.ServerPort == "" {
		cfg.ServerPort = "8080"
	}
	if len(cfg.CORSAllowedOrigins) == 0 {
		cfg.CORSAllowedOrigins = splitCSV("http://localhost:3000,http://localhost:5173")
	}

	if cfg.KafkaConsumerGroup == "" {
		cfg.KafkaConsumerGroup = "alert-service-group"
	}
	if len(cfg.KafkaConsumptionTopics) == 0 && cfg.KafkaConsumptionTopic != "" {
		cfg.KafkaConsumptionTopics = splitCSV(cfg.KafkaConsumptionTopic)
	}
	cfg.KafkaConsumptionTopics = normalizeKafkaConsumptionTopics(cfg.KafkaConsumptionTopics, cfg.KafkaConsumptionTopic)
	cfg.KafkaConsumptionTopic = primaryKafkaConsumptionTopic(cfg.KafkaConsumptionTopics)
	cfg.KafkaAlertCreatedTopic = normalizeKafkaAlertTopic(cfg.KafkaAlertCreatedTopic)
	cfg.KafkaBrokers = normalizeKafkaBrokersForRuntime(cfg.KafkaBrokers, cfg.Environment)
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
		} else if list := getServicesListEntry(servicesData, c.ServiceName); len(list) > 0 {
			serviceData = list
		}
	}
	kafkaData, err := fetchConfigMap(client, fmt.Sprintf("%s/api/v1/config/kafka", baseURL))
	if err != nil {
		return err
	}

	if c.ServerPort == "" {
		c.ServerPort = getString(serviceData, "serverPort", "server_port", "port")
	}
	if len(c.CORSAllowedOrigins) == 0 {
		c.CORSAllowedOrigins = getStringSlice(
			serviceData,
			"corsAllowedOrigins",
			"cors_allowed_origins",
			"allowedOrigins",
			"allowed_origins",
		)
	}
	if c.KafkaConsumerGroup == "" {
		c.KafkaConsumerGroup = getString(serviceData, "kafkaConsumerGroup", "kafka_consumer_group", "consumerGroup", "groupId")
	}
	if len(c.KafkaConsumptionTopics) == 0 {
		c.KafkaConsumptionTopics = getStringSlice(serviceData, "kafkaConsumptionTopics", "kafka_consumption_topics", "consumptionTopics", "topics")
	}
	if c.KafkaConsumptionTopic == "" {
		c.KafkaConsumptionTopic = getString(serviceData, "kafkaConsumptionTopic", "kafka_consumption_topic", "consumptionTopic", "topic")
	}
	if c.KafkaAlertCreatedTopic == "" {
		c.KafkaAlertCreatedTopic = getString(
			serviceData,
			"kafkaAlertsTopic",
			"kafka_alerts_topic",
			"alertsTopic",
			"kafkaAlertCreatedTopic",
			"kafka_alert_created_topic",
			"alertCreatedTopic",
		)
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
	if len(c.KafkaConsumptionTopics) == 0 && c.KafkaConsumptionTopic != "" {
		c.KafkaConsumptionTopics = splitCSV(c.KafkaConsumptionTopic)
	}
	c.KafkaConsumptionTopics = normalizeKafkaConsumptionTopics(c.KafkaConsumptionTopics, c.KafkaConsumptionTopic)
	c.KafkaConsumptionTopic = primaryKafkaConsumptionTopic(c.KafkaConsumptionTopics)
	c.KafkaAlertCreatedTopic = normalizeKafkaAlertTopic(c.KafkaAlertCreatedTopic)

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

func getStringSlice(values map[string]any, keys ...string) []string {
	for _, key := range keys {
		value, ok := values[key]
		if !ok {
			continue
		}

		switch typed := value.(type) {
		case string:
			parts := splitCSV(typed)
			if len(parts) > 0 {
				return parts
			}
		case []any:
			out := make([]string, 0, len(typed))
			for _, entry := range typed {
				asString, ok := entry.(string)
				if !ok {
					continue
				}
				trimmed := strings.TrimSpace(asString)
				if trimmed != "" {
					out = append(out, trimmed)
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}

	return nil
}

func getServicesListEntry(values map[string]any, serviceName string) map[string]any {
	servicesRaw, ok := values["services"]
	if !ok {
		return map[string]any{}
	}
	services, ok := servicesRaw.([]any)
	if !ok {
		return map[string]any{}
	}
	for _, item := range services {
		service, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := getString(service, "name", "serviceName", "service_name")
		if strings.EqualFold(strings.TrimSpace(name), strings.TrimSpace(serviceName)) {
			return service
		}
	}
	return map[string]any{}
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

func normalizeKafkaBrokersForRuntime(brokers []string, environment string) []string {
	_ = environment

	normalized := make([]string, 0, len(brokers))
	seen := make(map[string]struct{}, len(brokers))
	for _, broker := range brokers {
		trimmed := strings.TrimSpace(broker)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
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

func getCSVEnv(keys []string, defaultValue string) []string {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return splitCSV(value)
		}
	}

	return splitCSV(defaultValue)
}

func splitEnv(key string, defaultValue string) []string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}

	return splitCSV(value)
}

func getTopicsFromEnv() []string {
	topics := splitEnv("KAFKA_CONSUMPTION_TOPICS", "")
	if len(topics) > 0 {
		return topics
	}

	legacyTopic := getFirstEnv([]string{"KAFKA_CONSUMPTION_TOPIC", "KAFKA_TOPIC_DEVICE_READING_CREATED"}, "")
	if legacyTopic == "" {
		return nil
	}

	return splitCSV(legacyTopic)
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

func requiredKafkaConsumptionTopics() []string {
	return []string{
		"energy.events",
		"analytics.events",
	}
}

func normalizeKafkaConsumptionTopics(topics []string, legacyTopic string) []string {
	normalized := make([]string, 0, len(topics)+1)
	seen := make(map[string]struct{}, len(topics)+1)

	for _, topic := range append(topics, legacyTopic) {
		mapped, ok := mapToGroupedKafkaTopic(topic)
		if !ok {
			continue
		}
		if _, exists := seen[mapped]; exists {
			continue
		}
		seen[mapped] = struct{}{}
		normalized = append(normalized, mapped)
	}

	if len(normalized) == 0 {
		return requiredKafkaConsumptionTopics()
	}

	return mergeTopics(normalized, requiredKafkaConsumptionTopics())
}

func primaryKafkaConsumptionTopic(topics []string) string {
	if len(topics) == 0 {
		return "energy.events"
	}
	return topics[0]
}

func normalizeKafkaAlertTopic(topic string) string {
	mapped, ok := mapToGroupedKafkaTopic(topic)
	if ok && mapped == "alerts.events" {
		return mapped
	}
	if strings.TrimSpace(topic) == "" {
		return "alerts.events"
	}
	if strings.EqualFold(strings.TrimSpace(topic), "alert.created") {
		return "alerts.events"
	}
	return strings.TrimSpace(topic)
}

func mapToGroupedKafkaTopic(topic string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(topic)) {
	case "":
		return "", false
	case "energy.events", "energy.consumption.recorded", "energy.reading.created":
		return "energy.events", true
	case "analytics.events", "analytics.anomaly.detected", "analytics.recommendation.generated":
		return "analytics.events", true
	case "alerts.events", "alert.created", "monitoring.alert.created":
		return "alerts.events", true
	default:
		return strings.TrimSpace(topic), false
	}
}

func mergeTopics(base []string, extras []string) []string {
	seen := make(map[string]struct{}, len(base)+len(extras))
	merged := make([]string, 0, len(base)+len(extras))

	for _, topic := range append(base, extras...) {
		trimmed := strings.TrimSpace(topic)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		merged = append(merged, trimmed)
	}

	return merged
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
