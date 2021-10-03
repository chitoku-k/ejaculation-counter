package action_test

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PyuppyuManagerShindanmaker", func() {
	var (
		ctrl                       *gomock.Controller
		c                          *client.MockShindanmaker
		env                        config.Environment
		pyuppyuManagerShindanmaker service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = client.NewMockShindanmaker(ctrl)
		env = config.Environment{
			Mastodon: config.Mastodon{
				UserID: "1",
			},
		}
		pyuppyuManagerShindanmaker = action.NewPyuppyuManagerShindanmaker(c, env)
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
			Context("message is a reply from the admin", func() {
				It("returns false", func() {
					actual := pyuppyuManagerShindanmaker.Target(service.Message{
						IsReblog:    false,
						InReplyToID: "1",
						Account: service.Account{
							ID: "1",
						},
					})
					Expect(actual).To(BeFalse())
				})
			})

			Context("message is not a reply from the admin", func() {
				Context("message does not match pattern", func() {
					It("returns false", func() {
						actual := pyuppyuManagerShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "診断して",
						})
						Expect(actual).To(BeFalse())
					})
				})

				Context("message matches ぴゅっぴゅしても良い？", func() {
					It("returns true", func() {
						actual := pyuppyuManagerShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "ぴゅっぴゅしても良い？",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches ぴゅっぴゅしてもよい？", func() {
					It("returns true", func() {
						actual := pyuppyuManagerShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "ぴゅっぴゅしてもよい？",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches ぴゅっぴゅしてもいい？", func() {
					It("returns true", func() {
						actual := pyuppyuManagerShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "ぴゅっぴゅしてもいい？",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches ぴゅっぴゅしていい？", func() {
					It("returns true", func() {
						actual := pyuppyuManagerShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "ぴゅっぴゅしていい？",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches ぴゅっぴゅしていい?", func() {
					It("returns true", func() {
						actual := pyuppyuManagerShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "ぴゅっぴゅしていい?",
						})
						Expect(actual).To(BeTrue())
					})
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
				c.EXPECT().Do(context.Background(), "テスト", "https://shindanmaker.com/a/503598").Return(
					"",
					errors.New(`failed to fetch shindan result: Get "https://shindanmaker.com/a/503598": dial tcp [::1]:443: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				_, index, err := pyuppyuManagerShindanmaker.Event(context.Background(), service.Message{
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
				c.EXPECT().Do(context.Background(), "テスト", "https://shindanmaker.com/a/503598").Return(
					"診断結果",
					nil,
				)
			})

			It("returns an event", func() {
				event, index, err := pyuppyuManagerShindanmaker.Event(context.Background(), service.Message{
					ID:       "1",
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content:    "ぴゅっぴゅしていい？",
					Visibility: "private",
				})
				Expect(event).To(Equal(service.ReplyEvent{
					InReplyToID: "1",
					Acct:        "@test",
					Body:        io.NopCloser(strings.NewReader("診断結果")),
					Visibility:  "private",
				}))
				Expect(index).To(Equal(0))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
