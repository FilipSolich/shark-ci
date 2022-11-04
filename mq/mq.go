package mq

import (
	"encoding/json"
	"fmt"

	"github.com/FilipSolich/ci-server/db"
	amqp "github.com/rabbitmq/amqp091-go"
)

const queueName = "job"

var MQ *MessageQueue

type MessageQueue struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
}

func NewMQ(host string, port string, username string, password string) (*MessageQueue, error) {
	var mq MessageQueue
	var err error
	mq.conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port))
	if err != nil {
		return nil, err
	}

	mq.ch, err = mq.conn.Channel()
	if err != nil {
		mq.conn.Close()
		return nil, err
	}

	mq.queue, err = mq.ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		mq.ch.Close()
		mq.conn.Close()
		return nil, err
	}
	return &mq, nil
}

func (mq *MessageQueue) Close() {
	mq.ch.Close()
	mq.conn.Close()
}

func (mq *MessageQueue) PublishJob(job *db.Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	pub := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         data,
	}
	err = mq.ch.Publish("", queueName, false, false, pub)
	return err
}
