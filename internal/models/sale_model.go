package models

type Sale struct {
	Name         string `json:"name"`
	Url          string `json:"url"`
	Image        string `json:"image"`
	Discount     string `json:"discount"`
	InitialPrice string `json:"initial_price"`
	FinalPrice   string `json:"final_price"`
}
