package config_test

import (
	"os"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
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

		Context("all required vars set", func() {
			BeforeEach(func() {
				os.Setenv("MASTODON_SERVER_URL", "mastodon")
				os.Setenv("MASTODON_STREAM", "direct")
				os.Setenv("MASTODON_ACCESS_TOKEN", "token")
				os.Setenv("MQ_HOST", "mq")
				os.Setenv("MQ_USERNAME", "shiko")
				os.Setenv("MQ_PASSWORD", "shiko")
				os.Setenv("PORT", "8080")
			})

			It("returns config", func() {
				env, err := config.Get()
				Expect(env).To(Equal(config.Environment{
					Mastodon: config.Mastodon{
						ServerURL:   "mastodon",
						Stream:      "direct",
						AccessToken: "token",
					},
					Queue: config.Queue{
						Host:     "mq",
						Username: "shiko",
						Password: "shiko",
					},
					Port:     "8080",
					LogLevel: logrus.InfoLevel,
				}))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("all vars set", func() {
			Context("invalid log level is given", func() {
				BeforeEach(func() {
					os.Setenv("MASTODON_SERVER_URL", "mastodon")
					os.Setenv("MASTODON_STREAM", "direct")
					os.Setenv("MASTODON_ACCESS_TOKEN", "token")
					os.Setenv("MQ_HOST", "mq")
					os.Setenv("MQ_USERNAME", "shiko")
					os.Setenv("MQ_PASSWORD", "shiko")
					os.Setenv("PORT", "8080")
					os.Setenv("LOG_LEVEL", "unknown")
				})

				It("returns an error", func() {
					_, err := config.Get()
					Expect(err).To(MatchError(MatchRegexp("^failed to parse log level: ")))
				})
			})

			Context("valid log level is given", func() {
				BeforeEach(func() {
					os.Setenv("MASTODON_SERVER_URL", "mastodon")
					os.Setenv("MASTODON_STREAM", "direct")
					os.Setenv("MASTODON_ACCESS_TOKEN", "token")
					os.Setenv("MQ_HOST", "mq")
					os.Setenv("MQ_USERNAME", "shiko")
					os.Setenv("MQ_PASSWORD", "shiko")
					os.Setenv("PORT", "8080")
					os.Setenv("LOG_LEVEL", "debug")
				})

				It("returns config", func() {
					env, err := config.Get()
					Expect(env).To(Equal(config.Environment{
						Mastodon: config.Mastodon{
							ServerURL:   "mastodon",
							Stream:      "direct",
							AccessToken: "token",
						},
						Queue: config.Queue{
							Host:     "mq",
							Username: "shiko",
							Password: "shiko",
						},
						Port:     "8080",
						LogLevel: logrus.DebugLevel,
					}))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
