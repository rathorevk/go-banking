package main

import (
	"log"

	"github.com/rathorevk/GoBanking/app"
)

func main() {
	log.Println("Initializing application...")

	// Start the server
	app.StartServer()
}
