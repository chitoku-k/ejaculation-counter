package service_test

import (
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tick", func() {
	Context("Name()", func() {
		var (
			t service.Tick
		)

		It("returns packet name", func() {
			actual := t.Name()
			Expect(actual).To(Equal("packets.tick"))
		})
	})

	Context("HashCode()", func() {
		var (
			t service.Tick
		)

		Context("when values are default", func() {
			It("returns code", func() {
				actual := t.HashCode()
				Expect(actual).To(Equal(int64(208537)))
			})
		})

		Context("when values are set", func() {
			BeforeEach(func() {
				t = service.Tick{
					Year:  2006,
					Month: 1,
					Day:   2,
				}
			})

			It("returns code", func() {
				actual := t.HashCode()
				Expect(actual).To(Equal(int64(2136336)))
			})
		})
	})
})

var _ = Describe("Message", func() {
	Context("Name()", func() {
		var (
			m service.Message
		)

		It("returns packet name", func() {
			actual := m.Name()
			Expect(actual).To(Equal("packets.message"))
		})
	})

	Context("HashCode()", func() {
		var (
			m service.Message
		)

		Context("when values are default", func() {
			It("returns code", func() {
				actual := m.HashCode()
				Expect(actual).To(Equal(int64(6727)))
			})
		})

		Context("when values are set", func() {
			BeforeEach(func() {
				m = service.Message{
					ID: "1",
					Account: service.Account{
						ID: "1",
					},
				}
			})

			It("returns code", func() {
				actual := m.HashCode()
				Expect(actual).To(Equal(int64(6759)))
			})
		})
	})
})
