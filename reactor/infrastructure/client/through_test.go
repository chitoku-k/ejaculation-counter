package client_test

import (
	"context"
	"net/http"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Through", func() {
	var (
		server    *ghttp.Server
		serverURL string
		through   client.Through
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		serverURL = server.URL()
		through = client.NewThrough(server.HTTPTestServer.Client())
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Do()", func() {
		Context("fetching fails", func() {
			BeforeEach(func() {
				server.Close()
			})

			It("returns an error", func() {
				actual, err := through.Do(context.Background(), serverURL+"/through")
				Expect(actual).To(BeNil())
				Expect(err).To(MatchError(HavePrefix("failed to fetch challenge result:")))
			})
		})

		Context("fetching succeeds", func() {
			Context("decoding fails", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/through"),
							ghttp.RespondWith(http.StatusOK, "["),
						),
					)
				})

				It("returns the result", func() {
					_, err := through.Do(context.Background(), serverURL+"/through")
					Expect(err).To(MatchError("failed to decode challenge result: unexpected EOF"))
				})
			})

			Context("decoding succeeds", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/through"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, []string{"through"}),
						),
					)
				})

				It("returns the result", func() {
					actual, err := through.Do(context.Background(), serverURL+"/through")
					Expect(actual).To(Equal(client.ThroughResult{"through"}))
					Expect(err).To(BeNil())
				})
			})
		})
	})
})
