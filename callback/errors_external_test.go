package callback_test

import (
	"errors"
	"testing"

	"github.com/gowsp/go-wecom/callback"
)

func TestPublicErrorClassification(t *testing.T) {
	_, err := callback.NewCallbackCrypto("", "", "")
	if !errors.Is(err, callback.ErrConfig) {
		t.Fatalf("configuration error not classifiable: %v", err)
	}
	c, err := callback.NewCallbackCrypto("token", "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG", "corp")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.VerifyURL("invalid", "1", "nonce", "invalid")
	if !errors.Is(err, callback.ErrSignature) {
		t.Fatalf("signature error not classifiable: %v", err)
	}
}
