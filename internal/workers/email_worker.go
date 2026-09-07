package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	emailtemplate "github.com/bayuanugerah/insurance-core-api/internal/templates/email"
	"github.com/nats-io/nats.go"
)

const (
	DefaultEmailWorkerQueueGroup = "email-notification-workers"
	DefaultCustomerAppBaseURL    = "http://localhost:3000"
)

type EmailNotificationWorker struct {
	subscriber    ports.MessageSubscriber
	mailer        ports.Mailer
	renderer      *emailtemplate.Renderer
	baseURL       string
	queueGroup    string
	subscriptions []*nats.Subscription
	mu            sync.Mutex
	running       bool
}

func NewEmailNotificationWorker(subscriber ports.MessageSubscriber, mailer ports.Mailer, renderer *emailtemplate.Renderer, baseURL ...string) *EmailNotificationWorker {
	url := ""
	if len(baseURL) > 0 && strings.TrimSpace(baseURL[0]) != "" {
		url = strings.TrimRight(strings.TrimSpace(baseURL[0]), "/")
	}
	if url == "" {
		if env := strings.TrimSpace(os.Getenv("CUSTOMER_APP_BASE_URL")); env != "" {
			url = strings.TrimRight(env, "/")
		} else if env := strings.TrimSpace(os.Getenv("APP_BASE_URL")); env != "" {
			url = strings.TrimRight(env, "/")
		} else {
			url = DefaultCustomerAppBaseURL
		}
	}
	return &EmailNotificationWorker{
		subscriber: subscriber,
		mailer:     mailer,
		renderer:   renderer,
		baseURL:    url,
		queueGroup: DefaultEmailWorkerQueueGroup,
	}
}

func (w *EmailNotificationWorker) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return nil
	}
	if w.subscriber == nil {
		return errors.New("subscriber is required")
	}
	if w.mailer == nil {
		return errors.New("mailer is required")
	}
	if w.renderer == nil {
		renderer, err := emailtemplate.NewRenderer()
		if err != nil {
			return fmt.Errorf("init email renderer: %w", err)
		}
		w.renderer = renderer
	}

	topics := []struct {
		subject string
		handler func(context.Context, []byte) error
	}{
		{dtos.TopicApplicationSubmitted, w.handleApplicationSubmitted},
		{dtos.TopicApplicationApproved, w.handleApplicationApproved},
		{dtos.TopicApplicationRejected, w.handleApplicationRejected},
		{dtos.TopicApplicationRFIRequested, w.handleApplicationRFIRequested},
	}

	for _, t := range topics {
		sub, err := w.subscriber.QueueSubscribe(t.subject, w.queueGroup, t.handler)
		if err != nil {
			w.unsubscribeAllLocked()
			return fmt.Errorf("subscribe to %s: %w", t.subject, err)
		}
		w.subscriptions = append(w.subscriptions, sub)
		log.Printf("[EmailNotificationWorker] subscribed to topic '%s' (queueGroup: %s)", t.subject, w.queueGroup)
	}

	w.running = true
	return nil
}

func (w *EmailNotificationWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}
	w.unsubscribeAllLocked()
	w.running = false
	log.Printf("[EmailNotificationWorker] stopped all subscriptions")
}

func (w *EmailNotificationWorker) unsubscribeAllLocked() {
	for _, sub := range w.subscriptions {
		if sub != nil && sub.IsValid() {
			if err := sub.Unsubscribe(); err != nil {
				log.Printf("[EmailNotificationWorker] unsubscribe error: %v", err)
			}
		}
	}
	w.subscriptions = nil
}

func (w *EmailNotificationWorker) handleApplicationSubmitted(ctx context.Context, data []byte) error {
	var event dtos.ApplicationSubmittedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal ApplicationSubmittedEvent: %w", err)
	}

	if event.Email == "" {
		return errors.New("recipient email is required")
	}

	paymentFreq := event.PaymentFrequency
	if paymentFreq == "" {
		paymentFreq = "tahun"
	}

	portalURL := fmt.Sprintf("%s/portal/status/%s", w.baseURL, event.ApplicationID)

	textBody, htmlBody, err := w.renderer.RenderApplicationSubmitted(emailtemplate.ApplicationSubmittedData{
		FullName:            event.FullName,
		ProductName:         event.ProductName,
		ApplicationID:       event.ApplicationID,
		SumAssuredFormatted: emailtemplate.FormatIDR(event.SumAssured),
		PremiumFormatted:    emailtemplate.FormatIDR(event.Premium),
		PaymentFrequency:    paymentFreq,
		PortalURL:           portalURL,
	})
	if err != nil {
		return fmt.Errorf("render application submitted email: %w", err)
	}

	subject := fmt.Sprintf("[KONFIRMASI] Pengajuan Aplikasi Asuransi #%s Berhasil Diterima", event.ApplicationID)
	msg := ports.EmailMessage{
		To:       []string{event.Email},
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}

	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := w.mailer.Send(sendCtx, msg); err != nil {
		return fmt.Errorf("send application submitted email: %w", err)
	}

	log.Printf("[EmailNotificationWorker] sent submission email to %s for application #%s", event.Email, event.ApplicationID)
	return nil
}

func (w *EmailNotificationWorker) handleApplicationApproved(ctx context.Context, data []byte) error {
	var event dtos.ApplicationApprovedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal ApplicationApprovedEvent: %w", err)
	}

	if event.Email == "" {
		return errors.New("recipient email is required")
	}

	policyNo := event.PolicyNumber
	if policyNo == "" {
		policyNo = fmt.Sprintf("POL-%s", event.ApplicationID)
	}

	now := time.Now().UTC()
	termYears := event.PaymentTerm
	if termYears <= 0 {
		termYears = 10
	}
	endYear := now.AddDate(termYears, 0, 0)
	protectionPeriod := fmt.Sprintf("%s – %s (%d Tahun)", now.Format("02 Jan 2006"), endYear.Format("02 Jan 2006"), termYears)

	policyDownloadURL := fmt.Sprintf("%s/portal/policy/%s/download", w.baseURL, policyNo)
	portalURL := fmt.Sprintf("%s/portal/policy/%s", w.baseURL, policyNo)

	textBody, htmlBody, err := w.renderer.RenderApplicationApproved(emailtemplate.ApplicationApprovedData{
		FullName:            event.FullName,
		PolicyNumber:        policyNo,
		ProductName:         event.ProductName,
		ApplicationID:       event.ApplicationID,
		SumAssuredFormatted: emailtemplate.FormatIDR(event.SumAssured),
		ProtectionPeriod:    protectionPeriod,
		PremiumFormatted:    emailtemplate.FormatIDR(event.Premium),
		PolicyDownloadURL:   policyDownloadURL,
		PortalURL:           portalURL,
	})
	if err != nil {
		return fmt.Errorf("render application approved email: %w", err)
	}

	subject := fmt.Sprintf("[SELAMAT] Polis Asuransi Anda Telah Terbit - #%s (Proteksi Aktif)", policyNo)
	msg := ports.EmailMessage{
		To:       []string{event.Email},
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}

	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := w.mailer.Send(sendCtx, msg); err != nil {
		return fmt.Errorf("send application approved email: %w", err)
	}

	log.Printf("[EmailNotificationWorker] sent policy issued email to %s for policy #%s", event.Email, policyNo)
	return nil
}

func (w *EmailNotificationWorker) handleApplicationRejected(ctx context.Context, data []byte) error {
	var event dtos.ApplicationRejectedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal ApplicationRejectedEvent: %w", err)
	}

	if event.Email == "" {
		return errors.New("recipient email is required")
	}

	code := event.RejectionCode
	if code == "" {
		code = "UW-DEC-401"
	}
	reason := event.RejectionReason
	if reason == "" {
		reason = "Berdasarkan hasil evaluasi underwriting, profil risiko saat ini belum memenuhi kriteria penerimaan produk."
	}
	underwriter := event.LeadUnderwriterName
	if underwriter == "" {
		underwriter = "Tim Komite Underwriting Medis"
	}
	nip := event.LeadUnderwriterNIP
	if nip == "" {
		nip = "UW-2026-042"
	}

	consultationURL := fmt.Sprintf("%s/consultation/%s", w.baseURL, event.ApplicationID)

	textBody, htmlBody, err := w.renderer.RenderApplicationRejected(emailtemplate.ApplicationRejectedData{
		FullName:              event.FullName,
		ProductName:           event.ProductName,
		ApplicationID:         event.ApplicationID,
		RejectionCode:         code,
		RejectionReason:       reason,
		LeadUnderwriterName:   underwriter,
		LeadUnderwriterNIP:    nip,
		RefundAmountFormatted: emailtemplate.FormatIDR(event.Premium),
		ConsultationURL:       consultationURL,
	})
	if err != nil {
		return fmt.Errorf("render application rejected email: %w", err)
	}

	subject := fmt.Sprintf("[PEMBERITAHUAN RESMI] Hasil Keputusan Underwriting Aplikasi Polis #%s", event.ApplicationID)
	msg := ports.EmailMessage{
		To:       []string{event.Email},
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}

	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := w.mailer.Send(sendCtx, msg); err != nil {
		return fmt.Errorf("send application rejected email: %w", err)
	}

	log.Printf("[EmailNotificationWorker] sent rejection email to %s for application #%s", event.Email, event.ApplicationID)
	return nil
}

func (w *EmailNotificationWorker) handleApplicationRFIRequested(ctx context.Context, data []byte) error {
	var event dtos.ApplicationRFIRequestedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal ApplicationRFIRequestedEvent: %w", err)
	}

	if event.Email == "" {
		return errors.New("recipient email is required")
	}

	notes := event.Notes
	if notes == "" {
		notes = "Mohon bantuannya untuk mengunggah dokumen pendukung agar proses evaluasi underwriting dapat dilanjutkan."
	}

	docs := event.RequiredDocs
	if len(docs) == 0 {
		docs = []string{
			"1. Foto Ulang Fisik e-KTP (Resolusi Tinggi & Tanpa Pantulan Cahaya)",
			"2. Slip Gaji 3 Bulan Terakhir / Rekening Koran Legalisir Bank",
		}
	}

	sla := event.SLADeadline
	if sla == "" {
		deadline := time.Now().UTC().Add(72 * time.Hour)
		sla = fmt.Sprintf("3 x 24 Jam (Maks. %s WIB)", deadline.Format("02 Jan 2006, 15:04"))
	}

	uploadURL := fmt.Sprintf("%s/portal/rfi/%s", w.baseURL, event.ApplicationID)

	textBody, htmlBody, err := w.renderer.RenderApplicationRFI(emailtemplate.ApplicationRFIData{
		FullName:        event.FullName,
		ProductName:     event.ProductName,
		ApplicationID:   event.ApplicationID,
		Notes:           notes,
		RequiredDocs:    docs,
		SLADeadline:     sla,
		UploadPortalURL: uploadURL,
	})
	if err != nil {
		return fmt.Errorf("render application RFI email: %w", err)
	}

	subject := fmt.Sprintf("[TINDAK LANJUT DIPERLUKAN] Permintaan Dokumen Tambahan untuk Aplikasi Polis #%s", event.ApplicationID)
	msg := ports.EmailMessage{
		To:       []string{event.Email},
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}

	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := w.mailer.Send(sendCtx, msg); err != nil {
		return fmt.Errorf("send application RFI email: %w", err)
	}

	log.Printf("[EmailNotificationWorker] sent RFI email to %s for application #%s", event.Email, event.ApplicationID)
	return nil
}
