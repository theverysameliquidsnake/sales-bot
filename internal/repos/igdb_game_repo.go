package repos

import (
	"encoding/json"
	"fmt"

	"github.com/theverysameliquidsnake/sales-bot/internal/configs"
	"github.com/theverysameliquidsnake/sales-bot/internal/models/igdb"
)

func GetIgdbGames(slugs []string) ([]igdb.Game, error) {
	var games []igdb.Game
	for _, slug := range slugs {
		gameJson, err := configs.GetValkeyValue("igdb:" + slug)
		if err != nil {
			return nil, fmt.Errorf("repository: could not query igdb game: %w", err)
		}

		if len(gameJson) == 0 {
			continue
		}

		var game igdb.Game
		err = json.Unmarshal([]byte(gameJson), &game)
		if err != nil {
			return nil, fmt.Errorf("repository: could not map igdb game: %w", err)
		}

		games = append(games, game)
	}

	return games, nil
}

func InsertIgdbGames(games []igdb.Game) error {
	for _, game := range games {
		gameJson, err := json.Marshal(game)
		if err != nil {
			return fmt.Errorf("repository: could not marshal igdb game: %w", err)
		}

		if err = configs.SetValkeyValue("igdb:"+game.Slug, string(gameJson)); err != nil {
			return fmt.Errorf("repository: could not insert igdb game: %w", err)
		}
	}

	return nil
}
