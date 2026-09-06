package smtp

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/ports"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient(Config{Host: " smtp.example.com ", Port: 587, FromEmail: " no-reply@example.com ", FromName: " Insurance Core "})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client.host != "smtp.example.com" || client.port != 587 || client.fromEmail != "no-reply@example.com" || client.fromName != "Insurance Core" || client.encryption != EncryptionStartTLS || client.timeout != 10*time.Second {
		t.Fatalf("NewClient() = %+v, want normalized config", client)
	}
}

func TestNewClientValidatesConfig(t *testing.T) {
	tests := []Config{
		{Host: "", Port: 587, FromEmail: "no-reply@example.com"},
		{Host: "smtp.example.com", Port: 0, FromEmail: "no-reply@example.com"},
		{Host: "smtp.example.com", Port: 70000, FromEmail: "no-reply@example.com"},
		{Host: "smtp.example.com", Port: 587, FromEmail: ""},
		{Host: "smtp.example.com", Port: 587, FromEmail: "no-reply@example.com", Encryption: "invalid"},
	}
	for _, tt := range tests {
		if _, err := NewClient(tt); err == nil {
			t.Fatalf("NewClient(%+v) error = nil, want error", tt)
		}
	}
}

func TestValidateMessage(t *testing.T) {
	valid := ports.EmailMessage{To: []string{"user@example.com"}, Subject: "Subject", TextBody: "Body"}
	if err := validateMessage(valid); err != nil {
		t.Fatalf("validateMessage() error = %v", err)
	}

	tests := []ports.EmailMessage{
		{Subject: "Subject", TextBody: "Body"},
		{To: []string{" "}, Subject: "Subject", TextBody: "Body"},
		{To: []string{"not-an-email"}, Subject: "Subject", TextBody: "Body"},
		{To: []string{"user@example.com"}, TextBody: "Body"},
		{To: []string{"user@example.com"}, Subject: "Subject"},
	}
	for _, tt := range tests {
		if err := validateMessage(tt); err == nil {
			t.Fatalf("validateMessage(%+v) error = nil, want error", tt)
		}
	}

	injectedRecipient := ports.EmailMessage{To: []string{"user@example.com\r\nBcc: hidden@example.com"}, Subject: "Subject", TextBody: "Body"}
	if err := validateMessage(injectedRecipient); err == nil || !strings.Contains(err.Error(), "invalid email address") {
		t.Fatalf("validateMessage() error = %v, want invalid email address", err)
	}
}

func TestBuildMIMEMessage(t *testing.T) {
	message := buildMIMEMessage("Insurance Core <no-reply@example.com>", ports.EmailMessage{
		To:       []string{"user@example.com"},
		Subject:  "Halo Bayu",
		TextBody: "plain body",
		HTMLBody: "<p>html body</p>",
	})
	content := string(message)
	for _, expected := range []string{
		"From: Insurance Core <no-reply@example.com>",
		"To: user@example.com",
		"Subject: Halo Bayu",
		"Content-Type: multipart/alternative",
		"plain body",
		"<p>html body</p>",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("message does not contain %q: %s", expected, content)
		}
	}
}

func TestBuildMIMEMessageSanitizesHeadersAndBody(t *testing.T) {
	message := buildMIMEMessage("Insurance Core\r\nX-Evil: yes", ports.EmailMessage{
		To:       []string{"user@example.com\r\nBcc: hidden@example.com"},
		Subject:  "Halo\r\nX-Injected: yes",
		TextBody: "line 1\r\nline 2",
		HTMLBody: "<p>html\r\nbody</p>",
	})
	content := string(message)
	for _, unexpected := range []string{"X-Evil", "Bcc:", "X-Injected", "\r\nline 2", "\r\nbody"} {
		if strings.Contains(content, unexpected) {
			t.Fatalf("message contains %q: %s", unexpected, content)
		}
	}
	for _, expected := range []string{"line 1 line 2", "<p>html body</p>"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("message does not contain %q: %s", expected, content)
		}
	}
}

func TestSendValidatesBeforeDial(t *testing.T) {
	client, err := NewClient(Config{Host: "smtp.example.com", Port: 587, FromEmail: "no-reply@example.com"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.dial = func(ctx context.Context, network string, address string) (net.Conn, error) {
		t.Fatal("dial should not be called for invalid message")
		return nil, nil
	}
	if err := client.Send(context.Background(), ports.EmailMessage{}); err == nil {
		t.Fatal("Send() error = nil, want validation error")
	}
}
