package auth

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var dummyPasswordHash = mustDummyBcryptHash()

func mustDummyBcryptHash() string {
	h, err := bcrypt.GenerateFromPassword([]byte("dummy"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(h)
}

// ComparePasswordHash returns true when plain matches stored bcrypt hash.
func ComparePasswordHash(storedHash, plain string) bool {
	storedHash = strings.TrimSpace(storedHash)
	if storedHash == "" {
		_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(plain))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plain)) == nil
}
