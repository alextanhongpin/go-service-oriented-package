package otp

import (
	"context"
	"errors"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
	"github.com/alextanhongpin/go-service-oriented-package/domain/phones"
)

var (
	ErrIdempotentKeyRequired = errors.New("otp: idempotent key required")
	ErrTopicRequired         = errors.New("otp: topic required")
	ErrTooManyRequests       = errors.New("otp: too many requests")
)

//go:generate mockery --name sendOtp --case underscore --exported=true
type sendOtp interface {
	Allow(ctx context.Context, dto SendOtpDto) error
	GenerateOtp(ctx context.Context) (domain.OTP, error)
	CacheOtp(ctx context.Context, dto SendOtpDto, otp domain.OTP) error
	SendMessage(ctx context.Context, dto SendOtpDto, otp domain.OTP) error
}

type SendOtpDto struct {
	PhoneNumber   string `example:"+601243567890" desc:"Phone number in E164 format"`
	IdempotentKey string `desc:"unique key to ensure the request is unique, e.g. using the md5 hash of the request"`
	Topic         string `example:"payout" desc:"Unique topic of the OTP"`
}

func (dto SendOtpDto) Validate() error {
	if err := phones.PhoneNumber(dto.PhoneNumber).Validate(); err != nil {
		return err
	}

	if dto.IdempotentKey == "" {
		return ErrIdempotentKeyRequired
	}

	if dto.Topic == "" {
		return ErrTopicRequired
	}

	return nil
}

func SendOtp(ctx context.Context, steps sendOtp, dto SendOtpDto) error {
	if err := dto.Validate(); err != nil {
		return err
	}

	if err := steps.Allow(ctx, dto); err != nil {
		return err
	}

	otp, err := steps.GenerateOtp(ctx)
	if err != nil {
		return err
	}

	if err := steps.CacheOtp(ctx, dto, otp); err != nil {
		return err
	}

	return steps.SendMessage(ctx, dto, otp)
}

//func Allow(ctx context.Context, cache Cache, dto SendOtpDto) error {
//key := ""
//v, err := cache.Get(ctx, key)
//if errors.Is(err, ErrKeyNotFound) {
//sleep := 1 * time.Minute
//if err := cache.TTL(key, sleep); err != nil {
//return err
//}

//return nil
//}
//if err != nil {
//return err
//}
//return ErrTooManyRequests
//}
