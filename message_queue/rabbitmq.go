package message_queue

import (
	"context"
	"encoding/json"
	"log"

	"github.com/FilipSolich/shark-ci/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn      *amqp.Connection
	queueName string
}

var _ MessageQueuer = &RabbitMQ{}

func NewRabbitMQ(rabbitMQURI string) (*RabbitMQ, error) {
	rmq := &RabbitMQ{queueName: "jobs"}
	var err error

	log.Printf("Connecting to RabbitMQ")
	rmq.conn, err = amqp.Dial(rabbitMQURI)
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

	_, err = channel.QueueDeclare(rmq.queueName, true, false, false, false, nil)
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
	err = channel.PublishWithContext(ctx, "", rmq.queueName, false, false, pub)
	return err
}

func (rmq *RabbitMQ) JobChannel() (jobChannel, error) {
	channel, err := rmq.conn.Channel()
	if err != nil {
		return nil, err
	}

	// TODO: Why this doesn't work?
	// err = channel.Qos(1, 0, false)
	// if err != nil {
	// 	channel.Close()
	// 	return nil, err
	// }

	msgChannel, err := channel.Consume(rmq.queueName, "", false, false, false, false, nil)
	if err != nil {
		channel.Close()
		return nil, err
	}

	jobCh := make(jobChannel)
	go func() {
		for msg := range msgChannel {
			var job models.Job
			err := json.Unmarshal(msg.Body, &job)
			if err != nil {
				log.Println(err)
				msg.Nack(false, false)
				continue
			}

			job.Ack = func() error {
				return msg.Ack(false)
			}
			job.Nack = func() error {
				return msg.Nack(false, true)
			}

			jobCh <- job
		}
		channel.Close()
	}()

	return jobCh, nil
}
