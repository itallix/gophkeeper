package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

// Key Management Service (KMS) interface.
type KMS interface {
	GenerateDataKey() ([]byte, []byte, error)
	DecryptDataKey(encryptedDataKey []byte) ([]byte, error)
}

// RSA-based KMS implementation.
type RSAKMS struct {
	masterKey *rsa.PrivateKey
}

const BitSize = 2048

func NewRSAKMS() (*RSAKMS, error) {
	// For demonstration purposes, generating a new one each time
	masterKey, err := rsa.GenerateKey(rand.Reader, BitSize)
	if err != nil {
		return nil, err
	}
	return &RSAKMS{masterKey: masterKey}, nil
}

const DataKeyLength = 32

// GenerateDataKey returns generated data key as is and it's encrypted with master key version.
func (kms *RSAKMS) GenerateDataKey() ([]byte, []byte, error) {
	dataKey := make([]byte, DataKeyLength)
	_, err := rand.Read(dataKey)
	if err != nil {
		return nil, nil, err
	}

	// Encrypt the data key with the master key
	encryptedDataKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		&kms.masterKey.PublicKey,
		dataKey,
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	return dataKey, encryptedDataKey, nil
}

// DecryptDataKey decrypts encrypted data key using master key.
func (kms *RSAKMS) DecryptDataKey(encryptedDataKey []byte) ([]byte, error) {
	return rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		kms.masterKey,
		encryptedDataKey,
		nil,
	)
}
