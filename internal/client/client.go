package client

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	pid             string
	username        string
	isClosing       bool
	lastTypingEvent time.Time
	lastMessageSent time.Time
	mutex           sync.Mutex
}

func New() *Client {
	return &Client{
		pid:      uuid.NewString(),
		username: fmt.Sprintf("Anonymous%d", rand.Int()),
	}
}

func (client *Client) PID() string {
	return client.pid
}

func (client *Client) Username() string {
	return client.username
}

func (client *Client) AllowEvent(now time.Time, minInterval time.Duration, typing bool) bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	var lastEvent *time.Time
	if typing {
		lastEvent = &client.lastTypingEvent
	} else {
		lastEvent = &client.lastMessageSent
	}

	if !lastEvent.IsZero() && now.Sub(*lastEvent) < minInterval {
		return false
	}

	*lastEvent = now

	return true
}

func (client *Client) BeginClosing() bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.isClosing {
		return false
	}

	client.isClosing = true

	return true
}

func (client *Client) WithLock(fn func(isClosing bool)) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	fn(client.isClosing)
}

func (client *Client) MarkClosing() {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.isClosing = true
}
