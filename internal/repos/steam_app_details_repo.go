package repos

import (
	"encoding/json"
	"fmt"

	"github.com/theverysameliquidsnake/sales-bot/internal/configs"
	"github.com/theverysameliquidsnake/sales-bot/internal/models/steam"
)

func GetSteamAppsDetails(appIds []uint64, countryCode string) ([]steam.AppDetails, error) {
	var steamAppsDetails []steam.AppDetails
	for _, appId := range appIds {
		steamAppDetailsJson, err := configs.GetValkeyValue("steam:" + countryCode + ":" + fmt.Sprint(appId))
		if err != nil {
			return nil, fmt.Errorf("repository: could not query steam app details: %w", err)
		}

		if len(steamAppDetailsJson) == 0 {
			continue
		}

		var steamAppDetails steam.AppDetails
		err = json.Unmarshal([]byte(steamAppDetailsJson), &steamAppDetails)
		if err != nil {
			return nil, fmt.Errorf("repository: could not map steam app details: %w", err)
		}

		steamAppsDetails = append(steamAppsDetails, steamAppDetails)
	}

	return steamAppsDetails, nil
}

func InsertSteamAppsDetails(steamAppsDetails []steam.AppDetails, countryCode string) error {
	for _, steamAppDetails := range steamAppsDetails {
		steamAppDetailsJson, err := json.Marshal(steamAppDetails)
		if err != nil {
			return fmt.Errorf("repository: could not marshal steam app details: %w", err)
		}

		if err = configs.SetValkeyValue("steam:"+countryCode+":"+fmt.Sprint(steamAppDetails.SteamAppId), string(steamAppDetailsJson)); err != nil {
			return fmt.Errorf("repository: could not insert steam app details: %w", err)
		}
	}

	return nil
}
