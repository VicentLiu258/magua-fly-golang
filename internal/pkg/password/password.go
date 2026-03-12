package password

import (
	"golang.org/x/crypto/bcrypt"
)

// Hash hashes a plain password with bcrypt.
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Check returns true if plain matches hash.
func Check(hash, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}
