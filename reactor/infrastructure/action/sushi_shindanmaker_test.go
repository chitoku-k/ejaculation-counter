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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("SushiShindanmaker", func() {
	var (
		ctrl              *gomock.Controller
		c                 *client.MockShindanmaker
		env               config.Environment
		sushiShindanmaker service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = client.NewMockShindanmaker(ctrl)
		env = config.Environment{
			Mastodon: config.Mastodon{
				UserID: "1",
			},
		}
		sushiShindanmaker = action.NewSushiShindanmaker(c, env)
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
			Context("message is a reply from the admin", func() {
				It("returns false", func() {
					actual := sushiShindanmaker.Target(service.Message{
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
				Context("message does not contain sushi emoji", func() {
					It("returns true", func() {
						actual := sushiShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
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
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
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
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
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
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
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
							Content: "診断して",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches すしにぎ", func() {
					It("returns true", func() {
						actual := sushiShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "すしにぎ",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches すし握", func() {
					It("returns true", func() {
						actual := sushiShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "すし握",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches 寿司握", func() {
					It("returns true", func() {
						actual := sushiShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "寿司握",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches ちんちん握", func() {
					It("returns true", func() {
						actual := sushiShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "ちんちん握",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches ちんぽ握", func() {
					It("returns true", func() {
						actual := sushiShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "ちんぽ握",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches ちんこ握", func() {
					It("returns true", func() {
						actual := sushiShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "ちんこ握",
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
				c.EXPECT().Do(context.Background(), "テスト", "https://shindanmaker.com/a/577901").Return(
					"",
					errors.New(`failed to fetch shindan result: Get "https://shindanmaker.com/a/577901": dial tcp [::1]:443: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				_, index, err := sushiShindanmaker.Event(context.Background(), service.Message{
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
				c.EXPECT().Do(context.Background(), "テスト", "https://shindanmaker.com/a/577901").Return(
					"診断結果",
					nil,
				)
			})

			Context("toot matches with emoji", func() {
				It("returns an event", func() {
					event, index, err := sushiShindanmaker.Event(context.Background(), service.Message{
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

			Context("toot does not start with name", func() {
				It("returns an event", func() {
					event, index, err := sushiShindanmaker.Event(context.Background(), service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "テスト。寿司握。",
						Visibility: "private",
					})
					Expect(event).To(Equal(service.ReplyEvent{
						InReplyToID: "1",
						Acct:        "@test",
						Body:        io.NopCloser(strings.NewReader("診断結果")),
						Visibility:  "private",
					}))
					Expect(index).To(Equal(12))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("toot starts with name", func() {
				It("returns an event", func() {
					event, index, err := sushiShindanmaker.Event(context.Background(), service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "寿司握",
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
})
