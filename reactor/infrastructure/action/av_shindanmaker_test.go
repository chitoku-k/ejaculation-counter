package action_test

import (
	"errors"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AVShindanmaker", func() {
	var (
		ctrl           *gomock.Controller
		c              *client.MockShindanmaker
		avShindanmaker service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = client.NewMockShindanmaker(ctrl)
		avShindanmaker = action.NewAVShindanmaker(c)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := avShindanmaker.Name()
			Expect(actual).To(Equal("同人AVタイトルジェネレーター"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := avShindanmaker.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message contains shindanmaker tag", func() {
				It("returns false", func() {
					actual := avShindanmaker.Target(service.Message{
						IsReblog: false,
						Tags: []service.Tag{
							{
								Name: "同人avタイトルジェネレーター",
							},
						},
						Content: "テストのAV",
					})
					Expect(actual).To(BeFalse())
				})
			})

			Context("message does not contain shindanmaker tag", func() {
				Context("message does not match pattern", func() {
					It("returns false", func() {
						actual := avShindanmaker.Target(service.Message{
							IsReblog: false,
							Content:  "診断して",
						})
						Expect(actual).To(BeFalse())
					})
				})

				Context("message matches without a space before AV", func() {
					It("returns true", func() {
						actual := avShindanmaker.Target(service.Message{
							IsReblog: false,
							Content:  "テストのAV",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches with a space before AV", func() {
					It("returns true", func() {
						actual := avShindanmaker.Target(service.Message{
							IsReblog: false,
							Content:  "テストの AV",
						})
						Expect(actual).To(BeTrue())
					})
				})
			})
		})
	})

	Describe("Event()", func() {
		Context("fetching fails", func() {
			BeforeEach(func() {
				c.EXPECT().Do("テスト", "https://shindanmaker.com/a/794363").Return(
					"",
					errors.New(`failed to fetch shindan result: Get "https://shindanmaker.com/a/794363": dial tcp [::1]:443: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				_, index, err := avShindanmaker.Event(service.Message{
					IsReblog: false,
					Content:  "テストの AV",
				})
				Expect(index).To(Equal(13))
				Expect(err).To(MatchError(`failed to create event: failed to fetch shindan result: Get "https://shindanmaker.com/a/794363": dial tcp [::1]:443: connect: connection refused`))
			})
		})

		Context("fetching succeeds", func() {
			BeforeEach(func() {
				c.EXPECT().Do("テスト", "https://shindanmaker.com/a/794363").Return(
					"診断結果",
					nil,
				)
			})

			Context("toot does not start with name", func() {
				Context("name includes kun", func() {
					It("returns an event", func() {
						event, index, err := avShindanmaker.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								Acct: "@test",
							},
							Content:    "テスト。テストくんの AV",
							Visibility: "private",
						})
						Expect(event).To(Equal(&service.ReplyEvent{
							InReplyToID: "1",
							Acct:        "@test",
							Body:        "診断結果",
							Visibility:  "private",
						}))
						Expect(index).To(Equal(31))
						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("name includes chan", func() {
					It("returns an event", func() {
						event, index, err := avShindanmaker.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								Acct: "@test",
							},
							Content:    "テスト。テストちゃんの AV",
							Visibility: "private",
						})
						Expect(event).To(Equal(&service.ReplyEvent{
							InReplyToID: "1",
							Acct:        "@test",
							Body:        "診断結果",
							Visibility:  "private",
						}))
						Expect(index).To(Equal(34))
						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("name includes neither kun nor chan", func() {
					It("returns an event", func() {
						event, index, err := avShindanmaker.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								Acct: "@test",
							},
							Content:    "テスト。テストの AV",
							Visibility: "private",
						})
						Expect(event).To(Equal(&service.ReplyEvent{
							InReplyToID: "1",
							Acct:        "@test",
							Body:        "診断結果",
							Visibility:  "private",
						}))
						Expect(index).To(Equal(25))
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})

			Context("toot starts with name", func() {
				Context("name includes kun", func() {
					It("returns an event", func() {
						event, index, err := avShindanmaker.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								Acct: "@test",
							},
							Content:    "テストくんの AV",
							Visibility: "private",
						})
						Expect(event).To(Equal(&service.ReplyEvent{
							InReplyToID: "1",
							Acct:        "@test",
							Body:        "診断結果",
							Visibility:  "private",
						}))
						Expect(index).To(Equal(19))
						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("name includes chan", func() {
					It("returns an event", func() {
						event, index, err := avShindanmaker.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								Acct: "@test",
							},
							Content:    "テストちゃんの AV",
							Visibility: "private",
						})
						Expect(event).To(Equal(&service.ReplyEvent{
							InReplyToID: "1",
							Acct:        "@test",
							Body:        "診断結果",
							Visibility:  "private",
						}))
						Expect(index).To(Equal(22))
						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("name includes neither kun nor chan", func() {
					It("returns an event", func() {
						event, index, err := avShindanmaker.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								Acct: "@test",
							},
							Content:    "テストの AV",
							Visibility: "private",
						})
						Expect(event).To(Equal(&service.ReplyEvent{
							InReplyToID: "1",
							Acct:        "@test",
							Body:        "診断結果",
							Visibility:  "private",
						}))
						Expect(index).To(Equal(13))
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})
		})
	})
})
