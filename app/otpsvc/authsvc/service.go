package authsvc

import (
	"context"
)

type service interface {
	Login(ctx context.Context) error
}

var _ service = (*Service)(nil)

type Service struct {
}

type LoginDto struct {
	Email    domain.Email
	Password domain.Plaintext
}

func (s *Service) Login(ctx context.Context) error {
	return nil
}

func (s *Service) Register(ctx context.Context) error {
	return nil
}
