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

	out := make([]string, 0, len(cfg.KafkaBrokers))
	seen := make(map[string]struct{}, len(cfg.KafkaBrokers))
	for _, b := range cfg.KafkaBrokers {
		t := strings.TrimSpace(b)
		if t == "" {
			continue
		}
		if _, exists := seen[t]; exists {
			continue
		}
		seen[t] = struct{}{}
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
