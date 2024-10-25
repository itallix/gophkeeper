package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// EncryptionService handles encryption and decryption of streams using data keys.
type EncryptionService interface {
	EncryptStream(src io.Reader, dst io.Writer) ([]byte, []byte, error)
	EncryptStreamWithKey(src io.Reader, dst io.Writer, encryptedDataKey []byte) error
	DecryptStream(src io.Reader, dst io.Writer, encryptedDataKey []byte) error
}

type StreamEncryptionService struct {
	kms KMS
}

func NewStreamEncryptionService(kms KMS) *StreamEncryptionService {
	return &StreamEncryptionService{kms: kms}
}

func (s *StreamEncryptionService) EncryptStream(src io.Reader, dst io.Writer) ([]byte, []byte, error) {
	// Generate a new data key for this encryption operation
	dataKey, encryptedDataKey, err := s.kms.GenerateDataKey()
	if err != nil {
		return nil, nil, err
	}

	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, err
	}

	if _, err = dst.Write(iv); err != nil {
		return nil, nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	writer := &cipher.StreamWriter{S: stream, W: dst}

	_, err = io.Copy(writer, src)
	if err != nil {
		return nil, nil, err
	}

	// Return the data key and encrypted data key, which will be stored with the file metadata
	return dataKey, encryptedDataKey, nil
}

func (s *StreamEncryptionService) EncryptStreamWithKey(src io.Reader, dst io.Writer, dataKey []byte) error {
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	if _, err = dst.Write(iv); err != nil {
		return err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	writer := &cipher.StreamWriter{S: stream, W: dst}

	_, err = io.Copy(writer, src)
	if err != nil {
		return err
	}

	return nil
}

func (s *StreamEncryptionService) DecryptStream(src io.Reader, dst io.Writer, encryptedDataKey []byte) error {
	// Decrypt the data key using the master key
	dataKey, err := s.kms.DecryptDataKey(encryptedDataKey)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err = io.ReadFull(src, iv); err != nil {
		return err
	}

	stream := cipher.NewCFBDecrypter(block, iv)
	reader := &cipher.StreamReader{S: stream, R: src}

	_, err = io.Copy(dst, reader)
	return err
}
