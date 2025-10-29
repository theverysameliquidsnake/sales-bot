package requests

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/theverysameliquidsnake/sales-bot/internal/models/steam"
)

func RequestAppDetailsFromSteam(appDetailsIds []uint64, countryCode string) ([]steam.AppDetails, error) {
	var appsDetails []steam.AppDetails
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	for _, appDetailId := range appDetailsIds {
		response, err := http.Get(fmt.Sprintf("https://store.steampowered.com/api/appdetails/?appids=%d&l=english&cc=%s", appDetailId, countryCode))
		if err != nil {
			return nil, errors.Join(errors.New("request: could not create request to steam:"), err)
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Join(errors.New("request: could not read response from steam:"), err)
		}

		if !jsoniter.Get(body, strconv.FormatUint(uint64(appDetailId), 10), "success").ToBool() {
			return nil, errors.New("reguest: could not confirm success from steam api")
		}

		var appDetails steam.AppDetails
		err = json.Unmarshal([]byte(jsoniter.Get(body, strconv.FormatUint(uint64(appDetailId), 10), "data").ToString()), &appDetails)
		if err != nil {
			return nil, errors.Join(errors.New("request: could not map response from steam to variable:"), err)
		}

		appsDetails = append(appsDetails, appDetails)

		// Wait to match rate limits
		time.Sleep(2 * time.Second)
	}

	return appsDetails, nil
}
