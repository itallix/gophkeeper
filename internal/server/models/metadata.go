package models

import (
	"time"
)

type Visitor interface {
	VisitCard(card *Card) error
	VisitLogin(login *Login) error
	VisitNote(note *Note) error
	VisitBinary(binary *Binary) error
}

// Secret interface to implement double dispatch with algorithm decoupling.
type Secret interface {
	Accept(visitor Visitor) error
}

// SecretMetadata represents secret metadata to be saved in DB.
type SecretMetadata struct {
	SecretID         int64
	Path             string
	CustomMeta       map[string]string
	CreatedAt        time.Time
	ModifiedAt       time.Time
	EncryptedDataKey []byte
	CreatedBy        string
	ModifiedBy       string
}

type Login struct {
	LoginID  int64
	Login    string
	Password []byte

	SecretMetadata
}

func (login *Login) Accept(v Visitor) error {
	return v.VisitLogin(login)
}

type Card struct {
	CardID         int64
	CardholderName string
	Number         []byte
	ExpiryMonth    int8
	ExpiryYear     int8
	Cvc            []byte

	SecretMetadata
}

func (card *Card) Accept(v Visitor) error {
	return v.VisitCard(card)
}

type Note struct {
}

func (note *Note) Accept(v Visitor) error {
	return v.VisitNote(note)
}

type Binary struct {
}

func (binary *Binary) Accept(v Visitor) error {
	return v.VisitBinary(binary)
}
