package domain

import (
	"errors"
	"strconv"
)

var (
	ErrOTPInvalidFormat = errors.New("otp: invalid format")
)

type OTP string

func (otp OTP) Validate() error {
	if len(otp) == 0 {
		return ErrOTPInvalidFormat
	}

	_, err := strconv.ParseInt(string(otp), 10, 64)
	if err != nil {
		return ErrOTPInvalidFormat
	}

	return nil
}
