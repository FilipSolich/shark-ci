package message_queue

import (
	"context"
	"encoding/json"

	"github.com/FilipSolich/shark-ci/shared/model"
	"github.com/FilipSolich/shark-ci/shared/types"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/exp/slog"
)

type RabbitMQ struct {
	conn      *amqp.Connection
	queueName string
}

var _ MessageQueuer = &RabbitMQ{}

func NewRabbitMQ(rabbitMQURI string) (*RabbitMQ, error) {
	mq := &RabbitMQ{queueName: "jobs"}
	var err error

	mq.conn, err = amqp.Dial(rabbitMQURI)
	if err != nil {
		return nil, err
	}

	channel, err := mq.conn.Channel()
	if err != nil {
		mq.conn.Close()
		return nil, err
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(mq.queueName, true, false, false, false, nil)
	if err != nil {
		mq.conn.Close()
		return nil, err
	}

	return mq, nil
}

func (mq *RabbitMQ) Close(ctx context.Context) error {
	return mq.conn.Close()
}

func (mq *RabbitMQ) SendWork(ctx context.Context, work types.Work) error {
	data, err := json.Marshal(work)
	if err != nil {
		return err
	}

	channel, err := mq.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	pub := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         data,
	}
	err = channel.PublishWithContext(ctx, "", mq.queueName, false, false, pub)
	return err
}

func (mq *RabbitMQ) JobChannel() (jobChannel, error) {
	channel, err := mq.conn.Channel()
	if err != nil {
		return nil, err
	}

	// TODO: Why this doesn't work?
	// err = channel.Qos(1, 0, false)
	// if err != nil {
	// 	channel.Close()
	// 	return nil, err
	// }

	msgChannel, err := channel.Consume(mq.queueName, "", false, false, false, false, nil)
	if err != nil {
		channel.Close()
		return nil, err
	}

	jobCh := make(jobChannel)
	go func() {
		for msg := range msgChannel {
			var job model.Job
			err := json.Unmarshal(msg.Body, &job)
			if err != nil {
				slog.Error("cannot unmarshal job from message queue", "err", err)
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
