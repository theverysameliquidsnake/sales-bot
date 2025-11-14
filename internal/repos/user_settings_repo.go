package repos

import (
	"context"
	"fmt"

	"github.com/theverysameliquidsnake/sales-bot/internal/configs"
	"github.com/theverysameliquidsnake/sales-bot/internal/models"
)

func FindUserSettings(userId int64) (*models.UserSettings, error) {
	pool := configs.GetPostgresPool()
	sql := "SELECT backloggd_profile, country_code FROM user_settings WHERE user_id = $1"

	var userSettings models.UserSettings
	err := pool.QueryRow(context.Background(), sql, userId).Scan(&userSettings.BackloggdProfile, &userSettings.CountryCode)
	if err != nil {
		return nil, fmt.Errorf("repository: could not query user settings by user id: %w", err)
	}

	return &userSettings, nil
}

func GetUserSettings() ([]models.UserSettings, error) {
	pool := configs.GetPostgresPool()
	sql := "SELECT user_id, backloggd_profile, country_code FROM user_settings"

	rows, err := pool.Query(context.Background(), sql)
	if err != nil {
		return nil, fmt.Errorf("repository: could not query user settings: %w", err)
	}
	defer rows.Close()

	var usersSettings []models.UserSettings
	for rows.Next() {
		var userSettings models.UserSettings
		err := rows.Scan(&userSettings.UserId, &userSettings.BackloggdProfile, &userSettings.CountryCode)
		if err != nil {
			return nil, fmt.Errorf("repository: could not map user settings: %w", err)
		}

		usersSettings = append(usersSettings, userSettings)
	}

	return usersSettings, nil
}

func UpsertBackloggdProfileSetting(userId int64, backloggdProfileUrl string) error {
	pool := configs.GetPostgresPool()
	sql := "INSERT INTO user_settings (user_id, backloggd_profile) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET backloggd_profile = $2"

	_, err := pool.Exec(context.Background(), sql, userId, backloggdProfileUrl)
	if err != nil {
		return fmt.Errorf("repository: could not insert or update user backloggd profile url: %w", err)
	}

	return nil
}

func UpsertCountrySetting(userId int64, countryCode string, currencyCode string) error {
	pool := configs.GetPostgresPool()
	sql := "INSERT INTO user_settings (user_id, country_code) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET country_code = $2"

	_, err := pool.Exec(context.Background(), sql, userId, countryCode)
	if err != nil {
		return fmt.Errorf("repository: could not insert or update user country and/or currency: %w", err)
	}

	return nil
}

func DeleteUserSettings(userId int64) error {
	pool := configs.GetPostgresPool()
	sql := "DELETE FROM user_settings WHERE user_id = $1"

	_, err := pool.Exec(context.Background(), sql, userId)
	if err != nil {
		return fmt.Errorf("repository: could not delete user settings: %w", err)
	}

	return nil
}
