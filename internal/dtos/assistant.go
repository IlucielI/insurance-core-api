package dtos

type AssistantChatRequest struct {
	Message     string               `json:"message"`
	ProductSlug string               `json:"product_slug,omitempty"`
	Quote       *ProductQuoteRequest `json:"quote,omitempty"`
	Slug        string               `json:"slug,omitempty"`
}

type AssistantSource struct {
	Title      string  `json:"title"`
	SourceType string  `json:"source_type"`
	Score      float64 `json:"score"`
	Excerpt    string  `json:"excerpt"`
}

type AssistantChatResponse struct {
	Answer  string            `json:"answer"`
	Sources []AssistantSource `json:"sources"`
}
