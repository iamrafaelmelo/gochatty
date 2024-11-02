package main

import (
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type Client struct {
	isClosing bool
	mutex     sync.Mutex
}

var clients = make(map[*websocket.Conn]*Client)
var register = make(chan *websocket.Conn)
var broadcast = make(chan string)
var unregister = make(chan *websocket.Conn)

func handler() {
	for {
		select {
		case connection := <-register:
			clients[connection] = &Client{}
			log.Println("[info] connection registered")
		case message := <-broadcast:
			log.Println(message)

			for connection, client := range clients {
				go func(connection *websocket.Conn, client *Client) {
					client.mutex.Lock()
					defer client.mutex.Unlock()

					if client.isClosing {
						return
					}

					if error := connection.WriteMessage(websocket.TextMessage, []byte(message)); error != nil {
						log.Println("[error] write error: ", error)

						client.isClosing = true
						connection.WriteMessage(websocket.CloseMessage, []byte{})
						connection.Close()
						unregister <- connection
					}
				}(connection, client)
			}
		case connection := <-unregister:
			delete(clients, connection)
			log.Println("[info] connection unregistered")
		}
	}
}

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

	go handler()

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

			if messageType == websocket.TextMessage {
				log.Println("[info] received message: ", string(message))
				broadcast <- string(message)
			} else {
				log.Println("websocket message received of type", messageType)
			}
		}
	}))

	if error := app.Listen(":8080"); error != nil {
		log.Fatal(error)
	}
}
