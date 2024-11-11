package operation

import (
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
	encDataKey, err := enc.encryptionService.Encrypt(login.Password, &buf)
	if err != nil {
		return fmt.Errorf("cannot encrypt password: %w", err)
	}

	login.EncryptedDataKey = encDataKey
	login.Password = buf.Bytes()

	return nil
}

func (enc *Encryptor) VisitCard(card *models.Card) error {
	var buf buffer.Buffer

	encDataKey, err := enc.encryptionService.Encrypt(card.Number, &buf)
	if err != nil {
		return fmt.Errorf("cannot encrypt card number: %w", err)
	}
	card.EncryptedDataKey = encDataKey
	card.Number = append([]byte(nil), buf.Bytes()...)
	buf.Reset()

	err = enc.encryptionService.EncryptWithKey(card.CVC, &buf, encDataKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt cvc code: %w", err)
	}
	card.CVC = buf.Bytes()

	return nil
}

func (enc *Encryptor) VisitNote(note *models.Note) error {
	var buf buffer.Buffer

	encDataKey, err := enc.encryptionService.Encrypt(note.Text, &buf)
	if err != nil {
		return fmt.Errorf("cannot encrypt note text: %w", err)
	}
	note.EncryptedDataKey = encDataKey
	note.Text = buf.Bytes()

	return nil
}

func (enc *Encryptor) VisitBinary(binary *models.Binary) error {
	if binary.IsLast() {
		return nil
	}

	var buf buffer.Buffer
	if binary.EncryptedDataKey != nil {
		err := enc.encryptionService.EncryptWithKey(binary.Data, &buf, binary.EncryptedDataKey)
		if err != nil {
			return fmt.Errorf("cannot encrypt binary data: %w", err)
		}
	} else {
		encDataKey, err := enc.encryptionService.Encrypt(binary.Data, &buf)
		if err != nil {
			return fmt.Errorf("cannot encrypt binary data: %w", err)
		}
		binary.EncryptedDataKey = encDataKey
	}
	binary.Data = buf.Bytes()

	return nil
}

func (enc *Encryptor) GetResult() any {
	return nil
}
