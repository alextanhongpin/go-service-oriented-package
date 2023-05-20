package users

import (
	"github.com/alextanhongpin/go-service-oriented-package/domain/emails"
	"github.com/alextanhongpin/go-service-oriented-package/domain/password"
	"github.com/google/uuid"
)

type User struct {
	ID                uuid.UUID
	Name              string
	Email             emails.Email
	EncryptedPassword password.Ciphertext
}

type CreateUserParams struct {
	Name              string
	Email             string
	EncryptedPassword string
	Meta              map[string]any
}
