package action_test

import (
	"context"
	"io"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/repository"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Doublet", func() {
	var (
		ctrl           *gomock.Controller
		repo           *repository.MockDoubletRepository
		mastodonUserID string
		doublet        service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		repo = repository.NewMockDoubletRepository(ctrl)
		mastodonUserID = "1"
		doublet = action.NewDoublet(repo, mastodonUserID)
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
			Context("message is a reply from the admin", func() {
				It("returns false", func() {
					actual := doublet.Target(service.Message{
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
						actual := doublet.Target(service.Message{
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

				Context("message matches 今日の doublet", func() {
					It("returns true", func() {
						actual := doublet.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "今日の doublet",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches 今日の 二重語", func() {
					It("returns true", func() {
						actual := doublet.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "今日の 二重語",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches doublet ガチャ", func() {
					It("returns true", func() {
						actual := doublet.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "doublet ガチャ",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches 二重語 ガチャ", func() {
					It("returns true", func() {
						actual := doublet.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "二重語 ガチャ",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches doubletガチャ", func() {
					It("returns true", func() {
						actual := doublet.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "doubletガチャ",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches 二重語ガチャ", func() {
					It("returns true", func() {
						actual := doublet.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "二重語ガチャ",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches doubletｶﾞﾁｬ", func() {
					It("returns true", func() {
						actual := doublet.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "doubletｶﾞﾁｬ",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches 二重語ｶﾞﾁｬ", func() {
					It("returns true", func() {
						actual := doublet.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "二重語ｶﾞﾁｬ",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches 10 連 doublet ガチャ", func() {
					It("returns true", func() {
						actual := doublet.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "10 連 doublet ガチャ",
						})
						Expect(actual).To(BeTrue())
					})
				})
			})
		})
	})

	Describe("Event()", func() {
		Context("with count", func() {
			BeforeEach(func() {
				repo.EXPECT().Get().Return(
					[]string{"診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果"},
				)
			})

			Context("toot does not start with name", func() {
				It("returns an event", func() {
					event, index, err := doublet.Event(context.Background(), service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "テスト。10 連二重語ガチャ。",
						Visibility: "private",
					})
					Expect(event).To(ReplyEventEqual(service.ReplyEvent{
						InReplyToID: "1",
						Acct:        "@test",
						Body:        io.NopCloser(strings.NewReader("診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果")),
						Visibility:  "private",
					}))
					Expect(index).To(Equal(12))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("toot starts with name", func() {
				It("returns an event", func() {
					event, index, err := doublet.Event(context.Background(), service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "10 連二重語ガチャ",
						Visibility: "private",
					})
					Expect(event).To(ReplyEventEqual(service.ReplyEvent{
						InReplyToID: "1",
						Acct:        "@test",
						Body:        io.NopCloser(strings.NewReader("診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果")),
						Visibility:  "private",
					}))
					Expect(index).To(Equal(0))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("without count", func() {
			BeforeEach(func() {
				repo.EXPECT().Get().Return(
					[]string{"診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果", "診断結果"},
				)
			})

			Context("toot does not start with name", func() {
				It("returns an event", func() {
					event, index, err := doublet.Event(context.Background(), service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "テスト。10 連二重語ガチャ。",
						Visibility: "private",
					})
					Expect(event).To(ReplyEventEqual(service.ReplyEvent{
						InReplyToID: "1",
						Acct:        "@test",
						Body:        io.NopCloser(strings.NewReader("診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果\n診断結果")),
						Visibility:  "private",
					}))
					Expect(index).To(Equal(12))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("toot starts with name", func() {
				It("returns an event", func() {
					event, index, err := doublet.Event(context.Background(), service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "二重語ガチャ",
						Visibility: "private",
					})
					Expect(event).To(ReplyEventEqual(service.ReplyEvent{
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
