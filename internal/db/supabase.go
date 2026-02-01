package db

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const SupabaseDatabaseURLKey = "SUPABASE_DB_URL"

func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	_ = godotenv.Load()

	databaseURL := os.Getenv(SupabaseDatabaseURLKey)
	if databaseURL == "" {
		return nil, errors.New("SUPABASE_DB_URL is not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, config)
	if err != nil {
		return nil, err
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
