package otp

import (
	"context"
	"crypto/subtle"
	"errors"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
)

var (
	ErrIdempotentKeyRequired = errors.New("otp: idempotent key required")
	ErrRequestModified       = errors.New("otp: request has been modified")
	ErrTopicRequired         = errors.New("otp: topic required")
)

//go:generate mockery --name sendOtp --case underscore --exported=true
type sendOtp interface {
	Allow(ctx context.Context, dto SendOtpDto) error
	GenerateOtp(ctx context.Context) (domain.OTP, error)
	CreateSession(ctx context.Context, dto SendOtpDto, otp domain.OTP) error
	SendMessage(ctx context.Context, dto SendOtpDto, otp domain.OTP) error
}

type verifyOtp interface {
	Verify(ctx context.Context, dto VerifyOtpDto) (idempotentKey string, err error)
	ClearSession(ctx context.Context, dto VerifyOtpDto) error
}

type SendOtpDto struct {
	PhoneNumber   string `example:"+601243567890" desc:"Phone number in E164 format"`
	IdempotentKey string `desc:"unique key to ensure the request is unique, e.g. using the md5 hash of the request"`
	Topic         string `example:"payout" desc:"Unique topic of the OTP"`
}

func (dto SendOtpDto) Validate() error {
	if err := domain.PhoneNumber(dto.PhoneNumber).Validate(); err != nil {
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

	if err := steps.CreateSession(ctx, dto, otp); err != nil {
		return err
	}

	return steps.SendMessage(ctx, dto, otp)
}

type VerifyOtpDto struct {
	PhoneNumber   string `example:"+601243567890" desc:"Phone number in E164 format"`
	IdempotentKey string `desc:"unique key to ensure the request is unique, e.g. using the md5 hash of the request"`
	Topic         string `example:"payout" desc:"Unique topic of the OTP"`
	OTP           string `example:"123456"`
}

func (dto VerifyOtpDto) Validate() error {
	if err := domain.PhoneNumber(dto.PhoneNumber).Validate(); err != nil {
		return err
	}

	if dto.IdempotentKey == "" {
		return ErrIdempotentKeyRequired
	}

	if dto.Topic == "" {
		return ErrTopicRequired
	}

	return domain.OTP(dto.OTP).Validate()
}

func VerifyOtp(ctx context.Context, steps verifyOtp, dto VerifyOtpDto) error {
	if err := dto.Validate(); err != nil {
		return err
	}

	idempotencyKey, err := steps.Verify(ctx, dto)
	if err != nil {
		return err
	}

	match := subtle.ConstantTimeCompare([]byte(idempotencyKey), []byte(dto.IdempotentKey)) == 1
	if !match {
		return ErrRequestModified
	}

	return steps.ClearSession(ctx, dto)
}
