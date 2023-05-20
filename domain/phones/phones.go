package phones

import (
	"errors"
	"fmt"

	"github.com/nyaruka/phonenumbers"
)

const DefaultRegion = "US"

var (
	ErrInvalidPhoneNumber       = errors.New("phone: invalid phone number")
	ErrInvalidPhoneNumberFormat = errors.New("phone: invalid phone number format")
)

func New(maybePhoneNumber, region string) (PhoneNumber, error) {
	phone, err := phonenumbers.Parse(maybePhoneNumber, region)
	if err != nil {
		return "", err
	}
	if !phonenumbers.IsValidNumber(phone) {
		return "", ErrInvalidPhoneNumber
	}

	e164 := phonenumbers.Format(phone, phonenumbers.E164)
	return PhoneNumber(e164), nil
}

// PhoneNumber
type PhoneNumber string

// Validate validates that the number is a valid E164 format.
func (pn PhoneNumber) Validate() error {
	// If the phone number is already in E164 format, the
	// DefaultRegion plays no role here.
	phone, err := phonenumbers.Parse(string(pn), DefaultRegion)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidPhoneNumber, err)
	}

	if !phonenumbers.IsValidNumber(phone) {
		return ErrInvalidPhoneNumber
	}

	e164 := phonenumbers.Format(phone, phonenumbers.E164)
	if e164 != pn.String() {
		return ErrInvalidPhoneNumberFormat
	}

	return nil
}

func (pn PhoneNumber) String() string {
	return string(pn)
}
