package handlers

import (
	"errors"
	"fmt"

	"github.com/theverysameliquidsnake/sales-bot/internal/models"
	"github.com/theverysameliquidsnake/sales-bot/internal/models/igdb"
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

	// End if no user setting provided
	if len(userSettings) == 0 {
		return nil
	}

	var wishlists []models.Wishlist
	for _, setting := range userSettings {
		if isUserSettingsValid(setting) {
			slugs, err := parsers.ParseBackloggdWishlist(setting.BackloggdProfile)
			if err != nil {
				return errors.Join(fmt.Errorf("handler: could not parse profile: %s", setting.BackloggdProfile), err)
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

	// Get data from IGDB for games in wishlists
	for _, wishlist := range wishlists {
		existingGames, err := repos.GetIgdbGames(wishlist.SlugList)
		if err != nil {
			return errors.Join(fmt.Errorf("handler: could not check for existing records: %d", wishlist.UserId), err)
		}

		// Skip wishlist if no new games to request
		slugsToRequest := getMissingSlugs(wishlist.SlugList, existingGames)
		if len(slugsToRequest) == 0 {
			continue
		}

		games, err := requests.RequestGamesFromIgdb(slugsToRequest)
		if err != nil {
			return errors.Join(fmt.Errorf("handler: could not get games from igdb: %d", wishlist.UserId), err)
		}

		if err = repos.InsertIgdbGames(games); err != nil {
			return errors.Join(fmt.Errorf("handler: could not insert games from igdb: %d", wishlist.UserId), err)
		}
	}

	// Get data from Steam for linked IGDB games

	return nil
}

func isUserSettingsValid(userSettings models.UserSettings) bool {
	return len(userSettings.BackloggdProfile) > 0 && len(userSettings.CountryCode) > 0 && len(userSettings.CurrencyCode) > 0
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
