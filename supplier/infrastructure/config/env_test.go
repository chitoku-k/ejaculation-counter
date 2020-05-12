package config_test

import (
	"os"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Environment", func() {
	Describe("Get()", func() {
		Context("some vars missing", func() {
			BeforeEach(func() {
				os.Clearenv()
			})

			It("returns an error", func() {
				_, err := config.Get()
				Expect(err).To(MatchError(MatchRegexp("^missing env\\(s\\): ")))
			})
		})

		Context("all vars set", func() {
			BeforeEach(func() {
				os.Setenv("MASTODON_USER_ID", "user1")
				os.Setenv("MASTODON_SERVER_URL", "mastodon")
				os.Setenv("MASTODON_STREAM", "direct")
				os.Setenv("MASTODON_ACCESS_TOKEN", "token")
				os.Setenv("MQ_HOST", "mq")
				os.Setenv("MQ_USERNAME", "shiko")
				os.Setenv("MQ_PASSWORD", "shiko")
				os.Setenv("REACTOR_HOST", "reactor")
				os.Setenv("PORT", "8080")
			})

			It("returns config", func() {
				env, err := config.Get()
				Expect(env).To(Equal(config.Environment{
					Mastodon: config.Mastodon{
						UserID:      "user1",
						ServerURL:   "mastodon",
						Stream:      "direct",
						AccessToken: "token",
					},
					Reactor: config.Reactor{
						Host: "reactor",
					},
					Queue: config.Queue{
						Host:     "mq",
						Username: "shiko",
						Password: "shiko",
					},
					Port: "8080",
				}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
