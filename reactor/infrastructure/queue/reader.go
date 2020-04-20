package queue

import (
	"context"
	"encoding/json"
	"log"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/streadway/amqp"
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
	ctx         context.Context
	Exchange    string
	QueueName   string
	RoutingKey  string
	Environment config.Environment
	Channel     *amqp.Channel
	Delivery    <-chan amqp.Delivery
}

func NewReader(
	ctx context.Context,
	exchange string,
	queueName string,
	routingKey string,
	environment config.Environment,
) (service.QueueReader, error) {
	r := &reader{
		ctx:         ctx,
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
		return errors.Wrap(err, "failed to parse MQ URI")
	}

	uri.Username = r.Environment.Queue.Username
	uri.Password = r.Environment.Queue.Password

	conn, err := amqp.Dial(uri.String())
	if err != nil {
		return errors.Wrap(err, "failed to connect to MQ broker")
	}

	r.Channel, err = conn.Channel()
	if err != nil {
		return errors.Wrap(err, "failed to open a channel for MQ connection")
	}

	q, err := r.Channel.QueueDeclare(
		r.QueueName,
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to declare queue in MQ channel")
	}

	err = r.Channel.QueueBind(
		q.Name,
		r.RoutingKey,
		r.Exchange,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to bind queue for MQ channel")
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

func (r *reader) Consume() (<-chan service.Event, error) {
	ch := make(chan service.Event)

	go func() {
		for {
			select {
			case <-r.ctx.Done():
				return

			case message := <-r.Delivery:
				DeliveredMessageTotal.WithLabelValues(message.Type).Inc()

				switch message.Type {
				case "events.reply":
					var event service.ReplyEvent
					err := json.Unmarshal(message.Body, &event)
					if err != nil {
						log.Println("Failed to decode message (" + message.Type + "): " + err.Error())
						continue
					}
					ch <- &event

				case "events.update":
					var event service.UpdateEvent
					err := json.Unmarshal(message.Body, &event)
					if err != nil {
						log.Println("Failed to decode message (" + message.Type + "): " + err.Error())
						continue
					}
					ch <- &event

				case "events.increment":
					var event service.IncrementEvent
					err := json.Unmarshal(message.Body, &event)
					if err != nil {
						log.Println("Failed to decode message (" + message.Type + "): " + err.Error())
						continue
					}
					ch <- &event

				case "events.administration":
					var event service.AdministrationEvent
					err := json.Unmarshal(message.Body, &event)
					if err != nil {
						log.Println("Failed to decode message (" + message.Type + "): " + err.Error())
						continue
					}
					ch <- &event

				default:
					ch <- &service.ErrorEvent{
						Raw: string(message.Body),
					}
				}
			}
		}
	}()

	return ch, nil
}
