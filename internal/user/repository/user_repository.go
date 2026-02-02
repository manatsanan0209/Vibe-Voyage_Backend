package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type pgUserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) domain.UserRepository {
	return &pgUserRepository{pool: pool}
}

func (r *pgUserRepository) Create(ctx context.Context, user *domain.User) error {
	sql := `INSERT INTO users (email, name) VALUES ($1, $2) RETURNING id`
	return r.pool.QueryRow(ctx, sql, user.Email, user.Name).Scan(&user.ID)
}

func (r *pgUserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	// ... logic query ...
	return &domain.User{ID: id, Name: "Mock User"}, nil // ตัวอย่าง
}