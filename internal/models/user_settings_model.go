package models

type UserSettings struct {
	UserId           int64  `bson:"user_id"`
	BackloggdProfile string `bson:"backloggd_profile"`
	CountryCode      string `bson:"country_code"`
	CurrencyCode     string `bson:"currency_code"`
}
