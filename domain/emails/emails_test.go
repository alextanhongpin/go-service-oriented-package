package emails_test

import (
	"testing"

	"github.com/alextanhongpin/go-service-oriented-package/domain/emails"
	"github.com/stretchr/testify/assert"
)

func TestEmail(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		assert := assert.New(t)
		assert.True(emails.Email("a@b.com").Valid())
		assert.True(emails.Email("john.doe@mail.com").Valid())
		assert.True(emails.Email("john.doe+1@mail.com").Valid())
		assert.True(emails.Email("john.doe@hello.com").Valid())
	})

	t.Run("invalid", func(t *testing.T) {
		assert := assert.New(t)
		assert.False(emails.Email("").Valid(), "zero")
		assert.False(emails.Email("john.doe").Valid(), "no domain")
		assert.False(emails.Email("john.doe@com").Valid())
	})
}
