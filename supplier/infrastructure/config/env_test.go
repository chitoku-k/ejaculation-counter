package config_test

import (
	"log/slog"
	"os"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	. "github.com/onsi/ginkgo/v2"
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
				Expect(err).To(MatchError(HavePrefix("missing:")))
			})
		})

		Context("all required vars set", func() {
			BeforeEach(func() {
				err := os.Setenv("MASTODON_SERVER_URL", "mastodon")
				Expect(err).NotTo(HaveOccurred())

				err = os.Setenv("MASTODON_STREAM", "direct")
				Expect(err).NotTo(HaveOccurred())

				err = os.Setenv("MASTODON_ACCESS_TOKEN", "token")
				Expect(err).NotTo(HaveOccurred())

				err = os.Setenv("MQ_HOST", "mq")
				Expect(err).NotTo(HaveOccurred())

				err = os.Setenv("MQ_USERNAME", "shiko")
				Expect(err).NotTo(HaveOccurred())

				err = os.Setenv("MQ_PASSWORD", "shiko")
				Expect(err).NotTo(HaveOccurred())

				err = os.Setenv("PORT", "8080")
				Expect(err).NotTo(HaveOccurred())
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
					LogLevel: slog.LevelInfo,
				}))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("all vars set", func() {
			Context("invalid log level is given", func() {
				BeforeEach(func() {
					err := os.Setenv("MASTODON_SERVER_URL", "mastodon")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("MASTODON_STREAM", "direct")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("MASTODON_ACCESS_TOKEN", "token")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("MQ_HOST", "mq")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("MQ_USERNAME", "shiko")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("MQ_PASSWORD", "shiko")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("PORT", "8080")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("LOG_LEVEL", "unknown")
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, err := config.Get()
					Expect(err).To(MatchError(HavePrefix("LOG_LEVEL is invalid:")))
				})
			})

			Context("valid log level is given", func() {
				BeforeEach(func() {
					err := os.Setenv("MASTODON_SERVER_URL", "mastodon")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("MASTODON_STREAM", "direct")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("MASTODON_ACCESS_TOKEN", "token")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("MQ_HOST", "mq")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("MQ_USERNAME", "shiko")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("MQ_PASSWORD", "shiko")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("PORT", "8080")
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv("LOG_LEVEL", "debug")
					Expect(err).NotTo(HaveOccurred())
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
						LogLevel: slog.LevelDebug,
					}))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
