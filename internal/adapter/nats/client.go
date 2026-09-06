package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

const defaultNATSTimeout = 5 * time.Second

type Config struct {
	Host    string
	Port    int
	Token   string
	Name    string
	Timeout time.Duration
}

type Client struct {
	serverURL string
	token     string
	name      string
	timeout   time.Duration

	mu            sync.Mutex
	conn          *nats.Conn
	connectFunc   func(string, ...nats.Option) (*nats.Conn, error)
	closeFunc     func(*nats.Conn)
	publishFunc   func(*nats.Conn, *nats.Msg) error
	requestFunc   func(context.Context, *nats.Conn, *nats.Msg, time.Duration) (*nats.Msg, error)
	subscribeFunc func(conn *nats.Conn, subject, queueGroup string, handler nats.MsgHandler) (*nats.Subscription, error)
}

func NewClient(config Config) (*Client, error) {
	serverURL, err := normalizeServerURL(config.Host, config.Port)
	if err != nil {
		return nil, err
	}

	client := &Client{
		serverURL: serverURL,
		token:     strings.TrimSpace(config.Token),
		name:      strings.TrimSpace(config.Name),
		timeout:   normalizeNATSTimeout(config.Timeout),
	}
	client.connectFunc = nats.Connect
	client.closeFunc = func(conn *nats.Conn) {
		conn.Close()
	}
	client.publishFunc = func(conn *nats.Conn, message *nats.Msg) error {
		return conn.PublishMsg(message)
	}
	client.requestFunc = func(ctx context.Context, conn *nats.Conn, message *nats.Msg, timeout time.Duration) (*nats.Msg, error) {
		reqCtx := ctx
		if timeout > 0 {
			var cancel context.CancelFunc
			reqCtx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		return conn.RequestMsgWithContext(reqCtx, message)
	}
	client.subscribeFunc = func(conn *nats.Conn, subject, queueGroup string, handler nats.MsgHandler) (*nats.Subscription, error) {
		if strings.TrimSpace(queueGroup) != "" {
			return conn.QueueSubscribe(subject, queueGroup, handler)
		}
		return conn.Subscribe(subject, handler)
	}
	return client, nil
}

func (client *Client) Close() {
	if client == nil {
		return
	}
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.conn == nil {
		return
	}
	if client.closeFunc != nil {
		client.closeFunc(client.conn)
	} else {
		client.conn.Close()
	}
	client.conn = nil
}

func (client *Client) PublishJSON(ctx context.Context, subject string, payload any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	message, err := buildJSONMessage(subject, payload)
	if err != nil {
		return err
	}

	conn, err := client.ensureConn()
	if err != nil {
		return err
	}

	return client.publishFunc(conn, message)
}

func (client *Client) RequestJSON(ctx context.Context, subject string, payload any, result any, timeout time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	message, err := buildJSONMessage(subject, payload)
	if err != nil {
		return err
	}

	conn, err := client.ensureConn()
	if err != nil {
		return err
	}

	response, err := client.requestFunc(ctx, conn, message, normalizeRequestTimeout(timeout, client.timeout))
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	return json.Unmarshal(response.Data, result)
}

func (client *Client) Subscribe(subject string, handler func(context.Context, []byte) error) (*nats.Subscription, error) {
	return client.QueueSubscribe(subject, "", handler)
}

func (client *Client) QueueSubscribe(subject, queueGroup string, handler func(context.Context, []byte) error) (*nats.Subscription, error) {
	if strings.TrimSpace(subject) == "" {
		return nil, errors.New("subject is required")
	}
	if handler == nil {
		return nil, errors.New("handler is required")
	}

	conn, err := client.ensureConn()
	if err != nil {
		return nil, err
	}

	return client.subscribeFunc(conn, subject, strings.TrimSpace(queueGroup), func(msg *nats.Msg) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[NATS] panic recovered in topic %s: %v", subject, r)
			}
		}()

		ctx := context.Background()
		if err := handler(ctx, msg.Data); err != nil {
			log.Printf("[NATS] handler error in topic %s: %v", subject, err)
		}
	})
}

func (client *Client) ensureConn() (*nats.Conn, error) {
	if client == nil {
		return nil, errors.New("nats client is required")
	}

	client.mu.Lock()
	defer client.mu.Unlock()
	if client.timeout <= 0 {
		client.timeout = defaultNATSTimeout
	}
	if client.connectFunc == nil {
		client.connectFunc = nats.Connect
	}
	if client.closeFunc == nil {
		client.closeFunc = func(conn *nats.Conn) {
			conn.Close()
		}
	}
	if client.publishFunc == nil {
		client.publishFunc = func(conn *nats.Conn, message *nats.Msg) error {
			return conn.PublishMsg(message)
		}
	}
	if client.requestFunc == nil {
		client.requestFunc = func(ctx context.Context, conn *nats.Conn, message *nats.Msg, timeout time.Duration) (*nats.Msg, error) {
			reqCtx := ctx
			if timeout > 0 {
				var cancel context.CancelFunc
				reqCtx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}
			return conn.RequestMsgWithContext(reqCtx, message)
		}
	}
	if client.subscribeFunc == nil {
		client.subscribeFunc = func(conn *nats.Conn, subject, queueGroup string, handler nats.MsgHandler) (*nats.Subscription, error) {
			if strings.TrimSpace(queueGroup) != "" {
				return conn.QueueSubscribe(subject, queueGroup, handler)
			}
			return conn.Subscribe(subject, handler)
		}
	}

	if client.conn != nil && client.conn.IsConnected() {
		return client.conn, nil
	}

	options := []nats.Option{
		nats.Timeout(client.timeout),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Printf("[NATS] disconnected: %v", err)
			} else {
				log.Printf("[NATS] disconnected")
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("[NATS] reconnected to %s", nc.ConnectedUrl())
		}),
	}
	if client.name != "" {
		options = append(options, nats.Name(client.name))
	}
	if client.token != "" {
		options = append(options, nats.Token(client.token))
	}

	conn, err := client.connectFunc(client.serverURL, options...)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}
	client.conn = conn
	return client.conn, nil
}

func buildJSONMessage(subject string, payload any) (*nats.Msg, error) {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return nil, errors.New("subject is required")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	return &nats.Msg{Subject: subject, Data: body}, nil
}

func normalizeServerURL(host string, port int) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", errors.New("Host is required")
	}
	if !strings.Contains(host, "://") && strings.Contains(host, ":") {
		if strings.ContainsAny(host, "/?#") {
			return "", errors.New("Host must be a hostname or base URL without path")
		}
		return "nats://" + host, nil
	}
	if strings.Contains(host, "://") {
		parsed, err := url.Parse(host)
		if err != nil {
			return "", errors.New("Host must be a valid URL or hostname")
		}
		if parsed.Scheme != "nats" && parsed.Scheme != "tls" && parsed.Scheme != "ws" && parsed.Scheme != "wss" {
			return "", errors.New("Host scheme must be nats, tls, ws, or wss")
		}
		if parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return "", errors.New("Host must not contain path, query, or fragment")
		}
		return strings.TrimRight(host, "/"), nil
	}
	if strings.ContainsAny(host, "/?#") {
		return "", errors.New("Host must be a hostname or base URL without path")
	}
	if port <= 0 {
		port = 4222
	}
	return fmt.Sprintf("nats://%s:%d", host, port), nil
}

func normalizeNATSTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultNATSTimeout
	}
	return timeout
}

func normalizeRequestTimeout(timeout, fallback time.Duration) time.Duration {
	if timeout <= 0 {
		return fallback
	}
	return timeout
}
