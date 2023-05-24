package message_queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/shark-ci/shark-ci/models"
)

type RabbitMQ struct {
	conn      *amqp.Connection
	queueName string
}

var _ MessageQueuer = &RabbitMQ{}

func NewRabbitMQ(host string, port string, username string, password string) (*RabbitMQ, error) {
	rmq := &RabbitMQ{queueName: "jobs"}
	var err error

	log.Printf("Connecting to RabbitMQ: amqp://%s:%s@%s:%s\n", username, password, host, port)
	rmq.conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s", username, password, host, port))
	if err != nil {
		return nil, err
	}
	log.Println("RabbitMQ connected")

	channel, err := rmq.conn.Channel()
	if err != nil {
		rmq.conn.Close()
		return nil, err
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(rmq.queueName, false, false, false, false, nil)
	if err != nil {
		rmq.conn.Close()
		return nil, err
	}

	return rmq, nil
}

func (rmq *RabbitMQ) Close(ctx context.Context) error {
	return rmq.conn.Close()
}

func (rmq *RabbitMQ) SendJob(ctx context.Context, job *models.Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	channel, err := rmq.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	pub := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         data,
	}
	err = channel.Publish("", rmq.queueName, false, false, pub)
	return err
}

func (rmq *RabbitMQ) RegisterJobHandler(handler func(job *models.Job) error) error {
	channel, err := rmq.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	msgChannel, err := channel.Consume(rmq.queueName, "", true, false, false, false, nil)
	for msg := range msgChannel {
		var job models.Job
		err := json.Unmarshal(msg.Body, &job)
		if err != nil {
			return err
		}

		err = handler(&job)
		if err != nil {
			return err
		}
	}
	return err
}
