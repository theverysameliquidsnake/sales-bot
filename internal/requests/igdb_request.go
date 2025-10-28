package requests

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/theverysameliquidsnake/sales-bot/internal/configs"
	"github.com/theverysameliquidsnake/sales-bot/internal/models/igdb"
)

func RequestGamesFromIgdb(slugs []string) ([]igdb.Game, error) {
	client := &http.Client{}
	payload := "fields *, external_games.*, external_games.external_game_source.*, cover.*; where slug = (\"" + strings.Join(slugs, "\", \"") + "\");"

	request, err := http.NewRequest("POST", "https://api.igdb.com/v4/games", strings.NewReader(payload))
	if err != nil {
		return nil, errors.Join(errors.New("request: could not create request to igdb:"), err)
	}

	request.Header.Set("Content-Type", "text/plain")
	for key, value := range configs.ConstructAdditionalHeadersForIgdb() {
		request.Header.Set(key, value)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, errors.Join(errors.New("request: could not do request to igdb:"), err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Join(errors.New("request: could not read response from igdb:"), err)
	}

	var games []igdb.Game
	err = json.Unmarshal([]byte(jsoniter.Get(body).ToString()), &games)
	if err != nil {
		return nil, errors.Join(errors.New("request: could not map response from igdb to variable:"), err)
	}

	return games, nil
}
