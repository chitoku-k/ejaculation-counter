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

var _ = Describe("Doublet", func() {
	var (
		ctrl    *gomock.Controller
		c       *client.MockDoublet
		doublet service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = client.NewMockDoublet(ctrl)
		doublet = action.NewDoublet(c)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := doublet.Name()
			Expect(actual).To(Equal("二重語ガチャ"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := doublet.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message does not match pattern", func() {
				It("returns false", func() {
					actual := doublet.Target(service.Message{
						IsReblog: false,
						Content:  "診断して",
					})
					Expect(actual).To(BeFalse())
				})
			})

			Context("message matches 今日の doublet", func() {
				It("returns true", func() {
					actual := doublet.Target(service.Message{
						IsReblog: false,
						Content:  "今日の doublet",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches 今日の 二重語", func() {
				It("returns true", func() {
					actual := doublet.Target(service.Message{
						IsReblog: false,
						Content:  "今日の 二重語",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches doublet ガチャ", func() {
				It("returns true", func() {
					actual := doublet.Target(service.Message{
						IsReblog: false,
						Content:  "doublet ガチャ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches 二重語 ガチャ", func() {
				It("returns true", func() {
					actual := doublet.Target(service.Message{
						IsReblog: false,
						Content:  "二重語 ガチャ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches doubletガチャ", func() {
				It("returns true", func() {
					actual := doublet.Target(service.Message{
						IsReblog: false,
						Content:  "doubletガチャ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches 二重語ガチャ", func() {
				It("returns true", func() {
					actual := doublet.Target(service.Message{
						IsReblog: false,
						Content:  "二重語ガチャ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches doubletｶﾞﾁｬ", func() {
				It("returns true", func() {
					actual := doublet.Target(service.Message{
						IsReblog: false,
						Content:  "doubletｶﾞﾁｬ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches 二重語ｶﾞﾁｬ", func() {
				It("returns true", func() {
					actual := doublet.Target(service.Message{
						IsReblog: false,
						Content:  "二重語ｶﾞﾁｬ",
					})
					Expect(actual).To(BeTrue())
				})
			})

			Context("message matches 10 連 doublet ガチャ", func() {
				It("returns true", func() {
					actual := doublet.Target(service.Message{
						IsReblog: false,
						Content:  "10 連 doublet ガチャ",
					})
					Expect(actual).To(BeTrue())
				})
			})
		})
	})

	Describe("Event()", func() {
		Context("fetching fails", func() {
			BeforeEach(func() {
				c.EXPECT().Do("http://localhost/doublet").Return(
					client.DoubletResult{},
					errors.New(`failed to fetch challenge result: Get "http://localhost/doublet": dial tcp [::1]:80: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				_, index, err := doublet.Event(service.Message{
					IsReblog: false,
					Account: service.Account{
						DisplayName: "テスト",
						Acct:        "@test",
					},
					Content: "二重語ガチャ",
				})
				Expect(index).To(Equal(0))
				Expect(err).To(MatchError(`failed to create event: failed to fetch challenge result: Get "http://localhost/doublet": dial tcp [::1]:80: connect: connection refused`))
			})
		})

		Context("fetching succeeds", func() {
			Context("with count", func() {
				BeforeEach(func() {
					c.EXPECT().Do("http://localhost/doublet").Return(
						client.DoubletResult{"診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果"},
						nil,
					)
				})

				Context("toot does not start with name", func() {
					It("returns an event", func() {
						event, index, err := doublet.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								DisplayName: "テスト",
								Acct:        "@test",
							},
							Content:    "テスト。10 連二重語ガチャ。",
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
						event, index, err := doublet.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								DisplayName: "テスト",
								Acct:        "@test",
							},
							Content:    "10 連二重語ガチャ",
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
					c.EXPECT().Do("http://localhost/doublet").Return(
						client.DoubletResult{"診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果"},
						nil,
					)
				})

				Context("toot does not start with name", func() {
					It("returns an event", func() {
						event, index, err := doublet.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								DisplayName: "テスト",
								Acct:        "@test",
							},
							Content:    "テスト。10 連二重語ガチャ。",
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
						event, index, err := doublet.Event(service.Message{
							ID:       "1",
							IsReblog: false,
							Account: service.Account{
								DisplayName: "テスト",
								Acct:        "@test",
							},
							Content:    "二重語ガチャ",
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
