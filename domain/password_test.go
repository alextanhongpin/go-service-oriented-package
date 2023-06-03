package domain_test

import (
	"testing"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/stretchr/testify/assert"
)

func TestPasswordFormat(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "empty",
			password: "",
			wantErr:  domain.ErrPasswordTooShort,
		},
		{
			name:     "too short",
			password: "1234567",
			wantErr:  domain.ErrPasswordTooShort,
		},
		{
			name:     "spaces preserved",
			password: "        ",
			wantErr:  nil,
		},
		{
			name:     "spaces in between",
			password: "1234 678",
			wantErr:  nil,
		},
		{
			name:     "spaces in the beginning",
			password: " 1234567",
			wantErr:  nil,
		},
		{
			name:     "spaces at the end",
			password: "1234567 ",
			wantErr:  nil,
		},
		{
			name:     "special characters",
			password: "!@#$%^&*",
			wantErr:  nil,
		},
		{
			name:     "valid",
			password: "12345678",
			wantErr:  nil,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			err := domain.Plaintext(tc.password).Validate()
			assert.ErrorIs(err, tc.wantErr, err)
		})
	}
}

func TestEncryptCompare(t *testing.T) {
	assert := assert.New(t)

	pwd := domain.Plaintext("12345678")
	assert.Nil(pwd.Validate())

	ciphertext, err := pwd.Encrypt()
	assert.Nil(err)

	assert.True(ciphertext.Compare(pwd))

	anotherPwd := domain.Plaintext("87654321")
	assert.False(ciphertext.Compare(anotherPwd))
}
