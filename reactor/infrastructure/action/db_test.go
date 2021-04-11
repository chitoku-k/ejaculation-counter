package action_test

import (
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DB", func() {
	var (
		ctrl *gomock.Controller
		env  config.Environment
		db   service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		env = config.Environment{
			Mastodon: config.Mastodon{
				UserID: "1",
			},
		}
		db = action.NewDB(env)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := db.Name()
			Expect(actual).To(Equal("DB"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := db.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message is not from admin", func() {
				It("returns false", func() {
					actual := db.Target(service.Message{
						IsReblog: false,
						Account: service.Account{
							ID: "2",
						},
						Content: "SQL: SELECT 1",
					})
					Expect(actual).To(BeFalse())
				})
			})

			Context("message is from admin", func() {
				Context("message does not match pattern", func() {
					It("returns false", func() {
						actual := db.Target(service.Message{
							IsReblog: false,
							Account: service.Account{
								ID: "1",
							},
							Content: "SELECT 1",
						})
						Expect(actual).To(BeFalse())
					})
				})

				Context("message matches pattern", func() {
					It("returns false", func() {
						actual := db.Target(service.Message{
							IsReblog: false,
							Account: service.Account{
								ID: "1",
							},
							Content: "SQL: SELECT 1",
						})
						Expect(actual).To(BeTrue())
					})
				})
			})
		})
	})

	Describe("Event()", func() {
		It("returns an event", func() {
			actual, index, err := db.Event(service.Message{
				ID: "1",
				Account: service.Account{
					Acct: "@test",
				},
				Content: "SQL: SELECT 1",
			})
			Expect(actual).To(Equal(&service.AdministrationEvent{
				InReplyToID: "1",
				Acct:        "@test",
				Type:        "DB",
				Command:     "SELECT 1",
			}))
			Expect(index).To(Equal(0))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
