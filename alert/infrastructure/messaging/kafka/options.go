package kafka

import (
	"crypto/tls"
	"strings"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"

	"microservice-alert-service/alert/infrastructure/configuration"
)

func normalizeKafkaHosts(cfg configuration.Config) []string {
	if len(cfg.KafkaBrokers) == 0 {
		return nil
	}

	if !shouldRewriteKafkaHosts(cfg.Environment) {
		return append([]string(nil), cfg.KafkaBrokers...)
	}

	out := make([]string, 0, len(cfg.KafkaBrokers))
	for _, b := range cfg.KafkaBrokers {
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

func buildDialer(cfg configuration.Config) (*kafka.Dialer, error) {
	dialer := &kafka.Dialer{
		ClientID: strings.TrimSpace(cfg.KafkaClientID),
	}

	protocol := strings.ToUpper(strings.TrimSpace(cfg.KafkaSecurityProtocol))
	if protocol == "SSL" || protocol == "SASL_SSL" {
		dialer.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	mechanism, err := buildSASLMechanism(cfg)
	if err != nil {
		return nil, err
	}
	dialer.SASLMechanism = mechanism

	return dialer, nil
}

func buildSASLMechanism(cfg configuration.Config) (sasl.Mechanism, error) {
	mechanism := strings.ToUpper(strings.TrimSpace(cfg.KafkaSASLMechanism))
	if mechanism == "" || strings.TrimSpace(cfg.KafkaUsername) == "" {
		return nil, nil
	}

	switch mechanism {
	case "PLAIN":
		return plain.Mechanism{
			Username: cfg.KafkaUsername,
			Password: cfg.KafkaPassword,
		}, nil
	case "SCRAM-SHA-256":
		return scram.Mechanism(scram.SHA256, cfg.KafkaUsername, cfg.KafkaPassword)
	case "SCRAM-SHA-512":
		return scram.Mechanism(scram.SHA512, cfg.KafkaUsername, cfg.KafkaPassword)
	default:
		return nil, nil
	}
}

func shouldRewriteKafkaHosts(environment string) bool {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "", "local", "development", "dev":
		return true
	default:
		return false
	}
}
