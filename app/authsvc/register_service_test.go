package authsvc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alextanhongpin/go-service-oriented-package/app/authsvc"
	"github.com/alextanhongpin/go-service-oriented-package/app/authsvc/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterService(t *testing.T) {
	type TWant bool
	type VWant bool

	type stub struct {
		create    *VWant
		createErr error
	}

	tWant := TWant(true)
	vWant := VWant(true)
	wantErr := errors.New("want")

	testCases := []struct {
		name    string
		argsFn  func(authsvc.RegisterDto[TWant]) authsvc.RegisterDto[TWant]
		stubFn  func(*stub)
		want    *VWant
		wantErr error
	}{
		{
			name: "success",
			want: &vWant,
		},
		{
			name: "create error",
			stubFn: func(stub *stub) {
				stub.create = nil
				stub.createErr = wantErr
			},
			wantErr: wantErr,
		},
		{
			name: "email exists",
			stubFn: func(stub *stub) {
				stub.create = nil
				stub.createErr = authsvc.ErrEmailExists
			},
			wantErr: authsvc.ErrEmailExists,
		},
	}

	for _, tc := range testCases {
		tc := tc

		args := authsvc.RegisterDto[TWant]{
			Name:     "John Doe",
			Email:    "john.doe@mail.com",
			Password: "12345678",
			Data:     tWant,
		}

		if tc.argsFn != nil {
			args = tc.argsFn(args)
		}

		stub := stub{
			create: &vWant,
		}
		if tc.stubFn != nil {
			tc.stubFn(&stub)
		}

		matchCreateParams := mock.MatchedBy(func(dto authsvc.CreateUserParams[TWant]) bool {
			return true
		})

		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewRegisterRepo[TWant, VWant](t)
			repo.On("Create", mock.Anything, matchCreateParams).Return(stub.create, stub.createErr).Once()

			ctx := context.Background()

			svc := authsvc.NewRegisterService[TWant, VWant](repo)
			got, err := svc.Register(ctx, args)

			assert := assert.New(t)
			assert.Equal(tc.want, got)
			assert.ErrorIs(err, tc.wantErr, err)
		})
	}
}
