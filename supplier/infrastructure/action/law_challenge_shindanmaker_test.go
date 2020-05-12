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

var _ = Describe("LawChallengeShindanmaker", func() {
	var (
		ctrl                     *gomock.Controller
		c                        *client.MockShindanmaker
		lawChallengeShindanmaker service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = client.NewMockShindanmaker(ctrl)
		lawChallengeShindanmaker = action.NewLawChallengeShindanmaker(c)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := lawChallengeShindanmaker.Name()
			Expect(actual).To(Equal("法律ギリギリチャレンジ"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := lawChallengeShindanmaker.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message contains shindanmaker tag", func() {
				It("returns false", func() {
					actual := lawChallengeShindanmaker.Target(service.Message{
						IsReblog: false,
						Tags: []service.Tag{
							{
								Name: "法律ギリギリチャレンジ",
							},
						},
						Content: "法律ギリギリチャレンジ",
					})
					Expect(actual).To(BeFalse())
				})
			})

			Context("message does not contain shindanmaker tag", func() {
				Context("message does not match pattern", func() {
					It("returns false", func() {
						actual := lawChallengeShindanmaker.Target(service.Message{
							IsReblog: false,
							Content:  "診断して",
						})
						Expect(actual).To(BeFalse())
					})
				})

				Context("message matches 法律ギリギリチャレンジ", func() {
					It("returns true", func() {
						actual := lawChallengeShindanmaker.Target(service.Message{
							IsReblog: false,
							Content:  "法律ギリギリチャレンジ",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches 法律ｷﾞﾘｷﾞﾘﾁｬﾚﾝｼﾞ", func() {
					It("returns true", func() {
						actual := lawChallengeShindanmaker.Target(service.Message{
							IsReblog: false,
							Content:  "法律ｷﾞﾘｷﾞﾘﾁｬﾚﾝｼﾞ",
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
				c.EXPECT().Do("テスト", "https://shindanmaker.com/a/877845").Return(
					"",
					errors.New(`failed to fetch shindan result: Get "https://shindanmaker.com/a/877845": dial tcp [::1]:443: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				_, index, err := lawChallengeShindanmaker.Event(service.Message{
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content: "法律ギリギリチャレンジ",
				})
				Expect(index).To(Equal(0))
				Expect(err).To(MatchError(`failed to create event: failed to fetch shindan result: Get "https://shindanmaker.com/a/877845": dial tcp [::1]:443: connect: connection refused`))
			})
		})

		Context("fetching succeeds", func() {
			BeforeEach(func() {
				c.EXPECT().Do("テスト", "https://shindanmaker.com/a/877845").Return(
					"診断結果",
					nil,
				)
			})

			Context("toot does not start with name", func() {
				It("returns an event", func() {
					event, index, err := lawChallengeShindanmaker.Event(service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "テスト。法律ギリギリチャレンジ。",
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
					event, index, err := lawChallengeShindanmaker.Event(service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "法律ギリギリチャレンジ",
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
