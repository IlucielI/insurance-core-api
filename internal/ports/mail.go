package ports

import "context"

type EmailMessage struct {
	To       []string
	Subject  string
	TextBody string
	HTMLBody string
}

type Mailer interface {
	Send(ctx context.Context, message EmailMessage) error
}
