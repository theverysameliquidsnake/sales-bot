package main

import (
	"context"
	"log"
	"os"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/theverysameliquidsnake/sales-bot/internal/configs"
	"github.com/theverysameliquidsnake/sales-bot/internal/handlers"
)

func main() {
	// Postres
	if err := configs.ConnectToPostgres(); err != nil {
		log.Fatal(err)
	}
	defer configs.DisconnectFromPostgres()

	if err := configs.PingPostgres(); err != nil {
		log.Fatal(err)
	}

	// Valkey
	if err := configs.CreateValkeyClient(); err != nil {
		log.Fatal(err)
	}
	defer configs.DestroyValkeyClient()

	// Playwright
	if err := configs.InstallPlaywright(); err != nil {
		log.Fatal(err)
	}

	if err := configs.StartPlaywright(); err != nil {
		log.Fatal(err)
	}
	defer configs.StopPlaywright()

	// IGDB token
	if err := configs.RequestIgdbToken(); err != nil {
		log.Fatal(err)
	}

	// Run bot
	bot, err := telego.NewBot(os.Getenv("BOT_TOKEN"), telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatal(err)
	}

	updates, err := bot.UpdatesViaLongPolling(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	botHandler, err := telegohandler.NewBotHandler(bot, updates)
	if err != nil {
		log.Fatal(err)
	}
	defer botHandler.Stop()

	botHandler.Handle(handlers.StartHandler, telegohandler.CommandEqual("start"))
	botHandler.Handle(handlers.ProfileHandler, telegohandler.CommandEqual("profile"))
	botHandler.Handle(handlers.CountryHandler, telegohandler.CommandEqual("country"))
	botHandler.Handle(handlers.RefreshHandler, telegohandler.CommandEqual("refresh"))

	if err = botHandler.Start(); err != nil {
		log.Fatal(err)
	}
}
