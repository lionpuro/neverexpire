package api_test

import (
	"testing"

	"github.com/lionpuro/neverexpire/api"
)

func TestAPIKey(t *testing.T) {
	var rawKey string
	var storedHash string

	t.Run("Generate API key", func(t *testing.T) {
		k, err := api.GenerateKey()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		t.Logf("raw key: %s", k)
		rawKey = k
		storedHash = api.HashKey([]byte(k))
	})

	t.Run("Compare raw key to its hash", func(t *testing.T) {
		match := api.CompareKey(rawKey, storedHash)
		if !match {
			t.Errorf("expected true, got false")
		}
	})

	t.Run("Compare a new key to another keys hash", func(t *testing.T) {
		k, err := api.GenerateKey()
		if err != nil {
			t.Errorf("failed to generate test key: %v", err)
		}
		match := api.CompareKey(k, storedHash)
		if match {
			t.Errorf("expected false, got true")
		}
	})
}
