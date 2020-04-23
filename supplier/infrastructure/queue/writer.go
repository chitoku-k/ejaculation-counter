package queue

import (
	"context"
	"encoding/json"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/streadway/amqp"
)

var (
	QueuedMessageTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "queued_message_total",
		Help:      "Total number of messages attempted to write to message queue.",
	})
	QueuedMessageErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "queued_message_error_total",
		Help:      "Total number of errors when writing to message queue.",
	})
)

type writer struct {
	ctx           context.Context
	Exchange      string
	RoutingKey    string
	Environment   config.Environment
	Channel       *amqp.Channel
	Confirmations chan amqp.Confirmation
}

func NewWriter(
	ctx context.Context,
	exchange string,
	routingKey string,
	environment config.Environment,
) (service.QueueWriter, error) {
	w := &writer{
		ctx:         ctx,
		Exchange:    exchange,
		RoutingKey:  routingKey,
		Environment: environment,
	}

	return w, w.connect()
}

func (w *writer) connect() error {
	uri, err := amqp.ParseURI(w.Environment.Queue.Host)
	if err != nil {
		return errors.Wrap(err, "failed to parse MQ URI")
	}

	uri.Username = w.Environment.Queue.Username
	uri.Password = w.Environment.Queue.Password

	conn, err := amqp.Dial(uri.String())
	if err != nil {
		return errors.Wrap(err, "failed to connect to MQ broker")
	}

	w.Channel, err = conn.Channel()
	if err != nil {
		return errors.Wrap(err, "failed to open a channel for MQ connection")
	}

	err = w.Channel.ExchangeDeclare(
		w.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to declare exchange in MQ channel")
	}

	err = w.Channel.Confirm(false)
	if err != nil {
		return errors.Wrap(err, "failed to put channel into confirm mode")
	}

	go func() {
		<-w.ctx.Done()
		w.disconnect()
	}()

	go func() {
		for {
			_, ok := <-w.Confirmations
			if !ok {
				return
			}
		}
	}()

	w.Confirmations = w.Channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	return nil
}

func (w *writer) disconnect() error {
	<-w.Confirmations
	return errors.Wrap(w.Channel.Close(), "failed to close the MQ channel")
}

func (w *writer) Publish(event service.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "failed to marshal event")
	}

	err = w.Channel.Publish(
		w.Exchange,
		w.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Type:        event.Name(),
			Body:        body,
		},
	)
	QueuedMessageTotal.Inc()

	if err != nil {
		QueuedMessageErrorTotal.Inc()
		return errors.Wrap(err, "failed to publish message")
	}

	return nil
}
