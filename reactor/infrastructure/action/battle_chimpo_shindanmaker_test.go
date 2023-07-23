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
	"go.uber.org/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BattleChimpoShindanmaker", func() {
	var (
		ctrl                     *gomock.Controller
		c                        *client.MockShindanmaker
		env                      config.Environment
		battleChimpoShindanmaker service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = client.NewMockShindanmaker(ctrl)
		env = config.Environment{
			Mastodon: config.Mastodon{
				UserID: "1",
			},
		}
		battleChimpoShindanmaker = action.NewBattleChimpoShindanmaker(c, env)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := battleChimpoShindanmaker.Name()
			Expect(actual).To(Equal("絶対おちんぽなんかに負けない！"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := battleChimpoShindanmaker.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message is a reply from the admin", func() {
				It("returns false", func() {
					actual := battleChimpoShindanmaker.Target(service.Message{
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
						actual := battleChimpoShindanmaker.Target(service.Message{
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

				Context("message does not contain なんかに", func() {
					Context("message matches 勝てない", func() {
						Context("message matches おちんぽに勝てない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "おちんぽに勝てない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches おちんちんに勝てない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "おちんちんに勝てない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches おちんこに勝てない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "おちんこに勝てない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches ちんぽに勝てない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "ちんぽに勝てない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches ちんちんに勝てない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "ちんちんに勝てない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches ちんこに勝てない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "ちんこに勝てない",
								})
								Expect(actual).To(BeTrue())
							})
						})
					})

					Context("message matches 負けない", func() {
						Context("message matches おちんぽに負けない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "おちんぽに負けない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches おちんちんに負けない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "おちんちんに負けない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches おちんこに負けない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "おちんこに負けない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches ちんぽに負けない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "ちんぽに負けない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches ちんちんに負けない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "ちんちんに負けない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches ちんこに負けない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "ちんこに負けない",
								})
								Expect(actual).To(BeTrue())
							})
						})
					})
				})

				Context("message contains なんかに", func() {
					Context("message matches 勝てない", func() {
						Context("message matches おちんぽなんかに勝てない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "おちんぽなんかに勝てない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches おちんちんなんかに勝てない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "おちんちんなんかに勝てない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches おちんこなんかに勝てない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "おちんこなんかに勝てない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches ちんぽなんかに勝てない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "ちんぽなんかに勝てない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches ちんちんなんかに勝てない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "ちんちんなんかに勝てない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches ちんこなんかに勝てない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "ちんこなんかに勝てない",
								})
								Expect(actual).To(BeTrue())
							})
						})
					})

					Context("message matches 負けない", func() {
						Context("message matches おちんぽなんかに負けない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "おちんぽなんかに負けない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches おちんちんなんかに負けない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "おちんちんなんかに負けない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches おちんこなんかに負けない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "おちんこなんかに負けない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches ちんぽなんかに負けない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "ちんぽなんかに負けない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches ちんちんなんかに負けない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "ちんちんなんかに負けない",
								})
								Expect(actual).To(BeTrue())
							})
						})

						Context("message matches ちんこなんかに負けない", func() {
							It("returns true", func() {
								actual := battleChimpoShindanmaker.Target(service.Message{
									IsReblog:    false,
									InReplyToID: "",
									Account: service.Account{
										ID: "1",
									},
									Content: "ちんこなんかに負けない",
								})
								Expect(actual).To(BeTrue())
							})
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
				c.EXPECT().Do(context.Background(), "テスト", "https://shindanmaker.com/a/584238").Return(
					"",
					errors.New(`failed to fetch shindan result: Get "https://shindanmaker.com/a/584238": dial tcp [::1]:443: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				_, index, err := battleChimpoShindanmaker.Event(context.Background(), service.Message{
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content: "絶対おちんぽなんかに負けない！",
				})
				Expect(index).To(Equal(9))
				Expect(err).To(MatchError(`failed to create event: failed to fetch shindan result: Get "https://shindanmaker.com/a/584238": dial tcp [::1]:443: connect: connection refused`))
			})
		})

		Context("fetching succeeds", func() {
			BeforeEach(func() {
				c.EXPECT().Do(context.Background(), "テスト", "https://shindanmaker.com/a/584238").Return(
					"診断結果",
					nil,
				)
			})

			It("returns an event", func() {
				event, index, err := battleChimpoShindanmaker.Event(context.Background(), service.Message{
					ID:       "1",
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content:    "絶対おちんぽなんかに負けない！",
					Visibility: "private",
				})
				Expect(event).To(Equal(service.ReplyEvent{
					InReplyToID: "1",
					Acct:        "@test",
					Body:        io.NopCloser(strings.NewReader("診断結果")),
					Visibility:  "private",
				}))
				Expect(index).To(Equal(9))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
