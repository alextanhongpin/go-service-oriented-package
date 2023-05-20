package phones_test

import (
	"errors"
	"testing"

	"github.com/alextanhongpin/go-service-oriented-package/domain/phones"
)

func TestPhone(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pn, err := phones.New("80123456", "SG")
		if err != nil {
			t.Error(err)
		}

		if err := pn.Validate(); err != nil {
			t.Error(err)
		}

		if want, got := "+6580123456", pn.String(); want != got {
			t.Fatalf("phone number: want %s, got %s", want, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		pn := phones.PhoneNumber("80123456")
		err := pn.Validate()
		if !errors.Is(err, phones.ErrInvalidPhoneNumber) {
			t.Fatalf("phone number: want %q to be invalid, got %v", pn, err)

		}
	})
}
