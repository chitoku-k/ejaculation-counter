package queue

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueSize         = 1024
	ConnectionTimeout = 30 * time.Second
	ReconnectNone     = 0 * time.Second
	ReconnectInitial  = 5 * time.Second
	ReconnectMax      = 320 * time.Second

	DeadLetterSuffix = ".dl"
	DeadLetterTTL    = 5 * time.Minute
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
	ch          chan service.Packet
	Exchange    string
	QueueName   string
	RoutingKey  string
	Environment config.Environment
	Connection  *amqp.Connection
	Channel     *amqp.Channel
	Delivery    <-chan amqp.Delivery
	Closes      chan *amqp.Error
}

func NewReader(
	exchange string,
	queueName string,
	routingKey string,
	environment config.Environment,
) (service.QueueReader, error) {
	r := &reader{
		ch:          make(chan service.Packet, QueueSize),
		Exchange:    exchange,
		QueueName:   queueName,
		RoutingKey:  routingKey,
		Environment: environment,
	}

	return r, r.connect()
}

func (r *reader) dial(url string) (*amqp.Connection, net.Conn, error) {
	tlsConfig := &tls.Config{}

	var sasl []amqp.Authentication
	if r.Environment.Queue.Username == "" && r.Environment.Queue.Password == "" {
		sasl = append(sasl, &amqp.ExternalAuth{})
	}

	if r.Environment.Queue.SSLCert != "" && r.Environment.Queue.SSLKey != "" {
		cert, err := tls.LoadX509KeyPair(r.Environment.Queue.SSLCert, r.Environment.Queue.SSLKey)
		if err != nil {
			return nil, nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if r.Environment.Queue.SSLRootCert != "" {
		ca, err := os.ReadFile(r.Environment.Queue.SSLRootCert)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read CA file for queue: %w", err)
		}

		tlsConfig.RootCAs = x509.NewCertPool()
		tlsConfig.RootCAs.AppendCertsFromPEM(ca)
	}

	var nc net.Conn
	conn, err := amqp.DialConfig(url, amqp.Config{
		TLSClientConfig: tlsConfig,
		SASL:            sasl,
		Dial: func(network, addr string) (net.Conn, error) {
			var err error
			nc, err = amqp.DefaultDial(ConnectionTimeout)(network, addr)
			return nc, err
		},
	})

	return conn, nc, err
}

func (r *reader) connect() error {
	slog.Debug("Connecting to MQ broker...")

	uri, err := amqp.ParseURI(r.Environment.Queue.Host)
	if err != nil {
		return fmt.Errorf("failed to parse MQ URI: %w", err)
	}

	uri.Username = r.Environment.Queue.Username
	uri.Password = r.Environment.Queue.Password

	var nc net.Conn
	r.Connection, nc, err = r.dial(uri.String())
	if err != nil {
		return fmt.Errorf("failed to connect to MQ broker: %w", err)
	}

	r.Channel, err = r.Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel for MQ connection: %w", err)
	}

	r.Closes = r.Connection.NotifyClose(make(chan *amqp.Error, 1))

	slog.Debug("Declaring queues in MQ...")

	q, err := r.Channel.QueueDeclare(
		r.QueueName,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-queue-type":              "quorum",
			"x-dead-letter-exchange":    r.Exchange + DeadLetterSuffix,
			"x-dead-letter-routing-key": r.RoutingKey,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue in MQ channel: %w", err)
	}

	dlq, err := r.Channel.QueueDeclare(
		r.QueueName+DeadLetterSuffix,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-queue-type":              "quorum",
			"x-dead-letter-exchange":    r.Exchange,
			"x-dead-letter-routing-key": r.RoutingKey,
			"x-message-ttl":             DeadLetterTTL.Milliseconds(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue for dead letters in MQ channel: %w", err)
	}

	slog.Debug("Binding queues in MQ...")

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

	err = r.Channel.QueueBind(
		dlq.Name,
		r.RoutingKey,
		r.Exchange+DeadLetterSuffix,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue for dead letters in MQ channel: %w", err)
	}

	slog.Debug("Consuming from MQ...")

	r.Delivery, err = r.Channel.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume from MQ: %w", err)
	}

	slog.Info("Connected to MQ", slog.Any("remote", nc.RemoteAddr()))
	return nil
}

func (r *reader) disconnect() error {
	err := r.Channel.Close()
	if err != nil {
		return fmt.Errorf("failed to close the MQ channel: %w", err)
	}

	err = r.Connection.Close()
	if err != nil {
		return fmt.Errorf("failed to close the MQ connection: %w", err)
	}

	return nil
}

func (r *reader) reconnect(ctx context.Context) error {
	reconnect := ReconnectNone

	for {
		_ = r.disconnect()
		err := r.connect()
		if err == nil {
			return nil
		}

		slog.Error("Error from MQ", slog.Any("err", err))

		reconnect = time.Duration(
			min(
				max(
					reconnect*2,
					ReconnectInitial,
				),
				ReconnectMax,
			),
		)

		slog.Info(fmt.Sprintf("Reconnecting in %v...", reconnect))

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-time.After(reconnect):
			continue
		}
	}
}

func (r *reader) Ack(tag uint64) error {
	return r.Channel.Ack(tag, false)
}

func (r *reader) Reject(tag uint64) error {
	return r.Channel.Reject(tag, false)
}

func (r *reader) Packets() <-chan service.Packet {
	return r.ch
}

func (r *reader) Close(exit bool) error {
	if exit {
		close(r.ch)
		r.ch = nil
	}
	err := r.disconnect()
	if !errors.Is(err, amqp.ErrClosed) {
		return err
	}
	return nil
}

func (r *reader) Consume(ctx context.Context) {
	for r.Closes != nil && r.Delivery != nil {
		select {
		case amqperr, ok := <-r.Closes:
			if !ok {
				r.Closes = nil
				continue
			}
			slog.Info("Disconnected from MQ", slog.Any("err", amqperr))
			err := r.reconnect(ctx)
			if err != nil {
				continue
			}

		case packet, ok := <-r.Delivery:
			if !ok {
				r.Delivery = nil
				continue
			}
			DeliveredMessageTotal.WithLabelValues(packet.Type).Inc()

			switch packet.Type {
			case "packets.tick":
				tick := service.NewTick(packet.DeliveryTag, packet.Timestamp)
				err := json.Unmarshal(packet.Body, &tick)
				if err != nil {
					slog.Error("Failed to decode message", slog.String("packet-type", packet.Type), slog.Any("err", err))
					continue
				}
				r.ch <- tick

			case "packets.message":
				message := service.NewMessage(packet.DeliveryTag, packet.Timestamp)
				err := json.Unmarshal(packet.Body, &message)
				if err != nil {
					slog.Error("Failed to decode message", slog.String("packet-type", packet.Type), slog.Any("err", err))
					continue
				}
				r.ch <- message
			}
		}
	}
}
