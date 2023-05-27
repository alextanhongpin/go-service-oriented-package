package auth

import (
	"context"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/alextanhongpin/go-service-oriented-package/domain/emails"
	"github.com/alextanhongpin/go-service-oriented-package/domain/password"
)

// Flow to change password for a logged-in User. The User must provide old and
// new password for this process.
type changePassword interface {
	// 1. Login the  user before allowing password change.
	Authenticate(ctx context.Context, email emails.Email, oldPassword password.Plaintext) error

	// 2. Additional handling when new password is the same as the old password.
	// Return an error if this is not allowed.
	WhenPasswordIsReused(ctx context.Context, isPasswordReused bool) error

	// 3. Update the password.
	UpdatePassword(ctx context.Context, email emails.Email, newPassword password.Ciphertext) error
}

type ChangePasswordDto struct {
	Email       string
	OldPassword string
	NewPassword string
}

func (d ChangePasswordDto) Validate() error {
	return domain.Validate(
		emails.Email(d.Email),
		password.Plaintext(d.OldPassword),
		password.Plaintext(d.NewPassword),
	)
}

func ChangePassword(ctx context.Context, steps changePassword, dto ChangePasswordDto) error {
	if err := domain.Validate(dto); err != nil {
		return err
	}

	email := emails.Email(dto.Email)
	oldPwd := password.Plaintext(dto.OldPassword)
	newPwd := password.Plaintext(dto.NewPassword)

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
