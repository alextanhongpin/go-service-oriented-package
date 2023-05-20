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

type registerRepo[T any] interface {
	Create(ctx context.Context, params CreateUserParams[T]) error
}

type RegisterService[T any] struct {
	repo registerRepo[T]
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

func (s *RegisterService[T]) Register(ctx context.Context, dto RegisterDto[T]) error {
	if err := domain.Validate(dto); err != nil {
		return err
	}

	pwd := password.Plaintext(dto.Password)
	ciphertext, err := pwd.Encrypt()
	if err != nil {
		return err
	}

	err = s.repo.Create(ctx, CreateUserParams[T]{
		Name:              dto.Name,
		Email:             dto.Email,
		EncryptedPassword: string(ciphertext),
		Data:              dto.Data,
	})
	if errors.Is(err, ErrEmailExists) {
		return ErrEmailExists
	}

	return err
}
