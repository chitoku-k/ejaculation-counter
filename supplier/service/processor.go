package service

import (
	"context"

	"github.com/sirupsen/logrus"
)

type processor struct {
	Scheduler Scheduler
	Streaming Streaming
	Writer    QueueWriter
}

type Processor interface {
	Execute(ctx context.Context, scheduler <-chan Tick, stream <-chan Status)
}

func NewProcessor(writer QueueWriter) Processor {
	return &processor{
		Writer: writer,
	}
}

func (ps *processor) Execute(ctx context.Context, scheduler <-chan Tick, stream <-chan Status) {
	for scheduler != nil && stream != nil {
		select {
		case tick, ok := <-scheduler:
			if !ok {
				scheduler = nil
				continue
			}
			err := ps.Writer.Publish(ctx, tick)
			if err != nil {
				logrus.Errorf("Error in queueing: %v", err)
			}

		case status, ok := <-stream:
			if !ok {
				stream = nil
				continue
			}
			switch status := status.(type) {
			case Error:
				logrus.Errorf("Error in streaming: %v", status.Err)

			case Connection:
				logrus.Infof("Connected to streaming: %v", status.Server)

			case Disconnection:
				logrus.Infof("Disconnected from streaming: %v", status.Err)

			case Reconnection:
				logrus.Infof("Reconnecting to streaming in %v...", status.In)

			case Message:
				err := ps.Writer.Publish(ctx, status)
				if err != nil {
					logrus.Errorf("Error in publishing: %v", err)
				}
			}
		}
	}
}
