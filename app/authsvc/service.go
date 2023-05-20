package authsvc

import (
	"context"
	"errors"
	"strings"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/alextanhongpin/go-service-oriented-package/domain/emails"
	"github.com/alextanhongpin/go-service-oriented-package/domain/password"
	"github.com/alextanhongpin/go-service-oriented-package/domain/users"
)

type service interface {
	Login(ctx context.Context, dto LoginDto) error
}

var (
	ErrPasswordDoNotMatch = errors.New("authsvc: password do not match")
	ErrNameRequired       = errors.New("authsvc: name is required")
	ErrEmailExists        = errors.New("authsvc: email exists")
)

var _ service = (*Service)(nil)

type repository interface {
	FindUserByEmail(ctx context.Context, email string) (*users.User, error)
	Create(ctx context.Context, params users.CreateUserParams) error
}

type Service struct {
	repo repository
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

func (s *Service) Login(ctx context.Context, dto LoginDto) error {
	if err := domain.Validate(dto); err != nil {
		return err
	}

	u, err := s.repo.FindUserByEmail(ctx, dto.Email)
	if err != nil {
		return err
	}

	pwd := password.Plaintext(dto.Password)
	if !u.EncryptedPassword.Compare(pwd) {
		return ErrPasswordDoNotMatch
	}

	return nil
}

type RegisterDto struct {
	Name     string
	Email    string
	Password string
	Meta     map[string]any
}

func (dto RegisterDto) Validate() error {
	if strings.TrimSpace(dto.Name) == "" {
		return ErrNameRequired
	}

	return domain.Validate(
		emails.Email(dto.Email),
		password.Plaintext(dto.Password),
	)
}

func (s *Service) Register(ctx context.Context, dto RegisterDto) error {
	if err := domain.Validate(dto); err != nil {
		return err
	}

	pwd := password.Plaintext(dto.Password)
	ciphertext, err := pwd.Encrypt()
	if err != nil {
		return err
	}

	err = s.repo.Create(ctx, users.CreateUserParams{
		Name:              dto.Name,
		Email:             dto.Email,
		EncryptedPassword: string(ciphertext),
		Meta:              dto.Meta,
	})
	if errors.Is(err, ErrEmailExists) {
		return ErrEmailExists
	}

	return err
}
