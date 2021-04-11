package action_test

import (
	"errors"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ChimpoChallengeShindanmaker", func() {
	var (
		ctrl                        *gomock.Controller
		c                           *client.MockShindanmaker
		env                         config.Environment
		chimpoChallengeShindanmaker service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		env = config.Environment{
			Mastodon: config.Mastodon{
				UserID: "1",
			},
		}
		c = client.NewMockShindanmaker(ctrl)
		chimpoChallengeShindanmaker = action.NewChimpoChallengeShindanmaker(c, env)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := chimpoChallengeShindanmaker.Name()
			Expect(actual).To(Equal("ちんぽチャレンジ"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := chimpoChallengeShindanmaker.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message is a reply from the admin", func() {
				It("returns false", func() {
					actual := chimpoChallengeShindanmaker.Target(service.Message{
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
						actual := chimpoChallengeShindanmaker.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Tags: []service.Tag{
								{
									Name: "ちんぽチャレンジ",
								},
							},
							Content: "ちんぽチャレンジ",
						})
						Expect(actual).To(BeFalse())
					})
				})

				Context("message does not contain shindanmaker tag", func() {
					Context("message does not match pattern", func() {
						It("returns false", func() {
							actual := chimpoChallengeShindanmaker.Target(service.Message{
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

					Context("message matches ちんちんチャレンジ", func() {
						It("returns true", func() {
							actual := chimpoChallengeShindanmaker.Target(service.Message{
								IsReblog:    false,
								InReplyToID: "",
								Account: service.Account{
									ID: "1",
								},
								Content: "ちんちんチャレンジ",
							})
							Expect(actual).To(BeTrue())
						})
					})

					Context("message matches ちんぽチャレンジ", func() {
						It("returns true", func() {
							actual := chimpoChallengeShindanmaker.Target(service.Message{
								IsReblog:    false,
								InReplyToID: "",
								Account: service.Account{
									ID: "1",
								},
								Content: "ちんぽチャレンジ",
							})
							Expect(actual).To(BeTrue())
						})
					})

					Context("message matches ちんこチャレンジ", func() {
						It("returns true", func() {
							actual := chimpoChallengeShindanmaker.Target(service.Message{
								IsReblog:    false,
								InReplyToID: "",
								Account: service.Account{
									ID: "1",
								},
								Content: "ちんこチャレンジ",
							})
							Expect(actual).To(BeTrue())
						})
					})

					Context("message matches ちんぽﾁｬﾚﾝｼﾞ", func() {
						It("returns true", func() {
							actual := chimpoChallengeShindanmaker.Target(service.Message{
								IsReblog:    false,
								InReplyToID: "",
								Account: service.Account{
									ID: "1",
								},
								Content: "ちんぽﾁｬﾚﾝｼﾞ",
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
				c.EXPECT().Do("テスト", "https://shindanmaker.com/a/656461").Return(
					"",
					errors.New(`failed to fetch shindan result: Get "https://shindanmaker.com/a/656461": dial tcp [::1]:443: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				_, index, err := chimpoChallengeShindanmaker.Event(service.Message{
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content: "ちんぽチャレンジ",
				})
				Expect(index).To(Equal(0))
				Expect(err).To(MatchError(`failed to create event: failed to fetch shindan result: Get "https://shindanmaker.com/a/656461": dial tcp [::1]:443: connect: connection refused`))
			})
		})

		Context("fetching succeeds", func() {
			BeforeEach(func() {
				c.EXPECT().Do("テスト", "https://shindanmaker.com/a/656461").Return(
					"診断結果",
					nil,
				)
			})

			Context("toot does not start with name", func() {
				It("returns an event", func() {
					event, index, err := chimpoChallengeShindanmaker.Event(service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "テスト。ちんぽチャレンジ。",
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
					event, index, err := chimpoChallengeShindanmaker.Event(service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "ちんぽチャレンジ",
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
