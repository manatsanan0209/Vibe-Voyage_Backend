package service

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/token"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

const (
	tokenSecretKey     = "AUTH_TOKEN_SECRET"
	tokenTTLSecondsKey = "AUTH_TOKEN_TTL_SECONDS"
	defaultTTLSeconds  = 3600
)

type authService struct {
	userRepo domain.UserRepository
}

func NewAuthService(userRepo domain.UserRepository) domain.AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Register(ctx context.Context, user *domain.User) (*domain.AuthToken, error) {
	if user.Email == "" || user.Password == "" || user.Username == "" || user.FullName == "" {
		return nil, errors.New("the required fields are missing")
	}

	existUser, err := s.userRepo.GetByUsername(ctx, user.Username)
	if err == nil && existUser != nil {
		return nil, errors.New("username already exists")
	}

	existEmail, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existEmail != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user.Password = string(hashedPassword)
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	secret, ttl, err := tokenConfig()
	if err != nil {
		return nil, err
	}

	tok, exp, err := token.Generate(user.UserID, ttl, secret)
	if err != nil {
		return nil, err
	}

	return &domain.AuthToken{Token: tok, ExpiresAt: exp}, nil
}

func (s *authService) Login(ctx context.Context, username, password string) (*domain.User, *domain.AuthToken, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, nil, errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, nil, errors.New("invalid username or password")
	}

	secret, ttl, err := tokenConfig()
	if err != nil {
		return nil, nil, err
	}

	tok, exp, err := token.Generate(user.UserID, ttl, secret)
	if err != nil {
		return nil, nil, err
	}

	return user, &domain.AuthToken{Token: tok, ExpiresAt: exp}, nil
}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	_ = ctx
	secret, _, err := tokenConfig()
	if err != nil {
		return nil, err
	}

	claims, err := token.Validate(tokenString, secret)
	if err != nil {
		return nil, err
	}

	return &domain.TokenClaims{
		UserID:    claims.UserID,
		ExpiresAt: time.Unix(claims.Exp, 0),
	}, nil
}

func tokenConfig() (string, time.Duration, error) {
	secret := os.Getenv(tokenSecretKey)
	if secret == "" {
		return "", 0, errors.New("AUTH_TOKEN_SECRET is not set")
	}

	ttlSeconds := defaultTTLSeconds
	if v := os.Getenv(tokenTTLSecondsKey); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return "", 0, errors.New("AUTH_TOKEN_TTL_SECONDS must be integer")
		}
		ttlSeconds = parsed
	}

	return secret, time.Duration(ttlSeconds) * time.Second, nil
}
