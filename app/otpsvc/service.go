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

const otpKey = "otp:%s:phone:%s:otp:%s"
const rateLimitKey = "otp:%s:phone:%s:ratelimit"
const DefaultOtpTTL = 3 * time.Minute

var (
	// NOTE: The adapter is responsible for returning the correct error type.
	// The service package should not depend on external dependencies.
	ErrKeyNotFound        = errors.New("cache: key not found")
	ErrExternalIDRequired = errors.New("external id required")
	ErrOTPNotFound        = errors.New("otp: not found")
	ErrOTPTooManyRequests = errors.New("otp: too many requests")
)

//go:generate mockery --name smsProvider --case underscore --exported=true
type smsProvider interface {
	Send(ctx context.Context, phoneNumber, message string) error
}

//go:generate mockery --name cache --case underscore --exported=true
type cache interface {
	Inc(ctx context.Context, key string) (int64, error)
	Set(ctx context.Context, key string, val string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	TTL(ctx context.Context, key string) (time.Duration, error)
}

type Config struct {
	// Domain specific identifier, e.g. withdrawal/payment.
	// This is used to ensure no concurrent request is made.
	Domain string `example:"payout"`

	// The app name, to be shown in the OTP.
	App string `example:"MyApp"`

	// The specific template for this OTP message
	Template string `example:"Your %s code is %d"`
}

type Service struct {
	config      Config
	cache       cache
	smsProvider smsProvider
}

func New(cfg Config, cache cache, smsProvider smsProvider) *Service {
	return &Service{
		config:      cfg,
		cache:       cache,
		smsProvider: smsProvider,
	}
}

type SendDto struct {
	PhoneNumber string // PhoneNumber in E164 format.
	ExternalID  string // ExternalID to identify the request. You can for example hash the request.
}

func (dto SendDto) Validate() error {
	if err := domain.PhoneNumber(dto.PhoneNumber).Validate(); err != nil {
		return err
	}
	if dto.ExternalID == "" {
		return ErrExternalIDRequired
	}

	return nil
}

func (s *Service) Send(ctx context.Context, dto SendDto) error {
	if err := domain.Validate(dto); err != nil {
		return err
	}

	// - Cooldown? Skip
	phone := domain.PhoneNumber(dto.PhoneNumber)
	if err := s.allow(ctx, phone.String()); err != nil {
		return err
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
		key := s.rateLimitKey(phone.String())
		val := dto.ExternalID
		ttl := RateLimitDurationByCount(count)
		if err := s.cache.Set(ctx, key, val, ttl); err != nil {
			return err
		}
	}

	otp := NewOTP()

	// Save OTP in cache with TTL 3m
	{
		key := s.otpKey(phone.String(), otp)
		val := dto.ExternalID
		ttl := DefaultOtpTTL
		if err := s.cache.Set(ctx, key, val, ttl); err != nil {
			return err
		}
	}

	// Send OTP
	return s.smsProvider.Send(ctx, phone.String(), fmt.Sprintf(s.config.Template, s.config.App, otp))
}

type VerifyDto struct {
	PhoneNumber string
	OTP         string
}

func (dto VerifyDto) Validate() error {
	return domain.Validate(
		domain.PhoneNumber(dto.PhoneNumber),
		domain.OTP(dto.OTP),
	)
}

// Verify verifies the phone number and OTP, and returns the associated
// external id.
func (s *Service) Verify(ctx context.Context, dto VerifyDto) (string, error) {
	if err := domain.Validate(dto); err != nil {
		return "", err
	}

	phone := domain.PhoneNumber(dto.PhoneNumber)
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

	if err := s.cache.Del(ctx, s.rateLimitKey(phone.String())); err != nil {
		return "", err
	}

	return externalID, nil
}

// TTL returns the remaining time before the user can make another request.
// Used in conjunction with ErrOTPTooManyRequests.
// Useful for UI to display.
func (s *Service) TTL(ctx context.Context, phoneNumber string) (time.Duration, error) {
	return s.cache.TTL(ctx, s.rateLimitKey(phoneNumber))
}

func (s *Service) allow(ctx context.Context, phoneNumber string) error {
	_, err := s.cache.Get(ctx, s.rateLimitKey(phoneNumber))
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return nil
		}

		return err
	}

	return ErrOTPTooManyRequests
}

func (s *Service) otpKey(phoneNumber, otp string) string {
	return fmt.Sprintf(otpKey, s.config.Domain, phoneNumber, otp)
}

func (s *Service) rateLimitKey(phoneNumber string) string {
	return fmt.Sprintf(rateLimitKey, s.config.Domain, phoneNumber)
}

func RateLimitDurationByCount(count int64) time.Duration {
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
