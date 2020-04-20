package service

type queue struct {
	Writer QueueWriter
}

type Queue interface {
	Write(event Event) error
}

type QueueWriter interface {
	Publish(event Event) error
}

func NewQueue(writer QueueWriter) Queue {
	return &queue{
		Writer: writer,
	}
}

func (qs *queue) Write(event Event) error {
	return qs.Writer.Publish(event)
}
