package middleware

import (
	"strings"
	"unicode"
)

// PasswordPolicy defines the password requirements
type PasswordPolicy struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
}

// DefaultPasswordPolicy returns the default password policy
func DefaultPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:        12,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumber:    true,
		RequireSpecial:   true,
	}
}

// PasswordValidationResult contains the result of password validation
type PasswordValidationResult struct {
	Valid    bool
	Errors   []string
	Strength string // "weak", "fair", "good", "strong"
}

// ValidatePassword validates a password against the policy
func (p *PasswordPolicy) ValidatePassword(password string) PasswordValidationResult {
	result := PasswordValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Check length
	if len(password) < p.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must be at least 12 characters long")
	}

	// Check for uppercase
	hasUppercase := false
	hasLowercase := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUppercase = true
		case unicode.IsLower(char):
			hasLowercase = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if p.RequireUppercase && !hasUppercase {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one uppercase letter")
	}

	if p.RequireLowercase && !hasLowercase {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one lowercase letter")
	}

	if p.RequireNumber && !hasNumber {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one number")
	}

	if p.RequireSpecial && !hasSpecial {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one special character")
	}

	// Check against common passwords
	if isCommonPassword(password) {
		result.Valid = false
		result.Errors = append(result.Errors, "Password is too common. Please choose a more unique password")
	}

	// Calculate strength
	strength := 0
	if len(password) >= 12 {
		strength++
	}
	if len(password) >= 16 {
		strength++
	}
	if hasUppercase {
		strength++
	}
	if hasLowercase {
		strength++
	}
	if hasNumber {
		strength++
	}
	if hasSpecial {
		strength++
	}

	switch {
	case strength <= 2:
		result.Strength = "weak"
	case strength <= 4:
		result.Strength = "fair"
	case strength <= 5:
		result.Strength = "good"
	default:
		result.Strength = "strong"
	}

	return result
}

// Common passwords to check against
var commonPasswords = map[string]bool{
	"password":       true,
	"password1":      true,
	"password123":    true,
	"123456":         true,
	"12345678":       true,
	"123456789":      true,
	"1234567890":     true,
	"qwerty":         true,
	"qwerty123":      true,
	"abc123":         true,
	"letmein":        true,
	"welcome":        true,
	"monkey":         true,
	"dragon":         true,
	"master":         true,
	"admin":          true,
	"administrator":  true,
	"login":          true,
	"passw0rd":       true,
	"iloveyou":       true,
	"sunshine":       true,
	"princess":       true,
	"football":       true,
	"baseball":       true,
	"trustno1":       true,
	"superman":       true,
	"batman":         true,
	"starwars":       true,
	"password!":      true,
	"password1!":     true,
	"p@ssword":       true,
	"p@ssw0rd":       true,
	"P@ssword1":      true,
	"P@ssw0rd1":      true,
	"Qwerty123":      true,
	"Qwerty123!":     true,
	"Welcome1":       true,
	"Welcome1!":      true,
	"Welcome123":     true,
	"Welcome123!":    true,
	"changeme":       true,
	"changeme1":      true,
	"letmein123":     true,
	"vibetable":      true,
	"vibetable123":   true,
}

func isCommonPassword(password string) bool {
	// Check the password in lowercase
	lower := strings.ToLower(password)
	return commonPasswords[lower]
}
