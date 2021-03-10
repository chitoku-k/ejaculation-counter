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

var _ = Describe("Mpyw", func() {
	var (
		ctrl *gomock.Controller
		c    *client.MockMpyw
		mpyw service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = client.NewMockMpyw(ctrl)
		mpyw = action.NewMpyw(c)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := mpyw.Name()
			Expect(actual).To(Equal("実務経験ガチャ"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := mpyw.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message does not match pattern", func() {
				It("returns false", func() {
					actual := mpyw.Target(service.Message{
						IsReblog: false,
						Content:  "診断して",
					})
					Expect(actual).To(BeFalse())
				})
			})

			Context("message matches mpyw ガチャ", func() {
				It("returns true", func() {
					actual := mpyw.Target(service.Message{
						IsReblog: false,
						Content:  "mpyw ガチャ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches mpywガチャ", func() {
				It("returns true", func() {
					actual := mpyw.Target(service.Message{
						IsReblog: false,
						Content:  "mpywガチャ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches まっぴーガチャ", func() {
				It("returns true", func() {
					actual := mpyw.Target(service.Message{
						IsReblog: false,
						Content:  "まっぴーガチャ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches 実務経験ガチャ", func() {
				It("returns true", func() {
					actual := mpyw.Target(service.Message{
						IsReblog: false,
						Content:  "実務経験ガチャ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches mpyw 10 連ガチャ", func() {
				It("returns true", func() {
					actual := mpyw.Target(service.Message{
						IsReblog: false,
						Content:  "mpyw 10 連ガチャ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches まっぴー 10 連ガチャ", func() {
				It("returns true", func() {
					actual := mpyw.Target(service.Message{
						IsReblog: false,
						Content:  "まっぴー 10 連ガチャ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches 実務経験 10 連ガチャ", func() {
				It("returns true", func() {
					actual := mpyw.Target(service.Message{
						IsReblog: false,
						Content:  "実務経験 10 連ガチャ",
					})
					Expect(actual).To(BeTrue())
				})
			})
		})
	})

	Describe("Event()", func() {
		Context("fetching fails", func() {
			BeforeEach(func() {
				c.EXPECT().Do("https://mpyw.hinanawi.net/api", 1).Return(
					client.MpywChallengeResult{},
					errors.New(`failed to fetch challenge result: Get "https://mpyw.hinanawi.net/api": dial tcp [::1]:443: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				_, index, err := mpyw.Event(service.Message{
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content: "実務経験ガチャ",
				})
				Expect(index).To(Equal(0))
				Expect(err).To(MatchError(`failed to create event: failed to fetch challenge result: Get "https://mpyw.hinanawi.net/api": dial tcp [::1]:443: connect: connection refused`))

			})
		})

		Context("fetching succeeds", func() {
			Context("with count", func() {
				BeforeEach(func() {
					c.EXPECT().Do("https://mpyw.hinanawi.net/api", 10).Return(
						client.MpywChallengeResult{
							Title:  "診断結果",
							Result: []string{"診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果"},
						},
						nil,
					)
				})

				Context("toot does not start with name", func() {
					It("returns an event", func() {
						event, index, err := mpyw.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								DisplayName: "テスト",
								Acct:        "@test",
							},
							Content:    "テスト。実務経験 10 連ガチャ。",
							Visibility: "private",
						})
						Expect(event).To(Equal(&service.ReplyEvent{
							InReplyToID: "1",
							Acct:        "@test",
							Body:        "診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果",
							Visibility:  "private",
						}))
						Expect(index).To(Equal(12))
						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("toot starts with name", func() {
					It("returns an event", func() {
						event, index, err := mpyw.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								DisplayName: "テスト",
								Acct:        "@test",
							},
							Content:    "実務経験 10 連ガチャ",
							Visibility: "private",
						})
						Expect(event).To(Equal(&service.ReplyEvent{
							InReplyToID: "1",
							Acct:        "@test",
							Body:        "診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果",
							Visibility:  "private",
						}))
						Expect(index).To(Equal(0))
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})

			Context("without count", func() {
				BeforeEach(func() {
					c.EXPECT().Do("https://mpyw.hinanawi.net/api", 1).Return(
						client.MpywChallengeResult{
							Title:  "診断結果",
							Result: []string{"診断結果"},
						},
						nil,
					)
				})

				Context("toot does not start with name", func() {
					It("returns an event", func() {
						event, index, err := mpyw.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								DisplayName: "テスト",
								Acct:        "@test",
							},
							Content:    "テスト。実務経験ガチャ。",
							Visibility: "private",
						})
						Expect(event).To(Equal(&service.ReplyEvent{
							InReplyToID: "1",
							Acct:        "@test",
							Body:        "診断結果",
							Visibility:  "private",
						}))
						Expect(index).To(Equal(12))
						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("toot starts with name", func() {
					It("returns an event", func() {
						event, index, err := mpyw.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								DisplayName: "テスト",
								Acct:        "@test",
							},
							Content:    "実務経験ガチャ",
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
	})
})
