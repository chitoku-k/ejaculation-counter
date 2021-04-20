package client_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/wrapper"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mpyw", func() {
	var (
		ctrl *gomock.Controller
		c    *wrapper.MockHttpClient
		mpyw client.Mpyw
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = wrapper.NewMockHttpClient(ctrl)
		mpyw = client.NewMpyw(c)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Do()", func() {
		Context("targetURL is incorrect", func() {
			It("returns an error", func() {
				_, err := mpyw.Do(":/", 1)
				Expect(err).To(MatchError(`failed to parse given targetURL: parse ":/": missing protocol scheme`))
			})
		})

		Context("targetURL is correct", func() {
			Context("fetching fails", func() {
				BeforeEach(func() {
					c.EXPECT().Get("http://[::1]?count=1").Return(
						nil,
						errors.New(`Get "http://[::1]?count=1": dial tcp [::1]:80: connect: connection refused`),
					)
				})

				It("returns an error", func() {
					_, err := mpyw.Do("http://[::1]", 1)
					Expect(err).To(MatchError(`failed to fetch challenge result: Get "http://[::1]?count=1": dial tcp [::1]:80: connect: connection refused`))
				})
			})

			Context("fetching succeeds", func() {
				var (
					res *http.Response
				)

				Context("decoding fails", func() {
					BeforeEach(func() {
						res = &http.Response{
							Body: io.NopCloser(bytes.NewBufferString("{")),
						}
						c.EXPECT().Get("https://mpyw.kb10uy.org/api?count=1").Return(res, nil)
					})

					It("returns an error", func() {
						_, err := mpyw.Do("https://mpyw.kb10uy.org/api", 1)
						Expect(err).To(MatchError("failed to decode challenge result: unexpected EOF"))
					})
				})

				Context("decoding succeeds", func() {
					BeforeEach(func() {
						res = &http.Response{
							Body: io.NopCloser(bytes.NewBufferString(`
								{
									"title": "実務経験ガチャ",
									"result": [
										"https://web.archive.org/web/20181111004435/https://detail.chiebukuro.yahoo.co.jp/qa/question_detail/q13198470468"
									]
								}
							`)),
						}
						c.EXPECT().Get("https://mpyw.kb10uy.org/api?count=1").Return(res, nil)
					})

					It("returns decoded result", func() {
						actual, err := mpyw.Do("https://mpyw.kb10uy.org/api", 1)
						Expect(actual.Result).To(Equal([]string{"https://web.archive.org/web/20181111004435/https://detail.chiebukuro.yahoo.co.jp/qa/question_detail/q13198470468"}))
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})
		})
	})
})
