package mq

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/shark-ci/shark-ci/db"
)

const queueName = "jobs"

var MQ *MessageQueue

type MessageQueue struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
}

type close func()

func InitMQ(host string, port string, username string, password string) (close, error) {
	MQ = &MessageQueue{}
	var err error
	MQ.conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port))
	if err != nil {
		return nil, err
	}

	MQ.ch, err = MQ.conn.Channel()
	if err != nil {
		MQ.conn.Close()
		return nil, err
	}

	closeFn := func() {
		MQ.ch.Close()
		MQ.conn.Close()
	}

	MQ.queue, err = MQ.ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		closeFn()
		return nil, err
	}

	return closeFn, nil
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

func (mq *MessageQueue) GetJobsChanel() (<-chan amqp.Delivery, error) {
	msg, err := mq.ch.Consume(queueName, "", true, false, false, false, nil)
	return msg, err
}
