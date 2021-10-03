package reader_test

import (
	"io"
	"testing/iotest"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/reader"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StringFuncReader", func() {
	var (
		ctrl *gomock.Controller
		mock *reader.MockStringGenerator
		r    io.Reader
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = reader.NewMockStringGenerator(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Read()", func() {
		Context("when length == 0", func() {
			BeforeEach(func() {
				r = reader.NewStringFuncReader("\n", 0, mock.Generate)
			})

			It("reads 0 bytes", func() {
				err := iotest.TestReader(r, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when length == 1", func() {
			BeforeEach(func() {
				mock.EXPECT().Generate().Return("The quick brown fox jumps over the lazy dog.")
				r = reader.NewStringFuncReader("\n", 1, mock.Generate)
			})

			It("reads once", func() {
				err := iotest.TestReader(r, []byte("The quick brown fox jumps over the lazy dog."))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when length == 5", func() {
			BeforeEach(func() {
				gomock.InOrder(
					mock.EXPECT().Generate().Return("The quick brown fox jumps over the lazy dog 1."),
					mock.EXPECT().Generate().Return("The quick brown fox jumps over the lazy dog 2."),
					mock.EXPECT().Generate().Return("The quick brown fox jumps over the lazy dog 3."),
					mock.EXPECT().Generate().Return("The quick brown fox jumps over the lazy dog 4."),
					mock.EXPECT().Generate().Return("The quick brown fox jumps over the lazy dog 5."),
				)
				r = reader.NewStringFuncReader("\n", 5, mock.Generate)
			})

			It("reads 5 times", func() {
				err := iotest.TestReader(r, []byte(
					"The quick brown fox jumps over the lazy dog 1.\n"+
						"The quick brown fox jumps over the lazy dog 2.\n"+
						"The quick brown fox jumps over the lazy dog 3.\n"+
						"The quick brown fox jumps over the lazy dog 4.\n"+
						"The quick brown fox jumps over the lazy dog 5.",
				))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
