package auth

import (
	"context"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
)

// Flow for when a non-logged in User forgots the password.
type requestResetPassword interface {
	// 1. Checks if the email exists.
	CheckEmailExists(ctx context.Context, email domain.Email) (bool, error)

	// 2. What error to return if the email do/do not exists?
	WhenEmailExists(ctx context.Context, exists bool) error

	// 3. Generate a jwt token with the email as the subject.
	// Add expiration when needed.
	GenerateToken(ctx context.Context, email domain.Email) (string, error)

	// 4. Send the email containing the reset password link and token to the
	// email.
	SendResetPasswordEmail(ctx context.Context, email domain.Email, token string) error
}

type RequestResetPasswordDto struct {
	Email string
}

func (d RequestResetPasswordDto) Validate() error {
	return domain.Email(d.Email).Validate()
}

func RequestResetPassword(ctx context.Context, steps requestResetPassword, dto RequestResetPasswordDto) error {
	if err := domain.Validate(dto); err != nil {
		return err
	}

	email := domain.Email(dto.Email)
	exists, err := steps.CheckEmailExists(ctx, email)
	if err != nil {
		return err
	}
	if err := steps.WhenEmailExists(ctx, exists); err != nil {
		return err
	}

	token, err := steps.GenerateToken(ctx, email)
	if err != nil {
		return err
	}

	return steps.SendResetPasswordEmail(ctx, email, token)
}
