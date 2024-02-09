package streaming_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/streaming"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/gorilla/websocket"
	mast "github.com/mattn/go-mastodon"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Mastodon", func() {
	var (
		ctrl *gomock.Controller
		conn *wrapper.MockConn
		d    *wrapper.MockDialer
		t    *wrapper.MockTimer
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		conn = wrapper.NewMockConn(ctrl)
		d = wrapper.NewMockDialer(ctrl)
		t = wrapper.NewMockTimer(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Run()", func() {
		var (
			env      config.Environment
			mastodon service.Streaming
			ctx      context.Context
			cancel   context.CancelFunc
			ch       chan time.Time
			stream   mast.Stream
		)

		Context("parsing server URL fails", func() {
			BeforeEach(func() {
				env = config.Environment{
					Mastodon: config.Mastodon{
						ServerURL:   ":/",
						AccessToken: "token",
					},
				}
				mastodon = streaming.NewMastodon(env, d, t)
			})

			It("returns an error", func() {
				err := mastodon.Run(context.Background())
				Expect(err).To(MatchError(`failed to parse server URL: parse ":/": missing protocol scheme`))
			})
		})

		Context("parsing server URL succeeds", func() {
			BeforeEach(func() {
				env = config.Environment{
					Mastodon: config.Mastodon{
						ServerURL:   "https://mastodon.example.com",
						Stream:      "user",
						AccessToken: "token",
					},
				}
				mastodon = streaming.NewMastodon(env, d, t)
				ctx, cancel = context.WithCancel(context.Background())
			})

			Context("websocket connection fails", func() {
				Context("HTTP response unavailable", func() {
					BeforeEach(func() {
						ch = make(chan time.Time, 1)
						ch <- time.Time{}

						t.EXPECT().After(5 * time.Second).Return(ch)

						gomock.InOrder(
							// (1)
							d.EXPECT().DialContext(
								ctx,
								"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
								nil,
							).Return(
								nil,
								nil,
								errors.New("dial tcp [::1]:443: connect: connection refused"),
							),
							// (2)
							d.EXPECT().DialContext(
								ctx,
								"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
								nil,
							).Do(func(context.Context, string, http.Header) {
								cancel()
							}).Return(
								nil,
								nil,
								context.Canceled,
							),
						)
					})

					It("eventually exits", func() {
						actual := mastodon.Statuses()
						go func() {
							defer GinkgoRecover()

							err := mastodon.Run(ctx)
							Expect(err).To(Equal(context.Canceled))
						}()

						Eventually(actual).Should(Receive(Equal(service.Error{
							Err: errors.New("dial tcp [::1]:443: connect: connection refused"),
						})))
						Eventually(actual).Should(Receive(Equal(service.Reconnection{
							In: 5 * time.Second,
						})))
						Eventually(ctx.Done()).Should(BeClosed())
					})
				})

				Context("HTTP response available", func() {
					BeforeEach(func() {
						ch = make(chan time.Time, 1)
						ch <- time.Time{}

						t.EXPECT().After(5 * time.Second).Return(ch)

						gomock.InOrder(
							// (1)
							d.EXPECT().DialContext(
								ctx,
								"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
								nil,
							).Return(
								nil,
								&http.Response{
									Status: "401 Unauthorized",
								},
								errors.New("websocket: bad handshake"),
							),
							// (2)
							d.EXPECT().DialContext(
								ctx,
								"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
								nil,
							).Do(func(context.Context, string, http.Header) {
								cancel()
							}).Return(
								nil,
								nil,
								context.Canceled,
							),
						)
					})

					It("eventually exits", func() {
						actual := mastodon.Statuses()
						go func() {
							defer GinkgoRecover()

							err := mastodon.Run(ctx)
							Expect(err).To(Equal(context.Canceled))
						}()

						Eventually(actual).Should(Receive(Equal(service.Error{
							Err: fmt.Errorf("failed to connect: 401 Unauthorized: %w", errors.New("websocket: bad handshake")),
						})))
						Eventually(actual).Should(Receive(Equal(service.Reconnection{
							In: 5 * time.Second,
						})))
						Eventually(ctx.Done()).Should(BeClosed())
					})
				})
			})

			Context("websocket connection succeeds", func() {
				Context("event cannot be read", func() {
					Context("connection cannot be closed", func() {
						BeforeEach(func() {
							gomock.InOrder(
								// (1)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									conn,
									&http.Response{
										Header: http.Header{
											"X-Served-By": []string{"192.0.2.1:4000"},
										},
									},
									nil,
								),
								// (2)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Do(func(context.Context, string, http.Header) {
									cancel()
								}).Return(
									nil,
									nil,
									context.Canceled,
								),
							)

							message := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Shutdown")
							conn.EXPECT().ReadJSON(&stream).Return(errors.New("dial tcp [::1]:443: connect: connection refused"))
							conn.EXPECT().WriteMessage(websocket.CloseMessage, message)
							conn.EXPECT().Close().Return(errors.New("connection cannot be closed"))
						})

						It("sends disconnection and eventually exits", func() {
							actual := mastodon.Statuses()
							go func() {
								defer GinkgoRecover()

								err := mastodon.Run(ctx)
								Expect(err).To(Equal(context.Canceled))
							}()

							Eventually(actual).Should(Receive(Equal(service.Connection{
								Server: "192.0.2.1:4000",
							})))
							Eventually(actual).Should(Receive(Equal(service.Disconnection{
								Err: errors.New("dial tcp [::1]:443: connect: connection refused"),
							})))
							Eventually(actual).Should(Receive(Equal(service.Error{
								Err: errors.New("connection cannot be closed"),
							})))
							Eventually(ctx.Done()).Should(BeClosed())
						})
					})

					Context("connection can be closed", func() {
						BeforeEach(func() {
							gomock.InOrder(
								// (1)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									conn,
									&http.Response{
										Header: http.Header{
											"X-Served-By": []string{"192.0.2.1:4000"},
										},
									},
									nil,
								),
								// (2)
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Do(func(context.Context, string, http.Header) {
									cancel()
								}).Return(
									nil,
									nil,
									context.Canceled,
								),
							)

							message := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Shutdown")
							conn.EXPECT().ReadJSON(&stream).Return(errors.New("dial tcp [::1]:443: connect: connection refused"))
							conn.EXPECT().WriteMessage(websocket.CloseMessage, message)
							conn.EXPECT().Close().Return(nil)
						})

						It("sends disconnection and eventually exits", func() {
							actual := mastodon.Statuses()
							go func() {
								defer GinkgoRecover()

								err := mastodon.Run(ctx)
								Expect(err).To(Equal(context.Canceled))
							}()

							Eventually(actual).Should(Receive(Equal(service.Connection{
								Server: "192.0.2.1:4000",
							})))
							Eventually(actual).Should(Receive(Equal(service.Disconnection{
								Err: errors.New("dial tcp [::1]:443: connect: connection refused"),
							})))
							Eventually(ctx.Done()).Should(BeClosed())
						})
					})
				})

				Context("event is read", func() {
					Context("event is not update", func() {
						BeforeEach(func() {
							d.EXPECT().DialContext(
								ctx,
								"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
								nil,
							).Return(
								conn,
								&http.Response{
									Header: http.Header{
										"X-Served-By": []string{"192.0.2.1:4000"},
									},
								},
								nil,
							)

							gomock.InOrder(
								// (1)
								conn.EXPECT().ReadJSON(&stream).Do(func(s *mast.Stream) {
									*s = mast.Stream{
										Event:   "notification",
										Payload: "{}",
									}
								}).Return(nil),
								// (2)
								conn.EXPECT().ReadJSON(&stream).Do(func(*mast.Stream) {
									cancel()
								}).Return(context.Canceled),
							)
						})

						It("eventually exits", func() {
							actual := mastodon.Statuses()
							go func() {
								defer GinkgoRecover()

								err := mastodon.Run(ctx)
								Expect(err).To(Equal(context.Canceled))
							}()

							Eventually(actual).Should(Receive(Equal(service.Connection{
								Server: "192.0.2.1:4000",
							})))
							Eventually(ctx.Done()).Should(BeClosed())
						})
					})

					Context("event is update", func() {
						Context("parsing fails", func() {
							BeforeEach(func() {
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									conn,
									&http.Response{
										Header: http.Header{
											"X-Served-By": []string{"192.0.2.1:4000"},
										},
									},
									nil,
								)

								gomock.InOrder(
									// (1)
									conn.EXPECT().ReadJSON(&stream).Do(func(s *mast.Stream) {
										*s = mast.Stream{
											Event:   "update",
											Payload: "{",
										}
									}).Return(nil),
									// (2)
									conn.EXPECT().ReadJSON(&stream).Do(func(*mast.Stream) {
										cancel()
									}).Return(context.Canceled),
								)
							})

							It("sends error and eventually exits", func() {
								actual := mastodon.Statuses()
								go func() {
									defer GinkgoRecover()

									err := mastodon.Run(ctx)
									Expect(err).To(Equal(context.Canceled))
								}()

								Eventually(actual).Should(Receive(Equal(service.Connection{
									Server: "192.0.2.1:4000",
								})))
								Eventually(actual).Should(Receive(WithTransform(func(m service.Error) error {
									return m.Err
								}, MatchError("unexpected EOF"))))
								Eventually(ctx.Done()).Should(BeClosed())
							})
						})

						Context("parsing succeeds", func() {
							BeforeEach(func() {
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									conn,
									&http.Response{
										Header: http.Header{
											"X-Served-By": []string{"192.0.2.1:4000"},
										},
									},
									nil,
								)

								gomock.InOrder(
									// (1)
									conn.EXPECT().ReadJSON(&stream).Do(func(s *mast.Stream) {
										*s = mast.Stream{
											Event: "update",
											Payload: `
												{
													"id": "2",
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
													"in_reply_to_id": "1",
													"tags": [
														{
															"name": "同人avタイトルジェネレーター"
														}
													],
													"visibility": "private"
												}
											`,
										}
									}).Return(nil),
									// (2)
									conn.EXPECT().ReadJSON(&stream).Do(func(*mast.Stream) {
										cancel()
									}).Return(context.Canceled),
								)
							})

							It("sends status and eventually exits", func() {
								actual := mastodon.Statuses()
								go func() {
									defer GinkgoRecover()

									err := mastodon.Run(ctx)
									Expect(err).To(Equal(context.Canceled))
								}()

								Eventually(actual).Should(Receive(Equal(service.Connection{
									Server: "192.0.2.1:4000",
								})))
								Eventually(actual).Should(Receive(Equal(service.Message{
									ID: "2",
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
									InReplyToID: "1",
									Visibility:  "private",
								})))
								Eventually(ctx.Done()).Should(BeClosed())
							})
						})
					})

					Context("event is conversation", func() {
						Context("parsing fails", func() {
							BeforeEach(func() {
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									conn,
									&http.Response{
										Header: http.Header{
											"X-Served-By": []string{"192.0.2.1:4000"},
										},
									},
									nil,
								)

								gomock.InOrder(
									// (1)
									conn.EXPECT().ReadJSON(&stream).Do(func(s *mast.Stream) {
										*s = mast.Stream{
											Event:   "conversation",
											Payload: "{",
										}
									}).Return(nil),
									// (2)
									conn.EXPECT().ReadJSON(&stream).Do(func(*mast.Stream) {
										cancel()
									}).Return(context.Canceled),
								)
							})

							It("sends error and eventually exits", func() {
								actual := mastodon.Statuses()
								go func() {
									defer GinkgoRecover()

									err := mastodon.Run(ctx)
									Expect(err).To(Equal(context.Canceled))
								}()

								Eventually(actual).Should(Receive(Equal(service.Connection{
									Server: "192.0.2.1:4000",
								})))
								Eventually(actual).Should(Receive(WithTransform(func(m service.Error) error {
									return m.Err
								}, MatchError("unexpected EOF"))))
								Eventually(ctx.Done()).Should(BeClosed())
							})
						})

						Context("parsing succeeds", func() {
							BeforeEach(func() {
								d.EXPECT().DialContext(
									ctx,
									"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
									nil,
								).Return(
									conn,
									&http.Response{
										Header: http.Header{
											"X-Served-By": []string{"192.0.2.1:4000"},
										},
									},
									nil,
								)

								gomock.InOrder(
									// (1)
									conn.EXPECT().ReadJSON(&stream).Do(func(s *mast.Stream) {
										*s = mast.Stream{
											Event: "conversation",
											Payload: `
												{
													"last_status": {
														"id": "2",
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
														"in_reply_to_id": "1",
														"tags": [
															{
																"name": "同人avタイトルジェネレーター"
															}
														],
														"visibility": "private"
													}
												}
											`,
										}
									}).Return(nil),
									// (2)
									conn.EXPECT().ReadJSON(&stream).Do(func(*mast.Stream) {
										cancel()
									}).Return(context.Canceled),
								)
							})

							It("sends status and eventually exits", func() {
								actual := mastodon.Statuses()
								go func() {
									defer GinkgoRecover()

									err := mastodon.Run(ctx)
									Expect(err).To(Equal(context.Canceled))
								}()

								Eventually(actual).Should(Receive(Equal(service.Connection{
									Server: "192.0.2.1:4000",
								})))
								Eventually(actual).Should(Receive(Equal(service.Message{
									ID: "2",
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
									InReplyToID: "1",
									Visibility:  "private",
								})))
								Eventually(ctx.Done()).Should(BeClosed())
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

						t.EXPECT().After(5 * time.Second).Return(ch)

						gomock.InOrder(
							// (1)
							d.EXPECT().DialContext(
								ctx,
								"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
								nil,
							).Return(
								nil,
								nil,
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
										"X-Served-By": []string{"192.0.2.1:4000"},
									},
								},
								nil,
							),
						)

						conn.EXPECT().ReadJSON(&stream).Do(func(*mast.Stream) {
							cancel()
						}).Return(context.Canceled)
					})

					It("eventually exits", func() {
						actual := mastodon.Statuses()
						go func() {
							defer GinkgoRecover()

							err := mastodon.Run(ctx)
							Expect(err).To(Equal(context.Canceled))
						}()

						Eventually(actual).Should(Receive(Equal(service.Error{
							Err: errors.New("dial tcp [::1]:443: connect: connection refused"),
						})))
						Eventually(actual).Should(Receive(Equal(service.Reconnection{
							In: 5 * time.Second,
						})))
						Eventually(actual).Should(Receive(Equal(service.Connection{
							Server: "192.0.2.1:4000",
						})))
						Eventually(ctx.Done()).Should(BeClosed())
					})
				})

				Context("websocket connection succeeds with 2 retries", func() {
					BeforeEach(func() {
						ch = make(chan time.Time, 2)
						ch <- time.Time{}
						ch <- time.Time{}

						gomock.InOrder(
							t.EXPECT().After(5*time.Second).Return(ch),
							t.EXPECT().After(10*time.Second).Return(ch),
						)

						gomock.InOrder(
							// (1)
							d.EXPECT().DialContext(
								ctx,
								"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
								nil,
							).Return(
								nil,
								nil,
								errors.New("dial tcp [::1]:443: connect: connection refused"),
							),
							// (2)
							d.EXPECT().DialContext(
								ctx,
								"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
								nil,
							).Return(
								nil,
								nil,
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
										"X-Served-By": []string{"192.0.2.1:4000"},
									},
								},
								nil,
							),
						)

						conn.EXPECT().ReadJSON(&stream).Do(func(*mast.Stream) {
							cancel()
						}).Return(context.Canceled)
					})

					It("eventually exits", func() {
						actual := mastodon.Statuses()
						go func() {
							defer GinkgoRecover()

							err := mastodon.Run(ctx)
							Expect(err).To(Equal(context.Canceled))
						}()

						Eventually(actual).Should(Receive(Equal(service.Error{
							Err: errors.New("dial tcp [::1]:443: connect: connection refused"),
						})))
						Eventually(actual).Should(Receive(Equal(service.Reconnection{
							In: 5 * time.Second,
						})))
						Eventually(actual).Should(Receive(Equal(service.Error{
							Err: errors.New("dial tcp [::1]:443: connect: connection refused"),
						})))
						Eventually(actual).Should(Receive(Equal(service.Reconnection{
							In: 10 * time.Second,
						})))
						Eventually(actual).Should(Receive(Equal(service.Connection{
							Server: "192.0.2.1:4000",
						})))
						Eventually(ctx.Done()).Should(BeClosed())
					})
				})

				Context("websocket connection succeeds with 3 retries", func() {
					BeforeEach(func() {
						ch = make(chan time.Time, 3)
						ch <- time.Time{}
						ch <- time.Time{}
						ch <- time.Time{}

						gomock.InOrder(
							t.EXPECT().After(5*time.Second).Return(ch),
							t.EXPECT().After(10*time.Second).Return(ch),
							t.EXPECT().After(20*time.Second).Return(ch),
						)

						gomock.InOrder(
							// (1)
							d.EXPECT().DialContext(
								ctx,
								"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
								nil,
							).Return(
								nil,
								nil,
								errors.New("dial tcp [::1]:443: connect: connection refused"),
							),
							// (2)
							d.EXPECT().DialContext(
								ctx,
								"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
								nil,
							).Return(
								nil,
								nil,
								errors.New("dial tcp [::1]:443: connect: connection refused"),
							),
							// (3)
							d.EXPECT().DialContext(
								ctx,
								"wss://mastodon.example.com/api/v1/streaming?access_token=token&stream=user",
								nil,
							).Return(
								nil,
								nil,
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
										"X-Served-By": []string{"192.0.2.1:4000"},
									},
								},
								nil,
							),
						)

						conn.EXPECT().ReadJSON(&stream).Do(func(*mast.Stream) {
							cancel()
						}).Return(context.Canceled)
					})

					It("eventually exits", func() {
						actual := mastodon.Statuses()
						go func() {
							defer GinkgoRecover()

							err := mastodon.Run(ctx)
							Expect(err).To(Equal(context.Canceled))
						}()

						Eventually(actual).Should(Receive(Equal(service.Error{
							Err: errors.New("dial tcp [::1]:443: connect: connection refused"),
						})))
						Eventually(actual).Should(Receive(Equal(service.Reconnection{
							In: 5 * time.Second,
						})))
						Eventually(actual).Should(Receive(Equal(service.Error{
							Err: errors.New("dial tcp [::1]:443: connect: connection refused"),
						})))
						Eventually(actual).Should(Receive(Equal(service.Reconnection{
							In: 10 * time.Second,
						})))
						Eventually(actual).Should(Receive(Equal(service.Error{
							Err: errors.New("dial tcp [::1]:443: connect: connection refused"),
						})))
						Eventually(actual).Should(Receive(Equal(service.Reconnection{
							In: 20 * time.Second,
						})))
						Eventually(actual).Should(Receive(Equal(service.Connection{
							Server: "192.0.2.1:4000",
						})))
						Eventually(ctx.Done()).Should(BeClosed())
					})
				})
			})
		})
	})

	Describe("Close()", func() {
		var (
			env      config.Environment
			mastodon service.Streaming
		)

		Context("not exit", func() {
			Context("connection not established", func() {
				BeforeEach(func() {
					env = config.Environment{
						Mastodon: config.Mastodon{
							ServerURL:   ":/",
							AccessToken: "token",
						},
					}
					mastodon = streaming.NewMastodon(env, d, t)
				})

				It("does nothing", func() {
					err := mastodon.Close(false)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("exit", func() {
			Context("connection not established", func() {
				BeforeEach(func() {
					env = config.Environment{
						Mastodon: config.Mastodon{
							ServerURL:   ":/",
							AccessToken: "token",
						},
					}
					mastodon = streaming.NewMastodon(env, d, t)
				})

				It("does nothing", func() {
					err := mastodon.Close(true)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
