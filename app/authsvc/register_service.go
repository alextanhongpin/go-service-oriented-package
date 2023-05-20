package authsvc

import (
	"context"
	"errors"
	"strings"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/alextanhongpin/go-service-oriented-package/domain/emails"
	"github.com/alextanhongpin/go-service-oriented-package/domain/password"
)

var (
	ErrEmailExists  = errors.New("authsvc.Register: email exists")
	ErrNameRequired = errors.New("authsvc.Register: name is required")
)

type CreateUserParams[T any] struct {
	Name              string
	Email             string
	EncryptedPassword string
	Data              T
}

//go:generate mockery --name registerRepo --case underscore --exported=true
type registerRepo[T, V any] interface {
	Create(ctx context.Context, params CreateUserParams[T]) (*V, error)
}

type RegisterService[T, V any] struct {
	repo registerRepo[T, V]
}

func NewRegisterService[T, V any](repo registerRepo[T, V]) *RegisterService[T, V] {
	return &RegisterService[T, V]{
		repo: repo,
	}
}

type RegisterDto[T any] struct {
	Name     string
	Email    string
	Password string
	Data     T
}

func (dto RegisterDto[T]) Validate() error {
	if strings.TrimSpace(dto.Name) == "" {
		return ErrNameRequired
	}

	return domain.Validate(
		emails.Email(dto.Email),
		password.Plaintext(dto.Password),
	)
}

func (s *RegisterService[T, V]) Register(ctx context.Context, dto RegisterDto[T]) (*V, error) {
	if err := domain.Validate(dto); err != nil {
		return nil, err
	}

	pwd := password.Plaintext(dto.Password)
	ciphertext, err := pwd.Encrypt()
	if err != nil {
		return nil, err
	}

	v, err := s.repo.Create(ctx, CreateUserParams[T]{
		Name:              dto.Name,
		Email:             dto.Email,
		EncryptedPassword: string(ciphertext),
		Data:              dto.Data,
	})
	if errors.Is(err, ErrEmailExists) {
		return nil, ErrEmailExists
	}

	return v, err
}
