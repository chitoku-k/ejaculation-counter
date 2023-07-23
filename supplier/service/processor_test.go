package service_test

import (
	"context"
	"errors"
	"time"

	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"go.uber.org/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Processor", func() {
	var (
		ctrl      *gomock.Controller
		qw        *service.MockQueueWriter
		processor service.Processor
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		qw = service.NewMockQueueWriter(ctrl)
		processor = service.NewProcessor(qw)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Execute()", func() {
		Context("context is done", func() {
			var (
				ctx    context.Context
				cancel context.CancelFunc
			)

			BeforeEach(func() {
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			})

			It("eventually exits", func() {
				processor.Execute(ctx, nil, nil)
			})
		})

		Context("context is still", func() {
			var (
				ctx       context.Context
				cancel    context.CancelFunc
				scheduler chan service.Tick
				stream    chan service.Status
			)

			Context("received from scheduler", func() {
				Context("queueing fails", func() {
					BeforeEach(func() {
						ctx, cancel = context.WithCancel(context.Background())
						scheduler = make(chan service.Tick)
						stream = make(chan service.Status)

						qw.EXPECT().Publish(ctx, service.Tick{
							Year:  2006,
							Month: 1,
							Day:   2,
						}).Do(func(context.Context, service.Packet) {
							cancel()
						}).Return(
							errors.New("dial tcp [::1]:5672: connect: connection refused"),
						)
					})

					It("sends a tick and eventually exits", func() {
						go processor.Execute(ctx, scheduler, stream)

						scheduler <- service.Tick{
							Year:  2006,
							Month: 1,
							Day:   2,
						}

						Eventually(scheduler).ShouldNot(Receive())
					})
				})

				Context("queueing succeeds", func() {
					BeforeEach(func() {
						ctx, cancel = context.WithCancel(context.Background())
						scheduler = make(chan service.Tick)
						stream = make(chan service.Status)

						qw.EXPECT().Publish(ctx, service.Tick{
							Year:  2006,
							Month: 1,
							Day:   2,
						}).Do(func(context.Context, service.Packet) {
							cancel()
						}).Return(nil)
					})

					It("sends an event and eventually exits", func() {
						go processor.Execute(ctx, scheduler, stream)

						scheduler <- service.Tick{
							Year:  2006,
							Month: 1,
							Day:   2,
						}

						Eventually(scheduler).ShouldNot(Receive())
					})
				})
			})

			Context("received from stream", func() {
				Context("status is error", func() {
					BeforeEach(func() {
						ctx, cancel = context.WithCancel(context.Background())
						scheduler = make(chan service.Tick)
						stream = make(chan service.Status)
					})

					AfterEach(func() {
						cancel()
					})

					It("sends an error and eventually exits", func() {
						go processor.Execute(ctx, scheduler, stream)

						stream <- service.Error{
							Err: errors.New("websocket: bad handshake"),
						}

						Eventually(stream).ShouldNot(Receive())
					})
				})

				Context("status is connection", func() {
					BeforeEach(func() {
						ctx, cancel = context.WithCancel(context.Background())
						scheduler = make(chan service.Tick)
						stream = make(chan service.Status)
					})

					AfterEach(func() {
						cancel()
					})

					It("sends a connection and eventually exits", func() {
						go processor.Execute(ctx, scheduler, stream)

						stream <- service.Connection{
							Server: "::1",
						}

						Eventually(stream).ShouldNot(Receive())
					})
				})

				Context("status is disconnection", func() {
					BeforeEach(func() {
						ctx, cancel = context.WithCancel(context.Background())
						scheduler = make(chan service.Tick)
						stream = make(chan service.Status)
					})

					AfterEach(func() {
						cancel()
					})

					It("sends a disconnection and eventually exits", func() {
						go processor.Execute(ctx, scheduler, stream)

						stream <- service.Disconnection{
							Err: errors.New("websocket: unexpected EOF"),
						}

						Eventually(stream).ShouldNot(Receive())
					})
				})

				Context("status is reconnection", func() {
					BeforeEach(func() {
						ctx, cancel = context.WithCancel(context.Background())
						scheduler = make(chan service.Tick)
						stream = make(chan service.Status)
					})

					AfterEach(func() {
						cancel()
					})

					It("sends a reconnection and eventually exits", func() {
						go processor.Execute(ctx, scheduler, stream)

						stream <- service.Reconnection{
							In: 5 * time.Second,
						}

						Eventually(stream).ShouldNot(Receive())
					})
				})

				Context("status is message", func() {
					Context("queueing fails", func() {
						var (
							message service.Message
						)

						BeforeEach(func() {
							ctx, cancel = context.WithCancel(context.Background())
							scheduler = make(chan service.Tick)
							stream = make(chan service.Status)

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

							qw.EXPECT().Publish(ctx, message).Do(func(context.Context, service.Packet) {
								cancel()
							}).Return(
								errors.New("dial tcp [::1]:5672: connect: connection refused"),
							)
						})

						It("processes a message and eventually exits", func() {
							go processor.Execute(ctx, scheduler, stream)

							stream <- message

							Eventually(stream).ShouldNot(Receive())
						})
					})

					Context("queueing succeeds", func() {
						var (
							message service.Message
						)

						BeforeEach(func() {
							ctx, cancel = context.WithCancel(context.Background())
							scheduler = make(chan service.Tick)
							stream = make(chan service.Status)

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

							qw.EXPECT().Publish(ctx, message).Do(func(context.Context, service.Packet) {
								cancel()
							}).Return(nil)
						})

						It("processes a message and eventually exits", func() {
							go processor.Execute(ctx, scheduler, stream)

							stream <- message

							Eventually(stream).ShouldNot(Receive())
						})
					})
				})
			})
		})
	})
})
