package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/alextanhongpin/go-service-oriented-package/app/otp"
	"github.com/alextanhongpin/go-service-oriented-package/domain"
)

var (
	ErrTooManyRequests = errors.New("otp: too many requests")
	ErrKeyNotFound     = errors.New("otp: key not found")
)

func main() {
	ctx := context.Background()
	{
		impl := new(SendOtp)
		impl.cache = new(mockCache)
		dto := otp.SendOtpDto{
			PhoneNumber:   "+60123456789",
			IdempotentKey: "abc",
			Topic:         "payout",
		}

		if err := otp.SendOtp(ctx, impl, dto); err != nil {
			panic(err)
		}
	}
	{

		impl := new(VerifyOtp)
		impl.cache = new(mockCache)
		dto := otp.VerifyOtpDto{
			PhoneNumber:   "+60123456789",
			IdempotentKey: "abc",
			Topic:         "payout",
			OTP:           "123456",
		}

		if err := otp.VerifyOtp(ctx, impl, dto); err != nil {
			panic(err)
		}
	}
	log.Println("done")
}

type cache interface {
	Get(ctx context.Context, key string) (value string, err error)
	Set(ctx context.Context, key, value string, expiresIn time.Duration) error
	Del(ctx context.Context, key string) error
}

type mockCache struct{}

func (c *mockCache) Get(ctx context.Context, key string) (string, error) {
	log.Println("cache: get", key)
	if strings.Contains(key, "code") {
		return "abc", nil
	}

	return "", ErrKeyNotFound
}

func (c *mockCache) Set(ctx context.Context, key, value string, expiresIn time.Duration) error {
	log.Println("cache: set", key, value, expiresIn)
	return nil
}

func (c *mockCache) Del(ctx context.Context, key string) error {
	log.Println("cache: del", key)
	return nil
}

type SendOtp struct {
	cache cache
}

func (s *SendOtp) Allow(ctx context.Context, dto otp.SendOtpDto) error {
	key := fmt.Sprintf("otpsvc:%s:phone:%s", dto.Topic, dto.PhoneNumber)
	_, err := s.cache.Get(ctx, key)
	if err == nil {
		return ErrTooManyRequests
	}

	if errors.Is(err, ErrKeyNotFound) {
		sleep := 1 * time.Minute
		if err := s.cache.Set(ctx, key, dto.IdempotentKey, sleep); err != nil {
			return err
		}

		return nil
	}

	return err
}

func (s *SendOtp) GenerateOtp(ctx context.Context) (domain.OTP, error) {
	return domain.OTP("123456"), nil
}

func (s *SendOtp) CreateSession(ctx context.Context, dto otp.SendOtpDto, otp domain.OTP) error {
	key := fmt.Sprintf("otpsvc:%s:phone:%s:code:%s", dto.Topic, dto.PhoneNumber, otp)
	err := s.cache.Set(ctx, key, dto.IdempotentKey, 3*time.Minute)
	if err != nil {
		return err
	}

	return nil
}

func (s *SendOtp) SendMessage(ctx context.Context, dto otp.SendOtpDto, otp domain.OTP) error {
	log.Println("message sent")
	return nil
}

type VerifyOtp struct {
	cache cache
}

func (s *VerifyOtp) Verify(ctx context.Context, dto otp.VerifyOtpDto) (string, error) {
	key := fmt.Sprintf("otpsvc:%s:phone:%s:code:%s", dto.Topic, dto.PhoneNumber, dto.OTP)
	idempotencyKey, err := s.cache.Get(ctx, key)
	if err != nil {
		return "", err
	}

	return idempotencyKey, nil
}

func (s *VerifyOtp) ClearSession(ctx context.Context, dto otp.VerifyOtpDto) error {
	{
		key := fmt.Sprintf("otpsvc:%s:phone:%s", dto.Topic, dto.PhoneNumber)
		err := s.cache.Del(ctx, key)
		if err != nil {
			return err
		}
	}

	{
		key := fmt.Sprintf("otpsvc:%s:phone:%s:code:%s", dto.Topic, dto.PhoneNumber, dto.OTP)
		err := s.cache.Del(ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}
