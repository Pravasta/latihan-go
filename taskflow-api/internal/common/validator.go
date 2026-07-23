package common

import (
	"regexp"
)

var emailRegex = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#~$%^&*()+|_.,<>?/{}\-]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}
