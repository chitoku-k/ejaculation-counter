package client_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/wrapper"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Shindanmaker", func() {
	var (
		ctrl         *gomock.Controller
		c            *wrapper.MockHttpClient
		r            *wrapper.MockReader
		shindanmaker client.Shindanmaker
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = wrapper.NewMockHttpClient(ctrl)
		r = wrapper.NewMockReader(ctrl)
		shindanmaker = client.NewShindanmaker(c)
	})

	AfterEach(func() {
		ctrl.Finish()
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
		var (
			top *http.Response
			res *http.Response
		)

		Describe("token", func() {
			Context("fetching fails", func() {
				BeforeEach(func() {
					c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(
						nil,
						errors.New("error"),
					)
				})

				It("returns an error", func() {
					actual, err := shindanmaker.Do("テスト", "https://shindanmaker.com/a/855159")
					Expect(actual).To(BeEmpty())
					Expect(err).To(MatchError("failed to fetch shindan top: error"))
				})
			})

			Context("fetching succeeds", func() {
				Context("reading fails", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(r),
						}
						r.EXPECT().Read(gomock.Any()).Return(
							0,
							errors.New("dial tcp [::1]:443: connect: connection refused"),
						)

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(
							top,
							nil,
						)
					})

					It("returns an error", func() {
						actual, err := shindanmaker.Do("テスト", "https://shindanmaker.com/a/855159")
						Expect(actual).To(BeEmpty())
						Expect(err).To(MatchError("failed to read shindan top: dial tcp [::1]:443: connect: connection refused"))
					})
				})

				Context("reading succeeds", func() {
					Context("parsing fails", func() {
						BeforeEach(func() {
							top = &http.Response{
								Body: ioutil.NopCloser(strings.NewReader(`
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
								`)),
							}

							c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(
								top,
								nil,
							)
						})

						It("returns an error", func() {
							actual, err := shindanmaker.Do("テスト", "https://shindanmaker.com/a/855159")
							Expect(actual).To(BeEmpty())
							Expect(err).To(MatchError("failed to parse shindan top"))
						})
					})

					Context("parsing succeeds", func() {
						BeforeEach(func() {
							top = &http.Response{
								Body: ioutil.NopCloser(strings.NewReader(`
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`)),
							}

							c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
							c.EXPECT().PostForm(
								"https://shindanmaker.com/855159",
								url.Values{
									"name":   []string{"テスト"},
									"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								},
							).Return(
								nil,
								errors.New("error"),
							)
						})

						It("passes token", func() {
							shindanmaker.Do("テスト", "https://shindanmaker.com/a/855159")
						})
					})
				})
			})
		})

		Describe("escape name", func() {
			Context("name does not include special characters", func() {
				Context("name includes neither at-sign nor parentheses", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"テスト"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name", func() {
						shindanmaker.Do("テスト", "https://shindanmaker.com/a/855159")
					})
				})

				Context("name begins with half-width at-sign", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"@test"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name", func() {
						shindanmaker.Do("@test", "https://shindanmaker.com/a/855159")
					})
				})

				Context("name begins with full-width at-sign", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"＠テスト"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name", func() {
						shindanmaker.Do("＠テスト", "https://shindanmaker.com/a/855159")
					})
				})

				Context("name begins with half-width parentheses", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"(テスト)"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name", func() {
						shindanmaker.Do("(テスト)", "https://shindanmaker.com/a/855159")
					})
				})

				Context("name begins with full-width at-sign", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"（テスト）"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name", func() {
						shindanmaker.Do("（テスト）", "https://shindanmaker.com/a/855159")
					})
				})

				Context("name includes half-width at-sign", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"テスト"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name before half-width at-sign", func() {
						shindanmaker.Do("テスト@がんばらない", "https://shindanmaker.com/a/855159")
					})
				})

				Context("name includes full-width at-sign", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"テスト"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name before full-width at-sign", func() {
						shindanmaker.Do("テスト＠がんばらない", "https://shindanmaker.com/a/855159")
					})
				})

				Context("name includes half-width parentheses", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"テスト"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name before half-width parentheses", func() {
						shindanmaker.Do("テスト(昨日: 1 / 今日: 1)", "https://shindanmaker.com/a/855159")
					})
				})

				Context("name includes full-width parentheses", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"テスト"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name before full-width parentheses", func() {
						shindanmaker.Do("テスト（昨日: 1 / 今日: 1）", "https://shindanmaker.com/a/855159")
					})
				})
			})

			Context("name includes special characters", func() {
				Context("name includes neither at-sign nor parentheses", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"\\$1\\\\1\\${10}\\\\{10}"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name", func() {
						shindanmaker.Do("$1\\1${10}\\{10}", "https://shindanmaker.com/a/855159")
					})
				})

				Context("name includes half-width at-sign", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"\\$1\\\\1\\${10}\\\\{10}"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name before half-width at-sign", func() {
						shindanmaker.Do("$1\\1${10}\\{10}@がんばらない", "https://shindanmaker.com/a/855159")
					})
				})

				Context("name includes full-width at-sign", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"\\$1\\\\1\\${10}\\\\{10}"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name before full-width at-sign", func() {
						shindanmaker.Do("$1\\1${10}\\{10}＠がんばらない", "https://shindanmaker.com/a/855159")
					})
				})

				Context("name includes half-width parentheses", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"\\$1\\\\1\\${10}\\\\{10}"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name before half-width parentheses", func() {
						shindanmaker.Do("$1\\1${10}\\{10}(昨日: 1 / 今日: 1)", "https://shindanmaker.com/a/855159")
					})
				})

				Context("name includes full-width parentheses", func() {
					BeforeEach(func() {
						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"\\$1\\\\1\\${10}\\\\{10}"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(
							nil,
							errors.New("error"),
						)
					})

					It("passes name before full-width parentheses", func() {
						shindanmaker.Do("$1\\1${10}\\{10}（昨日: 1 / 今日: 1）", "https://shindanmaker.com/a/855159")
					})
				})
			})
		})

		Describe("fetch", func() {
			Context("fetching fails", func() {
				BeforeEach(func() {
					top = &http.Response{
						Body: ioutil.NopCloser(strings.NewReader(`
							<!doctype html>
							<html lang="ja">
							<head>
								<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
								<title>ちんぽ揃えゲーム</title>
							</head>
							<body>
							</body>
							</html>
						`)),
					}

					c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
					c.EXPECT().PostForm(
						"https://shindanmaker.com/855159",
						url.Values{
							"name":   []string{"テスト"},
							"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
						},
					).Return(
						nil,
						errors.New(`Get "https://shindanmaker.com/855159": dial tcp [::1]:443: connect: connection refused`),
					)
				})

				It("returns an error", func() {
					actual, err := shindanmaker.Do("テスト", "https://shindanmaker.com/a/855159")
					Expect(actual).To(Equal(""))
					Expect(err).To(MatchError(`failed to fetch shindan result: Get "https://shindanmaker.com/855159": dial tcp [::1]:443: connect: connection refused`))
				})
			})

			Context("fetching succeeds", func() {
				Context("reading fails", func() {
					BeforeEach(func() {
						r.EXPECT().Read(gomock.Any()).Return(
							0,
							errors.New("dial tcp [::1]:443: connect: connection refused"),
						)

						top = &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(`
								<!doctype html>
								<html lang="ja">
								<head>
									<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
									<title>ちんぽ揃えゲーム</title>
								</head>
								<body>
								</body>
								</html>
							`)),
						}
						res = &http.Response{
							Body: ioutil.NopCloser(r),
						}

						c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
						c.EXPECT().PostForm(
							"https://shindanmaker.com/855159",
							url.Values{
								"name":   []string{"テスト"},
								"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
							},
						).Return(res, nil)
					})

					It("returns an error", func() {
						actual, err := shindanmaker.Do("テスト", "https://shindanmaker.com/a/855159")
						Expect(actual).To(BeEmpty())
						Expect(err).To(MatchError("failed to read shindan result: dial tcp [::1]:443: connect: connection refused"))
					})
				})

				Context("reading succeeds", func() {
					Context("parsing fails", func() {
						BeforeEach(func() {
							top = &http.Response{
								Body: ioutil.NopCloser(strings.NewReader(`
									<!doctype html>
									<html lang="ja">
									<head>
										<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
										<title>ちんぽ揃えゲーム</title>
									</head>
									<body>
									</body>
									</html>
								`)),
							}
							res = &http.Response{
								Body: ioutil.NopCloser(strings.NewReader(`
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
								`)),
							}

							c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
							c.EXPECT().PostForm(
								"https://shindanmaker.com/855159",
								url.Values{
									"name":   []string{"テスト"},
									"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
								},
							).Return(res, nil)
						})

						It("returns an error", func() {
							actual, err := shindanmaker.Do("テスト", "https://shindanmaker.com/a/855159")
							Expect(actual).To(BeEmpty())
							Expect(err).To(MatchError("failed to parse shindan result"))
						})
					})

					Context("parsing succeeds", func() {
						Context("result does not include special characters", func() {
							Context("result is less than 140 characters", func() {
								BeforeEach(func() {
									top = &http.Response{
										Body: ioutil.NopCloser(strings.NewReader(`
											<!doctype html>
											<html lang="ja">
											<head>
												<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
												<title>ちんぽ揃えゲーム</title>
											</head>
											<body>
											</body>
											</html>
										`)),
									}
									res = &http.Response{
										Body: ioutil.NopCloser(strings.NewReader(`
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
										`)),
									}

									c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
									c.EXPECT().PostForm(
										"https://shindanmaker.com/855159",
										url.Values{
											"name":   []string{"テスト"},
											"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
										},
									).Return(res, nil)
								})

								It("returns the result", func() {
									actual, err := shindanmaker.Do("テスト", "https://shindanmaker.com/a/855159")
									Expect(actual).To(Equal(`ちんんんんぽんんぽちぽちちぽぽぽちんぽ(ﾎﾞﾛﾝ

テストさんは19文字目でちんぽを出せました！

#ちんぽ揃えゲーム #shindanmaker
https://shindanmaker.com/855159`))
									Expect(err).NotTo(HaveOccurred())
								})
							})

							Context("result exceeds 140 characters", func() {
								BeforeEach(func() {
									top = &http.Response{
										Body: ioutil.NopCloser(strings.NewReader(`
											<!doctype html>
											<html lang="ja">
											<head>
												<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
												<title>ちんぽ揃えゲーム</title>
											</head>
											<body>
											</body>
											</html>
										`)),
									}
									res = &http.Response{
										Body: ioutil.NopCloser(strings.NewReader(`
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
										`)),
									}

									c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
									c.EXPECT().PostForm(
										"https://shindanmaker.com/855159",
										url.Values{
											"name":   []string{"テスト"},
											"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
										},
									).Return(res, nil)
								})

								It("returns the result", func() {
									actual, err := shindanmaker.Do("テスト", "https://shindanmaker.com/a/855159")
									Expect(actual).To(Equal(`んちちんんぽんちちちぽんちちんんんぽぽちちぽちぽぽぽぽんぽぽちんんぽんんんんちちぽぽちんちちんんぽんぽちちぽちぽんんちぽぽんんちんんちんちちぽんんんちちぽちちちちぽちぽんんぽんぽちちぽんちんちちぽんんちんんんぽちんんぽぽ…
#ちんぽ揃えゲーム #shindanmaker
https://shindanmaker.com/855159`))
									Expect(err).NotTo(HaveOccurred())
								})
							})
						})

						Context("result includes special characters", func() {
							BeforeEach(func() {
								top = &http.Response{
									Body: ioutil.NopCloser(strings.NewReader(`
										<!doctype html>
										<html lang="ja">
										<head>
											<meta name="csrf-token" content="theQuickBrownFoxJumpsOverTheLazyDog">
											<title>ちんぽ揃えゲーム</title>
										</head>
										<body>
										</body>
										</html>
									`)),
								}
								res = &http.Response{
									Body: ioutil.NopCloser(strings.NewReader(`
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
									`)),
								}

								c.EXPECT().Get("https://shindanmaker.com/a/855159").Return(top, nil)
								c.EXPECT().PostForm(
									"https://shindanmaker.com/855159",
									url.Values{
										"name":   []string{`<>"'&`},
										"_token": []string{"theQuickBrownFoxJumpsOverTheLazyDog"},
									},
								).Return(res, nil)
							})

							It("returns the result", func() {
								actual, err := shindanmaker.Do(`<>"'&`, "https://shindanmaker.com/a/855159")
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
})
