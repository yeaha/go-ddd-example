package e2e

import (
	"net/http"

	"github.com/gavv/httpexpect/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OAuth API", func() {
	var e *httpexpect.Expect

	BeforeEach(func() {
		e = newExpect()
	})

	It("生成facebook授权地址", func() {
		body := e.GET(`/login/oauth/facebook`).
			WithQuery("redirect_uri", "https://example.com/callback").
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		expectErrno(body, 0)

		nextURL := body.Value("data").Object().
			Value("next_url").String().Raw()

		Expect(nextURL).To(ContainSubstring("facebook.com"))
		Expect(nextURL).To(ContainSubstring("client_id=test-client-id"))
		Expect(nextURL).To(ContainSubstring("redirect_uri="))
	})

	It("未配置的站点返回404", func() {
		body := e.GET(`/login/oauth/google`).
			WithQuery("redirect_uri", "https://example.com/callback").
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		expectErrno(body, 40004)
	})

	It("缺少code的验证请求返回400", func() {
		// query内没有code时，应当在发起外部请求前就拒绝
		body := e.POST(`/login/oauth/facebook`).
			WithJSON(map[string]string{
				"redirect_uri": "https://example.com/callback",
				"query":        "state=foo",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		expectErrno(body, 40000)
	})

	It("无效oauth_token注册返回406", func() {
		body := e.POST(`/register/oauth`).
			WithJSON(map[string]string{
				"oauth_token": "missing",
				"email":       uniqueEmail(),
			}).
			Expect().
			Status(http.StatusNotAcceptable).
			JSON().Object()

		expectErrno(body, 40005)
	})
})
