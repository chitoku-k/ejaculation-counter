package client_test

import (
	"context"
	"net/http"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Doublet", func() {
	var (
		server    *ghttp.Server
		serverURL string
		doublet   client.Doublet
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		serverURL = server.URL()
		doublet = client.NewDoublet(server.HTTPTestServer.Client())
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
				actual, err := doublet.Do(context.Background(), serverURL+"/doublet")
				Expect(actual).To(BeNil())
				Expect(err).To(MatchError(HavePrefix("failed to fetch challenge result:")))
			})
		})

		Context("fetching succeeds", func() {
			Context("decoding fails", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/doublet"),
							ghttp.RespondWith(http.StatusOK, "["),
						),
					)
				})

				It("returns the result", func() {
					_, err := doublet.Do(context.Background(), serverURL+"/doublet")
					Expect(err).To(MatchError("failed to decode challenge result: unexpected EOF"))
				})
			})

			Context("decoding succeeds", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/doublet"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, []string{"doublet"}),
						),
					)
				})

				It("returns the result", func() {
					actual, err := doublet.Do(context.Background(), serverURL+"/doublet")
					Expect(actual).To(Equal(client.DoubletResult{"doublet"}))
					Expect(err).To(BeNil())
				})
			})
		})
	})
})
