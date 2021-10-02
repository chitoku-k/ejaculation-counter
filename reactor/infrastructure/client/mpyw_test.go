package client_test

import (
	"context"
	"net/http"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Mpyw", func() {
	var (
		server    *ghttp.Server
		serverURL string
		mpyw      client.Mpyw
	)

	BeforeEach(func() {
		server = ghttp.NewTLSServer()
		serverURL = server.URL()
		mpyw = client.NewMpyw(server.HTTPTestServer.Client())
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Do()", func() {
		Context("targetURL is incorrect", func() {
			It("returns an error", func() {
				_, err := mpyw.Do(context.Background(), ":/", 1)
				Expect(err).To(MatchError(`failed to parse given targetURL: parse ":/": missing protocol scheme`))
			})
		})

		Context("targetURL is correct", func() {
			Context("fetching fails", func() {
				BeforeEach(func() {
					server.Close()
				})

				It("returns an error", func() {
					_, err := mpyw.Do(context.Background(), serverURL+"/api", 1)
					Expect(err).To(MatchError(HavePrefix("failed to fetch challenge result:")))
				})
			})

			Context("fetching succeeds", func() {
				Context("decoding fails", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/api", "count=1"),
								ghttp.RespondWith(http.StatusOK, `{`),
							),
						)
					})

					It("returns an error", func() {
						_, err := mpyw.Do(context.Background(), serverURL+"/api", 1)
						Expect(err).To(MatchError("failed to decode challenge result: unexpected EOF"))
					})
				})

				Context("decoding succeeds", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/api", "count=1"),
								ghttp.RespondWith(http.StatusOK, `
									{
										"title": "実務経験ガチャ",
										"result": [
											"https://web.archive.org/web/20181111004435/https://detail.chiebukuro.yahoo.co.jp/qa/question_detail/q13198470468"
										]
									}
								`),
							),
						)
					})

					It("returns decoded result", func() {
						actual, err := mpyw.Do(context.Background(), serverURL+"/api", 1)
						Expect(actual.Result).To(Equal([]string{"https://web.archive.org/web/20181111004435/https://detail.chiebukuro.yahoo.co.jp/qa/question_detail/q13198470468"}))
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})
		})
	})
})
