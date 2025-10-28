package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/theverysameliquidsnake/sales-bot/internal/configs"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Make initial required connections and calls
	if err := configs.ConnectToMongo(); err != nil {
		log.Fatal(err)
	}
	defer configs.DisconnectFromMongo()

	if err := configs.RequestIgdbToken(); err != nil {
		log.Fatal(err)
	}
}
