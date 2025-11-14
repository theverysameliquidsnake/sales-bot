package configs

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

func ConnectToPostgres() error {
	conn, err := pgxpool.New(context.Background(), os.Getenv("POSTGRES_URI"))
	if err != nil {
		return errors.Join(errors.New("config: cound not connect to postgres db:"), err)
	}

	pool = conn

	return nil
}

func PingPostgres() error {
	if err := pool.Ping(context.Background()); err != nil {
		return errors.Join(errors.New("config: cound not ping to postgres db:"), err)
	}

	return nil
}

func DisconnectFromPostgres() {
	pool.Close()
}

func GetPostgresPool() *pgxpool.Pool {
	return pool
}
