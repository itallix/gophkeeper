package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

// KMS defines the interface for key management operations.
// It provides methods for generating and decrypting data keys used in
// application-level encryption.
type KMS interface {
	GenerateDataKey() ([]byte, []byte, error)
	DecryptDataKey(encryptedDataKey []byte) ([]byte, error)
}

// RSAKMS implements the KMS interface using RSA-based key encryption.
// It uses a master RSA key to protect an AES encryption key, which in turn
// protects the data keys.
type RSAKMS struct {
	EncryptionKey []byte // AES key used for data key encryption/decryption
}

// NewRSAKMS creates a new instance of RSAKMS using the provided master key and encrypted key files.
//
// Parameters:
//   - masterKeyPath: Path to the PEM-encoded RSA private key file (PKCS8 format)
//   - encryptedKeyPath: Path to the file containing the encrypted AES key
//
// Returns:
//   - *RSAKMS: A new RSAKMS instance
//   - error: Any error encountered during initialization
//
// The master key should be in PKCS8 PEM format, and the encrypted key should have been
// encrypted using the corresponding RSA public key with OAEP padding.
func NewRSAKMS(masterKeyPath, encryptedKeyPath string) (*RSAKMS, error) {
	masterKeyPEM, err := os.ReadFile(masterKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read master key file: %w", err)
	}
	block, _ := pem.Decode(masterKeyPEM)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing RSA private key")
	}

	privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}
	masterKey, ok := privateKeyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("parsed key is not an RSA private key")
	}

	encryptedKeyBytes, err := os.ReadFile(encryptedKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption key file: %w", err)
	}
	encryptionKey, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		masterKey,
		encryptedKeyBytes,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt encryption key: %w", err)
	}

	return &RSAKMS{EncryptionKey: encryptionKey}, nil
}

const (
	DataKeyLength = 32
)

// GenerateDataKey implements the KMS interface. It generates a new random data key
// and encrypts it with the KMS's encryption key using AES-GCM.
//
// Returns:
//   - []byte: The plaintext data key (32 bytes)
//   - []byte: The encrypted data key (including GCM nonce)
//   - error: Any error encountered during generation or encryption
func (kms *RSAKMS) GenerateDataKey() ([]byte, []byte, error) {
	dataKey := make([]byte, DataKeyLength)
	_, err := rand.Read(dataKey)
	if err != nil {
		return nil, nil, err
	}

	encryptedDataKey, err := EncryptAES(
		dataKey,
		kms.EncryptionKey,
	)
	if err != nil {
		return nil, nil, err
	}

	return dataKey, encryptedDataKey, nil
}

// DecryptDataKey implements the KMS interface. It decrypts an encrypted data key
// using the KMS's encryption key with AES-GCM.
//
// Parameters:
//   - encryptedDataKey: The encrypted data key, including the GCM nonce
//
// Returns:
//   - []byte: The decrypted data key
//   - error: Any error encountered during decryption
func (kms *RSAKMS) DecryptDataKey(encryptedDataKey []byte) ([]byte, error) {
	return DecryptAES(encryptedDataKey, kms.EncryptionKey)
}

// EncryptAES encrypts plaintext using AES-GCM with the provided key.
// The nonce is prepended to the ciphertext in the returned byte slice.
//
// Parameters:
//   - plaintext: The data to encrypt
//   - key: The AES key to use for encryption
//
// Returns:
//   - []byte: The encrypted data (nonce + ciphertext)
//   - error: Any error encountered during encryption
func EncryptAES(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Use AES-GCM for authenticated encryption
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptAES decrypts ciphertext using AES-GCM with the provided key.
// Expects the nonce to be prepended to the ciphertext.
//
// Parameters:
//   - ciphertext: The encrypted data (nonce + ciphertext)
//   - key: The AES key to use for decryption
//
// Returns:
//   - []byte: The decrypted data
//   - error: Any error encountered during decryption
func DecryptAES(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Use AES-GCM for authenticated encryption
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
