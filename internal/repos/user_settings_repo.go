package repos

import (
	"context"
	"errors"

	"github.com/theverysameliquidsnake/sales-bot/internal/configs"
	"github.com/theverysameliquidsnake/sales-bot/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetUserSettings() ([]models.UserSettings, error) {
	cursor, err := getUserSettingsCollection().Find(context.Background(), bson.D{{}})
	if err != nil {
		return nil, errors.Join(errors.New("repository: could not query user settings:"), err)
	}
	defer cursor.Close(context.Background())

	var results []models.UserSettings
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, errors.Join(errors.New("repository: could not map user settings:"), err)
	}

	return results, nil
}

func UpsertBackloggdProfileSetting(userId int64, backloggdProfileUrl string) error {
	filter := bson.D{{Key: "user_id", Value: userId}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "backloggd_profile", Value: backloggdProfileUrl}}}}
	opts := options.UpdateOne().SetUpsert(true)

	if _, err := getUserSettingsCollection().UpdateOne(context.Background(), filter, update, opts); err != nil {
		return errors.Join(errors.New("repository: could not insert or update user backloggd profile url:"), err)
	}

	return nil
}

func UpsertCountrySetting(userId int64, countryCode string, currencyCode string) error {
	filter := bson.D{{Key: "user_id", Value: userId}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "country_code", Value: countryCode}, {Key: "currency_code", Value: currencyCode}}}}
	opts := options.UpdateOne().SetUpsert(true)

	if _, err := getUserSettingsCollection().UpdateOne(context.Background(), filter, update, opts); err != nil {
		return errors.Join(errors.New("repository: could not insert or update user country and/or currency:"), err)
	}

	return nil
}

func DeleteUserSettings(userId int64) error {
	filter := bson.D{{Key: "user_id", Value: userId}}

	if _, err := getUserSettingsCollection().DeleteOne(context.Background(), filter); err != nil {
		return errors.Join(errors.New("repository: could not delete user settings:"), err)
	}

	return nil
}

func getUserSettingsCollection() *mongo.Collection {
	return configs.GetMongoDatabase().Collection("user_settings")
}
