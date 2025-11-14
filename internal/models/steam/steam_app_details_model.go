package steam

type priceOverview struct {
	DiscountPercent  int    `json:"discount_percent"`
	Initial          int    `json:"initial"`
	InitialFormatted string `json:"initial_formatted"`
	Final            int    `json:"final"`
	FinalFormatted   string `json:"final_formatted"`
}

type AppDetails struct {
	Name          string        `json:"name"`
	SteamAppId    uint64        `json:"steam_appid"`
	PriceOverview priceOverview `json:"price_overview"`
}
