package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type processor struct {
	ctx            context.Context
	Queue          QueueReader
	Reply          Reply
	Increment      Increment
	Update         Update
	Administration Administration
}

type Processor interface {
	Execute() error
}

func NewProcessor(
	ctx context.Context,
	queue QueueReader,
	reply Reply,
	increment Increment,
	update Update,
	administration Administration,
) Processor {
	return &processor{
		ctx:            ctx,
		Queue:          queue,
		Reply:          reply,
		Increment:      increment,
		Update:         update,
		Administration: administration,
	}
}

func (ps *processor) Execute() error {
	ch, err := ps.Queue.Consume()
	if err != nil {
		return errors.Wrap(err, "failed to read from queue")
	}

	go func() {
		for {
			select {
			case <-ps.ctx.Done():
				return

			case event := <-ch:
				switch e := event.(type) {
				case *ReplyEvent:
					err = ps.Reply.Send(*e)
					if err != nil {
						logrus.Errorln("Failed to send reply: " + err.Error())
						continue
					}

				case *IncrementEvent:
					err = ps.Increment.Do(*e)
					if err != nil {
						logrus.Errorln("Failed to update increment: " + err.Error())
						continue
					}

				case *UpdateEvent:
					err = ps.Update.Do(*e)
					if err != nil {
						logrus.Errorln("Failed to update: " + err.Error())
						continue
					}

				case *AdministrationEvent:
					err = ps.Administration.Do(*e)
					if err != nil {
						logrus.Errorln("Failed to execute administrative operation: " + err.Error())
						continue
					}

				case *ErrorEvent:
					logrus.Errorln("ErrorEvent: " + e.Raw)
				}
			}
		}
	}()

	return nil
}
