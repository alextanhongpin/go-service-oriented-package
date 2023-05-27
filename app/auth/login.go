package auth

import (
	"context"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/alextanhongpin/go-service-oriented-package/domain/emails"
	"github.com/alextanhongpin/go-service-oriented-package/domain/password"
)

type login interface {
	FindEncryptedPasswordByEmail(ctx context.Context, email emails.Email) (password.Ciphertext, error)
	WhenPasswordMatch(ctx context.Context, match bool) error
}

type LoginDto struct {
	Email    string
	Password string
}

func (d LoginDto) Validate() error {
	if err := emails.Email(d.Email).Validate(); err != nil {
		return err
	}

	if err := password.Plaintext(d.Password).Validate(); err != nil {
		return err
	}

	return nil
}

func Login(ctx context.Context, steps login, dto LoginDto) error {
	if err := domain.Validate(dto); err != nil {
		return err
	}

	email := emails.Email(dto.Email)
	password := password.Plaintext(dto.Password)

	ciphertext, err := steps.FindEncryptedPasswordByEmail(ctx, email)
	if err != nil {
		return err
	}

	if err := steps.WhenPasswordMatch(ctx, ciphertext.Compare(password)); err != nil {
		return err
	}

	return nil
}
