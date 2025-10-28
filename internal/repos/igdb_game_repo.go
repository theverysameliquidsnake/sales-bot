package repos

import (
	"context"
	"errors"

	"github.com/theverysameliquidsnake/sales-bot/internal/configs"
	"github.com/theverysameliquidsnake/sales-bot/internal/models/igdb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetIgdbGames(slugs []string) ([]igdb.Game, error) {
	filter := bson.M{"slug": bson.M{"$in": slugs}}

	cursor, err := getIgdbGamesCollection().Find(context.Background(), filter)
	if err != nil {
		return nil, errors.Join(errors.New("repository: could not query igdb games:"), err)
	}
	defer cursor.Close(context.Background())

	var results []igdb.Game
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, errors.Join(errors.New("repository: could not map igdb games:"), err)
	}

	return results, nil
}

func InsertIgdbGames(games []igdb.Game) error {
	if _, err := getIgdbGamesCollection().InsertMany(context.Background(), games); err != nil {
		return errors.Join(errors.New("repository: could not insert igdb games:"), err)
	}

	return nil
}

func DropIgdbGames() error {
	if err := getIgdbGamesCollection().Drop(context.Background()); err != nil {
		return errors.Join(errors.New("repository: could not delete igdb games collection:"), err)
	}

	return nil
}

func getIgdbGamesCollection() *mongo.Collection {
	return configs.GetMongoDatabase().Collection("igdb_games")
}
