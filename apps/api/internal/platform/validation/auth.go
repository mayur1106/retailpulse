package validation

import (
	"errors"
	"net/mail"
	"strings"

	"retailpulse/apps/api/internal/domain"
)

func RegisterRequest(organizationName string, name string, email string, password string, accountType string) error {
	if strings.TrimSpace(organizationName) == "" {
		return errors.Join(domain.ErrValidation, errors.New("organizationName is required"))
	}
	if strings.TrimSpace(name) == "" {
		return errors.Join(domain.ErrValidation, errors.New("name is required"))
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.Join(domain.ErrValidation, errors.New("email must be valid"))
	}
	if len(password) < 12 {
		return errors.Join(domain.ErrValidation, errors.New("password must be at least 12 characters"))
	}
	if accountType != "" && accountType != "owner" && accountType != "seller" {
		return errors.Join(domain.ErrValidation, errors.New("accountType must be owner or seller"))
	}
	return nil
}

func LoginRequest(email string, password string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.Join(domain.ErrValidation, errors.New("email must be valid"))
	}
	if password == "" {
		return errors.Join(domain.ErrValidation, errors.New("password is required"))
	}
	return nil
}
