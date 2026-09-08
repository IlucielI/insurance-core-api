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
	auditService        services.AuditLogService
	slaSource           ApplicationSLASource
	queueGroup          string
	subscriptions       []*nats.Subscription
	tickerInterval      time.Duration
	slaThresholdHours   int
	mu                  sync.Mutex
	running             bool
	slaTimer            *time.Timer
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

func (w *AdminNotificationWorker) WithAuditService(auditService services.AuditLogService) *AdminNotificationWorker {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.auditService = auditService
	return w
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

	w.running = true

	if w.slaSource != nil && w.tickerInterval > 0 {
		w.scheduleSLALoopLocked()
	}

	return nil
}

func (w *AdminNotificationWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	if w.slaTimer != nil {
		w.slaTimer.Stop()
		w.slaTimer = nil
	}
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

func (w *AdminNotificationWorker) scheduleSLALoopLocked() {
	if !w.running || w.tickerInterval <= 0 || w.slaSource == nil {
		return
	}

	w.slaTimer = time.AfterFunc(w.tickerInterval, func() {
		w.triggerSLACheck()

		w.mu.Lock()
		defer w.mu.Unlock()
		w.scheduleSLALoopLocked()
	})
}

func (w *AdminNotificationWorker) triggerSLACheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	count, err := w.CheckSLANow(ctx)
	if err != nil {
		log.Printf("[AdminNotificationWorker] SLA check failed: %v", err)
		return
	}
	if count > 0 {
		log.Printf("[AdminNotificationWorker] SLA check generated %d warning notifications", count)
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

	if len(pendingApps) == 0 {
		return 0, nil
	}

	existingTitles := make(map[string]bool)
	unreadOnly := true
	existingList, listErr := w.notificationService.List(ctx, dtos.NotificationQuery{
		UnreadOnly: &unreadOnly,
		Limit:      100,
	})
	if listErr != nil {
		log.Printf("[AdminNotificationWorker] warning: failed to fetch existing notifications for deduplication: %v", listErr)
	} else if existingList != nil {
		for _, notif := range existingList.Data {
			existingTitles[notif.Title] = true
		}
	}

	var batchRequests []dtos.CreateNotificationRequest
	for _, app := range pendingApps {
		title := fmt.Sprintf("SLA Warning: Aplikasi #%s", app.ID)
		if existingTitles[title] {
			continue
		}

		existingTitles[title] = true
		batchRequests = append(batchRequests, dtos.CreateNotificationRequest{
			Type:     "SLA_WARNING",
			Category: "underwriting",
			Severity: "WARNING",
			Title:    title,
			Message:  fmt.Sprintf("Aplikasi nasabah %s telah berada dalam antrean underwriting selama lebih dari %d jam.", app.FullName, w.slaThresholdHours),
			Link:     "/queue",
		})
	}

	if len(batchRequests) == 0 {
		return 0, nil
	}

	createdList, batchErr := w.notificationService.CreateBatch(ctx, batchRequests)
	if batchErr != nil {
		log.Printf("[AdminNotificationWorker] failed to batch create SLA warnings: %v", batchErr)
		return 0, batchErr
	}

	return len(createdList), nil
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

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := w.notificationService.Create(reqCtx, req)
	if err != nil {
		return fmt.Errorf("create in-app notification: %w", err)
	}

	if w.auditService != nil {
		if _, err := w.auditService.Record(reqCtx, dtos.CreateAuditLogRequest{
			ActorName:      event.FullName,
			ActorRole:      "Applicant",
			Action:         "APPLICATION_SUBMITTED",
			Category:       "underwriting",
			TargetResource: fmt.Sprintf("application:%s", event.ApplicationID),
			Status:         "success",
			Details: map[string]any{
				"product_name":      productName,
				"sum_assured":       event.SumAssured,
				"premium":           event.Premium,
				"payment_frequency": event.PaymentFrequency,
			},
		}); err != nil {
			log.Printf("[AdminNotificationWorker] warning: failed to create audit log for submitted app: %v", err)
		}
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

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := w.notificationService.Create(reqCtx, req)
	if err != nil {
		return fmt.Errorf("create in-app notification: %w", err)
	}

	if w.auditService != nil {
		if _, err := w.auditService.Record(reqCtx, dtos.CreateAuditLogRequest{
			ActorName:      reviewedBy,
			ActorRole:      "Underwriter",
			Action:         "POLICY_APPROVED",
			Category:       "underwriting",
			TargetResource: fmt.Sprintf("application:%s", event.ApplicationID),
			Status:         "success",
			Details: map[string]any{
				"policy_number": event.PolicyNumber,
				"product_name":  event.ProductName,
				"full_name":     event.FullName,
				"reviewed_by":   reviewedBy,
			},
		}); err != nil {
			log.Printf("[AdminNotificationWorker] warning: failed to create audit log for approved app: %v", err)
		}
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

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := w.notificationService.Create(reqCtx, req)
	if err != nil {
		return fmt.Errorf("create in-app notification: %w", err)
	}

	if w.auditService != nil {
		if _, err := w.auditService.Record(reqCtx, dtos.CreateAuditLogRequest{
			ActorName:      "Lead Underwriter",
			ActorRole:      "Underwriter",
			Action:         "APPLICATION_REJECTED",
			Category:       "underwriting",
			TargetResource: fmt.Sprintf("application:%s", event.ApplicationID),
			Status:         "success",
			Details: map[string]any{
				"rejection_reason": rejectionReason,
				"product_name":     event.ProductName,
			},
		}); err != nil {
			log.Printf("[AdminNotificationWorker] warning: failed to create audit log for rejected app: %v", err)
		}
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

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := w.notificationService.Create(reqCtx, req)
	if err != nil {
		return fmt.Errorf("create in-app notification: %w", err)
	}

	if w.auditService != nil {
		if _, err := w.auditService.Record(reqCtx, dtos.CreateAuditLogRequest{
			ActorName:      "Lead Underwriter",
			ActorRole:      "Underwriter",
			Action:         "RFI_REQUESTED",
			Category:       "underwriting",
			TargetResource: fmt.Sprintf("application:%s", event.ApplicationID),
			Status:         "success",
			Details: map[string]any{
				"required_docs": event.RequiredDocs,
				"notes":         event.Notes,
			},
		}); err != nil {
			log.Printf("[AdminNotificationWorker] warning: failed to create audit log for RFI app: %v", err)
		}
	}

	log.Printf("[AdminNotificationWorker] notification created for RFI app: %s (id: %s)", event.ApplicationID, resp.ID)
	return nil
}
