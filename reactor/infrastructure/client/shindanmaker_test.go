package client_test

import (
	"context"
	"net/http"
	"net/url"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Shindanmaker", func() {
	var (
		server       *ghttp.Server
		serverURL    string
		shindanmaker client.Shindanmaker
	)

	BeforeEach(func() {
		server = ghttp.NewTLSServer()
		serverURL = server.URL()
		shindanmaker = client.NewShindanmaker(server.HTTPTestServer.Client())
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Name()", func() {
		Context("DisplayName is not empty", func() {
			var (
				account service.Account
			)

			BeforeEach(func() {
				account = service.Account{
					DisplayName: "テスト",
					Username:    "test",
				}
			})

			It("returns DisplayName", func() {
				actual := shindanmaker.Name(account)
				Expect(actual).To(Equal("テスト"))
			})
		})

		Context("DisplayName is empty", func() {
			var (
				account service.Account
			)

			BeforeEach(func() {
				account = service.Account{
					DisplayName: "",
					Username:    "test",
				}
			})

			It("returns Username", func() {
				actual := shindanmaker.Name(account)
				Expect(actual).To(Equal("test"))
			})
		})
	})

	Describe("Do()", func() {
		Describe("token", func() {
			Context("fetching fails", func() {
				BeforeEach(func() {
					server.Close()
				})

				It("returns an error", func() {
					actual, err := shindanmaker.Do(context.Background(), "テスト", serverURL+"/a/855159")
					Expect(actual).To(BeEmpty())
					Expect(err).To(MatchError(HavePrefix("failed to fetch shindan page:")))
				})
			})

			Context("fetching succeeds", func() {
				Context("parsing fails", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusForbidden, `
									<html>
									<head><title>403 Forbidden</title></head>
									<body bgcolor="white">
									<center><h1>403 Forbidden</h1></center>
									</body>
									</html>
									<!-- a padding to disable MSIE and Chrome friendly error page -->
									<!-- a padding to disable MSIE and Chrome friendly error page -->
									<!-- a padding to disable MSIE and Chrome friendly error page -->
									<!-- a padding to disable MSIE and Chrome friendly error page -->
									<!-- a padding to disable MSIE and Chrome friendly error page -->
									<!-- a padding to disable MSIE and Chrome friendly error page -->
								`),
							),
						)
					})

					It("returns an error", func() {
						actual, err := shindanmaker.Do(context.Background(), "テスト", serverURL+"/a/855159")
						Expect(actual).To(BeEmpty())
						Expect(err).To(MatchError("failed to parse shindan page"))
					})
				})

				Context("parsing succeeds", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"テスト"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes token", func() {
						shindanmaker.Do(context.Background(), "テスト", serverURL+"/a/855159")
					})
				})
			})
		})

		Describe("escape name", func() {
			Context("name does not include special characters", func() {
				Context("name includes neither at-sign nor parentheses", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"テスト"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name", func() {
						shindanmaker.Do(context.Background(), "テスト", serverURL+"/a/855159")
					})
				})

				Context("name begins with half-width at-sign", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"@test"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name", func() {
						shindanmaker.Do(context.Background(), "@test", serverURL+"/a/855159")
					})
				})

				Context("name begins with full-width at-sign", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"＠テスト"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name", func() {
						shindanmaker.Do(context.Background(), "＠テスト", serverURL+"/a/855159")
					})
				})

				Context("name begins with half-width parentheses", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"(テスト)"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name", func() {
						shindanmaker.Do(context.Background(), "(テスト)", serverURL+"/a/855159")
					})
				})

				Context("name begins with full-width at-sign", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"（テスト）"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name", func() {
						shindanmaker.Do(context.Background(), "（テスト）", serverURL+"/a/855159")
					})
				})

				Context("name includes half-width at-sign", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"テスト"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name before half-width at-sign", func() {
						shindanmaker.Do(context.Background(), "テスト@がんばらない", serverURL+"/a/855159")
					})
				})

				Context("name includes full-width at-sign", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"テスト"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name before full-width at-sign", func() {
						shindanmaker.Do(context.Background(), "テスト＠がんばらない", serverURL+"/a/855159")
					})
				})

				Context("name includes half-width parentheses", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"テスト"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name before half-width parentheses", func() {
						shindanmaker.Do(context.Background(), "テスト(昨日: 1 / 今日: 1)", serverURL+"/a/855159")
					})
				})

				Context("name includes full-width parentheses", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"テスト"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name before full-width parentheses", func() {
						shindanmaker.Do(context.Background(), "テスト（昨日: 1 / 今日: 1）", serverURL+"/a/855159")
					})
				})
			})

			Context("name includes special characters", func() {
				Context("name includes neither at-sign nor parentheses", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"\\$1\\\\1\\${10}\\\\{10}"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name", func() {
						shindanmaker.Do(context.Background(), "$1\\1${10}\\{10}", serverURL+"/a/855159")
					})
				})

				Context("name includes half-width at-sign", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"\\$1\\\\1\\${10}\\\\{10}"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name before half-width at-sign", func() {
						shindanmaker.Do(context.Background(), "$1\\1${10}\\{10}@がんばらない", serverURL+"/a/855159")
					})
				})

				Context("name includes full-width at-sign", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"\\$1\\\\1\\${10}\\\\{10}"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name before full-width at-sign", func() {
						shindanmaker.Do(context.Background(), "$1\\1${10}\\{10}＠がんばらない", serverURL+"/a/855159")
					})
				})

				Context("name includes half-width parentheses", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"\\$1\\\\1\\${10}\\\\{10}"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name before half-width parentheses", func() {
						shindanmaker.Do(context.Background(), "$1\\1${10}\\{10}(昨日: 1 / 今日: 1)", serverURL+"/a/855159")
					})
				})

				Context("name includes full-width parentheses", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"\\$1\\\\1\\${10}\\\\{10}"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
							),
						)
					})

					It("passes name before full-width parentheses", func() {
						shindanmaker.Do(context.Background(), "$1\\1${10}\\{10}（昨日: 1 / 今日: 1）", serverURL+"/a/855159")
					})
				})
			})
		})

		Describe("fetch", func() {
			Context("fetching fails", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
							ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
							ghttp.RespondWith(http.StatusOK, `
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`),
						),
						func(w http.ResponseWriter, r *http.Request) {
							c, _, err := w.(http.Hijacker).Hijack()
							Expect(err).NotTo(HaveOccurred())
							Expect(c.Close()).NotTo(HaveOccurred())
						},
					)
				})

				It("returns an error", func() {
					actual, err := shindanmaker.Do(context.Background(), "テスト", serverURL+"/a/855159")
					Expect(actual).To(Equal(""))
					Expect(err).To(MatchError(HavePrefix("failed to fetch shindan result:")))
				})
			})

			Context("fetching succeeds", func() {
				Context("parsing fails", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.RespondWith(http.StatusOK, `
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest(http.MethodPost, "/855159"),
								ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
								ghttp.VerifyForm(url.Values{
									"type":        []string{"name"},
									"shindanName": []string{"テスト"},
									"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								}),
								ghttp.RespondWith(http.StatusForbidden, `
									<html>
									<head><title>403 Forbidden</title></head>
									<body bgcolor="white">
									<center><h1>403 Forbidden</h1></center>
									</body>
									</html>
									<!-- a padding to disable MSIE and Chrome friendly error page -->
									<!-- a padding to disable MSIE and Chrome friendly error page -->
									<!-- a padding to disable MSIE and Chrome friendly error page -->
									<!-- a padding to disable MSIE and Chrome friendly error page -->
									<!-- a padding to disable MSIE and Chrome friendly error page -->
									<!-- a padding to disable MSIE and Chrome friendly error page -->
								`),
							),
						)
					})

					It("returns an error", func() {
						actual, err := shindanmaker.Do(context.Background(), "テスト", serverURL+"/a/855159")
						Expect(actual).To(BeEmpty())
						Expect(err).To(MatchError("failed to parse shindan result"))
					})
				})

				Context("parsing succeeds", func() {
					Context("result does not include special characters", func() {
						Context("result is less than 140 characters", func() {
							BeforeEach(func() {
								server.AppendHandlers(
									ghttp.CombineHandlers(
										ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
										ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
										ghttp.RespondWith(http.StatusOK, `
											<!doctype html>
											<html lang="ja">
											<head>
												<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
												<title>ちんぽ揃えゲーム</title>
											</head>
											<body>
											</body>
											</html>
										`),
									),
									ghttp.CombineHandlers(
										ghttp.VerifyRequest(http.MethodPost, "/855159"),
										ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
										ghttp.VerifyForm(url.Values{
											"type":        []string{"name"},
											"shindanName": []string{"テスト"},
											"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
										}),
										ghttp.RespondWith(http.StatusOK, `
											<!doctype html>
											<html lang="ja">
											<head>
												<title>ちんぽ揃えゲーム</title>
											</head>
											<body>
												<div id="main-container">
													<div id="main">
														<div class="modal fade" id="shareModal">
															<div class="modal-dialog modal-dialog-scrollable modal-md modal-dialog-centered">
																<div class="modal-content">
																	<div class="modal-body">
																		<div class="mb-2">
																			<div class="tab-content" id="copyContent">
																				<div class="tab-pane fade show active" id="copy_140">
																					<div class="form-group mb-2">
																						<textarea class="form-control border-top-0 nav-tabs-copy-textarea" id="copy-textarea-140" rows="5">ちんんんんぽんんぽちぽちちぽぽぽちんぽ(ﾎﾞﾛﾝ&#10;&#10;テストさんは19文字目でちんぽを出せました！&#10;&#10;#ちんぽ揃えゲーム&ensp;#shindanmaker&#10;https://shindanmaker.com/855159</textarea>
																					</div>
																				</div>
																			</div>
																		</div>
																	</div>
																</div>
															</div>
														</div>
													</div>
												</div>
											</body>
											</html>
										`),
									),
								)
							})

							It("returns the result", func() {
								actual, err := shindanmaker.Do(context.Background(), "テスト", serverURL+"/a/855159")
								Expect(actual).To(Equal(`ちんんんんぽんんぽちぽちちぽぽぽちんぽ(ﾎﾞﾛﾝ

テストさんは19文字目でちんぽを出せました！

#ちんぽ揃えゲーム #shindanmaker
https://shindanmaker.com/855159`))
								Expect(err).NotTo(HaveOccurred())
							})
						})

						Context("result exceeds 140 characters", func() {
							BeforeEach(func() {
								server.AppendHandlers(
									ghttp.CombineHandlers(
										ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
										ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
										ghttp.RespondWith(http.StatusOK, `
											<!doctype html>
											<html lang="ja">
											<head>
												<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
												<title>ちんぽ揃えゲーム</title>
											</head>
											<body>
											</body>
											</html>
										`),
									),
									ghttp.CombineHandlers(
										ghttp.VerifyRequest(http.MethodPost, "/855159"),
										ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
										ghttp.VerifyForm(url.Values{
											"type":        []string{"name"},
											"shindanName": []string{"テスト"},
											"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
										}),
										ghttp.RespondWith(http.StatusOK, `
											<!DOCTYPE html>
											<html lang="ja">
											<head>
												<title>ちんぽ揃えゲーム</title>
											</head>
											<body>
												<div id="main-container">
													<div id="main">
														<div class="modal fade" id="shareModal">
															<div class="modal-dialog modal-dialog-scrollable modal-md modal-dialog-centered">
																<div class="modal-content">
																	<div class="modal-body">
																		<div class="mb-2">
																			<div class="tab-content" id="copyContent">
																				<div class="tab-pane fade show active" id="copy_140">
																					<div class="form-group mb-2">
																						<textarea class="form-control border-top-0 nav-tabs-copy-textarea" id="copy-textarea-140" rows="5">んちちんんぽんちちちぽんちちんんんぽぽちちぽちぽぽぽぽんぽぽちんんぽんんんんちちぽぽちんちちんんぽんぽちちぽちぽんんちぽぽんんちんんちんちちぽんんんちちぽちちちちぽちぽんんぽんぽちちぽんちんちちぽんんちんんんぽちんんぽぽ…&#10;#ちんぽ揃えゲーム&ensp;#shindanmaker&#10;https://shindanmaker.com/855159</textarea>
																					</div>
																				</div>
																				<div class="tab-pane fade" id="copy_all">
																					<div class="form-group mb-2">
																						<textarea class="form-control border-top-0 nav-tabs-copy-textarea" id="copy-textarea-all" rows="5">んちちんんぽんちちちぽんちちんんんぽぽちちぽちぽぽぽぽんぽぽちんんぽんんんんちちぽぽちんちちんんぽんぽちちぽちぽんんちぽぽんんちんんちんちちぽんんんちちぽちちちちぽちぽんんぽんぽちちぽんちんちちぽんんちんんんぽちんんぽぽちんんちぽちんんちぽんちちぽぽぽちぽんんちんちちちちぽぽぽぽぽぽんぽんんちちちんちぽぽちぽちぽんんちちぽんんちんんぽちちんぽ(ﾎﾞﾛﾝ&#10;&#10;テストさんは172文字目でちんぽを出せました！&#10;&#10;#ちんぽ揃えゲーム&ensp;#shindanmaker&#10;https://shindanmaker.com/855159</textarea>
																					</div>
																				</div>
																			</div>
																		</div>
																	</div>
																</div>
															</div>
														</div>
													</div>
												</div>
											</body>
											</html>
										`),
									),
								)
							})

							It("returns the result", func() {
								actual, err := shindanmaker.Do(context.Background(), "テスト", serverURL+"/a/855159")
								Expect(actual).To(Equal(`んちちんんぽんちちちぽんちちんんんぽぽちちぽちぽぽぽぽんぽぽちんんぽんんんんちちぽぽちんちちんんぽんぽちちぽちぽんんちぽぽんんちんんちんちちぽんんんちちぽちちちちぽちぽんんぽんぽちちぽんちんちちぽんんちんんんぽちんんぽぽ…
#ちんぽ揃えゲーム #shindanmaker
https://shindanmaker.com/855159`))
								Expect(err).NotTo(HaveOccurred())
							})
						})
					})

					Context("result includes special characters", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								ghttp.CombineHandlers(
									ghttp.VerifyRequest(http.MethodGet, "/a/855159"),
									ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
									ghttp.RespondWith(http.StatusOK, `
										<!doctype html>
										<html lang="ja">
										<head>
											<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
											<title>ちんぽ揃えゲーム</title>
										</head>
										<body>
										</body>
										</html>
									`),
								),
								ghttp.CombineHandlers(
									ghttp.VerifyRequest(http.MethodPost, "/855159"),
									ghttp.VerifyHeaderKV("User-Agent", "Mozilla/5.0 (compatible)"),
									ghttp.VerifyForm(url.Values{
										"type":        []string{"name"},
										"shindanName": []string{`<>"'&`},
										"_token":      []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
									}),
									ghttp.RespondWith(http.StatusOK, `
										<!doctype html>
										<html lang="ja">
										<head>
											<title>ちんぽ揃えゲーム</title>
										</head>
										<body>
											<div id="main-container">
												<div id="main">
													<div class="modal fade" id="shareModal">
														<div class="modal-dialog modal-dialog-scrollable modal-md modal-dialog-centered">
															<div class="modal-content">
																<div class="modal-body">
																	<div class="mb-2">
																		<div class="tab-content" id="copyContent">
																			<div class="tab-pane fade show active" id="copy_140">
																				<div class="form-group mb-2">
																					<textarea class="form-control border-top-0 nav-tabs-copy-textarea" id="copy-textarea-140" rows="5">ちんんんんぽんんぽちぽちちぽぽぽちんぽ(ﾎﾞﾛﾝ&#10;&#10;&lt;&gt;&quot;&#039;&amp;さんは19文字目でちんぽを出せました！&#10;&#10;#ちんぽ揃えゲーム&ensp;#shindanmaker&#10;https://shindanmaker.com/855159</textarea>
																				</div>
																			</div>
																		</div>
																	</div>
																</div>
															</div>
														</div>
													</div>
												</div>
											</div>
										</body>
										</html>
									`),
								),
							)
						})

						It("returns the result", func() {
							actual, err := shindanmaker.Do(context.Background(), `<>"'&`, serverURL+"/a/855159")
							Expect(actual).To(Equal(`ちんんんんぽんんぽちぽちちぽぽぽちんぽ(ﾎﾞﾛﾝ

<>"'&さんは19文字目でちんぽを出せました！

#ちんぽ揃えゲーム #shindanmaker
https://shindanmaker.com/855159`))
							Expect(err).NotTo(HaveOccurred())
						})
					})
				})
			})
		})
	})
})
