package operation

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"gophkeeper.com/internal/server/models"
)

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

const (
	MinLoginLen      = 3
	MinPasswordLen   = 6
	MinCardNumberLen = 13
	MaxCardNumberLen = 19
	MinCVCLen        = 3
	MaxCVCLen        = 4
)

func (v *Validator) VisitLogin(login *models.Login) error {
	var errs []error

	if len(login.Login) < MinLoginLen {
		errs = append(errs, fmt.Errorf("login should be at least %d characters", MinLoginLen))
	}
	if len(string(login.Password)) < MinPasswordLen {
		errs = append(errs, fmt.Errorf("password should be at least %d characters", MinPasswordLen))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func validateExpiry(month, year int64) error {
	now := time.Now()
	currentYear := int64(now.Year())
	currentMonth := int64(now.Month())

	// Basic range checks
	if month < 1 || month > 12 {
		return errors.New("invalid month")
	}

	if year < currentYear || year > currentYear+20 {
		return errors.New("invalid year")
	}

	// Check if card is expired
	if year == currentYear && month < currentMonth {
		return errors.New("card is expired")
	}

	return nil
}

func (v *Validator) VisitCard(card *models.Card) error {
	var errs []error

	// Card Number validation
	cardNumberLen := len(string(card.Number))
	if cardNumberLen < MinCardNumberLen || cardNumberLen > MaxCardNumberLen {
		errs = append(errs, fmt.Errorf("card number length should be between %d and %d", MinCardNumberLen, MaxCardNumberLen))
	}
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(string(card.Number)) {
		errs = append(errs, errors.New("card number must contain only digits"))
	}

	// Expiry validation
	if err := validateExpiry(card.ExpiryMonth, card.ExpiryYear); err != nil {
		errs = append(errs, err)
	}

	// CVC validation
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(string(card.CVC)) {
		errs = append(errs, errors.New("CVC must contain only digits"))
	}
	cvcLen := len(string(card.CVC))
	if cvcLen < MinCVCLen || cvcLen > MaxCVCLen {
		errs = append(errs, fmt.Errorf("CVC length should be between %d and %d", MinCVCLen, MaxCVCLen))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (v *Validator) VisitNote(_ *models.Note) error {
	return nil
}

func (v *Validator) VisitBinary(_ *models.Binary) error {
	return nil
}

func (v *Validator) GetResult() any {
	return nil
}
