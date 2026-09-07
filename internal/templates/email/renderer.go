package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	texttemplate "text/template"
)

//go:embed *.html *.txt
var templateFiles embed.FS

type ApplicationSubmittedData struct {
	FullName            string
	ProductName         string
	ApplicationID       string
	SumAssuredFormatted string
	PremiumFormatted    string
	PaymentFrequency    string
	PortalURL           string
}

type ApplicationApprovedData struct {
	FullName            string
	PolicyNumber        string
	ProductName         string
	ApplicationID       string
	SumAssuredFormatted string
	ProtectionPeriod    string
	PremiumFormatted    string
	PolicyDownloadURL   string
	PortalURL           string
}

type ApplicationRejectedData struct {
	FullName              string
	ProductName           string
	ApplicationID         string
	RejectionCode         string
	RejectionReason       string
	LeadUnderwriterName   string
	LeadUnderwriterNIP    string
	RefundAmountFormatted string
	ConsultationURL       string
}

type ApplicationRFIData struct {
	FullName        string
	ProductName     string
	ApplicationID   string
	Notes           string
	RequiredDocs    []string
	SLADeadline     string
	UploadPortalURL string
}

type Renderer struct {
	htmlTemplates *template.Template
	textTemplates *texttemplate.Template
}

func NewRenderer() (*Renderer, error) {
	htmlTemplates, err := template.ParseFS(templateFiles, "*.html")
	if err != nil {
		return nil, fmt.Errorf("parse html templates: %w", err)
	}

	textTemplates, err := texttemplate.ParseFS(templateFiles, "*.txt")
	if err != nil {
		return nil, fmt.Errorf("parse text templates: %w", err)
	}

	return &Renderer{
		htmlTemplates: htmlTemplates,
		textTemplates: textTemplates,
	}, nil
}

func (renderer *Renderer) RenderApplicationSubmitted(data ApplicationSubmittedData) (string, string, error) {
	sanitizedData := sanitizeApplicationSubmittedData(data)
	var textBuffer bytes.Buffer
	if err := renderer.textTemplates.ExecuteTemplate(&textBuffer, "application_submitted.txt", sanitizedData); err != nil {
		return "", "", fmt.Errorf("render text template: %w", err)
	}

	var htmlBuffer bytes.Buffer
	if err := renderer.htmlTemplates.ExecuteTemplate(&htmlBuffer, "application_submitted.html", sanitizedData); err != nil {
		return "", "", fmt.Errorf("render html template: %w", err)
	}

	return textBuffer.String(), htmlBuffer.String(), nil
}

func (renderer *Renderer) RenderApplicationApproved(data ApplicationApprovedData) (string, string, error) {
	sanitizedData := sanitizeApplicationApprovedData(data)
	var textBuffer bytes.Buffer
	if err := renderer.textTemplates.ExecuteTemplate(&textBuffer, "application_approved.txt", sanitizedData); err != nil {
		return "", "", fmt.Errorf("render text template: %w", err)
	}

	var htmlBuffer bytes.Buffer
	if err := renderer.htmlTemplates.ExecuteTemplate(&htmlBuffer, "application_approved.html", sanitizedData); err != nil {
		return "", "", fmt.Errorf("render html template: %w", err)
	}

	return textBuffer.String(), htmlBuffer.String(), nil
}

func (renderer *Renderer) RenderApplicationRejected(data ApplicationRejectedData) (string, string, error) {
	sanitizedData := sanitizeApplicationRejectedData(data)
	var textBuffer bytes.Buffer
	if err := renderer.textTemplates.ExecuteTemplate(&textBuffer, "application_rejected.txt", sanitizedData); err != nil {
		return "", "", fmt.Errorf("render text template: %w", err)
	}

	var htmlBuffer bytes.Buffer
	if err := renderer.htmlTemplates.ExecuteTemplate(&htmlBuffer, "application_rejected.html", sanitizedData); err != nil {
		return "", "", fmt.Errorf("render html template: %w", err)
	}

	return textBuffer.String(), htmlBuffer.String(), nil
}

func (renderer *Renderer) RenderApplicationRFI(data ApplicationRFIData) (string, string, error) {
	sanitizedData := sanitizeApplicationRFIData(data)
	var textBuffer bytes.Buffer
	if err := renderer.textTemplates.ExecuteTemplate(&textBuffer, "application_rfi.txt", sanitizedData); err != nil {
		return "", "", fmt.Errorf("render text template: %w", err)
	}

	var htmlBuffer bytes.Buffer
	if err := renderer.htmlTemplates.ExecuteTemplate(&htmlBuffer, "application_rfi.html", sanitizedData); err != nil {
		return "", "", fmt.Errorf("render html template: %w", err)
	}

	return textBuffer.String(), htmlBuffer.String(), nil
}

func sanitizeApplicationSubmittedData(data ApplicationSubmittedData) ApplicationSubmittedData {
	data.FullName = sanitizeEmailText(data.FullName)
	data.ProductName = sanitizeEmailText(data.ProductName)
	data.ApplicationID = sanitizeEmailText(data.ApplicationID)
	data.SumAssuredFormatted = sanitizeEmailText(data.SumAssuredFormatted)
	data.PremiumFormatted = sanitizeEmailText(data.PremiumFormatted)
	data.PaymentFrequency = sanitizeEmailText(data.PaymentFrequency)
	data.PortalURL = sanitizeEmailURL(data.PortalURL)
	return data
}

func sanitizeApplicationApprovedData(data ApplicationApprovedData) ApplicationApprovedData {
	data.FullName = sanitizeEmailText(data.FullName)
	data.PolicyNumber = sanitizeEmailText(data.PolicyNumber)
	data.ProductName = sanitizeEmailText(data.ProductName)
	data.ApplicationID = sanitizeEmailText(data.ApplicationID)
	data.SumAssuredFormatted = sanitizeEmailText(data.SumAssuredFormatted)
	data.ProtectionPeriod = sanitizeEmailText(data.ProtectionPeriod)
	data.PremiumFormatted = sanitizeEmailText(data.PremiumFormatted)
	data.PolicyDownloadURL = sanitizeEmailURL(data.PolicyDownloadURL)
	data.PortalURL = sanitizeEmailURL(data.PortalURL)
	return data
}

func sanitizeApplicationRejectedData(data ApplicationRejectedData) ApplicationRejectedData {
	data.FullName = sanitizeEmailText(data.FullName)
	data.ProductName = sanitizeEmailText(data.ProductName)
	data.ApplicationID = sanitizeEmailText(data.ApplicationID)
	data.RejectionCode = sanitizeEmailText(data.RejectionCode)
	data.RejectionReason = sanitizeEmailText(data.RejectionReason)
	data.LeadUnderwriterName = sanitizeEmailText(data.LeadUnderwriterName)
	data.LeadUnderwriterNIP = sanitizeEmailText(data.LeadUnderwriterNIP)
	data.RefundAmountFormatted = sanitizeEmailText(data.RefundAmountFormatted)
	data.ConsultationURL = sanitizeEmailURL(data.ConsultationURL)
	return data
}

func sanitizeApplicationRFIData(data ApplicationRFIData) ApplicationRFIData {
	data.FullName = sanitizeEmailText(data.FullName)
	data.ProductName = sanitizeEmailText(data.ProductName)
	data.ApplicationID = sanitizeEmailText(data.ApplicationID)
	data.Notes = sanitizeEmailText(data.Notes)
	for i, doc := range data.RequiredDocs {
		data.RequiredDocs[i] = sanitizeEmailText(doc)
	}
	data.SLADeadline = sanitizeEmailText(data.SLADeadline)
	data.UploadPortalURL = sanitizeEmailURL(data.UploadPortalURL)
	return data
}

func sanitizeEmailURL(value string) string {
	value = sanitizeEmailText(value)
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") || strings.HasPrefix(lower, "vbscript:") {
		return "#"
	}
	return value
}

func sanitizeEmailText(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}

// FormatIDR formats an integer amount as Indonesian Rupiah (e.g. 1000000 -> "Rp 1.000.000")
func FormatIDR(amount int64) string {
	sign := ""
	var uAmount uint64
	if amount < 0 {
		sign = "-"
		// Safe conversion preventing math.MinInt64 two's complement negation overflow
		uAmount = uint64(-(amount + 1)) + 1
	} else {
		uAmount = uint64(amount)
	}
	s := strconv.FormatUint(uAmount, 10)
	n := len(s)
	if n <= 3 {
		return fmt.Sprintf("%sRp %s", sign, s)
	}

	var b strings.Builder
	b.WriteString(sign)
	b.WriteString("Rp ")

	remainder := n % 3
	if remainder > 0 {
		b.WriteString(s[:remainder])
		if remainder < n {
			b.WriteString(".")
		}
	}

	for i := remainder; i < n; i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < n {
			b.WriteString(".")
		}
	}

	return b.String()
}
