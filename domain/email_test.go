package domain_test

import (
	"testing"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/stretchr/testify/assert"
)

func TestEmail(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		assert := assert.New(t)
		assert.True(domain.Email("a@b.com").Valid())
		assert.True(domain.Email("john.doe@mail.com").Valid())
		assert.True(domain.Email("john.doe+1@mail.com").Valid())
		assert.True(domain.Email("john.doe@hello.com").Valid())
		assert.Nil(domain.Email("john.doe@hello.com").Validate())
	})

	t.Run("invalid", func(t *testing.T) {
		assert := assert.New(t)
		assert.False(domain.Email("").Valid(), "zero")
		assert.False(domain.Email("john.doe").Valid(), "no domain")
		assert.False(domain.Email("john.doe@com").Valid())
		assert.ErrorIs(domain.Email("john.doe@com").Validate(), domain.ErrEmailInvalid)
	})
}
