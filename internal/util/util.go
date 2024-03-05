package util

import "unicode"

func PasswordIsValid(password string) bool {
	if len(password) < 6 {
		return false
	}
	var (
		hasUpper, hasLower, hasDigit, hasSpecial bool
	)
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsDigit(char) {
			hasDigit = true
		} else if isSpecial(char) {
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func isSpecial(char rune) bool {
	specialCharacters := "!@#$%^&*()_+{}[]|:;<>,.?/~"
	for _, special := range specialCharacters {
		if char == special {
			return true
		}
	}
	return false
}
