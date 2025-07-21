package testutils

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/lionpuro/neverexpire/users"
)

func RandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func NewTestUser() (users.User, error) {
	id, err := RandomString(24)
	if err != nil {
		return users.User{}, err
	}
	email := id[:8] + "@example.com"
	user := users.User{ID: id, Email: email}
	return user, nil
}
