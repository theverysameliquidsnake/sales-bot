package configs

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	jsoniter "github.com/json-iterator/go"
)

var igdbToken string

func RequestIgdbToken() error {
	response, err := http.Post(fmt.Sprintf("https://id.twitch.tv/oauth2/token?client_id=%s&client_secret=%s&grant_type=client_credentials", os.Getenv("IGDB_ID"), os.Getenv("IGDB_SECRET")), "text/plain", nil)
	if err != nil {
		return errors.Join(errors.New("config: could not make request to igdb to request token:"), err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.Join(errors.New("config: could not read response from igdb to request token:"), err)
	}

	igdbToken = jsoniter.Get(body, "access_token").ToString()

	return nil
}

func ConstructAdditionalHeadersForIgdb() map[string]string {
	headers := make(map[string]string)

	headers["Client-ID"] = os.Getenv("IGDB_ID")
	headers["Authorization"] = fmt.Sprintf("Bearer %s", igdbToken)

	return headers
}
