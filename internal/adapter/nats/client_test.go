package nats

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestNewClientValidatesConfig(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Fatal("NewClient() error = nil, want error")
	}
	if _, err := NewClient(Config{Host: "localhost"}); err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
}

func TestNormalizeServerURLWithPort(t *testing.T) {
	serverURL, err := normalizeServerURL("localhost:4222", 0)
	if err != nil {
		t.Fatalf("normalizeServerURL() error = %v", err)
	}
	if serverURL != "nats://localhost:4222" {
		t.Fatalf("normalizeServerURL() = %q", serverURL)
	}
}

func TestNormalizeServerURL(t *testing.T) {
	serverURL, err := normalizeServerURL("localhost", 4222)
	if err != nil {
		t.Fatalf("normalizeServerURL() error = %v", err)
	}
	if serverURL != "nats://localhost:4222" {
		t.Fatalf("normalizeServerURL() = %q", serverURL)
	}
}

func TestPublishJSON(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.connectFunc = func(serverURL string, options ...nats.Option) (*nats.Conn, error) {
		if serverURL != "nats://localhost:4222" {
			t.Fatalf("serverURL = %q", serverURL)
		}
		return &nats.Conn{}, nil
	}
	client.publishFunc = func(conn *nats.Conn, message *nats.Msg) error {
		if message.Subject != "events.created" || string(message.Data) != `{"id":"1"}` {
			t.Fatalf("message = %+v", message)
		}
		return nil
	}

	if err := client.PublishJSON(context.Background(), "events.created", map[string]string{"id": "1"}); err != nil {
		t.Fatalf("PublishJSON() error = %v", err)
	}
}

func TestRequestJSON(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost", Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.connectFunc = func(serverURL string, options ...nats.Option) (*nats.Conn, error) {
		return &nats.Conn{}, nil
	}
	client.requestFunc = func(ctx context.Context, conn *nats.Conn, message *nats.Msg, timeout time.Duration) (*nats.Msg, error) {
		if timeout != time.Second {
			t.Fatalf("timeout = %s", timeout)
		}
		if ctx == nil {
			t.Fatal("ctx is nil")
		}
		return &nats.Msg{Data: []byte(`{"ok":true}`)}, nil
	}

	var result map[string]bool
	if err := client.RequestJSON(context.Background(), "events.status", map[string]string{"id": "1"}, &result, 0); err != nil {
		t.Fatalf("RequestJSON() error = %v", err)
	}
	if !reflect.DeepEqual(result, map[string]bool{"ok": true}) {
		t.Fatalf("RequestJSON() = %+v", result)
	}
}

func TestPublishAndRequestCancelledContext(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := client.PublishJSON(ctx, "events.created", map[string]string{"id": "1"}); err == nil {
		t.Fatal("PublishJSON() error = nil, want context canceled")
	}

	if err := client.RequestJSON(ctx, "events.created", map[string]string{"id": "1"}, nil, 0); err == nil {
		t.Fatal("RequestJSON() error = nil, want context canceled")
	}
}

func TestSubscribeAndQueueSubscribe(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.connectFunc = func(serverURL string, options ...nats.Option) (*nats.Conn, error) {
		return &nats.Conn{}, nil
	}

	var capturedSubject, capturedGroup string
	var capturedHandler nats.MsgHandler

	client.subscribeFunc = func(conn *nats.Conn, subject, queueGroup string, handler nats.MsgHandler) (*nats.Subscription, error) {
		capturedSubject = subject
		capturedGroup = queueGroup
		capturedHandler = handler
		return &nats.Subscription{}, nil
	}

	var receivedData []byte
	_, err = client.QueueSubscribe("orders.created", "order-workers", func(ctx context.Context, data []byte) error {
		receivedData = data
		return nil
	})
	if err != nil {
		t.Fatalf("QueueSubscribe() error = %v", err)
	}
	if capturedSubject != "orders.created" || capturedGroup != "order-workers" {
		t.Fatalf("QueueSubscribe captured subject=%q group=%q", capturedSubject, capturedGroup)
	}

	// Trigger handler
	capturedHandler(&nats.Msg{Subject: "orders.created", Data: []byte(`{"id":"order-1"}`)})
	if string(receivedData) != `{"id":"order-1"}` {
		t.Fatalf("handler receivedData = %s", string(receivedData))
	}

	// Regular subscribe
	_, err = client.Subscribe("orders.updated", func(ctx context.Context, data []byte) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	if capturedSubject != "orders.updated" || capturedGroup != "" {
		t.Fatalf("Subscribe captured subject=%q group=%q", capturedSubject, capturedGroup)
	}
}

func TestSubscribeRecoversFromPanic(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.connectFunc = func(serverURL string, options ...nats.Option) (*nats.Conn, error) {
		return &nats.Conn{}, nil
	}

	var capturedHandler nats.MsgHandler
	client.subscribeFunc = func(conn *nats.Conn, subject, queueGroup string, handler nats.MsgHandler) (*nats.Subscription, error) {
		capturedHandler = handler
		return &nats.Subscription{}, nil
	}

	_, err = client.Subscribe("events.risky", func(ctx context.Context, data []byte) error {
		panic("simulated worker crash")
	})
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	// Invoking the message handler should recover cleanly without panicking out of the test
	capturedHandler(&nats.Msg{Subject: "events.risky", Data: []byte(`{}`)})
}

func TestCloseThreadSafe(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.connectFunc = func(serverURL string, options ...nats.Option) (*nats.Conn, error) {
		return &nats.Conn{}, nil
	}
	var closed bool
	client.closeFunc = func(conn *nats.Conn) {
		closed = true
	}

	// Close on nil conn
	client.Close()

	// Ensure conn and close
	_, err = client.ensureConn()
	if err != nil {
		t.Fatalf("ensureConn() error = %v", err)
	}
	client.Close()
	if !closed {
		t.Fatal("closeFunc was not called")
	}
	if client.conn != nil {
		t.Fatal("client.conn != nil after Close()")
	}
}

func TestBuildJSONMessageValidatesInput(t *testing.T) {
	if _, err := buildJSONMessage(" ", map[string]string{"id": "1"}); err == nil {
		t.Fatal("buildJSONMessage() error = nil, want error")
	}
	if _, err := buildJSONMessage("events.created", func() {}); err == nil {
		t.Fatal("buildJSONMessage() error = nil, want marshal error")
	}
}

func TestNormalizeServerURLRejectsPath(t *testing.T) {
	if _, err := normalizeServerURL("nats://example.com/foo", 4222); err == nil {
		t.Fatal("normalizeServerURL() error = nil, want path validation error")
	}
}

func TestNormalizeServerURLRejectsInvalidScheme(t *testing.T) {
	if _, err := normalizeServerURL("ftp://example.com", 4222); err == nil {
		t.Fatal("normalizeServerURL() error = nil, want error")
	}
}
