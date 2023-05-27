package auth

import (
	"context"
	"errors"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/alextanhongpin/go-service-oriented-package/domain/emails"
	"github.com/alextanhongpin/go-service-oriented-package/domain/password"
)

type register interface {
	CheckEmailExists(ctx context.Context, email emails.Email) (bool, error)
	WhenEmailExists(ctx context.Context, exists bool) error
	CreateUser(ctx context.Context, name string, email emails.Email, ciphertext password.Ciphertext) error
}

type RegisterDto struct {
	Name     string
	Email    string
	Password string
}

func (d RegisterDto) Validate() error {
	if d.Name == "" {
		return errors.New("name is required")
	}

	return domain.Validate(
		emails.Email(d.Email),
		password.Plaintext(d.Password),
	)
}

func Register(ctx context.Context, steps register, dto RegisterDto) error {
	if err := domain.Validate(dto); err != nil {
		return err
	}

	name := dto.Name
	email := emails.Email(dto.Email)

	exists, err := steps.CheckEmailExists(ctx, email)
	if err != nil {
		return err
	}

	if err := steps.WhenEmailExists(ctx, exists); err != nil {
		return err
	}

	password := password.Plaintext(dto.Password)
	ciphertext, err := password.Encrypt()
	if err != nil {
		return err
	}

	return steps.CreateUser(ctx, name, email, ciphertext)
}
