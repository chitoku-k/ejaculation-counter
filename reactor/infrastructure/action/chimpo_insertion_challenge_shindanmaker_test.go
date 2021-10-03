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

var _ = Describe("ChimpoInsertionChallengeShindanmaker", func() {
	var (
		ctrl                                 *gomock.Controller
		c                                    *client.MockShindanmaker
		env                                  config.Environment
		chimpoInsertionChallengeShindanmaker service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = client.NewMockShindanmaker(ctrl)
		env = config.Environment{
			Mastodon: config.Mastodon{
				UserID: "1",
			},
		}
		chimpoInsertionChallengeShindanmaker = action.NewChimpoInsertionChallengeShindanmaker(c, env)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := chimpoInsertionChallengeShindanmaker.Name()
			Expect(actual).To(Equal("おちんぽ挿入チャレンジ"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := chimpoInsertionChallengeShindanmaker.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message is a reply from the admin", func() {
				It("returns false", func() {
					actual := chimpoInsertionChallengeShindanmaker.Target(service.Message{
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
				Context("message contains shindanmaker tag", func() {
					It("returns false", func() {
						actual := chimpoInsertionChallengeShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Tags: []service.Tag{
								{
									Name: "おちんぽ挿入チャレンジ",
								},
							},
							Content: "おちんぽ挿入チャレンジ",
						})
						Expect(actual).To(BeFalse())
					})
				})

				Context("message does not contain shindanmaker tag", func() {
					Context("message does not match pattern", func() {
						It("returns false", func() {
							actual := chimpoInsertionChallengeShindanmaker.Target(service.Message{
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

					Context("message matches ちんちん挿入チャレンジ", func() {
						It("returns true", func() {
							actual := chimpoInsertionChallengeShindanmaker.Target(service.Message{
								IsReblog:    false,
								InReplyToID: "",
								Account: service.Account{
									ID: "1",
								},
								Content: "ちんちん挿入チャレンジ",
							})
							Expect(actual).To(BeTrue())
						})
					})

					Context("message matches ちんぽ挿入チャレンジ", func() {
						It("returns true", func() {
							actual := chimpoInsertionChallengeShindanmaker.Target(service.Message{
								IsReblog:    false,
								InReplyToID: "",
								Account: service.Account{
									ID: "1",
								},
								Content: "ちんぽ挿入チャレンジ",
							})
							Expect(actual).To(BeTrue())
						})
					})

					Context("message matches ちんこ挿入チャレンジ", func() {
						It("returns true", func() {
							actual := chimpoInsertionChallengeShindanmaker.Target(service.Message{
								IsReblog:    false,
								InReplyToID: "",
								Account: service.Account{
									ID: "1",
								},
								Content: "ちんこ挿入チャレンジ",
							})
							Expect(actual).To(BeTrue())
						})
					})

					Context("message matches ちんぽ挿入ﾁｬﾚﾝｼﾞ", func() {
						It("returns true", func() {
							actual := chimpoInsertionChallengeShindanmaker.Target(service.Message{
								IsReblog:    false,
								InReplyToID: "",
								Account: service.Account{
									ID: "1",
								},
								Content: "ちんぽ挿入ﾁｬﾚﾝｼﾞ",
							})
							Expect(actual).To(BeTrue())
						})
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
				c.EXPECT().Do(context.Background(), "テスト", "https://shindanmaker.com/a/670773").Return(
					"",
					errors.New(`failed to fetch shindan result: Get "https://shindanmaker.com/a/670773": dial tcp [::1]:443: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				_, index, err := chimpoInsertionChallengeShindanmaker.Event(context.Background(), service.Message{
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content: "おちんぽ挿入チャレンジ",
				})
				Expect(index).To(Equal(3))
				Expect(err).To(MatchError(`failed to create event: failed to fetch shindan result: Get "https://shindanmaker.com/a/670773": dial tcp [::1]:443: connect: connection refused`))
			})
		})

		Context("fetching succeeds", func() {
			BeforeEach(func() {
				c.EXPECT().Do(context.Background(), "テスト", "https://shindanmaker.com/a/670773").Return(
					"診断結果",
					nil,
				)
			})

			Context("toot does not start with name", func() {
				It("returns an event", func() {
					event, index, err := chimpoInsertionChallengeShindanmaker.Event(context.Background(), service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "テスト。おちんぽ挿入チャレンジ。",
						Visibility: "private",
					})
					Expect(event).To(Equal(service.ReplyEvent{
						InReplyToID: "1",
						Acct:        "@test",
						Body:        io.NopCloser(strings.NewReader("診断結果")),
						Visibility:  "private",
					}))
					Expect(index).To(Equal(15))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("toot starts with name", func() {
				It("returns an event", func() {
					event, index, err := chimpoInsertionChallengeShindanmaker.Event(context.Background(), service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "おちんぽ挿入チャレンジ",
						Visibility: "private",
					})
					Expect(event).To(Equal(service.ReplyEvent{
						InReplyToID: "1",
						Acct:        "@test",
						Body:        io.NopCloser(strings.NewReader("診断結果")),
						Visibility:  "private",
					}))
					Expect(index).To(Equal(3))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
