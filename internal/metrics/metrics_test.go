package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestMetricsRegistration(t *testing.T) {
	// Проверяем, что метрики зарегистрированы
	registry := prometheus.NewRegistry()
	registry.MustRegister(RequestsTotal, RequestDuration, ActiveRequests)

	assert.NotNil(t, RequestsTotal)
	assert.NotNil(t, RequestDuration)
	assert.NotNil(t, ActiveRequests)
}

func TestRequestsTotalCounter(t *testing.T) {
	// Создаем тестовый регистр
	registry := prometheus.NewRegistry()
	registry.MustRegister(RequestsTotal)

	// Инкрементируем счетчик
	RequestsTotal.WithLabelValues("GET", "/test", "200").Inc()
	RequestsTotal.WithLabelValues("POST", "/test", "201").Inc()
	RequestsTotal.WithLabelValues("GET", "/test", "500").Inc()

	// Проверяем значения
	metrics, err := registry.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)
}

func TestRequestDurationHistogram(t *testing.T) {
	// Создаем тестовый регистр
	registry := prometheus.NewRegistry()
	registry.MustRegister(RequestDuration)

	// Наблюдаем значения
	RequestDuration.WithLabelValues("GET", "/test").Observe(0.1)
	RequestDuration.WithLabelValues("GET", "/test").Observe(0.2)
	RequestDuration.WithLabelValues("POST", "/api").Observe(0.5)

	// Проверяем, что метрики собраны
	metrics, err := registry.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)
}

func TestActiveRequestsGauge(t *testing.T) {
	// Создаем тестовый регистр
	registry := prometheus.NewRegistry()
	registry.MustRegister(ActiveRequests)

	// Изменяем значение
	ActiveRequests.Inc()
	ActiveRequests.Inc()
	ActiveRequests.Dec()

	metrics, err := registry.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)
}
