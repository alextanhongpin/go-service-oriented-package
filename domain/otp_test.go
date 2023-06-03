package domain_test

import (
	"testing"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/stretchr/testify/assert"
)

func TestOTP(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		assert := assert.New(t)

		otp := domain.OTP("123456")
		assert.Nil(otp.Validate())
	})

	t.Run("invalid", func(t *testing.T) {
		assert := assert.New(t)

		assert.ErrorIs(domain.OTP("").Validate(), domain.ErrOTPInvalidFormat)
		assert.ErrorIs(domain.OTP("abc").Validate(), domain.ErrOTPInvalidFormat)
	})
}
