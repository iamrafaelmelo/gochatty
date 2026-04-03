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
	presence := readPresenceMessage(t, connection)

	if setup.Type != "setup" {
		t.Fatalf("expected setup type, got %q", setup.Type)
	}
	if setup.Pid == "" {
		t.Fatal("expected server to assign a pid")
	}
	if !strings.HasPrefix(setup.Username, "Anonymous") {
		t.Fatalf("expected anonymous username, got %q", setup.Username)
	}
	if len(presence.Users) != 1 || presence.Users[0].Pid != setup.Pid {
		t.Fatalf("expected presence to include only the connected user, got %#v", presence.Users)
	}
}

func TestPresenceBroadcastIncludesConnectedUsers(t *testing.T) {
	url := startTestServer(t, config.Config{})

	firstConnection := dialWebsocket(t, url, testOrigin(url))
	secondConnection := dialWebsocket(t, url, testOrigin(url))

	firstSetup := readSetupMessage(t, firstConnection)
	_ = readPresenceMessage(t, firstConnection)

	secondSetup := readSetupMessage(t, secondConnection)
	secondPresence := readPresenceMessage(t, secondConnection)
	firstUpdatedPresence := readPresenceMessage(t, firstConnection)

	assertPresenceMembers(t, secondPresence.Users, firstSetup.Pid, secondSetup.Pid)
	assertPresenceMembers(t, firstUpdatedPresence.Users, firstSetup.Pid, secondSetup.Pid)
}

func TestPresenceBroadcastRemovesDisconnectedUsers(t *testing.T) {
	url := startTestServer(t, config.Config{})

	remainingConnection := dialWebsocket(t, url, testOrigin(url))
	departingConnection := dialWebsocket(t, url, testOrigin(url))

	remainingSetup := readSetupMessage(t, remainingConnection)
	_ = readPresenceMessage(t, remainingConnection)

	departingSetup := readSetupMessage(t, departingConnection)
	_ = readPresenceMessage(t, departingConnection)
	_ = readPresenceMessage(t, remainingConnection)

	if err := departingConnection.Close(); err != nil {
		t.Fatalf("close departing connection: %v", err)
	}

	updatedPresence := readPresenceMessage(t, remainingConnection)
	assertPresenceMembers(t, updatedPresence.Users, remainingSetup.Pid)

	for _, user := range updatedPresence.Users {
		if user.Pid == departingSetup.Pid {
			t.Fatalf("did not expect disconnected user %q to remain in presence list", departingSetup.Pid)
		}
	}
}

func TestMessageBroadcastUsesServerIdentity(t *testing.T) {
	url := startTestServer(t, config.Config{})
	sender := dialWebsocket(t, url, testOrigin(url))
	recipient := dialWebsocket(t, url, testOrigin(url))

	senderSetup := readSetupMessage(t, sender)
	_ = readPresenceMessage(t, sender)
	recipientSetup := readSetupMessage(t, recipient)
	_ = readPresenceMessage(t, recipient)
	_ = readPresenceMessage(t, sender)

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
	_ = readPresenceMessage(t, sender)
	_ = readSetupMessage(t, recipient)
	_ = readPresenceMessage(t, recipient)
	_ = readPresenceMessage(t, sender)

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
	_ = readPresenceMessage(t, sender)
	recipientSetup := readSetupMessage(t, recipient)
	_ = readPresenceMessage(t, recipient)
	_ = readPresenceMessage(t, sender)

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
	_ = readPresenceMessage(t, sender)
	_ = readSetupMessage(t, recipient)
	_ = readPresenceMessage(t, recipient)
	_ = readPresenceMessage(t, sender)

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
	_ = readPresenceMessage(t, sender)
	_ = readSetupMessage(t, recipient)
	_ = readPresenceMessage(t, recipient)
	_ = readPresenceMessage(t, sender)

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
	if response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %#v", response)
	}
}

func TestMissingAllowedOriginsRejectsUpgrade(t *testing.T) {
	url := startTestServer(t, config.Config{}, withoutDefaultAllowedOrigins())

	headers := http.Header{"Origin": []string{"http://localhost:8080"}}
	connection, response, err := clientwebsocket.DefaultDialer.Dial(url, headers)
	if connection != nil {
		_ = connection.Close()
	}
	if err == nil {
		t.Fatal("expected bad handshake when allowed origins are unset")
	}
	if response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %#v", response)
	}
}

func TestOversizedMessageIsRejected(t *testing.T) {
	url := startTestServer(t, config.Config{MaxMessageSize: 64})
	sender := dialWebsocket(t, url, testOrigin(url))
	recipient := dialWebsocket(t, url, testOrigin(url))

	senderSetup := readSetupMessage(t, sender)
	_ = readPresenceMessage(t, sender)
	recipientSetup := readSetupMessage(t, recipient)
	_ = readPresenceMessage(t, recipient)
	_ = readPresenceMessage(t, sender)

	sendJSONMessage(t, sender, map[string]string{
		"type":    "message",
		"content": strings.Repeat("x", 256),
	})

	nextMessage := readPresenceMessage(t, recipient)
	assertPresenceMembers(t, nextMessage.Users, recipientSetup.Pid)

	for _, user := range nextMessage.Users {
		if user.Pid == senderSetup.Pid {
			t.Fatalf("did not expect oversized message sender %q to remain connected", senderSetup.Pid)
		}
	}
}

type testServerOption func(*config.Config, net.Listener)

func startTestServer(t *testing.T, cfg config.Config, options ...testServerOption) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	applyDefaultAllowedOrigins := true
	for _, option := range options {
		option(&cfg, listener)
		if len(cfg.AllowedOrigins) == 0 {
			applyDefaultAllowedOrigins = false
		}
	}

	if applyDefaultAllowedOrigins && len(cfg.AllowedOrigins) == 0 {
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

func withoutDefaultAllowedOrigins() testServerOption {
	return func(cfg *config.Config, _ net.Listener) {
		cfg.AllowedOrigins = []string{}
	}
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

func readPresenceMessage(t *testing.T, connection *clientwebsocket.Conn) message.Presence {
	t.Helper()

	_, payload, err := connection.ReadMessage()
	if err != nil {
		t.Fatalf("read presence message: %v", err)
	}

	presence := message.Presence{}
	if err := json.Unmarshal(payload, &presence); err != nil {
		t.Fatalf("unmarshal presence message: %v", err)
	}

	return presence
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

func assertPresenceMembers(t *testing.T, users []message.PresenceUser, expectedPIDs ...string) {
	t.Helper()

	if len(users) != len(expectedPIDs) {
		t.Fatalf("expected %d presence users, got %d", len(expectedPIDs), len(users))
	}

	seen := make(map[string]bool, len(users))
	for _, user := range users {
		seen[user.Pid] = true
	}

	for _, expectedPID := range expectedPIDs {
		if !seen[expectedPID] {
			t.Fatalf("expected pid %q in presence users %#v", expectedPID, users)
		}
	}
}
