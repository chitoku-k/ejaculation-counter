//go:generate mockgen -source=random.go -destination=random_mock.go -package=wrapper -self_package=github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper

package wrapper

type Random interface {
	Intn(n int) int
}
