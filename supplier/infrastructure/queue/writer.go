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

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
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

	CacheSize = 1024
	CacheTTL  = 1 * time.Minute
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
	QueueName     string
	RoutingKey    string
	Environment   config.Environment
	Connection    *amqp.Connection
	Channel       *amqp.Channel
	Confirmations chan amqp.Confirmation
	Closes        chan *amqp.Error
	Queue         chan service.Packet
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
		Queue:       make(chan service.Packet, QueueSize),
	}

	return w, w.connect(ctx)
}

func (w *writer) dial(url string) (*amqp.Connection, net.Conn, error) {
	tlsConfig := &tls.Config{}

	var sasl []amqp.Authentication
	if w.Environment.Queue.Username == "" && w.Environment.Queue.Password == "" {
		sasl = append(sasl, &amqp.ExternalAuth{})
	}

	if w.Environment.Queue.SSLCert != "" && w.Environment.Queue.SSLKey != "" {
		cert, err := tls.LoadX509KeyPair(w.Environment.Queue.SSLCert, w.Environment.Queue.SSLKey)
		if err != nil {
			return nil, nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if w.Environment.Queue.SSLRootCert != "" {
		ca, err := os.ReadFile(w.Environment.Queue.SSLRootCert)
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

func (w *writer) connect(ctx context.Context) error {
	slog.Debug("Connecting to MQ broker...")

	uri, err := amqp.ParseURI(w.Environment.Queue.Host)
	if err != nil {
		return fmt.Errorf("failed to parse MQ URI: %w", err)
	}

	uri.Username = w.Environment.Queue.Username
	uri.Password = w.Environment.Queue.Password

	var nc net.Conn
	w.Connection, nc, err = w.dial(uri.String())
	if err != nil {
		return fmt.Errorf("failed to connect to MQ broker: %w", err)
	}

	w.Channel, err = w.Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel for MQ connection: %w", err)
	}

	w.Closes = w.Connection.NotifyClose(make(chan *amqp.Error, 1))
	w.Confirmations = w.Channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	slog.Debug("Declaring exchanges in MQ...")

	err = w.Channel.ExchangeDeclare(
		w.Exchange,
		"x-message-deduplication",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-cache-size": CacheSize,
			"x-cache-ttl":  CacheTTL.Milliseconds(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange in MQ channel: %w", err)
	}

	err = w.Channel.ExchangeDeclare(
		w.Exchange+DeadLetterSuffix,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange for dead letters in MQ channel: %w", err)
	}

	err = w.Channel.Confirm(false)
	if err != nil {
		return fmt.Errorf("failed to put channel into confirm mode: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case err := <-w.Closes:
				slog.Info("Disconnected from MQ", slog.Any("err", err))
				w.reconnect(ctx)
				return

			case packet := <-w.Queue:
				err := w.Publish(ctx, packet)
				if err != nil {
					slog.Error("Error in publishing from queue", slog.Any("err", err))
				}
			}
		}
	}()

	slog.Info("Connected to MQ", slog.Any("remote", nc.RemoteAddr()))
	return nil
}

func (w *writer) disconnect() error {
	err := w.Channel.Close()
	if err != nil {
		return fmt.Errorf("failed to close the MQ channel: %w", err)
	}
	err = w.Connection.Close()
	if err != nil {
		return fmt.Errorf("failed to close the MQ connection: %w", err)
	}
	return nil
}

func (w *writer) reconnect(ctx context.Context) {
	reconnect := ReconnectNone

	for {
		_ = w.disconnect()
		err := w.connect(ctx)
		if err == nil {
			return
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
			return

		case <-time.After(reconnect):
			continue
		}
	}
}

func (w *writer) Close() error {
	err := w.disconnect()
	if !errors.Is(err, amqp.ErrClosed) {
		return err
	}
	return nil
}

func (w *writer) Publish(ctx context.Context, packet service.Packet) error {
	body, err := json.Marshal(packet)
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %w", err)
	}

	err = w.Channel.PublishWithContext(
		ctx,
		w.Exchange,
		w.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Timestamp:   packet.Timestamp(),
			Type:        packet.Name(),
			Headers: amqp.Table{
				"x-deduplication-header": fmt.Sprintf("%v-%v", packet.Name(), packet.HashCode()),
			},
			Body: body,
		},
	)
	<-w.Confirmations
	QueuedMessageTotal.Inc()

	if err != nil {
		QueuedMessageErrorTotal.Inc()
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed to publish message (%w): %w", ctx.Err(), err)

		case w.Queue <- packet:
			return fmt.Errorf("failed to publish message (requeued): %w", err)

		default:
			return fmt.Errorf("failed to publish message (queue is full): %w", err)
		}
	}

	return nil
}
