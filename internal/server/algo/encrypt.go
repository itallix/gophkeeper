package algo

import (
	"bytes"
	"fmt"

	"go.uber.org/zap/buffer"

	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/service"
)

type EncryptionVisitor struct {
	encryptionService service.EncryptionService
}

func NewEncryptionVisitor(service service.EncryptionService) *EncryptionVisitor {
	return &EncryptionVisitor{
		encryptionService: service,
	}
}

func (enc *EncryptionVisitor) VisitLogin(login *models.Login) error {
	var buf buffer.Buffer
	_, encDataKey, err := enc.encryptionService.EncryptStream(bytes.NewReader(login.Password), &buf)
	if err != nil {
		return fmt.Errorf("cannot encrypt password: %w", err)
	}

	login.EncryptedDataKey = encDataKey
	login.Password = buf.Bytes()

	return nil
}

func (enc *EncryptionVisitor) VisitCard(card *models.Card) error {
	var buf buffer.Buffer

	dataKey, encDataKey, err := enc.encryptionService.EncryptStream(bytes.NewReader(card.Number), &buf)
	if err != nil {
		return fmt.Errorf("cannot encrypt card number: %w", err)
	}
	card.EncryptedDataKey = encDataKey
	card.Number = append([]byte(nil), buf.Bytes()...)
	buf.Reset()

	err = enc.encryptionService.EncryptStreamWithKey(bytes.NewReader(card.Cvc), &buf, dataKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt cvc code: %w", err)
	}
	card.Cvc = buf.Bytes()

	return nil
}

func (enc *EncryptionVisitor) VisitNote(_ *models.Note) error {
	return nil
}

func (enc *EncryptionVisitor) VisitBinary(_ *models.Binary) error {
	return nil
}
