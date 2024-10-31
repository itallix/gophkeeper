package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// EncryptionService handles encryption and decryption using data keys.
type EncryptionService interface {
	Encrypt(src []byte, dst io.Writer) ([]byte, []byte, error)
	EncryptWithKey(src []byte, dst io.Writer, encryptedDataKey []byte) error
	Decrypt(src []byte, dst io.Writer, encryptedDataKey []byte) error
}

type StandardEncryptionService struct {
	kms KMS
}

func NewStandardEncryptionService(kms KMS) *StandardEncryptionService {
	return &StandardEncryptionService{kms: kms}
}

func (s *StandardEncryptionService) Encrypt(src []byte, dst io.Writer) ([]byte, []byte, error) {
	// Generate a new data key for this encryption operation
	dataKey, encryptedDataKey, err := s.kms.GenerateDataKey()
	if err != nil {
		return nil, nil, err
	}

	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, src, nil)

	_, err = dst.Write(ciphertext)
	if err != nil {
		return nil, nil, err
	}

	// Return the data key and encrypted data key, which will be stored with the file metadata
	return dataKey, encryptedDataKey, nil
}

func (s *StandardEncryptionService) EncryptWithKey(src []byte, dst io.Writer, encryptedDataKey []byte) error {
	dataKey, err := s.kms.DecryptDataKey(encryptedDataKey)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, src, nil)

	_, err = dst.Write(ciphertext)
	if err != nil {
		return err
	}

	return nil
}

func (s *StandardEncryptionService) Decrypt(src []byte, dst io.Writer, encryptedDataKey []byte) error {
	// Decrypt the data key using the master key
	dataKey, err := s.kms.DecryptDataKey(encryptedDataKey)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := src[:nonceSize], src[nonceSize:]

	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}
	_, err = dst.Write(decrypted)
	if err != nil {
		return err
	}
	return nil
}
