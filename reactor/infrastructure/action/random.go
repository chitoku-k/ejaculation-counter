//go:generate mockgen -source=random.go -destination=random_mock.go -package=action -self_package=github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action

package action

type Random interface {
	IntN(n int) int
}
