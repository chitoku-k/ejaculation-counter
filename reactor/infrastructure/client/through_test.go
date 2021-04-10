package client_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/wrapper"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Through", func() {
	var (
		ctrl    *gomock.Controller
		c       *wrapper.MockHttpClient
		through client.Through
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = wrapper.NewMockHttpClient(ctrl)
		through = client.NewThrough(c)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Do()", func() {
		Context("fetching fails", func() {
			BeforeEach(func() {
				c.EXPECT().Get("http://reactor/through").Return(
					nil,
					errors.New(`Get "http://reactor/through": dial tcp [::1]:80: connect: connection refused`),
				)
			})

			It("returns an error", func() {
				actual, err := through.Do("http://reactor/through")
				Expect(actual).To(BeNil())
				Expect(err).To(MatchError(`failed to fetch challenge result: Get "http://reactor/through": dial tcp [::1]:80: connect: connection refused`))
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

					c.EXPECT().Get("http://reactor/through").Return(res, nil)
				})

				It("returns the result", func() {
					_, err := through.Do("http://reactor/through")
					Expect(err).To(MatchError("failed to decode challenge result: unexpected EOF"))
				})
			})

			Context("decoding succeeds", func() {
				BeforeEach(func() {
					res = &http.Response{
						Body: ioutil.NopCloser(strings.NewReader(`
							[
								"through"
							]
						`)),
					}

					c.EXPECT().Get("http://reactor/through").Return(res, nil)
				})

				It("returns the result", func() {
					actual, err := through.Do("http://reactor/through")
					Expect(actual).To(Equal(client.ThroughResult{"through"}))
					Expect(err).To(BeNil())
				})
			})
		})
	})
})
