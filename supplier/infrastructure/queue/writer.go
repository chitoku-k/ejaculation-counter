package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	QueueSize         = 1024
	ConnectionTimeout = 30 * time.Second
	ReconnectInitial  = 5 * time.Second
	ReconnectMax      = 320 * time.Second
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
	Closes        chan *amqp.Error
	Queue         chan service.Event
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
		Queue:       make(chan service.Event, QueueSize),
	}

	return w, w.connect(ctx)
}

func (w *writer) dial(url string) (*amqp.Connection, net.Conn, error) {
	var nc net.Conn
	conn, err := amqp.DialConfig(url, amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			var err error
			nc, err = amqp.DefaultDial(ConnectionTimeout)(network, addr)
			return nc, err
		},
	})
	return conn, nc, err
}

func (w *writer) connect(ctx context.Context) error {
	uri, err := amqp.ParseURI(w.Environment.Queue.Host)
	if err != nil {
		return fmt.Errorf("failed to parse MQ URI: %w", err)
	}

	uri.Username = w.Environment.Queue.Username
	uri.Password = w.Environment.Queue.Password

	conn, nc, err := w.dial(uri.String())
	if err != nil {
		return fmt.Errorf("failed to connect to MQ broker: %w", err)
	}

	w.Channel, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel for MQ connection: %w", err)
	}

	w.Closes = conn.NotifyClose(make(chan *amqp.Error, 1))
	w.Confirmations = w.Channel.NotifyPublish(make(chan amqp.Confirmation, 1))

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
			select {
			case <-ctx.Done():
				w.disconnect()
				return

			case err := <-w.Closes:
				logrus.Infof("Disconnected from MQ: %v", err)
				w.reconnect(ctx)
				return

			case event := <-w.Queue:
				err := w.Publish(event)
				if err != nil {
					logrus.Errorf("Error in publishing from queue: %v", err)
				}

			case <-w.Confirmations:
				return
			}
		}
	}()

	logrus.Infof("Connected to MQ: %v", nc.RemoteAddr())
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

func (w *writer) reconnect(ctx context.Context) {
	reconnect := ReconnectInitial

	for {
		reconnect = time.Duration(
			math.Min(
				math.Max(
					float64(reconnect*2),
					float64(ReconnectInitial),
				),
				float64(ReconnectMax),
			),
		)

		logrus.Infof("Reconnecting in %v...", reconnect)

		select {
		case <-ctx.Done():
			return

		case <-time.After(reconnect):
			w.disconnect()
			err := w.connect(ctx)
			if err != nil {
				continue
			}
			return
		}
	}
}

func (w *writer) Publish(event service.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
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
		select {
		case w.Queue <- event:
			return fmt.Errorf("failed to publish message (requeued): %w", err)

		default:
			return fmt.Errorf("failed to publish message (queue is full): %w", err)
		}
	}

	return nil
}
