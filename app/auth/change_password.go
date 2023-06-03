package auth

import (
	"context"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
)

// Flow to change password for a logged-in User. The User must provide old and
// new password for this process.
type changePassword interface {
	// 1. Login the  user before allowing password change.
	Authenticate(ctx context.Context, email domain.Email, oldPassword domain.Plaintext) error

	// 2. Additional handling when new password is the same as the old domain.
	// Return an error if this is not allowed.
	WhenPasswordIsReused(ctx context.Context, isPasswordReused bool) error

	// 3. Update the domain.
	UpdatePassword(ctx context.Context, email domain.Email, newPassword domain.Ciphertext) error
}

type ChangePasswordDto struct {
	Email       string
	OldPassword string
	NewPassword string
}

func (d ChangePasswordDto) Validate() error {
	return domain.Validate(
		domain.Email(d.Email),
		domain.Plaintext(d.OldPassword),
		domain.Plaintext(d.NewPassword),
	)
}

func ChangePassword(ctx context.Context, steps changePassword, dto ChangePasswordDto) error {
	if err := domain.Validate(dto); err != nil {
		return err
	}

	email := domain.Email(dto.Email)
	oldPwd := domain.Plaintext(dto.OldPassword)
	newPwd := domain.Plaintext(dto.NewPassword)

	if err := steps.Authenticate(ctx, email, oldPwd); err != nil {
		return err
	}

	if err := steps.WhenPasswordIsReused(ctx, oldPwd.Equal(newPwd)); err != nil {
		return err
	}

	pwd, err := newPwd.Encrypt()
	if err != nil {
		return err
	}

	return steps.UpdatePassword(ctx, email, pwd)
}
