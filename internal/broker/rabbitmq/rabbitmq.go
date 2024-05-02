package rabbitmq

import (
	"context"
	"fmt"
	"log"

	"user-managment-service/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Broker struct {
	ch        *amqp.Channel
	queueName string
}

func New(cfg config.Broker) (*Broker, error) {
	log.Println(cfg.URL)
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		cfg.QueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Broker{
		ch:        ch,
		queueName: cfg.QueueName,
	}, nil
}

func (b *Broker) ResetPassword(ctx context.Context, email string) error {
	const op = "ResetPassword"

	err := b.ch.PublishWithContext(ctx,
		"",
		b.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(email),
		},
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
