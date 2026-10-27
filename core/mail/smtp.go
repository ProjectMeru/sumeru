// Package mail sends transactional email via SMTP when configured in sumeru.conf.
package mail

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"sumeru/core/applog"
)

// SMTPConfig holds outbound mail settings from INI.
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

var smtpCfg SMTPConfig

// Configure sets global SMTP settings (call once at startup).
func Configure(c SMTPConfig) {
	smtpCfg = c
	if smtpCfg.Port <= 0 {
		smtpCfg.Port = 587
	}
}

// Configured reports whether SMTP host and from address are set.
func Configured() bool {
	return strings.TrimSpace(smtpCfg.Host) != "" && strings.TrimSpace(smtpCfg.From) != ""
}

// Send delivers a plain-text message to one recipient.
func Send(ctx context.Context, to, subject, body string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("recipient required")
	}
	if !Configured() {
		return fmt.Errorf("smtp not configured")
	}
	from := strings.TrimSpace(smtpCfg.From)
	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", strings.TrimSpace(smtpCfg.Host), smtpCfg.Port)
	var auth smtp.Auth
	if u := strings.TrimSpace(smtpCfg.User); u != "" {
		auth = smtp.PlainAuth("", u, smtpCfg.Password, smtpCfg.Host)
	}

	if smtpCfg.Port == 465 {
		return sendSMTPS(addr, auth, from, []string{to}, []byte(msg))
	}
	if err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg)); err != nil {
		applog.Warn(ctx, applog.Event{
			Message:   "smtp send failed",
			Component: "mail",
			Operation: "send",
			Status:    "failed",
			Context:   map[string]interface{}{"to": to, "subject": subject},
			Err:       err,
		})
		return err
	}
	return nil
}

func sendSMTPS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

// SendMultipart delivers text/plain and text/html alternative parts.
func SendMultipart(ctx context.Context, to, subject, textBody, htmlBody string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("recipient required")
	}
	if !Configured() {
		return fmt.Errorf("smtp not configured")
	}
	boundary := "sum-mail-boundary"
	from := strings.TrimSpace(smtpCfg.From)
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + "\r\n\r\n")
	writePart := func(contentType, body string) {
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString("Content-Type: " + contentType + "; charset=UTF-8\r\n")
		b.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
		b.WriteString(base64.StdEncoding.EncodeToString([]byte(body)))
		b.WriteString("\r\n")
	}
	writePart("text/plain", textBody)
	writePart("text/html", htmlBody)
	b.WriteString("--" + boundary + "--\r\n")
	return sendRawMessage(ctx, to, subject, []byte(b.String()))
}

func sendRawMessage(ctx context.Context, to, subject string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", strings.TrimSpace(smtpCfg.Host), smtpCfg.Port)
	var auth smtp.Auth
	if u := strings.TrimSpace(smtpCfg.User); u != "" {
		auth = smtp.PlainAuth("", u, smtpCfg.Password, smtpCfg.Host)
	}
	from := strings.TrimSpace(smtpCfg.From)
	if smtpCfg.Port == 465 {
		return sendSMTPS(addr, auth, from, []string{to}, msg)
	}
	if err := smtp.SendMail(addr, auth, from, []string{to}, msg); err != nil {
		applog.Warn(ctx, applog.Event{
			Message:   "smtp send failed",
			Component: "mail",
			Operation: "send",
			Status:    "failed",
			Context:   map[string]interface{}{"to": to, "subject": subject},
			Err:       err,
		})
		return err
	}
	return nil
}

// SendPasswordResetEmail notifies a user that an administrator requested a password reset.
func SendPasswordResetEmail(ctx context.Context, to, login, loginURL string) error {
	subject := "Sumeru password reset requested"
	body := fmt.Sprintf("Hello,\n\nAn administrator requested a password reset for account %q.\n\nOpen this link to choose a new password:\n%s\n\nIf you did not expect this message, contact your administrator.\n", login, loginURL)
	return Send(ctx, to, subject, body)
}
