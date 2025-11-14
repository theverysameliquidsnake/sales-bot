package igdb

type externalGameSource struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

type externalGame struct {
	Id                 uint64             `json:"id"`
	Uid                string             `json:"uid"`
	ExternalGameSource externalGameSource `json:"external_game_source"`
}

type cover struct {
	Id      uint64 `json:"id"`
	ImageId string `json:"image_id"`
}

type Game struct {
	Id            uint64         `json:"id"`
	Name          string         `json:"name"`
	Cover         cover          `json:"cover"`
	ExternalGames []externalGame `json:"external_games"`
	Slug          string         `json:"slug"`
}
