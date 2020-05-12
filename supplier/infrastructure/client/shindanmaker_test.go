package client_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Reader(m gomock.Matcher) gomock.Matcher {
	return &readerMatcher{m, nil}
}

type readerMatcher struct {
	m    gomock.Matcher
	data []byte
}

func (r *readerMatcher) Matches(x interface{}) bool {
	var err error
	r.data, err = ioutil.ReadAll(x.(io.Reader))
	if err != nil {
		return false
	}
	return r.m.Matches(r.data)
}

func (r *readerMatcher) Got(got interface{}) string {
	f, ok := r.m.(gomock.GotFormatter)
	if ok {
		return "data(" + f.Got(r.data) + ")"
	}
	return fmt.Sprintf("%#v", r.data)
}

func (r *readerMatcher) String() string {
	return "data(" + r.m.String() + ")"
}

func Query(m gomock.Matcher) gomock.Matcher {
	return &queryMatcher{m}
}

type queryMatcher struct {
	m gomock.Matcher
}

func (q *queryMatcher) Matches(x interface{}) bool {
	values, err := url.ParseQuery(string(x.([]byte)))
	if err != nil {
		return false
	}
	return q.m.Matches(values)
}

func (q *queryMatcher) Got(got interface{}) string {
	values, _ := url.ParseQuery(string(got.([]byte)))
	return fmt.Sprintf("url.Values(%v)", values)
}

func (q *queryMatcher) String() string {
	return "url.Values(" + q.m.String() + ")"
}

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
		Describe("escape name", func() {
			Context("name does not include special characters", func() {
				Context("name includes neither at-sign nor parentheses", func() {
					BeforeEach(func() {
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"テスト"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"@test"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"＠テスト"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"(テスト)"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"（テスト）"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"テスト"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"テスト"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"テスト"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"テスト"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"\\$1\\\\1\\${10}\\\\{10}"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"\\$1\\\\1\\${10}\\\\{10}"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"\\$1\\\\1\\${10}\\\\{10}"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"\\$1\\\\1\\${10}\\\\{10}"},
										},
									),
								),
							),
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
						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"\\$1\\\\1\\${10}\\\\{10}"},
										},
									),
								),
							),
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
					c.EXPECT().Post(
						"https://shindanmaker.com/a/855159",
						"application/x-www-form-urlencoded",
						Reader(
							Query(
								gomock.Eq(
									url.Values{
										"u": []string{"テスト"},
									},
								),
							),
						),
					).Return(
						nil,
						errors.New(`Get "https://shindanmaker.com/a/855159": dial tcp [::1]:443: connect: connection refused`),
					)
				})

				It("returns an error", func() {
					actual, err := shindanmaker.Do("テスト", "https://shindanmaker.com/a/855159")
					Expect(actual).To(Equal(""))
					Expect(err).To(MatchError(`failed to fetch shindan result: Get "https://shindanmaker.com/a/855159": dial tcp [::1]:443: connect: connection refused`))
				})
			})

			Context("fetching succeeds", func() {
				var (
					res *http.Response
				)

				Context("reading fails", func() {
					BeforeEach(func() {
						r.EXPECT().Read(gomock.Any()).Return(
							0,
							errors.New("dial tcp [::1]:443: connect: connection refused"),
						)

						res = &http.Response{
							Body: ioutil.NopCloser(r),
						}

						c.EXPECT().Post(
							"https://shindanmaker.com/a/855159",
							"application/x-www-form-urlencoded",
							Reader(
								Query(
									gomock.Eq(
										url.Values{
											"u": []string{"テスト"},
										},
									),
								),
							),
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

							c.EXPECT().Post(
								"https://shindanmaker.com/a/855159",
								"application/x-www-form-urlencoded",
								Reader(
									Query(
										gomock.Eq(
											url.Values{
												"u": []string{"テスト"},
											},
										),
									),
								),
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
									res = &http.Response{
										Body: ioutil.NopCloser(strings.NewReader(`
											<!DOCTYPE html>
											<html lang="ja">
											<head>
												<title>ちんぽ揃えゲーム</title>
											</head>
											<body>
												<div class="tab-content">
													<div role="tabpanel" class="tab-pane active" id="copy_panel_140">
														<form id="forcopy" name="forcopy" method="post" action="/855159" enctype="multipart/form-data" onSubmit="return false">
															<textarea id="copy_text_140" class="form-control" rows="2">ちんんんんぽんんぽちぽちちぽぽぽちんぽ(ﾎﾞﾛﾝ

テストさんは19文字目でちんぽを出せました！

#ちんぽ揃えゲーム #shindanmaker
https://shindanmaker.com/855159</textarea>
														</form>
													</div>
												</div>
											</body>
											</html>
										`)),
									}

									c.EXPECT().Post(
										"https://shindanmaker.com/a/855159",
										"application/x-www-form-urlencoded",
										Reader(
											Query(
												gomock.Eq(
													url.Values{
														"u": []string{"テスト"},
													},
												),
											),
										),
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
									res = &http.Response{
										Body: ioutil.NopCloser(strings.NewReader(`
											<!DOCTYPE html>
											<html lang="ja">
											<head>
												<title>ちんぽ揃えゲーム</title>
											</head>
											<body>
												<div class="tab-content">
													<div role="tabpanel" class="tab-pane active" id="copy_panel_140">
														<form id="forcopy" name="forcopy" method="post" action="/855159" enctype="multipart/form-data" onSubmit="return false">
															<textarea id="copy_text_140" class="form-control" rows="2">んちちんんぽんちちちぽんちちんんんぽぽちちぽちぽぽぽぽんぽぽちんんぽんんんんちちぽぽちんちちんんぽんぽちちぽちぽんんちぽぽんんちんんちんちちぽんんんちちぽちちちちぽちぽんんぽんぽちちぽんちんちちぽんんちんんんぽちんんぽぽ…
#ちんぽ揃えゲーム #shindanmaker
https://shindanmaker.com/855159</textarea>
														</form>
													</div>
													<div role="tabpanel" class="tab-pane" id="copy_panel_all">
														<textarea id="copy_text" class="form-control" rows="2">んちちんんぽんちちちぽんちちんんんぽぽちちぽちぽぽぽぽんぽぽちんんぽんんんんちちぽぽちんちちんんぽんぽちちぽちぽんんちぽぽんんちんんちんちちぽんんんちちぽちちちちぽちぽんんぽんぽちちぽんちんちちぽんんちんんんぽちんんぽぽちんんちぽちんんちぽんちちぽぽぽちぽんんちんちちちちぽぽぽぽぽぽんぽんんちちちんちぽぽちぽちぽんんちちぽんんちんんぽちちんぽ(ﾎﾞﾛﾝ

テストさんは172文字目でちんぽを出せました！

#ちんぽ揃えゲーム #shindanmaker
https://shindanmaker.com/855159</textarea>
													</div>
												</div>
											</body>
											</html>
										`)),
									}

									c.EXPECT().Post(
										"https://shindanmaker.com/a/855159",
										"application/x-www-form-urlencoded",
										Reader(
											Query(
												gomock.Eq(
													url.Values{
														"u": []string{"テスト"},
													},
												),
											),
										),
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
								res = &http.Response{
									Body: ioutil.NopCloser(strings.NewReader(`
										<!DOCTYPE html>
										<html lang="ja">
										<head>
											<title>ちんぽ揃えゲーム</title>
										</head>
										<body>
											<div class="tab-content">
												<div role="tabpanel" class="tab-pane active" id="copy_panel_140">
													<form id="forcopy" name="forcopy" method="post" action="/855159" enctype="multipart/form-data" onSubmit="return false">
														<textarea id="copy_text_140" class="form-control" rows="2">ちんんんんぽんんぽちぽちちぽぽぽちんぽ(ﾎﾞﾛﾝ

&lt;&gt;&quot;&#039;&amp;さんは19文字目でちんぽを出せました！

#ちんぽ揃えゲーム #shindanmaker
https://shindanmaker.com/855159</textarea>
													</form>
												</div>
											</div>
										</body>
										</html>
									`)),
								}

								c.EXPECT().Post(
									"https://shindanmaker.com/a/855159",
									"application/x-www-form-urlencoded",
									Reader(
										Query(
											gomock.Eq(
												url.Values{
													"u": []string{`<>"'&`},
												},
											),
										),
									),
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
