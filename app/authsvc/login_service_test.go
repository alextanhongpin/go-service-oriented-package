package authsvc_test

import (
	"context"
	"testing"

	"github.com/alextanhongpin/go-service-oriented-package/app/authsvc"
	"github.com/alextanhongpin/go-service-oriented-package/app/authsvc/mocks"
	"github.com/alextanhongpin/go-service-oriented-package/domain/password"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoginService(t *testing.T) {
	type TWant bool

	type stub struct {
		user    *authsvc.User[TWant]
		userErr error
	}

	want := TWant(true)
	testCases := []struct {
		name    string
		argsFn  func(authsvc.LoginDto) authsvc.LoginDto
		stubFn  func(*stub)
		want    *TWant
		wantErr error
	}{
		{
			name: "success",
			want: &want,
		},
		{
			name: "wrong password",
			argsFn: func(dto authsvc.LoginDto) authsvc.LoginDto {
				dto.Password = "87654321"
				return dto
			},
			wantErr: authsvc.ErrPasswordDoNotMatch,
		},
	}

	ciphertext, err := password.Plaintext("12345678").Encrypt()
	assert.Nil(t, err)

	for _, tc := range testCases {
		tc := tc

		args := authsvc.LoginDto{
			Email:    "john.doe@mail.com",
			Password: "12345678",
		}

		if tc.argsFn != nil {
			args = tc.argsFn(args)
		}

		want := TWant(true)
		stub := stub{
			user: &authsvc.User[TWant]{
				EncryptedPassword: ciphertext,
				Data:              &want,
			},
		}
		if tc.stubFn != nil {
			tc.stubFn(&stub)
		}

		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewLoginRepo[TWant](t)
			repo.On("FindUserByEmail", mock.Anything, mock.Anything).Return(stub.user, stub.userErr).Once()

			ctx := context.Background()

			svc := authsvc.NewLoginService[TWant](repo)
			got, err := svc.Login(ctx, args)

			assert := assert.New(t)
			assert.Equal(tc.want, got)
			assert.ErrorIs(err, tc.wantErr, err)
		})
	}
}
