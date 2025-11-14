package handlers

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/theverysameliquidsnake/sales-bot/internal/repos"
)

func StartHandler(ctx *telegohandler.Context, update telego.Update) error {
	message := "Not so fast, boss. First, send me a link to your Backloggd profile using the /profile <url> command, both short and long ones are acceptable. Then tell me the country you have your Steam account registered in using the /country <country code> command so I could display the right currency."
	if err := sendMessage(ctx, update, message); err != nil {
		errStack := fmt.Errorf("handler: could not handle /start command: %w", err)
		log.Println(errStack)
		return errStack
	}

	return nil
}

func ProfileHandler(ctx *telegohandler.Context, update telego.Update) error {
	if len(strings.Split(strings.TrimSpace(update.Message.Text), " ")) < 2 {
		message := "Boss, send me a link to your Backloggd profile using the /profile <url> command."
		if err := sendMessage(ctx, update, message); err != nil {
			errStack := fmt.Errorf("handler: could not handle /profile command: not enough arguments: %w", err)
			log.Println(errStack)
			return errStack
		}

		return nil
	}

	profileLink := strings.Split(update.Message.Text, " ")[1]
	if !isBackloggdLink(profileLink) {
		message := "Cannot confirm this is a Backloggd link. Try another one, boss."
		if err := sendMessage(ctx, update, message); err != nil {
			errStack := fmt.Errorf("handler: could not handle /profile command: not a Backloggd link: %w", err)
			log.Println(errStack)
			return errStack
		}

		return nil
	}

	if err := repos.UpsertBackloggdProfileSetting(update.Message.Chat.ID, profileLink); err != nil {
		message := "Couldn't update your profile link for some reason. Try again later, boss."
		if err := sendMessage(ctx, update, message); err != nil {
			errStack := fmt.Errorf("handler: could not handle /profile command: could not upsert link: %w", err)
			log.Println(errStack)
			return errStack
		}

		errStack := fmt.Errorf("handler: could not handle /profile command: could not upsert link: %w", err)
		log.Println(errStack)
		return errStack
	}

	message := "Got your link updated, boss."
	if err := sendMessage(ctx, update, message); err != nil {
		errStack := fmt.Errorf("handler: could not handle /profile command: send confirmation message: %w", err)
		log.Println(errStack)
		return errStack
	}

	return nil
}

func CountryHandler(ctx *telegohandler.Context, update telego.Update) error {
	if len(strings.Split(strings.TrimSpace(update.Message.Text), " ")) < 2 {
		message := "Boss, tell me your country you have Steam registered in using the /country <country code> command."
		if err := sendMessage(ctx, update, message); err != nil {
			return errors.Join(errors.New("handler: could not handle /country command: not enough arguments:"), err)
		}
	}

	countryCode := strings.Split(update.Message.Text, " ")[1]
	if !isCountryCode(countryCode) {
		message := "Cannot confirm this is a country code. Try another one, boss."
		if err := sendMessage(ctx, update, message); err != nil {
			return errors.Join(errors.New("handler: could not handle /country command: not a country code:"), err)
		}
	}

	if err := repos.UpsertCountrySetting(update.Message.Chat.ID, countryCode, ""); err != nil {
		message := "Couldn't update your country for some reason. Try again later, boss."
		if err := sendMessage(ctx, update, message); err != nil {
			return errors.Join(errors.New("handler: could not handle /country command: could not upsert country code:"), err)
		}
	}

	message := "Got your country updated, boss."
	if err := sendMessage(ctx, update, message); err != nil {
		return errors.Join(errors.New("handler: could not handle /country command: send confirmation message:"), err)
	}

	return nil
}

func RefreshHandler(ctx *telegohandler.Context, update telego.Update) error {
	if err := RunScheduledNotifications(ctx, update); err != nil {
		return errors.Join(errors.New("handler: could not handle /refresh command:"), err)
	}

	return nil
}

func sendMessage(ctx *telegohandler.Context, update telego.Update, message string) error {
	if _, err := ctx.Bot().SendMessage(ctx, telegoutil.Message(
		telegoutil.ID(update.Message.Chat.ID),
		message,
	)); err != nil {
		return errors.Join(errors.New("could not send message:"), err)
	}

	return nil
}

func isBackloggdLink(profileLink string) bool {
	return strings.HasPrefix(profileLink, "https://backloggd.com/u/") || strings.HasPrefix(profileLink, "https://bckl.gg/")
}

func isCountryCode(countryCode string) bool {
	return len(countryCode) == 2
}
