package middleware

import (
	"net/http"
)

// RemoveServerHeader удаляет заголовок Server
// Требование #5: Маскировка версий серверного ПО
func RemoveServerHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Del("Server")
		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders добавляет заголовки безопасности
// Требование #5: Удаление информации о версии
// Требование #12: Строгая Content Security Policy
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Требование #5: Удаляем все заголовки с версиями ПО
		w.Header().Del("Server")
		w.Header().Del("X-Powered-By")
		w.Header().Del("X-AspNet-Version")

		// Защита от MIME-sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Защита от кликджекинга
		w.Header().Set("X-Frame-Options", "DENY")

		// XSS Protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Политика реферера
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Требование #12: Строгая Content Security Policy
		// - script-src: свой домен + Chart.js CDN (jsdelivr.net — репутативный CDN)
		// - style-src: свой домен + 'unsafe-inline' (необходимо для динамических стилей)
		// - img-src: свой домен + data: (для inline изображений)
		// - connect-src: только свой домен (AJAX/fetch)
		// - font-src: только свой домен
		// - frame-ancestors: запрет встраивания
		// - base-uri: только свой домен
		// - form-action: только свой домен
		w.Header().Set("Content-Security-Policy",
			"default-src 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'; "+
				"script-src 'self' https://cdn.jsdelivr.net; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; "+
				"font-src 'self'; "+
				"connect-src 'self'; "+
				"media-src 'self'; "+
				"object-src 'none'; "+
				"frame-ancestors 'none'; "+
				"upgrade-insecure-requests;",
		)

		// HSTS для HTTPS
		if r.TLS != nil {
			w.Header().Set(
				"Strict-Transport-Security",
				"max-age=63072000; includeSubDomains; preload",
			)
		}

		// Запрет кеширования для авторизованных страниц
		if r.Header.Get("Authorization") != "" {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		next.ServeHTTP(w, r)
	})
}

// LogoutHeaders добавляет заголовки для принудительной инвалидации сессии
// Требование #1: Явное указание браузеру на удаление cookies
func LogoutHeaders() http.Header {
	h := make(http.Header)
	h.Set("Set-Cookie", "token=; Path=/; HttpOnly; Secure; SameSite=Strict; Max-Age=0; Expires=Thu, 01 Jan 1970 00:00:00 GMT")
	h.Set("Cache-Control", "no-store, no-cache, must-revalidate")
	h.Set("Pragma", "no-cache")
	h.Set("Expires", "0")
	return h
}
