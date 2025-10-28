package igdb

type externalGameSource struct {
	Id   uint64 `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
}

type externalGame struct {
	Id                 uint64             `json:"id" bson:"id"`
	Uid                string             `json:"uid" bson:"uid"`
	ExternalGameSource externalGameSource `json:"external_game_source" bson:"external_game_source"`
}

type cover struct {
	Id      uint64 `json:"id" bson:"id"`
	ImageId string `json:"image_id" bson:"image_id"`
}

type Game struct {
	Id            uint64         `json:"id" bson:"id"`
	Name          string         `json:"name" bson:"name"`
	Cover         cover          `json:"cover" bson:"cover"`
	ExternalGames []externalGame `json:"external_games" bson:"external_games"`
	Slug          string         `json:"slug" bson:"slug"`
}
