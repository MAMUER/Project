// internal/middleware/response_signer.go
package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/MAMUER/Project/internal/auth"
)

// ResponseSigner перехватывает тело ответа и добавляет подпись
// Требование #11: Подпись критических ответов
type ResponseSigner struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	written    bool
}

func NewResponseSigner(w http.ResponseWriter) *ResponseSigner {
	return &ResponseSigner{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
}

func (rs *ResponseSigner) Write(b []byte) (int, error) {
	if !rs.written {
		rs.written = true
	}
	return rs.body.Write(b)
}

func (rs *ResponseSigner) WriteHeader(code int) {
	if !rs.written {
		rs.statusCode = code
		rs.written = true
	}
	rs.statusCode = code
}

func (rs *ResponseSigner) Flush() {
	// Отправляем оригинальное тело
	_, _ = rs.ResponseWriter.Write(rs.body.Bytes())
}

// SignCriticalResponses подписывает JSON ответы для критических эндпоинтов
// Требование #11: Клиент проверяет подпись перед отрисовкой
func SignCriticalResponses(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			signer := NewResponseSigner(w)
			next.ServeHTTP(signer, r)

			// Подписываем только успешные JSON ответы
			if signer.statusCode >= 200 && signer.statusCode < 300 {
				signature, err := auth.SignResponse(signer.body.String(), secret)
				if err == nil {
					w.Header().Set("X-Response-Signature", signature)
					w.Header().Set("X-Signature-Algorithm", "HMAC-SHA256")
				}
			}

			// Отправляем тело
			signer.Flush()
		})
	}
}

// SignJSONString вычисляет HMAC-SHA256 подпись строки
func SignJSONString(data, secret string) (string, error) {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// VerifyJSONString проверяет подпись JSON строки
func VerifyJSONString(data, signature, secret string) bool {
	expected, err := SignJSONString(data, secret)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(expected), []byte(signature))
}

// CopyBody копирует тело ответа для последующей обработки
func CopyBody(r *http.Response) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer func() { _ = r.Body.Close() }()
	return io.ReadAll(r.Body)
}
