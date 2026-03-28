package server

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	clientwebsocket "github.com/fasthttp/websocket"

	"github.com/iamrafaelmelo/simple-golang-chat/internal/config"
	"github.com/iamrafaelmelo/simple-golang-chat/internal/message"
)

func TestSetupAssignsAnonymousIdentity(t *testing.T) {
	url := startTestServer(t, config.Config{})

	connection := dialWebsocket(t, url, testOrigin(url))
	setup := readSetupMessage(t, connection)

	if setup.Type != "setup" {
		t.Fatalf("expected setup type, got %q", setup.Type)
	}
	if setup.Pid == "" {
		t.Fatal("expected server to assign a pid")
	}
	if !strings.HasPrefix(setup.Username, "Anonymous") {
		t.Fatalf("expected anonymous username, got %q", setup.Username)
	}
}

func TestMessageBroadcastUsesServerIdentity(t *testing.T) {
	url := startTestServer(t, config.Config{})
	sender := dialWebsocket(t, url, testOrigin(url))
	recipient := dialWebsocket(t, url, testOrigin(url))

	senderSetup := readSetupMessage(t, sender)
	recipientSetup := readSetupMessage(t, recipient)

	sendJSONMessage(t, sender, map[string]string{
		"type":     "message",
		"content":  "hello team",
		"pid":      recipientSetup.Pid,
		"username": "SpoofedUser",
	})

	outbound := readOutboundMessage(t, recipient)
	if outbound.Type != "message" {
		t.Fatalf("expected message type, got %q", outbound.Type)
	}
	if outbound.Pid != senderSetup.Pid {
		t.Fatalf("expected pid %q, got %q", senderSetup.Pid, outbound.Pid)
	}
	if outbound.Username != senderSetup.Username {
		t.Fatalf("expected username %q, got %q", senderSetup.Username, outbound.Username)
	}
	if outbound.Content != "hello team" {
		t.Fatalf("expected message content to be preserved, got %q", outbound.Content)
	}
}

func TestInvalidJSONDoesNotBroadcast(t *testing.T) {
	url := startTestServer(t, config.Config{})
	sender := dialWebsocket(t, url, testOrigin(url))
	recipient := dialWebsocket(t, url, testOrigin(url))

	_ = readSetupMessage(t, sender)
	_ = readSetupMessage(t, recipient)

	if err := sender.WriteMessage(clientwebsocket.TextMessage, []byte("{invalid")); err != nil {
		t.Fatalf("write invalid payload: %v", err)
	}

	if err := expectNoClientMessage(recipient, 250*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestTypingBroadcastUsesServerIdentity(t *testing.T) {
	url := startTestServer(t, config.Config{})
	sender := dialWebsocket(t, url, testOrigin(url))
	recipient := dialWebsocket(t, url, testOrigin(url))

	senderSetup := readSetupMessage(t, sender)
	recipientSetup := readSetupMessage(t, recipient)

	sendJSONMessage(t, sender, map[string]string{
		"type":     "typing",
		"content":  "abc",
		"pid":      recipientSetup.Pid,
		"username": "SpoofedUser",
	})

	outbound := readOutboundMessage(t, recipient)
	expected := senderSetup.Username + " is typing..."
	if outbound.Type != "typing" {
		t.Fatalf("expected typing type, got %q", outbound.Type)
	}
	if outbound.Content != expected {
		t.Fatalf("expected typing content %q, got %q", expected, outbound.Content)
	}
	if outbound.Username != senderSetup.Username {
		t.Fatalf("expected username %q, got %q", senderSetup.Username, outbound.Username)
	}
}

func TestTypingEventsAreRateLimited(t *testing.T) {
	url := startTestServer(t, config.Config{})
	sender := dialWebsocket(t, url, testOrigin(url))
	recipient := dialWebsocket(t, url, testOrigin(url))

	_ = readSetupMessage(t, sender)
	_ = readSetupMessage(t, recipient)

	sendJSONMessage(t, sender, map[string]string{"type": "typing", "content": "h"})
	sendJSONMessage(t, sender, map[string]string{"type": "typing", "content": "he"})
	_ = readOutboundMessage(t, recipient)

	if err := expectNoClientMessage(recipient, 250*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestTypingIgnoresEmptyContent(t *testing.T) {
	url := startTestServer(t, config.Config{})
	sender := dialWebsocket(t, url, testOrigin(url))
	recipient := dialWebsocket(t, url, testOrigin(url))

	_ = readSetupMessage(t, sender)
	_ = readSetupMessage(t, recipient)

	sendJSONMessage(t, sender, map[string]string{"type": "typing", "content": "   "})
	if err := expectNoClientMessage(recipient, 250*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestDisallowedOriginRejected(t *testing.T) {
	url := startTestServer(t, config.Config{
		AllowedOrigins: []string{"http://allowed.example"},
	})

	headers := http.Header{"Origin": []string{"http://blocked.example"}}
	connection, response, err := clientwebsocket.DefaultDialer.Dial(url, headers)
	if connection != nil {
		_ = connection.Close()
	}
	if err == nil {
		t.Fatal("expected bad handshake for disallowed origin")
	}
	if response == nil || response.StatusCode != http.StatusUpgradeRequired {
		t.Fatalf("expected upgrade required, got %#v", response)
	}
}

func TestOversizedMessageIsRejected(t *testing.T) {
	url := startTestServer(t, config.Config{MaxMessageSize: 64})
	sender := dialWebsocket(t, url, testOrigin(url))
	recipient := dialWebsocket(t, url, testOrigin(url))

	_ = readSetupMessage(t, sender)
	_ = readSetupMessage(t, recipient)

	sendJSONMessage(t, sender, map[string]string{
		"type":    "message",
		"content": strings.Repeat("x", 256),
	})

	if err := expectNoClientMessage(recipient, 300*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func startTestServer(t *testing.T, cfg config.Config) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	if len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = []string{"http://" + listener.Addr().String()}
	}
	if cfg.Port == "" {
		cfg.Port = config.DefaultPort
	}
	if cfg.MaxMessageSize == 0 {
		cfg.MaxMessageSize = config.DefaultMaxMessageSize
	}
	if cfg.TypingMinInterval == 0 {
		cfg.TypingMinInterval = config.DefaultTypingMinInterval
	}
	if cfg.MessageMinInterval == 0 {
		cfg.MessageMinInterval = config.DefaultMessageMinInterval
	}

	app := New(cfg)
	go func() {
		_ = app.ListenOn(listener)
	}()

	t.Cleanup(func() {
		_ = app.Shutdown()
		_ = listener.Close()
	})

	return "ws://" + listener.Addr().String() + "/ws"
}

func dialWebsocket(t *testing.T, url string, origin string) *clientwebsocket.Conn {
	t.Helper()

	headers := http.Header{"Origin": []string{origin}}
	connection, response, err := clientwebsocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected websocket upgrade, got %#v", response)
	}

	t.Cleanup(func() {
		_ = connection.Close()
	})

	return connection
}

func readSetupMessage(t *testing.T, connection *clientwebsocket.Conn) message.Setup {
	t.Helper()

	_, payload, err := connection.ReadMessage()
	if err != nil {
		t.Fatalf("read setup message: %v", err)
	}

	setup := message.Setup{}
	if err := json.Unmarshal(payload, &setup); err != nil {
		t.Fatalf("unmarshal setup message: %v", err)
	}

	return setup
}

func readOutboundMessage(t *testing.T, connection *clientwebsocket.Conn) message.Outbound {
	t.Helper()

	_, payload, err := connection.ReadMessage()
	if err != nil {
		t.Fatalf("read outbound message: %v", err)
	}

	outbound := message.Outbound{}
	if err := json.Unmarshal(payload, &outbound); err != nil {
		t.Fatalf("unmarshal outbound message: %v", err)
	}

	return outbound
}

func sendJSONMessage(t *testing.T, connection *clientwebsocket.Conn, payload map[string]string) {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal websocket payload: %v", err)
	}

	if err := connection.WriteMessage(clientwebsocket.TextMessage, body); err != nil {
		t.Fatalf("write websocket payload: %v", err)
	}
}

func testOrigin(url string) string {
	host := strings.TrimPrefix(url, "ws://")
	host = strings.TrimSuffix(host, "/ws")
	return "http://" + host
}

func expectNoClientMessage(connection *clientwebsocket.Conn, timeout time.Duration) error {
	_ = connection.SetReadDeadline(time.Now().Add(timeout))
	defer func() {
		_ = connection.SetReadDeadline(time.Time{})
	}()

	_, _, err := connection.ReadMessage()
	if err == nil {
		return errors.New("unexpected websocket message received")
	}

	netError, ok := err.(interface{ Timeout() bool })
	if ok && netError.Timeout() {
		return nil
	}

	if clientwebsocket.IsCloseError(err, clientwebsocket.CloseNormalClosure, clientwebsocket.CloseGoingAway) {
		return nil
	}

	return nil
}
