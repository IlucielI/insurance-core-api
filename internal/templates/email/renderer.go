package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
	texttemplate "text/template"
)

//go:embed *.html *.txt
var templateFiles embed.FS

type ApplicationSubmittedData struct {
	FullName      string
	ProductName   string
	ApplicationID string
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
	sanitizedData := sanitizeApplicationSubmittedTextData(data)
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

func sanitizeApplicationSubmittedTextData(data ApplicationSubmittedData) ApplicationSubmittedData {
	data.FullName = sanitizeEmailText(data.FullName)
	data.ProductName = sanitizeEmailText(data.ProductName)
	data.ApplicationID = sanitizeEmailText(data.ApplicationID)
	return data
}

func sanitizeEmailText(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}
