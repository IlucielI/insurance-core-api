package workers

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	emailtemplate "github.com/bayuanugerah/insurance-core-api/internal/templates/email"
	"github.com/nats-io/nats.go"
)

type fakeSubscriber struct {
	handlers map[string]func(context.Context, []byte) error
}

func newFakeSubscriber() *fakeSubscriber {
	return &fakeSubscriber{handlers: make(map[string]func(context.Context, []byte) error)}
}

func (s *fakeSubscriber) Subscribe(subject string, handler func(context.Context, []byte) error) (*nats.Subscription, error) {
	s.handlers[subject] = handler
	return &nats.Subscription{}, nil
}

func (s *fakeSubscriber) QueueSubscribe(subject, queueGroup string, handler func(context.Context, []byte) error) (*nats.Subscription, error) {
	s.handlers[subject] = handler
	return &nats.Subscription{}, nil
}

type fakeMailer struct {
	messages []ports.EmailMessage
	err      error
}

func (m *fakeMailer) Send(ctx context.Context, message ports.EmailMessage) error {
	if m.err != nil {
		return m.err
	}
	m.messages = append(m.messages, message)
	return nil
}

func TestEmailNotificationWorkerLifecycle(t *testing.T) {
	sub := newFakeSubscriber()
	mailer := &fakeMailer{}
	renderer, err := emailtemplate.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	worker := NewEmailNotificationWorker(sub, mailer, renderer, "https://app.example.com")
	if err := worker.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	expectedTopics := []string{
		dtos.TopicApplicationSubmitted,
		dtos.TopicApplicationApproved,
		dtos.TopicApplicationRejected,
		dtos.TopicApplicationRFIRequested,
	}

	for _, topic := range expectedTopics {
		if _, exists := sub.handlers[topic]; !exists {
			t.Errorf("expected handler for topic %s", topic)
		}
	}

	worker.Stop()
}

func TestEmailNotificationWorkerHandleSubmitted(t *testing.T) {
	sub := newFakeSubscriber()
	mailer := &fakeMailer{}
	renderer, err := emailtemplate.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	worker := NewEmailNotificationWorker(sub, mailer, renderer, "https://app.example.com")
	_ = worker.Start(context.Background())

	event := dtos.ApplicationSubmittedEvent{
		ApplicationID:    "APP-123",
		ProductID:        "prod-1",
		ProductName:      "Secure Life Plus",
		FullName:         "Budi Santoso",
		Email:            "budi@example.com",
		SumAssured:       500000000,
		Premium:          1500000,
		PaymentFrequency: "tahun",
	}
	data, _ := json.Marshal(event)

	handler := sub.handlers[dtos.TopicApplicationSubmitted]
	if err := handler(context.Background(), data); err != nil {
		t.Fatalf("handler error = %v", err)
	}

	if len(mailer.messages) != 1 {
		t.Fatalf("expected 1 email, got %d", len(mailer.messages))
	}

	msg := mailer.messages[0]
	if msg.To[0] != "budi@example.com" {
		t.Errorf("To = %v, want budi@example.com", msg.To)
	}
	if !strings.Contains(msg.Subject, "APP-123") {
		t.Errorf("Subject = %s, want APP-123", msg.Subject)
	}
	if !strings.Contains(msg.HTMLBody, "Rp 500.000.000") {
		t.Errorf("expected formatted sum assured in HTML body")
	}
}

func TestEmailNotificationWorkerHandleApproved(t *testing.T) {
	sub := newFakeSubscriber()
	mailer := &fakeMailer{}
	renderer, err := emailtemplate.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	worker := NewEmailNotificationWorker(sub, mailer, renderer, "https://app.example.com")
	_ = worker.Start(context.Background())

	event := dtos.ApplicationApprovedEvent{
		ApplicationID:    "APP-123",
		PolicyNumber:     "POL-2026-123",
		ProductName:      "Secure Life Plus",
		FullName:         "Budi Santoso",
		Email:            "budi@example.com",
		SumAssured:       1000000000,
		Premium:          2760000,
		PaymentTerm:      10,
		PaymentFrequency: "tahun",
	}
	data, _ := json.Marshal(event)

	handler := sub.handlers[dtos.TopicApplicationApproved]
	if err := handler(context.Background(), data); err != nil {
		t.Fatalf("handler error = %v", err)
	}

	if len(mailer.messages) != 1 {
		t.Fatalf("expected 1 email, got %d", len(mailer.messages))
	}

	msg := mailer.messages[0]
	if !strings.Contains(msg.Subject, "POL-2026-123") {
		t.Errorf("Subject = %s, want POL-2026-123", msg.Subject)
	}
	if !strings.Contains(msg.HTMLBody, "e-Polis_Resmi_BayuInsurance_POL-2026-123.pdf") {
		t.Errorf("expected pdf attachment text in HTML body")
	}
}

func TestEmailNotificationWorkerHandleRejected(t *testing.T) {
	sub := newFakeSubscriber()
	mailer := &fakeMailer{}
	renderer, err := emailtemplate.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	worker := NewEmailNotificationWorker(sub, mailer, renderer, "https://app.example.com")
	_ = worker.Start(context.Background())

	event := dtos.ApplicationRejectedEvent{
		ApplicationID:       "APP-123",
		ProductName:         "Secure Life Plus",
		FullName:            "Budi Santoso",
		Email:               "budi@example.com",
		Premium:             2760000,
		RejectionCode:       "UW-DEC-401",
		RejectionReason:     "Riwayat kardiovaskular melebihi toleransi",
		LeadUnderwriterName: "dr. Hendra Kurniawan",
		LeadUnderwriterNIP:  "UW-2026-042",
	}
	data, _ := json.Marshal(event)

	handler := sub.handlers[dtos.TopicApplicationRejected]
	if err := handler(context.Background(), data); err != nil {
		t.Fatalf("handler error = %v", err)
	}

	if len(mailer.messages) != 1 {
		t.Fatalf("expected 1 email, got %d", len(mailer.messages))
	}

	msg := mailer.messages[0]
	if !strings.Contains(msg.HTMLBody, "UW-DEC-401") {
		t.Errorf("expected rejection code in HTML body")
	}
	if !strings.Contains(msg.HTMLBody, "dr. Hendra Kurniawan") {
		t.Errorf("expected underwriter name in HTML body")
	}
}

func TestEmailNotificationWorkerHandleRFI(t *testing.T) {
	sub := newFakeSubscriber()
	mailer := &fakeMailer{}
	renderer, err := emailtemplate.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	worker := NewEmailNotificationWorker(sub, mailer, renderer, "https://app.example.com")
	_ = worker.Start(context.Background())

	event := dtos.ApplicationRFIRequestedEvent{
		ApplicationID: "APP-123",
		ProductName:   "Secure Life Plus",
		FullName:      "Budi Santoso",
		Email:         "budi@example.com",
		Notes:         "Mohon upload ulang KTP",
		RequiredDocs:  []string{"Scan KTP HD"},
		SLADeadline:   "3 x 24 Jam",
	}
	data, _ := json.Marshal(event)

	handler := sub.handlers[dtos.TopicApplicationRFIRequested]
	if err := handler(context.Background(), data); err != nil {
		t.Fatalf("handler error = %v", err)
	}

	if len(mailer.messages) != 1 {
		t.Fatalf("expected 1 email, got %d", len(mailer.messages))
	}

	msg := mailer.messages[0]
	if !strings.Contains(msg.HTMLBody, "Mohon upload ulang KTP") {
		t.Errorf("expected notes in HTML body")
	}
}

func TestEmailNotificationWorkerValidationErrors(t *testing.T) {
	sub := newFakeSubscriber()
	mailer := &fakeMailer{}

	worker := NewEmailNotificationWorker(sub, mailer, nil)
	_ = worker.Start(context.Background())

	handler := sub.handlers[dtos.TopicApplicationSubmitted]

	// Invalid json
	if err := handler(context.Background(), []byte("{invalid")); err == nil {
		t.Errorf("expected error for invalid json")
	}

	// Missing email
	event := dtos.ApplicationSubmittedEvent{ApplicationID: "APP-1"}
	data, _ := json.Marshal(event)
	if err := handler(context.Background(), data); err == nil {
		t.Errorf("expected error for missing email")
	}

	// Mailer error
	mailer.err = errors.New("smtp connection failed")
	event.Email = "test@example.com"
	data, _ = json.Marshal(event)
	if err := handler(context.Background(), data); err == nil {
		t.Errorf("expected error when mailer fails")
	}
}
