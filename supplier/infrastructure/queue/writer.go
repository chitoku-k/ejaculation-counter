package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
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
		Exchange:    exchange,
		RoutingKey:  routingKey,
		Environment: environment,
	}

	go func() {
		<-ctx.Done()
		w.disconnect()
	}()

	return w, w.connect()
}

func (w *writer) connect() error {
	uri, err := amqp.ParseURI(w.Environment.Queue.Host)
	if err != nil {
		return fmt.Errorf("failed to parse MQ URI: %w", err)
	}

	uri.Username = w.Environment.Queue.Username
	uri.Password = w.Environment.Queue.Password

	conn, err := amqp.Dial(uri.String())
	if err != nil {
		return fmt.Errorf("failed to connect to MQ broker: %w", err)
	}

	w.Channel, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel for MQ connection: %w", err)
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
		return fmt.Errorf("failed to declare exchange in MQ channel: %w", err)
	}

	err = w.Channel.Confirm(false)
	if err != nil {
		return fmt.Errorf("failed to put channel into confirm mode: %w", err)
	}

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
	err := w.Channel.Close()
	if err != nil {
		return fmt.Errorf("failed to close the MQ channel: %w", err)
	}
	return nil
}

func (w *writer) Publish(event service.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return w.publish(event.Name(), body, true)
}

func (w *writer) publish(name string, body []byte, retry bool) error {
	err := w.Channel.Publish(
		w.Exchange,
		w.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Type:        name,
			Body:        body,
		},
	)
	QueuedMessageTotal.Inc()

	if err != nil {
		var errs []error
		if retry {
			w.disconnect()
			errs = append(errs, w.connect())
			errs = append(errs, w.publish(name, body, false))
			if errs == nil {
				return nil
			}
		}

		QueuedMessageErrorTotal.Inc()
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
