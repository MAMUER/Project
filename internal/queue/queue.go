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

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
	msgs    <-chan amqp.Delivery
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

func NewConsumer(url, queue string) (*Consumer, error) {
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

	msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &Consumer{conn: conn, channel: ch, queue: queue, msgs: msgs}, nil
}

func (c *Consumer) Messages() <-chan amqp.Delivery {
	return c.msgs
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
