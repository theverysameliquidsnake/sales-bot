package steam

type priceOverview struct {
	DiscountPercent  int    `json:"discount_percent" bson:"discount_percent"`
	Initial          int    `json:"initial" bson:"initial"`
	InitialFormatted string `json:"initial_formatted" bson:"initial_formatted"`
	Final            int    `json:"final" bson:"final"`
	FinalFormatted   string `json:"final_formatted" bson:"final_formatted"`
}

type AppDetails struct {
	Name          string        `json:"name" bson:"name"`
	SteamAppId    uint64        `json:"steam_appid" bson:"steam_appid"`
	PriceOverview priceOverview `json:"price_overview" bson:"price_overview"`
}
