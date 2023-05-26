package otp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alextanhongpin/go-service-oriented-package/app/otp"
	"github.com/alextanhongpin/go-service-oriented-package/app/otp/mocks"
	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendOtp(t *testing.T) {
	type stub struct {
		allowErr         error
		generateOtp      domain.OTP
		generateOtpErr   error
		createSessionErr error
		sendMessageErr   error
	}

	wantErr := errors.New("want")

	testCases := []struct {
		name    string
		stubFn  func(*stub)
		wantErr error
	}{
		{
			name:    "success",
			stubFn:  func(stub *stub) {},
			wantErr: nil,
		},
		{
			name: "allow error",
			stubFn: func(stub *stub) {
				stub.allowErr = wantErr
			},
			wantErr: wantErr,
		},
		{
			name: "generate otp error",
			stubFn: func(stub *stub) {
				stub.generateOtpErr = wantErr
			},
			wantErr: wantErr,
		},
		{
			name: "create session error",
			stubFn: func(stub *stub) {
				stub.createSessionErr = wantErr
			},
			wantErr: wantErr,
		},
		{
			name: "send message error",
			stubFn: func(stub *stub) {
				stub.sendMessageErr = wantErr
			},
			wantErr: wantErr,
		},
	}
	for _, tc := range testCases {
		tc := tc

		args := otp.SendOtpDto{
			PhoneNumber:   "+60123456789",
			Topic:         "payout",
			IdempotentKey: "md5(req)",
		}

		stub := stub{
			generateOtp: domain.OTP("134256"),
		}
		tc.stubFn(&stub)

		t.Run(tc.name, func(t *testing.T) {
			steps := mocks.NewSendOtp(t)
			steps.On("Allow", mock.Anything, args).Return(stub.allowErr).Maybe()
			steps.On("GenerateOtp", mock.Anything).Return(stub.generateOtp, stub.generateOtpErr).Maybe()
			steps.On("CreateSession", mock.Anything, args, stub.generateOtp).Return(stub.createSessionErr).Maybe()
			steps.On("SendMessage", mock.Anything, args, stub.generateOtp).Return(stub.sendMessageErr).Maybe()

			ctx := context.Background()
			err := otp.SendOtp(ctx, steps, args)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
