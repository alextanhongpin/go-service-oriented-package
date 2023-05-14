// package otpsvc is an example otp implementation. Don't use this in
// production. Each SMS provider should have a verify method readily available.
// Use that instead of implementing your own.
package otpsvc

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
)

var (
	ErrKeyNotFound = errors.New("cache: key not found")
	ErrOTPNotFound = errors.New("otp: not found")
	ErrCooldown    = errors.New("otp: cooldown")
)

type smsProvider interface {
	Send(ctx context.Context, phoneNumber, message string) error
}

type cache interface {
	Inc(ctx context.Context, key string) (int64, error)
	Set(ctx context.Context, key string, val string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

type Service struct {
	cache       cache
	smsProvider smsProvider

	// Domain specific identifier, e.g. withdrawal/payment.
	// This is used to ensure no concurrent request is made.
	domain string

	// The specific template for this OTP message
	template string
}

func New() *Service {
	return &Service{
		domain:   "unknown",
		template: "OTP is %s",
	}
}

type SendDto struct {
	PhoneNumber string // PhoneNumber in E164 format.
	ExternalID  string // ExternalID to identify the request. You can for example hash the request.
}

func (s *Service) Send(ctx context.Context, dto SendDto) error {
	phone := domain.PhoneNumber(dto.PhoneNumber)
	if err := phone.Validate(); err != nil {
		return fmt.Errorf("%w: %q", err, dto.PhoneNumber)
	}

	// - Cooldown? Skip
	if s.isCooldown(ctx, phone.String()) {
		return ErrCooldown
	}

	// Increment the count.
	count, err := s.cache.Inc(ctx, dto.ExternalID)
	if err != nil {
		return err
	}

	// - count > 10?  Set cooldown = 24h
	// - count > 5?  Set cooldown = 1h
	// - count > 3? Set cooldown = 10m
	// - set cooldown = 1m
	{
		key := s.cooldownKey(phone.String())
		val := dto.ExternalID
		ttl := cooldownDurationByCount(count)
		if err := s.cache.Set(ctx, key, val, ttl); err != nil {
			return err
		}
	}

	otp := NewOTP()

	// Save OTP in cache with TTL 3m
	{
		key := s.otpKey(phone.String(), otp)
		val := dto.ExternalID
		ttl := 3 * time.Minute
		if err := s.cache.Set(ctx, key, val, ttl); err != nil {
			return err
		}
	}

	// Send OTP
	return s.smsProvider.Send(ctx, phone.String(), fmt.Sprintf(s.template, otp))
}

type VerifyDto struct {
	PhoneNumber string
	OTP         string
}

// Verify verifies the phone number and OTP, and returns the associated
// external id.
func (s *Service) Verify(ctx context.Context, dto VerifyDto) (string, error) {
	phone := domain.PhoneNumber(dto.PhoneNumber)
	if err := phone.Validate(); err != nil {
		return "", fmt.Errorf("%w: %q", err, dto.PhoneNumber)
	}

	externalID, err := s.cache.Get(ctx, s.otpKey(phone.String(), dto.OTP))
	if errors.Is(err, ErrKeyNotFound) {
		return "", ErrOTPNotFound
	}
	if err != nil {
		return "", err
	}

	if err := s.cache.Del(ctx, s.otpKey(phone.String(), dto.OTP)); err != nil {
		return "", err
	}

	if err := s.cache.Del(ctx, s.cooldownKey(phone.String())); err != nil {
		return "", err
	}

	return externalID, nil
}

func (s *Service) isCooldown(ctx context.Context, phoneNumber string) bool {
	_, err := s.cache.Get(ctx, s.cooldownKey(phoneNumber))
	return !errors.Is(err, ErrKeyNotFound)
}

func (s *Service) otpKey(phoneNumber, otp string) string {
	return fmt.Sprintf("otp:%s:pn:%s:otp:%s", s.domain, phoneNumber, otp)
}

func (s *Service) cooldownKey(phoneNumber string) string {
	return fmt.Sprintf("otp:%s:pn:%s:cooldown", s.domain, phoneNumber)
}

func cooldownDurationByCount(count int64) time.Duration {
	switch {
	case count > 10:
		return 24 * time.Hour
	case count > 5:
		return 1 * time.Hour
	case count > 3:
		return 10 * time.Minute
	default:
		return 1 * time.Minute
	}
}

// NewOTP returns a new 6 digit OTP.
func NewOTP() string {
	// TODO: Replace with TOTP.
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
