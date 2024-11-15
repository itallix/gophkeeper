package operation

import (
	"fmt"

	"go.uber.org/zap/buffer"

	"github.com/itallix/gophkeeper/internal/server/models"
	"github.com/itallix/gophkeeper/internal/server/service"
)

type Decryptor struct {
	encryptionService service.EncryptionService
}

func NewDecryptor(service service.EncryptionService) *Decryptor {
	return &Decryptor{
		encryptionService: service,
	}
}

func (enc *Decryptor) VisitLogin(login *models.Login) error {
	var buf buffer.Buffer
	err := enc.encryptionService.Decrypt(login.Password, &buf, login.EncryptedDataKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt password: %w", err)
	}

	login.Password = buf.Bytes()

	return nil
}

func (enc *Decryptor) VisitCard(card *models.Card) error {
	var buf buffer.Buffer
	key := card.EncryptedDataKey

	err := enc.encryptionService.Decrypt(card.Number, &buf, key)
	if err != nil {
		return fmt.Errorf("cannot decrypt card number: %w", err)
	}
	card.Number = append([]byte(nil), buf.Bytes()...)
	buf.Reset()

	err = enc.encryptionService.Decrypt(card.CVC, &buf, key)
	if err != nil {
		return fmt.Errorf("cannot decrypt cvc code: %w", err)
	}
	card.CVC = buf.Bytes()

	return nil
}

func (enc *Decryptor) VisitNote(note *models.Note) error {
	var buf buffer.Buffer
	err := enc.encryptionService.Decrypt(note.Text, &buf, note.EncryptedDataKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt note: %w", err)
	}

	note.Text = buf.Bytes()

	return nil
}

func (enc *Decryptor) VisitBinary(binary *models.Binary) error {
	if binary.IsLast() {
		return nil
	}

	var buf buffer.Buffer
	err := enc.encryptionService.Decrypt(binary.Data, &buf, binary.EncryptedDataKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt binary: %w", err)
	}
	binary.Data = buf.Bytes()

	return nil
}

func (enc *Decryptor) GetResult() any {
	return nil
}
