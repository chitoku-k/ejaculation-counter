//go:generate mockgen -source=scheduler.go -destination=scheduler_mock.go -package=service -self_package=github.com/chitoku-k/ejaculation-counter/supplier/service

package service

type Scheduler interface {
	Start() <-chan Tick
	Stop()
}
