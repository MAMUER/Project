package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
)

// SignResponse вычисляет HMAC-SHA256 подпись JSON-ответа
func SignResponse(data interface{}, secret string) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(jsonData)
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// VerifyResponse проверяет подпись ответа
func VerifyResponse(data interface{}, signature, secret string) bool {
	expected, err := SignResponse(data, secret)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(expected), []byte(signature))
}
