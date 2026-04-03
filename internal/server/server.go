package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"github.com/iamrafaelmelo/simple-golang-chat/internal/client"
	"github.com/iamrafaelmelo/simple-golang-chat/internal/config"
	"github.com/iamrafaelmelo/simple-golang-chat/internal/message"
)

type registeredClient struct {
	connection *websocket.Conn
	client     *client.Client
}

type broadcastEnvelope struct {
	sender  *websocket.Conn
	message message.Outbound
}

type Server struct {
	app        *fiber.App
	config     config.Config
	clients    map[*websocket.Conn]*client.Client
	register   chan registeredClient
	unregister chan *websocket.Conn
	broadcast  chan broadcastEnvelope
	done       chan struct{}
	closeOnce  sync.Once
}

func New(cfg config.Config) *Server {
	server := &Server{
		app:        fiber.New(),
		config:     cfg,
		clients:    make(map[*websocket.Conn]*client.Client),
		register:   make(chan registeredClient),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan broadcastEnvelope),
		done:       make(chan struct{}),
	}

	server.registerRoutes()
	go server.handleMessages()

	return server
}

func (server *Server) Listen() error {
	return server.app.Listen(":" + server.config.Port)
}

func (server *Server) ListenOn(listener net.Listener) error {
	return server.app.Listener(listener)
}

func (server *Server) Shutdown() error {
	server.Close()
	return server.app.Shutdown()
}

func (server *Server) Close() {
	server.closeOnce.Do(func() {
		close(server.done)
	})
}

func (server *Server) registerRoutes() {
	server.app.Static("/", "./web/public", fiber.Static{Index: "index.html"})
	server.app.Get("/healthz", func(context *fiber.Ctx) error {
		return context.SendStatus(fiber.StatusOK)
	})
	server.app.Use("/ws", func(context *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(context) {
			if !server.isAllowedOrigin(context.Get("Origin")) {
				return context.SendStatus(fiber.StatusForbidden)
			}
			return context.Next()
		}

		return context.SendStatus(fiber.StatusUpgradeRequired)
	})
	server.app.Get("/ws", websocket.New(server.handleSocket, websocket.Config{
		Origins: server.config.AllowedOrigins,
	}))
}

func (server *Server) isAllowedOrigin(origin string) bool {
	if len(server.config.AllowedOrigins) == 0 {
		return false
	}

	for _, allowedOrigin := range server.config.AllowedOrigins {
		if allowedOrigin == origin {
			return true
		}
	}

	return false
}

func (server *Server) handleSocket(connection *websocket.Conn) {
	currentClient := client.New()

	if err := server.sendSetupMessage(connection, currentClient); err != nil {
		log.Println("[error] setup write error:", err)
		_ = connection.Close()
		return
	}

	server.register <- registeredClient{connection: connection, client: currentClient}
	connection.SetReadLimit(server.config.MaxMessageSize)

	defer func() {
		if currentClient.BeginClosing() {
			server.unregister <- connection
			_ = connection.Close()
		}
	}()

	for {
		messageType, payload, err := connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("[error] read error:", err)
			}
			return
		}

		if messageType != websocket.TextMessage {
			log.Println("[warn] ignoring non-text websocket message")
			continue
		}

		outbound, ok := server.buildOutboundMessage(currentClient, payload)
		if !ok {
			continue
		}

		server.broadcast <- broadcastEnvelope{sender: connection, message: outbound}
	}
}

func (server *Server) buildOutboundMessage(currentClient *client.Client, payload []byte) (message.Outbound, bool) {
	inbound := message.Inbound{}
	if err := json.Unmarshal(payload, &inbound); err != nil {
		log.Println("[warn] invalid json payload:", err)
		return message.Outbound{}, false
	}

	now := time.Now()

	switch inbound.Type {
	case "message":
		content := strings.TrimSpace(inbound.Content)
		if content == "" {
			return message.Outbound{}, false
		}
		if !currentClient.AllowEvent(now, server.config.MessageMinInterval, false) {
			return message.Outbound{}, false
		}
		return server.newOutboundMessage(currentClient, inbound.Type, content, now), true
	case "typing":
		content := strings.TrimSpace(inbound.Content)
		if content == "" {
			return message.Outbound{}, false
		}
		if !currentClient.AllowEvent(now, server.config.TypingMinInterval, true) {
			return message.Outbound{}, false
		}
		return server.newOutboundMessage(currentClient, inbound.Type, "", now), true
	default:
		log.Println("[warn] unsupported message type:", inbound.Type)
		return message.Outbound{}, false
	}
}

func (server *Server) newOutboundMessage(currentClient *client.Client, messageType string, content string, now time.Time) message.Outbound {
	return message.Outbound{
		Type:     messageType,
		Pid:      currentClient.PID(),
		Username: currentClient.Username(),
		Content:  content,
		DateTime: now.Format(time.TimeOnly),
	}
}

func (server *Server) sendSetupMessage(connection *websocket.Conn, currentClient *client.Client) error {
	setup := message.Setup{
		Type:     "setup",
		Pid:      currentClient.PID(),
		Username: currentClient.Username(),
	}

	payload, err := json.Marshal(setup)
	if err != nil {
		return err
	}

	var writeErr error
	currentClient.WithLock(func(isClosing bool) {
		if isClosing {
			writeErr = fmt.Errorf("client is closing")
			return
		}
		writeErr = connection.WriteMessage(websocket.TextMessage, payload)
	})

	return writeErr
}

func (server *Server) handleMessages() {
	for {
		select {
		case registration := <-server.register:
			server.clients[registration.connection] = registration.client
			log.Println("[info] connection registered")
			server.broadcastPresenceSnapshot()
		case event := <-server.broadcast:
			server.broadcastMessage(event)
		case connection := <-server.unregister:
			server.removeConnection(connection)
			server.broadcastPresenceSnapshot()
		case <-server.done:
			return
		}
	}
}

func (server *Server) broadcastMessage(event broadcastEnvelope) {
	for connection, connectedClient := range server.clients {
		if connection == event.sender {
			continue
		}

		payload := event.message
		if payload.Type == "typing" {
			payload.Content = fmt.Sprintf("%s is typing...", event.message.Username)
		}

		go server.writeMessage(connection, connectedClient, payload)
	}
}

func (server *Server) broadcastPresenceSnapshot() {
	users := make([]message.PresenceUser, 0, len(server.clients))
	for _, connectedClient := range server.clients {
		users = append(users, message.PresenceUser{
			Pid:      connectedClient.PID(),
			Username: connectedClient.Username(),
		})
	}

	sort.Slice(users, func(left int, right int) bool {
		if users[left].Username == users[right].Username {
			return users[left].Pid < users[right].Pid
		}
		return users[left].Username < users[right].Username
	})

	payload := message.Presence{
		Type:  "presence",
		Users: users,
	}

	for _, entry := range mapsKeys(server.clients) {
		go server.writePresence(entry.connection, entry.client, payload)
	}
}

func (server *Server) writeMessage(connection *websocket.Conn, currentClient *client.Client, outbound message.Outbound) {
	payload, err := json.Marshal(outbound)
	if err != nil {
		log.Println("[error] converting to json error:", err)
		return
	}

	var writeErr error
	currentClient.WithLock(func(isClosing bool) {
		if isClosing {
			return
		}
		writeErr = connection.WriteMessage(websocket.TextMessage, payload)
	})

	if writeErr != nil {
		log.Println("[error] write error:", writeErr)
		currentClient.MarkClosing()
		_ = connection.WriteMessage(websocket.CloseMessage, []byte{})
		_ = connection.Close()
		server.unregister <- connection
	}
}

func (server *Server) writePresence(connection *websocket.Conn, currentClient *client.Client, outbound message.Presence) {
	payload, err := json.Marshal(outbound)
	if err != nil {
		log.Println("[error] converting presence to json error:", err)
		return
	}

	var writeErr error
	currentClient.WithLock(func(isClosing bool) {
		if isClosing {
			return
		}
		writeErr = connection.WriteMessage(websocket.TextMessage, payload)
	})

	if writeErr != nil {
		log.Println("[error] presence write error:", writeErr)
		currentClient.MarkClosing()
		_ = connection.WriteMessage(websocket.CloseMessage, []byte{})
		_ = connection.Close()
		server.unregister <- connection
	}
}

func (server *Server) removeConnection(connection *websocket.Conn) {
	if _, ok := server.clients[connection]; !ok {
		return
	}

	delete(server.clients, connection)
	log.Println("[info] connection unregistered")
}

func mapsKeys(clients map[*websocket.Conn]*client.Client) []struct {
	connection *websocket.Conn
	client     *client.Client
} {
	keys := make([]struct {
		connection *websocket.Conn
		client     *client.Client
	}, 0, len(clients))

	for connection, currentClient := range clients {
		keys = append(keys, struct {
			connection *websocket.Conn
			client     *client.Client
		}{
			connection: connection,
			client:     currentClient,
		})
	}

	return keys
}
