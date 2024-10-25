package algo

import (
	"bytes"
	"fmt"

	"go.uber.org/zap/buffer"

	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/service"
)

type DecryptionVisitor struct {
	encryptionService service.EncryptionService
}

func NewDecryptionVisitor(service service.EncryptionService) *DecryptionVisitor {
	return &DecryptionVisitor{
		encryptionService: service,
	}
}

func (enc *DecryptionVisitor) VisitLogin(login *models.Login) error {
	var buf buffer.Buffer
	err := enc.encryptionService.DecryptStream(bytes.NewReader(login.Password), &buf, login.EncryptedDataKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt password: %w", err)
	}

	login.Password = buf.Bytes()

	return nil
}

func (enc *DecryptionVisitor) VisitCard(card *models.Card) error {
	var buf buffer.Buffer
	key := card.EncryptedDataKey

	err := enc.encryptionService.DecryptStream(bytes.NewReader(card.Number), &buf, key)
	if err != nil {
		return fmt.Errorf("cannot decrypt card number: %w", err)
	}
	card.Number = append([]byte(nil), buf.Bytes()...)
	buf.Reset()

	err = enc.encryptionService.DecryptStream(bytes.NewReader(card.Cvc), &buf, key)
	if err != nil {
		return fmt.Errorf("cannot decrypt cvc code: %w", err)
	}
	card.Cvc = buf.Bytes()

	return nil
}

func (enc *DecryptionVisitor) VisitNote(_ *models.Note) error {
	return nil
}

func (enc *DecryptionVisitor) VisitBinary(_ *models.Binary) error {
	return nil
}
