package action_test

import (
	"context"
	"time"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var JST = time.FixedZone("JST", int(9*time.Hour.Seconds()))

var _ = Describe("PyuUpdate", func() {
	var (
		ctrl      *gomock.Controller
		env       config.Environment
		pyuUpdate service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		env = config.Environment{
			Mastodon: config.Mastodon{
				UserID: "1",
			},
		}
		pyuUpdate = action.NewPyuUpdate(env)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := pyuUpdate.Name()
			Expect(actual).To(Equal("ぴゅっ♡"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := pyuUpdate.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message is not mime", func() {
				It("returns false", func() {
					actual := pyuUpdate.Target(service.Message{
						IsReblog: false,
						Account: service.Account{
							ID: "2",
						},
						Content: "ぴゅっ♡",
					})
					Expect(actual).To(BeFalse())
				})
			})

			Context("message is mime", func() {
				Context("message does not match pattern", func() {
					It("returns false", func() {
						actual := pyuUpdate.Target(service.Message{
							IsReblog: false,
							Account: service.Account{
								ID: "1",
							},
							Content: "ぴゅっ！",
						})
						Expect(actual).To(BeFalse())
					})
				})

				Context("message matches pattern", func() {
					It("returns false", func() {
						actual := pyuUpdate.Target(service.Message{
							IsReblog: false,
							Account: service.Account{
								ID: "1",
							},
							Content: "ぴゅっ♡",
						})
						Expect(actual).To(BeTrue())
					})
				})
			})
		})
	})

	Describe("Event()", func() {
		It("returns an event", func() {
			actual, index, err := pyuUpdate.Event(context.Background(), service.Message{
				ID:        "1",
				CreatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, JST),
				Account: service.Account{
					ID:   "1",
					Acct: "@test",
				},
				Content: "ぴゅっ♡",
			})
			Expect(actual).To(Equal(&service.IncrementEvent{
				Year:  2006,
				Month: 1,
				Day:   2,
			}))
			Expect(index).To(Equal(0))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
