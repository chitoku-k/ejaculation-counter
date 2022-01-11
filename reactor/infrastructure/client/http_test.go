package client_test

import (
	"net/http"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HttpClient", func() {
	Describe("NewHttpClient()", func() {
		Context("when creating", func() {
			var (
				httpClient *http.Client
				err        error
			)

			BeforeEach(func() {
				httpClient, err = client.NewHttpClient()
			})

			It("returns client", func() {
				Expect(httpClient.Jar).NotTo(BeNil())
				Expect(httpClient.Transport).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
