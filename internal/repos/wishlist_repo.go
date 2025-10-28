package repos

import (
	"context"
	"errors"

	"github.com/theverysameliquidsnake/sales-bot/internal/configs"
	"github.com/theverysameliquidsnake/sales-bot/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetWishlists() ([]models.Wishlist, error) {
	cursor, err := getWishlistCollection().Find(context.Background(), bson.D{{}})
	if err != nil {
		return nil, errors.Join(errors.New("repository: could not query wishlists:"), err)
	}
	defer cursor.Close(context.Background())

	var results []models.Wishlist
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, errors.Join(errors.New("repository: could not map wishlists:"), err)
	}

	return results, nil
}

func InsertWishlists(wishlists []models.Wishlist) error {
	if _, err := getWishlistCollection().InsertMany(context.Background(), wishlists); err != nil {
		return errors.Join(errors.New("repository: could not insert wishlists:"), err)
	}

	return nil
}

func DropWishlists() error {
	if err := getWishlistCollection().Drop(context.Background()); err != nil {
		return errors.Join(errors.New("repository: could not delete wishlists collection:"), err)
	}

	return nil
}

func getWishlistCollection() *mongo.Collection {
	return configs.GetMongoDatabase().Collection("wishlists")
}
