package handlers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/theverysameliquidsnake/sales-bot/internal/models"
	"github.com/theverysameliquidsnake/sales-bot/internal/models/igdb"
	"github.com/theverysameliquidsnake/sales-bot/internal/models/steam"
	"github.com/theverysameliquidsnake/sales-bot/internal/parsers"
	"github.com/theverysameliquidsnake/sales-bot/internal/repos"
	"github.com/theverysameliquidsnake/sales-bot/internal/requests"
	"github.com/theverysameliquidsnake/sales-bot/internal/types"
)

func RunScheduledNotifications() error {
	// Parse user profiles in database and create wishlists
	userSettings, err := repos.GetUserSettings()
	if err != nil {
		return errors.Join(errors.New("handler: could not get all user settings from mongo db:"), err)
	}

	userSettingsMap := convertUserSettingsToMap(userSettings)

	// End if no user setting provided
	if len(userSettingsMap) == 0 {
		return nil
	}

	var wishlists []models.Wishlist
	for _, setting := range userSettingsMap {
		if isUserSettingsValid(setting) {
			slugs, err := parsers.ParseBackloggdWishlist(setting.BackloggdProfile)
			if err != nil {
				return errors.Join(fmt.Errorf("handler: could not parse profile: %s", setting.BackloggdProfile), err)
			}

			if len(slugs) == 0 {
				continue
			}

			wishlist := models.Wishlist{
				UserId:   setting.UserId,
				SlugList: slugs,
			}
			wishlists = append(wishlists, wishlist)
		}
	}

	// End if no wishlists to insert
	if len(wishlists) == 0 {
		return nil
	}

	if err = repos.InsertWishlists(wishlists); err != nil {
		return errors.Join(errors.New("handler: could not insert wishlists:"), err)
	}

	for _, wishlist := range wishlists {
		// Get data from IGDB for games in wishlists
		existingGames, err := repos.GetIgdbGames(wishlist.SlugList)
		if err != nil {
			return errors.Join(fmt.Errorf("handler: could not check for existing igdb records: %d", wishlist.UserId), err)
		}

		// Add missing games if any
		slugsToRequest := getMissingSlugs(wishlist.SlugList, existingGames)
		if len(slugsToRequest) > 0 {
			games, err := requests.RequestGamesFromIgdb(slugsToRequest)
			if err != nil {
				return errors.Join(fmt.Errorf("handler: could not get games from igdb: %d", wishlist.UserId), err)
			}

			if err = repos.InsertIgdbGames(games); err != nil {
				return errors.Join(fmt.Errorf("handler: could not insert games from igdb: %d", wishlist.UserId), err)
			}
		}

		// Get data from Steam for linked IGDB games
		games, err := repos.GetIgdbGames(wishlist.SlugList)
		if err != nil {
			return errors.Join(fmt.Errorf("handler: could not get igdb games from mongo db: %d", wishlist.UserId), err)
		}

		var steamAppIds []uint64
		for _, game := range games {
			for _, externalGame := range game.ExternalGames {
				if externalGame.ExternalGameSource.Name == "Steam" {
					steamAppId, err := strconv.ParseUint(externalGame.Uid, 10, 64)
					if err != nil {
						return errors.Join(fmt.Errorf("handler: could not parse external game uid to uint: %s", externalGame.Uid), err)
					}

					steamAppIds = append(steamAppIds, steamAppId)
				}
			}
		}

		// Add missing apps if any
		if len(steamAppIds) > 0 {
			existingSteamAppsDetails, err := repos.GetSteamAppsDetails(steamAppIds)
			if err != nil {
				return errors.Join(errors.New("handler: could not check for existing steam record:"), err)
			}

			idsToRequest, err := getMissingSteamAppsIds(steamAppIds, existingSteamAppsDetails)
			if err != nil {
				return errors.Join(errors.New("handler: could not get missing steam apps ids difference:"), err)
			}

			if len(idsToRequest) > 0 {
				appsDetails, err := requests.RequestAppDetailsFromSteam(idsToRequest, userSettingsMap[wishlist.UserId].CountryCode)
				if err != nil {
					return errors.Join(errors.New("handler: could not get apps details from steam:"), err)
				}

				if err = repos.InsertSteamAppsDetails(appsDetails); err != nil {
					return errors.Join(errors.New("handler: could not insert apps details from steam:"), err)
				}
			}

			appsDetails, err := repos.GetSteamAppsDetails(steamAppIds)
			if err != nil {
				return errors.Join(fmt.Errorf("handler: could not get app details from mongo db: %d", wishlist.UserId), err)
			}

			// Currently only Steam
			var sales []models.Sale
			for _, appDetails := range appsDetails {
				if appDetails.PriceOverview.DiscountPercent > 0 {
					sale := models.Sale{
						Name:         appDetails.Name,
						Url:          fmt.Sprintf("https://store.steampowered.com/app/%d/", appDetails.SteamAppId),
						Discount:     fmt.Sprintf("-%d%%", appDetails.PriceOverview.DiscountPercent),
						InitialPrice: appDetails.PriceOverview.InitialFormatted,
						FinalPrice:   appDetails.PriceOverview.FinalFormatted,
					}

					sales = append(sales, sale)
				}
			}

			// Send notifications
			if len(sales) > 0 {
				fmt.Println(sales)
			}
		}
	}

	// Clean up
	if err = repos.DropWishlists(); err != nil {
		return errors.Join(errors.New("handler: could not drop wishlists:"), err)
	}

	if err = repos.DropIgdbGames(); err != nil {
		return errors.Join(errors.New("handler: could not drop igdb games:"), err)
	}

	if err = repos.DropSteamAppsDetails(); err != nil {
		return errors.Join(errors.New("handler: could not drop steam apps details:"), err)
	}

	return nil
}

func isUserSettingsValid(userSettings models.UserSettings) bool {
	return len(userSettings.BackloggdProfile) > 0 && len(userSettings.CountryCode) > 0 && len(userSettings.CurrencyCode) > 0
}

func convertUserSettingsToMap(userSettingsList []models.UserSettings) map[int64]models.UserSettings {
	userSettingsMap := make(map[int64]models.UserSettings)
	for _, userSettings := range userSettingsList {
		userSettingsMap[userSettings.UserId] = userSettings
	}

	return userSettingsMap
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
