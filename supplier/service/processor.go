package service

import (
	"context"
	"fmt"
	"log/slog"
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
				slog.Error("Error in queueing", slog.Any("err", err))
			}

		case status, ok := <-stream:
			if !ok {
				stream = nil
				continue
			}
			switch status := status.(type) {
			case Error:
				slog.Error("Error in streaming", slog.Any("err", status.Err))

			case Connection:
				slog.Info("Connected to streaming", slog.String("server", status.Server))

			case Disconnection:
				slog.Info("Disconnected from streaming", slog.Any("err", status.Err))

			case Reconnection:
				slog.Info(fmt.Sprintf("Reconnecting to streaming in %v...", status.In))

			case Message:
				err := ps.Writer.Publish(ctx, status)
				if err != nil {
					slog.Error("Error in publishing", slog.Any("err", err))
				}
			}
		}
	}
}
