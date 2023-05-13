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
	Set(ctx context.Context, key string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

type Service struct {
	cache       cache
	smsProvider smsProvider
}

type SendDto struct {
	PhoneNumberE164 string // PhoneNumber in E164 format.
	ExternalID      string // ExternalID to identify the request.
	// Template?
}

func (s *Service) Send(ctx context.Context, dto SendDto) error {
	phone := domain.PhoneNumber(dto.PhoneNumberE164)
	if err := phone.Validate(); err != nil {
		return fmt.Errorf("%w: %q", err, dto.PhoneNumberE164)
	}

	// - Cooldown? Skip
	if s.isCooldown(ctx, dto.ExternalID) {
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
	cooldownDuration := cooldownDurationByCount(count)
	if err := s.cache.Set(ctx, cooldownKey(dto.ExternalID), cooldownDuration); err != nil {
		return err
	}

	otp := NewOTP()

	// Save OTP in cache with TTL 3m
	otpDuration := 3 * time.Minute
	if err := s.cache.Set(ctx, otpKey(dto.ExternalID, otp), otpDuration); err != nil {
		return err
	}

	// Send OTP
	return s.smsProvider.Send(ctx, phone.String(), fmt.Sprintf("OTP is %s", otp))
}

type VerifyDto struct {
	OTP        string
	ExternalID string
}

func (s *Service) Verify(ctx context.Context, dto VerifyDto) error {
	_, err := s.cache.Get(ctx, otpKey(dto.ExternalID, dto.OTP))
	if errors.Is(err, ErrKeyNotFound) {
		return ErrOTPNotFound
	}
	if err != nil {
		return err
	}

	if err := s.cache.Del(ctx, otpKey(dto.ExternalID, dto.OTP)); err != nil {
		return err
	}

	if err := s.cache.Del(ctx, cooldownKey(dto.ExternalID)); err != nil {
		return err
	}

	return nil
}

func (s *Service) isCooldown(ctx context.Context, id string) bool {
	_, err := s.cache.Get(ctx, cooldownKey(id))
	return !errors.Is(err, ErrKeyNotFound)
}

func otpKey(id, otp string) string {
	return fmt.Sprintf("otp:%s:%s", id, otp)
}

func cooldownKey(id string) string {
	return fmt.Sprintf("otp:cooldown:%s", id)
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
