package otpsvc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alextanhongpin/go-service-oriented-package/app/otpsvc/mocks"
	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendDto(t *testing.T) {
	testCases := []struct {
		name    string
		dtoFn   func(SendDto) SendDto
		wantErr error
	}{
		{
			name: "no phone number",
			dtoFn: func(dto SendDto) SendDto {
				dto.PhoneNumber = ""
				return dto
			},
			wantErr: domain.ErrInvalidPhoneNumber,
		},
		{
			name: "invalid phone number format",
			dtoFn: func(dto SendDto) SendDto {
				dto.PhoneNumber = "60123456789"
				return dto
			},
			wantErr: domain.ErrInvalidPhoneNumber,
		},
		{
			name: "no external id",
			dtoFn: func(dto SendDto) SendDto {
				dto.ExternalID = ""
				return dto
			},
			wantErr: ErrExternalIDRequired,
		},
	}

	for _, tc := range testCases {
		tc := tc

		// Arrange.
		dto := tc.dtoFn(SendDto{
			PhoneNumber: "+60123456789",
			ExternalID:  "some-uuid",
		})

		t.Run(tc.name, func(t *testing.T) {
			// Act.
			svc := New(Config{}, nil, nil)
			gotErr := svc.Send(context.Background(), dto)

			// Assert.
			assert.ErrorIs(t, gotErr, tc.wantErr)
		})
	}
}

func TestSend(t *testing.T) {
	type stub struct {
		getErr             error
		inc                int64
		incErr             error
		setRateLimitKeyErr error
		setOtpKeyErr       error
		sendErr            error
	}

	wantErr := errors.New("want")

	testCases := []struct {
		name    string
		stubFn  func(*stub)
		wantErr error
	}{
		{
			name: "success",
		},
		{
			name: "rate limited",
			stubFn: func(stub *stub) {
				stub.getErr = nil
			},
			wantErr: ErrOTPTooManyRequests,
		},
		{
			name: "get error",
			stubFn: func(stub *stub) {
				stub.getErr = wantErr
			},
			wantErr: wantErr,
		},
		{
			name: "increment error",
			stubFn: func(stub *stub) {
				stub.incErr = wantErr
			},
			wantErr: wantErr,
		},
		{
			name: "set rate limit key error",
			stubFn: func(stub *stub) {
				stub.setRateLimitKeyErr = wantErr
			},
			wantErr: wantErr,
		},
		{
			name: "set otp key error",
			stubFn: func(stub *stub) {
				stub.setOtpKeyErr = wantErr
			},
			wantErr: wantErr,
		},
		{
			name: "send sms error",
			stubFn: func(stub *stub) {
				stub.sendErr = wantErr
			},
			wantErr: wantErr,
		},
	}

	for _, tc := range testCases {
		tc := tc

		// Arrange.
		dto := SendDto{
			PhoneNumber: "+60123456789",
			ExternalID:  "some-uuid",
		}
		stub := stub{
			inc:    10,
			getErr: ErrKeyNotFound,
		}
		if tc.stubFn != nil {
			tc.stubFn(&stub)
		}

		otpKeyPrefix := fmt.Sprintf("otp:Payout:phone:%s:otp:", dto.PhoneNumber)
		rateLimitKey := fmt.Sprintf("otp:Payout:phone:%s:ratelimit", dto.PhoneNumber)
		ttl := RateLimitDurationByCount(stub.inc)

		// Matchers.
		matchByOtpKeyPrefix := mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, otpKeyPrefix)
		})

		t.Run(tc.name, func(t *testing.T) {
			cache := new(mocks.Cache)
			cache.On("Get", mock.Anything, rateLimitKey).Return("", stub.getErr).Once()
			cache.On("Inc", mock.Anything, dto.ExternalID).Return(stub.inc, stub.incErr).Once()
			cache.On("Set", mock.Anything, rateLimitKey, dto.ExternalID, ttl).Return(stub.setRateLimitKeyErr).Once()
			cache.On("Set", mock.Anything, matchByOtpKeyPrefix, dto.ExternalID, DefaultOtpTTL).Return(stub.setOtpKeyErr).Once()
			smsProvider := new(mocks.SmsProvider)
			smsProvider.On("Send", mock.Anything, dto.PhoneNumber, mock.Anything).Return(stub.sendErr).Once()

			cfg := Config{
				App:      "MyApp",
				Template: "Your %stub OTP is %d",
				Domain:   "Payout",
			}

			// Act.
			svc := New(cfg, cache, smsProvider)
			gotErr := svc.Send(context.Background(), dto)

			// Assert.
			assert := assert.New(t)
			assert.ErrorIs(gotErr, tc.wantErr)
		})
	}

}

func TestVerifyDto(t *testing.T) {
	testCases := []struct {
		name    string
		dtoFn   func(VerifyDto) VerifyDto
		wantErr error
	}{
		{
			name: "when no otp",
			dtoFn: func(dto VerifyDto) VerifyDto {
				dto.OTP = ""
				return dto
			},
			wantErr: domain.ErrOTPInvalidFormat,
		},
		{
			name: "when otp is not a number",
			dtoFn: func(dto VerifyDto) VerifyDto {
				dto.OTP = "xyzabc"
				return dto
			},
			wantErr: domain.ErrOTPInvalidFormat,
		},
		{
			name: "when phone number is invalid",
			dtoFn: func(dto VerifyDto) VerifyDto {
				dto.PhoneNumber = "0123456789"
				return dto
			},
			wantErr: domain.ErrInvalidPhoneNumber,
		},
	}

	for _, tc := range testCases {
		tc := tc

		// Arrange.
		dto := tc.dtoFn(VerifyDto{
			PhoneNumber: "+60123456789",
			OTP:         "987654",
		})

		t.Run(tc.name, func(t *testing.T) {
			svc := New(Config{}, nil, nil)
			_, gotErr := svc.Verify(context.Background(), dto)
			assert.ErrorIs(t, gotErr, tc.wantErr)
		})
	}
}

func TestVerify(t *testing.T) {
	type stub struct {
		get                string
		getErr             error
		delOtpKeyErr       error
		delRateLimitKeyErr error
	}

	wantErr := errors.New("want")

	testCases := []struct {
		name    string
		stubFn  func(*stub)
		want    string
		wantErr error
	}{
		{
			name: "success",
			want: "some-uuid",
		},
		{
			name: "when get otp key failed",
			stubFn: func(stub *stub) {
				stub.getErr = wantErr
			},
			wantErr: wantErr,
		},
		{
			name: "when get otp key not found",
			stubFn: func(stub *stub) {
				stub.getErr = ErrKeyNotFound
			},
			wantErr: ErrOTPNotFound,
		},
		{
			name: "when get del otp key failed",
			stubFn: func(stub *stub) {
				stub.delOtpKeyErr = wantErr
			},
			wantErr: wantErr,
		},
		{
			name: "when get del rate limit key failed",
			stubFn: func(stub *stub) {
				stub.delRateLimitKeyErr = wantErr
			},
			wantErr: wantErr,
		},
	}

	for _, tc := range testCases {
		tc := tc

		// Arrange.
		dto := VerifyDto{
			PhoneNumber: "+60123456789",
			OTP:         "987654",
		}

		stub := stub{
			get: "some-uuid",
		}
		if tc.stubFn != nil {
			tc.stubFn(&stub)
		}

		otpKey := fmt.Sprintf("otp:Payout:phone:%s:otp:%s", dto.PhoneNumber, dto.OTP)
		rateLimitKey := fmt.Sprintf("otp:Payout:phone:%s:ratelimit", dto.PhoneNumber)

		t.Run(tc.name, func(t *testing.T) {
			cache := new(mocks.Cache)
			cache.On("Get", mock.Anything, otpKey).Return(stub.get, stub.getErr).Once()
			cache.On("Del", mock.Anything, otpKey).Return(stub.delOtpKeyErr).Once()
			cache.On("Del", mock.Anything, rateLimitKey).Return(stub.delRateLimitKeyErr).Once()

			smsProvider := new(mocks.SmsProvider)

			cfg := Config{
				App:      "MyApp",
				Template: "Your %stub OTP is %d",
				Domain:   "Payout",
			}

			svc := New(cfg, cache, smsProvider)
			got, gotErr := svc.Verify(context.Background(), dto)
			assert := assert.New(t)
			assert.Equal(tc.want, got)
			assert.ErrorIs(gotErr, tc.wantErr)
		})
	}
}

func TestTTL(t *testing.T) {
	type stub struct {
		ttl    time.Duration
		ttlErr error
	}

	wantErr := errors.New("want")

	testCases := []struct {
		name    string
		stubFn  func(*stub)
		want    time.Duration
		wantErr error
	}{
		{
			name: "success",
			want: 10 * time.Second,
		},
		{
			name: "failed",
			stubFn: func(stub *stub) {
				stub.ttl = 0
				stub.ttlErr = wantErr
			},
			wantErr: wantErr,
		},
	}

	for _, tc := range testCases {
		tc := tc
		stub := stub{
			ttl:    10 * time.Second,
			ttlErr: nil,
		}
		if tc.stubFn != nil {
			tc.stubFn(&stub)
		}

		t.Run(tc.name, func(t *testing.T) {
			cache := new(mocks.Cache)
			cache.On("TTL", mock.Anything, "otp:Payout:phone:+60123456789:ratelimit").Return(stub.ttl, stub.ttlErr).Once()
			smsProvider := new(mocks.SmsProvider)

			cfg := Config{
				App:      "MyApp",
				Template: "Your %stub OTP is %d",
				Domain:   "Payout",
			}

			svc := New(cfg, cache, smsProvider)
			got, gotErr := svc.TTL(context.Background(), "+60123456789")
			assert := assert.New(t)
			assert.Equal(tc.want, got)
			assert.ErrorIs(gotErr, tc.wantErr)
		})
	}
}

func TestRateLimitDurationByCount(t *testing.T) {

	testCases := []struct {
		name  string
		count int64
		want  time.Duration
	}{
		{
			name:  ">10",
			count: 11,
			want:  24 * time.Hour,
		},
		{
			name:  "=10",
			count: 10,
			want:  1 * time.Hour,
		},
		{
			name:  ">5",
			count: 6,
			want:  1 * time.Hour,
		},
		{
			name:  "=5",
			count: 5,
			want:  10 * time.Minute,
		},
		{
			name:  "=3",
			count: 3,
			want:  1 * time.Minute,
		},
		{
			name:  "0",
			count: 0,
			want:  1 * time.Minute,
		},
		{
			name:  "<0",
			count: -1,
			want:  1 * time.Minute,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, RateLimitDurationByCount(tc.count))
		})
	}
}

func TestNewOTP(t *testing.T) {
	otp := NewOTP()
	assert := assert.New(t)
	assert.Equal(6, len(otp))
}