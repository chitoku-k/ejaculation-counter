package action_test

import (
	"context"
	"io"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/repository"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Through", func() {
	var (
		ctrl    *gomock.Controller
		repo    *repository.MockThroughRepository
		env     config.Environment
		through service.Action
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		repo = repository.NewMockThroughRepository(ctrl)
		env = config.Environment{
			Port: "80",
			Mastodon: config.Mastodon{
				UserID: "1",
			},
		}
		through = action.NewThrough(repo, env)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Name()", func() {
		It("returns the name", func() {
			actual := through.Name()
			Expect(actual).To(Equal("駿河茶"))
		})
	})

	Describe("Target()", func() {
		Context("message is reblog", func() {
			It("returns false", func() {
				actual := through.Target(service.Message{
					IsReblog: true,
				})
				Expect(actual).To(BeFalse())
			})
		})

		Context("message is not reblog", func() {
			Context("message is a reply from the admin", func() {
				It("returns false", func() {
					actual := through.Target(service.Message{
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
						actual := through.Target(service.Message{
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

				Context("message matches 駿河茶", func() {
					It("returns true", func() {
						actual := through.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "駿河茶",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches 今日の through", func() {
					It("returns true", func() {
						actual := through.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "今日の through",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches through ガチャ", func() {
					It("returns true", func() {
						actual := through.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "through ガチャ",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches throughガチャ", func() {
					It("returns true", func() {
						actual := through.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "throughガチャ",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches throughｶﾞﾁｬ", func() {
					It("returns true", func() {
						actual := through.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "throughｶﾞﾁｬ",
						})
						Expect(actual).To(BeTrue())
					})
				})

				Context("message matches 10 連駿河茶", func() {
					It("returns true", func() {
						actual := through.Target(service.Message{
							IsReblog:    false,
							InReplyToID: "",
							Account: service.Account{
								ID: "1",
							},
							Content: "10 連駿河茶",
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
					event, index, err := through.Event(context.Background(), service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "テスト。10 連駿河茶。",
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
					event, index, err := through.Event(context.Background(), service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "10 連駿河茶",
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
					event, index, err := through.Event(context.Background(), service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "テスト。10 連駿河茶。",
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
					event, index, err := through.Event(context.Background(), service.Message{
						ID:       "1",
						IsReblog: false,
						Account: service.Account{
							DisplayName: "テスト",
							Acct:        "@test",
						},
						Content:    "駿河茶",
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
