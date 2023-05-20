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

type AuthUsecase struct {
	// Method steps
	loginSteps interface {
		Login(ctx context.Context, dto authsvc.LoginDto) (*User, error)
		GenerateToken(ctx context.Context, user *User) (string, error)
	}
	// anotherMethodSteps ..
}

type loginSteps struct {
	*authsvc.LoginService[User]
	*generateTokenStep
}

type LoginDto = authsvc.LoginDto

func (uc *AuthUsecase) Login(ctx context.Context, dto LoginDto) (string, error) {
	defer func(start time.Time) {
		log.Println("took: ", time.Since(start))
	}(time.Now())

	// Add other steps, like monitoring etc.
	user, err := uc.loginSteps.Login(ctx, dto)
	if err != nil {
		return "", err
	}

	return uc.loginSteps.GenerateToken(ctx, user)
}

func main() {
	svc := authsvc.NewLoginService[User](&repository{})
	ctx := context.Background()

	uc := &AuthUsecase{loginSteps: &loginSteps{
		LoginService:      svc,
		generateTokenStep: &generateTokenStep{},
	}}
	user, err := uc.Login(ctx, LoginDto{
		Email:    "john.doe@mail.com",
		Password: "12345678",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(user)
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
