//go:generate mockgen -source=doublet.go -destination=doublet_mock.go -package=repository -self_package=github.com/chitoku-k/ejaculation-counter/reactor/repository

package repository

type DoubletRepository interface {
	Get() []string
}
