package validator

import (
	"fmt"
	"net/mail"
	"regexp"
	"slices"

	"github.com/valkyraycho/bank_project/utils"
)

var (
	isValidUsername = regexp.MustCompile("^[a-z0-9_]+$").MatchString
	isValidFullName = regexp.MustCompile("^[a-zA-Z\\s]+$").MatchString
)

func ValidateString(s string, minLength, maxLength int) error {
	n := len(s)

	if n < minLength || n > maxLength {
		return fmt.Errorf("must contain from %d-%d characters", minLength, maxLength)
	}
	return nil
}

func ValidateUsername(name string) error {
	if err := ValidateString(name, 3, 100); err != nil {
		return err
	}
	if !isValidUsername(name) {
		return fmt.Errorf("must contain only lower case letters, digits, or underscores")
	}
	return nil
}

func ValidatePassword(pwd string) error {
	return ValidateString(pwd, 8, 100)
}

func ValidateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email address")
	}
	return nil
}

func ValidateFullName(name string) error {
	if err := ValidateString(name, 3, 100); err != nil {
		return err
	}
	if !isValidFullName(name) {
		return fmt.Errorf("must contain only letters or spaces")
	}
	return nil
}

func ValidateID(id int32) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	return nil
}

func ValidateCurrency(currency string) error {
	if !slices.Contains(utils.SupportedCurrencies, currency) {
		return fmt.Errorf("unsupported currency")
	}
	return nil
}
