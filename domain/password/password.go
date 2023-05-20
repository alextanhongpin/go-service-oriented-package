package password

import (
	"errors"

	"github.com/alextanhongpin/passwd"
)

// PasswordMinLen rules
const PasswordMinLen = 8

// Password errors.
var (
	ErrPasswordTooShort = errors.New("password: too short")
)

// Plaintext representation of password.
type Plaintext string

// Validate validates the plaintext min length.
func (p Plaintext) Validate() error {
	if len(string(p)) < PasswordMinLen {
		return ErrPasswordTooShort
	}

	return nil
}

func (p Plaintext) String() string {
	return "*PASSWORD REDACTED*"
}

// Encrypt encrypts a plaintext into cipthertext.
func (p Plaintext) Encrypt() (Ciphertext, error) {
	ciphertext, err := passwd.Encrypt([]byte(p))
	if err != nil {
		return "", err
	}

	return Ciphertext(ciphertext), nil
}

// Equal checks if two plaintext matches in constant time.
func (p Plaintext) Equal(other Plaintext) bool {
	return passwd.ConstantTimeCompare(string(p), string(other))
}

// Ciphertext is encrypted plaintext.
type Ciphertext string

// Compare checks if the cipthertext is derived from the plaintext password.
func (c Ciphertext) Compare(p Plaintext) bool {
	match, err := passwd.Compare(string(c), []byte(p))
	return err == nil && match
}
