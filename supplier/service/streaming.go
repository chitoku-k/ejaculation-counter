package service

import "context"

type Streaming interface {
	Run(ctx context.Context) (<-chan Status, error)
}
