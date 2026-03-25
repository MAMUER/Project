package queue

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPublisher(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test - use -short flag to skip")
	}

	url := "amqp://guest:guest@localhost:5672/"
	queue := "test_queue"

	pub, err := NewPublisher(url, queue)
	if err != nil {
		t.Skip("RabbitMQ not available, skipping test")
	}
	defer pub.Close()

	assert.NotNil(t, pub)
	assert.NotNil(t, pub.conn)
	assert.NotNil(t, pub.channel)
	assert.Equal(t, queue, pub.queue)
}

func TestNewPublisherInvalidURL(t *testing.T) {
	pub, err := NewPublisher("amqp://invalid:5672/", "test_queue")
	assert.Error(t, err)
	assert.Nil(t, pub)
}

func TestPublish(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test - use -short flag to skip")
	}

	url := "amqp://guest:guest@localhost:5672/"
	queue := "test_queue"

	pub, err := NewPublisher(url, queue)
	if err != nil {
		t.Skip("RabbitMQ not available, skipping test")
	}
	defer pub.Close()

	tests := []struct {
		name  string
		event interface{}
	}{
		{"simple map", map[string]interface{}{"test": "message", "id": 123}},
		{"struct", struct{ Name string }{"test"}},
		{"array", []string{"item1", "item2"}},
		{"complex nested", map[string]interface{}{
			"user_id": "user-123",
			"metrics": map[string]interface{}{
				"heart_rate": 72,
				"spo2":       98,
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := pub.Publish(ctx, tt.event)
			assert.NoError(t, err)
		})
	}
}

func TestNewConsumer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test - use -short flag to skip")
	}

	url := "amqp://guest:guest@localhost:5672/"
	queue := "test_queue"

	// Создаем publisher, чтобы очередь существовала
	pub, err := NewPublisher(url, queue)
	if err != nil {
		t.Skip("RabbitMQ not available, skipping test")
	}
	pub.Close()

	consumer, err := NewConsumer(url, queue)
	if err != nil {
		t.Skip("RabbitMQ not available, skipping test")
	}
	defer consumer.Close()

	assert.NotNil(t, consumer)
	assert.NotNil(t, consumer.msgs)
}

func TestPublishAndConsume(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test - use -short flag to skip")
	}

	url := "amqp://guest:guest@localhost:5672/"
	queue := "test_queue"

	// Создаем publisher
	pub, err := NewPublisher(url, queue)
	if err != nil {
		t.Skip("RabbitMQ not available, skipping test")
	}
	defer pub.Close()

	// Создаем consumer
	consumer, err := NewConsumer(url, queue)
	if err != nil {
		t.Skip("RabbitMQ not available, skipping test")
	}
	defer consumer.Close()

	// Канал для получения сообщений
	received := make(chan map[string]interface{}, 1)

	go func() {
		for msg := range consumer.Messages() {
			var data map[string]interface{}
			if err := json.Unmarshal(msg.Body, &data); err == nil {
				received <- data
				msg.Ack(false)
			}
		}
	}()

	// Публикуем сообщение
	event := map[string]interface{}{
		"test": "consume",
		"id":   12345,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pub.Publish(ctx, event)
	require.NoError(t, err)

	// Ждем сообщение
	select {
	case receivedEvent := <-received:
		assert.Equal(t, "consume", receivedEvent["test"])
		assert.Equal(t, float64(12345), receivedEvent["id"])
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestPublisherClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test - use -short flag to skip")
	}

	url := "amqp://guest:guest@localhost:5672/"
	queue := "test_queue"

	pub, err := NewPublisher(url, queue)
	if err != nil {
		t.Skip("RabbitMQ not available, skipping test")
	}

	err = pub.Close()
	assert.NoError(t, err)

	// Повторный close не должен вызывать ошибку
	err = pub.Close()
	assert.NoError(t, err)
}

func TestConsumerClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test - use -short flag to skip")
	}

	url := "amqp://guest:guest@localhost:5672/"
	queue := "test_queue"

	// Создаем publisher, чтобы очередь существовала
	pub, err := NewPublisher(url, queue)
	if err != nil {
		t.Skip("RabbitMQ not available, skipping test")
	}
	pub.Close()

	consumer, err := NewConsumer(url, queue)
	if err != nil {
		t.Skip("RabbitMQ not available, skipping test")
	}

	err = consumer.Close()
	assert.NoError(t, err)

	// Повторный close не должен вызывать ошибку
	err = consumer.Close()
	assert.NoError(t, err)
}
