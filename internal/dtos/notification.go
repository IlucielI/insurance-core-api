package dtos

type CreateNotificationRequest struct {
	Type     string `json:"type"`
	Category string `json:"category"`
	Severity string `json:"severity,omitempty"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Link     string `json:"link,omitempty"`
}

type NotificationResponse struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Category  string  `json:"category"`
	Severity  string  `json:"severity"`
	Title     string  `json:"title"`
	Message   string  `json:"message"`
	Link      string  `json:"link"`
	IsRead    bool    `json:"is_read"`
	CreatedAt string  `json:"created_at"`
	ReadAt    *string `json:"read_at,omitempty"`
}

type NotificationQuery struct {
	Category   string `query:"category"`
	Severity   string `query:"severity"`
	UnreadOnly *bool  `query:"unread_only"`
	Limit      int    `query:"limit"`
	Offset     int    `query:"offset"`
}

type NotificationListResponse struct {
	Data        []NotificationResponse `json:"data"`
	Total       int64                  `json:"total"`
	UnreadCount int64                  `json:"unread_count"`
}

type MarkReadResponse struct {
	ID     string `json:"id"`
	IsRead bool   `json:"is_read"`
	ReadAt string `json:"read_at"`
}

type MarkAllReadResponse struct {
	UpdatedCount int64 `json:"updated_count"`
	UnreadCount  int64 `json:"unread_count"`
}
