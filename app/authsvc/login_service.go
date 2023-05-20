package authsvc

import (
	"context"
	"errors"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/alextanhongpin/go-service-oriented-package/domain/emails"
	"github.com/alextanhongpin/go-service-oriented-package/domain/password"
)

var (
	ErrPasswordDoNotMatch = errors.New("authsvc.Login: password do not match")
)

type User[T any] struct {
	EncryptedPassword password.Ciphertext
	Data              *T
}

//go:generate mockery --name loginRepo --case underscore --exported=true
type loginRepo[T any] interface {
	FindUserByEmail(ctx context.Context, email string) (*User[T], error)
}

type LoginService[T any] struct {
	repo loginRepo[T]
}

func NewLoginService[T any](repo loginRepo[T]) *LoginService[T] {
	return &LoginService[T]{
		repo: repo,
	}
}

type LoginDto struct {
	Email    string
	Password string
}

func (dto LoginDto) Validate() error {
	return domain.Validate(
		emails.Email(dto.Email),
		password.Plaintext(dto.Password),
	)
}

func (s *LoginService[T]) Login(ctx context.Context, dto LoginDto) (*T, error) {
	if err := domain.Validate(dto); err != nil {
		return nil, err
	}

	u, err := s.repo.FindUserByEmail(ctx, dto.Email)
	if err != nil {
		return nil, err
	}

	pwd := password.Plaintext(dto.Password)
	if !u.EncryptedPassword.Compare(pwd) {
		return nil, ErrPasswordDoNotMatch
	}

	return u.Data, nil
}
