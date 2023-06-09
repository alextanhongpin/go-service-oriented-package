package domain

import (
	"errors"
	"regexp"
)

var emailRe *regexp.Regexp

func init() {
	// JavaScript's version from https://emailregex.com/
	emailRe = regexp.MustCompile(`^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`)
}

// Email errors.
var (
	ErrEmailInvalid = errors.New("email: invalid format")
)

// Email ...
type Email string

// Valid returns true if the format is valid.
func (e Email) Valid() bool {
	return emailRe.MatchString(string(e))
}

// Validate validates tthe email format.
func (e Email) Validate() error {
	if !e.Valid() {
		return ErrEmailInvalid
	}

	return nil
}
