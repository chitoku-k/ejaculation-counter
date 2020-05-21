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

var _ = Describe("SushiShindanmaker", func() {
	var (
		ctrl              *gomock.Controller
		c                 *client.MockShindanmaker
		sushiShindanmaker service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = client.NewMockShindanmaker(ctrl)
		sushiShindanmaker = action.NewSushiShindanmaker(c)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := sushiShindanmaker.Name()
			Expect(actual).To(Equal("寿司職人"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := sushiShindanmaker.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message does not contain sushi emoji", func() {
				It("returns true", func() {
					actual := sushiShindanmaker.Target(service.Message{
						IsReblog: false,
						Emojis: []service.Emoji{
							{Shortcode: "naruhodo"},
							{Shortcode: "atamawarui"},
						},
						Content: "診断して",
					})
					Expect(actual).To(BeFalse())
				})
			})

			Context("message contains thinking_sushi emoji", func() {
				It("returns true", func() {
					actual := sushiShindanmaker.Target(service.Message{
						IsReblog: false,
						Emojis: []service.Emoji{
							{Shortcode: "thinking_sushi"},
						},
						Content: "診断して",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message contains ios_big_sushi_* emoji", func() {
				It("returns true", func() {
					actual := sushiShindanmaker.Target(service.Message{
						IsReblog: false,
						Emojis: []service.Emoji{
							{Shortcode: "ios_big_sushi_1"},
							{Shortcode: "ios_big_sushi_2"},
							{Shortcode: "ios_big_sushi_3"},
							{Shortcode: "ios_big_sushi_4"},
						},
						Content: "診断して",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message contains *_*_sushi emoji", func() {
				It("returns true", func() {
					actual := sushiShindanmaker.Target(service.Message{
						IsReblog: false,
						Emojis: []service.Emoji{
							{Shortcode: "top_left_sushi"},
							{Shortcode: "top_center_sushi"},
							{Shortcode: "top_right_sushi"},
							{Shortcode: "middle_left_sushi"},
							{Shortcode: "middle_right_sushi"},
							{Shortcode: "bottom_left_sushi"},
							{Shortcode: "bottom_center_sushi"},
							{Shortcode: "bottom_right_sushi"},
						},
						Content: "診断して",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches すしにぎ", func() {
				It("returns true", func() {
					actual := sushiShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "すしにぎ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches すし握", func() {
				It("returns true", func() {
					actual := sushiShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "すし握",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches 寿司握", func() {
				It("returns true", func() {
					actual := sushiShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "寿司握",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ちんちん握", func() {
				It("returns true", func() {
					actual := sushiShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ちんちん握",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ちんぽ握", func() {
				It("returns true", func() {
					actual := sushiShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ちんぽ握",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ちんこ握", func() {
				It("returns true", func() {
					actual := sushiShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ちんこ握",
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
				c.EXPECT().Do("テスト", "https://shindanmaker.com/a/577901").Return(
					"",
					errors.New(`failed to fetch shindan result: Get "https://shindanmaker.com/a/577901": dial tcp [::1]:443: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				_, index, err := sushiShindanmaker.Event(service.Message{
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content: "ちんぽチャレンジ",
				})
				Expect(index).To(Equal(0))
				Expect(err).To(MatchError(`failed to create event: failed to fetch shindan result: Get "https://shindanmaker.com/a/577901": dial tcp [::1]:443: connect: connection refused`))
			})
		})

		Context("fetching succeeds", func() {
			BeforeEach(func() {
				c.EXPECT().Do("テスト", "https://shindanmaker.com/a/577901").Return(
					"診断結果",
					nil,
				)
			})

			Context("toot matches with emoji", func() {
				It("returns an event", func() {
					event, index, err := sushiShindanmaker.Event(service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Emojis: []service.Emoji{
							{Shortcode: "top_left_sushi"},
							{Shortcode: "top_center_sushi"},
							{Shortcode: "top_right_sushi"},
							{Shortcode: "middle_left_sushi"},
							{Shortcode: "middle_right_sushi"},
							{Shortcode: "bottom_left_sushi"},
							{Shortcode: "bottom_center_sushi"},
							{Shortcode: "bottom_right_sushi"},
						},
						Content:    "診断して",
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

			Context("toot does not start with name", func() {
				It("returns an event", func() {
					event, index, err := sushiShindanmaker.Event(service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "テスト。寿司握。",
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
					event, index, err := sushiShindanmaker.Event(service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "寿司握",
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
