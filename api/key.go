package api

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"time"
)

type Key struct {
	ID        string    `db:"id"`
	Hash      string    `db:"hash"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

func GenerateKey() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	raw := hex.EncodeToString(b)
	return raw, nil
}

func HashKey(key []byte) string {
	hash := sha256.Sum256(key)
	return hex.EncodeToString(hash[:])
}

func CompareKey(input, hash string) (match bool) {
	hashed := HashKey([]byte(input))
	return subtle.ConstantTimeCompare([]byte(hashed), []byte(hash)) == 1
}

func NewKey(raw, userID string) (*Key, error) {
	h := HashKey([]byte(raw))
	id := raw[:8]
	key := &Key{
		ID:     id,
		Hash:   h,
		UserID: userID,
	}
	return key, nil
}
