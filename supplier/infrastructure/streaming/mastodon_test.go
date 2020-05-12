package streaming_test

import (
	"errors"
	"net/http"
	"time"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/streaming"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/golang/mock/gomock"
	mast "github.com/mattn/go-mastodon"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mastodon", func() {
	var (
		ctrl *gomock.Controller
		conn *wrapper.MockConn
		ctx  *wrapper.MockContext
		d    *wrapper.MockDialer
		t    *wrapper.MockTicker
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		conn = wrapper.NewMockConn(ctrl)
		ctx = wrapper.NewMockContext(ctrl)
		d = wrapper.NewMockDialer(ctrl)
		t = wrapper.NewMockTicker(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Run()", func() {
		Context("parsing server URL fails", func() {
			var (
				env      config.Environment
				mastodon service.Streaming
			)

			BeforeEach(func() {
				env = config.Environment{
					Mastodon: config.Mastodon{
						ServerURL:   ":/",
						AccessToken: "token",
					},
				}
				mastodon = streaming.NewMastodon(ctx, env, d, t)
			})

			It("returns an error", func() {
				actual, err := mastodon.Run()
				Expect(actual).To(BeClosed())
				Expect(err).To(MatchError(`failed to parse server URL: parse ":/": missing protocol scheme`))
			})
		})

		Context("parsing server URL succeeds", func() {
			var (
				env      config.Environment
				mastodon service.Streaming
			)

			BeforeEach(func() {
				env = config.Environment{
					Mastodon: config.Mastodon{
						ServerURL:   "https://mastodon.example.com",
						Stream:      "user",
						AccessToken: "token",
					},
				}
				mastodon = streaming.NewMastodon(ctx, env, d, t)
			})

			Context("context is done", func() {
				BeforeEach(func() {
					ctx.EXPECT().Err().Return(errors.New("closed"))
				})

				It("succeeds and eventually exits", func() {
					actual, err := mastodon.Run()
					Expect(err).NotTo(HaveOccurred())
					Eventually(actual).Should(BeClosed())
				})
			})

			Context("context is still", func() {
				var (
					ch chan time.Time
				)

				Context("websocket connection fails", func() {
					BeforeEach(func() {
						ch = make(chan time.Time, 1)
						ch <- time.Time{}

						gomock.InOrder(
							ctx.EXPECT().Err().Return(nil),
							ctx.EXPECT().Err().Return(errors.New("closed")),
						)

						t.EXPECT().Tick(5 * time.Second).Return(ch)

						d.EXPECT().DialContext(
							ctx,
							"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
							nil,
						).Return(
							nil,
							&http.Response{},
							errors.New("dial tcp [::1]:443: connect: connection refused"),
						)
					})

					AfterEach(func() {
						close(ch)
					})

					It("returns channel and eventually exits", func() {
						actual, err := mastodon.Run()
						Expect(err).NotTo(HaveOccurred())

						Eventually(actual).Should(Receive(Equal(service.MessageStatus{
							Error: errors.New("dial tcp [::1]:443: connect: connection refused"),
						})))
						Eventually(actual).Should(BeClosed())
					})
				})

				Context("websocket connection succeeds", func() {
					BeforeEach(func() {
						d.EXPECT().DialContext(
							ctx,
							"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
							nil,
						).Return(
							conn,
							&http.Response{
								Header: http.Header{
									"X-Served-Server": []string{"192.0.2.1:4000"},
								},
							},
							nil,
						)
					})

					Context("context is done", func() {
						BeforeEach(func() {
							gomock.InOrder(
								ctx.EXPECT().Err().Return(nil),
								ctx.EXPECT().Err().Return(errors.New("closed")),
							)
							conn.EXPECT().Close().Return(nil)
						})

						It("returns channel and eventually exits", func() {
							actual, err := mastodon.Run()
							Expect(err).NotTo(HaveOccurred())
							Eventually(actual).Should(BeClosed())
						})
					})

					Context("context is still", func() {
						Context("event cannot be read", func() {
							Context("connection cannot be closed", func() {
								BeforeEach(func() {
									gomock.InOrder(
										ctx.EXPECT().Err().Return(nil),
										ctx.EXPECT().Err().Return(nil),
										ctx.EXPECT().Err().Return(errors.New("closed")),
									)

									conn.EXPECT().ReadJSON(gomock.Any()).Return(errors.New("dial tcp [::1]:443: connect: connection refused"))
									conn.EXPECT().Close().Return(nil)
								})

								It("sends error and eventually exits", func() {
									actual, err := mastodon.Run()
									Expect(err).NotTo(HaveOccurred())

									Eventually(actual).Should(Receive(Equal(service.MessageStatus{
										Error: errors.New("dial tcp [::1]:443: connect: connection refused"),
									})))
									Eventually(actual).Should(BeClosed())
								})
							})

							Context("connection can be closed", func() {
								BeforeEach(func() {
									gomock.InOrder(
										ctx.EXPECT().Err().Return(nil),
										ctx.EXPECT().Err().Return(nil),
										ctx.EXPECT().Err().Return(errors.New("closed")),
									)

									conn.EXPECT().ReadJSON(gomock.Any()).Return(errors.New("dial tcp [::1]:443: connect: connection refused"))
									conn.EXPECT().Close().Return(errors.New("connection cannot be closed"))
								})

								It("sends error and eventually exits", func() {
									actual, err := mastodon.Run()
									Expect(err).NotTo(HaveOccurred())

									Eventually(actual).Should(Receive(Equal(service.MessageStatus{
										Error: errors.New("dial tcp [::1]:443: connect: connection refused"),
									})))
									Eventually(actual).Should(Receive(Equal(service.MessageStatus{
										Error: errors.New("connection cannot be closed"),
									})))
									Eventually(actual).Should(BeClosed())
								})
							})
						})

						Context("event is read", func() {
							Context("event is not update", func() {
								BeforeEach(func() {
									gomock.InOrder(
										ctx.EXPECT().Err().Return(nil),
										ctx.EXPECT().Err().Return(nil),
										ctx.EXPECT().Err().Return(errors.New("closed")),
									)

									conn.EXPECT().ReadJSON(gomock.Any()).Do(func(s *mast.Stream) {
										*s = mast.Stream{
											Event:   "notification",
											Payload: "{}",
										}
									}).Return(nil)
									conn.EXPECT().Close().Return(nil)
								})

								It("eventually exits", func() {
									actual, err := mastodon.Run()
									Expect(err).NotTo(HaveOccurred())
									Eventually(actual).Should(BeClosed())
								})
							})

							Context("event is update", func() {
								Context("parsing fails", func() {
									BeforeEach(func() {
										gomock.InOrder(
											ctx.EXPECT().Err().Return(nil),
											ctx.EXPECT().Err().Return(nil),
											ctx.EXPECT().Err().Return(errors.New("closed")),
										)

										conn.EXPECT().ReadJSON(gomock.Any()).Do(func(s *mast.Stream) {
											*s = mast.Stream{
												Event:   "update",
												Payload: "{",
											}
										}).Return(nil)
										conn.EXPECT().Close().Return(nil)
									})

									It("sends status and eventually exits", func() {
										actual, err := mastodon.Run()
										Expect(err).NotTo(HaveOccurred())

										Eventually(actual).Should(Receive(WithTransform(func(m service.MessageStatus) error {
											return m.Error
										}, MatchError("unexpected end of JSON input"))))
										Eventually(actual).Should(BeClosed())
									})
								})

								Context("parsing succeeds", func() {
									BeforeEach(func() {
										gomock.InOrder(
											ctx.EXPECT().Err().Return(nil),
											ctx.EXPECT().Err().Return(nil),
											ctx.EXPECT().Err().Return(errors.New("closed")),
										)

										conn.EXPECT().ReadJSON(gomock.Any()).Do(func(s *mast.Stream) {
											*s = mast.Stream{
												Event: "update",
												Payload: `
													{
														"id": "1",
														"account": {
															"id": "1",
															"acct": "@test",
															"display_name": "テスト",
															"username": "test"
														},
														"content": "<p>テスト</p>",
														"emojis": [
															{
																"shortcode": "ios_big_sushi_1"
															},
															{
																"shortcode": "ios_big_sushi_2"
															},
															{
																"shortcode": "ios_big_sushi_3"
															},
															{
																"shortcode": "ios_big_sushi_4"
															}
														],
														"reblog": null,
														"tags": [
															{
																"name": "同人avタイトルジェネレーター"
															}
														],
														"visibility": "private"
													}
												`,
											}
										}).Return(nil)
										conn.EXPECT().Close().Return(nil)
									})

									It("sends status and eventually exits", func() {
										actual, err := mastodon.Run()
										Expect(err).NotTo(HaveOccurred())

										Eventually(actual).Should(Receive(Equal(service.MessageStatus{
											Message: service.Message{
												ID: "1",
												Account: service.Account{
													ID:          "1",
													Acct:        "@test",
													DisplayName: "テスト",
													Username:    "test",
												},
												Content: "テスト",
												Emojis: []service.Emoji{
													{Shortcode: "ios_big_sushi_1"},
													{Shortcode: "ios_big_sushi_2"},
													{Shortcode: "ios_big_sushi_3"},
													{Shortcode: "ios_big_sushi_4"},
												},
												Tags: []service.Tag{
													{Name: "同人avタイトルジェネレーター"},
												},
												IsReblog:    false,
												InReplyToID: "<nil>",
												Visibility:  "private",
											},
										})))
										Eventually(actual).Should(BeClosed())
									})
								})
							})
						})
					})
				})

				Context("websocket connection succeeds with retries", func() {
					Context("websocket connection succeeds with 1 retry", func() {
						BeforeEach(func() {
							ch = make(chan time.Time, 1)
							ch <- time.Time{}

							gomock.InOrder(
								ctx.EXPECT().Err().Return(nil),
								ctx.EXPECT().Err().Return(nil),
								ctx.EXPECT().Err().Return(errors.New("closed")),
							)

							t.EXPECT().Tick(5 * time.Second).Return(ch)

							gomock.InOrder(
								// (1)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									nil,
									&http.Response{
										Header: http.Header{},
									},
									errors.New("dial tcp [::1]:443: connect: connection refused"),
								),
								// (2)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									conn,
									&http.Response{
										Header: http.Header{
											"X-Served-Server": []string{"192.0.2.1:4000"},
										},
									},
									nil,
								),
							)
							conn.EXPECT().Close().Return(nil)
						})

						AfterEach(func() {
							close(ch)
						})

						It("returns channel and eventually exits", func() {
							actual, err := mastodon.Run()
							Expect(err).NotTo(HaveOccurred())

							Eventually(actual).Should(Receive(Equal(service.MessageStatus{
								Error: errors.New("dial tcp [::1]:443: connect: connection refused"),
							})))
							Eventually(actual).Should(BeClosed())
						})
					})

					Context("websocket connection succeeds with 2 retries", func() {
						BeforeEach(func() {
							ch = make(chan time.Time, 2)
							ch <- time.Time{}
							ch <- time.Time{}

							gomock.InOrder(
								ctx.EXPECT().Err().Return(nil),
								ctx.EXPECT().Err().Return(nil),
								ctx.EXPECT().Err().Return(nil),
								ctx.EXPECT().Err().Return(errors.New("closed")),
							)

							gomock.InOrder(
								t.EXPECT().Tick(5*time.Second).Return(ch),
								t.EXPECT().Tick(10*time.Second).Return(ch),
							)

							gomock.InOrder(
								// (1)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									nil,
									&http.Response{
										Header: http.Header{},
									},
									errors.New("dial tcp [::1]:443: connect: connection refused"),
								),
								// (2)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									nil,
									&http.Response{
										Header: http.Header{},
									},
									errors.New("dial tcp [::1]:443: connect: connection refused"),
								),
								// (3)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									conn,
									&http.Response{
										Header: http.Header{
											"X-Served-Server": []string{"192.0.2.1:4000"},
										},
									},
									nil,
								),
							)
							conn.EXPECT().Close().Return(nil)
						})

						AfterEach(func() {
							close(ch)
						})

						It("returns channel and eventually exits", func() {
							actual, err := mastodon.Run()
							Expect(err).NotTo(HaveOccurred())

							Eventually(actual).Should(Receive(Equal(service.MessageStatus{
								Error: errors.New("dial tcp [::1]:443: connect: connection refused"),
							})))
							Eventually(actual).Should(Receive(Equal(service.MessageStatus{
								Error: errors.New("dial tcp [::1]:443: connect: connection refused"),
							})))
							Eventually(actual).Should(BeClosed())
						})
					})

					Context("websocket connection succeeds with 3 retries", func() {
						BeforeEach(func() {
							ch = make(chan time.Time, 3)
							ch <- time.Time{}
							ch <- time.Time{}
							ch <- time.Time{}

							gomock.InOrder(
								ctx.EXPECT().Err().Return(nil),
								ctx.EXPECT().Err().Return(nil),
								ctx.EXPECT().Err().Return(nil),
								ctx.EXPECT().Err().Return(nil),
								ctx.EXPECT().Err().Return(errors.New("closed")),
							)

							gomock.InOrder(
								t.EXPECT().Tick(5*time.Second).Return(ch),
								t.EXPECT().Tick(10*time.Second).Return(ch),
								t.EXPECT().Tick(20*time.Second).Return(ch),
							)

							gomock.InOrder(
								// (1)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									nil,
									&http.Response{
										Header: http.Header{},
									},
									errors.New("dial tcp [::1]:443: connect: connection refused"),
								),
								// (2)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									nil,
									&http.Response{
										Header: http.Header{},
									},
									errors.New("dial tcp [::1]:443: connect: connection refused"),
								),
								// (3)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									nil,
									&http.Response{
										Header: http.Header{},
									},
									errors.New("dial tcp [::1]:443: connect: connection refused"),
								),
								// (4)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									conn,
									&http.Response{
										Header: http.Header{
											"X-Served-Server": []string{"192.0.2.1:4000"},
										},
									},
									nil,
								),
							)
							conn.EXPECT().Close().Return(nil)
						})

						AfterEach(func() {
							close(ch)
						})

						It("returns channel and eventually exits", func() {
							actual, err := mastodon.Run()
							Expect(err).NotTo(HaveOccurred())

							Eventually(actual).Should(Receive(Equal(service.MessageStatus{
								Error: errors.New("dial tcp [::1]:443: connect: connection refused"),
							})))
							Eventually(actual).Should(Receive(Equal(service.MessageStatus{
								Error: errors.New("dial tcp [::1]:443: connect: connection refused"),
							})))
							Eventually(actual).Should(Receive(Equal(service.MessageStatus{
								Error: errors.New("dial tcp [::1]:443: connect: connection refused"),
							})))
							Eventually(actual).Should(BeClosed())
						})
					})
				})
			})
		})
	})
})
