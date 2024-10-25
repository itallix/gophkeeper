package operation

import (
	"bytes"
	"fmt"

	"go.uber.org/zap/buffer"

	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/service"
)

type Encryptor struct {
	encryptionService service.EncryptionService
}

func NewEncryptor(service service.EncryptionService) *Encryptor {
	return &Encryptor{
		encryptionService: service,
	}
}

func (enc *Encryptor) VisitLogin(login *models.Login) error {
	var buf buffer.Buffer
	_, encDataKey, err := enc.encryptionService.EncryptStream(bytes.NewReader(login.Password), &buf)
	if err != nil {
		return fmt.Errorf("cannot encrypt password: %w", err)
	}

	login.EncryptedDataKey = encDataKey
	login.Password = buf.Bytes()

	return nil
}

func (enc *Encryptor) VisitCard(card *models.Card) error {
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

func (enc *Encryptor) VisitNote(_ *models.Note) error {
	return nil
}

func (enc *Encryptor) VisitBinary(_ *models.Binary) error {
	return nil
}

func (enc *Encryptor) GetResult() any {
	return nil
}
