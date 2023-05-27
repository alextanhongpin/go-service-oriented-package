package auth

import (
	"context"
	"errors"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/alextanhongpin/go-service-oriented-package/domain/emails"
	"github.com/alextanhongpin/go-service-oriented-package/domain/password"
)

// A continuation of the request reset password for non-logged in user.
type resetPassword interface {
	// 1. Verify the jwt token provided. It should have the email as the subject.
	VerifyToken(ctx context.Context, token string) (emails.Email, error)

	// 2. Encrypt the password, and update the password for the given email.
	UpdatePassword(ctx context.Context, email emails.Email, newPassword password.Ciphertext) error
}

type ResetPasswordDto struct {
	Token       string
	NewPassword string
}

func (d ResetPasswordDto) Validate() error {
	if d.Token == "" {
		return errors.New("auth: token required")
	}

	return domain.Validate(password.Plaintext(d.NewPassword))
}

func ResetPassword(ctx context.Context, steps resetPassword, dto ResetPasswordDto) error {
	if err := domain.Validate(dto); err != nil {
		return err
	}

	email, err := steps.VerifyToken(ctx, dto.Token)
	if err != nil {
		return err
	}

	pwd := password.Plaintext(dto.NewPassword)
	newPwd, err := pwd.Encrypt()
	if err != nil {
		return err
	}

	return steps.UpdatePassword(ctx, email, newPwd)
}
