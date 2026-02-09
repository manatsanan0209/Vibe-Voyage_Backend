package repository

import "github.com/jackc/pgx/v5/pgxpool"

type pgAuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *pgAuthRepository {
	return &pgAuthRepository{pool: pool}
}