package msession

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// randomTokenBytes is the number of random bytes used to generate a session token.
const randomTokenBytes = 32

// generateRandomString returns a cryptographically random, URL-safe session token.
func generateRandomString() (string, error) {
	b := make([]byte, randomTokenBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(b)

	return token, nil
}

// HashSessionToken deterministically hashes a raw session token with secret (HMAC-SHA256),
// producing the value that should be stored and looked up in session storage. The raw
// token itself should never be persisted.
func HashSessionToken(secret, token string) string {
	hashToken := hmac.New(sha256.New, []byte(secret))
	hashToken.Write([]byte(token))
	return base64.URLEncoding.EncodeToString(hashToken.Sum(nil))
}
