package handlers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/theverysameliquidsnake/sales-bot/internal/models"
	"github.com/theverysameliquidsnake/sales-bot/internal/models/igdb"
	"github.com/theverysameliquidsnake/sales-bot/internal/models/steam"
	"github.com/theverysameliquidsnake/sales-bot/internal/parsers"
	"github.com/theverysameliquidsnake/sales-bot/internal/repos"
	"github.com/theverysameliquidsnake/sales-bot/internal/requests"
	"github.com/theverysameliquidsnake/sales-bot/internal/types"
)

func RunScheduledNotifications(ctx *telegohandler.Context, update telego.Update) error {
	if err := runCleanup(); err != nil {
		return errors.Join(errors.New("handler: could not clean mongo db:"), err)
	}

	userSettings, err := repos.GetUserSettings()
	if err != nil {
		return errors.Join(errors.New("handler: could not get all user settings from mongo db:"), err)
	}

	for _, settings := range userSettings {
		sales, err := processWishlist(settings)
		if err != nil {
			return errors.Join(fmt.Errorf("handler: could not handle profile: %s", settings.BackloggdProfile), err)
		}

		if len(sales) > 0 {
			var fullMessage string
			for _, sale := range sales {
				message := fmt.Sprintf("<a href=\"%s\"><b>%s</b></a>\n%s %s <s>%s</s>\n", sale.Url, sale.Name, sale.FinalPrice, sale.Discount, sale.InitialPrice)
				fullMessage = fullMessage + message
			}

			if _, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
				ChatID:    telegoutil.ID(settings.UserId),
				ParseMode: "HTML",
				Text:      fullMessage,
			}); err != nil {
				return errors.Join(errors.New("handler: could not send message:"), err)
			}
		}
	}

	return nil
}

func getMissingSlugs(wishlistSlugs []string, existingGames []igdb.Game) []string {
	slugsSet := types.NewSet()
	for _, wishlistSlug := range wishlistSlugs {
		slugsSet.Add(wishlistSlug)
	}

	for _, existingGame := range existingGames {
		if slugsSet.Contains(existingGame.Slug) {
			slugsSet.Remove(existingGame.Slug)
		}
	}

	return slugsSet.Values()
}

func getMissingSteamAppsIds(parsedSteamAppsIds []uint64, existingSteamAppsDetails []steam.AppDetails) ([]uint64, error) {
	idsSet := types.NewSet()
	for _, parsedSteamAppId := range parsedSteamAppsIds {
		idsSet.Add(strconv.FormatUint(parsedSteamAppId, 10))
	}

	for _, existingSteamAppDetails := range existingSteamAppsDetails {
		if idsSet.Contains(strconv.FormatUint(existingSteamAppDetails.SteamAppId, 10)) {
			idsSet.Remove(strconv.FormatUint(existingSteamAppDetails.SteamAppId, 10))
		}
	}

	var idsToRequest []uint64
	for _, id := range idsSet.Values() {
		parsedId, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("could not parse steam app id from string back to uint: %s", id), err)
		}

		idsToRequest = append(idsToRequest, parsedId)
	}

	return idsToRequest, nil
}

func runCleanup() error {
	if err := repos.DropWishlists(); err != nil {
		return errors.Join(errors.New("could not drop wishlists:"), err)
	}

	if err := repos.DropIgdbGames(); err != nil {
		return errors.Join(errors.New("could not drop igdb games:"), err)
	}

	if err := repos.DropSteamAppsDetails(); err != nil {
		return errors.Join(errors.New("could not drop steam apps details:"), err)
	}

	return nil
}

func obtainIgdbGames(slugs []string) ([]igdb.Game, error) {
	existingGames, err := repos.GetIgdbGames(slugs)
	if err != nil {
		return nil, errors.Join(errors.New("could not check for existing igdb records:"), err)
	}

	// Add missing games if any
	slugsToRequest := getMissingSlugs(slugs, existingGames)
	if len(slugsToRequest) > 0 {
		games, err := requests.RequestGamesFromIgdb(slugsToRequest)
		if err != nil {
			return nil, errors.Join(errors.New("could not get games from igdb:"), err)
		}

		if err = repos.InsertIgdbGames(games); err != nil {
			return nil, errors.Join(errors.New("could not insert games from igdb:"), err)
		}
	}

	games, err := repos.GetIgdbGames(slugs)
	if err != nil {
		return nil, errors.Join(errors.New("could not get igdb games from mongo db:"), err)
	}

	return games, nil
}

func extractSteamAppsIdsFromExternalIgdbGames(igdbGames []igdb.Game) ([]uint64, error) {
	var steamAppsIds []uint64
	for _, igdbGame := range igdbGames {
		for _, externalGame := range igdbGame.ExternalGames {
			if externalGame.ExternalGameSource.Name == "Steam" {
				steamAppId, err := strconv.ParseUint(externalGame.Uid, 10, 64)
				if err != nil {
					return nil, errors.Join(fmt.Errorf("could not parse external game uid to uint: %s", externalGame.Uid), err)
				}

				steamAppsIds = append(steamAppsIds, steamAppId)
			}
		}
	}

	return steamAppsIds, nil
}

func obtainSteamAppsDetails(steamAppsIds []uint64, userSettings models.UserSettings) ([]steam.AppDetails, error) {
	existingSteamAppsDetails, err := repos.GetSteamAppsDetails(steamAppsIds)
	if err != nil {
		return nil, errors.Join(errors.New("could not check for existing steam record:"), err)
	}

	// Add missing apps details if any
	idsToRequest, err := getMissingSteamAppsIds(steamAppsIds, existingSteamAppsDetails)
	if err != nil {
		return nil, errors.Join(errors.New("could not get missing steam apps ids difference:"), err)
	}

	if len(idsToRequest) > 0 {
		appsDetails, err := requests.RequestAppDetailsFromSteam(idsToRequest, userSettings.CountryCode)
		if err != nil {
			return nil, errors.Join(errors.New("could not get apps details from steam:"), err)
		}

		if err = repos.InsertSteamAppsDetails(appsDetails); err != nil {
			return nil, errors.Join(errors.New("could not insert apps details from steam:"), err)
		}
	}

	appsDetails, err := repos.GetSteamAppsDetails(steamAppsIds)
	if err != nil {
		return nil, errors.Join(errors.New("could not get app details from mongo db:"), err)
	}

	return appsDetails, nil
}

func processWishlist(userSettings models.UserSettings) ([]models.Sale, error) {
	slugs, err := parsers.ParseBackloggdWishlist(userSettings.BackloggdProfile)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("could not parse profile: %s", userSettings.BackloggdProfile), err)
	}

	if len(slugs) == 0 {
		return []models.Sale{}, nil
	}

	wishlist := models.Wishlist{
		UserId:   userSettings.UserId,
		SlugList: slugs,
	}

	if err = repos.InsertWishlists([]models.Wishlist{wishlist}); err != nil {
		return nil, errors.Join(fmt.Errorf("could not insert wishlists: %s", userSettings.BackloggdProfile), err)
	}

	igdbGames, err := obtainIgdbGames(slugs)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("could not obtain igdb games: %s", userSettings.BackloggdProfile), err)
	}

	// Only Steam for now
	steamAppsIds, err := extractSteamAppsIdsFromExternalIgdbGames(igdbGames)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("could not extract steam apps ids from igdb games: %s", userSettings.BackloggdProfile), err)
	}

	if len(steamAppsIds) == 0 {
		return []models.Sale{}, nil
	}

	steamAppsDetails, err := obtainSteamAppsDetails(steamAppsIds, userSettings)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("could not obtain steam apps details: %s", userSettings.BackloggdProfile), err)
	}

	var steamSales []models.Sale
	for _, steamAppDetails := range steamAppsDetails {
		if steamAppDetails.PriceOverview.DiscountPercent > 0 {
			steamSale := models.Sale{
				Name:         steamAppDetails.Name,
				Url:          fmt.Sprintf("https://store.steampowered.com/app/%d/", steamAppDetails.SteamAppId),
				Discount:     fmt.Sprintf("-%d%%", steamAppDetails.PriceOverview.DiscountPercent),
				InitialPrice: steamAppDetails.PriceOverview.InitialFormatted,
				FinalPrice:   steamAppDetails.PriceOverview.FinalFormatted,
			}

			steamSales = append(steamSales, steamSale)
		}
	}

	return steamSales, nil
}
