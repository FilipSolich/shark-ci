package message_queue

import (
	"context"
	"encoding/json"

	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/shark-ci/shark-ci/shared/types"
)

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

var _ MessageQueuer = &RabbitMQ{}

func NewRabbitMQ(rabbitMQURI string) (*RabbitMQ, error) {
	mq := &RabbitMQ{queueName: "work"}
	var err error

	mq.conn, err = amqp.Dial(rabbitMQURI)
	if err != nil {
		return nil, err
	}

	mq.channel, err = mq.conn.Channel()
	if err != nil {
		mq.conn.Close()
		return nil, err
	}

	_, err = mq.channel.QueueDeclare(mq.queueName, true, false, false, false, nil)
	if err != nil {
		mq.channel.Close()
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

	pub := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         data,
	}
	err = mq.channel.PublishWithContext(ctx, "", mq.queueName, false, false, pub)
	return err
}

func (mq *RabbitMQ) WorkChannel() (chan types.Work, error) {
	// TODO: Why this doesn't work?
	// err = channel.Qos(1, 0, false)
	// if err != nil {
	// 	channel.Close()
	// 	return nil, err
	// }

	msgChannel, err := mq.channel.Consume(mq.queueName, "", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	workCh := make(chan types.Work, 100)
	go func() {
		for msg := range msgChannel {
			var work types.Work
			err := json.Unmarshal(msg.Body, &work)
			if err != nil {
				slog.Error("cannot unmarshal job from message queue", "err", err)
				// TODO: Send error to CI server. Reason 500 internal server error.
				continue
			}

			workCh <- work
		}
	}()

	return workCh, nil
}
