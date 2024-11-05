package jwt

import (
	"encoding/json"
	"fmt"
	"os"
)

type TokenProvider struct {
	filename string
}

func NewTokenProvider(filename string) *TokenProvider {
	return &TokenProvider{
		filename: filename,
	}
}

type TokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewToken(accessToken, refreshToken string) *TokenData {
	return &TokenData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

// Token storage with file permissions.
func (p *TokenProvider) SaveToken(tokenData *TokenData) error {
	jsonData, err := json.Marshal(tokenData)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Create file with user-only read/write permissions
	if err = os.WriteFile(p.filename, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

func (p *TokenProvider) LoadToken() (*TokenData, error) {
	data, err := os.ReadFile(p.filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var tokenData TokenData

	err = json.Unmarshal(data, &tokenData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &tokenData, nil
}
