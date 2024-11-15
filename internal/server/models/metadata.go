package models

import (
	"time"
)

type SecretVisitor interface {
	VisitCard(card *Card) error
	VisitLogin(login *Login) error
	VisitNote(note *Note) error
	VisitBinary(binary *Binary) error
	GetResult() any
}

// Secret interface to implement double dispatch with algorithm decoupling.
type Secret interface {
	Accept(visitor SecretVisitor) error
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

func (login *Login) Accept(v SecretVisitor) error {
	return v.VisitLogin(login)
}

type Card struct {
	CardID         int64
	CardholderName string
	Number         []byte
	ExpiryMonth    int64
	ExpiryYear     int64
	CVC            []byte

	SecretMetadata
}

func (card *Card) Accept(v SecretVisitor) error {
	return v.VisitCard(card)
}

type Note struct {
	NoteID int64
	Text   []byte

	SecretMetadata
}

func (note *Note) Accept(v SecretVisitor) error {
	return v.VisitNote(note)
}

type Binary struct {
	BinaryID int64
	ChunkID  int64
	Chunks   int64
	Hash     string
	Data     []byte

	SecretMetadata
}

func (binary *Binary) Accept(v SecretVisitor) error {
	return v.VisitBinary(binary)
}

// IsLast indicates the final chunk which doesn't have any data, but contains full file hash.
func (binary *Binary) IsLast() bool {
	return binary.Data == nil && binary.Chunks > 0
}
