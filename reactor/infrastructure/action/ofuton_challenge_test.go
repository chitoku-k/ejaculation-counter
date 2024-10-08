package action_test

import (
	"context"
	"io"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("OfutonChallenge", func() {
	var (
		ctrl            *gomock.Controller
		r               *action.MockRandom
		mastodonUserID  string
		ofutonChallenge service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		r = action.NewMockRandom(ctrl)
		mastodonUserID = "1"
		ofutonChallenge = action.NewOfufutonChallenge(r, mastodonUserID)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := ofutonChallenge.Name()
			Expect(actual).To(Equal("おふとんチャレンジ"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := ofutonChallenge.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message is a reply from the admin", func() {
				It("returns false", func() {
					actual := ofutonChallenge.Target(service.Message{
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
						actual := ofutonChallenge.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Tags: []service.Tag{
								{
									Name: "おふとんチャレンジ",
								},
							},
							Content: "おふとんチャレンジ",
						})
						Expect(actual).To(BeFalse())
					})
				})

				Context("message does not contain shindanmaker tag", func() {
					Context("message does not match pattern", func() {
						It("returns false", func() {
							actual := ofutonChallenge.Target(service.Message{
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

					Context("message matches ふとんチャレンジ", func() {
						It("returns true", func() {
							actual := ofutonChallenge.Target(service.Message{
								IsReblog:    false,
								InReplyToID: "",
								Account: service.Account{
									ID: "1",
								},
								Content: "ふとんチャレンジ",
							})
							Expect(actual).To(BeTrue())
						})
					})

					Context("message matches ふとんﾁｬﾚﾝｼﾞ", func() {
						It("returns true", func() {
							actual := ofutonChallenge.Target(service.Message{
								IsReblog:    false,
								InReplyToID: "",
								Account: service.Account{
									ID: "1",
								},
								Content: "ふとんﾁｬﾚﾝｼﾞ",
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
			gomock.InOrder(
				r.EXPECT().IntN(4).Return(2),
				r.EXPECT().IntN(4).Return(0),
				r.EXPECT().IntN(4).Return(1),
				r.EXPECT().IntN(4).Return(0),
			)
		})

		It("returns an event", func() {
			event, index, err := ofutonChallenge.Event(context.Background(), service.Message{
				ID:       "1",
				IsReblog: false,
				Account: service.Account{
					DisplayName: "テスト",
					Acct:        "@test",
				},
				Content:    "おふとんチャレンジ",
				Visibility: "private",
			})
			Expect(event).To(ReplyEventEqual(service.ReplyEvent{
				InReplyToID: "1",
				Acct:        "@test",
				Body:        io.NopCloser(strings.NewReader("とおふお\n#おふとんチャレンジ")),
				Visibility:  "private",
			}))
			Expect(index).To(Equal(3))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
