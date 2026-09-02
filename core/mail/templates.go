package mail

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
)

// Template renders a named mail body with data.
type Template struct {
	Name    string
	Subject string
	Body    string
}

var mailTemplates = map[string]Template{}

// RegisterTemplate registers a mail template by name.
func RegisterTemplate(t Template) {
	name := strings.TrimSpace(t.Name)
	if name == "" {
		return
	}
	mailTemplates[name] = t
}

// RenderTemplate executes a registered template.
func RenderTemplate(name string, data map[string]interface{}) (subject, body string, err error) {
	t, ok := mailTemplates[strings.TrimSpace(name)]
	if !ok {
		return "", "", fmt.Errorf("mail template %q not found", name)
	}
	subject, err = executeTemplate(t.Subject, data)
	if err != nil {
		return "", "", err
	}
	body, err = executeTemplate(t.Body, data)
	if err != nil {
		return "", "", err
	}
	return subject, body, nil
}

// SendTemplate sends mail using a registered template.
func SendTemplate(ctx context.Context, to, templateName string, data map[string]interface{}) error {
	subject, body, err := RenderTemplate(templateName, data)
	if err != nil {
		return err
	}
	return Send(ctx, to, subject, body)
}

func executeTemplate(raw string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("mail").Parse(raw)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func init() {
	RegisterTemplate(Template{
		Name:    "report_delivery",
		Subject: "Report: {{.ReportName}}",
		Body:    "Hello,\n\nYour scheduled report {{.ReportName}} is ready.\n\n— Sumeru",
	})
}
