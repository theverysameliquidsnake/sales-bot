package repos

import (
	"context"
	"errors"

	"github.com/theverysameliquidsnake/sales-bot/internal/configs"
	"github.com/theverysameliquidsnake/sales-bot/internal/models/steam"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetSteamAppsDetails(appIds []uint64) ([]steam.AppDetails, error) {
	filter := bson.M{"steam_appid": bson.M{"$in": appIds}}

	cursor, err := getSteamAppDetailsCollection().Find(context.Background(), filter)
	if err != nil {
		return nil, errors.Join(errors.New("repository: could not query steam apps details:"), err)
	}
	defer cursor.Close(context.Background())

	var results []steam.AppDetails
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, errors.Join(errors.New("repository: could not map steam apps details:"), err)
	}

	return results, nil
}

func InsertSteamAppsDetails(steamAppsDetails []steam.AppDetails) error {
	if _, err := getSteamAppDetailsCollection().InsertMany(context.Background(), steamAppsDetails); err != nil {
		return errors.Join(errors.New("repository: could not insert steam apps details:"), err)
	}

	return nil
}

func DropSteamAppsDetails() error {
	if err := getSteamAppDetailsCollection().Drop(context.Background()); err != nil {
		return errors.Join(errors.New("repository: could not delete steam apps details collection:"), err)
	}

	return nil
}

func getSteamAppDetailsCollection() *mongo.Collection {
	return configs.GetMongoDatabase().Collection("steam_app_details")
}
