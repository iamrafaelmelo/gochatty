package main

import (
	"log"

	"github.com/iamrafaelmelo/simple-golang-chat/internal/config"
	"github.com/iamrafaelmelo/simple-golang-chat/internal/server"
)

func main() {
	app := server.New(config.Load())
	defer app.Close()

	if error := app.Listen(); error != nil {
		log.Fatal(error)
	}
}
