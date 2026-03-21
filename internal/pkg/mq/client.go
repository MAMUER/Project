package mq

import (
    "fmt"

    amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

func NewClient(url string) (*Client, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("failed to open channel: %w", err)
    }

    return &Client{
        conn:    conn,
        channel: ch,
    }, nil
}

func (c *Client) DeclareQueue(name string, durable bool) error {
    _, err := c.channel.QueueDeclare(
        name,    // name
        durable, // durable
        false,   // delete when unused
        false,   // exclusive
        false,   // no-wait
        nil,     // arguments
    )
    return err
}

func (c *Client) Publish(queue string, body []byte) error {
    return c.channel.Publish(
        "",    // exchange
        queue, // routing key
        false, // mandatory
        false, // immediate
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
        },
    )
}

func (c *Client) Consume(queue string) (<-chan amqp.Delivery, error) {
    return c.channel.Consume(
        queue, // queue
        "",    // consumer
        false, // auto-ack
        false, // exclusive
        false, // no-local
        false, // no-wait
        nil,   // args
    )
}

func (c *Client) Close() error {
    if c.channel != nil {
        c.channel.Close()
    }
    if c.conn != nil {
        return c.conn.Close()
    }
    return nil
}