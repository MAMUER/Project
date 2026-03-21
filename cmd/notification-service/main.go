package main

import (
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

type Notification struct {
    UserID    string `json:"user_id"`
    Type      string `json:"type"`
    Title     string `json:"title"`
    Body      string `json:"body"`
    CreatedAt string `json:"created_at"`
}

func main() {
    log.Println("Notification service starting...")

    rabbitURL := os.Getenv("RABBITMQ_URL")
    if rabbitURL == "" {
        rabbitURL = "amqp://guest:guest@localhost:5672/"
    }

    conn, err := amqp.Dial(rabbitURL)
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        log.Fatalf("Failed to open channel: %v", err)
    }
    defer ch.Close()

    // Объявление очереди уведомлений
    queueName := "notifications"
    _, err = ch.QueueDeclare(
        queueName,
        true,  // durable
        false, // delete when unused
        false, // exclusive
        false, // no-wait
        nil,   // arguments
    )
    if err != nil {
        log.Fatalf("Failed to declare queue: %v", err)
    }

    msgs, err := ch.Consume(
        queueName,
        "",
        false, // auto-ack
        false, // exclusive
        false, // no-local
        false, // no-wait
        nil,   // args
    )
    if err != nil {
        log.Fatalf("Failed to register consumer: %v", err)
    }

    log.Println("Notification service ready, waiting for messages...")

    go func() {
        for msg := range msgs {
            var notification Notification
            if err := json.Unmarshal(msg.Body, &notification); err != nil {
                log.Printf("Failed to unmarshal notification: %v", err)
                msg.Ack(false)
                continue
            }

            // Отправка уведомления (в реальном приложении - push, email, sms)
            log.Printf("Sending notification to user %s: %s - %s",
                notification.UserID, notification.Title, notification.Body)

            // Имитация отправки
            time.Sleep(100 * time.Millisecond)

            msg.Ack(false)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Notification service shutting down...")
}