package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alextanhongpin/go-service-oriented-package/app/authsvc"
	"github.com/alextanhongpin/go-service-oriented-package/domain/password"
	"github.com/google/uuid"
)

type User struct {
	ID    uuid.UUID
	Name  string
	Email string
}

type LoginUsecase struct {
	steps interface {
		Login(ctx context.Context, dto authsvc.LoginDto) (*User, error)
		GenerateToken(ctx context.Context, user *User) (string, error)
	}
}

type steps struct {
	*authsvc.LoginService[User]
	*generateTokenStep
}

type LoginDto = authsvc.LoginDto

func (uc *LoginUsecase) Login(ctx context.Context, dto LoginDto) (string, error) {
	defer func(start time.Time) {
		log.Println("took: ", time.Since(start))
	}(time.Now())

	// Add other steps, like monitoring etc.
	user, err := uc.steps.Login(ctx, dto)
	if err != nil {
		return "", err
	}

	return uc.steps.GenerateToken(ctx, user)
}

func main() {
	svc := authsvc.NewLoginService[User](&repository{})
	ctx := context.Background()

	uc := &LoginUsecase{steps: &steps{
		LoginService:      svc,
		generateTokenStep: &generateTokenStep{},
	}}
	token, err := uc.Login(ctx, LoginDto{
		Email:    "john.doe@mail.com",
		Password: "12345678",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("token:", token)
}

type repository struct{}

func (r *repository) FindUserByEmail(ctx context.Context, email string) (*authsvc.User[User], error) {
	ciphertext, err := password.Plaintext("12345678").Encrypt()
	if err != nil {
		return nil, err
	}

	return &authsvc.User[User]{
		EncryptedPassword: ciphertext,
		Data: &User{
			ID:    uuid.New(),
			Name:  "John Doe",
			Email: "john.doe@mail.com",
		},
	}, nil
}

type generateTokenStep struct {
}

func (s *generateTokenStep) GenerateToken(ctx context.Context, u *User) (string, error) {
	return u.ID.String(), nil
}
