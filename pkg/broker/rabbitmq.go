package broker

import (
	"context"
	"encoding/json"
	"log"
	"wallet-service/internal/domain"

	ampqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	channel *ampqp.Channel
}

func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
	conn, err := ampqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declare the queue to ensure it exists
	_, err = ch.QueueDeclare(
		"topup_success_event", // Queue name
		true,                  // Durable (survive RabbitMQ restarts)
		false,                 // Auto-delete
		false,                 // Exclusive
		false,                 // No-wait
		nil,                   // Arguments
	)
	if err != nil {
		return nil, err
	}

	log.Println("RabbitMQ connected successfully!")
	return &RabbitMQPublisher{channel: ch}, nil
}

func (r *RabbitMQPublisher) PublishTopUpSuccess(ctx context.Context, tx *domain.Transaction) error {
	body, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	return r.channel.PublishWithContext(ctx,
		"",                     // Exchange (default)
		"topup_success_events", // Routing key (matches queue name)
		false,                  // Mandatory
		false,                  // Immediate
		ampqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: ampqp.Persistent, // Ensure messages are saved to disk
			Body:         body,
		},
	)
}
