package service_test

import (
	"context"
	"errors"
	"time"

	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Processor", func() {
	var (
		ctrl      *gomock.Controller
		sc        *service.MockScheduler
		st        *service.MockStreaming
		qw        *service.MockQueueWriter
		a1        *service.MockAction
		a2        *service.MockAction
		a3        *service.MockAction
		processor service.Processor
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		sc = service.NewMockScheduler(ctrl)
		st = service.NewMockStreaming(ctrl)
		qw = service.NewMockQueueWriter(ctrl)
		a1 = service.NewMockAction(ctrl)
		a2 = service.NewMockAction(ctrl)
		a3 = service.NewMockAction(ctrl)
		processor = service.NewProcessor(sc, st, qw, []service.Action{a1, a2, a3})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Execute()", func() {
		Context("streaming fails", func() {
			var (
				ctx    context.Context
				cancel context.CancelFunc
			)

			BeforeEach(func() {
				ctx, cancel = context.WithCancel(context.Background())
				st.EXPECT().Run(ctx).Do(func(context.Context) {
					cancel()
				}).Return(
					nil,
					errors.New("failed in streaming"),
				)
			})

			It("eventually exits", func() {
				processor.Execute(ctx)
				Eventually(ctx.Done()).Should(BeClosed())
			})
		})

		Context("streaming succeeds", func() {
			Context("context is done", func() {
				var (
					ctx    context.Context
					cancel context.CancelFunc
				)

				BeforeEach(func() {
					ctx, cancel = context.WithCancel(context.Background())
					st.EXPECT().Run(ctx).Return(make(chan service.Status), nil)
					sc.EXPECT().Start().Do(cancel).Return(make(chan service.Event))
				})

				It("eventually exits", func() {
					processor.Execute(ctx)
					Eventually(ctx.Done()).Should(BeClosed())
				})
			})

			Context("context is still", func() {
				var (
					ctx    context.Context
					cancel context.CancelFunc
					events chan service.Event
					stream chan service.Status
				)

				Context("received from scheduler", func() {
					Context("queueing fails", func() {
						BeforeEach(func() {
							ctx, cancel = context.WithCancel(context.Background())
							events = make(chan service.Event)
							stream = make(chan service.Status)

							st.EXPECT().Run(ctx).Return(stream, nil)
							sc.EXPECT().Start().Return(events)
							qw.EXPECT().Publish(&service.IncrementEvent{
								Date:   "2006-01-02 15:04:05 +0900 JST",
								UserID: "1",
							}).Do(func(service.Event) {
								cancel()
							}).Return(
								errors.New("dial tcp [::1]:5672: connect: connection refused"),
							)
						})

						AfterEach(func() {
							close(events)
							close(stream)
						})

						It("sends an event and eventually exits", func() {
							processor.Execute(ctx)

							events <- &service.IncrementEvent{
								Date:   "2006-01-02 15:04:05 +0900 JST",
								UserID: "1",
							}

							Eventually(ctx.Done()).Should(BeClosed())
						})
					})

					Context("queueing succeeds", func() {
						BeforeEach(func() {
							ctx, cancel = context.WithCancel(context.Background())
							events = make(chan service.Event)
							stream = make(chan service.Status)

							st.EXPECT().Run(ctx).Return(stream, nil)
							sc.EXPECT().Start().Return(events)
							qw.EXPECT().Publish(&service.IncrementEvent{
								Date:   "2006-01-02 15:04:05 +0900 JST",
								UserID: "1",
							}).Return(nil)

						})

						AfterEach(func() {
							close(events)
							close(stream)
						})

						It("sends an event and eventually exits", func() {
							processor.Execute(ctx)

							events <- &service.IncrementEvent{
								Date:   "2006-01-02 15:04:05 +0900 JST",
								UserID: "1",
							}
							cancel()

							Eventually(ctx.Done()).Should(BeClosed())
						})
					})
				})

				Context("received from stream", func() {
					Context("status is error", func() {
						BeforeEach(func() {
							ctx, cancel = context.WithCancel(context.Background())
							events = make(chan service.Event)
							stream = make(chan service.Status, 1)

							stream <- service.Error{
								Err: errors.New("websocket: bad handshake"),
							}

							st.EXPECT().Run(ctx).Return(stream, nil)
							sc.EXPECT().Start().Return(events)
						})

						AfterEach(func() {
							close(events)
							close(stream)
						})

						It("sends an error and eventually exits", func() {
							processor.Execute(ctx)

							// HACK
							time.Sleep(500 * time.Millisecond)
							cancel()

							Eventually(ctx.Done()).Should(BeClosed())
						})
					})

					Context("status is reconnection", func() {
						BeforeEach(func() {
							ctx, cancel = context.WithCancel(context.Background())
							events = make(chan service.Event)
							stream = make(chan service.Status, 1)

							stream <- service.Reconnection{
								In: 5 * time.Second,
							}

							st.EXPECT().Run(ctx).Return(stream, nil)
							sc.EXPECT().Start().Return(events)
						})

						AfterEach(func() {
							close(events)
							close(stream)
						})

						It("sends a reconnection and eventually exits", func() {
							processor.Execute(ctx)

							// HACK
							time.Sleep(500 * time.Millisecond)
							cancel()

							Eventually(ctx.Done()).Should(BeClosed())
						})
					})

					Context("status is message", func() {
						BeforeEach(func() {
							ctx, cancel = context.WithCancel(context.Background())
							events = make(chan service.Event)
							stream = make(chan service.Status, 2)

							st.EXPECT().Run(ctx).Return(stream, nil)
							sc.EXPECT().Start().Return(events)
						})

						AfterEach(func() {
							close(events)
							close(stream)
						})

						Context("message has already been replied", func() {
							var (
								message1 service.Message
								message2 service.Message
								e        service.Event
							)

							BeforeEach(func() {
								message1 = service.Message{
									ID: "1",
									Account: service.Account{
										ID:          "1",
										Acct:        "@test",
										DisplayName: "テスト",
										Username:    "test",
									},
									Content: "test",
								}
								message2 = service.Message{
									ID:          "2",
									InReplyToID: "1",
									Account: service.Account{
										ID:          "1",
										Acct:        "@test",
										DisplayName: "テスト",
										Username:    "test",
									},
									Content: "test reply",
								}

								stream <- message1
								stream <- message2

								e = &service.ReplyEvent{
									Acct:        "@test",
									Body:        "テスト！",
									InReplyToID: "1",
									Visibility:  "private",
								}

								gomock.InOrder(
									a1.EXPECT().Target(message1).Return(false),
									a2.EXPECT().Target(message1).Return(true),
									a3.EXPECT().Target(message1).Return(false),
								)

								gomock.InOrder(
									a2.EXPECT().Event(message1).Return(e, 0, nil),
									a2.EXPECT().Name().Do(cancel).Return("a2"),
								)

								qw.EXPECT().Publish(e).Do(func(service.Event) {
									cancel()
								}).Return(nil)
							})

							It("processes message1 and eventually exits", func() {
								processor.Execute(ctx)

								Eventually(ctx.Done()).Should(BeClosed())
							})
						})

						Context("message is not yet replied", func() {
							var (
								message service.Message
								e1      service.Event
								e2      service.Event
								e3      service.Event
							)

							BeforeEach(func() {
								message = service.Message{
									ID: "1",
									Account: service.Account{
										ID:          "1",
										Acct:        "@test",
										DisplayName: "テスト",
										Username:    "test",
									},
									Content: "test",
								}

								stream <- message
							})

							Context("no actions are correspondent", func() {
								BeforeEach(func() {
									gomock.InOrder(
										a1.EXPECT().Target(message).Return(false),
										a2.EXPECT().Target(message).Return(false),
										a3.EXPECT().Target(message).Do(func(service.Message) {
											cancel()
										}).Return(false),
									)
								})

								It("eventually exits", func() {
									processor.Execute(ctx)

									Eventually(ctx.Done()).Should(BeClosed())
								})
							})

							Context("1 action is correspondent", func() {
								Context("action cannot be processed", func() {
									BeforeEach(func() {
										gomock.InOrder(
											a1.EXPECT().Target(message).Return(false),
											a2.EXPECT().Target(message).Return(true),
											a3.EXPECT().Target(message).Return(false),
										)

										gomock.InOrder(
											a2.EXPECT().Event(message).Return(nil, 0, errors.New("failed to create event")),
											a2.EXPECT().Name().Return("a2"),
											a2.EXPECT().Name().Return("a2"),
											a2.EXPECT().Name().Do(cancel).Return("a2"),
										)

										qw.EXPECT().Publish(&service.ReplyErrorEvent{
											InReplyToID: "1",
											Acct:        "@test",
											ActionName:  "a2",
										}).Do(func(service.Event) {
											cancel()
										}).Return(nil)
									})

									It("does nothing and eventually exits", func() {
										processor.Execute(ctx)

										Eventually(ctx.Done()).Should(BeClosed())
									})
								})

								Context("action is processed", func() {
									Context("queueing fails", func() {
										BeforeEach(func() {
											e1 = &service.ReplyEvent{
												Acct:        "@test",
												Body:        "テスト！",
												InReplyToID: "1",
												Visibility:  "private",
											}

											gomock.InOrder(
												a1.EXPECT().Target(message).Return(false),
												a2.EXPECT().Target(message).Return(true),
												a3.EXPECT().Target(message).Return(false),
											)

											gomock.InOrder(
												a2.EXPECT().Event(message).Return(e1, 0, nil),
												a2.EXPECT().Name().Do(cancel).Return("a2"),
											)

											qw.EXPECT().Publish(e1).Return(errors.New("failed to write message"))
										})

										It("does nothing and eventually exits", func() {
											processor.Execute(ctx)

											Eventually(ctx.Done()).Should(BeClosed())
										})
									})

									Context("queueing succeeds", func() {
										BeforeEach(func() {
											e1 = &service.ReplyEvent{
												Acct:        "@test",
												Body:        "テスト！",
												InReplyToID: "1",
												Visibility:  "private",
											}

											gomock.InOrder(
												a1.EXPECT().Target(message).Return(false),
												a2.EXPECT().Target(message).Return(true),
												a3.EXPECT().Target(message).Return(false),
											)

											gomock.InOrder(
												a2.EXPECT().Event(message).Return(e1, 0, nil),
												a2.EXPECT().Name().Do(cancel).Return("a2"),
											)

											qw.EXPECT().Publish(e1).Return(nil)
										})

										It("writes to queue and eventually exits", func() {
											processor.Execute(ctx)

											Eventually(ctx.Done()).Should(BeClosed())
										})
									})
								})
							})

							Context("multiple actions are correspondent", func() {
								Context("actions cannot be processed", func() {
									BeforeEach(func() {
										gomock.InOrder(
											a1.EXPECT().Target(message).Return(false),
											a2.EXPECT().Target(message).Return(true),
											a3.EXPECT().Target(message).Return(true),
										)

										gomock.InOrder(
											a2.EXPECT().Event(message).Return(nil, 0, errors.New("failed to create event")),
											a2.EXPECT().Name().Return("a2"),
											a2.EXPECT().Name().Return("a2"),
											a2.EXPECT().Name().Return("a2"),
										)

										gomock.InOrder(
											a3.EXPECT().Event(message).Return(nil, 0, errors.New("failed to create event")),
											a3.EXPECT().Name().Return("a3"),
											a3.EXPECT().Name().Return("a3"),
											a3.EXPECT().Name().Do(cancel).Return("a3"),
										)

										qw.EXPECT().Publish(&service.ReplyErrorEvent{
											InReplyToID: "1",
											Acct:        "@test",
											ActionName:  "a2",
										}).Return(nil)

										qw.EXPECT().Publish(&service.ReplyErrorEvent{
											InReplyToID: "1",
											Acct:        "@test",
											ActionName:  "a3",
										}).Do(func(service.Event) {
											cancel()
										}).Return(nil)
									})

									It("does nothing and eventually exits", func() {
										processor.Execute(ctx)
										Eventually(ctx.Done()).Should(BeClosed())
									})
								})

								Context("actions are processed", func() {
									Context("queueing fails", func() {
										BeforeEach(func() {
											e1 = &service.ReplyEvent{
												Acct:        "@test",
												Body:        "テスト！",
												InReplyToID: "1",
												Visibility:  "private",
											}
											e2 = &service.UpdateEvent{
												Date:   "2006-01-02 15:04:05 +0900 JST",
												UserID: "1",
											}
											e3 = &service.AdministrationEvent{
												Acct:        "@test",
												Command:     "SELECT 1",
												InReplyToID: "1",
												Type:        "test",
											}

											gomock.InOrder(
												a1.EXPECT().Target(message).Return(true),
												a2.EXPECT().Target(message).Return(true),
												a3.EXPECT().Target(message).Return(true),
											)

											gomock.InOrder(
												a1.EXPECT().Event(message).Return(e1, 1, nil),
												a1.EXPECT().Name().Return("a1"),
											)

											gomock.InOrder(
												a2.EXPECT().Event(message).Return(e2, 2, nil),
												a2.EXPECT().Name().Return("a2"),
											)

											gomock.InOrder(
												a3.EXPECT().Event(message).Return(e3, 0, nil),
												a3.EXPECT().Name().Do(cancel).Return("a3"),
											)

											gomock.InOrder(
												qw.EXPECT().Publish(e3).Return(errors.New("failed to write message")),
												qw.EXPECT().Publish(e1).Return(errors.New("failed to write message")),
												qw.EXPECT().Publish(e2).Return(errors.New("failed to write message")),
											)
										})

										It("does nothing and eventually exits", func() {
											processor.Execute(ctx)

											Eventually(ctx.Done()).Should(BeClosed())
										})
									})

									Context("queueing succeeds", func() {
										BeforeEach(func() {
											e1 = &service.ReplyEvent{
												Acct:        "@test",
												Body:        "テスト！",
												InReplyToID: "1",
												Visibility:  "private",
											}
											e2 = &service.UpdateEvent{
												Date:   "2006-01-02 15:04:05 +0900 JST",
												UserID: "1",
											}
											e3 = &service.AdministrationEvent{
												Acct:        "@test",
												Command:     "SELECT 1",
												InReplyToID: "1",
												Type:        "test",
											}

											gomock.InOrder(
												a1.EXPECT().Target(message).Return(true),
												a2.EXPECT().Target(message).Return(true),
												a3.EXPECT().Target(message).Return(true),
											)

											gomock.InOrder(
												a1.EXPECT().Event(message).Return(e1, 1, nil),
												a1.EXPECT().Name().Return("a1"),
											)

											gomock.InOrder(
												a2.EXPECT().Event(message).Return(e2, 2, nil),
												a2.EXPECT().Name().Return("a2"),
											)

											gomock.InOrder(
												a3.EXPECT().Event(message).Return(e3, 0, nil),
												a3.EXPECT().Name().Do(cancel).Return("a3"),
											)

											gomock.InOrder(
												qw.EXPECT().Publish(e3).Return(nil),
												qw.EXPECT().Publish(e1).Return(nil),
												qw.EXPECT().Publish(e2).Return(nil),
											)
										})

										It("writes to queue and eventually exits", func() {
											processor.Execute(ctx)

											Eventually(ctx.Done()).Should(BeClosed())
										})
									})
								})
							})
						})
					})
				})
			})
		})
	})
})
