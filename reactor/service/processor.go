package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type processor struct {
	Queue          QueueReader
	Reply          Reply
	Increment      Increment
	Update         Update
	Administration Administration
}

type Processor interface {
	Execute(ctx context.Context) error
}

func NewProcessor(
	queue QueueReader,
	reply Reply,
	increment Increment,
	update Update,
	administration Administration,
) Processor {
	return &processor{
		Queue:          queue,
		Reply:          reply,
		Increment:      increment,
		Update:         update,
		Administration: administration,
	}
}

func (ps *processor) Execute(ctx context.Context) error {
	ch, err := ps.Queue.Consume(ctx)
	if err != nil {
		return fmt.Errorf("failed to read from queue: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case event := <-ch:
				switch e := event.(type) {
				case *ReplyEvent:
					err = ps.Reply.Send(ctx, *e)
					if err != nil {
						logrus.Errorf("Failed to send reply: %v", err)
						continue
					}

				case *ReplyErrorEvent:
					err = ps.Reply.SendError(ctx, *e)
					if err != nil {
						logrus.Errorf("Failed to send reply (error): %v", err)
						continue
					}

				case *IncrementEvent:
					err = ps.Increment.Do(ctx, *e)
					if err != nil {
						logrus.Errorf("Failed to update increment: %v", err)
						continue
					}

				case *UpdateEvent:
					err = ps.Update.Do(ctx, *e)
					if err != nil {
						logrus.Errorf("Failed to update: %v", err)
						continue
					}

				case *AdministrationEvent:
					err = ps.Administration.Do(ctx, *e)
					if err != nil {
						logrus.Errorf("Failed to execute administrative operation: %v", err)
						continue
					}

				case *ErrorEvent:
					logrus.Errorf("ErrorEvent: %v", e.Raw)
				}
			}
		}
	}()

	return nil
}
