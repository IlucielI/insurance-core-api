package dtos

type ServiceHealthStatus string

const (
	ServiceHealthOnline   ServiceHealthStatus = "online"
	ServiceHealthDegraded ServiceHealthStatus = "degraded"
	ServiceHealthOffline  ServiceHealthStatus = "offline"
)

type ServiceHealthItem struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Type             string              `json:"type"`
	Endpoint         string              `json:"endpoint"`
	Status           ServiceHealthStatus `json:"status"`
	LatencyMs        float64             `json:"latency_ms"`
	UptimePercentage float64             `json:"uptime_percentage"`
	LastChecked      string              `json:"last_checked"`
}

type DatabasePoolStats struct {
	OpenConnections    int `json:"open_connections"`
	InUse              int `json:"in_use"`
	Idle               int `json:"idle"`
	MaxOpenConnections int `json:"max_open_connections"`
}

type SystemHealthOverviewResponse struct {
	OverallStatus       ServiceHealthStatus `json:"overall_status"`
	ActiveServicesCount int                 `json:"active_services_count"`
	TotalServicesCount  int                 `json:"total_services_count"`
	AvgLatencyMs        float64             `json:"avg_latency_ms"`
	Services            []ServiceHealthItem `json:"services"`
	DatabaseStats       DatabasePoolStats   `json:"database_stats"`
	RecentAuditLogs     []AuditLogResponse  `json:"recent_audit_logs"`
	Uptime              string              `json:"uptime,omitempty"`
	Version             string              `json:"version,omitempty"`
	GitHash             string              `json:"git_hash,omitempty"`
}

type PingServicesRequest struct {
	ServiceID string `json:"service_id,omitempty"`
}

type PingServicesResponse struct {
	Services []ServiceHealthItem `json:"services"`
}

type RouteLatencyProbeItem struct {
	ID          string  `json:"id"`
	Method      string  `json:"method"`
	Path        string  `json:"path"`
	LatencyMs   float64 `json:"latency_ms"`
	StatusCode  int     `json:"status_code"`
	Description string  `json:"description"`
}

type PingRoutesResponse struct {
	Routes []RouteLatencyProbeItem `json:"routes"`
}
