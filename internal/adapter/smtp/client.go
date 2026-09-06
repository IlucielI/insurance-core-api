package smtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	stdsmtp "net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/ports"
)

type Encryption string

const (
	EncryptionNone     Encryption = "none"
	EncryptionStartTLS Encryption = "starttls"
	EncryptionTLS      Encryption = "tls"
)

type Config struct {
	Host       string
	Port       int
	Username   string
	Password   string
	FromEmail  string
	FromName   string
	Encryption Encryption
	Timeout    time.Duration
}

type Client struct {
	host       string
	port       int
	username   string
	password   string
	fromEmail  string
	fromName   string
	encryption Encryption
	timeout    time.Duration
	dial       func(ctx context.Context, network string, address string) (net.Conn, error)
}

func NewClient(config Config) (*Client, error) {
	host := strings.TrimSpace(config.Host)
	if host == "" {
		return nil, errors.New("Host is required")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return nil, errors.New("Port must be between 1 and 65535")
	}

	fromEmail := strings.TrimSpace(config.FromEmail)
	if fromEmail == "" {
		return nil, errors.New("FromEmail is required")
	}

	encryption := config.Encryption
	if encryption == "" {
		encryption = EncryptionStartTLS
	}
	if encryption != EncryptionNone && encryption != EncryptionStartTLS && encryption != EncryptionTLS {
		return nil, fmt.Errorf("unsupported Encryption %q", encryption)
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	dialer := &net.Dialer{Timeout: timeout}
	return &Client{
		host:       host,
		port:       config.Port,
		username:   strings.TrimSpace(config.Username),
		password:   config.Password,
		fromEmail:  fromEmail,
		fromName:   strings.TrimSpace(config.FromName),
		encryption: encryption,
		timeout:    timeout,
		dial:       dialer.DialContext,
	}, nil
}

func (client *Client) Send(ctx context.Context, message ports.EmailMessage) error {
	if err := validateMessage(message); err != nil {
		return err
	}

	address := net.JoinHostPort(client.host, strconv.Itoa(client.port))
	conn, err := client.dial(ctx, "tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return err
		}
	} else if client.timeout > 0 {
		if err := conn.SetDeadline(time.Now().Add(client.timeout)); err != nil {
			return err
		}
	}

	if client.encryption == EncryptionTLS {
		conn = tls.Client(conn, &tls.Config{ServerName: client.host, MinVersion: tls.VersionTLS12})
		if err := conn.(*tls.Conn).HandshakeContext(ctx); err != nil {
			return err
		}
	}

	smtpClient, err := stdsmtp.NewClient(conn, client.host)
	if err != nil {
		return err
	}
	defer smtpClient.Quit()

	if client.encryption == EncryptionStartTLS {
		ok, extensionParam := smtpClient.Extension("STARTTLS")
		if !ok {
			return fmt.Errorf("smtp server does not support STARTTLS: %s", extensionParam)
		}

		if err := smtpClient.StartTLS(&tls.Config{ServerName: client.host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}

	if client.username != "" {
		if err := smtpClient.Auth(stdsmtp.PlainAuth("", client.username, client.password, client.host)); err != nil {
			return err
		}
	}

	if err := smtpClient.Mail(client.fromEmail); err != nil {
		return err
	}
	for _, recipient := range message.To {
		if err := smtpClient.Rcpt(strings.TrimSpace(recipient)); err != nil {
			return err
		}
	}

	writer, err := smtpClient.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(buildMIMEMessage(client.fromAddress(), message)); err != nil {
		writer.Close()
		return err
	}

	return writer.Close()
}

func validateMessage(message ports.EmailMessage) error {
	if len(message.To) == 0 {
		return errors.New("To is required")
	}
	for _, recipient := range message.To {
		if strings.TrimSpace(recipient) == "" {
			return errors.New("To contains an empty recipient")
		}
		if _, err := mail.ParseAddress(recipient); err != nil {
			return fmt.Errorf("invalid email address %q: %w", recipient, err)
		}
	}
	if strings.TrimSpace(message.Subject) == "" {
		return errors.New("Subject is required")
	}
	if strings.TrimSpace(message.TextBody) == "" && strings.TrimSpace(message.HTMLBody) == "" {
		return errors.New("TextBody or HTMLBody is required")
	}
	return nil
}

func (client *Client) fromAddress() string {
	if client.fromName == "" {
		return client.fromEmail
	}
	return mime.QEncoding.Encode("utf-8", client.fromName) + " <" + client.fromEmail + ">"
}

func buildMIMEMessage(from string, message ports.EmailMessage) []byte {
	var buffer bytes.Buffer
	writeHeader(&buffer, "From", from)
	writeHeader(&buffer, "To", strings.Join(message.To, ", "))
	writeHeader(&buffer, "Subject", mime.QEncoding.Encode("utf-8", message.Subject))
	writeHeader(&buffer, "MIME-Version", "1.0")

	textBody := sanitizeEmailBody(message.TextBody)
	htmlBody := sanitizeEmailBody(message.HTMLBody)
	if strings.TrimSpace(htmlBody) == "" {
		writeHeader(&buffer, "Content-Type", `text/plain; charset="UTF-8"`)
		writeHeader(&buffer, "Content-Transfer-Encoding", "8bit")
		buffer.WriteString("\r\n")
		buffer.WriteString(textBody)
		return buffer.Bytes()
	}
	if strings.TrimSpace(textBody) == "" {
		writeHeader(&buffer, "Content-Type", `text/html; charset="UTF-8"`)
		writeHeader(&buffer, "Content-Transfer-Encoding", "8bit")
		buffer.WriteString("\r\n")
		buffer.WriteString(htmlBody)
		return buffer.Bytes()
	}

	boundary := "insurance-core-api-boundary"
	writeHeader(&buffer, "Content-Type", `multipart/alternative; boundary="`+boundary+`"`)
	buffer.WriteString("\r\n")
	buffer.WriteString("--" + boundary + "\r\n")
	buffer.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	buffer.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	buffer.WriteString(textBody + "\r\n")
	buffer.WriteString("--" + boundary + "\r\n")
	buffer.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	buffer.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	buffer.WriteString(htmlBody + "\r\n")
	buffer.WriteString("--" + boundary + "--\r\n")
	return buffer.Bytes()
}

func writeHeader(buffer *bytes.Buffer, key string, value string) {
	buffer.WriteString(key)
	buffer.WriteString(": ")
	buffer.WriteString(sanitizeHeaderValue(value))
	buffer.WriteString("\r\n")
}

func sanitizeHeaderValue(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return value
}

func sanitizeEmailBody(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}
