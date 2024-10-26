package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"gophkeeper.com/internal/server/storage"
)

// AuthenticationService handles user authentication.
type AuthenticationService interface {
	GetToken(username string) (string, error)
	Authenticate(ctx context.Context, username, password string) (string, error)
	ValidateToken(token string) (string, error)
}

type JWTAuthService struct {
	userRepo *storage.UserRepo
	jwtKey   []byte
	tokenTTL time.Duration
}

func NewJWTAuthService(userRepo *storage.UserRepo, jwtKey []byte, tokenTTL time.Duration) *JWTAuthService {
	return &JWTAuthService{
		userRepo: userRepo,
		jwtKey:   jwtKey,
		tokenTTL: tokenTTL,
	}
}

func (s *JWTAuthService) GetToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(s.tokenTTL).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *JWTAuthService) Authenticate(ctx context.Context, username, password string) (string, error) {
	hashedPassword, err := s.userRepo.GetPasswordHash(ctx, username)
	if err != nil {
		return "", errors.New("user not found")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(s.tokenTTL).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *JWTAuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, exists := claims["username"].(string)
		if !exists {
			return "", errors.New("invalid token claims")
		}
		return username, nil
	}

	return "", errors.New("invalid token")
}
