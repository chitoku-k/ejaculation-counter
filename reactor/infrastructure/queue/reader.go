package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	ReconnectInitial = 5 * time.Second
	ReconnectMax     = 320 * time.Second
)

var (
	DeliveredMessageTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "delivered_message_total",
		Help:      "Total number of messages delivered from message queue.",
	}, []string{"type"})
	DeliveredMessageErrorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "delivered_message_error_total",
		Help:      "Total number of errors when delivered from message queue.",
	}, []string{"type"})
)

type reader struct {
	Exchange    string
	QueueName   string
	RoutingKey  string
	Environment config.Environment
	Channel     *amqp.Channel
	Delivery    <-chan amqp.Delivery
}

func NewReader(
	exchange string,
	queueName string,
	routingKey string,
	environment config.Environment,
) (service.QueueReader, error) {
	r := &reader{
		Exchange:    exchange,
		QueueName:   queueName,
		RoutingKey:  routingKey,
		Environment: environment,
	}

	return r, r.connect()
}

func (r *reader) connect() error {
	uri, err := amqp.ParseURI(r.Environment.Queue.Host)
	if err != nil {
		return fmt.Errorf("failed to parse MQ URI: %w", err)
	}

	uri.Username = r.Environment.Queue.Username
	uri.Password = r.Environment.Queue.Password

	conn, err := amqp.Dial(uri.String())
	if err != nil {
		return fmt.Errorf("failed to connect to MQ broker: %w", err)
	}

	r.Channel, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel for MQ connection: %w", err)
	}

	q, err := r.Channel.QueueDeclare(
		r.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue in MQ channel: %w", err)
	}

	err = r.Channel.QueueBind(
		q.Name,
		r.RoutingKey,
		r.Exchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue for MQ channel: %w", err)
	}

	r.Delivery, err = r.Channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	return err
}

func (r *reader) disconnect() error {
	err := r.Channel.Close()
	if err != nil {
		return fmt.Errorf("failed to close the MQ channel: %w", err)
	}
	return nil
}

func (r *reader) Consume(ctx context.Context) (<-chan service.Event, error) {
	ch := make(chan service.Event)

	go func() {
		for {
			reconnect := ReconnectInitial

			select {
			case <-ctx.Done():
				r.disconnect()
				return

			case err := <-r.Channel.NotifyClose(make(chan *amqp.Error)):
				logrus.Errorf("Channel closed: %v", err.Error())

			case message := <-r.Delivery:
				DeliveredMessageTotal.WithLabelValues(message.Type).Inc()

				switch message.Type {
				case "events.reply":
					var event service.ReplyEvent
					err := json.Unmarshal(message.Body, &event)
					if err != nil {
						logrus.Errorln("Failed to decode message (" + message.Type + "): " + err.Error())
						continue
					}
					ch <- &event

				case "events.update":
					var event service.UpdateEvent
					err := json.Unmarshal(message.Body, &event)
					if err != nil {
						logrus.Errorln("Failed to decode message (" + message.Type + "): " + err.Error())
						continue
					}
					ch <- &event

				case "events.increment":
					var event service.IncrementEvent
					err := json.Unmarshal(message.Body, &event)
					if err != nil {
						logrus.Errorln("Failed to decode message (" + message.Type + "): " + err.Error())
						continue
					}
					ch <- &event

				case "events.administration":
					var event service.AdministrationEvent
					err := json.Unmarshal(message.Body, &event)
					if err != nil {
						logrus.Errorln("Failed to decode message (" + message.Type + "): " + err.Error())
						continue
					}
					ch <- &event

				default:
					ch <- &service.ErrorEvent{
						Raw: string(message.Body),
					}
				}

				continue
			}

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
				<-time.Tick(reconnect)

				r.disconnect()
				err := r.connect()
				if err == nil {
					break
				}
			}
		}
	}()

	return ch, nil
}
