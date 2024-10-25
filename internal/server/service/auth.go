package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthenticationService handles user authentication.
type AuthenticationService interface {
	Authenticate(username, password string) (string, error)
	ValidateToken(token string) (bool, error)
}

type JWTAuthService struct {
	users    map[string]string // username -> hashed password
	jwtKey   []byte
	tokenTTL time.Duration
}

func NewJWTAuthService(jwtKey []byte, tokenTTL time.Duration) *JWTAuthService {
	return &JWTAuthService{
		users:    make(map[string]string),
		jwtKey:   jwtKey,
		tokenTTL: tokenTTL,
	}
}

func (s *JWTAuthService) RegisterUser(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s.users[username] = string(hashedPassword)
	return nil
}

func (s *JWTAuthService) Authenticate(username, password string) (string, error) {
	hashedPassword, ok := s.users[username]
	if !ok {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
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
