package reader_test

import (
	"bytes"
	"io"
	"testing/iotest"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/reader"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JsonStreamReader", func() {
	var (
		base io.ReadWriter
		r    io.ReadCloser
	)

	BeforeEach(func() {
		base = &bytes.Buffer{}
		r = reader.NewJsonStreamReader("\n", 1, io.NopCloser(base))
	})

	Describe("Read()", func() {
		Context("when no items emitted", func() {
			BeforeEach(func() {
				_, err := io.WriteString(base, `[]`)
				Expect(err).NotTo(HaveOccurred())
			})

			It("reads 0 bytes", func() {
				err := iotest.TestReader(r, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when items emitted", func() {
			BeforeEach(func() {
				_, err := io.WriteString(base, `[
					"item1",
					"item2",
					"item3",
					"item4",
					"item5"
				]`)
				Expect(err).NotTo(HaveOccurred())
			})

			It("reads items", func() {
				err := iotest.TestReader(r, []byte("item1\nitem2\nitem3\nitem4\nitem5"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when invalid json given", func() {
			BeforeEach(func() {
				_, err := io.WriteString(base, `[
					"item1",
					"item2",
					"item3",
					"item4",
					"item5",,,,,,,,,,
				]`)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				actual, err := io.ReadAll(r)
				Expect(actual).To(Equal([]byte("item1\nitem2\nitem3\nitem4\nitem5")))
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Close()", func() {
		Context("when closing", func() {
			It("returns", func() {
				err := r.Close()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
