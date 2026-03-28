package client

import (
	"testing"
	"time"
)

func TestAllowEventRateLimitsPerEventType(t *testing.T) {
	currentClient := New()
	now := time.Now()

	if !currentClient.AllowEvent(now, time.Second, false) {
		t.Fatal("expected first message event to be allowed")
	}
	if currentClient.AllowEvent(now.Add(500*time.Millisecond), time.Second, false) {
		t.Fatal("expected message event inside interval to be rejected")
	}
	if !currentClient.AllowEvent(now.Add(500*time.Millisecond), time.Second, true) {
		t.Fatal("expected typing event to keep an independent rate limit")
	}
	if !currentClient.AllowEvent(now.Add(1100*time.Millisecond), time.Second, false) {
		t.Fatal("expected message event after interval to be allowed")
	}
}

func TestBeginClosingOnlySucceedsOnce(t *testing.T) {
	currentClient := New()

	if !currentClient.BeginClosing() {
		t.Fatal("expected first close transition to succeed")
	}
	if currentClient.BeginClosing() {
		t.Fatal("expected second close transition to be ignored")
	}
}

func TestMarkClosingUpdatesLockState(t *testing.T) {
	currentClient := New()

	currentClient.MarkClosing()

	currentClient.WithLock(func(isClosing bool) {
		if !isClosing {
			t.Fatal("expected client to be marked closing")
		}
	})
}
