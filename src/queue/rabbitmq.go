package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher notifies external consumers (the RAG ingestion worker) that a
// book's file is ready to be processed.
type Publisher struct {
	channel   *amqp.Channel
	queueName string
}

// NewPublisherFromEnv builds a Publisher from RABBITMQ_USER,
// RABBITMQ_PASSWORD, RABBITMQ_HOST, RABBITMQ_PORT, and RABBITMQ_QUEUE,
// falling back to this project's local-dev RabbitMQ defaults for whichever
// aren't set.
func NewPublisherFromEnv() (*Publisher, error) {
	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		getEnv("RABBITMQ_USER", "rabbitmq"),
		getEnv("RABBITMQ_PASSWORD", "rabbitmq123"),
		getEnv("RABBITMQ_HOST", "localhost"),
		getEnv("RABBITMQ_PORT", "5672"),
	)
	return NewPublisher(url, getEnv("RABBITMQ_QUEUE", "book-ingestion"))
}

func NewPublisher(url, queueName string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if _, err := channel.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &Publisher{channel: channel, queueName: queueName}, nil
}

// ingestionMessage mirrors lexi-rag-ms's IngestionMessage: file_path holds
// the object key in the shared MinIO bucket, not a filesystem path.
type ingestionMessage struct {
	Book     string `json:"book"`
	FilePath string `json:"file_path"`
}

func (p *Publisher) NotifyBookCreated(ctx context.Context, book, fileKey string) error {
	body, err := json.Marshal(ingestionMessage{Book: book, FilePath: fileKey})
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(ctx, "", p.queueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
