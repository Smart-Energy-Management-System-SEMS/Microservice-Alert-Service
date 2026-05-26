package outboundservices

import "context"

type EmailSender interface {
    Send(ctx context.Context, to string, subject string, body string) error
}

type SmsSender interface {
    Send(ctx context.Context, to string, body string) error
}
