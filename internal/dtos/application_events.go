package dtos

const (
	TopicApplicationSubmitted    = "application.submitted"
	TopicApplicationApproved     = "application.approved"
	TopicApplicationRejected     = "application.rejected"
	TopicApplicationRFIRequested = "application.rfi_requested"
)

type ApplicationSubmittedEvent struct {
	ApplicationID    string `json:"application_id"`
	ProductID        string `json:"product_id"`
	ProductName      string `json:"product_name"`
	FullName         string `json:"full_name"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	SumAssured       int64  `json:"sum_assured"`
	Premium          int64  `json:"premium"`
	PaymentFrequency string `json:"payment_frequency"`
	Status           string `json:"status"`
}

type ApplicationApprovedEvent struct {
	ApplicationID    string `json:"application_id"`
	PolicyNumber     string `json:"policy_number"`
	ProductID        string `json:"product_id"`
	ProductName      string `json:"product_name"`
	FullName         string `json:"full_name"`
	Email            string `json:"email"`
	SumAssured       int64  `json:"sum_assured"`
	Premium          int64  `json:"premium"`
	PaymentTerm      int    `json:"payment_term"`
	PaymentFrequency string `json:"payment_frequency"`
	ReviewedBy       string `json:"reviewed_by"`
}

type ApplicationRejectedEvent struct {
	ApplicationID       string `json:"application_id"`
	ProductID           string `json:"product_id"`
	ProductName         string `json:"product_name"`
	FullName            string `json:"full_name"`
	Email               string `json:"email"`
	Premium             int64  `json:"premium"`
	RejectionCode       string `json:"rejection_code"`
	RejectionReason     string `json:"rejection_reason"`
	LeadUnderwriterName string `json:"lead_underwriter_name"`
	LeadUnderwriterNIP  string `json:"lead_underwriter_nip"`
}

type ApplicationRFIRequestedEvent struct {
	ApplicationID string   `json:"application_id"`
	ProductID     string   `json:"product_id"`
	ProductName   string   `json:"product_name"`
	FullName      string   `json:"full_name"`
	Email         string   `json:"email"`
	Notes         string   `json:"notes"`
	RequiredDocs  []string `json:"required_docs"`
	SLADeadline   string   `json:"sla_deadline"`
}
