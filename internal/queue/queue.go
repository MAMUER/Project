package queue

import (
    "context"
    "encoding/json"
    amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
    conn    *amqp.Connection
    channel *amqp.Channel
    queue   string
}

func NewPublisher(url, queue string) (*Publisher, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, err
    }
    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, err
    }
    _, err = ch.QueueDeclare(queue, true, false, false, false, nil)
    if err != nil {
        ch.Close()
        conn.Close()
        return nil, err
    }
    return &Publisher{conn: conn, channel: ch, queue: queue}, nil
}

func (p *Publisher) Publish(ctx context.Context, event interface{}) error {
    body, err := json.Marshal(event)
    if err != nil {
        return err
    }
    return p.channel.PublishWithContext(ctx, "", p.queue, false, false,
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
        })
}

func (p *Publisher) Close() error {
    if p.channel != nil {
        p.channel.Close()
    }
    if p.conn != nil {
        return p.conn.Close()
    }
    return nil
}