package auth

import (
	"context"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
)

type login interface {
	FindEncryptedPasswordByEmail(ctx context.Context, email domain.Email) (domain.Ciphertext, error)
	WhenPasswordMatch(ctx context.Context, match bool) error
}

type LoginDto struct {
	Email    string
	Password string
}

func (d LoginDto) Validate() error {
	if err := domain.Email(d.Email).Validate(); err != nil {
		return err
	}

	if err := domain.Plaintext(d.Password).Validate(); err != nil {
		return err
	}

	return nil
}

func Login(ctx context.Context, steps login, dto LoginDto) error {
	if err := domain.Validate(dto); err != nil {
		return err
	}

	email := domain.Email(dto.Email)
	password := domain.Plaintext(dto.Password)

	ciphertext, err := steps.FindEncryptedPasswordByEmail(ctx, email)
	if err != nil {
		return err
	}

	if err := steps.WhenPasswordMatch(ctx, ciphertext.Compare(password)); err != nil {
		return err
	}

	return nil
}
