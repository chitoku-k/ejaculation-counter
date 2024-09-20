package action_test

import (
	"context"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("DB", func() {
	var (
		ctrl           *gomock.Controller
		mastodonUserID string
		db             service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mastodonUserID = "1"
		db = action.NewDB(mastodonUserID)
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
			actual, index, err := db.Event(context.Background(), service.Message{
				ID: "1",
				Account: service.Account{
					Acct: "@test",
				},
				Content:    "SQL: SELECT 1",
				Visibility: "private",
			})
			Expect(actual).To(Equal(service.AdministrationEvent{
				InReplyToID: "1",
				Acct:        "@test",
				Type:        "DB",
				Command:     "SELECT 1",
				Visibility:  "private",
			}))
			Expect(index).To(Equal(0))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
