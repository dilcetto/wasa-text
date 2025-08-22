package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

var jwtKey = []byte("your-very-secret-key")

type tokenClaims struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
}

func sign(data string) []byte {
	h := hmac.New(sha256.New, jwtKey)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func createToken(userID string) (string, error) {
	h := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	c := tokenClaims{
		Sub: userID,
		Exp: time.Now().Add(24 * time.Hour).Unix(),
	}
	payloadBytes, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	// Create the token payload
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	sig := sign(h + "." + payload)
	token := h + "." + payload + "." + base64.RawURLEncoding.EncodeToString(sig)
	return token, nil
}

func ParseToken(tokenString string) (string, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "", ErrUnauthorized
	}
	header, payload, sigB64 := parts[0], parts[1], parts[2]
	sig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return "", ErrUnauthorized
	}
	expected := sign(header + "." + payload)
	if !hmac.Equal(sig, expected) {
		return "", ErrUnauthorized
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", ErrUnauthorized
	}
	var claims tokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", ErrUnauthorized
	}
	if claims.Exp < time.Now().Unix() || claims.Sub == "" {
		return "", ErrUnauthorized
	}
	return claims.Sub, nil
}
