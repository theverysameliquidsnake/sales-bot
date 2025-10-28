package models

type Wishlist struct {
	UserId   int64    `bson:"user_id"`
	SlugList []string `bson:"slug_list"`
}
