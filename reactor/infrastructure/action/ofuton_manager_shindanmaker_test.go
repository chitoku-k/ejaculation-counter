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

var _ = Describe("OfutonManagerShindanmaker", func() {
	var (
		ctrl                      *gomock.Controller
		c                         *client.MockShindanmaker
		ofutonManagerShindanmaker service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = client.NewMockShindanmaker(ctrl)
		ofutonManagerShindanmaker = action.NewOfutonManagerShindanmaker(c)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := ofutonManagerShindanmaker.Name()
			Expect(actual).To(Equal("おふとん管理官の毎日"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := ofutonManagerShindanmaker.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message does not match pattern", func() {
				It("returns false", func() {
					actual := ofutonManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "診断して",
					})
					Expect(actual).To(BeFalse())
				})
			})

			Context("message matches おふとんしても良い？", func() {
				It("returns true", func() {
					actual := ofutonManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "おふとんしても良い？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches おふとんしてもよい？", func() {
				It("returns true", func() {
					actual := ofutonManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "おふとんしてもよい？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches おふとんしてもいい？", func() {
				It("returns true", func() {
					actual := ofutonManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "おふとんしてもいい？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches おふとんしていい？", func() {
				It("returns true", func() {
					actual := ofutonManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "おふとんしていい？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ふとんしていい？", func() {
				It("returns true", func() {
					actual := ofutonManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ふとんしていい？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ふとん入っていい？", func() {
				It("returns true", func() {
					actual := ofutonManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ふとん入っていい？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ふとんはいっていい？", func() {
				It("returns true", func() {
					actual := ofutonManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ふとんはいっていい？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ふとんいっていい？", func() {
				It("returns true", func() {
					actual := ofutonManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ふとんいっていい？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ふとん行っていい？", func() {
				It("returns true", func() {
					actual := ofutonManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ふとん行っていい？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ふとん潜っていい？", func() {
				It("returns true", func() {
					actual := ofutonManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ふとん潜っていい？",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches ふとんもぐっていい？", func() {
				It("returns true", func() {
					actual := ofutonManagerShindanmaker.Target(service.Message{
						IsReblog: false,
						Content:  "ふとんもぐっていい？",
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
				c.EXPECT().Do("テスト", "https://shindanmaker.com/a/503598").Return(
					"",
					errors.New(`failed to fetch shindan result: Get "https://shindanmaker.com/a/503598": dial tcp [::1]:443: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				_, index, err := ofutonManagerShindanmaker.Event(service.Message{
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content: "おふとんしても良い？",
				})
				Expect(index).To(Equal(3))
				Expect(err).To(MatchError(`failed to create event: failed to fetch shindan result: Get "https://shindanmaker.com/a/503598": dial tcp [::1]:443: connect: connection refused`))
			})
		})

		Context("fetching succeeds", func() {
			BeforeEach(func() {
				c.EXPECT().Do("テスト", "https://shindanmaker.com/a/503598").Return(
					`むむ……このおちんちんしゅっしゅは…………よしよし、ちゃんと申請してありますね♥えらいえらい♥ もっとぴゅっぴゅってさせたげる♥ あは、おちんちんびくびくしちゃってる♥
あっ、今日のぴゅっぴゅは…………ちょっと！おちんちんぴゅっぴゅの許可は下りてないじゃない！ぴゅっぴゅするの駄目っ！おちんちんやめなさいっ！！
むむ……今日のおちんちんしこしこは…………こらっ！おちんちんぴゅっぴゅの許可は下りてないじゃない！こらーっ！おちんちんしこしこするなっ！ぴゅっぴゅしちゃ駄目でしょ！！
んっ？このおちんちんしゅっしゅは…………こらっ！おちんちんぴゅっぴゅ申請してないじゃないっ！おちんちんピクピクさせちゃ駄目っ！いじるの禁止です！！
#shindanmaker
https://shindanmaker.com/503598`,
					nil,
				)
			})

			It("returns an event", func() {
				event, index, err := ofutonManagerShindanmaker.Event(service.Message{
					ID:       "1",
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content:    "おふとんしても良い？",
					Visibility: "private",
				})
				Expect(event).To(Equal(&service.ReplyEvent{
					InReplyToID: "1",
					Acct:        "@test",
					Body: `むむ……このおふとんもふもふは…………よしよし、ちゃんと申請してありますね♥えらいえらい♥ もっともふもふってさせたげる♥ あは、おふとんびくびくしちゃってる♥
あっ、今日のおふとんは…………ちょっと！おふとんおふとんの許可は下りてないじゃない！おふとんするの駄目っ！おふとんやめなさいっ！！
むむ……今日のおふとんもふもふは…………こらっ！おふとんおふとんの許可は下りてないじゃない！こらーっ！おふとんもふもふするなっ！おふとんしちゃ駄目でしょ！！
んっ？このおふとんもふもふは…………こらっ！おふとんおふとん申請してないじゃないっ！おふとんピクピクさせちゃ駄目っ！おふとん禁止です！！
#shindanmaker
https://shindanmaker.com/503598`,
					Visibility: "private",
				}))
				Expect(index).To(Equal(3))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
