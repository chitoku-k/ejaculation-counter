//go:generate mockgen -source=streaming.go -destination=streaming_mock.go -package=service -self_package=github.com/chitoku-k/ejaculation-counter/supplier/service

package service

import "context"

type Streaming interface {
	Run(ctx context.Context) error
	Statuses() <-chan Status
	Close(exit bool) error
}
