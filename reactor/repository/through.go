//go:generate go tool mockgen -source=through.go -destination=through_mock.go -package=repository -self_package=github.com/chitoku-k/ejaculation-counter/reactor/repository

package repository

type ThroughRepository interface {
	Get() []string
}
