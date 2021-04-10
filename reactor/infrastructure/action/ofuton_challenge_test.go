package action_test

import (
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/wrapper"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OfutonChallenge", func() {
	var (
		ctrl            *gomock.Controller
		r               *wrapper.MockRandom
		ofutonChallenge service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		r = wrapper.NewMockRandom(ctrl)
		ofutonChallenge = action.NewOfufutonChallenge(r)
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
			Context("message contains shindanmaker tag", func() {
				It("returns false", func() {
					actual := ofutonChallenge.Target(service.Message{
						IsReblog: false,
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
							IsReblog: false,
							Content:  "診断して",
						})
						Expect(actual).To(BeFalse())
					})
				})

				Context("message matches ふとんチャレンジ", func() {
					It("returns true", func() {
						actual := ofutonChallenge.Target(service.Message{
							IsReblog: false,
							Content:  "ふとんチャレンジ",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches ふとんﾁｬﾚﾝｼﾞ", func() {
					It("returns true", func() {
						actual := ofutonChallenge.Target(service.Message{
							IsReblog: false,
							Content:  "ふとんﾁｬﾚﾝｼﾞ",
						})
						Expect(actual).To(BeTrue())
					})
				})
			})
		})
	})

	Describe("Event()", func() {
		BeforeEach(func() {
			gomock.InOrder(
				r.EXPECT().Intn(4).Return(2),
				r.EXPECT().Intn(4).Return(0),
				r.EXPECT().Intn(4).Return(1),
				r.EXPECT().Intn(4).Return(0),
			)
		})

		It("returns an event", func() {
			event, index, err := ofutonChallenge.Event(service.Message{
				ID:       "1",
				IsReblog: false,
				Account: service.Account{
					DisplayName: "テスト",
					Acct:        "@test",
				},
				Content:    "おふとんチャレンジ",
				Visibility: "private",
			})
			Expect(event).To(Equal(&service.ReplyEvent{
				InReplyToID: "1",
				Acct:        "@test",
				Body:        "とおふお\n#おふとんチャレンジ",
				Visibility:  "private",
			}))
			Expect(index).To(Equal(3))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
