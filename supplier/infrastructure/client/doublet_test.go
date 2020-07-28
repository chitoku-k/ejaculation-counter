package client_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Doublet", func() {
	var (
		ctrl    *gomock.Controller
		c       *wrapper.MockHttpClient
		doublet client.Doublet
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = wrapper.NewMockHttpClient(ctrl)
		doublet = client.NewDoublet(c)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Do()", func() {
		Context("fetching fails", func() {
			BeforeEach(func() {
				c.EXPECT().Get("http://reactor/doublet").Return(
					nil,
					errors.New(`Get "http://reactor/doublet": dial tcp [::1]:80: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				actual, err := doublet.Do("http://reactor/doublet")
				Expect(actual).To(BeNil())
				Expect(err).To(MatchError(`failed to fetch challenge result: Get "http://reactor/doublet": dial tcp [::1]:80: connect: connection refused`))
			})
		})

		Context("fetching succeeds", func() {
			var (
				res *http.Response
			)

			Context("decoding fails", func() {
				BeforeEach(func() {
					res = &http.Response{
						Body: ioutil.NopCloser(strings.NewReader("[")),
					}

					c.EXPECT().Get("http://reactor/doublet").Return(res, nil)
				})

				It("returns the result", func() {
					_, err := doublet.Do("http://reactor/doublet")
					Expect(err).To(MatchError("failed to decode challenge result: unexpected EOF"))
				})
			})

			Context("decoding succeeds", func() {
				BeforeEach(func() {
					res = &http.Response{
						Body: ioutil.NopCloser(strings.NewReader(`
							[
								"doublet"
							]
						`)),
					}

					c.EXPECT().Get("http://reactor/doublet").Return(res, nil)
				})

				It("returns the result", func() {
					actual, err := doublet.Do("http://reactor/doublet")
					Expect(actual).To(Equal(client.DoubletResult{"doublet"}))
					Expect(err).To(BeNil())
				})
			})
		})
	})
})
