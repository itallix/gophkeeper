package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/itallix/gophkeeper/internal/server/storage"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidCreds     = errors.New("invalid credentials")
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidClaims    = errors.New("invalid token claims")
	ErrInvalidSignature = errors.New("unexpected signing method")
)

// TokenType represents the type of JWT token.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the custom JWT claims.
type Claims struct {
	Username string    `json:"username"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

// TokenPair represents an access and refresh token pair.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// AuthenticationService handles user authentication.
type AuthenticationService interface {
	GetTokenPair(username string) (*TokenPair, error)
	Authenticate(ctx context.Context, username, password string) (*TokenPair, error)
	ValidateAccessToken(accessToken string) (string, error)
	RefreshTokens(refreshToken string) (*TokenPair, error)
}

type JWTAuthService struct {
	userRepo        *storage.UserRepo
	accessTokenKey  []byte
	refreshTokenKey []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTAuthService(userRepo *storage.UserRepo, accessTokenKey []byte,
	refreshTokenKey []byte, accessTokenTTL time.Duration, refreshTokenTTL time.Duration) *JWTAuthService {
	return &JWTAuthService{
		userRepo:        userRepo,
		accessTokenKey:  accessTokenKey,
		refreshTokenKey: refreshTokenKey,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// createToken generates a new JWT token with the specified claims.
func (s *JWTAuthService) createToken(username string, tokenType TokenType, ttl time.Duration,
	key []byte) (string, error) {
	now := time.Now()
	claims := Claims{
		Username: username,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

func (s *JWTAuthService) createAccessToken(username string) (string, error) {
	return s.createToken(username, AccessToken, s.accessTokenTTL, s.accessTokenKey)
}

func (s *JWTAuthService) createRefreshToken(username string) (string, error) {
	return s.createToken(username, RefreshToken, s.refreshTokenTTL, s.refreshTokenKey)
}

func (s *JWTAuthService) GetTokenPair(username string) (*TokenPair, error) {
	accessToken, err := s.createAccessToken(username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.createRefreshToken(username)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *JWTAuthService) Authenticate(ctx context.Context, username, password string) (*TokenPair, error) {
	hashedPassword, err := s.userRepo.GetPasswordHash(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, ErrInvalidCreds
	}

	return s.GetTokenPair(username)
}

// parseAndValidateToken parses and validates a JWT token.
func (s *JWTAuthService) parseAndValidateToken(tokenString string, key []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return key, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func (s *JWTAuthService) ValidateAccessToken(accessToken string) (string, error) {
	claims, err := s.parseAndValidateToken(accessToken, s.accessTokenKey)
	if err != nil {
		return "", err
	}

	if claims.Type != AccessToken {
		return "", ErrInvalidToken
	}

	return claims.Username, nil
}

func (s *JWTAuthService) RefreshTokens(refreshToken string) (*TokenPair, error) {
	claims, err := s.parseAndValidateToken(refreshToken, s.refreshTokenKey)
	if err != nil {
		return nil, err
	}

	if claims.Type != RefreshToken {
		return nil, ErrInvalidToken
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, ErrTokenExpired
	}

	return s.GetTokenPair(claims.Username)
}
