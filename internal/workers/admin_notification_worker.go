package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/nats-io/nats.go"
)

const (
	DefaultAdminNotificationQueueGroup = "admin-notification-workers"
	DefaultSLACheckInterval           = 15 * time.Minute
	DefaultSLAThresholdHours          = 20
)

type ApplicationSLASource interface {
	FindPendingSLABreach(ctx context.Context, olderThan time.Time) ([]models.Application, error)
}

type AdminNotificationWorker struct {
	subscriber          ports.MessageSubscriber
	notificationService services.NotificationService
	slaSource           ApplicationSLASource
	queueGroup          string
	subscriptions       []*nats.Subscription
	tickerInterval      time.Duration
	slaThresholdHours   int
	mu                  sync.Mutex
	running             bool
	stopChan            chan struct{}
}

func NewAdminNotificationWorker(
	subscriber ports.MessageSubscriber,
	notificationService services.NotificationService,
	slaSource ...ApplicationSLASource,
) *AdminNotificationWorker {
	var source ApplicationSLASource
	if len(slaSource) > 0 {
		source = slaSource[0]
	}

	return &AdminNotificationWorker{
		subscriber:          subscriber,
		notificationService: notificationService,
		slaSource:           source,
		queueGroup:          DefaultAdminNotificationQueueGroup,
		tickerInterval:      DefaultSLACheckInterval,
		slaThresholdHours:   DefaultSLAThresholdHours,
	}
}

func (w *AdminNotificationWorker) SetTickerInterval(d time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.tickerInterval = d
}

func (w *AdminNotificationWorker) SetSLAThresholdHours(hours int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.slaThresholdHours = hours
}

func (w *AdminNotificationWorker) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return nil
	}
	if w.subscriber == nil {
		return errors.New("subscriber is required")
	}
	if w.notificationService == nil {
		return errors.New("notificationService is required")
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
		log.Printf("[AdminNotificationWorker] subscribed to topic '%s' (queueGroup: %s)", t.subject, w.queueGroup)
	}

	w.stopChan = make(chan struct{})
	w.running = true

	if w.slaSource != nil && w.tickerInterval > 0 {
		go w.runSLALoop()
	}

	return nil
}

func (w *AdminNotificationWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	close(w.stopChan)
	w.unsubscribeAllLocked()
	w.running = false
	log.Printf("[AdminNotificationWorker] stopped all subscriptions and SLA loop")
}

func (w *AdminNotificationWorker) unsubscribeAllLocked() {
	for _, sub := range w.subscriptions {
		if sub != nil && sub.IsValid() {
			if err := sub.Unsubscribe(); err != nil {
				log.Printf("[AdminNotificationWorker] unsubscribe error: %v", err)
			}
		}
	}
	w.subscriptions = nil
}

func (w *AdminNotificationWorker) runSLALoop() {
	ticker := time.NewTicker(w.tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			count, err := w.CheckSLANow(ctx)
			cancel()
			if err != nil {
				log.Printf("[AdminNotificationWorker] SLA check failed: %v", err)
			} else if count > 0 {
				log.Printf("[AdminNotificationWorker] SLA check generated %d warning notifications", count)
			}
		}
	}
}

func (w *AdminNotificationWorker) CheckSLANow(ctx context.Context) (int, error) {
	if w.slaSource == nil || w.notificationService == nil {
		return 0, nil
	}

	thresholdTime := time.Now().UTC().Add(-time.Duration(w.slaThresholdHours) * time.Hour)
	pendingApps, err := w.slaSource.FindPendingSLABreach(ctx, thresholdTime)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, app := range pendingApps {
		req := dtos.CreateNotificationRequest{
			Type:     "SLA_WARNING",
			Category: "underwriting",
			Severity: "WARNING",
			Title:    fmt.Sprintf("SLA Warning: Aplikasi #%s", app.ID),
			Message:  fmt.Sprintf("Aplikasi nasabah %s telah berada dalam antrean underwriting selama lebih dari %d jam.", app.FullName, w.slaThresholdHours),
			Link:     "/queue",
		}

		resp, createErr := w.notificationService.Create(ctx, req)
		if createErr != nil {
			log.Printf("[AdminNotificationWorker] failed to generate SLA warning for app %s: %v", app.ID, createErr)
		} else if resp != nil {
			count++
		}
	}

	return count, nil
}

func (w *AdminNotificationWorker) handleApplicationSubmitted(ctx context.Context, data []byte) error {
	var event dtos.ApplicationSubmittedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal ApplicationSubmittedEvent: %w", err)
	}

	productName := event.ProductName
	if productName == "" {
		productName = "Asuransi"
	}

	req := dtos.CreateNotificationRequest{
		Type:     "APPLICATION_SUBMITTED",
		Category: "underwriting",
		Severity: "INFO",
		Title:    fmt.Sprintf("Aplikasi Baru: %s (#%s)", event.FullName, event.ApplicationID),
		Message:  fmt.Sprintf("Pengajuan polis %s baru masuk dan menunggu review underwriting.", productName),
		Link:     "/queue",
	}

	resp, err := w.notificationService.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("create in-app notification: %w", err)
	}

	log.Printf("[AdminNotificationWorker] notification created for submitted app: %s (id: %s)", event.ApplicationID, resp.ID)
	return nil
}

func (w *AdminNotificationWorker) handleApplicationApproved(ctx context.Context, data []byte) error {
	var event dtos.ApplicationApprovedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal ApplicationApprovedEvent: %w", err)
	}

	reviewedBy := event.ReviewedBy
	if reviewedBy == "" {
		reviewedBy = "Underwriter"
	}

	req := dtos.CreateNotificationRequest{
		Type:     "APPLICATION_APPROVED",
		Category: "underwriting",
		Severity: "SUCCESS",
		Title:    fmt.Sprintf("Aplikasi Disetujui: #%s", event.ApplicationID),
		Message:  fmt.Sprintf("Aplikasi nasabah %s (%s) telah disetujui oleh %s.", event.FullName, event.ProductName, reviewedBy),
		Link:     "/queue",
	}

	resp, err := w.notificationService.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("create in-app notification: %w", err)
	}

	log.Printf("[AdminNotificationWorker] notification created for approved app: %s (id: %s)", event.ApplicationID, resp.ID)
	return nil
}

func (w *AdminNotificationWorker) handleApplicationRejected(ctx context.Context, data []byte) error {
	var event dtos.ApplicationRejectedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal ApplicationRejectedEvent: %w", err)
	}

	rejectionReason := event.RejectionReason
	if rejectionReason == "" {
		rejectionReason = "Kriteria underwriting tidak terpenuhi"
	}

	req := dtos.CreateNotificationRequest{
		Type:     "APPLICATION_REJECTED",
		Category: "underwriting",
		Severity: "WARNING",
		Title:    fmt.Sprintf("Aplikasi Ditolak: #%s", event.ApplicationID),
		Message:  fmt.Sprintf("Aplikasi %s (%s) ditolak. Alasan: %s", event.ApplicationID, event.ProductName, rejectionReason),
		Link:     "/queue",
	}

	resp, err := w.notificationService.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("create in-app notification: %w", err)
	}

	log.Printf("[AdminNotificationWorker] notification created for rejected app: %s (id: %s)", event.ApplicationID, resp.ID)
	return nil
}

func (w *AdminNotificationWorker) handleApplicationRFIRequested(ctx context.Context, data []byte) error {
	var event dtos.ApplicationRFIRequestedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal ApplicationRFIRequestedEvent: %w", err)
	}

	docs := strings.Join(event.RequiredDocs, ", ")
	if docs == "" {
		docs = "Kelengkapan berkas nasabah"
	}

	req := dtos.CreateNotificationRequest{
		Type:     "APPLICATION_RFI",
		Category: "underwriting",
		Severity: "INFO",
		Title:    fmt.Sprintf("Permintaan Dokumen: #%s", event.ApplicationID),
		Message:  fmt.Sprintf("Dokumen tambahan (%s) diminta untuk nasabah %s.", docs, event.FullName),
		Link:     "/queue",
	}

	resp, err := w.notificationService.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("create in-app notification: %w", err)
	}

	log.Printf("[AdminNotificationWorker] notification created for RFI app: %s (id: %s)", event.ApplicationID, resp.ID)
	return nil
}
