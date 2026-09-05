package constants

import "errors"

const (
	ErrAssistantBodyInvalid        = "invalid assistant request body"
	ErrAssistantMessageRequired    = "message is required"
	ErrAssistantMessageTooLong     = "message is too long"
	ErrAssistantServiceUnavailable = "assistant service is unavailable"
	ErrAssistantChatFailed         = "failed to answer assistant request"
)

var ErrAssistantMessageRequiredError = errors.New(ErrAssistantMessageRequired)
var ErrAssistantMessageTooLongError = errors.New(ErrAssistantMessageTooLong)
var ErrAssistantServiceUnavailableError = errors.New(ErrAssistantServiceUnavailable)

const (
	AssistantSourceTypeProduct         = "product"
	AssistantSourceTypeQuote           = "quote"
	AssistantSourceTypeApplicationFlow = "application_flow"
	AssistantSourceTypeUnderwriting    = "underwriting"
	AssistantSourceTypeCompany         = "company"
	AssistantSourceTypeFAQ             = "faq"
)

const (
	AssistantEmbeddingDimension = 1024
	AssistantChunkTokenLimit    = 1024
	AssistantChunkOverlap       = 128
	AssistantTopK               = 4
	AssistantMaxMessageSize     = 4000
	AssistantMaxDistance        = 0.65
)
