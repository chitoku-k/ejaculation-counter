package action_test

import (
	"errors"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PyuppyuManagerShindanmaker", func() {
	var (
		ctrl                       *gomock.Controller
		c                          *client.MockShindanmaker
		pyuppyuManagerShindanmaker service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = client.NewMockShindanmaker(ctrl)
		pyuppyuManagerShindanmaker = action.NewPyuppyuManagerShindanmaker(c)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := pyuppyuManagerShindanmaker.Name()
			Expect(actual).To(Equal("おちんちんぴゅっぴゅ管理官の毎日"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := pyuppyuManagerShindanmaker.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message does not match pattern", func() {
				It("returns false", func() {
					actual := pyuppyuManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "診断して",
					})
					Expect(actual).To(BeFalse())
				})
			})

			Context("message matches ぴゅっぴゅしても良い？", func() {
				It("returns true", func() {
					actual := pyuppyuManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ぴゅっぴゅしても良い？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ぴゅっぴゅしてもよい？", func() {
				It("returns true", func() {
					actual := pyuppyuManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ぴゅっぴゅしてもよい？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ぴゅっぴゅしてもいい？", func() {
				It("returns true", func() {
					actual := pyuppyuManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ぴゅっぴゅしてもいい？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ぴゅっぴゅしていい？", func() {
				It("returns true", func() {
					actual := pyuppyuManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ぴゅっぴゅしていい？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ぴゅっぴゅしていい?", func() {
				It("returns true", func() {
					actual := pyuppyuManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ぴゅっぴゅしていい?",
					})
					Expect(actual).To(BeTrue())
				})
			})
		})
	})

	Describe("Event()", func() {
		BeforeEach(func() {
			c.EXPECT().Name(service.Account{
				DisplayName: "テスト",
				Acct:        "@test",
			}).Return("テスト")
		})

		Context("fetching fails", func() {
			BeforeEach(func() {
				c.EXPECT().Do("テスト", "https://shindanmaker.com/a/503598").Return(
					"",
					errors.New(`failed to fetch shindan result: Get "https://shindanmaker.com/a/503598": dial tcp [::1]:443: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				_, index, err := pyuppyuManagerShindanmaker.Event(service.Message{
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content: "ぴゅっぴゅしていい？",
				})
				Expect(index).To(Equal(0))
				Expect(err).To(MatchError(`failed to create event: failed to fetch shindan result: Get "https://shindanmaker.com/a/503598": dial tcp [::1]:443: connect: connection refused`))
			})
		})

		Context("fetching succeeds", func() {
			BeforeEach(func() {
				c.EXPECT().Do("テスト", "https://shindanmaker.com/a/503598").Return(
					"診断結果",
					nil,
				)
			})

			It("returns an event", func() {
				event, index, err := pyuppyuManagerShindanmaker.Event(service.Message{
					ID:       "1",
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content:    "ぴゅっぴゅしていい？",
					Visibility: "private",
				})
				Expect(event).To(Equal(&service.ReplyEvent{
					InReplyToID: "1",
					Acct:        "@test",
					Body:        "診断結果",
					Visibility:  "private",
				}))
				Expect(index).To(Equal(0))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
