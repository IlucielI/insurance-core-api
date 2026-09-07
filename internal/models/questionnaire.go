package models

import "time"

type Questionnaire struct {
	ID          string     `gorm:"primaryKey;type:varchar(64)" json:"id"`
	ProductID   *string    `gorm:"type:varchar(64);index" json:"product_id,omitempty"`
	Product     *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Category    string     `gorm:"type:varchar(32);not null;default:'all';index" json:"category"`
	Title       string     `gorm:"type:varchar(120);not null" json:"title"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	Version     int        `gorm:"not null;default:1" json:"version"`
	IsActive    bool       `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Questions   []Question `gorm:"foreignKey:QuestionnaireID" json:"questions,omitempty"`
}

func (Questionnaire) TableName() string {
	return "questionnaires"
}

type QuestionOption struct {
	Value       string  `json:"value"`
	Label       string  `json:"label"`
	Multiplier  float64 `json:"multiplier,omitempty"` // actuarial pricing multiplier (e.g. 1.25, 0.95, 1.0)
	RiskWeight  float64 `json:"risk_weight,omitempty"`
	RiskImpact  string  `json:"risk_impact,omitempty"`
	RfiRequired bool    `json:"rfi_required,omitempty"`
}

type Question struct {
	ID                  string           `gorm:"primaryKey;type:varchar(64)" json:"id"`
	QuestionnaireID     string           `gorm:"type:varchar(64);not null;index" json:"questionnaire_id"`
	StepNumber          int              `gorm:"not null;index" json:"step_number"`
	PillarType          string           `gorm:"type:varchar(32);not null;index" json:"pillar_type"`
	Code                string           `gorm:"type:varchar(64);not null" json:"code"`
	Label               string           `gorm:"type:text;not null" json:"label"`
	HelpText            string           `gorm:"type:text" json:"help_text,omitempty"`
	InputType           string           `gorm:"type:varchar(32);not null" json:"input_type"`
	Placeholder         string           `gorm:"type:varchar(120)" json:"placeholder,omitempty"`
	OrderIndex          int              `gorm:"not null;default:0" json:"order_index"`
	ValidationRules     map[string]any   `gorm:"type:jsonb;serializer:json" json:"validation_rules"`
	Options             []QuestionOption `gorm:"type:jsonb;serializer:json" json:"options"`
	ParentQuestionID    *string          `gorm:"type:varchar(64);index" json:"parent_question_id,omitempty"`
	ShowIfParentValue   any              `gorm:"type:jsonb;serializer:json" json:"show_if_parent_value,omitempty"`
	AffectsPricingField string           `gorm:"type:varchar(64)" json:"affects_pricing_field,omitempty"`
	UnderwritingRules   map[string]any   `gorm:"type:jsonb;serializer:json" json:"underwriting_rules"`
	IsActive            bool             `gorm:"not null;default:true" json:"is_active"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

func (Question) TableName() string {
	return "questions"
}

type ApplicationAnswer struct {
	ID            string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	ApplicationID string    `gorm:"type:varchar(64);not null;index" json:"application_id"`
	QuestionID    string    `gorm:"type:varchar(64);not null;index" json:"question_id"`
	Code          string    `gorm:"type:varchar(64);not null" json:"code"`
	AnswerValue   any       `gorm:"type:jsonb;serializer:json" json:"answer_value"`
	CreatedAt     time.Time `json:"created_at"`
}

func (ApplicationAnswer) TableName() string {
	return "application_answers"
}
