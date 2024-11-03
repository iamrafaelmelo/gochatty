package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Client struct {
	pid       string
	username  string
	isClosing bool
	mutex     sync.Mutex
}

type SetupMessage struct {
	Type     string `json:"type"`
	Pid      string `json:"pid"`
	Username string `json:"username"`
}

type DefaultMessage struct {
	Type     string `json:"type"`
	Pid      string `json:"pid"`
	Username string `json:"username"`
	Content  string `json:"content"`
	DateTime string `json:"datetime"`
}

var clients = make(map[*websocket.Conn]*Client)
var register = make(chan *websocket.Conn)
var broadcast = make(chan string)
var unregister = make(chan *websocket.Conn)

func main() {
	app := fiber.New()
	app.Static("/", "./web/public", fiber.Static{
		Index: "index.html",
	})

	app.Use(func(context *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(context) {
			return context.Next()
		}

		return context.SendStatus(fiber.StatusUpgradeRequired)
	})

	go handleMessages()

	app.Get("/ws", websocket.New(func(connection *websocket.Conn) {
		defer func() {
			unregister <- connection
			connection.Close()
		}()

		register <- connection

		for {
			messageType, message, error := connection.ReadMessage()

			if error != nil {
				if websocket.IsUnexpectedCloseError(error, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("[error] read error:", error)
				}

				return
			}

			if messageType != websocket.TextMessage {
				return
			}

			log.Println("[info] received message: ", string(message))
			broadcast <- string(message)
		}
	}))

	if error := app.Listen(":8080"); error != nil {
		log.Fatal(error)
	}
}

func handleMessages() {
	for {
		select {
		case connection := <-register:
			registerConnection(connection)
			log.Println("[info] connection registered")
		case message := <-broadcast:
			log.Println("[info] received message: ", message)
			broadcastMessage(message)
		case connection := <-unregister:
			removeConnection(connection)
			log.Println("[info] connection unregistered")
		}
	}
}

func registerConnection(connection *websocket.Conn) {
	pid := uuid.NewString()
	username := fmt.Sprintf("Anonymous%d", rand.Int())

	clients[connection] = &Client{
		pid:      pid,
		username: username,
	}

	message := &SetupMessage{
		Type:     "setup",
		Pid:      pid,
		Username: username,
	}

	payload, error := json.Marshal(message)

	if error != nil {
		log.Println("[error] converting to json error: ", error)
	}

	connection.WriteMessage(websocket.TextMessage, payload)
}

func broadcastMessage(content string) {
	for connection, client := range clients {
		go func(connection *websocket.Conn, client *Client) {
			client.mutex.Lock()
			defer client.mutex.Unlock()

			if client.isClosing {
				return
			}

			message := &DefaultMessage{}

			if error := json.Unmarshal([]byte(content), message); error != nil {
				log.Println("[error] converting json to struct error: ", error)
			}

			if message.Pid == client.pid {
				return
			}

			payload, error := json.Marshal(message)

			if error != nil {
				log.Println("[error] converting to json error: ", error)
			}

			if error := connection.WriteMessage(websocket.TextMessage, payload); error != nil {
				log.Println("[error] write error: ", error)

				client.isClosing = true
				connection.WriteMessage(websocket.CloseMessage, []byte{})
				connection.Close()
				unregister <- connection
			}
		}(connection, client)
	}
}

func removeConnection(connection *websocket.Conn) {
	delete(clients, connection)
}
