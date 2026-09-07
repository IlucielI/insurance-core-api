package dtos

type CreateAuditLogRequest struct {
	ActorName      string         `json:"actor_name"`
	ActorRole      string         `json:"actor_role"`
	Action         string         `json:"action"`
	Category       string         `json:"category"`
	TargetResource string         `json:"target_resource"`
	IPAddress      string         `json:"ip_address,omitempty"`
	Status         string         `json:"status,omitempty"`
	Details        map[string]any `json:"details,omitempty"`
}

type AuditLogResponse struct {
	ID             string         `json:"id"`
	Timestamp      string         `json:"timestamp"`
	ActorName      string         `json:"actor_name"`
	ActorRole      string         `json:"actor_role"`
	Action         string         `json:"action"`
	Category       string         `json:"category"`
	TargetResource string         `json:"target_resource"`
	IPAddress      string         `json:"ip_address"`
	Status         string         `json:"status"`
	Details        map[string]any `json:"details"`
	Hash           string         `json:"hash"`
}

type AuditLogQuery struct {
	Category string `query:"category"`
	Status   string `query:"status"`
	Search   string `query:"search"`
	Limit    int    `query:"limit"`
	Offset   int    `query:"offset"`
}

type AuditLogListResponse struct {
	Data  []AuditLogResponse `json:"data"`
	Total int64              `json:"total"`
}
