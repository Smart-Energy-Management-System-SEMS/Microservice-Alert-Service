// Package outboundservices defines the outbound "ports" (interfaces) the
// Application layer needs from the outside world: persistence repositories
// and notification senders. Concrete adapters live in the infrastructure
// layer and implement these contracts (Ports & Adapters / Hexagonal style).
package outboundservices

import "context"

// EmailSender is the port for delivering an email message.
type EmailSender interface {
	Send(ctx context.Context, to string, subject string, body string) error
}

// SmsSender is the port for delivering an SMS message.
type SmsSender interface {
	Send(ctx context.Context, to string, body string) error
}
